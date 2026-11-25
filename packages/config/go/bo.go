package config

import (
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"
)

type BOAPIConfig struct {
	App                 AppConfig
	BOAPI               ServiceConfig
	BOOrigin, APIOrigin string
	Development         bool
	TrustedProxyCIDRs   []netip.Prefix
	CSRFSecret          string
}

// LoadBOAPI selects no database or other service's credentials.
func LoadBOAPI(options Options) (BOAPIConfig, error) {
	d, err := loadDocument(options, "boApi")
	if err != nil {
		return BOAPIConfig{}, err
	}
	var c BOAPIConfig
	if c.App, c.BOAPI, err = d.selected("boApi"); err != nil {
		return BOAPIConfig{}, err
	}
	for _, f := range []struct {
		key string
		out *string
	}{{"boOrigin", &c.BOOrigin}, {"apiOrigin", &c.APIOrigin}, {"csrfSecret", &c.CSRFSecret}} {
		if *f.out, err = d.stringAt("boApi", f.key); err != nil {
			return BOAPIConfig{}, err
		}
	}
	mode, err := d.valueAt("boApi", "development")
	if err != nil {
		return BOAPIConfig{}, err
	}
	switch v := mode.(type) {
	case bool:
		c.Development = v
	case string:
		if v != "true" && v != "false" {
			return BOAPIConfig{}, d.fail("boApi.development", "expected boolean")
		}
		c.Development = v == "true"
	default:
		return BOAPIConfig{}, d.fail("boApi.development", "expected boolean")
	}
	if len(c.CSRFSecret) < 32 || len(c.CSRFSecret) > 4096 {
		return BOAPIConfig{}, d.fail("boApi.csrfSecret", "expected 32–4096 bytes")
	}
	origins := make([]*url.URL, 0, 2)
	for _, value := range []string{c.BOOrigin, c.APIOrigin} {
		u, e := url.Parse(value)
		if e != nil || !validOrigin(value) || u.Path != "" || u.RawPath != "" || u.Opaque != "" || u.Hostname() == "" || strings.ToLower(u.Host) != u.Host {
			return BOAPIConfig{}, d.fail("boApi.origins", "expected canonical origins without path")
		}
		if port := u.Port(); port != "" {
			n, e := strconv.Atoi(port)
			if e != nil || n < 1 || n > 65535 {
				return BOAPIConfig{}, d.fail("boApi.origins", "invalid port")
			}
		}
		origins = append(origins, u)
	}
	if c.Development {
		if c.App.Environment != "development" && c.App.Environment != "test" {
			return BOAPIConfig{}, d.fail("boApi.development", "not allowed in staging/production")
		}
		a, b := origins[0], origins[1]
		if a.Scheme != "http" || b.Scheme != "http" || a.Hostname() != b.Hostname() || !loopbackHost(a.Hostname()) {
			return BOAPIConfig{}, d.fail("boApi.origins", "development requires matching loopback HTTP hosts")
		}
	} else {
		sites := make([]string, 0, 2)
		for _, u := range origins {
			site, e := publicsuffix.EffectiveTLDPlusOne(u.Hostname())
			if u.Scheme != "https" || e != nil || net.ParseIP(u.Hostname()) != nil {
				return BOAPIConfig{}, d.fail("boApi.origins", "production requires same-site HTTPS DNS origins")
			}
			sites = append(sites, site)
		}
		if sites[0] != sites[1] {
			return BOAPIConfig{}, d.fail("boApi.origins", "cross-site origins forbidden")
		}
	}
	raw := d.root["boApi"].(map[string]any)["trustedProxyCIDRs"].([]any)
	for i := range raw {
		s, e := d.stringAt("boApi", "trustedProxyCIDRs", strconv.Itoa(i))
		if e != nil {
			return BOAPIConfig{}, e
		}
		p, e := netip.ParsePrefix(s)
		if e != nil || p != p.Masked() {
			return BOAPIConfig{}, d.fail("boApi.trustedProxyCIDRs", "expected canonical CIDR")
		}
		c.TrustedProxyCIDRs = append(c.TrustedProxyCIDRs, p)
	}
	return c, nil
}

func loopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	a, e := netip.ParseAddr(host)
	return e == nil && a.IsLoopback()
}
