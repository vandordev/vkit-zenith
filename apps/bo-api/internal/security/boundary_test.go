package security

import (
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func TestBOCSRFSignedBinding(t *testing.T) {
	b := &Boundary{Clock: func() time.Time { return time.Unix(100000, 0) }, Secret: []byte("fixture-only-32-bytes-secret-value")}
	a := [32]byte{1}
	token, err := b.issue(a)
	if err != nil {
		t.Fatal(err)
	}
	if !b.valid(token, a) {
		t.Fatal("actual token rejected")
	}
	if b.valid(token, [32]byte{2}) || b.valid(token+"x", a) {
		t.Fatal("invalid signature/binding accepted")
	}
	b.Clock = func() time.Time { return time.Unix(101800, 0) }
	if b.valid(token, a) {
		t.Fatal("exact-expiry token accepted")
	}
}

func TestBOSourceUntrustedForwarding(t *testing.T) {
	b := &Boundary{}
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.4:1234"
	r.Header.Set("X-Forwarded-For", "203.0.113.5")
	got, err := b.Source(r)
	if err != nil || got != "192.0.2.4" {
		t.Fatal("untrusted forwarding accepted")
	}
}

func TestBOSourceTrustedProxyChain(t *testing.T) {
	b := &Boundary{TrustedProxies: []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")}}
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.4:1"
	r.Header.Set("X-Forwarded-For", "203.0.113.123, 198.51.100.8, 192.0.2.5")
	got, err := b.Source(r)
	if err != nil || got != "198.51.100.8" {
		t.Fatal("proxy trust boundary crossed")
	}
	r.Header.Set("X-Forwarded-For", "invalid")
	if _, err = b.Source(r); err == nil {
		t.Fatal("malformed proxy input accepted")
	}
}
