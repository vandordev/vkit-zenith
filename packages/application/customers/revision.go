package customers

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func canonical(t time.Time) time.Time { return t.UTC().Truncate(time.Microsecond) }
func digest(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("customers: invalid canonical revision")
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func UserRevision(v User) string {
	v.CreatedAt, v.UpdatedAt = canonical(v.CreatedAt), canonical(v.UpdatedAt)
	if v.EmailVerifiedAt != nil {
		at := canonical(*v.EmailVerifiedAt)
		v.EmailVerifiedAt = &at
	}
	return digest(v)
}
func OrganizationRevision(v Organization) string {
	v.CreatedAt, v.UpdatedAt = canonical(v.CreatedAt), canonical(v.UpdatedAt)
	return digest(v)
}
func MembershipRevision(v Membership) string {
	v.CreatedAt, v.UpdatedAt = canonical(v.CreatedAt), canonical(v.UpdatedAt)
	return digest(v)
}
func revision(actual, expected string) error {
	if len(expected) != 64 {
		return staff.ErrConflict
	}
	if _, err := hex.DecodeString(expected); err != nil {
		return staff.ErrConflict
	}
	if subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) != 1 {
		return staff.ErrConflict
	}
	return nil
}
