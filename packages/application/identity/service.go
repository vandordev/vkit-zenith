package identity

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/netip"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	db "github.com/vandordev/vkit-zenith/packages/db"
)

type Service struct {
	pool        *pgxpool.Pool
	clock       Clock
	sender      DeliverySender
	limiterKey  [32]byte
	csrfKey     [32]byte
	authOrigin  string
	development bool
}

func NewService(pool *pgxpool.Pool, clock Clock, sender DeliverySender, limiterKey, csrfKey [32]byte, authOrigin string, development bool) *Service {
	if pool == nil || clock == nil || sender == nil || limiterKey == ([32]byte{}) || csrfKey == ([32]byte{}) || limiterKey == csrfKey || authOrigin == "" {
		panic("identity: explicit isolated dependencies required")
	}
	return &Service{pool: pool, clock: clock, sender: sender, limiterKey: limiterKey, csrfKey: csrfKey, authOrigin: authOrigin, development: development}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{ErrValidation, ErrConflict, ErrUnauthenticated, ErrDenied, ErrNotFound, ErrUnavailable, ErrOverloaded, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, known) {
			return known
		}
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505", "40001", "40P01", "55P03":
			return ErrConflict
		case "23514", "23503", "22003", "22021":
			return ErrValidation
		}
	}
	return ErrUnavailable
}
func mutationFailure(phase string, err error) error {
	failure := &MutationFailure{phase: phase, cause: mapError(err)}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		failure.sqlState = pg.Code
		failure.constraint = pg.ConstraintName
	}
	return failure
}
func identityStamp(at time.Time) time.Time { return at.UTC().Truncate(time.Microsecond) }
func newOpaque(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", ErrUnavailable
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func secretDigest(s string) [32]byte { return sha256.Sum256([]byte(s)) }

func (s *Service) admission(ctx context.Context, scope, purpose string, meta RequestMeta, account string, accountLimit int, accountWindow, accountCooldown time.Duration, sourceLimit int, sourceWindow time.Duration) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	source, err := trustedSource(meta.Source)
	if err != nil {
		return false, ErrValidation
	}
	keys := []struct {
		dimension, value string
		limit            int
		window           time.Duration
	}{
		{"account-hour", strings.ToLower(account) + "\x00" + purpose, accountLimit, accountWindow},
		{"account-day", strings.ToLower(account) + "\x00" + purpose, 10, 24 * time.Hour},
		{"source-hour", source, sourceLimit, sourceWindow},
		{"source-day", source, 100, 24 * time.Hour},
		{"global-hour", scope, 500, time.Hour},
	}
	if accountCooldown > 0 {
		keys = append(keys, struct {
			dimension, value string
			limit            int
			window           time.Duration
		}{"account-cooldown", strings.ToLower(account) + "\x00" + purpose, 1, accountCooldown})
	}
	type bucket struct {
		key       []byte
		dimension string
		limit     int
		window    time.Duration
	}
	buckets := make([]bucket, 0, len(keys))
	keyTime := identityStamp(s.clock.Now())
	for _, entry := range keys {
		mac := hmac.New(sha256.New, s.limiterKey[:])
		_, _ = fmt.Fprintf(mac, "%s\x00%s\x00%s\x00%d", scope, entry.dimension, entry.value, keyTime.UnixNano()/entry.window.Nanoseconds())
		buckets = append(buckets, bucket{mac.Sum(nil), entry.dimension, entry.limit, entry.window})
	}
	sort.Slice(buckets, func(i, j int) bool { return bytes.Compare(buckets[i].key, buckets[j].key) < 0 })
	var suppressed bool
	err = db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		if _, err := q.IdentityLockLimiterControl(ctx); err != nil {
			return err
		}
		at := identityStamp(s.clock.Now())
		for _, entry := range buckets {
			key := entry.key
			var current struct {
				WindowStartedAt pgtype.Timestamptz
				ExpiresAt       pgtype.Timestamptz
				Attempts        int32
			}
			var blocked bool
			err := tx.QueryRow(ctx, `SELECT window_started_at,expires_at,attempts FROM identity_limiter_buckets WHERE key_hash=$1 FOR UPDATE`, key).Scan(&current.WindowStartedAt, &current.ExpiresAt, &current.Attempts)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			if errors.Is(err, pgx.ErrNoRows) {
				_, err = tx.Exec(ctx, `INSERT INTO identity_limiter_buckets(key_hash,window_started_at,expires_at,attempts) VALUES($1,$2,$3,1) ON CONFLICT(key_hash) DO NOTHING`, key, at, at.Add(entry.window))
				if err != nil {
					return err
				}
				if _, err = tx.Exec(ctx, `SELECT key_hash FROM identity_limiter_buckets WHERE key_hash=$1 FOR UPDATE`, key); err != nil {
					return err
				}
				if err = tx.QueryRow(ctx, `SELECT window_started_at,expires_at,attempts FROM identity_limiter_buckets WHERE key_hash=$1`, key).Scan(&current.WindowStartedAt, &current.ExpiresAt, &current.Attempts); err != nil {
					return err
				}
			} else if !at.Before(current.ExpiresAt.Time) {
				_, err = tx.Exec(ctx, `UPDATE identity_limiter_buckets SET window_started_at=$2,expires_at=$3,attempts=1 WHERE key_hash=$1`, key, at, at.Add(entry.window))
				if err != nil {
					return err
				}
				current.Attempts = 1
			} else if int(current.Attempts) >= entry.limit {
				blocked = true
			} else {
				_, err = tx.Exec(ctx, `UPDATE identity_limiter_buckets SET attempts=attempts+1 WHERE key_hash=$1`, key)
				if err != nil {
					return err
				}
			}
			isAccount := strings.HasPrefix(entry.dimension, "account-")
			if blocked && !isAccount {
				return ErrOverloaded
			}
			if blocked && isAccount {
				suppressed = true
			}
		}
		// Bounded cleanup remains in this same singleton→sorted-bucket ordering.
		_, err := tx.Exec(ctx, `WITH expired AS (SELECT key_hash FROM identity_limiter_buckets WHERE expires_at<=$1 ORDER BY expires_at,key_hash LIMIT 100 FOR UPDATE) DELETE FROM identity_limiter_buckets b USING expired e WHERE b.key_hash=e.key_hash`, at)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `DELETE FROM identity_enrollment_receipts WHERE secret_hash IN (SELECT secret_hash FROM identity_enrollment_receipts WHERE expires_at<=$1 ORDER BY expires_at,secret_hash LIMIT 100)`, at)
		return err
	})
	return suppressed, mapError(err)
}

// Register always performs a bounded password KDF and returns the same generic
// acceptance/receipt shape. Duplicate normalized addresses receive an unbound
// receipt and no account-derived state. Delivery is attempted only after commit;
// provider failure never changes the public accepted response.
func (s *Service) Register(ctx context.Context, in RegisterInput, meta RequestMeta) (RegistrationResult, error) {
	var result RegistrationResult
	normalized, err := NormalizeRegistration(in)
	if err != nil {
		return result, err
	}
	if meta.Source != "" {
		if _, err = trustedSource(meta.Source); err != nil {
			return result, ErrValidation
		}
	}
	// Calculate the exact eventual UTF-8 payload size before quota writes/KDF.
	if len(normalized.DisplayName)+len(normalized.Email)+len(normalized.Telephone)+len(normalized.Password)+256 > 16*1024 {
		return result, ErrValidation
	}
	suppressed, err := s.admission(ctx, "register", string(ActionVerifyEmail), meta, normalized.Email, 5, time.Hour, time.Minute, 30, time.Hour)
	if err != nil {
		if errors.Is(err, ErrOverloaded) {
			return result, err
		}
		return result, err
	}
	passwordHash, err := HashPassword(ctx, normalized.Password)
	if err != nil {
		return result, mapError(err)
	}
	receipt, receiptHash, err := cookieSecret()
	if err != nil {
		return result, err
	}
	sessionSecret, sessionHash, err := cookieSecret()
	if err != nil {
		return result, err
	}
	fullToken, err := newOpaque(32)
	if err != nil {
		return result, err
	}
	tokenHash := secretDigest(fullToken)
	attempt, err := newOpaque(24)
	if err != nil {
		return result, err
	}
	if suppressed {
		return RegistrationResult{Accepted: PublicAccepted{}, EnrollmentSecret: receipt, SessionSecret: sessionSecret}, nil
	}
	acceptedAt := identityStamp(s.clock.Now())
	var delivery *ActionDelivery
	var created bool
	err = db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		// accepted time and KDF result are revalidated inside the complete write
		// transaction; no status write is made, so owner trigger does not fire.
		at := identityStamp(s.clock.Now())
		if at.Before(acceptedAt) {
			at = acceptedAt
		}
		acceptedAt = at
		q := persistence.New(tx)
		user, insertErr := q.IdentityInsertUser(ctx, persistence.IdentityInsertUserParams{Email: normalized.Email, DisplayName: normalized.DisplayName, CreatedAt: pgtype.Timestamptz{Time: at, Valid: true}, Telephone: pgtype.Text{String: normalized.Telephone, Valid: true}})
		if insertErr != nil && !errors.Is(insertErr, pgx.ErrNoRows) {
			return insertErr
		}
		if user.ID.Valid {
			created = true
			if err := q.IdentityInsertCredential(ctx, persistence.IdentityInsertCredentialParams{UserID: user.ID, PasswordHash: passwordHash, PasswordChangedAt: pgtype.Timestamptz{Time: at, Valid: true}}); err != nil {
				return err
			}
			if err := q.IdentityInsertRestrictedSession(ctx, persistence.IdentityInsertRestrictedSessionParams{UserID: user.ID, SecretHash: sessionHash[:], AuthenticatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: at.Add(RestrictedSessionLifetime), Valid: true}, IdentityVersion: pgtype.Int8{Int64: 1, Valid: true}, CredentialGeneration: pgtype.Int8{Int64: 1, Valid: true}}); err != nil {
				return err
			}
			if err := q.IdentityInsertEnrollmentReceipt(ctx, persistence.IdentityInsertEnrollmentReceiptParams{SecretHash: receiptHash[:], UserID: user.ID, IdentityVersion: pgtype.Int8{Int64: 1, Valid: true}, CredentialGeneration: pgtype.Int8{Int64: 1, Valid: true}, CreatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: at.Add(RestrictedSessionLifetime), Valid: true}}); err != nil {
				return err
			}
			expires := at.Add(VerifyEmailLifetime)
			if err := q.IdentityInsertAction(ctx, persistence.IdentityInsertActionParams{UserID: user.ID, Purpose: string(ActionVerifyEmail), TargetEmailNormalized: normalized.Email, TokenHash: tokenHash[:], CreatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true}, IdentityVersion: pgtype.Int8{Int64: 1, Valid: true}, CredentialGeneration: pgtype.Int8{Int64: 1, Valid: true}, EnrollmentHash: receiptHash[:]}); err != nil {
				return err
			}
			if err := q.IdentityInsertAudit(ctx, persistence.IdentityInsertAuditParams{OccurredAt: pgtype.Timestamptz{Time: at, Valid: true}, Action: "identity.register", TargetType: pgtype.Text{String: "user", Valid: true}, TargetID: pgtype.Text{String: user.ID.String(), Valid: true}, Column5: auditDetail("email", "credentials", "telephone", "identity_version")}); err != nil {
				return err
			}
			delivery = &ActionDelivery{AttemptID: attempt, Recipient: normalized.Email, DisplayName: normalized.DisplayName, URL: actionURL(s.authOrigin, ActionVerifyEmail, fullToken), Purpose: ActionVerifyEmail, ExpiresAt: expires}
		}
		// Both branches issue a same-size opaque receipt; only a created account
		// has a server-side user binding. Duplicate receipt never authenticates it.
		if !created {
			if err := q.IdentityInsertEnrollmentReceipt(ctx, persistence.IdentityInsertEnrollmentReceiptParams{SecretHash: receiptHash[:], CreatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: at.Add(RestrictedSessionLifetime), Valid: true}}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return result, mapError(err)
	}
	result = RegistrationResult{Accepted: PublicAccepted{}, EnrollmentSecret: receipt, SessionSecret: sessionSecret}
	if delivery != nil && s.sender != nil {
		_ = s.sender.Send(ctx, *delivery)
	}
	return result, nil
}

// RequestPasswordAction returns the same accepted DTO for unknown, disabled,
// passwordless, unverified and eligible email addresses. A passwordless active
// account receives the distinct first-password purpose; an existing credential
// receives reset purpose. Raw token exists only in the post-commit delivery.
func (s *Service) RequestPasswordAction(ctx context.Context, emailAddress string, meta RequestMeta) (PublicAccepted, error) {
	address, ok := canonicalEmail(emailAddress)
	if !ok {
		return PublicAccepted{}, ErrValidation
	}
	suppressed, err := s.admission(ctx, "password-action", "password_action", meta, address, 5, time.Hour, time.Minute, 30, time.Hour)
	if err != nil {
		return PublicAccepted{}, err
	}
	if suppressed {
		return PublicAccepted{}, nil
	}
	token, err := newOpaque(32)
	if err != nil {
		return PublicAccepted{}, err
	}
	tokenHash := secretDigest(token)
	attempt, err := newOpaque(24)
	if err != nil {
		return PublicAccepted{}, err
	}
	var delivery *ActionDelivery
	err = db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		unlocked, err := q.IdentityFindUserByEmail(ctx, address)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		user, err := q.IdentityLockUser(ctx, unlocked.ID)
		if err != nil {
			return err
		}
		at := identityStamp(s.clock.Now())
		if user.Status != "active" || user.EmailNormalized != address {
			return nil
		}
		credential, credentialErr := q.IdentityLockCredential(ctx, user.ID)
		if credentialErr != nil && !errors.Is(credentialErr, pgx.ErrNoRows) {
			return credentialErr
		}
		purpose := ActionSetupPassword
		generation := int64(0)
		version := int64(1)
		if user.IdentityVersion.Valid {
			version = user.IdentityVersion.Int64
		}
		if credentialErr == nil {
			purpose = ActionResetPassword
			generation = 1
			if credential.Generation.Valid {
				generation = credential.Generation.Int64
			}
		}
		if _, err = tx.Exec(ctx, `SELECT id FROM email_action_tokens WHERE user_id=$1 AND purpose=$2 AND consumed_at IS NULL AND revoked_at IS NULL ORDER BY id FOR UPDATE`, user.ID, string(purpose)); err != nil {
			return err
		}
		var live int
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens WHERE user_id=$1 AND purpose=$2 AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$3`, user.ID, string(purpose), at).Scan(&live); err != nil {
			return err
		}
		if live >= 10 {
			return nil
		}
		expires := at.Add(SetupPasswordLifetime)
		if purpose == ActionResetPassword {
			expires = at.Add(ResetPasswordLifetime)
		}
		if user.IdentityVersion.Valid && user.IdentityVersion.Int64 != version {
			return ErrConflict
		}
		if err = q.IdentityInsertAction(ctx, persistence.IdentityInsertActionParams{UserID: user.ID, Purpose: string(purpose), TargetEmailNormalized: address, TokenHash: tokenHash[:], CreatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true}, IdentityVersion: pgtype.Int8{Int64: version, Valid: true}, CredentialGeneration: pgtype.Int8{Int64: generation, Valid: true}}); err != nil {
			return err
		}
		delivery = &ActionDelivery{AttemptID: attempt, Recipient: address, DisplayName: user.DisplayName, URL: actionURL(s.authOrigin, purpose, token), Purpose: purpose, ExpiresAt: expires}
		return nil
	})
	if err != nil {
		return PublicAccepted{}, mapError(err)
	}
	if delivery != nil {
		_ = s.sender.Send(ctx, *delivery)
	}
	return PublicAccepted{}, nil
}

// RequestVerification is restricted-session-only. The caller supplies hashes
// derived from the independent host-only session and enrollment cookies; the
// action trigger participates in the existing user-first fence.
func (s *Service) RequestVerification(ctx context.Context, meta RequestMeta) (PublicAccepted, error) {
	if meta.SessionHash == ([32]byte{}) || meta.EnrollmentHash == ([32]byte{}) {
		return PublicAccepted{}, ErrUnauthenticated
	}
	var scope struct {
		userID pgtype.UUID
		email  string
	}
	if err := s.pool.QueryRow(ctx, `SELECT u.id,u.email_normalized FROM auth_sessions s JOIN users u ON u.id=s.user_id WHERE s.secret_hash=$1`, meta.SessionHash[:]).Scan(&scope.userID, &scope.email); err != nil {
		return PublicAccepted{}, nil
	}
	suppressed, err := s.admission(ctx, "verify-email", string(ActionVerifyEmail), meta, scope.email, 5, time.Hour, time.Minute, 30, time.Hour)
	if err != nil {
		return PublicAccepted{}, err
	}
	if suppressed {
		return PublicAccepted{}, nil
	}
	token, err := newOpaque(32)
	if err != nil {
		return PublicAccepted{}, err
	}
	tokenHash := secretDigest(token)
	attempt, err := newOpaque(24)
	if err != nil {
		return PublicAccepted{}, err
	}
	var delivery *ActionDelivery
	err = db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		user, err := q.IdentityLockUser(ctx, scope.userID)
		if err != nil {
			return err
		}
		credential, err := q.IdentityLockCredential(ctx, user.ID)
		if err != nil {
			return err
		}
		sessionID, err := q.IdentityFindSessionByHash(ctx, persistence.IdentityFindSessionByHashParams{SecretHash: meta.SessionHash[:], UserID: user.ID})
		if err != nil {
			return nil
		}
		session, err := q.IdentityLockSession(ctx, sessionID.ID)
		if err != nil {
			return err
		}
		at := identityStamp(s.clock.Now())
		if user.Status != "active" || user.EmailNormalized != scope.email || user.EmailVerifiedAt.Valid || !session.Kind.Valid || session.Kind.String != string(SessionRestricted) || !session.IdentityVersion.Valid || !user.IdentityVersion.Valid || session.IdentityVersion.Int64 != user.IdentityVersion.Int64 || !session.CredentialGeneration.Valid || !credential.Generation.Valid || session.CredentialGeneration.Int64 != credential.Generation.Int64 || session.RevokedAt.Valid || !BeforeExpiry(at, session.ExpiresAt.Time) {
			return nil
		}
		receipt, err := q.IdentityLockEnrollmentReceipt(ctx, meta.EnrollmentHash[:])
		if err != nil || !receipt.UserID.Valid || receipt.UserID != user.ID || !receipt.IdentityVersion.Valid || receipt.IdentityVersion.Int64 != user.IdentityVersion.Int64 || !receipt.CredentialGeneration.Valid || receipt.CredentialGeneration.Int64 != credential.Generation.Int64 || !BeforeExpiry(at, receipt.ExpiresAt.Time) {
			return nil
		}
		_, err = tx.Exec(ctx, `SELECT id FROM email_action_tokens WHERE user_id=$1 AND purpose='verify_email' AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$2 ORDER BY id FOR UPDATE`, user.ID, at)
		if err != nil {
			return err
		}
		var live int
		if err = tx.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens WHERE user_id=$1 AND purpose='verify_email' AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$2`, user.ID, at).Scan(&live); err != nil {
			return err
		}
		if live >= 10 {
			return nil
		}
		expires := at.Add(VerifyEmailLifetime)
		if err = q.IdentityInsertAction(ctx, persistence.IdentityInsertActionParams{UserID: user.ID, Purpose: string(ActionVerifyEmail), TargetEmailNormalized: user.EmailNormalized, TokenHash: tokenHash[:], CreatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true}, IdentityVersion: user.IdentityVersion, CredentialGeneration: credential.Generation, EnrollmentHash: meta.EnrollmentHash[:]}); err != nil {
			return err
		}
		delivery = &ActionDelivery{AttemptID: attempt, Recipient: user.EmailNormalized, DisplayName: user.DisplayName, URL: actionURL(s.authOrigin, ActionVerifyEmail, token), Purpose: ActionVerifyEmail, ExpiresAt: expires}
		return nil
	})
	if err != nil {
		return PublicAccepted{}, mapError(err)
	}
	if delivery != nil {
		_ = s.sender.Send(ctx, *delivery)
	}
	return PublicAccepted{}, nil
}

// Login performs bounded password work before entering the user-first lock chain.
// Unknown accounts use the same dummy PHC profile; disabled/ineligible accounts
// never receive a session and no result contains account-existence detail.
func (s *Service) Login(ctx context.Context, in LoginInput, meta RequestMeta) (SessionIssue, error) {
	var result SessionIssue
	address, ok := canonicalEmail(in.Email)
	if !ok || !validPassword(in.Password) {
		return result, ErrValidation
	}
	if meta.Source == "" {
		return result, ErrValidation
	}
	suppressed, err := s.admission(ctx, "login", "password_login", meta, address, 10, 15*time.Minute, 0, 50, 15*time.Minute)
	if err != nil {
		return result, err
	}
	if suppressed {
		return result, ErrUnauthenticated
	}
	var user persistence.User
	var credential persistence.PasswordCredential
	err = s.pool.QueryRow(ctx, `SELECT id,email,email_normalized,display_name,status,email_verified_at,created_at,updated_at,telephone,identity_version FROM users WHERE email_normalized=$1`, address).Scan(&user.ID, &user.Email, &user.EmailNormalized, &user.DisplayName, &user.Status, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt, &user.Telephone, &user.IdentityVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		if err = VerifyUnknownPassword(ctx, in.Password); err != nil {
			return result, mapError(err)
		}
		return result, ErrUnauthenticated
	}
	if err != nil {
		return result, ErrUnavailable
	}
	err = s.pool.QueryRow(ctx, `SELECT user_id,password_hash,password_changed_at,generation FROM password_credentials WHERE user_id=$1`, user.ID).Scan(&credential.UserID, &credential.PasswordHash, &credential.PasswordChangedAt, &credential.Generation)
	if errors.Is(err, pgx.ErrNoRows) {
		if err = VerifyUnknownPassword(ctx, in.Password); err != nil {
			return result, mapError(err)
		}
		return result, ErrUnauthenticated
	}
	if err != nil {
		return result, ErrUnavailable
	}
	if user.Status != "active" || !credential.Generation.Valid || !user.IdentityVersion.Valid {
		if err = VerifyUnknownPassword(ctx, in.Password); err != nil {
			return result, mapError(err)
		}
		return result, ErrUnauthenticated
	}
	valid, err := verifyIdentityPassword(ctx, credential.PasswordHash, in.Password)
	if errors.Is(err, ErrOverloaded) {
		return result, err
	}
	if err != nil {
		return result, mapError(err)
	}
	if !valid || user.Status != "active" {
		return result, ErrUnauthenticated
	}
	secret, secretHash, err := cookieSecret()
	if err != nil {
		return result, err
	}
	sessionHash := SecretHash(secretHash)
	csrf, err := s.NewSessionCSRF(sessionHash)
	if err != nil {
		return result, err
	}
	enrollmentSecret := ""
	var issued persistence.User
	var full bool
	var reusedSession bool
	var deadline, authenticatedAt time.Time
	err = db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		locked, err := q.IdentityLockUser(ctx, user.ID)
		if err != nil {
			return err
		}
		current, err := q.IdentityLockCredential(ctx, locked.ID)
		if err != nil {
			return err
		}
		if locked.Status != "active" || locked.EmailNormalized != address || !locked.IdentityVersion.Valid || !current.Generation.Valid || locked.IdentityVersion.Int64 != user.IdentityVersion.Int64 || current.Generation.Int64 != credential.Generation.Int64 || current.PasswordHash != credential.PasswordHash {
			return ErrUnauthenticated
		}
		at := identityStamp(s.clock.Now())
		kind := SessionRestricted
		duration := RestrictedSessionLifetime
		if locked.EmailVerifiedAt.Valid {
			kind = SessionFull
			duration = SessionIdleLifetime
			full = true
		}
		deadline = at.Add(duration)
		authenticatedAt = at
		issued = locked
		if meta.SessionHash != ([32]byte{}) {
			previous, previousErr := q.IdentityFindSessionByHash(ctx, persistence.IdentityFindSessionByHashParams{SecretHash: meta.SessionHash[:], UserID: locked.ID})
			if previousErr == nil {
				previous, previousErr = q.IdentityLockSession(ctx, previous.ID)
				if previousErr != nil {
					return previousErr
				}
				if !previous.RevokedAt.Valid {
					rows, rotateErr := q.IdentityRotateLoginSession(ctx, persistence.IdentityRotateLoginSessionParams{ID: previous.ID, SecretHash: secretHash[:], AuthenticatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: deadline, Valid: true}, Kind: pgtype.Text{String: string(kind), Valid: true}, IdentityVersion: locked.IdentityVersion, CredentialGeneration: current.Generation, SecretHash_2: meta.SessionHash[:]})
					if rotateErr != nil {
						return rotateErr
					}
					if rows != 1 {
						return ErrConflict
					}
					reusedSession = true
				}
			} else if !errors.Is(previousErr, pgx.ErrNoRows) {
				return previousErr
			}
		}
		if !reusedSession {
			if err = q.IdentityInsertLoginSession(ctx, persistence.IdentityInsertLoginSessionParams{UserID: locked.ID, SecretHash: secretHash[:], AuthenticatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: deadline, Valid: true}, Kind: pgtype.Text{String: string(kind), Valid: true}, IdentityVersion: locked.IdentityVersion, CredentialGeneration: current.Generation}); err != nil {
				return err
			}
		}
		if kind == SessionRestricted {
			var enrollmentHash [32]byte
			enrollmentSecret, enrollmentHash, err = cookieSecret()
			if err != nil {
				return err
			}
			if err = q.IdentityInsertEnrollmentReceipt(ctx, persistence.IdentityInsertEnrollmentReceiptParams{SecretHash: enrollmentHash[:], UserID: locked.ID, IdentityVersion: locked.IdentityVersion, CredentialGeneration: current.Generation, CreatedAt: pgtype.Timestamptz{Time: at, Valid: true}, ExpiresAt: pgtype.Timestamptz{Time: at.Add(RestrictedSessionLifetime), Valid: true}}); err != nil {
				return err
			}
		}
		action := "identity.login.restricted"
		if kind == SessionFull {
			action = "identity.login.full"
		}
		return q.IdentityInsertAudit(ctx, persistence.IdentityInsertAuditParams{OccurredAt: pgtype.Timestamptz{Time: at, Valid: true}, Action: action, TargetType: pgtype.Text{String: "user", Valid: true}, TargetID: pgtype.Text{String: locked.ID.String(), Valid: true}, Column5: auditDetail("session")})
	})
	if err != nil {
		return result, mapError(err)
	}
	kind := SessionRestricted
	if full {
		kind = SessionFull
	}
	result = SessionIssue{Secret: secret, CSRFToken: csrf, EnrollmentSecret: enrollmentSecret, ExpiresAt: deadline, CookieMaxAge: int(deadline.Sub(authenticatedAt).Seconds()), View: SessionView{Kind: kind, Subject: issued.ID.String(), DisplayName: issued.DisplayName, Email: issued.EmailNormalized, EmailVerified: issued.EmailVerifiedAt.Valid, AuthenticatedAt: authenticatedAt, IdleExpiresAt: deadline}}
	return result, nil
}

// Current never extends either the server deadline or browser cookie deadline.
func (s *Service) Current(ctx context.Context, hash SecretHash) (SessionView, error) {
	var view SessionView
	if s == nil || s.pool == nil || s.clock == nil || hash == ([32]byte{}) {
		return view, ErrUnauthenticated
	}
	var kind, status string
	var verified, revoked, browserExpires pgtype.Timestamptz
	var sessionVersion, sessionGeneration, credentialGeneration, userVersion pgtype.Int8
	err := s.pool.QueryRow(ctx, `SELECT s.kind,u.id,s.id,u.display_name,u.email_normalized,u.email_verified_at,s.authenticated_at,s.expires_at,s.browser_cookie_expires_at,s.revoked_at,s.identity_version,s.credential_generation,c.generation,u.status,u.identity_version FROM auth_sessions s JOIN users u ON u.id=s.user_id JOIN password_credentials c ON c.user_id=u.id WHERE s.secret_hash=$1`, hash[:]).Scan(&kind, &view.Subject, &view.SessionID, &view.DisplayName, &view.Email, &verified, &view.AuthenticatedAt, &view.IdleExpiresAt, &browserExpires, &revoked, &sessionVersion, &sessionGeneration, &credentialGeneration, &status, &userVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionView{}, ErrUnauthenticated
		}
		return SessionView{}, ErrUnavailable
	}
	at := identityStamp(s.clock.Now())
	if status != "active" || revoked.Valid || !sessionVersion.Valid || !sessionGeneration.Valid || !credentialGeneration.Valid || !userVersion.Valid || sessionVersion.Int64 != userVersion.Int64 || sessionGeneration.Int64 != credentialGeneration.Int64 || !BeforeExpiry(at, view.IdleExpiresAt) || !browserExpires.Valid || !BeforeExpiry(at, browserExpires.Time) {
		return SessionView{}, ErrUnauthenticated
	}
	if kind != string(SessionRestricted) && kind != string(SessionFull) {
		return SessionView{}, ErrUnauthenticated
	}
	if kind == string(SessionFull) && !verified.Valid {
		return SessionView{}, ErrUnauthenticated
	}
	if kind == string(SessionRestricted) && verified.Valid {
		return SessionView{}, ErrUnauthenticated
	}
	view.Kind = SessionKind(kind)
	view.EmailVerified = verified.Valid
	view.CredentialGeneration = sessionGeneration.Int64
	return view, nil
}

// Logout is idempotent and user-first. Invalid/unknown browser credentials have
// the same public outcome and cannot select an arbitrary user.
func (s *Service) Logout(ctx context.Context, hash SecretHash) error {
	if s == nil || s.pool == nil || hash == ([32]byte{}) {
		return nil
	}
	var owner persistence.AuthSession
	if err := s.pool.QueryRow(ctx, `SELECT id,user_id FROM auth_sessions WHERE secret_hash=$1`, hash[:]).Scan(&owner.ID, &owner.UserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return ErrUnavailable
	}
	err := db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		user, err := q.IdentityLockUser(ctx, owner.UserID)
		if err != nil {
			return err
		}
		if _, err = q.IdentityLockCredential(ctx, user.ID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		locked, err := q.IdentityLockSession(ctx, owner.ID)
		if err != nil {
			return err
		}
		if !hmac.Equal(locked.SecretHash, hash[:]) || locked.UserID != user.ID {
			return nil
		}
		at := identityStamp(s.clock.Now())
		if locked.RevokedAt.Valid {
			return nil
		}
		if _, err = q.IdentityRevokeSession(ctx, persistence.IdentityRevokeSessionParams{ID: locked.ID, RevokedAt: pgtype.Timestamptz{Time: at, Valid: true}}); err != nil {
			return err
		}
		return q.IdentityInsertAudit(ctx, persistence.IdentityInsertAuditParams{OccurredAt: pgtype.Timestamptz{Time: at, Valid: true}, Action: "identity.session.logout", TargetType: pgtype.Text{String: "user", Valid: true}, TargetID: pgtype.Text{String: user.ID.String(), Valid: true}, Column5: auditDetail("session")})
	})
	return mapError(err)
}

// TouchFullSession records accepted foreground activity no more than once per
// minute. Read-only Current/Login paths never call this transition.
func (s *Service) TouchFullSession(ctx context.Context, hash SecretHash) (time.Time, int, error) {
	var deadline time.Time
	if hash == ([32]byte{}) {
		return deadline, 0, ErrUnauthenticated
	}
	var owner pgtype.UUID
	if err := s.pool.QueryRow(ctx, `SELECT user_id FROM auth_sessions WHERE secret_hash=$1`, hash[:]).Scan(&owner); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deadline, 0, ErrUnauthenticated
		}
		return deadline, 0, ErrUnavailable
	}
	err := db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		user, err := q.IdentityLockUser(ctx, owner)
		if err != nil {
			return err
		}
		credential, err := q.IdentityLockCredential(ctx, user.ID)
		if err != nil {
			return err
		}
		session, err := q.IdentityFindSessionByHash(ctx, persistence.IdentityFindSessionByHashParams{SecretHash: hash[:], UserID: user.ID})
		if err != nil {
			return ErrUnauthenticated
		}
		session, err = q.IdentityLockSession(ctx, session.ID)
		if err != nil {
			return err
		}
		at := identityStamp(s.clock.Now())
		if user.Status != "active" || !user.EmailVerifiedAt.Valid || !user.IdentityVersion.Valid || !credential.Generation.Valid || !session.Kind.Valid || session.Kind.String != string(SessionFull) || !session.IdentityVersion.Valid || session.IdentityVersion.Int64 != user.IdentityVersion.Int64 || !session.CredentialGeneration.Valid || session.CredentialGeneration.Int64 != credential.Generation.Int64 || session.RevokedAt.Valid || !BeforeExpiry(at, session.ExpiresAt.Time) || !session.BrowserCookieExpiresAt.Valid || !BeforeExpiry(at, session.BrowserCookieExpiresAt.Time) {
			return ErrUnauthenticated
		}
		if !session.LastActivityAt.Valid || at.Before(session.LastActivityAt.Time.Add(time.Minute)) {
			deadline = session.ExpiresAt.Time
			return nil
		}
		updated, err := q.IdentityTouchFullSession(ctx, persistence.IdentityTouchFullSessionParams{ID: session.ID, LastActivityAt: pgtype.Timestamptz{Time: at, Valid: true}})
		if err != nil {
			return err
		}
		if updated != 1 {
			return ErrConflict
		}
		deadline = at.Add(SessionIdleLifetime)
		return nil
	})
	if err != nil {
		return time.Time{}, 0, mapError(err)
	}
	remaining := int(deadline.Sub(identityStamp(s.clock.Now())).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return deadline, remaining, nil
}

// NewAnonymousCSRF issues a short-lived, signed, session-free token for public
// same-origin form operations. It is returned in JSON and must remain in memory.
func (s *Service) NewAnonymousCSRF() (string, error) {
	return s.newCSRF("anonymous", [32]byte{})
}

func (s *Service) NewSessionCSRF(sessionHash SecretHash) (string, error) {
	if sessionHash == ([32]byte{}) {
		return "", ErrUnauthenticated
	}
	return s.newCSRF("session", sessionHash)
}

func (s *Service) newCSRF(scope string, sessionHash SecretHash) (string, error) {
	var payload [40]byte
	if _, err := rand.Read(payload[:32]); err != nil {
		return "", ErrUnavailable
	}
	deadline := identityStamp(s.clock.Now()).Add(10 * time.Minute).Unix()
	binary.BigEndian.PutUint64(payload[32:], uint64(deadline))
	mac := hmac.New(sha256.New, s.csrfKey[:])
	_, _ = mac.Write(payload[:])
	_, _ = mac.Write([]byte("\x00" + scope + "\x00" + s.authOrigin))
	if scope == "session" {
		_, _ = mac.Write(sessionHash[:])
	}
	signed := append(payload[:], mac.Sum(nil)...)
	return base64.RawURLEncoding.EncodeToString(signed), nil
}

func (s *Service) ValidateAnonymousCSRF(token string) bool {
	return s.validateCSRF("anonymous", [32]byte{}, token)
}

func (s *Service) ValidateSessionCSRF(sessionHash SecretHash, token string) bool {
	return sessionHash != ([32]byte{}) && s.validateCSRF("session", sessionHash, token)
}

func (s *Service) ValidateCSRF(sessionHash SecretHash, token string) bool {
	if sessionHash != ([32]byte{}) {
		return s.ValidateSessionCSRF(sessionHash, token)
	}
	return s.ValidateAnonymousCSRF(token)
}

func (s *Service) validateCSRF(scope string, sessionHash SecretHash, token string) bool {
	if s == nil || len(token) != 96 {
		return false
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(raw) != 72 || base64.RawURLEncoding.EncodeToString(raw) != token {
		return false
	}
	mac := hmac.New(sha256.New, s.csrfKey[:])
	_, _ = mac.Write(raw[:40])
	_, _ = mac.Write([]byte("\x00" + scope + "\x00" + s.authOrigin))
	if scope == "session" {
		_, _ = mac.Write(sessionHash[:])
	}
	if !hmac.Equal(raw[40:], mac.Sum(nil)) {
		return false
	}
	deadline := int64(binary.BigEndian.Uint64(raw[32:40]))
	now := s.clock.Now().Unix()
	return deadline > now && deadline <= now+int64(10*time.Minute/time.Second)
}

func trustedSource(value string) (string, error) {
	if len(value) > 256 || strings.ContainsAny(value, "\r\n\x00") {
		return "", ErrValidation
	}
	if value == "" {
		return "unknown", nil
	}
	ip := net.ParseIP(value)
	if ip == nil {
		return "", ErrValidation
	}
	normalized := netip.MustParseAddr(ip.String()).String()
	return normalized, nil
}

func safeDisplayName(value string) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && utf8.RuneCountInString(value) <= 200 && strings.TrimSpace(value) != ""
}
func canonicalEmail(value string) (string, bool) {
	if !utf8.ValidString(value) || len(value) > 254 || strings.ContainsAny(value, "\x00\r\n") {
		return "", false
	}
	value = strings.Trim(value, " ")
	if !emailPattern.MatchString(value) {
		return "", false
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return "", false
	}
	return strings.ToLower(value), true
}
func cookieSecret() (string, [32]byte, error) {
	value, err := newOpaque(32)
	if err != nil {
		return "", [32]byte{}, err
	}
	return value, secretDigest(value), nil
}
func actionURL(origin string, purpose ActionPurpose, token string) string {
	path := map[ActionPurpose]string{ActionVerifyEmail: "/verify-email", ActionSetupPassword: "/set-password", ActionResetPassword: "/reset-password"}[purpose]
	return origin + path + "?token=" + token
}
func auditID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", ErrUnavailable
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
func auditDetail(fields ...string) []byte {
	data, _ := json.Marshal(map[string]any{"changed_fields": fields})
	return data
}
