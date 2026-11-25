package staff

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

// A valid public work factor, not a real account credential. Unknown-user work
// uses the identical gate/Argon2 path; this constant must never provision staff.
const DummyPasswordHash = "$argon2id$v=19$m=65536,t=3,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

var processKDFGate = make(chan struct{}, 2)
var processKDFQueue = make(chan struct{}, 16)

type Passwords struct{}

func NewPasswords() *Passwords { return &Passwords{} }

func validPassword(password string) bool {
	if len(password) > 512 || !utf8.ValidString(password) {
		return false
	}
	n := utf8.RuneCountInString(password)
	return n >= 15 && n <= 128
}

func acquireKDF(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case processKDFQueue <- struct{}{}:
	default:
		return nil, ErrOverloaded
	}
	defer func() { <-processKDFQueue }()
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case processKDFGate <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-processKDFGate
			return nil, err
		}
		return func() { <-processKDFGate }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, ErrOverloaded
	}
}

func (p *Passwords) Hash(ctx context.Context, password string) (string, error) {
	if !validPassword(password) {
		return "", ErrValidation
	}
	release, err := acquireKDF(ctx)
	if err != nil {
		return "", err
	}
	defer release()
	salt := make([]byte, 16)
	if _, err = rand.Read(salt); err != nil {
		return "", ErrUnavailable
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 65536, 1, 32)
	if err = ctx.Err(); err != nil {
		return "", err
	}
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=1$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

func parsePHC(phc string) (salt, key []byte, iterations uint32, err error) {
	if len(phc) > 256 {
		return nil, nil, 0, ErrValidation
	}
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" {
		return nil, nil, 0, ErrValidation
	}
	// Explicit supported bounded profiles only. t=2 is reserved for an explicit
	// old-profile upgrade; no arbitrary attacker-selected allocation parameters.
	switch parts[3] {
	case "m=65536,t=3,p=1":
		iterations = 3
	case "m=65536,t=2,p=1":
		iterations = 2
	default:
		return nil, nil, 0, ErrValidation
	}
	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != 16 || base64.RawStdEncoding.EncodeToString(salt) != parts[4] {
		return nil, nil, 0, ErrValidation
	}
	key, err = base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(key) != 32 || base64.RawStdEncoding.EncodeToString(key) != parts[5] {
		return nil, nil, 0, ErrValidation
	}
	return salt, key, iterations, nil
}
func (p *Passwords) Verify(ctx context.Context, phc, password string) (bool, error) {
	if !validPassword(password) {
		return false, ErrValidation
	}
	salt, key, iterations, err := parsePHC(phc)
	if err != nil {
		return false, err
	}
	release, err := acquireKDF(ctx)
	if err != nil {
		return false, err
	}
	defer release()
	actual := argon2.IDKey([]byte(password), salt, iterations, 65536, 1, 32)
	if err = ctx.Err(); err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(actual, key) == 1, nil
}
func (p *Passwords) VerifyDummy(ctx context.Context, password string) error {
	_, err := p.Verify(ctx, DummyPasswordHash, password)
	return err
}
