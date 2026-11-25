package handlers

import (
	"net/http"
	"testing"
	"time"
)

func TestAuthCookieProductionAndDevelopmentSeparation(t *testing.T) {
	expires := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name        string
		development bool
		cookieName  string
		secure      bool
	}{
		{"production host cookie", false, "__Host-zenith_auth", true},
		{"development loopback cookie", true, "zenith_auth_dev", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cookie := authSessionCookie("opaque-session", expires, 1800, tc.development)
			if cookie.Name != tc.cookieName || cookie.Value != "opaque-session" || cookie.Path != "/" || cookie.Domain != "" || !cookie.HttpOnly || cookie.Secure != tc.secure || cookie.SameSite != http.SameSiteLaxMode || !cookie.Expires.Equal(expires) || cookie.MaxAge != 1800 {
				t.Fatal("session cookie violated its environment-specific contract")
			}
			cleared := clearingAuthCookie(tc.development)
			if cleared.Name != tc.cookieName || cleared.Value != "" || cleared.MaxAge != -1 || cleared.Path != "/" || cleared.Domain != "" || !cleared.HttpOnly || cleared.Secure != tc.secure {
				t.Fatal("session cookie clearing did not match issued scope")
			}
			enrollment := clearingEnrollmentCookie(tc.development)
			wantEnrollment := "__Host-zenith_enrollment"
			if tc.development {
				wantEnrollment = "zenith_enrollment_dev"
			}
			if enrollment.Name != wantEnrollment || enrollment.MaxAge != -1 || enrollment.Path != "/" || enrollment.Domain != "" || !enrollment.HttpOnly || enrollment.Secure != tc.secure {
				t.Fatal("enrollment cookie clearing crossed runtime scopes")
			}
		})
	}
}
