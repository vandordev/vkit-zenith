package staff

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
)

const SessionLifetime = 720 * time.Hour

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	validEmail := utf8.ValidString(email) && len(email) <= 254 && emailPattern.MatchString(email)
	key := email
	if !validEmail {
		key = ""
	}
	a, err := s.admit(ctx, "login", key, input.Source)
	if err != nil {
		return LoginResult{}, err
	}
	fail := func(e error) (LoginResult, error) {
		if err := s.failure(ctx, a); err != nil {
			return LoginResult{}, err
		}
		return LoginResult{}, e
	}
	if !validEmail || !validPassword(input.Password) {
		return fail(ErrValidation)
	}
	q := persistence.New(s.pool)
	r, err := q.StaffLoginCredential(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		if err = s.passwords.VerifyDummy(ctx, input.Password); err != nil {
			return LoginResult{}, err
		}
		return fail(ErrUnauthenticated)
	}
	if err != nil {
		return LoginResult{}, safeError(err)
	}
	ok, err := s.passwords.Verify(ctx, r.PasswordHash, input.Password)
	if err != nil {
		return LoginResult{}, err
	}
	if !ok || r.Status != "active" {
		return fail(ErrUnauthenticated)
	}
	verified := persistence.PlatformStaffCredential{StaffID: r.ID, PasswordHash: r.PasswordHash, ChangedAt: r.ChangedAt}
	token, hash, err := NewSessionToken()
	if err != nil {
		return LoginResult{}, err
	}
	id, err := newID()
	if err != nil {
		return LoginResult{}, err
	}
	var result LoginResult
	err = s.transaction(ctx, func(q *persistence.Queries, _ pgx.Tx, at time.Time) error {
		rows, err := q.StaffLockAccounts(ctx, []pgtype.UUID{r.ID})
		if err != nil {
			return err
		}
		credentials, err := q.StaffLockCredentials(ctx, []pgtype.UUID{r.ID})
		if err != nil {
			return err
		}
		if len(rows) != 1 || len(credentials) != 1 || rows[0].Status != "active" || rows[0].EmailNormalized != email || !credentialMatches(verified, credentials[0]) {
			return ErrUnauthenticated
		}
		expires := at.Add(SessionLifetime)
		if err = q.StaffIssueSession(ctx, persistence.StaffIssueSessionParams{ID: id, StaffID: r.ID, TokenHash: hash[:], CreatedAt: stamp(at), ExpiresAt: stamp(expires)}); err != nil {
			return err
		}
		if err = audit(ctx, q, r.ID, r.ID, "staff.login", at, map[string]any{}); err != nil {
			return err
		}
		result = LoginResult{Staff: staffFromRow(rows[0]), Token: token, ExpiresAt: expires}
		return nil
	})
	if errors.Is(err, ErrUnauthenticated) {
		return fail(err)
	}
	if err != nil {
		return LoginResult{}, err
	}
	return result, nil
}
