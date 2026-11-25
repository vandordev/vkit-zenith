package identity

import "time"

const (
	SessionIdleLifetime       = 49 * 24 * time.Hour
	VerifyEmailLifetime       = 24 * time.Hour
	SetupPasswordLifetime     = 24 * time.Hour
	ResetPasswordLifetime     = 30 * time.Minute
	RestrictedSessionLifetime = 30 * time.Minute
	AuthorizationLifetime     = 10 * time.Minute
	CodeLifetime              = time.Minute
	TokenLifetime             = 5 * time.Minute
	ForegroundInterval        = time.Minute
	CookieRenewalWindow       = 7 * 24 * time.Hour
	LogoutReceiptLifetime     = 5 * time.Minute
	RefreshLeaseLifetime      = 15 * time.Second
)

// BeforeExpiry rejects the exact deadline and unset clocks/deadlines. Callers
// select the acceptance instant after acquiring their live eligibility locks.
func BeforeExpiry(now, deadline time.Time) bool {
	return !now.IsZero() && !deadline.IsZero() && now.Before(deadline)
}
