// Package identity owns customer identity contracts. HTTP, protocol-library and
// persistence adapters remain outside these transport-independent types.
package identity

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrValidation      = errors.New("identity: invalid fields")
	ErrConflict        = errors.New("identity: conflicting state")
	ErrUnauthenticated = errors.New("identity: authentication required")
	ErrDenied          = errors.New("identity: action denied")
	ErrNotFound        = errors.New("identity: resource unavailable")
	ErrUnavailable     = errors.New("identity: dependency unavailable")
	ErrOverloaded      = errors.New("identity: admission capacity unavailable")
)

// MutationFailure retains a fixed, non-secret operation phase for in-process
// diagnostics. The wrapped cause is a safe sentinel only, never driver/input text.
type MutationFailure struct {
	phase      string
	cause      error
	sqlState   string
	constraint string
}

func (e *MutationFailure) Error() string      { return "identity: action transaction failed" }
func (e *MutationFailure) Unwrap() error      { return e.cause }
func (e *MutationFailure) Phase() string      { return e.phase }
func (e *MutationFailure) SQLState() string   { return e.sqlState }
func (e *MutationFailure) Constraint() string { return e.constraint }

type Clock interface{ Now() time.Time }
type SecretHash [32]byte

func HashSecret(value string) SecretHash { return sha256.Sum256([]byte(value)) }

type SessionKind string

const (
	SessionRestricted SessionKind = "restricted"
	SessionFull       SessionKind = "full"
)

type ActionPurpose string

const (
	ActionVerifyEmail   ActionPurpose = "verify_email"
	ActionSetupPassword ActionPurpose = "setup_password"
	ActionResetPassword ActionPurpose = "reset_password"
)

// RequestMeta contains trusted boundary output, never a caller-selected user ID.
type RequestMeta struct {
	Source         string
	SessionHash    SecretHash
	EnrollmentHash SecretHash
}
type RegisterInput struct{ DisplayName, Email, Telephone, Password string }
type LoginInput struct{ Email, Password string }
type ChangePasswordInput struct {
	CurrentPassword, NewPassword string
	SessionHash                  SecretHash
}
type RequestActionInput struct {
	Email                       string
	Purpose                     ActionPurpose
	SessionHash, EnrollmentHash SecretHash
}
type RedeemActionInput struct {
	Token, Password, Telephone string
	// EnrollmentHash is selected by the server from validated browser binding.
	EnrollmentHash, SessionHash SecretHash
}
type InspectActionInput struct{ Token string }

func redactIdentityValue(w fmt.State)            { _, _ = io.WriteString(w, "[redacted customer identity input]") }
func (RegisterInput) Format(w fmt.State, _ rune) { redactIdentityValue(w) }
func (RegisterInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("customer registration input is not serializable")
}
func (LoginInput) Format(w fmt.State, _ rune) { redactIdentityValue(w) }
func (LoginInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("customer login input is not serializable")
}
func (ChangePasswordInput) Format(w fmt.State, _ rune) { redactIdentityValue(w) }
func (ChangePasswordInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("customer password change input is not serializable")
}
func (RedeemActionInput) Format(w fmt.State, _ rune) { redactIdentityValue(w) }
func (RedeemActionInput) MarshalJSON() ([]byte, error) {
	return nil, errors.New("customer action input is not serializable")
}

// PublicAccepted has the same public representation for eligible/ineligible
// requests. Registration receipt authority must never be inferred from this DTO.
type PublicAccepted struct{}

// RegistrationResult carries the same-shaped opaque receipt on both creation
// and duplicate paths. It contains no user identity or eligibility flag.
type RegistrationResult struct {
	Accepted         PublicAccepted
	EnrollmentSecret string
	SessionSecret    string
}

func (RegistrationResult) Format(w fmt.State, _ rune) {
	_, _ = io.WriteString(w, "[redacted identity registration result]")
}
func (RegistrationResult) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity registration result is not serializable")
}

type ActionInspection struct {
	Purpose       ActionPurpose
	ExpiresAt     time.Time
	RequiresLogin bool
	DisplayName   string
}
type ActionRedemption struct {
	SessionSecret, CSRFToken string
	SessionExpiresAt         time.Time
	CookieMaxAge             int
	RequiresFreshLogin       bool
}

func (ActionRedemption) Format(w fmt.State, _ rune) {
	_, _ = io.WriteString(w, "[redacted identity action redemption]")
}
func (ActionRedemption) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity action redemption is not serializable")
}

type SessionView struct {
	Kind                                   SessionKind
	Subject, SessionID, DisplayName, Email string
	EmailVerified                          bool
	AuthenticatedAt, IdleExpiresAt         time.Time
	CredentialGeneration                   int64
}
type SessionIssue struct {
	Secret, CSRFToken string
	EnrollmentSecret  string
	View              SessionView
	ExpiresAt         time.Time
	CookieMaxAge      int
}

func (SessionIssue) Format(w fmt.State, _ rune) {
	_, _ = io.WriteString(w, "[redacted identity session issue]")
}
func (SessionIssue) MarshalJSON() ([]byte, error) {
	return nil, errors.New("identity session issue is not serializable")
}

// ActionDelivery exists only between committed issuance and bounded delivery.
// It must not be persisted, logged, or serialized into a public response.
type ActionDelivery struct {
	AttemptID, Recipient, DisplayName, URL string
	Purpose                                ActionPurpose
	ExpiresAt                              time.Time
}
type DeliverySender interface {
	Send(context.Context, ActionDelivery) error
}

func (ActionDelivery) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[redacted action delivery]")
}
func (ActionDelivery) MarshalJSON() ([]byte, error) {
	return nil, errors.New("action delivery is not serializable")
}
