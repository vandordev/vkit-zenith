//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/handlers/v1"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/server"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
	"github.com/vandordev/vkit-zenith/packages/tools/authtest"
)

const orgTestOrigin = "https://web.example.test"

var orgSecretKey = regexp.MustCompile(`(?i)(hash|ciphertext|secret|token|password|objectkey|^pan$|cardnumber|cvv)`)

func assertNoSecretKeys(t *testing.T, body []byte) {
	t.Helper()
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatal("response was not JSON")
	}
	var walk func(any)
	walk = func(value any) {
		switch value := value.(type) {
		case map[string]any:
			for key, nested := range value {
				if orgSecretKey.MatchString(key) && key != "hasCredentials" && key != "prefix" && key != "tokenPrefix" {
					t.Errorf("response contains secret-like field %q", key)
				}
				walk(nested)
			}
		case []any:
			for _, nested := range value {
				walk(nested)
			}
		}
	}
	walk(value)
}

type orgRecorder struct {
	router   *echo.Echo
	runtime  *identityhttp.Runtime
	pool     *pgxpool.Pool
	service  *identity.Service
	boundary *orgaccess.Boundary
	api      huma.API
	token    string
	csrf     string
	userID   string
	orgID    string
}

var orgFixture *authtest.Fixture

func TestMain(m *testing.M) {
	if listed := flag.Lookup("test.list"); listed != nil && listed.Value.String() != "" {
		os.Exit(m.Run())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	code := 1
	err := authtest.WithFixture(ctx, func(_ context.Context, fixture *authtest.Fixture) error {
		orgFixture = fixture
		code = m.Run()
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "organization fixture or cleanup failed")
		code = 1
	}
	os.Exit(code)
}

func newOrgRecorder(t *testing.T, fixture *authtest.Fixture) *orgRecorder {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Now().UTC().Truncate(time.Microsecond)
	userID, orgID, sessionID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sessionSecret := uuid.NewString()
	secretHash := identity.HashSecret(sessionSecret)
	for _, seed := range []struct {
		name  string
		query string
		args  []any
	}{
		{"user", `INSERT INTO users(id,email,email_normalized,display_name,status,email_verified_at,created_at,updated_at,identity_version) VALUES($1,$2,$2,'Member Test','active',$3,$3,$3,1)`, []any{userID, "member-" + userID + "@example.test", now}},
		{"credential", `INSERT INTO password_credentials(user_id,password_hash,password_changed_at,generation) VALUES($1,'$argon2id$v=19$m=65536,t=3,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',$2,1)`, []any{userID, now}},
		{"session", `INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,created_at,kind,identity_version,credential_generation,last_activity_at,browser_cookie_expires_at) VALUES($3,$1,$4,$2::timestamptz,$2::timestamptz,$2::timestamptz+interval '1176 hours',$2::timestamptz,'full',1,1,$2::timestamptz,$2::timestamptz+interval '48 hours')`, []any{userID, now, sessionID, secretHash[:]}},
	} {
		if _, err := fixture.Pool.Exec(ctx, seed.query, seed.args...); err != nil {
			var databaseError *pgconn.PgError
			if errors.As(err, &databaseError) {
				t.Fatalf("organization fixture %s seed failed (SQLSTATE %s, constraint %s)", seed.name, databaseError.Code, databaseError.ConstraintName)
			}
			t.Fatalf("organization fixture %s seed failed", seed.name)
		}
	}
	tx, err := fixture.Pool.Begin(ctx)
	if err != nil {
		t.Fatal("organization organization transaction unavailable")
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO organizations(id,name,slug,status,created_at,updated_at) VALUES($1,'Organization Example',$2,'active',$3,$3)`, orgID, "org-"+strings.ReplaceAll(orgID, "-", ""), now); err != nil {
		t.Fatal("organization organization seed failed")
	}
	if _, err := tx.Exec(ctx, `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,'active','owner',$4,$4)`, uuid.NewString(), orgID, userID, now); err != nil {
		t.Fatal("organization membership seed failed")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal("organization owner invariant seed failed")
	}
	service := identity.NewService(fixture.Pool, orgTestClock{now}, orgTestSender{}, [32]byte{1}, [32]byte{2}, orgTestOrigin, false)
	csrf, err := service.NewSessionCSRF(secretHash)
	if err != nil {
		t.Fatal("organization CSRF fixture failed")
	}
	runtime := identityhttp.NewRuntime(service, orgTestOrigin, false)
	e := server.NewHTTP(echoLogger(), orgTestOrigin)
	api := server.NewAPI(e)
	queries := persistence.New(fixture.Pool)
	boundary := orgaccess.NewBoundary(runtime, queries)
	for path, permission := range map[string]customers.Permission{"read": customers.PermissionOrganizationRead, "developers": customers.PermissionDevelopersManage, "manage": customers.PermissionOrganizationManage, "members": customers.PermissionMembersManage} {
		path, permission := path, permission
		operation := huma.Operation{OperationID: "testPermission" + path, Method: http.MethodGet, Path: "/v1/organizations/{organizationId}/permission/" + path, Errors: []int{401, 403, 404, 503}}
		huma.Register(api, operation, func(ctx context.Context, in *testOrgPath) (*testOutput, error) {
			if _, err := boundary.Read(ctx, in.OrganizationID, permission); err != nil {
				return nil, err
			}
			return &testOutput{}, nil
		})
	}
	huma.Register(api, huma.Operation{OperationID: "testOrganizationWrite", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/write", Errors: []int{401, 403, 404, 409, 503}}, func(ctx context.Context, in *testWriteInput) (*testOutput, error) {
		if _, err := boundary.Write(ctx, in.OrganizationID, customers.PermissionOrganizationRead, orgaccess.WriteHeaders{Origin: in.Origin, CSRF: in.CSRF}); err != nil {
			return nil, err
		}
		return &testOutput{}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "testSensitiveValidation", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/validation", Errors: []int{422}}, func(ctx context.Context, in *testSensitiveInput) (*testOutput, error) {
		return &testOutput{}, nil
	})
	for _, path := range []string{"/v1/auth/body-policy", "/v1/body-policy"} {
		path := path
		huma.Register(api, huma.Operation{OperationID: "testBodyPolicy" + strings.ReplaceAll(strings.TrimPrefix(path, "/v1/"), "/", ""), Method: http.MethodPost, Path: path}, func(ctx context.Context, in *testRawBodyInput) (*testOutput, error) {
			return &testOutput{}, nil
		})
	}
	huma.Register(api, huma.Operation{OperationID: "testBodyPolicyShortTimeout", Method: http.MethodPost, Path: "/v1/body-policy-short", BodyReadTimeout: 250 * time.Millisecond}, func(ctx context.Context, in *testRawBodyInput) (*testOutput, error) {
		return &testOutput{}, nil
	})
	if err := v1.NewListMyOrganizations(api, boundary, queries); err != nil {
		t.Fatal("organization list route registration failed")
	}
	if err := v1.NewGetOrganizationProfile(api, boundary); err != nil {
		t.Fatal("organization profile route registration failed")
	}
	if err := v1.NewListOrganizationMembers(api, boundary, queries); err != nil {
		t.Fatal("organization members route registration failed")
	}
	for _, register := range []func(huma.API, *orgaccess.Boundary) error{v1.NewListAPIKeys, v1.NewCreateAPIKey, v1.NewRevokeAPIKey, v1.NewUploadLogo} {
		if err := register(api, boundary); err != nil {
			t.Fatal("settings route registration failed")
		}
	}
	return &orgRecorder{router: e, runtime: runtime, pool: fixture.Pool, service: service, boundary: boundary, api: api, token: sessionSecret, csrf: csrf, userID: userID, orgID: orgID}
}

type testOrgPath struct {
	OrganizationID string `path:"organizationId"`
}
type testWriteInput struct {
	OrganizationID string `path:"organizationId"`
	Origin         string `header:"Origin"`
	CSRF           string `header:"X-CSRF-Token"`
}
type testOutput struct {
	Body struct {
		OK bool `json:"ok"`
	}
}
type testSensitiveInput struct {
	OrganizationID string `path:"organizationId"`
	Body           struct {
		Password string `json:"password" maxLength:"5"`
	}
}

type testRawBodyInput struct{ Body json.RawMessage }

type orgTestClock struct{ now time.Time }

func (c orgTestClock) Now() time.Time { return c.now }

type orgTestSender struct{}

func (orgTestSender) Send(context.Context, identity.ActionDelivery) error { return nil }

func echoLogger() *slog.Logger { return slog.New(slog.NewJSONHandler(io.Discard, nil)) }

func (r *orgRecorder) request(method, path, origin, csrf string) *httptest.ResponseRecorder {
	return r.requestBody(method, path, origin, csrf, "")
}

func (r *orgRecorder) requestBody(method, path, origin, csrf, body string) *httptest.ResponseRecorder {
	return r.requestBodyContentType(method, path, origin, csrf, body, echo.MIMEApplicationJSON)
}

func (r *orgRecorder) requestBodyContentType(method, path, origin, csrf, body, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, orgTestOrigin+path, strings.NewReader(body))
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if body != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	}
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	req.AddCookie(&http.Cookie{Name: "__Host-zenith_auth", Value: r.token})
	resp := httptest.NewRecorder()
	r.router.ServeHTTP(resp, req)
	return resp
}

func (r *orgRecorder) requestStream(method, path, origin, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, orgTestOrigin+path, nil)
	req.Body = io.NopCloser(strings.NewReader(body))
	req.ContentLength = -1
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, origin)
	req.AddCookie(&http.Cookie{Name: "__Host-zenith_auth", Value: r.token})
	resp := httptest.NewRecorder()
	r.router.ServeHTTP(resp, req)
	return resp
}

func (r *orgRecorder) requestDeadlineTimeout(method, path, origin string) (*httptest.ResponseRecorder, []time.Time) {
	req := httptest.NewRequest(method, orgTestOrigin+path, nil)
	req.ContentLength = -1
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderOrigin, origin)
	req.AddCookie(&http.Cookie{Name: "__Host-zenith_auth", Value: r.token})
	resp := httptest.NewRecorder()
	writer := &deadlineRecorder{ResponseRecorder: resp}
	req.Body = io.NopCloser(deadlineTimeoutReader{writer: writer})
	r.router.ServeHTTP(writer, req)
	return resp, writer.deadlines
}

type deadlineTimeoutReader struct{ writer *deadlineRecorder }

func (r deadlineTimeoutReader) Read([]byte) (int, error) {
	if len(r.writer.deadlines) == 0 || r.writer.deadlines[0].IsZero() {
		return 0, errors.New("read deadline was not installed before reading")
	}
	return 0, os.ErrDeadlineExceeded
}

type deadlineRecorder struct {
	*httptest.ResponseRecorder
	deadlines []time.Time
}

func (w *deadlineRecorder) Unwrap() http.ResponseWriter { return w.ResponseRecorder }
func (w *deadlineRecorder) SetReadDeadline(deadline time.Time) error {
	w.deadlines = append(w.deadlines, deadline)
	return nil
}
