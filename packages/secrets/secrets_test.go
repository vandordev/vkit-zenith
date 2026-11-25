package secrets

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func key(id string, fill byte) Key {
	var secret [32]byte
	for i := range secret {
		secret[i] = fill
	}
	return Key{ID: id, Secret: secret}
}

func TestKeyringValidation(t *testing.T) {
	for name, keys := range map[string][]Key{
		"empty":        nil,
		"empty id":     {key("", 1)},
		"duplicate id": {key("old", 1), key("old", 2)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewKeyring(keys); err == nil {
				t.Fatal("NewKeyring accepted invalid keys")
			}
		})
	}
}

func TestSealOpenAndRotation(t *testing.T) {
	old, _ := NewKeyring([]Key{key("old", 1)})
	ciphertext, id, err := old.Seal([]byte("secret value"), []byte("row aad"))
	if err != nil || id != "old" {
		t.Fatalf("Seal() = key %q, error %v", id, err)
	}
	rotated, err := NewKeyring([]Key{key("new", 2), key("old", 1)})
	if err != nil {
		t.Fatal(err)
	}
	opened, err := rotated.Open(ciphertext, id, []byte("row aad"))
	if err != nil || !bytes.Equal(opened, []byte("secret value")) {
		t.Fatalf("Open() = %q, error %v", opened, err)
	}
	_, id, err = rotated.Seal(nil, nil)
	if err != nil || id != "new" {
		t.Fatalf("Seal() after rotation = key %q, error %v", id, err)
	}
}

func TestOpenRejectsTamperingAndInvalidFormats(t *testing.T) {
	ring, _ := NewKeyring([]Key{key("k1", 1)})
	original, id, err := ring.Seal([]byte("plaintext marker"), []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func([]byte) []byte{
		"wrap nonce":    func(b []byte) []byte { b[1] ^= 1; return b },
		"wrapped key":   func(b []byte) []byte { b[13] ^= 1; return b },
		"data nonce":    func(b []byte) []byte { b[61] ^= 1; return b },
		"data":          func(b []byte) []byte { b[73] ^= 1; return b },
		"wrong version": func(b []byte) []byte { b[0]++; return b },
		"truncated":     func(b []byte) []byte { return b[:len(b)-1] },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			ciphertext := append([]byte(nil), original...)
			if _, err := ring.Open(mutate(ciphertext), id, []byte("aad")); !errors.Is(err, ErrOpen) {
				t.Fatalf("Open() error = %v, want ErrOpen", err)
			}
		})
	}
	for name, params := range map[string]struct {
		id  string
		aad []byte
	}{
		"wrong aad":      {id, []byte("other aad")},
		"unknown key id": {"missing", []byte("aad")},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ring.Open(original, params.id, params.aad); !errors.Is(err, ErrOpen) {
				t.Fatalf("Open() error = %v, want ErrOpen", err)
			}
		})
	}
}

func TestOpenErrorIsOpaque(t *testing.T) {
	ring, _ := NewKeyring([]Key{key("sensitive-key-id", 0xab)})
	if _, err := ring.Open([]byte("ciphertext marker"), "sensitive-key-id", nil); err == nil {
		t.Fatal("Open() accepted invalid ciphertext")
	} else if strings.Contains(err.Error(), "ciphertext marker") || strings.Contains(err.Error(), "sensitive-key-id") || strings.Contains(err.Error(), "abababab") {
		t.Fatalf("Open() error exposed input: %v", err)
	}
}

func TestRecordAADLengthPrefixesFields(t *testing.T) {
	if bytes.Equal(RecordAAD("ab", "c", "d", "e"), RecordAAD("a", "bc", "d", "e")) {
		t.Fatal("RecordAAD() ambiguously encoded adjacent fields")
	}
	if got := RecordAAD("table", "column", "org", "record"); len(got) <= len("tablecolumnorgrecord") {
		t.Fatalf("RecordAAD() omitted length prefixes: %x", got)
	}
}
