//go:build integration

package authtest

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type actionCapture struct{ messages []identity.ActionDelivery }

func (s *actionCapture) Send(_ context.Context, m identity.ActionDelivery) error {
	s.messages = append(s.messages, m)
	return nil
}

type failingPostCommitActionSender struct {
	pool              *pgxpool.Pool
	called, committed bool
}

func (s *failingPostCommitActionSender) Send(ctx context.Context, item identity.ActionDelivery) error {
	s.called = true
	parsed, err := url.Parse(item.URL)
	if err != nil {
		return errors.New("fake post-commit link malformed")
	}
	hash := sha256.Sum256([]byte(parsed.Query().Get("token")))
	if err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM email_action_tokens WHERE token_hash=$1 AND consumed_at IS NULL AND revoked_at IS NULL)`, hash[:]).Scan(&s.committed); err != nil {
		return err
	}
	return errors.New("fake provider timeout after commit")
}

func TestAuthAccountRequestActionEnumerationAndEligibility(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stage := "fixture"
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES('10000000-0000-4000-8000-000000000001','Setup@example.test','setup@example.test','Setup customer','active',$1,$1),('10000000-0000-4000-8000-000000000002','Disabled@example.test','disabled@example.test','Disabled','disabled',$1,$1)`, at); err != nil {
			return err
		}
		capture := &actionCapture{}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, capture, [32]byte{5}, [32]byte{6}, "https://auth.example.test", false)
		accepted, err := svc.RequestPasswordAction(ctx, "SETUP@example.test", identity.RequestMeta{Source: "203.0.113.10"})
		if err != nil || accepted != (identity.PublicAccepted{}) {
			return errors.New("eligible passwordless request not generically accepted")
		}
		if len(capture.messages) != 1 || capture.messages[0].Purpose != identity.ActionSetupPassword || capture.messages[0].ExpiresAt.Sub(at) != identity.SetupPasswordLifetime {
			return errors.New("passwordless account did not receive first-password action")
		}
		setupToken := strings.TrimPrefix(capture.messages[0].URL, "https://auth.example.test/set-password?token=")
		inspection, err := svc.InspectAction(ctx, identity.RedeemActionInput{Token: setupToken})
		if err != nil || inspection.Purpose != identity.ActionSetupPassword || inspection.DisplayName != "Setup customer" {
			return errors.New("read-only setup inspection unavailable")
		}
		tokenHash := sha256.Sum256([]byte(setupToken))
		var unconsumed bool
		if err = f.Pool.QueryRow(ctx, `SELECT consumed_at IS NULL FROM email_action_tokens WHERE token_hash=$1`, tokenHash[:]).Scan(&unconsumed); err != nil {
			return err
		}
		if !unconsumed {
			return errors.New("inspection consumed action token")
		}
		before := len(capture.messages)
		for _, emailAddress := range []string{"unknown@example.test", "disabled@example.test", "setup@example.test"} {
			at = at.Add(61 * time.Second)
			result, err := svc.RequestPasswordAction(ctx, emailAddress, identity.RequestMeta{Source: "203.0.113.10"})
			if err != nil || result != (identity.PublicAccepted{}) {
				return errors.New("ineligible/passwordless request changed generic response")
			}
		}
		if len(capture.messages) != before+1 {
			return errors.New("unknown/disabled leaked eligibility or credential presence")
		}
		var state string
		if err = f.Pool.QueryRow(ctx, `SELECT purpose FROM email_action_tokens WHERE user_id='10000000-0000-4000-8000-000000000001' ORDER BY created_at DESC LIMIT 1`).Scan(&state); err != nil {
			return err
		}
		if state != "setup_password" {
			return errors.New("passwordless recovery silently reset/provisioned credential")
		}
		if _, err = f.Pool.Exec(ctx, `UPDATE users SET identity_version=1 WHERE id='10000000-0000-4000-8000-000000000001'`); err != nil {
			return err
		}
		if _, err = f.Pool.Exec(ctx, `INSERT INTO password_credentials(user_id,password_hash,password_changed_at,generation) VALUES('10000000-0000-4000-8000-000000000001','fixture current hash',$1,1)`, at); err != nil {
			return err
		}
		at = at.Add(61 * time.Second)
		actionResult, err := svc.RequestPasswordAction(ctx, "setup@example.test", identity.RequestMeta{Source: "203.0.113.10"})
		if err != nil || actionResult != (identity.PublicAccepted{}) {
			return errors.New("eligible reset request not generically accepted")
		}
		last := capture.messages[len(capture.messages)-1]
		if last.Purpose != identity.ActionResetPassword || last.ExpiresAt.Sub(at) != identity.ResetPasswordLifetime {
			return errors.New("credential account did not receive bounded reset action")
		}
		stage = "manual-verification-setup"
		manualVerifiedAt := at
		if _, err = f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,email_verified_at,created_at,updated_at,identity_version) VALUES('10000000-0000-4000-8000-000000000003','manual-verified@example.test','manual-verified@example.test','Manual verified','active',$1,$1,$1,1)`, manualVerifiedAt); err != nil {
			var pg *pgconn.PgError
			if errors.As(err, &pg) {
				stage = fmt.Sprintf("manual-user-code-%s-constraint-%s", pg.Code, pg.ConstraintName)
			} else {
				stage = fmt.Sprintf("manual-user-error-%T", err)
			}
			return errors.New("manual-verified account fixture insert failed")
		}
		at = at.Add(61 * time.Second)
		stage = "manual-verified-request"
		if _, err = svc.RequestPasswordAction(ctx, "manual-verified@example.test", identity.RequestMeta{Source: "203.0.113.10"}); err != nil {
			return err
		}
		setupMail, err := url.Parse(capture.messages[len(capture.messages)-1].URL)
		if err != nil {
			return errors.New("manual-verified setup URL malformed")
		}
		stage = "manual-verified-redemption"
		setupResult, err := svc.RedeemAction(ctx, identity.RedeemActionInput{Token: setupMail.Query().Get("token"), Password: "Manual Verification Setup Password", Telephone: "+14155552672"})
		if err != nil || !setupResult.RequiresFreshLogin {
			stage = fmt.Sprintf("manual-verified-redeem-v%t-c%t-n%t-d%t-u%t", errors.Is(err, identity.ErrValidation), errors.Is(err, identity.ErrConflict), errors.Is(err, identity.ErrNotFound), errors.Is(err, identity.ErrDenied), errors.Is(err, identity.ErrUnavailable))
			return errors.New("manual-verified user could not complete first-password setup")
		}
		var preserved bool
		stage = "manual-verified-preserved-query"
		if err = f.Pool.QueryRow(ctx, `SELECT u.email_verified_at=$1::timestamptz AND c.generation=1 FROM users u JOIN password_credentials c ON c.user_id=u.id WHERE u.email_normalized='manual-verified@example.test'`, manualVerifiedAt).Scan(&preserved); err != nil {
			var pg *pgconn.PgError
			if errors.As(err, &pg) {
				stage = fmt.Sprintf("manual-verified-state-code-%s", pg.Code)
			} else {
				stage = fmt.Sprintf("manual-verified-state-error-%T", err)
			}
			return err
		}
		if !preserved {
			return errors.New("first-password setup overwrote existing manual verification time")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("owned PostgreSQL action-request contract failed at %s; details withheld", stage)
	}
}

func TestAuthAccountActionDeliveryFailureRetainsCommittedAcceptance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at,identity_version) VALUES('10000000-0000-4000-8000-000000000106','delivery-failure@example.test','delivery-failure@example.test','Delivery failure','active',$1,$1,1)`, at); err != nil {
			return err
		}
		sender := &failingPostCommitActionSender{pool: f.Pool}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, sender, [32]byte{25}, [32]byte{26}, "https://auth.example.test", false)
		accepted, err := svc.RequestPasswordAction(ctx, "delivery-failure@example.test", identity.RequestMeta{Source: "203.0.113.82"})
		if err != nil || accepted != (identity.PublicAccepted{}) || !sender.called || !sender.committed {
			return errors.New("provider failure changed generic post-commit request acceptance")
		}
		var committed int
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens WHERE user_id='10000000-0000-4000-8000-000000000106' AND purpose='setup_password' AND consumed_at IS NULL AND revoked_at IS NULL`).Scan(&committed); err != nil {
			return err
		}
		if committed != 1 {
			return errors.New("provider failure rolled back committed setup action")
		}
		return nil
	})
	if err != nil {
		t.Fatal("owned post-commit delivery-failure boundary failed; details withheld")
	}
}

type actionDeliveryCapture struct{ items []identity.ActionDelivery }

func (c *actionDeliveryCapture) Send(_ context.Context, item identity.ActionDelivery) error {
	c.items = append(c.items, item)
	return nil
}

func TestAuthAccountRegistrationDeliveryFailureIsPostCommit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		sender := &failingPostCommitActionSender{pool: f.Pool}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, sender, [32]byte{27}, [32]byte{28}, "https://auth.example.test", false)
		result, err := svc.Register(ctx, identity.RegisterInput{DisplayName: "Post commit", Email: "registration-delivery-failure@example.test", Telephone: "+14155552675", Password: "A Long Registration Password 123"}, identity.RequestMeta{Source: "203.0.113.83"})
		if err != nil || result.Accepted != (identity.PublicAccepted{}) || len(result.EnrollmentSecret) != 43 || len(result.SessionSecret) != 43 || !sender.called || !sender.committed {
			return errors.New("provider failure changed generic post-commit registration response")
		}
		var committed bool
		if err = f.Pool.QueryRow(ctx, `SELECT u.status='active' AND u.email_verified_at IS NULL AND c.generation=1 AND s.kind='restricted' AND e.consumed_at IS NULL AND EXISTS(SELECT 1 FROM audit_events a WHERE a.action='identity.register' AND a.target_id=u.id::text) FROM users u JOIN password_credentials c ON c.user_id=u.id JOIN auth_sessions s ON s.user_id=u.id JOIN email_action_tokens e ON e.user_id=u.id WHERE u.email_normalized='registration-delivery-failure@example.test'`).Scan(&committed); err != nil {
			return err
		}
		if !committed {
			return errors.New("mail failure rolled back new unverified customer state")
		}
		return nil
	})
	if err != nil {
		t.Fatal("owned registration delivery-failure boundary failed; details withheld")
	}
}

func TestAuthAccountVerifyRotatesRestrictedSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stage := "fixture"
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		clock := fixedIdentityClock{at: &at}
		sender := &actionDeliveryCapture{}
		svc := identity.NewService(f.Pool, clock, sender, [32]byte{7}, [32]byte{8}, "https://auth.example.test", false)
		registered, err := svc.Register(ctx, identity.RegisterInput{DisplayName: "Verifier", Email: "verify@example.test", Telephone: "+14155552671", Password: "A long correct horse battery"}, identity.RequestMeta{Source: "203.0.113.11"})
		if err != nil {
			return err
		}
		if len(sender.items) != 1 {
			return errors.New("verification action was not delivered")
		}
		parsed, err := url.Parse(sender.items[0].URL)
		if err != nil {
			return errors.New("verification URL malformed")
		}
		input := identity.RedeemActionInput{Token: parsed.Query().Get("token"), EnrollmentHash: identity.HashSecret(registered.EnrollmentSecret), SessionHash: identity.HashSecret(registered.SessionSecret)}
		stage = "same-browser-inspection"
		view, err := svc.InspectAction(ctx, input)
		if err != nil || view.RequiresLogin {
			return errors.New("same-browser verification binding rejected")
		}
		var consumed bool
		tokenHash := sha256.Sum256([]byte(input.Token))
		if err = f.Pool.QueryRow(ctx, `SELECT consumed_at IS NOT NULL FROM email_action_tokens WHERE token_hash=$1`, tokenHash[:]).Scan(&consumed); err != nil {
			return err
		}
		if consumed {
			return errors.New("GET inspection consumed verification token")
		}
		stage = "verification-expiry"
		originalIssue := at
		at = sender.items[0].ExpiresAt.Add(-time.Microsecond)
		if _, err = svc.InspectAction(ctx, input); err != nil {
			return errors.New("verification action rejected immediately before its exact deadline")
		}
		at = sender.items[0].ExpiresAt
		if _, err = svc.InspectAction(ctx, input); !errors.Is(err, identity.ErrNotFound) {
			return errors.New("verification action remained valid at its exact deadline")
		}
		at = originalIssue
		type redemptionResult struct {
			value identity.ActionRedemption
			err   error
		}
		stage = "concurrent-redemption"
		results := make([]redemptionResult, 2)
		var wg sync.WaitGroup
		for i := range results {
			wg.Add(1)
			go func(i int) { defer wg.Done(); results[i].value, results[i].err = svc.RedeemAction(ctx, input) }(i)
		}
		wg.Wait()
		var result identity.ActionRedemption
		winners := 0
		for _, attempt := range results {
			if attempt.err == nil {
				winners++
				result = attempt.value
				continue
			}
			if !errors.Is(attempt.err, identity.ErrNotFound) && !errors.Is(attempt.err, identity.ErrDenied) {
				var failure *identity.MutationFailure
				if errors.As(attempt.err, &failure) {
					return fmt.Errorf("verify mutation phase %s validation=%t conflict=%t unavailable=%t", failure.Phase(), errors.Is(attempt.err, identity.ErrValidation), errors.Is(attempt.err, identity.ErrConflict), errors.Is(attempt.err, identity.ErrUnavailable))
				}
				return errors.New("verification race failed with an unexpected error")
			}
		}
		if winners != 1 {
			return errors.New("verification token did not produce exactly one concurrent winner")
		}
		stage = "rotated-result"
		if result.RequiresFreshLogin || len(result.SessionSecret) != 43 || len(result.CSRFToken) != 96 || !result.SessionExpiresAt.Equal(at.Add(identity.SessionIdleLifetime)) {
			return errors.New("verified session did not rotate to bounded full session")
		}
		var valid bool
		rotatedHash := identity.HashSecret(result.SessionSecret)
		stage = "persisted-verified-session"
		if err = f.Pool.QueryRow(ctx, `SELECT u.email_verified_at=$2 AND s.kind='full' AND s.authenticated_at=$2 AND s.last_activity_at=$2 AND s.expires_at=$2::timestamptz+interval '1176 hours' AND s.secret_hash=$1 AND a.consumed_at=$2 FROM users u JOIN auth_sessions s ON s.user_id=u.id JOIN email_action_tokens a ON a.user_id=u.id WHERE u.email_normalized='verify@example.test'`, rotatedHash[:], at).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return errors.New("full session/email verification/token consumption not atomic")
		}
		var oldExists bool
		oldHash := identity.HashSecret(registered.SessionSecret)
		if err = f.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auth_sessions WHERE secret_hash=$1 AND kind='full')`, oldHash[:]).Scan(&oldExists); err != nil {
			return err
		}
		if oldExists {
			return errors.New("old restricted secret retained full authority")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("owned verification transition failed at %s; raw details withheld", stage)
	}
}

func TestAuthAccountActionExactExpiryBoundaries(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		start := at
		capture := &actionDeliveryCapture{}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, capture, [32]byte{10}, [32]byte{11}, "https://auth.example.test", false)
		if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at,identity_version) VALUES('10000000-0000-4000-8000-000000000101','setup-expiry@example.test','setup-expiry@example.test','Setup expiry','active',$1,$1,1),('10000000-0000-4000-8000-000000000102','reset-expiry@example.test','reset-expiry@example.test','Reset expiry','active',$1,$1,1)`, start); err != nil {
			return err
		}
		if _, err := f.Pool.Exec(ctx, `INSERT INTO password_credentials(user_id,password_hash,password_changed_at,generation) VALUES('10000000-0000-4000-8000-000000000102','fixture current hash',$1,1)`, start); err != nil {
			return err
		}
		for _, tc := range []struct {
			email   string
			purpose identity.ActionPurpose
			life    time.Duration
		}{
			{email: "setup-expiry@example.test", purpose: identity.ActionSetupPassword, life: identity.SetupPasswordLifetime},
			{email: "reset-expiry@example.test", purpose: identity.ActionResetPassword, life: identity.ResetPasswordLifetime},
		} {
			at = start
			if _, err := svc.RequestPasswordAction(ctx, tc.email, identity.RequestMeta{Source: "203.0.113.77"}); err != nil {
				return err
			}
			if len(capture.items) == 0 || capture.items[len(capture.items)-1].Purpose != tc.purpose || !capture.items[len(capture.items)-1].ExpiresAt.Equal(start.Add(tc.life)) {
				return errors.New("action purpose/lifetime did not match approved elapsed duration")
			}
			item := capture.items[len(capture.items)-1]
			parsed, err := url.Parse(item.URL)
			if err != nil {
				return errors.New("captured action URL malformed")
			}
			in := identity.RedeemActionInput{Token: parsed.Query().Get("token")}
			at = start.Add(tc.life).Add(-time.Microsecond)
			if _, err = svc.InspectAction(ctx, in); err != nil {
				return errors.New("action rejected immediately before exact deadline")
			}
			at = start.Add(tc.life)
			if _, err = svc.InspectAction(ctx, in); !errors.Is(err, identity.ErrNotFound) {
				return errors.New("action remained valid at exact deadline")
			}
			if tc.purpose == identity.ActionResetPassword {
				if _, err = svc.RedeemAction(ctx, identity.RedeemActionInput{Token: in.Token, Password: "A Valid Reset Password Is Long"}); !errors.Is(err, identity.ErrNotFound) {
					return errors.New("reset redemption accepted at its exact deadline")
				}
			}
			at = at.Add(time.Microsecond)
			if _, err = svc.InspectAction(ctx, in); !errors.Is(err, identity.ErrNotFound) {
				return errors.New("expired action became valid after deadline")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal("owned action expiry boundary failed; details withheld")
	}
}

func TestAuthAccountSetupRollbackAndConcurrentSingleWinner(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stage := "fixture"
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		capture := &actionDeliveryCapture{}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, capture, [32]byte{12}, [32]byte{13}, "https://auth.example.test", false)
		stage = "setup-seed-user"
		if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at,identity_version) VALUES('10000000-0000-4000-8000-000000000103','setup-race@example.test','setup-race@example.test','Setup race','active',$1,$1,1)`, at); err != nil {
			return err
		}
		stage = "setup-seed-organization"
		seedTx, err := f.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = seedTx.Exec(ctx, `INSERT INTO organizations(id,name,slug,status,created_at,updated_at) VALUES('20000000-0000-4000-8000-000000000103','Existing access','existing-access','active',$1,$1)`, at); err != nil {
			_ = seedTx.Rollback(ctx)
			return err
		}
		if _, err = seedTx.Exec(ctx, `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES('20000000-0000-4000-8000-000000000104','20000000-0000-4000-8000-000000000103','10000000-0000-4000-8000-000000000103','active','owner',$1,$1)`, at); err != nil {
			_ = seedTx.Rollback(ctx)
			return err
		}
		if err = seedTx.Commit(ctx); err != nil {
			return err
		}
		stage = "setup-request"
		if _, err := svc.RequestPasswordAction(ctx, "setup-race@example.test", identity.RequestMeta{Source: "203.0.113.78"}); err != nil {
			return err
		}
		parsed, err := url.Parse(capture.items[0].URL)
		if err != nil {
			return errors.New("setup link malformed")
		}
		input := identity.RedeemActionInput{Token: parsed.Query().Get("token"), Password: "A Correct Setup Password 123", Telephone: "+14155552671"}
		if _, err = f.Pool.Exec(ctx, `CREATE FUNCTION auth_test_fail_setup_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action='identity.password.setup' THEN RAISE EXCEPTION 'fixture audit rejection' USING ERRCODE='XX000'; END IF; RETURN NEW; END $$`); err != nil {
			return err
		}
		if _, err = f.Pool.Exec(ctx, `CREATE TRIGGER auth_test_fail_setup_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION auth_test_fail_setup_audit()`); err != nil {
			return err
		}
		stage = "setup-audit-rollback"
		if _, err = svc.RedeemAction(ctx, input); !errors.Is(err, identity.ErrUnavailable) {
			return errors.New("failed setup audit was not surfaced as safe unavailable")
		}
		var unchanged bool
		if err = f.Pool.QueryRow(ctx, `SELECT u.telephone IS NULL AND u.email_verified_at IS NULL AND NOT EXISTS(SELECT 1 FROM password_credentials c WHERE c.user_id=u.id) AND a.consumed_at IS NULL FROM users u JOIN email_action_tokens a ON a.user_id=u.id WHERE u.email_normalized='setup-race@example.test'`).Scan(&unchanged); err != nil {
			return err
		}
		if !unchanged {
			return errors.New("audit failure committed setup state or consumed token")
		}
		if _, err = f.Pool.Exec(ctx, `DROP TRIGGER auth_test_fail_setup_audit ON audit_events`); err != nil {
			return err
		}
		if _, err = f.Pool.Exec(ctx, `DROP FUNCTION auth_test_fail_setup_audit()`); err != nil {
			return err
		}
		stage = "setup-concurrent-redemption"
		const attempts = 2
		results := make([]identity.ActionRedemption, attempts)
		errorsOut := make([]error, attempts)
		var wg sync.WaitGroup
		for i := range results {
			wg.Add(1)
			go func(i int) { defer wg.Done(); results[i], errorsOut[i] = svc.RedeemAction(ctx, input) }(i)
		}
		wg.Wait()
		winners := 0
		for i, result := range results {
			if errorsOut[i] == nil {
				if !result.RequiresFreshLogin {
					return errors.New("setup produced an automatic session")
				}
				winners++
				continue
			}
			if !errors.Is(errorsOut[i], identity.ErrConflict) && !errors.Is(errorsOut[i], identity.ErrNotFound) {
				return errors.New("losing concurrent setup returned an unsafe error")
			}
		}
		if winners != 1 {
			return errors.New("setup action did not produce exactly one concurrent winner")
		}
		var completed bool
		if err = f.Pool.QueryRow(ctx, `SELECT u.telephone='+14155552671' AND u.email_verified_at IS NOT NULL AND c.generation=1 AND a.consumed_at IS NOT NULL AND EXISTS(SELECT 1 FROM organization_memberships m WHERE m.user_id=u.id AND m.organization_id='20000000-0000-4000-8000-000000000103' AND m.status='active' AND m.role='owner') AND (SELECT count(*)=1 FROM audit_events WHERE action='identity.password.setup') FROM users u JOIN password_credentials c ON c.user_id=u.id JOIN email_action_tokens a ON a.user_id=u.id WHERE u.email_normalized='setup-race@example.test'`).Scan(&completed); err != nil {
			return err
		}
		if !completed {
			return errors.New("winning setup did not atomically persist credential, phone, verification, consumption and audit")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("owned setup rollback/concurrency contract failed at %s; details withheld", stage)
	}
}

func TestAuthAccountLiveActionLimitSuppressesCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at,identity_version) VALUES('10000000-0000-4000-8000-000000000104','action-cap@example.test','action-cap@example.test','Action cap','active',$1,$1,1)`, at); err != nil {
			return err
		}
		if _, err := f.Pool.Exec(ctx, `INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation) SELECT gen_random_uuid(),'10000000-0000-4000-8000-000000000104','setup_password','action-cap@example.test',sha256(int4send(n)),$1,$1::timestamptz+interval '24 hours',1,0 FROM generate_series(1,10) n`, at); err != nil {
			return err
		}
		capture := &actionCapture{}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, capture, [32]byte{16}, [32]byte{17}, "https://auth.example.test", false)
		accepted, err := svc.RequestPasswordAction(ctx, "action-cap@example.test", identity.RequestMeta{Source: "203.0.113.80"})
		if err != nil || accepted != (identity.PublicAccepted{}) {
			return errors.New("action-cap response was not generic accepted")
		}
		var live int
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens WHERE user_id='10000000-0000-4000-8000-000000000104' AND purpose='setup_password' AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$1`, at).Scan(&live); err != nil {
			return err
		}
		if live != 10 || len(capture.messages) != 0 {
			return errors.New("live action cap created a new link or disclosed suppression")
		}
		return nil
	})
	if err != nil {
		t.Fatal("owned live action-cap contract failed; details withheld")
	}
}

func TestAuthFullSessionForegroundActivityHasNoAbsoluteCap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		start := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		at := start
		if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,email_verified_at,created_at,updated_at,identity_version) VALUES('10000000-0000-4000-8000-000000000105','continuity@example.test','continuity@example.test','Continuity','active',$1,$1,$1,1)`, start); err != nil {
			return err
		}
		if _, err := f.Pool.Exec(ctx, `INSERT INTO password_credentials(user_id,password_hash,password_changed_at,generation) VALUES('10000000-0000-4000-8000-000000000105','fixture continuity hash',$1,1)`, start); err != nil {
			return err
		}
		secret := identity.HashSecret("fake full customer session")
		if _, err := f.Pool.Exec(ctx, `INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,created_at,kind,identity_version,credential_generation,last_activity_at,browser_cookie_expires_at) VALUES('10000000-0000-4000-8000-000000000106','10000000-0000-4000-8000-000000000105',$1,$2,$2,$2::timestamptz+interval '1176 hours',$2,'full',1,1,$2,$2::timestamptz+interval '1176 hours')`, secret[:], start); err != nil {
			return err
		}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, &actionDeliveryCapture{}, [32]byte{18}, [32]byte{19}, "https://auth.example.test", false)
		view, err := svc.Current(ctx, secret)
		if err != nil || view.Kind != identity.SessionFull || !view.AuthenticatedAt.Equal(start) {
			return errors.New("initial full session current read failed")
		}
		var before time.Time
		if err = f.Pool.QueryRow(ctx, `SELECT last_activity_at FROM auth_sessions WHERE secret_hash=$1`, secret[:]).Scan(&before); err != nil {
			return err
		}
		at = at.Add(time.Hour)
		if _, err = svc.Current(ctx, secret); err != nil {
			return errors.New("current read rejected live full session")
		}
		var afterRead time.Time
		if err = f.Pool.QueryRow(ctx, `SELECT last_activity_at FROM auth_sessions WHERE secret_hash=$1`, secret[:]).Scan(&afterRead); err != nil {
			return err
		}
		if !afterRead.Equal(before) {
			return errors.New("read-only current extended idle activity")
		}
		for day := 0; day < 100; day++ {
			at = at.Add(23 * time.Hour)
			deadline, _, err := svc.TouchFullSession(ctx, secret)
			if err != nil {
				return errors.New("foreground activity failed beyond absolute age")
			}
			if !deadline.Equal(at.Add(identity.SessionIdleLifetime)) {
				return errors.New("foreground activity did not roll exact idle deadline")
			}
			view, err = svc.Current(ctx, secret)
			if err != nil || !view.AuthenticatedAt.Equal(start) {
				return errors.New("activity changed authentication instant or lost current session")
			}
		}
		var preserved bool
		if err = f.Pool.QueryRow(ctx, `SELECT authenticated_at=$2 AND last_activity_at=$3 AND expires_at=$3::timestamptz+interval '1176 hours' AND browser_cookie_expires_at=expires_at AND revoked_at IS NULL FROM auth_sessions WHERE secret_hash=$1`, secret[:], start, at).Scan(&preserved); err != nil {
			return err
		}
		if !preserved {
			return errors.New("long-running foreground session violated rolling/no-absolute-cap contract")
		}
		return nil
	})
	if err != nil {
		t.Fatal("owned rolling session activity contract failed; details withheld")
	}
}

func TestAuthFullSessionPasswordLoginAndCurrentBoundary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		clock := fixedIdentityClock{at: &at}
		sender := &actionDeliveryCapture{}
		svc := identity.NewService(f.Pool, clock, sender, [32]byte{23}, [32]byte{24}, "https://auth.example.test", false)
		registered, err := svc.Register(ctx, identity.RegisterInput{DisplayName: "Login customer", Email: "login@example.test", Telephone: "+14155552671", Password: "A Valid Password For Login 123"}, identity.RequestMeta{Source: "203.0.113.81"})
		if err != nil {
			return err
		}
		verifyURL, err := url.Parse(sender.items[0].URL)
		if err != nil {
			return errors.New("login fixture verification URL malformed")
		}
		verified, err := svc.RedeemAction(ctx, identity.RedeemActionInput{Token: verifyURL.Query().Get("token"), EnrollmentHash: identity.HashSecret(registered.EnrollmentSecret), SessionHash: identity.HashSecret(registered.SessionSecret)})
		if err != nil {
			return err
		}
		oldHash := identity.HashSecret(registered.SessionSecret)
		if _, err = svc.Current(ctx, oldHash); !errors.Is(err, identity.ErrUnauthenticated) {
			return errors.New("verification failed to revoke the old restricted cookie")
		}
		issue, err := svc.Login(ctx, identity.LoginInput{Email: "LOGIN@example.test", Password: "A Valid Password For Login 123"}, identity.RequestMeta{Source: "203.0.113.81", SessionHash: identity.HashSecret(verified.SessionSecret)})
		if err != nil {
			return errors.New("valid password login rejected")
		}
		if issue.View.Kind != identity.SessionFull || !issue.View.EmailVerified || issue.CookieMaxAge != int(identity.SessionIdleLifetime.Seconds()) || len(issue.Secret) != 43 || len(issue.CSRFToken) != 96 {
			return errors.New("full login session contract invalid")
		}
		newHash := identity.HashSecret(issue.Secret)
		if !svc.ValidateSessionCSRF(newHash, issue.CSRFToken) {
			return errors.New("login CSRF not bound to new cookie")
		}
		if _, err = svc.Current(ctx, newHash); err != nil {
			return errors.New("current rejected valid login session")
		}
		wrong, err := svc.Login(ctx, identity.LoginInput{Email: "login@example.test", Password: "A Different Valid Password 123"}, identity.RequestMeta{Source: "203.0.113.81"})
		if err == nil || wrong.Secret != "" {
			return errors.New("wrong password issued session material")
		}
		if err = svc.Logout(ctx, newHash); err != nil {
			return errors.New("logout failed")
		}
		if err = svc.Logout(ctx, newHash); err != nil {
			return errors.New("repeated logout was not idempotent")
		}
		if _, err = svc.Current(ctx, newHash); !errors.Is(err, identity.ErrUnauthenticated) {
			return errors.New("logout left session eligible")
		}
		return nil
	})
	if err != nil {
		t.Fatal("owned customer session login/current/logout contract failed; details withheld")
	}
}

func TestAuthAccountResetRevokesEverySession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stage := "fixture"
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		capture := &actionDeliveryCapture{}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, capture, [32]byte{14}, [32]byte{15}, "https://auth.example.test", false)
		stage = "register-reset-target"
		_, err := svc.Register(ctx, identity.RegisterInput{DisplayName: "Reset target", Email: "reset-target@example.test", Telephone: "+14155552671", Password: "An Initial Password That Is Long"}, identity.RequestMeta{Source: "203.0.113.79"})
		if err != nil {
			return err
		}
		var userID, firstSessionID string
		if err = f.Pool.QueryRow(ctx, `SELECT id::text FROM users WHERE email_normalized='reset-target@example.test'`).Scan(&userID); err != nil {
			return err
		}
		if err = f.Pool.QueryRow(ctx, `SELECT id::text FROM auth_sessions WHERE user_id=$1::uuid AND kind='restricted'`, userID).Scan(&firstSessionID); err != nil {
			return err
		}
		stage = "seed-second-session"
		if _, err = f.Pool.Exec(ctx, `INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,created_at) VALUES('10000000-0000-4000-8000-000000000202',$1,decode(repeat('02',32),'hex'),$2::timestamptz,$2::timestamptz,$2::timestamptz+interval '49 days',$2::timestamptz)`, userID, at); err != nil {
			var pg *pgconn.PgError
			if errors.As(err, &pg) {
				stage = fmt.Sprintf("seed-session-code-%s-constraint-%s", pg.Code, pg.ConstraintName)
				return errors.New("seed second session failed")
			}
			stage = fmt.Sprintf("seed-session-error-%T", err)
			return errors.New("seed second session failed")
		}
		if _, err = f.Pool.Exec(ctx, `INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation) VALUES('20000000-0000-4000-8000-000000000105',$1,'setup_password','reset-target@example.test',decode(repeat('05',32),'hex'),$2,$2::timestamptz+interval '24 hours',1,0)`, userID, at); err != nil {
			return err
		}
		at = at.Add(61 * time.Second)
		stage = "request-reset"
		if _, err = svc.RequestPasswordAction(ctx, "reset-target@example.test", identity.RequestMeta{Source: "203.0.113.79"}); err != nil {
			return err
		}
		parsed, err := url.Parse(capture.items[len(capture.items)-1].URL)
		if err != nil {
			return errors.New("reset URL malformed")
		}
		if _, err = f.Pool.Exec(ctx, `CREATE FUNCTION auth_test_fail_reset_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action='identity.password.reset' THEN RAISE EXCEPTION 'fixture audit rejection' USING ERRCODE='XX000'; END IF; RETURN NEW; END $$`); err != nil {
			return err
		}
		if _, err = f.Pool.Exec(ctx, `CREATE TRIGGER auth_test_fail_reset_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION auth_test_fail_reset_audit()`); err != nil {
			return err
		}
		stage = "reset-audit-rollback"
		if _, err = svc.RedeemAction(ctx, identity.RedeemActionInput{Token: parsed.Query().Get("token"), Password: "A Completely New Password Is Long"}); !errors.Is(err, identity.ErrUnavailable) {
			stage = fmt.Sprintf("reset-audit-category-v%t-c%t-d%t-n%t-u%t", errors.Is(err, identity.ErrValidation), errors.Is(err, identity.ErrConflict), errors.Is(err, identity.ErrDenied), errors.Is(err, identity.ErrNotFound), errors.Is(err, identity.ErrUnavailable))
			return errors.New("reset audit failure did not safely abort")
		}
		stage = "reset-audit-check-rollback"
		var rollbackIntact bool
		actionHash := sha256.Sum256([]byte(parsed.Query().Get("token")))
		var verifiedNull, credentialGeneration, actionsUnconsumed bool
		var sessionCount int
		if err = f.Pool.QueryRow(ctx, `SELECT email_verified_at IS NULL FROM users WHERE id=$1`, userID).Scan(&verifiedNull); err != nil {
			return err
		}
		if err = f.Pool.QueryRow(ctx, `SELECT generation=1 FROM password_credentials WHERE user_id=$1`, userID).Scan(&credentialGeneration); err != nil {
			return err
		}
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM auth_sessions WHERE user_id=$1`, userID).Scan(&sessionCount); err != nil {
			return err
		}
		if err = f.Pool.QueryRow(ctx, `SELECT consumed_at IS NULL AND (SELECT revoked_at IS NULL FROM email_action_tokens WHERE id='20000000-0000-4000-8000-000000000105') FROM email_action_tokens WHERE token_hash=$1`, actionHash[:]).Scan(&actionsUnconsumed); err != nil {
			return err
		}
		rollbackIntact = verifiedNull && credentialGeneration && sessionCount == 2 && actionsUnconsumed
		if err != nil {
			return err
		}
		if !rollbackIntact {
			stage = fmt.Sprintf("reset-audit-state-v%t-g%t-s%d-t%t", verifiedNull, credentialGeneration, sessionCount, actionsUnconsumed)
			return errors.New("reset audit failure partially mutated credentials or revocation state")
		}
		if _, err = f.Pool.Exec(ctx, `DROP TRIGGER auth_test_fail_reset_audit ON audit_events`); err != nil {
			return err
		}
		if _, err = f.Pool.Exec(ctx, `DROP FUNCTION auth_test_fail_reset_audit()`); err != nil {
			return err
		}
		stage = "redeem-reset"
		if _, err = svc.RedeemAction(ctx, identity.RedeemActionInput{Token: parsed.Query().Get("token"), Password: "A Completely New Password Is Long"}); err != nil {
			stage = fmt.Sprintf("reset-result-v%t-c%t-n%t-d%t-u%t", errors.Is(err, identity.ErrValidation), errors.Is(err, identity.ErrConflict), errors.Is(err, identity.ErrNotFound), errors.Is(err, identity.ErrDenied), errors.Is(err, identity.ErrUnavailable))
			return errors.New("password reset transaction rejected")
		}
		var revoked bool
		stage = "persisted-reset-state"
		if err = f.Pool.QueryRow(ctx, `SELECT u.email_verified_at IS NULL AND u.status='active' AND c.generation=2 AND (SELECT count(*)=2 AND bool_and(revoked_at IS NOT NULL) FROM auth_sessions WHERE user_id=u.id) AND NOT EXISTS(SELECT 1 FROM email_action_tokens a WHERE a.user_id=u.id AND a.purpose IN ('setup_password','reset_password') AND a.consumed_at IS NULL AND a.revoked_at IS NULL) FROM users u JOIN password_credentials c ON c.user_id=u.id WHERE u.id=$1`, userID).Scan(&revoked); err != nil {
			var pg *pgconn.PgError
			if errors.As(err, &pg) {
				stage = fmt.Sprintf("persisted-reset-code-%s-constraint-%s", pg.Code, pg.ConstraintName)
			} else {
				stage = fmt.Sprintf("persisted-reset-error-%T", err)
			}
			return err
		}
		if !revoked {
			return errors.New("reset did not revoke all extant sessions and actions")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("owned reset all-device revocation failed at %s; details withheld", stage)
	}
}

func TestAuthAccountAuthenticatedPasswordChangeUsesRevokeContract(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stage := "fixture"
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		capture := &actionDeliveryCapture{}
		svc := identity.NewService(f.Pool, fixedIdentityClock{at: &at}, capture, [32]byte{29}, [32]byte{30}, "https://auth.example.test", false)
		stage = "register-change-target"
		registration, err := svc.Register(ctx, identity.RegisterInput{DisplayName: "Password change", Email: "password-change@example.test", Telephone: "+14155552676", Password: "An Original Password For Customer"}, identity.RequestMeta{Source: "203.0.113.84"})
		if err != nil {
			return err
		}
		verification, err := url.Parse(capture.items[0].URL)
		if err != nil {
			return errors.New("password-change verification URL malformed")
		}
		stage = "verify-change-target"
		verified, err := svc.RedeemAction(ctx, identity.RedeemActionInput{Token: verification.Query().Get("token"), SessionHash: identity.HashSecret(registration.SessionSecret), EnrollmentHash: identity.HashSecret(registration.EnrollmentSecret)})
		if err != nil {
			return err
		}
		oldHash := identity.HashSecret(verified.SessionSecret)
		stage = "change-password"
		if err = svc.ChangePassword(ctx, identity.ChangePasswordInput{CurrentPassword: "An Original Password For Customer", NewPassword: "A Replacement Password For Customer", SessionHash: oldHash}, identity.RequestMeta{Source: "203.0.113.84", SessionHash: oldHash}); err != nil {
			var mutation *identity.MutationFailure
			if errors.As(err, &mutation) {
				stage = fmt.Sprintf("change-password-%s-%s-%s", mutation.Phase(), mutation.SQLState(), mutation.Constraint())
			} else {
				stage = fmt.Sprintf("change-password-v%t-c%t-n%t-d%t-u%t", errors.Is(err, identity.ErrValidation), errors.Is(err, identity.ErrConflict), errors.Is(err, identity.ErrNotFound), errors.Is(err, identity.ErrDenied), errors.Is(err, identity.ErrUnavailable))
			}
			return errors.New("authenticated password change was rejected")
		}
		if _, err = svc.Current(ctx, oldHash); !errors.Is(err, identity.ErrUnauthenticated) {
			return errors.New("password change did not revoke current session")
		}
		if _, err = svc.Login(ctx, identity.LoginInput{Email: "password-change@example.test", Password: "An Original Password For Customer"}, identity.RequestMeta{Source: "203.0.113.84"}); !errors.Is(err, identity.ErrUnauthenticated) {
			return errors.New("password change retained old credential")
		}
		fresh, err := svc.Login(ctx, identity.LoginInput{Email: "password-change@example.test", Password: "A Replacement Password For Customer"}, identity.RequestMeta{Source: "203.0.113.84"})
		if err != nil {
			return errors.New("new credential did not authenticate after fresh login")
		}
		if fresh.View.Kind != identity.SessionFull || !fresh.View.EmailVerified {
			return errors.New("password change altered verification/access state")
		}
		var valid bool
		if err = f.Pool.QueryRow(ctx, `SELECT c.generation=2 AND u.status='active' AND u.email_verified_at IS NOT NULL AND (SELECT count(*)=1 FROM audit_events WHERE action='identity.password.change') FROM users u JOIN password_credentials c ON c.user_id=u.id WHERE u.email_normalized='password-change@example.test'`).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return errors.New("password-change generation/audit/identity state invalid")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("owned authenticated password-change contract failed at %s; details withheld", stage)
	}
}
