package customers

import (
	"context"
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
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	db "github.com/vandordev/vkit-zenith/packages/db"
)

type Service struct {
	pool  *pgxpool.Pool
	clock func() time.Time
	staff *staff.Service
}

func NewService(pool *pgxpool.Pool, clock func() time.Time, staffService *staff.Service) *Service {
	if pool == nil || clock == nil || staffService == nil {
		panic("customers: explicit same-runtime dependencies required")
	}
	return &Service{pool, clock, staffService}
}

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func text(v string, max int) bool {
	return utf8.ValidString(v) && !strings.ContainsRune(v, 0) && utf8.RuneCountInString(v) <= max && strings.TrimSpace(v) != ""
}
func validEmail(v string) bool {
	return utf8.ValidString(v) && !strings.ContainsRune(v, 0) && len(strings.Trim(v, " ")) <= 254 && emailPattern.MatchString(strings.Trim(v, " "))
}
func validUser(v UserCreate) bool { return validEmail(v.Email) && text(v.DisplayName, 200) }
func validOrganization(name, slug string) bool {
	return text(name, 200) && len(slug) <= 64 && slugPattern.MatchString(slug)
}
func stamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: canonical(t), Valid: true}
}
func safe(err error) error {
	if err == nil {
		return nil
	}
	var limited *staff.RateLimitError
	if errors.As(err, &limited) {
		return limited
	}
	for _, e := range []error{staff.ErrValidation, staff.ErrUnauthenticated, staff.ErrDenied, staff.ErrNotFound, staff.ErrConflict, staff.ErrUnavailable, staff.ErrOverloaded, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, e) {
			return e
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return staff.ErrNotFound
	}
	var p *pgconn.PgError
	if errors.As(err, &p) {
		switch p.Code {
		case "23505", "40001", "40P01", "55P03":
			return staff.ErrConflict
		case "23514":
			if p.ConstraintName == "organization_active_owner_check" {
				return staff.ErrConflict
			}
			return staff.ErrValidation
		case "23503", "22003", "22021":
			return staff.ErrValidation
		}
	}
	return staff.ErrUnavailable
}
func finish(ctx context.Context, tx pgx.Tx, fn func(*persistence.Queries) error) error {
	if err := fn(persistence.New(tx)); err != nil {
		return safe(err)
	}
	_, err := tx.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`)
	return safe(err)
}
func (s *Service) transaction(ctx context.Context, a staff.Actor, fn func(*persistence.Queries, time.Time) error) error {
	return safe(db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s';SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		_, at, err := staff.AuthorizeMutationNow(ctx, tx, a, "any_staff", s.clock)
		if err != nil {
			return err
		}
		return finish(ctx, tx, func(q *persistence.Queries) error { return fn(q, at) })
	}))
}

func (s *Service) confirmed(ctx context.Context, a staff.Actor, in Sensitive, fn func(*persistence.Queries, time.Time) error) error {
	if !text(in.Reason, 1024) || (in.CurrentPassword != "" && strings.Contains(in.Reason, in.CurrentPassword)) || strings.Contains(in.Reason, "$argon2") {
		return staff.ErrValidation
	}
	// The callback maps customer constraints before staff's opaque safe boundary.
	return s.staff.ConfirmedMutation(ctx, a, in.CurrentPassword, func(ctx context.Context, tx pgx.Tx, at time.Time) error {
		return finish(ctx, tx, func(q *persistence.Queries) error { return fn(q, at) })
	})
}
func user(r persistence.User) User {
	v := User{ID: r.ID, Email: r.Email, DisplayName: r.DisplayName, Status: r.Status, CreatedAt: r.CreatedAt.Time, UpdatedAt: r.UpdatedAt.Time}
	if r.EmailVerifiedAt.Valid {
		at := r.EmailVerifiedAt.Time
		v.EmailVerifiedAt = &at
	}
	return v
}
func organization(r persistence.Organization) Organization {
	return Organization{r.ID, r.Name, r.Slug, r.Status, r.CreatedAt.Time, r.UpdatedAt.Time}
}
func membership(r persistence.OrganizationMembership) Membership {
	return Membership{r.ID, r.OrganizationID, r.UserID, r.Status, Role(r.Role), r.CreatedAt.Time, r.UpdatedAt.Time}
}
