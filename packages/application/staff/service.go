package staff

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
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
	pool      *pgxpool.Pool
	clock     func() time.Time
	passwords *Passwords
}

func NewService(pool *pgxpool.Pool, clock func() time.Time, passwords *Passwords) *Service {
	if pool == nil || clock == nil || passwords == nil {
		panic("staff: explicit dependencies required")
	}
	return &Service{pool: pool, clock: clock, passwords: passwords}
}

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)

func normalizeWrite(in StaffWrite) (StaffWrite, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if !utf8.ValidString(in.Email) || len(in.Email) > 254 || !emailPattern.MatchString(in.Email) || !utf8.ValidString(in.DisplayName) || utf8.RuneCountInString(in.DisplayName) > 200 || strings.TrimSpace(in.DisplayName) == "" ||
		(in.Role != "superadmin" && in.Role != "admin") || (in.Status != "active" && in.Status != "disabled") {
		return StaffWrite{}, ErrValidation
	}
	return in, nil
}
func stamp(at time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: at.UTC().Truncate(time.Microsecond), Valid: true}
}
func newID() (pgtype.UUID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return pgtype.UUID{}, ErrUnavailable
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return pgtype.UUID{Bytes: b, Valid: true}, nil
}

// Revision is canonical safe persisted state, never credential material or xmin.
func Revision(s Staff) string {
	b, _ := json.Marshal(struct{ ID, Email, Name, Role, Status, Updated string }{s.ID.String(), s.Email, s.DisplayName, s.Role, s.Status, s.UpdatedAt.UTC().Format(time.RFC3339Nano)})
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}
func matchRevision(s Staff, input string) bool {
	expected := Revision(s)
	return len(input) == 64 && subtle.ConstantTimeCompare([]byte(expected), []byte(input)) == 1
}

func safeError(err error) error {
	if err == nil {
		return nil
	}
	var limited *RateLimitError
	if errors.As(err, &limited) {
		return limited
	}
	for _, safe := range []error{ErrValidation, ErrUnauthenticated, ErrDenied, ErrConflict, ErrNotFound, ErrUnavailable, ErrOverloaded, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, safe) {
			return safe
		}
	}
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		switch pe.Code {
		case "23505", "40001", "40P01":
			return ErrConflict
		case "23514":
			if pe.ConstraintName == "platform_staff_last_superadmin_check" {
				return ErrConflict
			}
			return ErrValidation
		}
	}
	return ErrUnavailable
}
func (s *Service) transaction(ctx context.Context, fn func(*persistence.Queries, pgx.Tx, time.Time) error) error {
	return safeError(db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s';SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		if _, err := q.StaffFence(ctx); err != nil {
			return err
		}
		// Accepted time is chosen after admission, not before waiting on a fence.
		at := stamp(s.clock()).Time
		if err := fn(q, tx, at); err != nil {
			return err
		}
		// Surface deferred guards before db.InTx's safe opaque commit-error category.
		_, err := tx.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`)
		return err
	}))
}
func audit(ctx context.Context, q *persistence.Queries, actor, target pgtype.UUID, action string, at time.Time, detail any) error {
	id, err := newID()
	if err != nil {
		return err
	}
	b, err := json.Marshal(detail)
	if err != nil {
		return ErrUnavailable
	}
	return q.StaffAudit(ctx, persistence.StaffAuditParams{ID: id, OccurredAt: stamp(at), ActorStaffID: actor, Action: action, TargetID: pgtype.Text{String: target.String(), Valid: true}, Detail: b})
}
func staffFromRow(r persistence.StaffLockAccountsRow) Staff {
	return Staff{ID: r.ID, Email: r.EmailNormalized, DisplayName: r.DisplayName, Role: r.Role, Status: r.Status, UpdatedAt: r.UpdatedAt.Time}
}

// KDF occurs before locks. Snapshot is checked again under the mutation fence.
func (s *Service) confirmation(ctx context.Context, actor Actor, password string) (persistence.PlatformStaffCredential, error) {
	current, err := s.Authenticate(ctx, actor.SessionHash)
	if err != nil {
		return persistence.PlatformStaffCredential{}, err
	}
	if current.ID != actor.StaffID {
		return persistence.PlatformStaffCredential{}, ErrUnauthenticated
	}
	admission, err := s.admit(ctx, "confirmation", current.Email, actor.Source)
	if err != nil {
		return persistence.PlatformStaffCredential{}, err
	}
	confirmed, err := s.verifyConfirmation(ctx, actor, password)
	if errors.Is(err, ErrDenied) {
		if e := s.failure(ctx, admission); e != nil {
			return confirmed, e
		}
	}
	return confirmed, err
}

func (s *Service) verifyConfirmation(ctx context.Context, actor Actor, password string) (persistence.PlatformStaffCredential, error) {
	if !actor.StaffID.Valid || !validPassword(password) {
		return persistence.PlatformStaffCredential{}, ErrDenied
	}
	c, err := persistence.New(s.pool).StaffReadCredential(ctx, actor.StaffID)
	if errors.Is(err, pgx.ErrNoRows) {
		if e := s.passwords.VerifyDummy(ctx, password); e != nil {
			return c, e
		}
		return c, ErrDenied
	}
	if err != nil {
		return c, safeError(err)
	}
	ok, err := s.passwords.Verify(ctx, c.PasswordHash, password)
	if err != nil {
		return c, err
	}
	if !ok {
		return c, ErrDenied
	}
	return c, nil
}

type lockedGovernance struct {
	actor, target Staff
	credentials   map[pgtype.UUID]persistence.PlatformStaffCredential
}

func lockGovernance(ctx context.Context, q *persistence.Queries, actor Actor, target pgtype.UUID, required string, at time.Time) (lockedGovernance, error) {
	var locked lockedGovernance
	if required != "superadmin" && required != "any_staff" {
		return locked, ErrDenied
	}
	if !actor.StaffID.Valid {
		return locked, ErrUnauthenticated
	}
	if _, err := q.StaffFence(ctx); err != nil {
		return locked, err
	}
	ids := []pgtype.UUID{actor.StaffID}
	if target.Valid && target != actor.StaffID {
		ids = append(ids, target)
	}
	rows, err := q.StaffLockAccounts(ctx, ids)
	if err != nil {
		return locked, err
	}
	for _, r := range rows {
		v := staffFromRow(r)
		if v.ID == actor.StaffID {
			locked.actor = v
		}
		if v.ID == target {
			locked.target = v
		}
	}
	if !locked.actor.ID.Valid || locked.actor.Status != "active" {
		return locked, ErrUnauthenticated
	}
	if required == "superadmin" && locked.actor.Role != "superadmin" {
		return locked, ErrDenied
	}
	credentials, err := q.StaffLockCredentials(ctx, ids)
	if err != nil {
		return locked, err
	}
	locked.credentials = make(map[pgtype.UUID]persistence.PlatformStaffCredential, len(credentials))
	for _, c := range credentials {
		locked.credentials[c.StaffID] = c
	}
	if _, err = q.StaffActiveSession(ctx, persistence.StaffActiveSessionParams{StaffID: actor.StaffID, TokenHash: actor.SessionHash[:], CreatedAt: stamp(at)}); errors.Is(err, pgx.ErrNoRows) {
		return locked, ErrUnauthenticated
	} else if err != nil {
		return locked, err
	}
	return locked, nil
}
func credentialMatches(a, b persistence.PlatformStaffCredential) bool {
	// Hash comparison is internal, never a revision/DTO. Timestamp checks bind
	// confirmation to the read credential version even when the PHC is unchanged.
	return a.StaffID.Valid && a.StaffID == b.StaffID && a.ChangedAt == b.ChangedAt && len(a.PasswordHash) == len(b.PasswordHash) && subtle.ConstantTimeCompare([]byte(a.PasswordHash), []byte(b.PasswordHash)) == 1
}

// AuthorizeMutation uses the caller's transaction; customer mutations must call it before
// locking its own parent/children. No cached role or browser actor is trusted.
func AuthorizeMutation(ctx context.Context, tx pgx.Tx, actor Actor, requiredRole string, at time.Time) (Staff, error) {
	locked, err := lockGovernance(ctx, persistence.New(tx), actor, pgtype.UUID{}, requiredRole, at)
	return locked.actor, safeError(err)
}

// AuthorizeMutationNow makes the fence-before-clock protocol indivisible for
// caller-owned customer transactions. The returned time is used for the whole write.
func AuthorizeMutationNow(ctx context.Context, tx pgx.Tx, actor Actor, requiredRole string, clock func() time.Time) (Staff, time.Time, error) {
	if clock == nil {
		return Staff{}, time.Time{}, ErrUnavailable
	}
	if _, err := persistence.New(tx).StaffFence(ctx); err != nil {
		return Staff{}, time.Time{}, safeError(err)
	}
	at := stamp(clock()).Time
	current, err := AuthorizeMutation(ctx, tx, actor, requiredRole, at)
	return current, at, err
}
