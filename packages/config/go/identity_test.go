package config

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func authFixture(t *testing.T) Options {
	t.Helper()
	o := fixture(t, nil)
	for k, v := range map[string]string{
		"IDENTITY_AUTH_ORIGIN": "https://auth.example.com",
		"IDENTITY_EMAIL_FROM":  "identity@example.com", "IDENTITY_RESEND_KEY": "fake-resend-sensitive",
		"IDENTITY_CSRF_KEY": base64.StdEncoding.EncodeToString(bytesOf(1)), "IDENTITY_LIMITER_KEY": base64.StdEncoding.EncodeToString(bytesOf(2)),
	} {
		o.Environment[k] = v
	}
	return o
}
func bytesOf(v byte) []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = v
	}
	return b
}

func TestAuthConfigSelectors(t *testing.T) {
	o := authFixture(t)
	core, err := LoadIdentity(o)
	if err != nil {
		t.Fatal(err)
	}
	if core.Identity.AuthOrigin != "https://auth.example.com" {
		t.Fatal("wrong identity selection")
	}
	if _, err := LoadDatabase(o); err == nil {
		t.Fatal("Core DB DSN was invented")
	}
}

func TestAuthConfigRejectsUnsafeValues(t *testing.T) {
	for _, tc := range []struct {
		name, key, value string
	}{
		{"origin-path", "IDENTITY_AUTH_ORIGIN", "https://auth.example.com/dynamic"},
		{"forwarded-host", "IDENTITY_AUTH_ORIGIN", "https://auth.example.com@attacker.test"},
		{"publicsuffix", "IDENTITY_AUTH_ORIGIN", "https://com"},
		{"bad-key", "IDENTITY_CSRF_KEY", "fake-sensitive-invalid-base64"},
		{"zero-key", "IDENTITY_CSRF_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := authFixture(t)
			o.Environment[tc.key] = tc.value
			var err error
			_, err = LoadIdentity(o)
			if err == nil {
				t.Fatal("unsafe value accepted")
			}
			if strings.Contains(err.Error(), "fake-sensitive") {
				t.Fatal("secret diagnostic leak")
			}
		})
	}
}

func TestAuthConfigKeySeparation(t *testing.T) {
	o := authFixture(t)
	o.Environment["IDENTITY_LIMITER_KEY"] = o.Environment["IDENTITY_CSRF_KEY"]
	if _, err := LoadIdentity(o); err == nil {
		t.Fatal("CSRF/limiter key reuse accepted")
	}
}

func TestAuthConfigUnselectedGrammar(t *testing.T) {
	o := authFixture(t)
	file := filepath.Join(o.Root, "config/config.yaml")
	source, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(string(source), "csrfKey:", "unknownSensitiveKey:", 1)
	if err := os.WriteFile(file, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAPI(o); err == nil {
		t.Fatal("malformed unselected identity structure accepted")
	}
}

func TestAuthConfigUnselectedTemplate(t *testing.T) {
	o := fixture(t, func(s string) string { return strings.Replace(s, "${IDENTITY_RESEND_KEY}", "${broken", 1) })
	if _, err := LoadAPI(o); err == nil {
		t.Fatal("unselected malformed template accepted")
	}
}
