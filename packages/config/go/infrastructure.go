package config

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type WorkerConfig struct{ ShutdownTimeoutSeconds, MaxWorkers int }
type StorageConfig struct {
	Endpoint, Region, Bucket, AccessKeyID, SecretAccessKey string
	PresignTTLSeconds                                      int
}
type EncryptionKey struct {
	ID     string
	Secret [32]byte
}
type EncryptionConfig struct {
	Keys        []EncryptionKey
	ActiveKeyID string
}

var encryptionKeyID = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)

func (d *document) boundedInt(path string, min, max int) (int, error) {
	value, err := d.valueAt(strings.Split(path, ".")...)
	if err != nil {
		return 0, err
	}
	var n int
	switch v := value.(type) {
	case string:
		if !digits.MatchString(v) {
			return 0, d.fail(path, "expected bounded integer")
		}
		parsed, e := strconv.Atoi(v)
		if e != nil {
			return 0, d.fail(path, "expected bounded integer")
		}
		n = parsed
	case float64:
		if v != float64(int(v)) {
			return 0, d.fail(path, "expected bounded integer")
		}
		n = int(v)
	default:
		return 0, d.fail(path, "expected bounded integer")
	}
	if n < min || n > max {
		return 0, d.fail(path, "expected bounded integer")
	}
	return n, nil
}

func LoadWorker(options Options) (WorkerConfig, error) {
	d, err := loadDocument(options, "worker")
	if err != nil {
		return WorkerConfig{}, err
	}
	var out WorkerConfig
	if out.ShutdownTimeoutSeconds, err = d.boundedInt("worker.shutdownTimeoutSeconds", 1, 300); err != nil {
		return WorkerConfig{}, err
	}
	if out.MaxWorkers, err = d.boundedInt("worker.maxWorkers", 1, 1000); err != nil {
		return WorkerConfig{}, err
	}
	return out, nil
}

func LoadStorage(options Options) (StorageConfig, error) {
	d, err := loadDocument(options, "storage")
	if err != nil {
		return StorageConfig{}, err
	}
	var out StorageConfig
	for _, field := range []struct {
		key    string
		target *string
	}{
		{"endpoint", &out.Endpoint}, {"region", &out.Region}, {"bucket", &out.Bucket}, {"accessKeyId", &out.AccessKeyID}, {"secretAccessKey", &out.SecretAccessKey},
	} {
		if *field.target, err = d.stringAt("storage", field.key); err != nil {
			return StorageConfig{}, err
		}
	}
	if out.PresignTTLSeconds, err = d.boundedInt("storage.presignTtlSeconds", 60, 3600); err != nil {
		return StorageConfig{}, err
	}
	u, e := url.Parse(out.Endpoint)
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(out.Endpoint, "#") || (u.Path != "" && u.Path != "/") || u.Opaque != "" || strings.ContainsAny(out.Endpoint, " \t\r\n") {
		return StorageConfig{}, d.fail("storage.endpoint", "expected absolute host-only URL")
	}
	if u.Scheme != "https" {
		if u.Scheme != "http" || !loopbackHost(u.Hostname()) {
			return StorageConfig{}, d.fail("storage.endpoint", "expected HTTPS endpoint or loopback HTTP")
		}
		appEnv, e := d.stringAt("app", "environment")
		if e != nil {
			return StorageConfig{}, e
		}
		if appEnv != "development" && appEnv != "test" {
			return StorageConfig{}, d.fail("storage.endpoint", "loopback HTTP requires development or test")
		}
	}
	if u.Hostname() == "" {
		return StorageConfig{}, d.fail("storage.endpoint", "expected endpoint host")
	}
	if out.Region == "" || out.Bucket == "" || out.AccessKeyID == "" || out.SecretAccessKey == "" {
		return StorageConfig{}, d.fail("storage", "required field is empty")
	}
	return out, nil
}

func LoadEncryption(options Options) (EncryptionConfig, error) {
	d, err := loadDocument(options, "encryption")
	if err != nil {
		return EncryptionConfig{}, err
	}
	raw, err := d.rawAuth("encryption", "keys")
	if err != nil {
		return EncryptionConfig{}, err
	}
	entries, ok := raw.([]any)
	if !ok || len(entries) == 0 {
		return EncryptionConfig{}, d.fail("encryption.keys", "expected nonempty key sequence")
	}
	out := EncryptionConfig{Keys: make([]EncryptionKey, 0, len(entries))}
	seen := make(map[string]bool, len(entries))
	for i := range entries {
		entry, e := d.stringAt("encryption", "keys", strconv.Itoa(i))
		if e != nil {
			return EncryptionConfig{}, e
		}
		id, encoded, found := strings.Cut(entry, ":")
		if !found || !encryptionKeyID.MatchString(id) || seen[id] {
			return EncryptionConfig{}, d.fail(fmt.Sprintf("encryption.keys.%d", i), "invalid or duplicate key id")
		}
		secret, e := base64.StdEncoding.Strict().DecodeString(encoded)
		if e != nil || len(secret) != 32 || base64.StdEncoding.EncodeToString(secret) != encoded {
			return EncryptionConfig{}, d.fail(fmt.Sprintf("encryption.keys.%d", i), "expected canonical base64 of 32 bytes")
		}
		var key EncryptionKey
		key.ID = id
		copy(key.Secret[:], secret)
		if key.Secret == ([32]byte{}) {
			return EncryptionConfig{}, d.fail(fmt.Sprintf("encryption.keys.%d", i), "key must not be all zero")
		}
		seen[id] = true
		out.Keys = append(out.Keys, key)
	}
	out.ActiveKeyID = out.Keys[0].ID
	return out, nil
}
