package config

import (
	"encoding/base64"
	"net"
	"net/mail"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"
)

type IdentitySettings struct {
	AuthOrigin          string
	Development         bool
	TrustedProxyCIDRs   []netip.Prefix
	CSRFKey, LimiterKey [32]byte
	Email               struct{ From, ResendKey string }
}
type IdentityConfig struct {
	App      AppConfig
	API      ServiceConfig
	Identity IdentitySettings
}

// LoadIdentity does not select any database DSN. The Core
// database remains an explicit, separate LoadDatabase selection.
func LoadIdentity(options Options) (IdentityConfig, error) {
	d, err := loadDocument(options, "identity")
	if err != nil {
		return IdentityConfig{}, err
	}
	var c IdentityConfig
	if c.App, c.API, err = d.selectedService("webApi"); err != nil {
		return IdentityConfig{}, err
	}
	p := []string{"webApi", "identity"}
	get := func(key string) (string, error) { return d.stringAt(append(p, key)...) }
	if c.Identity.AuthOrigin, err = get("authOrigin"); err != nil {
		return IdentityConfig{}, err
	}
	if c.Identity.Development, err = d.authBool(append(p, "development")...); err != nil {
		return IdentityConfig{}, err
	}
	if err = d.authOrigins(c.App, c.Identity.Development, c.Identity.AuthOrigin); err != nil {
		return IdentityConfig{}, err
	}
	if c.Identity.TrustedProxyCIDRs, err = d.authProxies(append(p, "trustedProxyCIDRs")...); err != nil {
		return IdentityConfig{}, err
	}
	if c.Identity.CSRFKey, err = d.authKey(append(p, "csrfKey")...); err != nil {
		return IdentityConfig{}, err
	}
	if c.Identity.LimiterKey, err = d.authKey(append(p, "limiterKey")...); err != nil {
		return IdentityConfig{}, err
	}

	if c.Identity.CSRFKey == c.Identity.LimiterKey {
		return IdentityConfig{}, d.fail("webApi.identity.keys", "key purposes must be distinct")
	}
	if c.Identity.Email.From, err = d.stringAt("webApi", "identity", "email", "from"); err != nil {
		return IdentityConfig{}, err
	}
	if c.Identity.Email.ResendKey, err = d.stringAt("webApi", "identity", "email", "resendKey"); err != nil {
		return IdentityConfig{}, err
	}
	address, e := mail.ParseAddress(c.Identity.Email.From)
	if e != nil || address.Address == "" || len(c.Identity.Email.From) > 320 || strings.ContainsAny(c.Identity.Email.From, "\r\n") || len(c.Identity.Email.ResendKey) < 8 || len(c.Identity.Email.ResendKey) > 4096 {
		return IdentityConfig{}, d.fail("webApi.identity.email", "invalid sender or missing provider credential")
	}
	return c, nil
}
func (d *document) authBool(path ...string) (bool, error) {
	v, err := d.valueAt(path...)
	if err != nil {
		return false, err
	}
	switch b := v.(type) {
	case bool:
		return b, nil
	case string:
		if b == "true" || b == "false" {
			return b == "true", nil
		}
	}
	return false, d.fail(strings.Join(path, "."), "expected boolean")
}
func (d *document) authKey(path ...string) ([32]byte, error) {
	var out [32]byte
	s, err := d.stringAt(path...)
	if err != nil {
		return out, err
	}
	if len(s) != 44 {
		return out, d.fail(strings.Join(path, "."), "expected canonical base64 of 32 bytes")
	}
	b, e := base64.StdEncoding.Strict().DecodeString(s)
	if e != nil || len(b) != 32 || base64.StdEncoding.EncodeToString(b) != s {
		return out, d.fail(strings.Join(path, "."), "expected canonical base64 of 32 bytes")
	}
	copy(out[:], b)
	if out == ([32]byte{}) {
		return out, d.fail(strings.Join(path, "."), "key must not be all zero")
	}
	return out, nil
}

// rawAuth navigates already globally validated structure without interpolation.
func (d *document) rawAuth(path ...string) (any, error) {
	var current any = d.root
	for _, part := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, d.fail("identity", "invalid mapping")
		}
		current = m[part]
	}
	return current, nil
}
func (d *document) authProxies(path ...string) ([]netip.Prefix, error) {
	raw, err := d.rawAuth(path...)
	if err != nil {
		return nil, err
	}
	var out []netip.Prefix
	for i := range raw.([]any) {
		value, err := d.stringAt(append(path, strconv.Itoa(i))...)
		if err != nil {
			return nil, err
		}
		prefix, e := netip.ParsePrefix(value)
		if e != nil || prefix != prefix.Masked() {
			return nil, d.fail(strings.Join(path, "."), "expected canonical proxy CIDR")
		}
		out = append(out, prefix)
	}
	return out, nil
}
func (d *document) authOrigins(app AppConfig, development bool, values ...string) error {
	var site, loopback string
	if development && app.Environment != "development" && app.Environment != "test" {
		return d.fail("identity.development", "loopback exception forbidden in this environment")
	}
	for _, value := range values {
		u, err := url.Parse(value)
		if err != nil || !validOrigin(value) || u.Path != "" || u.RawPath != "" || u.Opaque != "" || u.Hostname() == "" || strings.ToLower(u.Host) != u.Host || strings.HasSuffix(u.Hostname(), ".") || u.ForceQuery {
			return d.fail("identity.origins", "expected canonical static origins")
		}
		if port := u.Port(); port != "" {
			n, e := strconv.Atoi(port)
			if e != nil || n < 1 || n > 65535 || strconv.Itoa(n) != port {
				return d.fail("identity.origins", "invalid port")
			}
		}
		if development {
			if u.Scheme != "http" || !loopbackHost(u.Hostname()) || (loopback != "" && loopback != u.Hostname()) {
				return d.fail("identity.origins", "development requires matching loopback HTTP hosts")
			}
			loopback = u.Hostname()
		} else {
			s, e := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
			if e != nil || net.ParseIP(u.Hostname()) != nil || u.Scheme != "https" || (site != "" && site != s) {
				return d.fail("identity.origins", "requires same-site HTTPS DNS origins")
			}
			site = s
		}
	}
	return nil
}
