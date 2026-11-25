package identity

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nyaruka/phonenumbers"
	"golang.org/x/crypto/argon2"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)
var phoneInputPattern = regexp.MustCompile(`^[0-9 +().-]+$`)
var identityKDFSlots = make(chan struct{}, 2)
var identityKDFQueue = make(chan struct{}, 16)

type NormalizedRegistration struct{ DisplayName, Email, Telephone, Password string }

// NormalizeRegistration accepts only explicit international phone numbers.
// Telephone shape plausibility is not a claim of assignment or ownership.
func NormalizeRegistration(in RegisterInput) (NormalizedRegistration, error) {
	if !utf8.ValidString(in.DisplayName) || strings.ContainsRune(in.DisplayName, 0) || utf8.RuneCountInString(in.DisplayName) > 200 || strings.TrimSpace(in.DisplayName) == "" ||
		!utf8.ValidString(in.Email) || strings.ContainsRune(in.Email, 0) || len(in.Email) > 254 ||
		!utf8.ValidString(in.Telephone) || len(in.Telephone) > 64 || !phoneInputPattern.MatchString(in.Telephone) ||
		!validPassword(in.Password) {
		return NormalizedRegistration{}, ErrValidation
	}
	name := strings.TrimSpace(in.DisplayName)
	address := strings.TrimSpace(in.Email)
	if !emailPattern.MatchString(address) || strings.ContainsAny(address, "\r\n") {
		return NormalizedRegistration{}, ErrValidation
	}
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Address != address || !strings.Contains(parsed.Address, "@") || strings.ContainsAny(parsed.Address, "\r\n") {
		return NormalizedRegistration{}, ErrValidation
	}
	normalizedEmail := strings.ToLower(parsed.Address)
	e164, err := NormalizeTelephone(in.Telephone)
	if err != nil {
		return NormalizedRegistration{}, err
	}
	return NormalizedRegistration{DisplayName: name, Email: normalizedEmail, Telephone: e164, Password: in.Password}, nil
}

func NormalizeTelephone(value string) (string, error) {
	if !utf8.ValidString(value) || len(value) > 64 || !phoneInputPattern.MatchString(value) {
		return "", ErrValidation
	}
	raw := strings.TrimSpace(value)
	if !strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "++") {
		return "", ErrValidation
	}
	number, err := phonenumbers.Parse(raw, "ZZ")
	if err != nil || !phonenumbers.IsValidNumber(number) {
		return "", ErrValidation
	}
	e164 := phonenumbers.Format(number, phonenumbers.E164)
	if len(e164) < 3 || len(e164) > 16 || e164[0] != '+' || e164[1] < '1' || e164[1] > '9' {
		return "", ErrValidation
	}
	for _, c := range e164[1:] {
		if c < '0' || c > '9' {
			return "", ErrValidation
		}
	}
	return e164, nil
}

const dummyIdentityPasswordHash = "$argon2id$v=19$m=65536,t=3,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

func validPassword(value string) bool {
	return len(value) <= 512 && utf8.ValidString(value) && utf8.RuneCountInString(value) >= 15 && utf8.RuneCountInString(value) <= 128
}

func acquireIdentityKDF(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case identityKDFQueue <- struct{}{}:
	default:
		return nil, ErrOverloaded
	}
	defer func() { <-identityKDFQueue }()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case identityKDFSlots <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-identityKDFSlots
			return nil, err
		}
		return func() { <-identityKDFSlots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, ErrOverloaded
	}
}

func HashPassword(ctx context.Context, password string) (string, error) {
	if !validPassword(password) {
		return "", ErrValidation
	}
	release, err := acquireIdentityKDF(ctx)
	if err != nil {
		return "", err
	}
	defer release()
	salt := make([]byte, 16)
	if _, err = rand.Read(salt); err != nil {
		return "", ErrUnavailable
	}
	derived := argon2.IDKey([]byte(password), salt, 3, 65536, 1, 32)
	if err = ctx.Err(); err != nil {
		return "", err
	}
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=1$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(derived)), nil
}

func verifyIdentityPassword(ctx context.Context, phc, password string) (bool, error) {
	if !validPassword(password) || len(phc) > 256 {
		return false, ErrValidation
	}
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=65536,t=3,p=1" {
		return false, ErrValidation
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != 16 || base64.RawStdEncoding.EncodeToString(salt) != parts[4] {
		return false, ErrValidation
	}
	want, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(want) != 32 || base64.RawStdEncoding.EncodeToString(want) != parts[5] {
		return false, ErrValidation
	}
	release, err := acquireIdentityKDF(ctx)
	if err != nil {
		return false, err
	}
	defer release()
	got := argon2.IDKey([]byte(password), salt, 3, 65536, 1, 32)
	if err = ctx.Err(); err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// VerifyUnknownPassword performs the same bounded real password work for unknown
// accounts. It never returns identity authority or distinguishes account state.
func VerifyUnknownPassword(ctx context.Context, password string) error {
	_, err := verifyIdentityPassword(ctx, dummyIdentityPasswordHash, password)
	return err
}
