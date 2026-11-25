package storage

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func settings() Settings {
	return Settings{Endpoint: "https://account.r2.cloudflarestorage.com", Region: "auto", Bucket: "private", AccessKeyID: "access", SecretAccessKey: "secret", PresignTTL: 5 * time.Minute}
}

func TestNewValidatesSettings(t *testing.T) {
	for _, endpoint := range []string{"http://example.com", "https://user:pass@example.com", "https://example.com/path", "https://example.com?x=1", "https://example.com?", "https://example.com#fragment", "https://example.com#", "https://", "https://:443"} {
		t.Run(endpoint, func(t *testing.T) {
			s := settings()
			s.Endpoint = endpoint
			if _, err := New(s); err == nil {
				t.Fatal("expected endpoint error")
			}
		})
	}
	for _, endpoint := range []string{"http://localhost", "http://localhost:9000", "http://127.0.0.1:9000", "http://[::1]:9000"} {
		t.Run(endpoint, func(t *testing.T) {
			s := settings()
			s.Endpoint = endpoint
			if _, err := New(s); err != nil {
				t.Fatalf("loopback endpoint rejected: %v", err)
			}
		})
	}
	for _, ttl := range []time.Duration{59 * time.Second, 3601 * time.Second} {
		t.Run(ttl.String(), func(t *testing.T) {
			s := settings()
			s.PresignTTL = ttl
			if _, err := New(s); err == nil {
				t.Fatal("expected ttl error")
			}
		})
	}
	if _, err := New(settings()); err != nil {
		t.Fatal(err)
	}
}

func TestKeyBuilders(t *testing.T) {
	org := uuid.New()
	for _, tc := range []struct {
		name, key, prefix, ext string
	}{
		{"logo png", mustLogo(org, "image/png"), "orgs/" + org.String() + "/logo/", "png"}, {"logo jpeg", mustLogo(org, "image/jpeg"), "orgs/" + org.String() + "/logo/", "jpg"}, {"logo webp", mustLogo(org, "image/webp"), "orgs/" + org.String() + "/logo/", "webp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !keyPattern.MatchString(tc.key) || !strings.HasPrefix(tc.key, tc.prefix) {
				t.Fatalf("invalid key %q", tc.key)
			}
			filename := strings.TrimPrefix(tc.key, tc.prefix)
			stem := strings.TrimSuffix(filename, "."+tc.ext)
			if stem == filename {
				t.Fatalf("expected .%s extension in %q", tc.ext, filename)
			}
			if _, err := uuid.Parse(stem); err != nil {
				t.Fatalf("expected UUID filename stem in %q: %v", filename, err)
			}
		})
	}
	for _, fn := range []func() error{func() error { _, e := LogoKey(org, "image/gif"); return e }, func() error { _, e := LogoKey(org, "application/pdf"); return e }} {
		if fn() == nil {
			t.Fatal("unsupported type accepted")
		}
	}
}

func TestPresignOffline(t *testing.T) {
	c, err := New(settings())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	put, headers, err := c.PresignPut(ctx, "orgs/a/logo/a.png", "image/png")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(put)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "account.r2.cloudflarestorage.com" || u.Path != "/private/orgs/a/logo/a.png" || u.Query().Get("X-Amz-Expires") != "300" {
		t.Fatalf("unexpected signed URL: %s", put)
	}
	if headers.Get("Content-Type") != "image/png" || !strings.Contains(u.Query().Get("X-Amz-SignedHeaders"), "content-type") {
		t.Fatalf("content type not signed: %#v %s", headers, u.Query().Get("X-Amz-SignedHeaders"))
	}
	get, err := c.PresignGet(ctx, "orgs/a/logo/a.png")
	if err != nil {
		t.Fatal(err)
	}
	gu, err := url.Parse(get)
	if err != nil {
		t.Fatal(err)
	}
	if gu.Host != u.Host || gu.Path != u.Path || gu.Query().Get("X-Amz-Expires") != "300" {
		t.Fatalf("unexpected get URL: %s", get)
	}
	if _, _, err := c.PresignPut(ctx, "bad key", "image/png"); err == nil {
		t.Fatal("invalid key accepted")
	}
	if _, _, err := c.PresignPut(ctx, "orgs/a/logo/a.png", "image/gif"); err == nil {
		t.Fatal("invalid logo type accepted")
	}
	logo, headers, err := c.PresignPut(ctx, "orgs/a/logo/a.png", "image/png")
	if err != nil {
		t.Fatal(err)
	}
	lu, err := url.Parse(logo)
	if err != nil || !strings.Contains(lu.Query().Get("X-Amz-SignedHeaders"), "content-type") || headers.Get("Content-Type") != "image/png" {
		t.Fatalf("logo content type not signed: %s %#v", logo, headers)
	}
	if _, _, err := c.PresignPut(ctx, "orgs/a/other/a.png", "image/png"); err == nil {
		t.Fatal("image accepted outside logo path")
	}
}

func mustLogo(id uuid.UUID, ct string) string { k, _ := LogoKey(id, ct); return k }
