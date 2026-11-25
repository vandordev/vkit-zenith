package identityhttp

import (
	"net/netip"
	"testing"
)

func TestAuthSourceTrustedForwardedBoundary(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("192.0.2.0/24")}
	for _, tc := range []struct {
		name, remote, forwarded, want string
		valid                         bool
	}{
		{"ignore forged header from direct peer", "198.51.100.8:4100", "203.0.113.9", "198.51.100.8", true},
		{"select nearest untrusted proxy-chain address", "10.0.0.2:4100", "203.0.113.9, 192.0.2.44, 10.0.0.3", "203.0.113.9", true},
		{"reject malformed trusted header", "10.0.0.2:4100", "attacker", "", false},
		{"reject overlong chain", "10.0.0.2:4100", "203.0.113.1,203.0.113.2,203.0.113.3,203.0.113.4,203.0.113.5,203.0.113.6,203.0.113.7,203.0.113.8,203.0.113.9,203.0.113.10,203.0.113.11,203.0.113.12,203.0.113.13,203.0.113.14,203.0.113.15,203.0.113.16,203.0.113.17", "", false},
		{"reject nonnumeric remote peer", "attacker:4100", "", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveSource(tc.remote, tc.forwarded, trusted)
			if (err == nil) != tc.valid {
				t.Fatal("source policy result mismatch")
			}
			if err == nil && got != tc.want {
				t.Fatal("source resolution mismatch")
			}
		})
	}
}
