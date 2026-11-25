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

// InspectAction is read-only: scanner/GET-style access never consumes a token.
// Verification requires both the enrollment receipt and matching live
// restricted-login browser context; cross-browser links guide the user to login.
func (s *Service) InspectAction(ctx context.Context, in RedeemActionInput) (ActionInspection, error) {
	if in.Token == "" || len(in.Token) > 128 {
		return ActionInspection{}, ErrValidation
	}
	secret := secretDigest(in.Token)
	var row persistence.EmailActionToken
	if err := s.pool.QueryRow(ctx, `SELECT id,user_id,purpose,target_email_normalized,token_hash,expires_at,consumed_at,revoked_at,created_at,identity_version,credential_generation,enrollment_hash FROM email_action_tokens WHERE token_hash=$1`, secret[:]).Scan(&row.ID, &row.UserID, &row.Purpose, &row.TargetEmailNormalized, &row.TokenHash, &row.ExpiresAt, &row.ConsumedAt, &row.RevokedAt, &row.CreatedAt, &row.IdentityVersion, &row.CredentialGeneration, &row.EnrollmentHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ActionInspection{}, ErrNotFound
		}
		return ActionInspection{}, ErrUnavailable
	}
	if !row.IdentityVersion.Valid || !BeforeExpiry(identityStamp(s.clock.Now()), row.ExpiresAt.Time) || row.ConsumedAt.Valid || row.RevokedAt.Valid {
		return ActionInspection{}, ErrNotFound
	}
	var user struct {
		email, displayName, status string
		verified                   pgtype.Timestamptz
		version                    pgtype.Int8
	}
	if err := s.pool.QueryRow(ctx, `SELECT email_normalized,display_name,status,email_verified_at,identity_version FROM users WHERE id=$1`, row.UserID).Scan(&user.email, &user.displayName, &user.status, &user.verified, &user.version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ActionInspection{}, ErrNotFound
		}
		return ActionInspection{}, ErrUnavailable
	}
	legacyUncredentialedSetup := row.Purpose == string(ActionSetupPassword) && !user.version.Valid && row.IdentityVersion.Int64 == 1
	legacyCredentialReset := row.Purpose == string(ActionResetPassword) && !user.version.Valid && row.IdentityVersion.Int64 == 1
	if user.status != "active" || (!legacyUncredentialedSetup && !legacyCredentialReset && (!user.version.Valid || user.version.Int64 != row.IdentityVersion.Int64)) || user.email != row.TargetEmailNormalized {
		return ActionInspection{}, ErrNotFound
	}
	requiresLogin := false
	if row.Purpose == string(ActionVerifyEmail) {
		if !row.CredentialGeneration.Valid || !row.IdentityVersion.Valid || user.verified.Valid || len(row.EnrollmentHash) != 32 {
			return ActionInspection{}, ErrNotFound
		}
		if !hmac.Equal(in.EnrollmentHash[:], row.EnrollmentHash) {
			requiresLogin = true
		} else {
			var valid bool
			if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auth_sessions s JOIN password_credentials c ON c.user_id=s.user_id WHERE s.secret_hash=$1 AND s.user_id=$2 AND s.kind='restricted' AND s.identity_version=$3 AND s.credential_generation=$4 AND s.revoked_at IS NULL AND s.expires_at>$5 AND c.generation=$4)`, in.SessionHash[:], row.UserID, row.IdentityVersion, row.CredentialGeneration, identityStamp(s.clock.Now())).Scan(&valid); err != nil {
				return ActionInspection{}, ErrUnavailable
			}
			requiresLogin = !valid
		}
	} else if row.Purpose == string(ActionSetupPassword) {
		var credential bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM password_credentials WHERE user_id=$1)`, row.UserID).Scan(&credential); err != nil {
			return ActionInspection{}, ErrUnavailable
		}
		if credential || !row.CredentialGeneration.Valid || row.CredentialGeneration.Int64 != 0 {
			return ActionInspection{}, ErrNotFound
		}
	} else if row.Purpose == string(ActionResetPassword) {
		var credential bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM password_credentials WHERE user_id=$1 AND (generation=$2 OR generation IS NULL AND $2=1))`, row.UserID, row.CredentialGeneration).Scan(&credential); err != nil {
			return ActionInspection{}, ErrUnavailable
		}
		if !credential || !row.CredentialGeneration.Valid || row.CredentialGeneration.Int64 < 1 {
			return ActionInspection{}, ErrNotFound
		}
	} else {
		return ActionInspection{}, ErrNotFound
	}
	displayName := user.displayName
	if requiresLogin {
		displayName = ""
	}
	return ActionInspection{Purpose: ActionPurpose(row.Purpose), ExpiresAt: row.ExpiresAt.Time, RequiresLogin: requiresLogin, DisplayName: displayName}, nil
}

// RedeemAction performs explicit confirmation. Token inspection is read-only;
// password KDF work completes before the mutation transaction. Every binding is
// rechecked under the application-owned user-first lock protocol.
func (s *Service) RedeemAction(ctx context.Context, in RedeemActionInput) (ActionRedemption, error) {
	var result ActionRedemption
	phase := "mutation"
	if in.Token == "" || len(in.Token) > 128 {
		return result, ErrValidation
	}
	hash := secretDigest(in.Token)
	var candidate persistence.EmailActionToken
	lookup := persistence.New(s.pool)
	if err := ctx.Err(); err != nil {
		return result, err
	}
	row, err := lookup.IdentityFindEmailActionByHash(ctx, hash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return result, ErrNotFound
		}
		return result, ErrUnavailable
	}
	candidate = row
	purpose := ActionPurpose(candidate.Purpose)
	if purpose != ActionVerifyEmail && purpose != ActionSetupPassword && purpose != ActionResetPassword {
		return result, ErrNotFound
	}
	telephone := ""
	passwordHash := ""
	if purpose == ActionSetupPassword {
		telephone, err = NormalizeTelephone(in.Telephone)
		if err != nil {
			return result, err
		}
	}
	if purpose == ActionSetupPassword || purpose == ActionResetPassword {
		passwordHash, err = HashPassword(ctx, in.Password)
		if err != nil {
			return result, mapError(err)
		}
	}
	newSession := ""
	newCSRF := ""
	var newSessionHash [32]byte
	if purpose == ActionVerifyEmail {
		newSession, newSessionHash, err = cookieSecret()
		if err != nil {
			return result, err
		}
		newCSRF, err = s.NewSessionCSRF(SecretHash(newSessionHash))
		if err != nil {
			return result, err
		}
	}
	var auditEvent string
	var rotatedAt time.Time
	err = db.InTx(ctx, s.pool, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		phase = "begin-transaction"
		if _, err := tx.Exec(ctx, `SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'`); err != nil {
			return err
		}
		q := persistence.New(tx)
		phase = "lock-user-and-credential"
		user, err := q.IdentityLockUser(ctx, candidate.UserID)
		if err != nil {
			return err
		}
		credential, credentialErr := q.IdentityLockCredential(ctx, user.ID)
		if credentialErr != nil && !errors.Is(credentialErr, pgx.ErrNoRows) {
			return credentialErr
		}
		var session persistence.AuthSession
		var receipt persistence.IdentityEnrollmentReceipt
		if purpose == ActionVerifyEmail {
			phase = "lock-session-and-enrollment"
			session, err = q.IdentityFindSessionByHash(ctx, persistence.IdentityFindSessionByHashParams{SecretHash: in.SessionHash[:], UserID: user.ID})
			if err != nil {
				return ErrDenied
			}
			session, err = q.IdentityLockSession(ctx, session.ID)
			if err != nil {
				return err
			}
			receipt, err = q.IdentityLockEnrollmentReceipt(ctx, in.EnrollmentHash[:])
			if err != nil {
				return ErrDenied
			}
		}
		if purpose == ActionResetPassword || purpose == ActionSetupPassword {
			phase = "lock-credential-descendants"
			if err = lockCredentialDescendants(ctx, tx, user.ID); err != nil {
				return err
			}
		}
		actions, err := q.IdentityLockEmailActions(ctx, user.ID)
		phase = "lock-action-history"
		if err != nil {
			return err
		}
		var action persistence.EmailActionToken
		found := false
		for _, item := range actions {
			if hmac.Equal(item.TokenHash, hash[:]) {
				action = item
				found = true
			}
		}
		if !found {
			return ErrNotFound
		}
		at := identityStamp(s.clock.Now())
		rotatedAt = at
		if !BeforeExpiry(at, action.ExpiresAt.Time) || action.ConsumedAt.Valid || action.RevokedAt.Valid || user.Status != "active" || user.EmailNormalized != action.TargetEmailNormalized || !action.IdentityVersion.Valid {
			return ErrNotFound
		}
		version := int64(1)
		if user.IdentityVersion.Valid {
			version = user.IdentityVersion.Int64
		}
		if action.IdentityVersion.Int64 != version {
			return ErrNotFound
		}
		if purpose == ActionVerifyEmail {
			phase = "verify-current-bindings"
			if user.EmailVerifiedAt.Valid || credentialErr != nil || !credential.Generation.Valid || !action.CredentialGeneration.Valid || credential.Generation.Int64 != action.CredentialGeneration.Int64 || len(action.EnrollmentHash) != 32 || !hmac.Equal(action.EnrollmentHash, in.EnrollmentHash[:]) || !receipt.UserID.Valid || receipt.UserID != user.ID || !receipt.IdentityVersion.Valid || receipt.IdentityVersion.Int64 != version || !receipt.CredentialGeneration.Valid || receipt.CredentialGeneration.Int64 != credential.Generation.Int64 || !BeforeExpiry(at, receipt.ExpiresAt.Time) || !session.Kind.Valid || session.Kind.String != string(SessionRestricted) || !session.IdentityVersion.Valid || session.IdentityVersion.Int64 != version || !session.CredentialGeneration.Valid || session.CredentialGeneration.Int64 != credential.Generation.Int64 || session.RevokedAt.Valid || !BeforeExpiry(at, session.ExpiresAt.Time) {
				return ErrDenied
			}
			if err = q.IdentitySetVerification(ctx, persistence.IdentitySetVerificationParams{ID: user.ID, EmailNormalized: user.EmailNormalized, At: pgtype.Timestamptz{Time: at, Valid: true}}); err != nil {
				return err
			}
			command, err := q.IdentityRotateVerifiedSession(ctx, persistence.IdentityRotateVerifiedSessionParams{ID: session.ID, SecretHash: newSessionHash[:], LastActivityAt: pgtype.Timestamptz{Time: at, Valid: true}})
			if err != nil {
				return err
			}
			if command != 1 {
				return ErrConflict
			}
			auditEvent = "identity.email.verify"
			result = ActionRedemption{SessionSecret: newSession, CSRFToken: newCSRF, SessionExpiresAt: at.Add(SessionIdleLifetime)}
		} else if purpose == ActionSetupPassword {
			phase = "setup-current-bindings"
			if credentialErr == nil || !action.CredentialGeneration.Valid || action.CredentialGeneration.Int64 != 0 {
				return ErrConflict
			}
			if err = q.IdentitySetTelephone(ctx, persistence.IdentitySetTelephoneParams{ID: user.ID, Telephone: pgtype.Text{String: telephone, Valid: true}, At: pgtype.Timestamptz{Time: at, Valid: true}}); err != nil {
				return err
			}
			if err = q.IdentitySetCredential(ctx, persistence.IdentitySetCredentialParams{UserID: user.ID, PasswordHash: passwordHash, PasswordChangedAt: pgtype.Timestamptz{Time: at, Valid: true}, Generation: pgtype.Int8{Int64: 1, Valid: true}}); err != nil {
				return err
			}
			if err = q.IdentityRevokeRestrictedSessions(ctx, persistence.IdentityRevokeRestrictedSessionsParams{UserID: user.ID, RevokedAt: pgtype.Timestamptz{Time: at, Valid: true}}); err != nil {
				return err
			}
			auditEvent = "identity.password.setup"
			result = ActionRedemption{RequiresFreshLogin: true}
		} else {
			phase = "reset-current-bindings"
			if credentialErr != nil || !action.CredentialGeneration.Valid {
				return ErrConflict
			}
			currentGeneration := int64(1)
			if credential.Generation.Valid {
				currentGeneration = credential.Generation.Int64
			}
			if action.CredentialGeneration.Int64 != currentGeneration {
				return ErrConflict
			}
			if !user.IdentityVersion.Valid {
				if _, err = tx.Exec(ctx, `UPDATE users SET identity_version=1 WHERE id=$1 AND identity_version IS NULL`, user.ID); err != nil {
					return err
				}
			}
			if err = q.IdentitySetCredential(ctx, persistence.IdentitySetCredentialParams{UserID: user.ID, PasswordHash: passwordHash, PasswordChangedAt: pgtype.Timestamptz{Time: at, Valid: true}, Generation: pgtype.Int8{Int64: currentGeneration + 1, Valid: true}}); err != nil {
				return err
			}
			if err = revokeCredentialAuthority(ctx, q, user.ID, pgtype.Timestamptz{Time: at, Valid: true}); err != nil {
				return err
			}
			auditEvent = "identity.password.reset"
			result = ActionRedemption{RequiresFreshLogin: true}
		}
		if purpose == ActionSetupPassword || purpose == ActionResetPassword {
			phase = "revoke-competing-actions"
			if err = revokeCompetingCredentialActions(ctx, tx, user.ID, at, action.ID); err != nil {
				return err
			}
		}
		if purpose == ActionVerifyEmail {
			if err = q.IdentityRevokeOtherPurposeActions(ctx, persistence.IdentityRevokeOtherPurposeActionsParams{UserID: user.ID, Purpose: string(ActionVerifyEmail), RevokedAt: pgtype.Timestamptz{Time: at, Valid: true}, ID: action.ID}); err != nil {
				return err
			}
		}
		command, err := tx.Exec(ctx, `UPDATE email_action_tokens SET consumed_at=$2 WHERE id=$1 AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$2`, action.ID, at)
		phase = "consume-token"
		if err != nil {
			return err
		}
		if command.RowsAffected() != 1 {
			return ErrConflict
		}
		detail := auditDetail("email_verified_at")
		if purpose == ActionSetupPassword {
			detail = auditDetail("credentials", "telephone", "email_verified_at", "identity_version")
		}
		if purpose == ActionResetPassword {
			detail = auditDetail("credentials", "sessions", "tokens")
		}
		id, err := auditID()
		phase = "write-success-audit"
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO audit_events(id,occurred_at,actor_kind,action,outcome,target_type,target_id,detail) VALUES($1,$2,'anonymous',$3,'success','user',$4,$5::jsonb)`, id, at, auditEvent, user.ID.String(), detail)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `SET CONSTRAINTS ALL IMMEDIATE`)
		phase = "deferred-constraints"
		return err
	})
	if err != nil {
		return ActionRedemption{}, mutationFailure(phase, err)
	}
	if purpose == ActionVerifyEmail {
		result.CookieMaxAge = int(result.SessionExpiresAt.Sub(rotatedAt).Seconds())
	}
	return result, nil
}

func credentialErrIsNil(err error) bool { return err == nil }
