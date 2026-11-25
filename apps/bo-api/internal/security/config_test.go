package security

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
)

func TestBOOriginConfigAndDevelopmentCookie(t *testing.T) {
	source, err := os.ReadFile("../../../../config/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.MkdirTemp("/tmp/opencode", "bo-security-config-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})
	if err = os.Mkdir(filepath.Join(root, "config"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "config/config.yaml"), source, 0600); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name, env, dev, bo, api string
		ok                      bool
	}{{"same_site", "production", "false", "https://bo.example.test", "https://bo-api.example.test", true}, {"http", "production", "false", "http://bo.example.test", "http://bo-api.example.test", false}, {"cross_site", "production", "false", "https://bo.example.test", "https://other.example.org", false}, {"prod_fallback", "production", "true", "http://localhost:4103", "http://localhost:4104", false}, {"loopback", "development", "true", "http://localhost:4103", "http://localhost:4104", true}, {"non_loopback", "development", "true", "http://bo.example.test", "http://bo.example.test:4104", false}, {"mismatch_host", "development", "true", "http://localhost:4103", "http://127.0.0.1:4104", false}} {
		t.Run(c.name, func(t *testing.T) {
			got, e := config.LoadBOAPI(config.Options{Root: root, Environment: map[string]string{"NODE_ENV": c.env, "BO_API_DEVELOPMENT": c.dev, "BO_ORIGIN": c.bo, "BO_API_BASE_URL": c.api, "BO_API_CSRF_SECRET": "fixture-only-long-enough-secret-for-tests", "BO_API_DOCS_ENABLED": "false"}})
			if (e == nil) != c.ok {
				t.Fatal("topology validation mismatch")
			}
			if e == nil {
				b := &Boundary{Development: got.Development}
				cookie := b.cookie(b.SessionName(), "opaque", time.Unix(100000, 0), 2592000)
				if got.Development && cookie.Secure {
					t.Fatal("development transport not explicit")
				}
				if !got.Development && (!cookie.Secure || cookie.Name != "__Host-zenith_bo_session") {
					t.Fatal("unsafe production cookie")
				}
			}
		})
	}
}
