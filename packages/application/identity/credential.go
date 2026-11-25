package identity

import (
	"context"
	"crypto/hmac"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	db "github.com/vandordev/vkit-zenith/packages/db"
)

// lockCredentialDescendants implements the user → credential → session →
// action row prefix. Each
// sequence is ordered by stable ID; SELECT is only a fence, never authority.
func lockCredentialDescendants(ctx context.Context, tx pgx.Tx, userID pgtype.UUID) error {
	for _, query := range []string{
		`SELECT id FROM auth_sessions WHERE user_id=$1 ORDER BY id FOR UPDATE`,
	} {
		rows, err := tx.Query(ctx, query, userID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id pgtype.UUID
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// revokeCredentialAuthority is shared by reset and authenticated credential
// change. The caller has already acquired the complete descendant lock chain.
func revokeCredentialAuthority(ctx context.Context, q *persistence.Queries, userID pgtype.UUID, at pgtype.Timestamptz) error {
	if err := q.IdentityRevokeAllSessions(ctx, persistence.IdentityRevokeAllSessionsParams{UserID: userID, RevokedAt: at}); err != nil {
		return err
	}
	return nil
}

func revokeCompetingCredentialActions(ctx context.Context, tx pgx.Tx, userID pgtype.UUID, at time.Time, acceptedID pgtype.UUID) error {
	_, err := tx.Exec(ctx, `UPDATE email_action_tokens SET revoked_at=greatest(created_at,$2) WHERE user_id=$1 AND purpose IN ('setup_password','reset_password') AND id<>$3 AND consumed_at IS NULL AND revoked_at IS NULL`, userID, at, acceptedID)
	return err
}

// ChangePassword is the customer-security internal contract for an explicit
// authenticated credential change. It requires a live full session and the
// current password, then atomically rotates generation and revokes every old
// session/action. No public settings endpoint is installed here.
func (s *Service) ChangePassword(ctx context.Context, in ChangePasswordInput, meta RequestMeta) error {
	if s == nil || s.pool == nil || s.clock == nil || in.SessionHash == ([32]byte{}) || !validPassword(in.CurrentPassword) || !validPassword(in.NewPassword) {
		return ErrValidation
	}
	view, err := s.Current(ctx, in.SessionHash)
	if err != nil {
		return err
	}
	if view.Kind != SessionFull {
		return ErrDenied
	}
	suppressed, err := s.admission(ctx, "password-change", "credential_change", meta, view.Email, 5, time.Hour, time.Minute, 30, time.Hour)
	if err != nil {
		return err
	}
	if suppressed {
		return ErrOverloaded
	}
	owner, err := persistence.New(s.pool).IdentityFindSessionOwnerByHash(ctx, in.SessionHash[:])
	if err != nil {
		return ErrUnauthenticated
	}
	var before persistence.PasswordCredential
	if err = s.pool.QueryRow(ctx, `SELECT user_id,password_hash,password_changed_at,generation FROM password_credentials WHERE user_id=$1`, owner.UserID).Scan(&before.UserID, &before.PasswordHash, &before.PasswordChangedAt, &before.Generation); err != nil || !before.Generation.Valid {
		return ErrUnauthenticated
	}
	valid, err := verifyIdentityPassword(ctx, before.PasswordHash, in.CurrentPassword)
	if errors.Is(err, ErrOverloaded) {
		return err
	}
	if err != nil {
		return mapError(err)
	}
	if !valid {
		return ErrUnauthenticated
	}
	newHash, err := HashPassword(ctx, in.NewPassword)
	if err != nil {
		return mapError(err)
	}
	phase := "begin-transaction"
	err = db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		phase = "lock-user-credential"
		user, err := q.IdentityLockUser(ctx, owner.UserID)
		if err != nil {
			return err
		}
		credential, err := q.IdentityLockCredential(ctx, user.ID)
		if err != nil {
			return err
		}
		phase = "lock-descendants"
		if err = lockCredentialDescendants(ctx, tx, user.ID); err != nil {
			return err
		}
		session, err := q.IdentityFindSessionByHash(ctx, persistence.IdentityFindSessionByHashParams{SecretHash: in.SessionHash[:], UserID: user.ID})
		if err != nil {
			return ErrUnauthenticated
		}
		session, err = q.IdentityLockSession(ctx, session.ID)
		if err != nil {
			return err
		}
		at := identityStamp(s.clock.Now())
		if user.Status != "active" || !user.EmailVerifiedAt.Valid || !user.IdentityVersion.Valid || !credential.Generation.Valid || credential.Generation.Int64 != before.Generation.Int64 || !hmac.Equal([]byte(credential.PasswordHash), []byte(before.PasswordHash)) || !session.Kind.Valid || session.Kind.String != string(SessionFull) || !session.IdentityVersion.Valid || session.IdentityVersion.Int64 != user.IdentityVersion.Int64 || !session.CredentialGeneration.Valid || session.CredentialGeneration.Int64 != credential.Generation.Int64 || session.RevokedAt.Valid || !BeforeExpiry(at, session.ExpiresAt.Time) {
			return ErrUnauthenticated
		}
		atValue := pgtype.Timestamptz{Time: at, Valid: true}
		phase = "set-credential"
		if err = q.IdentitySetCredential(ctx, persistence.IdentitySetCredentialParams{UserID: user.ID, PasswordHash: newHash, PasswordChangedAt: atValue, Generation: pgtype.Int8{Int64: credential.Generation.Int64 + 1, Valid: true}}); err != nil {
			return err
		}
		phase = "revoke-authority"
		if err = revokeCredentialAuthority(ctx, q, user.ID, atValue); err != nil {
			return err
		}
		phase = "revoke-actions"
		if _, err = tx.Exec(ctx, `UPDATE email_action_tokens SET revoked_at=greatest(created_at,$2) WHERE user_id=$1 AND purpose IN ('setup_password','reset_password') AND consumed_at IS NULL AND revoked_at IS NULL`, user.ID, at); err != nil {
			return err
		}
		id, err := auditID()
		if err != nil {
			return err
		}
		phase = "audit-success"
		if _, err = tx.Exec(ctx, `INSERT INTO audit_events(id,occurred_at,actor_kind,actor_user_id,action,outcome,target_type,target_id,detail) VALUES($1,$2,'user',$3,$4,'success','user',$5,$6::jsonb)`, id, at, user.ID, "identity.password.change", user.ID.String(), auditDetail("credentials", "sessions", "tokens")); err != nil {
			return err
		}
		phase = "deferred-constraints"
		_, err = tx.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`)
		return err
	})
	if err != nil {
		return mutationFailure(phase, err)
	}
	return nil
}
