//go:build integration

package securitytest

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/handlers"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/server"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
	"github.com/vandordev/vkit-zenith/packages/tools/authtest"
)

const authTestOrigin = "https://web.example.test"

type authTestClock struct{ at time.Time }

func (c *authTestClock) Now() time.Time { return c.at }

type authTestTransport struct{ capture *authtest.EmailCapture }

func (t authTestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.capture.RoundTrip(r)
}

func authRegistrationRouter(t *testing.T, pool *pgxpool.Pool, clock identity.Clock, capture *authtest.EmailCapture) *echo.Echo {
	t.Helper()
	limiter := [32]byte{1}
	csrfKey := [32]byte{2}
	sender, err := identityhttp.NewDeliverySender(authTestOrigin, false, "fake-resend-key", "noreply@example.test", authTestTransport{capture})
	if err != nil {
		t.Fatal("fake delivery boundary failed")
	}
	service := identity.NewService(pool, clock, sender, limiter, csrfKey, authTestOrigin, false)
	runtime := identityhttp.NewRuntime(service, authTestOrigin, false)
	e := server.NewHTTP(slog.New(slog.NewJSONHandler(io.Discard, nil)), authTestOrigin)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.SetRequest(c.Request().WithContext(identityhttp.WithRequestSource(c.Request().Context(), "127.0.0.1")))
			return next(c)
		}
	})
	c := huma.DefaultConfig("Zenith API fixture", "1")
	c.OpenAPIPath = ""
	c.DocsPath = ""
	c.SchemasPath = ""
	c.CreateHooks = nil
	api := humaecho.NewV4(e, c)
	for _, register := range []func(huma.API, *identityhttp.Runtime) error{handlers.NewGetAuthCSRF, handlers.NewRegisterCustomer, handlers.NewRequestPasswordAction, handlers.NewInspectEmailAction, handlers.NewConfirmEmailAction, handlers.NewResendVerification, handlers.NewLoginCustomer, handlers.NewGetCurrentCustomerSession, handlers.NewLogoutCustomer, handlers.NewTouchCustomerSession} {
		if err = register(api, runtime); err != nil {
			t.Fatal("Auth route registration failed")
		}
	}
	return e
}

func csrfFor(t *testing.T, e *echo.Echo, cookies ...*http.Cookie) string {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/csrf", nil)
	r.Header.Set("Origin", authTestOrigin)
	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatal("CSRF bootstrap unavailable")
	}
	if w.Header().Get("Access-Control-Allow-Origin") != authTestOrigin || w.Header().Get("Access-Control-Allow-Credentials") != "true" || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Referrer-Policy") != "no-referrer" {
		t.Fatal("Auth CSRF bootstrap headers do not match the exact credentialed/no-store policy")
	}
	var body struct {
		Token string `json:"token"`
	}
	if json.Unmarshal(w.Body.Bytes(), &body) != nil || len(body.Token) != 96 {
		t.Fatal("CSRF bootstrap contract invalid")
	}
	if len(w.Result().Cookies()) != 0 {
		t.Fatal("CSRF token persisted in a cookie instead of memory")
	}
	return body.Token
}

func performRegisterHTTP(t *testing.T, e *echo.Echo, csrf, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/register", strings.NewReader(body))
	r.Header.Set("Origin", authTestOrigin)
	r.Header.Set("X-CSRF-Token", csrf)
	r.Header.Set("Content-Type", "application/json")
	r.RemoteAddr = "127.0.0.1:43123"
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	return w
}

func TestAuthAccountHTTPRegistrationPrivacyAndCapture(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stage := "fixture"
	err := authtest.WithFixture(ctx, func(ctx context.Context, f *authtest.Fixture) error {
		clock := &authTestClock{at: time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)}
		capture := authtest.NewEmailCapture(16)
		router := authRegistrationRouter(t, f.Pool, clock, capture)
		if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES('10000000-0000-4000-8000-000000000098','disabled-login@example.test','disabled-login@example.test','Disabled login','disabled',$1,$1),('10000000-0000-4000-8000-000000000097','no-password@example.test','no-password@example.test','No password','active',$1,$1)`, clock.at); err != nil {
			return err
		}
		stage = "csrf-bootstrap"
		token := csrfFor(t, router)
		stage = "first-http-registration"
		first := performRegisterHTTP(t, router, token, `{"displayName":" Rani ","email":" PERSON+tag@Example.TEST ","telephone":"+62 812-3456-7890","password":"  Correct Horse Battery Staple  "}`)
		stage = fmt.Sprintf("first-http-status-%d", first.Code)
		if first.Code != http.StatusAccepted {
			return errors.New("registration status was not accepted")
		}
		if first.Body.String() != "{\"accepted\":true}\n" {
			return errors.New("registration body was not generic accepted shape")
		}
		stage = "first-cookies"
		var enroll, session *http.Cookie
		for _, c := range first.Result().Cookies() {
			if c.Name == "__Host-zenith_enrollment" {
				enroll = c
			}
			if c.Name == "__Host-zenith_auth" {
				session = c
			}
		}
		if enroll == nil || session == nil || enroll.Value == "" || session.Value == "" || len(enroll.Value) != 43 || len(session.Value) != 43 {
			return errors.New("generic authority cookies missing or malformed")
		}
		for _, c := range []*http.Cookie{enroll, session} {
			if !c.Secure || !c.HttpOnly || c.Path != "/" || c.Domain != "" || c.SameSite != http.SameSiteLaxMode || c.MaxAge != int(identity.RestrictedSessionLifetime.Seconds()) {
				return errors.New("registration cookie flags are unsafe")
			}
		}
		stage = "verification-capture"
		messages := capture.Snapshot()
		if len(messages) != 1 || !strings.Contains(messages[0].HTML, "Konfirmasi alamat email Anda") || !strings.Contains(messages[0].Text, "https://web.example.test/verify-email?") {
			return errors.New("HTTP journey did not capture actual verification template")
		}
		stage = "read-only-verification-inspection"
		prefix := "https://web.example.test/verify-email?"
		start := strings.Index(messages[0].Text, prefix)
		if start < 0 {
			return errors.New("captured verification action URL missing")
		}
		end := strings.Index(messages[0].Text[start:], "\n")
		if end < 0 {
			return errors.New("captured verification action URL malformed")
		}
		mailURL, err := url.Parse(strings.TrimSpace(messages[0].Text[start : start+end]))
		if err != nil {
			return errors.New("captured verification action URL malformed")
		}
		inspect := func(includeCookies bool) *httptest.ResponseRecorder {
			r := httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/action/inspect?token="+url.QueryEscape(mailURL.Query().Get("token")), nil)
			r.Header.Set("Origin", authTestOrigin)
			if includeCookies {
				r.AddCookie(session)
				r.AddCookie(enroll)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			return w
		}
		validInspect := inspect(true)
		if validInspect.Code != http.StatusOK || !strings.Contains(validInspect.Body.String(), `"purpose":"verify_email"`) || strings.Contains(validInspect.Body.String(), "person+tag@example.test") || !strings.Contains(validInspect.Body.String(), "Rani") {
			return errors.New("matching restricted browser could not safely inspect verification action")
		}
		crossBrowser := inspect(false)
		if crossBrowser.Code != http.StatusOK || !strings.Contains(crossBrowser.Body.String(), `"requiresLogin":true`) || strings.Contains(crossBrowser.Body.String(), "Rani") {
			return errors.New("cross-browser link exposed identity or bypassed enrollment binding")
		}
		crossCSRF := csrfFor(t, router)
		crossBody, _ := json.Marshal(map[string]string{"token": mailURL.Query().Get("token")})
		crossConfirm := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/action/confirm", strings.NewReader(string(crossBody)))
		crossConfirm.Header.Set("Origin", authTestOrigin)
		crossConfirm.Header.Set("X-CSRF-Token", crossCSRF)
		crossConfirm.Header.Set("Content-Type", "application/json")
		crossConfirm.RemoteAddr = "127.0.0.1:43123"
		crossResponse := httptest.NewRecorder()
		router.ServeHTTP(crossResponse, crossConfirm)
		if crossResponse.Code != http.StatusNotFound {
			return errors.New("cross-browser verification promoted an unbound enrollment")
		}
		var unconsumed bool
		tokenHash := sha256.Sum256([]byte(mailURL.Query().Get("token")))
		if err = f.Pool.QueryRow(ctx, `SELECT consumed_at IS NULL FROM email_action_tokens WHERE token_hash=$1`, tokenHash[:]).Scan(&unconsumed); err != nil || !unconsumed {
			return errors.New("GET action inspection consumed token")
		}
		stage = "restricted-explicit-verification-resend"
		resendCSRF := csrfFor(t, router, session, enroll)
		resendRequest := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/verification/resend", strings.NewReader(`{}`))
		resendRequest.Header.Set("Origin", authTestOrigin)
		resendRequest.Header.Set("X-CSRF-Token", resendCSRF)
		resendRequest.Header.Set("Content-Type", "application/json")
		resendRequest.RemoteAddr = "127.0.0.1:43123"
		resendRequest.AddCookie(session)
		resendRequest.AddCookie(enroll)
		resendResponse := httptest.NewRecorder()
		router.ServeHTTP(resendResponse, resendRequest)
		if resendResponse.Code != http.StatusAccepted || len(capture.Snapshot()) != 2 {
			return errors.New("restricted explicit resend did not preserve/send second valid link")
		}
		var validVerifyTokens int
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens WHERE user_id=(SELECT id FROM users WHERE email_normalized='person+tag@example.test') AND purpose='verify_email' AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$1`, clock.at).Scan(&validVerifyTokens); err != nil {
			return err
		}
		if validVerifyTokens != 2 {
			return errors.New("resend invalidated an older unconsumed verification link")
		}
		stage = "same-browser-verification-confirm"
		verifyCSRF := csrfFor(t, router, session, enroll)
		confirmBody, _ := json.Marshal(map[string]string{"token": mailURL.Query().Get("token")})
		confirmRequest := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/action/confirm", strings.NewReader(string(confirmBody)))
		confirmRequest.Header.Set("Origin", authTestOrigin)
		confirmRequest.Header.Set("X-CSRF-Token", verifyCSRF)
		confirmRequest.Header.Set("Content-Type", "application/json")
		confirmRequest.RemoteAddr = "127.0.0.1:43123"
		confirmRequest.AddCookie(session)
		confirmRequest.AddCookie(enroll)
		confirmResponse := httptest.NewRecorder()
		router.ServeHTTP(confirmResponse, confirmRequest)
		if confirmResponse.Code != http.StatusOK || !strings.Contains(confirmResponse.Body.String(), `"completed":true`) || !strings.Contains(confirmResponse.Body.String(), `"requiresFreshLogin":false`) {
			var problem struct {
				Errors []struct {
					Location []string `json:"location"`
				} `json:"errors"`
			}
			_ = json.Unmarshal(confirmResponse.Body.Bytes(), &problem)
			stage = fmt.Sprintf("confirm-http-%d-validation-fields-%v", confirmResponse.Code, problem.Errors)
			transport, deliveryErr := identityhttp.NewDeliverySender(authTestOrigin, false, "fake-resend-key", "noreply@example.test", authTestTransport{capture})
			if deliveryErr == nil {
				diagnostic := identity.NewService(f.Pool, clock, transport, [32]byte{1}, [32]byte{2}, authTestOrigin, false)
				_, redemptionErr := diagnostic.RedeemAction(ctx, identity.RedeemActionInput{Token: mailURL.Query().Get("token"), EnrollmentHash: identity.HashSecret(enroll.Value), SessionHash: identity.HashSecret(session.Value)})
				var phase *identity.MutationFailure
				if errors.As(redemptionErr, &phase) {
					stage = fmt.Sprintf("confirm-%d-%s-%v", confirmResponse.Code, phase.Phase(), errors.Is(redemptionErr, identity.ErrValidation))
				} else if redemptionErr == nil {
					stage = fmt.Sprintf("confirm-%d-direct-success", confirmResponse.Code)
				} else {
					stage = fmt.Sprintf("confirm-%d-direct-v%t-d%t-n%t-u%t", confirmResponse.Code, errors.Is(redemptionErr, identity.ErrValidation), errors.Is(redemptionErr, identity.ErrDenied), errors.Is(redemptionErr, identity.ErrNotFound), errors.Is(redemptionErr, identity.ErrUnavailable))
				}
			}
			return errors.New("same-browser verification did not explicitly confirm")
		}
		stage = "verification-cookie-rotation"
		var rotated *http.Cookie
		cookieNames := make([]string, 0, 2)
		for _, cookie := range confirmResponse.Result().Cookies() {
			cookieNames = append(cookieNames, cookie.Name)
			if cookie.Name == "__Host-zenith_auth" && cookie.Value != "" {
				rotated = cookie
			}
			if cookie.Name == "__Host-zenith_enrollment" && cookie.MaxAge >= 0 {
				stage = "enrollment-cookie-not-cleared"
				return errors.New("verification did not clear enrollment receipt cookie")
			}
		}
		if rotated == nil || rotated.Value == session.Value {
			stage = fmt.Sprintf("verification-session-rotation-count-%d-names-%s", len(confirmResponse.Result().Cookies()), strings.Join(cookieNames, ","))
			return errors.New("verification did not rotate central browser session")
		}
		if rotated.MaxAge != int(identity.SessionIdleLifetime.Seconds()) {
			stage = fmt.Sprintf("verification-cookie-max-age-%d", rotated.MaxAge)
			return errors.New("rotated browser cookie max-age differs from the application idle policy")
		}
		var confirmation struct {
			CSRFToken string `json:"csrfToken"`
		}
		if err = json.Unmarshal(confirmResponse.Body.Bytes(), &confirmation); err != nil || len(confirmation.CSRFToken) != 96 {
			return errors.New("verification did not return a fresh in-memory CSRF token")
		}
		delivery, deliveryErr := identityhttp.NewDeliverySender(authTestOrigin, false, "fake-resend-key", "noreply@example.test", authTestTransport{capture})
		if deliveryErr != nil {
			return errors.New("test delivery boundary unavailable")
		}
		csrfService := identity.NewService(f.Pool, clock, delivery, [32]byte{1}, [32]byte{2}, authTestOrigin, false)
		if !csrfService.ValidateSessionCSRF(identity.HashSecret(rotated.Value), confirmation.CSRFToken) {
			return errors.New("rotated session did not accept its returned CSRF token")
		}
		var verified, fullKind, authTime, sessionExpiry, sessionHashOK, tokenConsumed bool
		rotatedHash := sha256.Sum256([]byte(rotated.Value))
		stage = "verification-state-query"
		if err = f.Pool.QueryRow(ctx, `SELECT u.email_verified_at IS NOT NULL,s.kind='full',s.authenticated_at=$2,s.expires_at=$2::timestamptz+interval '1176 hours',s.secret_hash=$1,t.consumed_at IS NOT NULL FROM users u JOIN auth_sessions s ON s.user_id=u.id JOIN email_action_tokens t ON t.user_id=u.id WHERE u.email_normalized='person+tag@example.test' AND t.token_hash=$3`, rotatedHash[:], clock.at, tokenHash[:]).Scan(&verified, &fullKind, &authTime, &sessionExpiry, &sessionHashOK, &tokenConsumed); err != nil {
			return err
		}
		stage = "verification-state-assertion"
		if !verified || !fullKind || !authTime || !sessionExpiry || !sessionHashOK || !tokenConsumed {
			stage = fmt.Sprintf("verify-state-v%t-f%t-a%t-e%t-h%t-c%t", verified, fullKind, authTime, sessionExpiry, sessionHashOK, tokenConsumed)
			return errors.New("verification promotion/rotation or single-use state mismatch")
		}
		stage = "verification-competing-link-revocation"
		var competingRevoked bool
		if err = f.Pool.QueryRow(ctx, `SELECT count(*)=1 AND bool_and(revoked_at IS NOT NULL AND consumed_at IS NULL) FROM email_action_tokens WHERE user_id=(SELECT id FROM users WHERE email_normalized='person+tag@example.test') AND purpose='verify_email' AND id<>(SELECT id FROM email_action_tokens WHERE token_hash=$1)`, tokenHash[:]).Scan(&competingRevoked); err != nil {
			return err
		}
		if !competingRevoked {
			return errors.New("successful verification did not revoke competing link")
		}
		stage = "registration-account-cooldown"
		warmupDuplicate := performRegisterHTTP(t, router, csrfFor(t, router), `{"displayName":"Different","email":"person+tag@example.test","telephone":"+14155552671","password":"Another Correct Horse Battery"}`)
		if warmupDuplicate.Code != first.Code || warmupDuplicate.Body.String() != first.Body.String() || len(warmupDuplicate.Result().Cookies()) != 2 {
			stage = fmt.Sprintf("duplicate-warmup-%d-cookies-%d", warmupDuplicate.Code, len(warmupDuplicate.Result().Cookies()))
			return errors.New("duplicate request did not retain generic receipt shape")
		}
		cooldownResponse := performRegisterHTTP(t, router, csrfFor(t, router), `{"displayName":"Different","email":"person+tag@example.test","telephone":"+14155552671","password":"Another Correct Horse Battery"}`)
		cooldownCookies := cooldownResponse.Result().Cookies()
		if cooldownResponse.Code != first.Code || cooldownResponse.Body.String() != first.Body.String() || len(cooldownCookies) != 2 || len(capture.Snapshot()) != 2 {
			stage = fmt.Sprintf("duplicate-cooldown-%d-cookies-%d-mail-%d", cooldownResponse.Code, len(cooldownCookies), len(capture.Snapshot()))
			return errors.New("registration cooldown changed generic response or sent another email")
		}
		cooldownSession := cooldownCookies[1]
		cooldownCurrent := httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/session/current", nil)
		cooldownCurrent.Header.Set("Origin", authTestOrigin)
		cooldownCurrent.AddCookie(cooldownSession)
		cooldownCurrentResponse := httptest.NewRecorder()
		router.ServeHTTP(cooldownCurrentResponse, cooldownCurrent)
		if cooldownCurrentResponse.Code != http.StatusUnauthorized {
			stage = fmt.Sprintf("duplicate-cookie-current-%d", cooldownCurrentResponse.Code)
			return errors.New("cooldown receipt created authenticated authority")
		}
		stage = "first-persisted-account"
		var me bool
		if err := f.Pool.QueryRow(ctx, `SELECT (SELECT count(*)=1 FROM users WHERE email_normalized='person+tag@example.test' AND telephone='+6281234567890' AND email_verified_at IS NOT NULL) AND (SELECT count(*)=0 FROM organization_memberships) AND (SELECT count(*)=1 FROM auth_sessions WHERE secret_hash=$1 AND kind='full')`, rotatedHash[:]).Scan(&me); err != nil {
			return err
		}
		if !me {
			return errors.New("customer account got unintended business access")
		}
		clock.at = clock.at.Add(61 * time.Second)
		stage = "duplicate-http-registration"
		token = csrfFor(t, router)
		second := performRegisterHTTP(t, router, token, `{"displayName":" Rani ","email":" PERSON+tag@Example.TEST ","telephone":"+62 812-3456-7890","password":"  Correct Horse Battery Staple  "}`)
		if second.Code != first.Code || second.Body.String() != first.Body.String() {
			return errors.New("duplicate registration response enumerated account")
		}
		stage = "duplicate-cookies"
		cookies := second.Result().Cookies()
		if len(cookies) != 2 || cookies[0].Name != enroll.Name || cookies[1].Name != session.Name || len(cookies[0].Value) != 43 || len(cookies[1].Value) != 43 {
			return errors.New("duplicate cookie shape differs")
		}
		if cookies[0].Value == enroll.Value || cookies[1].Value == session.Value {
			return errors.New("duplicate credential material was reused")
		}
		duplicateCurrent := httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/session/current", nil)
		duplicateCurrent.Header.Set("Origin", authTestOrigin)
		duplicateCurrent.AddCookie(cookies[1])
		duplicateCurrentResponse := httptest.NewRecorder()
		router.ServeHTTP(duplicateCurrentResponse, duplicateCurrent)
		if duplicateCurrentResponse.Code != http.StatusUnauthorized {
			return errors.New("duplicate registration receipt issued an authenticated session")
		}
		stage = "duplicate-delivery-count"
		if len(capture.Snapshot()) != 2 {
			return errors.New("duplicate request sent verification email")
		}
		stage = "concurrent-duplicate-registration"
		clock.at = clock.at.Add(61 * time.Second)
		csrfConcurrent := csrfFor(t, router)
		const attempts = 2
		responses := make([]*httptest.ResponseRecorder, attempts)
		var wg sync.WaitGroup
		for i := range responses {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				body := `{"displayName":"Concurrent","email":" race@example.test ","telephone":"+14155552671","password":"Another Correct Horse Battery"}`
				if i == 1 {
					body = `{"displayName":"Different","email":"RACE@EXAMPLE.TEST","telephone":"+1 (415) 555-2671","password":"Another Correct Horse Battery"}`
				}
				r := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/register", strings.NewReader(body))
				r.Header.Set("Origin", authTestOrigin)
				r.Header.Set("X-CSRF-Token", csrfConcurrent)
				r.Header.Set("Content-Type", "application/json")
				r.RemoteAddr = "127.0.0.1:43123"
				w := httptest.NewRecorder()
				router.ServeHTTP(w, r)
				responses[i] = w
			}(i)
		}
		wg.Wait()
		for _, response := range responses {
			if response.Code != http.StatusAccepted || response.Body.String() != first.Body.String() || len(response.Result().Cookies()) != 2 {
				return errors.New("concurrent duplicate request exposed account state")
			}
		}
		var raceCounts int
		if err := f.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users WHERE email_normalized='race@example.test')+(SELECT count(*) FROM email_action_tokens WHERE target_email_normalized='race@example.test')`).Scan(&raceCounts); err != nil {
			return err
		}
		if raceCounts != 2 || len(capture.Snapshot()) != 3 {
			return errors.New("concurrent duplicate did not yield one user/action and safe responses")
		}
		stage = "passwordless-first-password-request"
		if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES('10000000-0000-4000-8000-000000000099','bo-setup@example.test','bo-setup@example.test','BO setup','active',$1,$1)`, clock.at); err != nil {
			return err
		}
		requestPassword := func(address string) *httptest.ResponseRecorder {
			body, _ := json.Marshal(map[string]string{"email": address})
			passwordCSRF := csrfFor(t, router)
			r := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/password/request", strings.NewReader(string(body)))
			r.Header.Set("Origin", authTestOrigin)
			r.Header.Set("X-CSRF-Token", passwordCSRF)
			r.Header.Set("Content-Type", "application/json")
			r.RemoteAddr = "127.0.0.1:43123"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			return w
		}
		setupCSRF := csrfFor(t, router)
		setupResponse := requestPassword("BO-SETUP@example.test")
		if setupResponse.Code != http.StatusAccepted {
			return errors.New("passwordless setup request not generically accepted")
		}
		messagesAfterSetup := capture.Snapshot()
		if len(messagesAfterSetup) != 4 || !strings.Contains(messagesAfterSetup[3].Subject, "kata sandi") {
			return errors.New("BO-style passwordless account did not receive setup mail")
		}
		setupMail := messagesAfterSetup[3].Text
		setupPrefix := "https://web.example.test/set-password?"
		setupStart := strings.Index(setupMail, setupPrefix)
		if setupStart < 0 {
			return errors.New("setup action URL missing")
		}
		setupEnd := strings.Index(setupMail[setupStart:], "\n")
		if setupEnd < 0 {
			return errors.New("setup action URL malformed")
		}
		setupURL, parseErr := url.Parse(strings.TrimSpace(setupMail[setupStart : setupStart+setupEnd]))
		if parseErr != nil {
			return errors.New("setup action URL malformed")
		}
		inspectRequest := httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/action/inspect?token="+url.QueryEscape(setupURL.Query().Get("token")), nil)
		inspectRequest.Header.Set("Origin", authTestOrigin)
		inspectResponse := httptest.NewRecorder()
		router.ServeHTTP(inspectResponse, inspectRequest)
		if inspectResponse.Code != http.StatusOK || !strings.Contains(inspectResponse.Body.String(), `"purpose":"setup_password"`) || !strings.Contains(inspectResponse.Body.String(), "BO setup") {
			return errors.New("first-password action inspection failed")
		}
		setupCSRF = csrfFor(t, router)
		setupConfirmBody, _ := json.Marshal(map[string]string{"token": setupURL.Query().Get("token"), "password": "A Setup Correct Horse Battery", "telephone": "+14155552671"})
		setupConfirm := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/action/confirm", strings.NewReader(string(setupConfirmBody)))
		setupConfirm.Header.Set("Origin", authTestOrigin)
		setupConfirm.Header.Set("X-CSRF-Token", setupCSRF)
		setupConfirm.Header.Set("Content-Type", "application/json")
		setupConfirm.RemoteAddr = "127.0.0.1:43123"
		setupConfirmResponse := httptest.NewRecorder()
		router.ServeHTTP(setupConfirmResponse, setupConfirm)
		if setupConfirmResponse.Code != http.StatusOK || !strings.Contains(setupConfirmResponse.Body.String(), `"requiresFreshLogin":true`) {
			return errors.New("first-password completion did not require fresh login")
		}
		var setupState bool
		if err = f.Pool.QueryRow(ctx, `SELECT u.telephone='+14155552671' AND u.identity_version=1 AND u.email_verified_at IS NOT NULL AND c.generation=1 AND NOT EXISTS(SELECT 1 FROM auth_sessions s WHERE s.user_id=u.id) AND t.consumed_at IS NOT NULL FROM users u JOIN password_credentials c ON c.user_id=u.id JOIN email_action_tokens t ON t.user_id=u.id WHERE u.email_normalized='bo-setup@example.test'`).Scan(&setupState); err != nil {
			return err
		}
		if !setupState {
			return errors.New("setup mutation did not commit minimum account changes atomically")
		}
		unknown := requestPassword("unknown@example.test")
		if unknown.Code != setupResponse.Code || unknown.Body.String() != setupResponse.Body.String() || len(capture.Snapshot()) != 4 {
			return errors.New("unknown password request enumerated account")
		}
		stage = "password-reset-http"
		clock.at = clock.at.Add(61 * time.Second)
		resetRequest := requestPassword("bo-setup@example.test")
		resetMessages := capture.Snapshot()
		stage = fmt.Sprintf("password-reset-http-%d-capture-%d-bodymatch-%t-subjectmatch-%t", resetRequest.Code, len(resetMessages), resetRequest.Body.String() == setupResponse.Body.String(), len(resetMessages) > 4 && strings.Contains(resetMessages[4].Subject, "Atur ulang kata sandi"))
		if resetRequest.Code != setupResponse.Code || resetRequest.Body.String() != setupResponse.Body.String() || len(capture.Snapshot()) != 5 || !strings.Contains(capture.Snapshot()[4].Subject, "Atur ulang kata sandi") {
			return errors.New("password-reset request response/message mismatch")
		}
		resetMail := capture.Snapshot()[4].Text
		resetPrefix := "https://web.example.test/reset-password?"
		resetStart := strings.Index(resetMail, resetPrefix)
		if resetStart < 0 {
			return errors.New("reset action URL missing")
		}
		resetEnd := strings.Index(resetMail[resetStart:], "\n")
		if resetEnd < 0 {
			return errors.New("reset action URL malformed")
		}
		resetURL, parseErr := url.Parse(strings.TrimSpace(resetMail[resetStart : resetStart+resetEnd]))
		if parseErr != nil {
			return errors.New("reset action URL malformed")
		}
		resetCSRF := csrfFor(t, router)
		resetBody, _ := json.Marshal(map[string]string{"token": resetURL.Query().Get("token"), "password": "A New Correct Horse Battery 123"})
		resetConfirm := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/action/confirm", strings.NewReader(string(resetBody)))
		resetConfirm.Header.Set("Origin", authTestOrigin)
		resetConfirm.Header.Set("X-CSRF-Token", resetCSRF)
		resetConfirm.Header.Set("Content-Type", "application/json")
		resetConfirm.RemoteAddr = "127.0.0.1:43123"
		resetResponse := httptest.NewRecorder()
		router.ServeHTTP(resetResponse, resetConfirm)
		stage = fmt.Sprintf("password-reset-confirm-%d-fresh-%t", resetResponse.Code, strings.Contains(resetResponse.Body.String(), `"requiresFreshLogin":true`))
		if resetResponse.Code != http.StatusOK || !strings.Contains(resetResponse.Body.String(), `"requiresFreshLogin":true`) {
			return errors.New("password reset did not require a fresh login")
		}
		stage = "password-reset-state"
		var resetVerified, active bool
		var generation int64
		var changed time.Time
		var sessionsRemaining, actionsRemaining, consumed int
		if err = f.Pool.QueryRow(ctx, `SELECT u.email_verified_at IS NOT NULL,u.status='active',c.generation,c.password_changed_at,(SELECT count(*) FROM auth_sessions s WHERE s.user_id=u.id AND s.revoked_at IS NULL),(SELECT count(*) FROM email_action_tokens a WHERE a.user_id=u.id AND a.consumed_at IS NULL AND a.revoked_at IS NULL),(SELECT count(*) FROM email_action_tokens a WHERE a.user_id=u.id AND a.purpose='reset_password' AND a.consumed_at IS NOT NULL) FROM users u JOIN password_credentials c ON c.user_id=u.id WHERE u.email_normalized='bo-setup@example.test'`).Scan(&resetVerified, &active, &generation, &changed, &sessionsRemaining, &actionsRemaining, &consumed); err != nil {
			return err
		}
		if !resetVerified || !active || generation != 2 || !changed.Equal(clock.at) || sessionsRemaining != 0 || actionsRemaining != 0 || consumed != 1 {
			stage = fmt.Sprintf("password-reset-state-v%t-a%t-g%d-time%t-s%d-a%d-c%d", resetVerified, active, generation, changed.Equal(clock.at), sessionsRemaining, actionsRemaining, consumed)
			return errors.New("password reset failed revoke/generation/state contract")
		}
		stage = "reset-action-limit"
		for i := 0; i < 12; i++ {
			clock.at = clock.at.Add(61 * time.Second)
			resp := requestPassword("bo-setup@example.test")
			if resp.Code != http.StatusAccepted {
				return errors.New("action-cap request response was not generic")
			}
		}
		var liveResetActions int
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens WHERE user_id=(SELECT id FROM users WHERE email_normalized='bo-setup@example.test') AND purpose='reset_password' AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$1`, clock.at).Scan(&liveResetActions); err != nil {
			return err
		}
		if liveResetActions != 3 {
			stage = fmt.Sprintf("reset-action-limit-live-%d", liveResetActions)
			return errors.New("per-purpose live action cap not enforced")
		}
		loginRequest := func(emailAddress, password string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
			csrf := csrfFor(t, router, cookies...)
			body, _ := json.Marshal(map[string]string{"email": emailAddress, "password": password})
			r := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/login", strings.NewReader(string(body)))
			r.Header.Set("Origin", authTestOrigin)
			r.Header.Set("X-CSRF-Token", csrf)
			r.Header.Set("Content-Type", "application/json")
			r.RemoteAddr = "127.0.0.1:43123"
			for _, cookie := range cookies {
				r.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)
			return w
		}
		stage = "unverified-reset-login-resend-recovery"
		stage = "request-unverified-reset"
		resetRaceRequest := requestPassword("race@example.test")
		if resetRaceRequest.Code != http.StatusAccepted {
			return errors.New("unverified account reset request was not generically accepted")
		}
		stage = "parse-unverified-reset-link"
		resetMailText := capture.Snapshot()[len(capture.Snapshot())-1].Text
		unverifiedResetPrefix := "https://web.example.test/reset-password?"
		unverifiedResetStart := strings.Index(resetMailText, unverifiedResetPrefix)
		if unverifiedResetStart < 0 {
			return errors.New("unverified reset action link missing")
		}
		unverifiedResetEnd := strings.Index(resetMailText[unverifiedResetStart:], "\n")
		if unverifiedResetEnd < 0 {
			unverifiedResetEnd = len(resetMailText) - unverifiedResetStart
		}
		resetRaceMail, parseErr := url.Parse(strings.TrimSpace(resetMailText[unverifiedResetStart : unverifiedResetStart+unverifiedResetEnd]))
		if parseErr != nil {
			return errors.New("unverified reset action URL malformed")
		}
		resetRaceBody, _ := json.Marshal(map[string]string{"token": resetRaceMail.Query().Get("token"), "password": "A Post Reset Password For Race 123"})
		resetRaceCSRF := csrfFor(t, router)
		resetRaceConfirm := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/action/confirm", strings.NewReader(string(resetRaceBody)))
		resetRaceConfirm.Header.Set("Origin", authTestOrigin)
		resetRaceConfirm.Header.Set("X-CSRF-Token", resetRaceCSRF)
		resetRaceConfirm.Header.Set("Content-Type", "application/json")
		resetRaceConfirm.RemoteAddr = "127.0.0.1:43123"
		stage = "confirm-unverified-reset"
		resetRaceResult := httptest.NewRecorder()
		router.ServeHTTP(resetRaceResult, resetRaceConfirm)
		if resetRaceResult.Code != http.StatusOK || !strings.Contains(resetRaceResult.Body.String(), `"requiresFreshLogin":true`) {
			return errors.New("unverified reset did not require explicit login")
		}
		stage = "login-after-unverified-reset"
		recoveryLogin := loginRequest("race@example.test", "A Post Reset Password For Race 123")
		if recoveryLogin.Code != http.StatusOK || !strings.Contains(recoveryLogin.Body.String(), `"kind":"restricted"`) || !strings.Contains(recoveryLogin.Body.String(), `"emailVerified":false`) {
			return errors.New("reset of unverified account bypassed restricted login policy")
		}
		var recoverySession, recoveryReceipt *http.Cookie
		for _, cookie := range recoveryLogin.Result().Cookies() {
			if cookie.Name == "__Host-zenith_auth" {
				recoverySession = cookie
			}
			if cookie.Name == "__Host-zenith_enrollment" {
				recoveryReceipt = cookie
			}
		}
		if recoverySession == nil || recoveryReceipt == nil {
			return errors.New("restricted post-reset login did not issue fresh enrollment binding")
		}
		stage = "recovery-session-resend"
		recoveryCSRF := csrfFor(t, router, recoverySession, recoveryReceipt)
		beforeResendMessages := len(capture.Snapshot())
		recoveryResend := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/verification/resend", strings.NewReader(`{}`))
		recoveryResend.Header.Set("Origin", authTestOrigin)
		recoveryResend.Header.Set("X-CSRF-Token", recoveryCSRF)
		recoveryResend.Header.Set("Content-Type", "application/json")
		recoveryResend.RemoteAddr = "127.0.0.1:43123"
		recoveryResend.AddCookie(recoverySession)
		recoveryResend.AddCookie(recoveryReceipt)
		recoveryResendResult := httptest.NewRecorder()
		router.ServeHTTP(recoveryResendResult, recoveryResend)
		if recoveryResendResult.Code != http.StatusAccepted || len(capture.Snapshot()) != beforeResendMessages+1 {
			stage = fmt.Sprintf("recovery-resend-status-%d-messages-%d", recoveryResendResult.Code, len(capture.Snapshot())-beforeResendMessages)
			return errors.New("post-reset restricted login could not safely request new verification")
		}
		var latestGeneration int64
		if err = f.Pool.QueryRow(ctx, `SELECT credential_generation FROM email_action_tokens WHERE user_id=(SELECT id FROM users WHERE email_normalized='race@example.test') AND purpose='verify_email' ORDER BY created_at DESC LIMIT 1`).Scan(&latestGeneration); err != nil {
			return err
		}
		if latestGeneration != 2 {
			return errors.New("post-reset verification action did not bind the new credential generation")
		}
		stage = "customer-login-current-logout"
		stage = "restricted-login"
		restrictedLogin := recoveryLogin
		if restrictedLogin.Code != http.StatusOK || !strings.Contains(restrictedLogin.Body.String(), `"kind":"restricted"`) || !strings.Contains(restrictedLogin.Body.String(), `"emailVerified":false`) {
			stage = fmt.Sprintf("restricted-login-status-%d", restrictedLogin.Code)
			return errors.New("unverified password proof did not receive restricted-only session")
		}
		var restrictedCookie, restrictedEnrollment *http.Cookie
		for _, cookie := range restrictedLogin.Result().Cookies() {
			if cookie.Name == "__Host-zenith_auth" {
				restrictedCookie = cookie
			}
			if cookie.Name == "__Host-zenith_enrollment" {
				restrictedEnrollment = cookie
			}
		}
		if restrictedCookie == nil || restrictedEnrollment == nil || restrictedCookie.MaxAge != int(identity.RestrictedSessionLifetime.Seconds()) {
			stage = fmt.Sprintf("restricted-login-cookies-auth-%t-enrollment-%t-maxage-%d", restrictedCookie != nil, restrictedEnrollment != nil, func() int {
				if restrictedCookie == nil {
					return -2
				}
				return restrictedCookie.MaxAge
			}())
			return errors.New("restricted login cookie did not preserve short lifetime")
		}
		stage = "restricted-current-read"
		restrictedCurrent := httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/session/current", nil)
		restrictedCurrent.Header.Set("Origin", authTestOrigin)
		restrictedCurrent.AddCookie(restrictedCookie)
		restrictedCurrentResponse := httptest.NewRecorder()
		router.ServeHTTP(restrictedCurrentResponse, restrictedCurrent)
		if restrictedCurrentResponse.Code != http.StatusOK || !strings.Contains(restrictedCurrentResponse.Body.String(), `"kind":"restricted"`) {
			stage = fmt.Sprintf("restricted-current-status-%d", restrictedCurrentResponse.Code)
			return errors.New("restricted session current view unavailable")
		}
		var restrictedBody struct {
			CSRFToken string `json:"csrfToken"`
		}
		if json.Unmarshal(restrictedLogin.Body.Bytes(), &restrictedBody) != nil || restrictedBody.CSRFToken == "" {
			return errors.New("restricted login did not return memory-only CSRF")
		}
		stage = "restricted-touch-denied"
		restrictedTouch := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/session/touch", strings.NewReader(`{}`))
		restrictedTouch.Header.Set("Origin", authTestOrigin)
		restrictedTouch.Header.Set("X-CSRF-Token", restrictedBody.CSRFToken)
		restrictedTouch.Header.Set("Content-Type", "application/json")
		restrictedTouch.RemoteAddr = "127.0.0.1:43123"
		restrictedTouch.AddCookie(restrictedCookie)
		restrictedTouchResponse := httptest.NewRecorder()
		router.ServeHTTP(restrictedTouchResponse, restrictedTouch)
		if restrictedTouchResponse.Code != http.StatusUnauthorized {
			stage = fmt.Sprintf("restricted-touch-status-%d", restrictedTouchResponse.Code)
			return errors.New("restricted session acquired full-session foreground renewal")
		}
		stage = "restricted-verification-resend"
		clock.at = clock.at.Add(61 * time.Second)
		var verifyCountBefore int
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens WHERE target_email_normalized='race@example.test' AND purpose='verify_email' AND consumed_at IS NULL AND revoked_at IS NULL`).Scan(&verifyCountBefore); err != nil {
			return err
		}
		resendToken := csrfFor(t, router, restrictedCookie, restrictedEnrollment)
		resendRestricted := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/verification/resend", strings.NewReader(`{}`))
		resendRestricted.Header.Set("Origin", authTestOrigin)
		resendRestricted.Header.Set("X-CSRF-Token", resendToken)
		resendRestricted.Header.Set("Content-Type", "application/json")
		resendRestricted.RemoteAddr = "127.0.0.1:43123"
		resendRestricted.AddCookie(restrictedCookie)
		resendRestricted.AddCookie(restrictedEnrollment)
		resendRestrictedResponse := httptest.NewRecorder()
		router.ServeHTTP(resendRestrictedResponse, resendRestricted)
		var verifyCountAfter int
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens WHERE target_email_normalized='race@example.test' AND purpose='verify_email' AND consumed_at IS NULL AND revoked_at IS NULL`).Scan(&verifyCountAfter); err != nil {
			return err
		}
		if resendRestrictedResponse.Code != http.StatusAccepted || verifyCountAfter != verifyCountBefore+1 {
			return errors.New("restricted login receipt could not authorize one explicit verification resend")
		}
		stage = "full-login"
		if _, err := f.Pool.Exec(ctx, `CREATE FUNCTION auth_test_fail_login_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action IN ('identity.login.full','identity.login.restricted') THEN RAISE EXCEPTION 'fixture login audit rejection' USING ERRCODE='XX000'; END IF; RETURN NEW; END $$`); err != nil {
			return err
		}
		if _, err := f.Pool.Exec(ctx, `CREATE TRIGGER auth_test_fail_login_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION auth_test_fail_login_audit()`); err != nil {
			return err
		}
		failedLogin := loginRequest("bo-setup@example.test", "A New Correct Horse Battery 123")
		if failedLogin.Code != http.StatusServiceUnavailable {
			return errors.New("login audit failure did not return safe unavailable")
		}
		var failedLoginRows int
		if err = f.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM auth_sessions WHERE user_id=(SELECT id FROM users WHERE email_normalized='bo-setup@example.test'))+(SELECT count(*) FROM audit_events WHERE action LIKE 'identity.login.%' AND target_id=(SELECT id::text FROM users WHERE email_normalized='bo-setup@example.test'))`).Scan(&failedLoginRows); err != nil {
			return err
		}
		if failedLoginRows != 0 {
			return errors.New("login audit failure committed session or success audit")
		}
		if _, err = f.Pool.Exec(ctx, `DROP TRIGGER auth_test_fail_login_audit ON audit_events`); err != nil {
			return err
		}
		if _, err = f.Pool.Exec(ctx, `DROP FUNCTION auth_test_fail_login_audit()`); err != nil {
			return err
		}
		login := loginRequest("bo-setup@example.test", "A New Correct Horse Battery 123")
		if login.Code != http.StatusOK || !strings.Contains(login.Body.String(), `"kind":"full"`) || !strings.Contains(login.Body.String(), `"emailVerified":true`) {
			stage = fmt.Sprintf("full-login-status-%d", login.Code)
			return errors.New("valid customer password did not issue verified full session")
		}
		var loginBody struct {
			CSRFToken string `json:"csrfToken"`
			Email     string `json:"email"`
		}
		if json.Unmarshal(login.Body.Bytes(), &loginBody) != nil || loginBody.Email != "bo-setup@example.test" || len(loginBody.CSRFToken) != 96 {
			return errors.New("login response identity/CSRF shape invalid")
		}
		var loginCookie *http.Cookie
		for _, cookie := range login.Result().Cookies() {
			if cookie.Name == "__Host-zenith_auth" {
				loginCookie = cookie
			}
		}
		if loginCookie == nil || !loginCookie.Secure || !loginCookie.HttpOnly || loginCookie.MaxAge != int(identity.SessionIdleLifetime.Seconds()) {
			return errors.New("full login cookie flags/deadline incorrect")
		}
		var initialLogin struct {
			CSRFToken string `json:"csrfToken"`
		}
		if json.Unmarshal(login.Body.Bytes(), &initialLogin) != nil || initialLogin.CSRFToken == "" {
			return errors.New("login CSRF bootstrap missing")
		}
		stage = "same-cookie-login-rotation"
		secondLogin := loginRequest("bo-setup@example.test", "A New Correct Horse Battery 123", loginCookie)
		if secondLogin.Code != http.StatusOK {
			stage = fmt.Sprintf("same-cookie-login-status-%d", secondLogin.Code)
			return errors.New("session-bound password proof did not rotate browser session")
		}
		var secondLoginBody struct {
			CSRFToken string `json:"csrfToken"`
		}
		if json.Unmarshal(secondLogin.Body.Bytes(), &secondLoginBody) != nil || secondLoginBody.CSRFToken == "" || secondLoginBody.CSRFToken == initialLogin.CSRFToken {
			return errors.New("login did not rotate session-bound CSRF token")
		}
		var rotatedLoginCookie *http.Cookie
		for _, cookie := range secondLogin.Result().Cookies() {
			if cookie.Name == "__Host-zenith_auth" {
				rotatedLoginCookie = cookie
			}
		}
		if rotatedLoginCookie == nil || rotatedLoginCookie.Value == loginCookie.Value {
			return errors.New("login reused the prior browser credential")
		}
		stage = "old-login-rejected"
		currentRequest := httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/session/current", nil)
		currentRequest.Header.Set("Origin", authTestOrigin)
		currentRequest.AddCookie(loginCookie)
		currentResponse := httptest.NewRecorder()
		router.ServeHTTP(currentResponse, currentRequest)
		if currentResponse.Code != http.StatusUnauthorized {
			return errors.New("login left prior same-browser session active")
		}
		stage = "current-full-session"
		currentRequest = httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/session/current", nil)
		currentRequest.Header.Set("Origin", authTestOrigin)
		currentRequest.AddCookie(rotatedLoginCookie)
		currentResponse = httptest.NewRecorder()
		router.ServeHTTP(currentResponse, currentRequest)
		if currentResponse.Code != http.StatusOK || !strings.Contains(currentResponse.Body.String(), `"email":"bo-setup@example.test"`) {
			stage = fmt.Sprintf("current-session-status-%d", currentResponse.Code)
			return errors.New("current-session read did not reflect full session")
		}
		var currentBeforeTouch struct {
			ExpiresAt string `json:"expiresAt"`
		}
		if json.Unmarshal(currentResponse.Body.Bytes(), &currentBeforeTouch) != nil {
			return errors.New("current deadline response malformed")
		}
		previousDeadline, parseErr := time.Parse(time.RFC3339Nano, currentBeforeTouch.ExpiresAt)
		if parseErr != nil {
			return errors.New("current deadline malformed")
		}
		clock.at = clock.at.Add(61 * time.Second)
		stage = "full-session-touch"
		touchRequest := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/session/touch", strings.NewReader(`{}`))
		touchRequest.Header.Set("Origin", authTestOrigin)
		touchRequest.Header.Set("X-CSRF-Token", secondLoginBody.CSRFToken)
		touchRequest.Header.Set("Content-Type", "application/json")
		touchRequest.RemoteAddr = "127.0.0.1:43123"
		touchRequest.AddCookie(rotatedLoginCookie)
		touchResponse := httptest.NewRecorder()
		router.ServeHTTP(touchResponse, touchRequest)
		var touched *http.Cookie
		for _, cookie := range touchResponse.Result().Cookies() {
			if cookie.Name == "__Host-zenith_auth" {
				touched = cookie
			}
		}
		var touchBody struct {
			ExpiresAt string `json:"expiresAt"`
		}
		if touchResponse.Code != http.StatusOK || touched == nil || touched.Value != rotatedLoginCookie.Value || json.Unmarshal(touchResponse.Body.Bytes(), &touchBody) != nil || touchBody.ExpiresAt == "" {
			stage = fmt.Sprintf("touch-status-%d", touchResponse.Code)
			return errors.New("foreground touch response/cookie invalid")
		}
		stage = "touch-deadline-assertion"
		touchedDeadline, parseErr := time.Parse(time.RFC3339Nano, touchBody.ExpiresAt)
		if parseErr != nil || !touchedDeadline.Equal(clock.at.Add(identity.SessionIdleLifetime)) || !touchedDeadline.After(previousDeadline) || touched.MaxAge != int(identity.SessionIdleLifetime.Seconds()) {
			stage = fmt.Sprintf("touch-expiry-parse-%t-equal-%t-after-%t-maxage-%d", parseErr == nil, parseErr == nil && touchedDeadline.Equal(clock.at.Add(identity.SessionIdleLifetime)), parseErr == nil && touchedDeadline.After(previousDeadline), touched.MaxAge)
			return errors.New("foreground touch did not extend only the same live session deadline")
		}
		stage = "touch-throttle-assertion"
		clock.at = clock.at.Add(10 * time.Second)
		touchAgain := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/session/touch", strings.NewReader(`{}`))
		touchAgain.Header.Set("Origin", authTestOrigin)
		touchAgain.Header.Set("X-CSRF-Token", secondLoginBody.CSRFToken)
		touchAgain.Header.Set("Content-Type", "application/json")
		touchAgain.RemoteAddr = "127.0.0.1:43123"
		touchAgain.AddCookie(rotatedLoginCookie)
		touchAgainResponse := httptest.NewRecorder()
		router.ServeHTTP(touchAgainResponse, touchAgain)
		var throttledBody struct {
			ExpiresAt string `json:"expiresAt"`
		}
		var throttledCookie *http.Cookie
		for _, cookie := range touchAgainResponse.Result().Cookies() {
			if cookie.Name == "__Host-zenith_auth" {
				throttledCookie = cookie
			}
		}
		if touchAgainResponse.Code != http.StatusOK || json.Unmarshal(touchAgainResponse.Body.Bytes(), &throttledBody) != nil || throttledBody.ExpiresAt != touchBody.ExpiresAt || throttledCookie == nil || throttledCookie.MaxAge >= int(identity.SessionIdleLifetime.Seconds()) {
			stage = fmt.Sprintf("touch-throttle-status-%d-same-%t-maxage-%d", touchAgainResponse.Code, throttledBody.ExpiresAt == touchBody.ExpiresAt, func() int {
				if throttledCookie == nil {
					return -2
				}
				return throttledCookie.MaxAge
			}())
			return errors.New("touch less than one minute apart extended the logical session again")
		}
		logoutCSRF := secondLoginBody.CSRFToken
		stage = "stale-csrf-rejected"
		staleLogout := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/session/logout", strings.NewReader(`{}`))
		staleLogout.Header.Set("Origin", authTestOrigin)
		staleLogout.Header.Set("X-CSRF-Token", initialLogin.CSRFToken)
		staleLogout.Header.Set("Content-Type", "application/json")
		staleLogout.RemoteAddr = "127.0.0.1:43123"
		staleLogout.AddCookie(rotatedLoginCookie)
		staleLogoutResponse := httptest.NewRecorder()
		router.ServeHTTP(staleLogoutResponse, staleLogout)
		if staleLogoutResponse.Code != http.StatusForbidden {
			return errors.New("login accepted the prior session's CSRF token")
		}
		stage = "logout"
		logoutRequest := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/session/logout", strings.NewReader(`{}`))
		logoutRequest.Header.Set("Origin", authTestOrigin)
		logoutRequest.Header.Set("X-CSRF-Token", logoutCSRF)
		logoutRequest.Header.Set("Content-Type", "application/json")
		logoutRequest.RemoteAddr = "127.0.0.1:43123"
		logoutRequest.AddCookie(rotatedLoginCookie)
		logoutResponse := httptest.NewRecorder()
		router.ServeHTTP(logoutResponse, logoutRequest)
		if logoutResponse.Code != http.StatusOK || !strings.Contains(logoutResponse.Body.String(), `"loggedOut":true`) {
			return errors.New("logout did not confirm session revocation")
		}
		currentRequest = httptest.NewRequest(http.MethodGet, authTestOrigin+"/v1/auth/session/current", nil)
		currentRequest.Header.Set("Origin", authTestOrigin)
		currentRequest.AddCookie(rotatedLoginCookie)
		currentResponse = httptest.NewRecorder()
		router.ServeHTTP(currentResponse, currentRequest)
		if currentResponse.Code != http.StatusUnauthorized {
			return errors.New("revoked session remained usable for current identity")
		}
		stage = "unknown-disabled-login-generic"
		unknownLogin := loginRequest("unknown-login@example.test", "A New Correct Horse Battery 123")
		disabledLogin := loginRequest("disabled-login@example.test", "A New Correct Horse Battery 123")
		noCredentialLogin := loginRequest("no-password@example.test", "A New Correct Horse Battery 123")
		badPasswordLogin := loginRequest("bo-setup@example.test", "A Different Correct Horse Password")
		if unknownLogin.Code != http.StatusUnauthorized || unknownLogin.Body.String() != disabledLogin.Body.String() || unknownLogin.Body.String() != noCredentialLogin.Body.String() || unknownLogin.Body.String() != badPasswordLogin.Body.String() {
			return errors.New("unknown/disabled/no-credential/bad-password login response enumerated account state")
		}
		stage = "registration-body-limit"
		oversized := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/register", strings.NewReader(strings.Repeat("x", 16*1024+1)))
		oversized.Header.Set("Origin", authTestOrigin)
		oversized.Header.Set("X-CSRF-Token", csrfConcurrent)
		oversized.Header.Set("Content-Type", "application/json")
		oversized.RemoteAddr = "127.0.0.1:43123"
		largeResponse := httptest.NewRecorder()
		router.ServeHTTP(largeResponse, oversized)
		if largeResponse.Code != http.StatusRequestEntityTooLarge {
			return errors.New("registration accepted body above 16 KiB")
		}
		if _, err := f.Pool.Exec(ctx, `CREATE FUNCTION auth_test_fail_register_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action='identity.register' THEN RAISE EXCEPTION 'fixture audit rejection' USING ERRCODE='XX000'; END IF; RETURN NEW; END $$`); err != nil {
			return err
		}
		if _, err := f.Pool.Exec(ctx, `CREATE TRIGGER auth_test_fail_register_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION auth_test_fail_register_audit()`); err != nil {
			return err
		}
		stage = "atomic-audit-rejection"
		token = csrfFor(t, router)
		failed := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/register", strings.NewReader(`{"displayName":"Rollback","email":"rollback@example.test","telephone":"+14155552671","password":"Correct Horse Battery Staple"}`))
		failed.Header.Set("Origin", authTestOrigin)
		failed.Header.Set("X-CSRF-Token", token)
		failed.Header.Set("Content-Type", "application/json")
		failed.RemoteAddr = "127.0.0.1:43123"
		failedResponse := httptest.NewRecorder()
		router.ServeHTTP(failedResponse, failed)
		if failedResponse.Code != http.StatusServiceUnavailable || len(failedResponse.Result().Cookies()) != 0 {
			return errors.New("audit rollback response leaked a successful registration")
		}
		var persisted int
		if err := f.Pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE email_normalized='rollback@example.test'`).Scan(&persisted); err != nil {
			return err
		}
		if persisted != 0 {
			return errors.New("audit failure committed user")
		}
		if _, err := f.Pool.Exec(ctx, `DROP TRIGGER auth_test_fail_register_audit ON audit_events`); err != nil {
			return err
		}
		if _, err := f.Pool.Exec(ctx, `DROP FUNCTION auth_test_fail_register_audit()`); err != nil {
			return err
		}
		stage = "origin-rejection"
		badOrigin := httptest.NewRequest(http.MethodPost, authTestOrigin+"/v1/auth/register", strings.NewReader(`{}`))
		badOrigin.Header.Set("Origin", "https://evil.example.test")
		badOrigin.Header.Set("X-CSRF-Token", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, badOrigin)
		if w.Code != http.StatusForbidden {
			return errors.New("cross-origin registration accepted")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("owned in-process Auth HTTP registration failed at %s; details withheld", stage)
	}
}
