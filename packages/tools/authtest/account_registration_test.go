//go:build integration

package authtest

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
	"github.com/vandordev/vkit-zenith/packages/email"
)

type fixedIdentityClock struct{ at *time.Time }

func (c fixedIdentityClock) Now() time.Time { return *c.at }

type registeredDelivery struct {
	pool      *pgxpool.Pool
	messages  []identity.ActionDelivery
	committed bool
}

func (s *registeredDelivery) Send(ctx context.Context, m identity.ActionDelivery) error {
	var ok bool
	if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*)=1 FROM users WHERE email_normalized=$1 AND telephone='+6281234567890') AND (SELECT count(*)=1 FROM email_action_tokens WHERE user_id=(SELECT id FROM users WHERE email_normalized=$1) AND identity_version=1 AND purpose='verify_email') AND EXISTS(SELECT 1 FROM audit_events WHERE action='identity.register')`, m.Recipient).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return errors.New("registration delivery ran before durable user/token/audit commit")
	}
	s.committed = true
	s.messages = append(s.messages, m)
	return nil
}

var _ identity.DeliverySender = (*registeredDelivery)(nil)

func TestAuthAccountRegisterDuplicateAndPostCommit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stage := "fixture"
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
		sender := &registeredDelivery{pool: f.Pool}
		limiter := [32]byte{1}
		csrf := [32]byte{2}
		svc := identity.NewService(f.Pool, fixedIdentityClock{&at}, sender, limiter, csrf, "https://auth.example.test", false)
		input := identity.RegisterInput{DisplayName: "  Rani  ", Email: "  PERSON+tag@Example.TEST  ", Telephone: "+62 812-3456-7890", Password: "  Correct Horse Battery Staple  "}
		stage = "first-registration-command"
		first, err := svc.Register(ctx, input, identity.RequestMeta{Source: "127.0.0.1"})
		if err != nil {
			return err
		}
		if first.Accepted != (identity.PublicAccepted{}) || len(first.EnrollmentSecret) != 43 || len(first.SessionSecret) != 43 || !sender.committed || len(sender.messages) != 1 {
			return errors.New("registration response/delivery contract mismatch")
		}
		message := sender.messages[0]
		if message.Purpose != identity.ActionVerifyEmail || message.Recipient != "person+tag@example.test" || message.ExpiresAt.Sub(at) != identity.VerifyEmailLifetime || !strings.HasPrefix(message.URL, "https://auth.example.test/verify-email?token=") || len(message.AttemptID) < 16 {
			return errors.New("registration delivery fields are not normalized/bounded")
		}
		receiptHash := sha256.Sum256([]byte(first.EnrollmentSecret))
		sessionHash := sha256.Sum256([]byte(first.SessionSecret))
		var valid bool
		stage = "first-persisted-contract"
		if err = f.Pool.QueryRow(ctx, `SELECT
 (SELECT count(*)=1 AND bool_and(display_name='Rani' AND email='person+tag@example.test' AND telephone='+6281234567890' AND status='active' AND email_verified_at IS NULL AND identity_version=1) FROM users)
 AND (SELECT count(*)=1 AND bool_and(generation=1 AND password_hash LIKE '$argon2id$v=19$m=65536,t=3,p=1$%') FROM password_credentials)
 AND (SELECT count(*)=1 AND bool_and(kind='restricted' AND identity_version=1 AND credential_generation=1 AND secret_hash=$2 AND expires_at=$3::timestamptz+interval '30 minutes') FROM auth_sessions)
 AND (SELECT count(*)=1 AND bool_and(user_id IS NOT NULL AND identity_version=1 AND credential_generation=1 AND secret_hash=$1 AND expires_at=$3::timestamptz+interval '30 minutes') FROM identity_enrollment_receipts)
 AND (SELECT count(*)=1 AND bool_and(purpose='verify_email' AND identity_version=1 AND credential_generation=1 AND enrollment_hash=$1 AND expires_at=$3::timestamptz+interval '24 hours' AND octet_length(token_hash)=32) FROM email_action_tokens)
 AND NOT EXISTS(SELECT 1 FROM organization_memberships)`, receiptHash[:], sessionHash[:], at).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return errors.New("registration did not persist exact restricted/new-generation contract")
		}
		stage = "duplicate-registration-command"
		at = at.Add(61 * time.Second)
		second, err := svc.Register(ctx, input, identity.RequestMeta{Source: "127.0.0.1"})
		if err != nil {
			return err
		}
		if second.Accepted != (identity.PublicAccepted{}) || len(second.EnrollmentSecret) != 43 || len(second.SessionSecret) != 43 || second.EnrollmentSecret == first.EnrollmentSecret || second.SessionSecret == first.SessionSecret || len(sender.messages) != 1 {
			return errors.New("duplicate registration distinguished identity or delivered another action")
		}
		stage = "duplicate-receipt-binding"
		duplicateHash := sha256.Sum256([]byte(second.EnrollmentSecret))
		if err = f.Pool.QueryRow(ctx, `SELECT user_id IS NULL AND identity_version IS NULL AND credential_generation IS NULL FROM identity_enrollment_receipts WHERE secret_hash=$1`, duplicateHash[:]).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return errors.New("duplicate enrollment receipt gained account authority")
		}
		stage = "duplicate-did-not-mutate"
		var users, credentials, sessions, actions int
		if err = f.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM password_credentials),(SELECT count(*) FROM auth_sessions),(SELECT count(*) FROM email_action_tokens)`).Scan(&users, &credentials, &sessions, &actions); err != nil {
			return err
		}
		if users != 1 {
			return errors.New("duplicate registration changed account count")
		}
		if credentials != 1 {
			return errors.New("duplicate registration changed credential count")
		}
		if sessions != 1 {
			return errors.New("duplicate registration changed session count")
		}
		if actions != 1 {
			return errors.New("duplicate registration changed action count")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("owned PostgreSQL registration contract failed at %s; details withheld", stage)
	}
}

var _ email.Transport = (*email.Resend)(nil)
