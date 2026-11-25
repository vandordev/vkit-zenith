package customers

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"time"
)

func insertUser(ctx context.Context, q *persistence.Queries, a staff.Actor, in UserCreate, at time.Time) (User, error) {
	r, err := q.CustomerInsertUser(ctx, persistence.CustomerInsertUserParams{Email: in.Email, DisplayName: in.DisplayName, At: stamp(at)})
	if err != nil {
		return User{}, err
	}
	v := user(r)
	err = audit(ctx, q, a, v.ID, pgtype.UUID{}, "user", "customer.user.create", at, "", "email", "display_name", "status")
	return v, err
}
func (s *Service) CreateUser(ctx context.Context, a staff.Actor, in UserCreate) (User, error) {
	if !validUser(in) {
		return User{}, staff.ErrValidation
	}
	var v User
	err := s.transaction(ctx, a, func(q *persistence.Queries, at time.Time) error {
		var err error
		v, err = insertUser(ctx, q, a, in, at)
		return err
	})
	if err != nil {
		return User{}, err
	}
	return v, nil
}
func (s *Service) UpdateUser(ctx context.Context, a staff.Actor, id pgtype.UUID, in UserProfile) (User, error) {
	if !id.Valid || !text(in.DisplayName, 200) {
		return User{}, staff.ErrValidation
	}
	var v User
	err := s.transaction(ctx, a, func(q *persistence.Queries, at time.Time) error {
		old, err := lockUser(ctx, q, id)
		if err != nil {
			return err
		}
		if err = revision(UserRevision(old), in.Revision); err != nil {
			return err
		}
		if old.DisplayName == in.DisplayName {
			return staff.ErrConflict
		}
		r, err := q.CustomerUpdateUserProfile(ctx, persistence.CustomerUpdateUserProfileParams{ID: id, DisplayName: in.DisplayName, At: stamp(at)})
		if err != nil {
			return err
		}
		v = user(r)
		return audit(ctx, q, a, id, pgtype.UUID{}, "user", "customer.user.update", at, "", "display_name")
	})
	if err != nil {
		return User{}, err
	}
	return v, nil
}
