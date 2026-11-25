package security

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

type Boundary struct {
	Service        *staff.Service
	Pool           *pgxpool.Pool
	Clock          func() time.Time
	Origin         string
	Development    bool
	Secret         []byte
	TrustedProxies []netip.Prefix
}

type requestKey struct{}
type requestState struct {
	actor    staff.Actor
	response http.ResponseWriter
	hash     [32]byte
	boundary *Boundary
}

// Huma's default validation details can echo submitted secret values. Retain
// field locations only from this fixed public DTO allowlist, never raw input.
func init() {
	huma.NewError = func(status int, _ string, errs ...error) huma.StatusError {
		p := Problem(status)
		if status == 422 {
			for _, err := range errs {
				var d huma.ErrorDetailer
				if errors.As(err, &d) {
					loc := d.ErrorDetail().Location
					switch loc {
					case "body.email", "body.password", "body.currentPassword", "body.newPassword", "body.displayName", "body.role", "body.status", "body.revision", "path.id", "query.page", "query.pageSize", "query.q":
					case "body.key", "body.name", "body.kind", "body.description", "body.productId", "body.planVersionId", "body.features", "body.permissionIds", "body.tier", "body.amount", "body.durationMonths", "body.currencyCode", "path.productId", "path.planId", "query.productId", "query.planVersionId", "query.status":
					case "query.from", "query.until", "query.actorStaffId", "query.action", "query.outcome", "query.selectedId", "query.kind":
					case "body.reason", "body.slug", "body.ownerUserId", "body.newOwner.email", "body.newOwner.displayName", "body.userId", "path.organizationId", "query.organizationId", "query.verification", "query.ownership":
					default:
						loc = "body"
					}
					p.Errors = append(p.Errors, &huma.ErrorDetail{Location: loc, Message: "Isian tidak valid."})
				}
			}
		}
		return p
	}
}

func Actor(ctx context.Context) (staff.Actor, error) {
	v, ok := ctx.Value(requestKey{}).(requestState)
	if !ok || !v.actor.StaffID.Valid {
		return staff.Actor{}, staff.ErrUnauthenticated
	}
	return v.actor, nil
}
func Source(ctx context.Context) string {
	v, _ := ctx.Value(requestKey{}).(requestState)
	return v.actor.Source
}
func (b *Boundary) Read(ctx context.Context, role string) (staff.Staff, error) {
	a, err := Actor(ctx)
	if err != nil {
		return staff.Staff{}, err
	}
	current, err := b.Service.Authenticate(ctx, a.SessionHash)
	if err != nil {
		return staff.Staff{}, err
	}
	if role == "superadmin" && current.Role != "superadmin" {
		return staff.Staff{}, staff.ErrDenied
	}
	return current, nil
}
func (b *Boundary) SessionName() string {
	if b.Development {
		return "zenith_bo_dev_session"
	}
	return "__Host-zenith_bo_session"
}
func (b *Boundary) CSRFName() string {
	if b.Development {
		return "zenith_bo_dev_csrf"
	}
	return "__Host-zenith_bo_csrf"
}
func (b *Boundary) cookie(name, value string, expires time.Time, maxAge int) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Path: "/", HttpOnly: true, Secure: !b.Development, SameSite: http.SameSiteStrictMode, MaxAge: maxAge, Expires: expires.UTC()}
}
func (b *Boundary) Clear(w http.ResponseWriter) {
	for _, name := range []string{b.SessionName(), b.CSRFName()} {
		http.SetCookie(w, b.cookie(name, "", time.Unix(1, 0), -1))
	}
}
func (b *Boundary) SetSession(ctx context.Context, token string, expires time.Time) error {
	v, ok := ctx.Value(requestKey{}).(requestState)
	if !ok {
		return staff.ErrUnavailable
	}
	http.SetCookie(v.response, b.cookie(b.SessionName(), token, expires, int(staff.SessionLifetime.Seconds())))
	http.SetCookie(v.response, b.cookie(b.CSRFName(), "", time.Unix(1, 0), -1))
	return nil
}
func (b *Boundary) ClearSession(ctx context.Context) {
	if v, ok := ctx.Value(requestKey{}).(requestState); ok {
		b.Clear(v.response)
	}
}
func (b *Boundary) CSRF(ctx context.Context) (string, error) {
	v, ok := ctx.Value(requestKey{}).(requestState)
	if !ok {
		return "", staff.ErrUnavailable
	}
	token, err := b.issue(v.hash)
	if err != nil {
		return "", err
	}
	http.SetCookie(v.response, b.cookie(b.CSRFName(), token, b.Clock().Add(30*time.Minute), 1800))
	return token, nil
}

func (b *Boundary) issue(binding [32]byte) (string, error) {
	if len(b.Secret) < 32 || b.Clock == nil {
		return "", staff.ErrUnavailable
	}
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", staff.ErrUnavailable
	}
	payload := "1." + strconv.FormatInt(b.Clock().Add(30*time.Minute).Unix(), 10) + "." + base64.RawURLEncoding.EncodeToString(nonce) + "." + hex.EncodeToString(binding[:])
	mac := hmac.New(sha256.New, b.Secret)
	_, _ = mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
func (b *Boundary) valid(token string, binding [32]byte) bool {
	if len(token) > 256 || len(b.Secret) < 32 || b.Clock == nil {
		return false
	}
	p := strings.Split(token, ".")
	if len(p) != 5 || p[0] != "1" || len(p[2]) != 43 || len(p[3]) != 64 || len(p[4]) != 43 {
		return false
	}
	exp, err := strconv.ParseInt(p[1], 10, 64)
	now := b.Clock().Unix()
	if err != nil || exp <= now || exp > now+1800 {
		return false
	}
	nonce, err := base64.RawURLEncoding.Strict().DecodeString(p[2])
	if err != nil || len(nonce) != 32 {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(p[3]), []byte(hex.EncodeToString(binding[:]))) != 1 {
		return false
	}
	sig, err := base64.RawURLEncoding.Strict().DecodeString(p[4])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, b.Secret)
	_, _ = mac.Write([]byte(strings.Join(p[:4], ".")))
	return hmac.Equal(sig, mac.Sum(nil))
}

func uniqueCookie(r *http.Request, name string) (string, error) {
	value := ""
	found := false
	for _, c := range r.Cookies() {
		if c.Name == name {
			if found || c.Value == "" {
				return "", staff.ErrUnauthenticated
			}
			found = true
			value = c.Value
		}
	}
	return value, nil
}
func (b *Boundary) trusted(a netip.Addr) bool {
	for _, p := range b.TrustedProxies {
		if p.Contains(a) {
			return true
		}
	}
	return false
}
func (b *Boundary) Source(r *http.Request) (string, error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", staff.ErrValidation
	}
	peer, err := netip.ParseAddr(host)
	if err != nil || peer.Zone() != "" {
		return "", staff.ErrValidation
	}
	peer = peer.Unmap()
	if b.trusted(peer) {
		header := r.Header.Get("X-Forwarded-For")
		if len(header) > 2048 {
			return "", staff.ErrValidation
		}
		if header != "" {
			parts := strings.Split(header, ",")
			if len(parts) > 32 {
				return "", staff.ErrValidation
			}
			for i := len(parts) - 1; i >= 0 && b.trusted(peer); i-- {
				peer, err = netip.ParseAddr(strings.TrimSpace(parts[i]))
				if err != nil || peer.Zone() != "" {
					return "", staff.ErrValidation
				}
				peer = peer.Unmap()
			}
		}
	}
	// Return the canonical address, not its /64 prefix; application owns budgets.
	return peer.String(), nil
}

func Problem(status int) *huma.ErrorModel {
	code, detail := "unavailable", "Layanan sementara tidak tersedia."
	switch status {
	case 401:
		code, detail = "unauthenticated", "Silakan masuk kembali."
	case 403:
		code, detail = "denied", "Akses ditolak."
	case 409:
		code, detail = "conflict", "Data berubah. Muat ulang sebelum mencoba kembali."
	case 422:
		code, detail = "validation", "Periksa isian Anda."
	case 429:
		code, detail = "rate_limited", "Terlalu banyak percobaan. Coba lagi nanti."
	}
	return &huma.ErrorModel{Type: "urn:zenith:bo:" + code, Title: http.StatusText(status), Status: status, Detail: detail}
}
func Error(ctx context.Context, err error) error {
	status := 503
	switch {
	case errors.Is(err, staff.ErrUnauthenticated):
		status = 401
	case errors.Is(err, staff.ErrDenied):
		status = 403
	case errors.Is(err, staff.ErrConflict):
		status = 409
	case errors.Is(err, staff.ErrValidation):
		status = 422
	case errors.Is(err, staff.ErrRateLimited):
		status = 429
	}
	if v, ok := ctx.Value(requestKey{}).(requestState); ok {
		if status == 401 && v.boundary != nil {
			v.boundary.Clear(v.response)
		}
		var limited *staff.RateLimitError
		if errors.As(err, &limited) {
			v.response.Header().Set("Retry-After", strconv.Itoa(limited.RetryAfter))
		}
	}
	return Problem(status)
}

func (b *Boundary) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		if !strings.HasPrefix(r.URL.Path, "/v1/") {
			return next(c)
		}
		w := c.Response()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Add("Vary", "Origin")
		deny := func(status int) error {
			if status == 401 {
				b.Clear(w)
			}
			w.Header().Set("Content-Type", "application/problem+json")
			return c.JSON(status, Problem(status))
		}
		if b.Service == nil || b.Pool == nil || b.Clock == nil || b.Origin == "" || len(b.Secret) < 32 {
			return deny(503)
		}
		origin := r.Header.Get("Origin")
		if origin != "" && origin != b.Origin {
			return deny(403)
		}
		if origin == b.Origin {
			w.Header().Set("Access-Control-Allow-Origin", b.Origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			if origin != b.Origin {
				return deny(403)
			}
			method := r.Header.Get("Access-Control-Request-Method")
			if method != "GET" && method != "POST" && method != "PATCH" && method != "PUT" {
				return deny(403)
			}
			for _, header := range strings.Split(r.Header.Get("Access-Control-Request-Headers"), ",") {
				switch strings.ToLower(strings.TrimSpace(header)) {
				case "", "content-type", "x-csrf-token":
				default:
					return deny(403)
				}
			}
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			return c.NoContent(204)
		}
		mutation := r.Method != "GET" && r.Method != "HEAD"
		if mutation || r.URL.Path == "/v1/auth/csrf" {
			if origin != b.Origin {
				return deny(403)
			}
		}
		if mutation {
			media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || media != "application/json" {
				return deny(403)
			}
		}
		source, err := b.Source(r)
		if err != nil {
			return deny(422)
		}
		state := requestState{response: w, actor: staff.Actor{Source: source}, boundary: b}
		raw, err := uniqueCookie(r, b.SessionName())
		if err != nil {
			return deny(401)
		}
		if raw != "" {
			hash, err := staff.HashSessionToken(raw)
			if err != nil {
				return deny(401)
			}
			current, err := b.Service.Authenticate(r.Context(), hash)
			if err != nil {
				if errors.Is(err, staff.ErrUnauthenticated) {
					return deny(401)
				}
				return deny(503)
			}
			state.hash = hash
			state.actor.StaffID = current.ID
			state.actor.SessionHash = hash
		}
		anonymous := r.URL.Path == "/v1/auth/login" || r.URL.Path == "/v1/auth/csrf"
		if !anonymous && !state.actor.StaffID.Valid {
			return deny(401)
		}
		if r.URL.Path == "/v1/auth/login" && state.actor.StaffID.Valid {
			return deny(409)
		}
		if mutation {
			cookie, err := uniqueCookie(r, b.CSRFName())
			header := r.Header.Get("X-CSRF-Token")
			if err != nil || len(r.Header.Values("X-CSRF-Token")) != 1 || cookie == "" || len(header) > 256 || subtle.ConstantTimeCompare([]byte(header), []byte(cookie)) != 1 || !b.valid(header, state.hash) {
				return deny(403)
			}
		}
		c.SetRequest(r.WithContext(context.WithValue(r.Context(), requestKey{}, state)))
		return next(c)
	}
}
