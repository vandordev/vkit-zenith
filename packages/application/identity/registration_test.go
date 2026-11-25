package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAuthNormalizeRegistration(t *testing.T) {
	cases := []struct {
		name                          string
		in                            RegisterInput
		valid                         bool
		displayName, email, telephone string
	}{
		{"normalizes address and international phone", RegisterInput{DisplayName: "  Rani  ", Email: " PERSON+tag@Example.TEST ", Telephone: "+62 812-3456-7890", Password: "correct horse battery staple"}, true, "Rani", "person+tag@example.test", "+6281234567890"},
		{"preserves plus-address identity", RegisterInput{DisplayName: "Name", Email: "A+B@EXAMPLE.TEST", Telephone: "+14155552671", Password: "0123456789abcdef"}, true, "Name", "a+b@example.test", "+14155552671"},
		{"rejects missing required telephone", RegisterInput{DisplayName: "Name", Email: "a@example.test", Password: "0123456789abcdef"}, false, "", "", ""},
		{"rejects unqualified local number", RegisterInput{DisplayName: "Name", Email: "a@example.test", Telephone: "08123456789", Password: "0123456789abcdef"}, false, "", "", ""},
		{"rejects invalid phone metadata", RegisterInput{DisplayName: "Name", Email: "a@example.test", Telephone: "+999123456789", Password: "0123456789abcdef"}, false, "", "", ""},
		{"rejects invalid email", RegisterInput{DisplayName: "Name", Email: "a @example.test", Telephone: "+14155552671", Password: "0123456789abcdef"}, false, "", "", ""},
		{"rejects empty display name", RegisterInput{DisplayName: " \t", Email: "a@example.test", Telephone: "+14155552671", Password: "0123456789abcdef"}, false, "", "", ""},
		{"rejects overlong display name", RegisterInput{DisplayName: strings.Repeat("界", 201), Email: "a@example.test", Telephone: "+14155552671", Password: "0123456789abcdef"}, false, "", "", ""},
		{"rejects input phone bytes above bound", RegisterInput{DisplayName: "Name", Email: "a@example.test", Telephone: "+1" + strings.Repeat("1", 64), Password: "0123456789abcdef"}, false, "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeRegistration(tc.in)
			if !tc.valid {
				if err == nil {
					t.Fatal("invalid enrollment accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("valid enrollment rejected")
			}
			if got.Email != tc.email || got.Telephone != tc.telephone || got.DisplayName != tc.displayName {
				t.Fatal("registration normalization mismatch")
			}
			if got.Password != tc.in.Password {
				t.Fatal("password was trimmed or normalized")
			}
		})
	}
}

func TestAuthIdentityPasswordProfileAndUnknownWork(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	password := "  correct horse battery staple  "
	hash, err := HashPassword(ctx, password)
	if err != nil {
		t.Fatal("valid password rejected")
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=3,p=1$") {
		t.Fatal("unsupported credential profile")
	}
	ok, err := verifyIdentityPassword(ctx, hash, password)
	if err != nil || !ok {
		t.Fatal("valid credential did not verify")
	}
	ok, err = verifyIdentityPassword(ctx, hash, "incorrect horse battery staple")
	if err != nil || ok {
		t.Fatal("invalid credential accepted")
	}
	if err = VerifyUnknownPassword(ctx, "incorrect horse battery staple"); err != nil {
		t.Fatal("bounded unknown-account KDF work failed")
	}
	for _, invalid := range []string{"short", strings.Repeat("界", 129), string([]byte{0xff})} {
		if _, err = HashPassword(ctx, invalid); err != ErrValidation {
			t.Fatal("invalid password accepted")
		}
	}
	if _, err = verifyIdentityPassword(ctx, "$argon2id$v=19$m=131072,t=99,p=9$", "correct horse battery staple"); err == nil {
		t.Fatal("caller-selected KDF allocation accepted")
	}
}

func TestAuthAnonymousCSRFFixedExpiryAndTamper(t *testing.T) {
	at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	service := &Service{clock: identityTestClock(func() time.Time { return at }), csrfKey: [32]byte{9}}
	token, err := service.NewAnonymousCSRF()
	if err != nil || len(token) != 96 || !service.ValidateAnonymousCSRF(token) {
		t.Fatal("valid bounded anonymous CSRF token rejected")
	}
	mutated := token[:95] + map[bool]string{true: "A", false: "B"}[token[95] != 'A']
	if service.ValidateAnonymousCSRF(mutated) {
		t.Fatal("tampered CSRF token accepted")
	}
	at = at.Add(10 * time.Minute)
	if service.ValidateAnonymousCSRF(token) {
		t.Fatal("expired CSRF token accepted")
	}
}

func TestAuthSessionCSRFIsBoundToRotatedCookie(t *testing.T) {
	at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	service := &Service{clock: identityTestClock(func() time.Time { return at }), csrfKey: [32]byte{9}, authOrigin: "https://auth.example.test"}
	session := HashSecret("new-browser-session")
	token, err := service.NewSessionCSRF(session)
	if err != nil || !service.ValidateSessionCSRF(session, token) {
		t.Fatal("session-bound CSRF token rejected")
	}
	if service.ValidateAnonymousCSRF(token) || service.ValidateSessionCSRF(HashSecret("other-session"), token) {
		t.Fatal("session-bound CSRF token accepted outside its rotated session")
	}
}

func TestAuthActionRedactionCannotSerializeBearerMaterial(t *testing.T) {
	value := ActionRedemption{SessionSecret: strings.Repeat("S", 43), CSRFToken: strings.Repeat("C", 96)}
	if got := fmt.Sprintf("%v", value); strings.Contains(got, value.SessionSecret) || strings.Contains(got, value.CSRFToken) {
		t.Fatal("action redemption formatting exposed bearer material")
	}
	if _, err := json.Marshal(value); err == nil {
		t.Fatal("action redemption serialization accepted bearer material")
	}
	session := SessionIssue{Secret: strings.Repeat("N", 43), CSRFToken: strings.Repeat("D", 96)}
	if got := fmt.Sprintf("%v", session); strings.Contains(got, session.Secret) || strings.Contains(got, session.CSRFToken) {
		t.Fatal("session issue formatting exposed bearer material")
	}
	if _, err := json.Marshal(session); err == nil {
		t.Fatal("session issue serialization accepted bearer material")
	}
}

func TestAuthKDFBoundedQueueRejectsOverflow(t *testing.T) {
	for i := 0; i < cap(identityKDFSlots); i++ {
		identityKDFSlots <- struct{}{}
	}
	for i := 0; i < cap(identityKDFQueue); i++ {
		identityKDFQueue <- struct{}{}
	}
	defer func() {
		for len(identityKDFQueue) > 0 {
			<-identityKDFQueue
		}
		for len(identityKDFSlots) > 0 {
			<-identityKDFSlots
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := HashPassword(ctx, "A valid fake password for KDF queue"); err != ErrOverloaded {
		t.Fatal("full KDF admission queue did not fail closed")
	}
}

type identityTestClock func() time.Time

func (c identityTestClock) Now() time.Time { return c() }
