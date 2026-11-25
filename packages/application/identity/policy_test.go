package identity

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAuthPolicyDurations(t *testing.T) {
	if SessionIdleLifetime != 49*24*time.Hour || VerifyEmailLifetime != 24*time.Hour || SetupPasswordLifetime != 24*time.Hour || ResetPasswordLifetime != 30*time.Minute || RestrictedSessionLifetime != 30*time.Minute {
		t.Fatal("fixed session/action contracts changed")
	}
	if AuthorizationLifetime != 10*time.Minute || CodeLifetime != time.Minute || TokenLifetime != 5*time.Minute {
		t.Fatal("protocol maximum defaults changed")
	}
	if ForegroundInterval != time.Minute || CookieRenewalWindow != 7*24*time.Hour || LogoutReceiptLifetime != 5*time.Minute || RefreshLeaseLifetime != 15*time.Second {
		t.Fatal("continuity defaults changed")
	}
}

func TestAuthPolicyExactExpiry(t *testing.T) {
	start := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	for name, lifetime := range map[string]time.Duration{
		"verify_email":       VerifyEmailLifetime,
		"setup_password":     SetupPasswordLifetime,
		"reset_password":     ResetPasswordLifetime,
		"restricted_session": RestrictedSessionLifetime,
		"full_session":       SessionIdleLifetime,
	} {
		t.Run(name, func(t *testing.T) {
			deadline := start.Add(lifetime)
			for _, tc := range []struct {
				now  time.Time
				want bool
			}{
				{deadline.Add(-time.Nanosecond), true}, {deadline, false}, {deadline.Add(time.Nanosecond), false},
			} {
				if got := BeforeExpiry(tc.now, deadline); got != tc.want {
					t.Fatalf("expiry boundary: got %v want %v", got, tc.want)
				}
			}
			if BeforeExpiry(time.Time{}, deadline) || BeforeExpiry(deadline, time.Time{}) {
				t.Fatal("unset instants must fail closed")
			}
		})
	}
}

func TestAuthDeliveryCannotEscape(t *testing.T) {
	d := ActionDelivery{AttemptID: "fake-attempt-sensitive", Recipient: "fake-sensitive@example.test", DisplayName: "fake-sensitive-name", URL: "https://auth.test/?token=fake-sensitive-token", Purpose: ActionVerifyEmail}
	for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		for _, value := range []any{d, &d} {
			if output := fmt.Sprintf(format, value); strings.Contains(output, "sensitive") {
				t.Fatalf("delivery leaks through %s", format)
			}
		}
	}
	for _, value := range []any{d, &d} {
		if data, err := json.Marshal(value); err == nil || strings.Contains(string(data), "sensitive") {
			t.Fatal("delivery may not be serialized")
		}
	}
}
