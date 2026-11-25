package staff

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBOPasswordBoundaries(t *testing.T) {
	p := NewPasswords()
	ctx := context.Background()
	for _, c := range []struct {
		password string
		valid    bool
	}{
		{strings.Repeat("a", 14), false}, {strings.Repeat("a", 15), true},
		{strings.Repeat("a", 128), true}, {strings.Repeat("a", 129), false},
		{strings.Repeat("😀", 128), true}, {strings.Repeat("😀", 129), false},
		{string([]byte{0xff}) + strings.Repeat("a", 15), false},
	} {
		hash, err := p.Hash(ctx, c.password)
		if c.valid {
			if err != nil {
				t.Fatal("valid password rejected")
			}
			ok, err := p.Verify(ctx, hash, c.password)
			if !ok || err != nil {
				t.Fatal("roundtrip failed")
			}
		} else if !errors.Is(err, ErrValidation) {
			t.Fatal("invalid password accepted")
		}
	}
	password := "  fixture password with spaces  "
	a, err := p.Hash(ctx, password)
	if err != nil {
		t.Fatal(err)
	}
	b, err := p.Hash(ctx, password)
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("salts repeated")
	}
	for _, candidate := range []string{strings.TrimSpace(password), "wrong fixture password"} {
		ok, err := p.Verify(ctx, a, candidate)
		if err != nil || ok {
			t.Fatal("password normalization/wrong password accepted")
		}
	}
}

func TestBOPasswordMalformedAndBounds(t *testing.T) {
	p := NewPasswords()
	for _, hash := range []string{"", strings.Repeat("x", 10000), "$argon2i$v=19$m=65536,t=3,p=1$AAAA$AAAA", strings.Replace(DummyPasswordHash, "m=65536", "m=4294967295", 1), strings.Replace(DummyPasswordHash, "t=3", "t=9999999", 1), strings.Replace(DummyPasswordHash, "p=1", "p=255", 1), strings.Replace(DummyPasswordHash, "v=19", "v=16", 1), DummyPasswordHash + "=", strings.Replace(DummyPasswordHash, "m=65536", "m=065536", 1)} {
		if ok, err := p.Verify(context.Background(), hash, "fixture password 123"); err == nil || ok {
			t.Fatal("malformed/unbounded PHC accepted")
		}
	}
	if err := p.VerifyDummy(context.Background(), "fixture password 123"); err != nil {
		t.Fatal("dummy work failed")
	}
}

func TestBOPasswordResourceBounds(t *testing.T) {
	// No KDF mock: occupy the actual process gate to test queue/cancellation.
	processKDFGate <- struct{}{}
	processKDFGate <- struct{}{}
	defer func() { <-processKDFGate; <-processKDFGate }()
	p := NewPasswords()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Hash(ctx, "fixture password 123"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation ignored")
	}
	start := time.Now()
	if _, err := p.Hash(context.Background(), "fixture password 123"); !errors.Is(err, ErrOverloaded) {
		t.Fatal("gate did not bound waiting")
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("acquisition deadline exceeded")
	}
}

func TestBOTokenEncoding(t *testing.T) {
	a, hash, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	b, other, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(a)
	if err != nil || len(decoded) != 32 || len(a) != 43 {
		t.Fatal("wrong token encoding")
	}
	if a == b || EqualTokenHashes(hash, other) || !EqualTokenHashes(hash, hash) {
		t.Fatal("token uniqueness/comparison")
	}
	got, err := HashSessionToken(a)
	if err != nil || !EqualTokenHashes(hash, got) {
		t.Fatal("token hash mismatch")
	}
	for _, raw := range []string{a + "=", a[:42], strings.Repeat("a", 44), " " + a, strings.Repeat("!", 43)} {
		if _, err := HashSessionToken(raw); err == nil {
			t.Fatal("noncanonical token accepted")
		}
	}
}
