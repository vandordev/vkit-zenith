package staff

import (
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

// Sentinel categories are safe across adapters. Never wrap raw SQL/KDF/input.
var (
	ErrValidation      = errors.New("staff: invalid fields")
	ErrUnauthenticated = errors.New("staff: authentication required")
	ErrDenied          = errors.New("staff: action denied")
	ErrConflict        = errors.New("staff: conflicting state")
	ErrNotFound        = errors.New("staff: resource unavailable")
	ErrUnavailable     = errors.New("staff: dependency unavailable")
	ErrOverloaded      = errors.New("staff: password capacity unavailable")
	ErrRateLimited     = errors.New("staff: attempt limit reached")
)

type RateLimitError struct{ RetryAfter int }

func (e *RateLimitError) Error() string { return ErrRateLimited.Error() }
func (e *RateLimitError) Unwrap() error { return ErrRateLimited }

type Actor struct {
	StaffID     pgtype.UUID
	SessionHash [32]byte
	Source      string
}
type Staff struct {
	ID                               pgtype.UUID
	Email, DisplayName, Role, Status string
	UpdatedAt                        time.Time
}
type LoginInput struct{ Email, Password, Source string }
type LoginResult struct {
	Staff     Staff
	Token     string
	ExpiresAt time.Time
}
type StaffWrite struct {
	Email, DisplayName, Role, Status    string
	Password, CurrentPassword, Revision string
}
type PasswordChange struct{ CurrentPassword, NewPassword string }
