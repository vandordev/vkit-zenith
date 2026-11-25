package config

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const infrastructureFixture = `worker:
  shutdownTimeoutSeconds: "${WORKER_SHUTDOWN_TIMEOUT_SECONDS:-30}"
  maxWorkers: "${WORKER_MAX_WORKERS:-10}"
storage:
  endpoint: "${R2_ENDPOINT}"
  region: "${R2_REGION:-auto}"
  bucket: "${R2_BUCKET}"
  accessKeyId: "${R2_ACCESS_KEY_ID}"
  secretAccessKey: "${R2_SECRET_ACCESS_KEY}"
  presignTtlSeconds: "${R2_PRESIGN_TTL_SECONDS:-300}"
encryption:
  keys: ["${ZENITH_KEK_ACTIVE}"]
`

func infraFixture(t *testing.T, transform func(string) string) Options {
	t.Helper()
	o := fixture(t, nil)
	path := filepath.Join(o.Root, "config/config.yaml")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	if !strings.Contains(text, "worker:\n") {
		text += infrastructureFixture
	}
	if transform != nil {
		text = transform(text)
	}
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	return o
}

func infraEnvironment() map[string]string {
	return map[string]string{
		"WORKER_SHUTDOWN_TIMEOUT_SECONDS": "30", "WORKER_MAX_WORKERS": "10",
		"R2_ENDPOINT": "https://acct.r2.cloudflarestorage.com", "R2_REGION": "auto",
		"R2_BUCKET": "private-bucket", "R2_ACCESS_KEY_ID": "fake-access-key",
		"R2_SECRET_ACCESS_KEY": "fake-secret-key", "R2_PRESIGN_TTL_SECONDS": "300",
		"ZENITH_KEK_ACTIVE": "active:" + base64.StdEncoding.EncodeToString(bytesOf(7)),
	}
}

func TestInfrastructureSelectorIsolation(t *testing.T) {
	o := infraFixture(t, nil)
	for name, tc := range map[string]struct {
		environment map[string]string
		load        func(Options) error
	}{
		"worker": {map[string]string{"WORKER_SHUTDOWN_TIMEOUT_SECONDS": "30", "WORKER_MAX_WORKERS": "10"}, func(o Options) error { _, e := LoadWorker(o); return e }},
		"storage": {map[string]string{
			"R2_ENDPOINT": "https://acct.r2.cloudflarestorage.com", "R2_REGION": "auto",
			"R2_BUCKET": "private-bucket", "R2_ACCESS_KEY_ID": "fake-access-key",
			"R2_SECRET_ACCESS_KEY": "fake-secret-key", "R2_PRESIGN_TTL_SECONDS": "300",
		}, func(o Options) error { _, e := LoadStorage(o); return e }},
		"encryption": {map[string]string{"ZENITH_KEK_ACTIVE": "active:" + base64.StdEncoding.EncodeToString(bytesOf(7))}, func(o Options) error { _, e := LoadEncryption(o); return e }},
	} {
		t.Run(name+"_ignores_unselected_secrets", func(t *testing.T) {
			if err := tc.load(Options{Root: o.Root, Environment: tc.environment}); err != nil {
				t.Fatalf("selector resolved unselected fields: %v", err)
			}
		})
	}
	// Existing selectors run with their own secrets only; infra templates stay unresolved.
	for name, load := range map[string]func(Options) error{
		"api": func(o Options) error {
			_, e := LoadAPI(Options{Root: o.Root, Environment: map[string]string{}})
			return e
		},
		"web-api": func(o Options) error {
			_, e := LoadWebAPI(Options{Root: o.Root, Environment: map[string]string{}})
			return e
		},
		"database": func(o Options) error {
			_, e := LoadDatabase(Options{Root: o.Root, Environment: map[string]string{
				"DB_DSN": "postgres://fixture@localhost/fixture",
			}})
			return e
		},
		"identity": func(o Options) error {
			_, e := LoadIdentity(Options{Root: o.Root, Environment: identityEnvironment()})
			return e
		},
	} {
		t.Run(name+"_ignores_infrastructure_secrets", func(t *testing.T) {
			if err := load(Options{Root: o.Root}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func identityEnvironment() map[string]string {
	return map[string]string{
		"IDENTITY_AUTH_ORIGIN": "https://auth.example.com",
		"IDENTITY_EMAIL_FROM":  "identity@example.com", "IDENTITY_RESEND_KEY": "fake-resend-key",
		"IDENTITY_CSRF_KEY":    base64.StdEncoding.EncodeToString(bytesOf(1)),
		"IDENTITY_LIMITER_KEY": base64.StdEncoding.EncodeToString(bytesOf(2)),
	}
}

func TestInfrastructureBoundsAndSafeDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name, key, value string
		load             func(Options) error
	}{
		{"worker-low", "WORKER_MAX_WORKERS", "0", func(o Options) error { _, e := LoadWorker(o); return e }},
		{"worker-max-workers-high", "WORKER_MAX_WORKERS", "1001", func(o Options) error { _, e := LoadWorker(o); return e }},
		{"worker-shutdown-low", "WORKER_SHUTDOWN_TIMEOUT_SECONDS", "0", func(o Options) error { _, e := LoadWorker(o); return e }},
		{"worker-high", "WORKER_SHUTDOWN_TIMEOUT_SECONDS", "301", func(o Options) error { _, e := LoadWorker(o); return e }},
		{"storage-ttl-low", "R2_PRESIGN_TTL_SECONDS", "59", func(o Options) error { _, e := LoadStorage(o); return e }},
		{"storage-ttl-high", "R2_PRESIGN_TTL_SECONDS", "3601", func(o Options) error { _, e := LoadStorage(o); return e }},
		{"storage-userinfo", "R2_ENDPOINT", "https://fake-secret@acct.r2.cloudflarestorage.com", func(o Options) error { _, e := LoadStorage(o); return e }},
		{"storage-path", "R2_ENDPOINT", "https://acct.r2.cloudflarestorage.com/path", func(o Options) error { _, e := LoadStorage(o); return e }},
		{"storage-empty-fragment", "R2_ENDPOINT", "https://acct.r2.cloudflarestorage.com/#", func(o Options) error { _, e := LoadStorage(o); return e }},
		{"storage-fragment", "R2_ENDPOINT", "https://acct.r2.cloudflarestorage.com/#fragment", func(o Options) error { _, e := LoadStorage(o); return e }},
		{"storage-empty-host", "R2_ENDPOINT", "https:///", func(o Options) error { _, e := LoadStorage(o); return e }},
		{"storage-empty-query", "R2_ENDPOINT", "https://acct.r2.cloudflarestorage.com/?", func(o Options) error { _, e := LoadStorage(o); return e }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := infraFixture(t, nil)
			o.Environment = infraEnvironment()
			o.Environment[tc.key] = tc.value
			err := tc.load(o)
			if err == nil {
				t.Fatal("invalid config accepted")
			}
			if strings.Contains(err.Error(), "fake-secret") || strings.Contains(err.Error(), "fake-access-key") {
				t.Fatalf("secret leaked: %v", err)
			}
		})
	}
}

func TestStorageLoopbackRequiresDevelopmentOrTest(t *testing.T) {
	for _, environment := range []string{"development", "test"} {
		o := infraFixture(t, func(s string) string {
			return strings.Replace(s, "environment: \"${NODE_ENV:-development}\"", "environment: "+environment, 1)
		})
		o.Environment = infraEnvironment()
		o.Environment["R2_ENDPOINT"] = "http://127.0.0.1:9000"
		if _, err := LoadStorage(o); err != nil {
			t.Fatalf("%s loopback rejected: %v", environment, err)
		}
	}
	o := infraFixture(t, func(s string) string {
		return strings.Replace(s, "environment: \"${NODE_ENV:-development}\"", "environment: production", 1)
	})
	o.Environment = infraEnvironment()
	o.Environment["R2_ENDPOINT"] = "http://127.0.0.1:9000"
	if _, err := LoadStorage(o); err == nil {
		t.Fatal("production loopback accepted")
	}
}

func TestEncryptionKeySequenceValidation(t *testing.T) {
	for _, tc := range []struct {
		name, keys string
		valid      bool
	}{
		{"valid rotation", `keys: ["old:` + base64.StdEncoding.EncodeToString(bytesOf(1)) + `", "active:` + base64.StdEncoding.EncodeToString(bytesOf(2)) + `"]`, true},
		{"duplicate id", `keys: ["same:` + base64.StdEncoding.EncodeToString(bytesOf(1)) + `", "same:` + base64.StdEncoding.EncodeToString(bytesOf(2)) + `"]`, false},
		{"bad id", `keys: ["Bad:` + base64.StdEncoding.EncodeToString(bytesOf(1)) + `"]`, false},
		{"wrong length", `keys: ["active:` + base64.StdEncoding.EncodeToString([]byte("short")) + `"]`, false},
		{"zero key", `keys: ["active:` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `"]`, false},
		{"malformed base64", `keys: ["active:not-base64"]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := infraFixture(t, func(s string) string { return strings.Replace(s, `keys: ["${ZENITH_KEK_ACTIVE}"]`, tc.keys, 1) })
			got, err := LoadEncryption(o)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v selected=%v err=%v", tc.valid, got.ActiveKeyID != "", err)
			}
			if tc.name == "valid rotation" {
				first, second := bytesOf(1), bytesOf(2)
				if got.ActiveKeyID != "old" || len(got.Keys) != 2 || got.Keys[0].ID != "old" || got.Keys[0].Secret != [32]byte(first) || got.Keys[1].ID != "active" || got.Keys[1].Secret != [32]byte(second) {
					t.Fatal("key order, decoded bytes, or active key incorrect")
				}
			}
		})
	}
}
