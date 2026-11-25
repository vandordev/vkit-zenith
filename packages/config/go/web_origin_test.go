package config

import "testing"

func TestLoadWebOriginSelectsOnlyWebOrigin(t *testing.T) {
	o := fixture(t, nil)
	o.Environment["IDENTITY_AUTH_ORIGIN"] = "https://web.example.com"
	got, err := LoadWebOrigin(o)
	if err != nil {
		t.Fatal(err)
	}
	if got != WebOrigin("https://web.example.com") {
		t.Fatalf("web origin = %q", got)
	}
}

func TestLoadWebOriginRejectsUntrustedOrigin(t *testing.T) {
	o := fixture(t, nil)
	o.Environment["IDENTITY_AUTH_ORIGIN"] = "https://web.example.com/path"
	if _, err := LoadWebOrigin(o); err == nil {
		t.Fatal("Web origin with a path must be rejected")
	}
	o = fixture(t, nil)
	o.Environment["NODE_ENV"] = "production"
	o.Environment["IDENTITY_AUTH_ORIGIN"] = "http://web.example.com"
	if _, err := LoadWebOrigin(o); err == nil {
		t.Fatal("production must reject an insecure Web origin")
	}
}
