package staff

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

func NewSessionToken() (raw string, hash [32]byte, err error) {
	token := make([]byte, 32)
	if _, err = rand.Read(token); err != nil {
		return "", hash, ErrUnavailable
	}
	raw = base64.RawURLEncoding.EncodeToString(token)
	return raw, sha256.Sum256(token), nil
}
func HashSessionToken(raw string) ([32]byte, error) {
	if len(raw) != 43 {
		return [32]byte{}, ErrUnauthenticated
	}
	token, err := base64.RawURLEncoding.Strict().DecodeString(raw)
	if err != nil || len(token) != 32 || base64.RawURLEncoding.EncodeToString(token) != raw {
		return [32]byte{}, ErrUnauthenticated
	}
	return sha256.Sum256(token), nil
}
func EqualTokenHashes(a, b [32]byte) bool { return subtle.ConstantTimeCompare(a[:], b[:]) == 1 }
