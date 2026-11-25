//go:build integration

package securitytest

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

type coreAuthClock struct{ at *time.Time }

func (c coreAuthClock) Now() time.Time { return *c.at }

type coreAuthDelivery struct{ items []identity.ActionDelivery }

func (d *coreAuthDelivery) Send(_ context.Context, item identity.ActionDelivery) error {
	d.items = append(d.items, item)
	return nil
}

func loadCoreAuthCustomer(ctx context.Context, pool *pgxpool.Pool, email string) (customers.User, error) {
	var user customers.User
	var verified pgtype.Timestamptz
	err := pool.QueryRow(ctx, `SELECT id,email,display_name,status,email_verified_at,created_at,updated_at FROM users WHERE email_normalized=$1`, email).Scan(&user.ID, &user.Email, &user.DisplayName, &user.Status, &verified, &user.CreatedAt, &user.UpdatedAt)
	if verified.Valid {
		at := verified.Time
		user.EmailVerifiedAt = &at
	}
	return user, err
}

func waitForCoreLockWaiters(ctx context.Context, pool *pgxpool.Pool, minimum int) error {
	deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var count int
		err := pool.QueryRow(deadline, `SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND pid<>pg_backend_pid() AND wait_event_type='Lock'`).Scan(&count)
		if err != nil {
			return err
		}
		if count >= minimum {
			return nil
		}
		select {
		case <-deadline.Done():
			return errors.New("bounded lock-wait barrier not reached")
		case <-ticker.C:
		}
	}
}

func TestBOCustomerRevocationCoreResetCompatibility(t *testing.T) {
	stage := "fixture"
	previousFixtureNow := fixtureNow
	defer func() { fixtureNow = previousFixtureNow }()
	governanceDatabase(t, func(ctx context.Context, pool *pgxpool.Pool, staffService *staff.Service) (callbackErr error) {
		defer func() {
			if callbackErr != nil {
				callbackErr = fmt.Errorf("%s: %w", stage, callbackErr)
			}
		}()
		commands, actor, err := customerCommands(ctx, pool, staffService)
		if err != nil {
			stage = "bootstrap-staff"
			return err
		}
		at := fixtureNow
		stage = "compose-core-identity-service"
		delivery := &coreAuthDelivery{}
		core := identity.NewService(pool, coreAuthClock{at: &at}, delivery, [32]byte{21}, [32]byte{22}, "https://auth.example.test", false)
		register := func(email string) error {
			stage = "register-" + email
			_, err := core.Register(ctx, identity.RegisterInput{DisplayName: "Core customer", Email: email, Telephone: "+14155552671", Password: "A long initial customer password"}, identity.RequestMeta{Source: "198.51.100.44"})
			return err
		}
		requestReset := func(email string) (string, error) {
			stage = "request-reset-" + email
			at = at.Add(61 * time.Second)
			_, err := core.RequestPasswordAction(ctx, email, identity.RequestMeta{Source: "198.51.100.44"})
			if err != nil {
				return "", err
			}
			parsed, err := url.Parse(delivery.items[len(delivery.items)-1].URL)
			if err != nil {
				return "", errors.New("fake reset URL malformed")
			}
			return parsed.Query().Get("token"), nil
		}
		// BO email replacement first: the old-email action must be unusable.
		stage = "register-email-replacement-race-target"
		firstRegistration, err := core.Register(ctx, identity.RegisterInput{DisplayName: "Core customer", Email: "replace-before-reset@example.test", Telephone: "+14155552671", Password: "A long initial customer password"}, identity.RequestMeta{Source: "198.51.100.44"})
		if err != nil {
			return err
		}
		verifyURL, err := url.Parse(delivery.items[len(delivery.items)-1].URL)
		if err != nil {
			return errors.New("fake verification URL malformed")
		}
		fullSession, err := core.RedeemAction(ctx, identity.RedeemActionInput{Token: verifyURL.Query().Get("token"), EnrollmentHash: identity.HashSecret(firstRegistration.EnrollmentSecret), SessionHash: identity.HashSecret(firstRegistration.SessionSecret)})
		if err != nil {
			return errors.New("verification before BO email replacement failed")
		}
		if _, err = core.Current(ctx, identity.HashSecret(fullSession.SessionSecret)); err != nil {
			return errors.New("verified full session was not live before BO email replacement")
		}
		user, err := loadCoreAuthCustomer(ctx, pool, "replace-before-reset@example.test")
		if err != nil {
			return err
		}
		beforeChange := fullSession.SessionSecret
		token, err := requestReset("replace-before-reset@example.test")
		if err != nil {
			return err
		}
		beforeBOEmailMailCount := len(delivery.items)
		fixtureNow = at
		user, err = commands.ChangeUserEmail(ctx, actor, user.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(user)), Email: "replaced-before-reset@example.test"})
		if err != nil {
			return err
		}
		if len(delivery.items) != beforeBOEmailMailCount {
			return errors.New("BO email replacement emitted customer email")
		}
		if _, err = core.Current(ctx, identity.HashSecret(beforeChange)); !errors.Is(err, identity.ErrUnauthenticated) {
			return errors.New("BO email replacement left the old Core session current")
		}
		stage = "email-replacement-before-reset"
		if _, err = core.RedeemAction(ctx, identity.RedeemActionInput{Token: token, Password: "A new reset password is long"}); !errors.Is(err, identity.ErrNotFound) {
			return errors.New("Core reset succeeded after actual BO email replacement")
		}
		var unchanged bool
		if err = pool.QueryRow(ctx, `SELECT u.telephone='+14155552671' AND c.generation=1 AND NOT EXISTS(SELECT 1 FROM email_action_tokens a WHERE a.user_id=u.id AND a.purpose='reset_password' AND a.revoked_at IS NULL) FROM users u JOIN password_credentials c ON c.user_id=u.id WHERE u.id=$1`, user.ID).Scan(&unchanged); err != nil {
			return err
		}
		if !unchanged {
			return errors.New("BO email replacement left stale Core reset credential/action state")
		}
		// BO disable first: a previously issued reset cannot reactivate the user.
		if err = register("disable-before-reset@example.test"); err != nil {
			return err
		}
		disabled, err := loadCoreAuthCustomer(ctx, pool, "disable-before-reset@example.test")
		if err != nil {
			return err
		}
		disabledSession, err := core.Login(ctx, identity.LoginInput{Email: "disable-before-reset@example.test", Password: "A long initial customer password"}, identity.RequestMeta{Source: "198.51.100.44"})
		if err != nil || disabledSession.View.Kind != identity.SessionRestricted {
			return errors.New("unverified login did not issue restricted session")
		}
		token, err = requestReset("disable-before-reset@example.test")
		if err != nil {
			return err
		}
		fixtureNow = at
		stage = "bo-disable-after-reset-request"
		disabled, err = commands.SetUserStatus(ctx, actor, disabled.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(disabled)), Status: "disabled"})
		if err != nil {
			return fmt.Errorf("BO disable command failed: %w", err)
		}
		if len(delivery.items) != 4 {
			return errors.New("BO disable emitted a customer email")
		}
		if _, err = core.Current(ctx, identity.HashSecret(disabledSession.Secret)); !errors.Is(err, identity.ErrUnauthenticated) {
			return errors.New("BO disable left the old Core session current")
		}
		if _, err = core.RedeemAction(ctx, identity.RedeemActionInput{Token: token, Password: "A new reset password is long"}); !errors.Is(err, identity.ErrNotFound) {
			return errors.New("Core reset reactivated a BO-disabled customer")
		}
		stage = "bo-reactivate"
		disabled, err = commands.SetUserStatus(ctx, actor, disabled.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(disabled)), Status: "active"})
		if err != nil {
			return fmt.Errorf("BO reactivation failed: %w", err)
		}
		stage = "bo-manual-verify"
		disabled, err = commands.VerifyUserEmail(ctx, actor, disabled.ID, sensitive(customers.UserRevision(disabled)))
		if err != nil {
			return fmt.Errorf("BO manual verification failed: %w", err)
		}
		if _, err = core.Current(ctx, identity.HashSecret(disabledSession.Secret)); !errors.Is(err, identity.ErrUnauthenticated) {
			return errors.New("BO reactivation/manual verification resurrected a revoked restricted session")
		}
		// Core reset first, followed by an actual BO email replacement, retains
		// the reset's credential generation and does not resurrect old action rows.
		if err = register("reset-before-replacement@example.test"); err != nil {
			return err
		}
		changing, err := loadCoreAuthCustomer(ctx, pool, "reset-before-replacement@example.test")
		if err != nil {
			return err
		}
		token, err = requestReset("reset-before-replacement@example.test")
		if err != nil {
			return err
		}
		if _, err = core.RedeemAction(ctx, identity.RedeemActionInput{Token: token, Password: "A second new customer password is long"}); err != nil {
			return errors.New("Core reset-before-BO replacement failed")
		}
		changing, err = commands.ChangeUserEmail(ctx, actor, changing.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(changing)), Email: "after-reset@example.test"})
		if err != nil {
			return err
		}
		var retained bool
		if err = pool.QueryRow(ctx, `SELECT u.email_normalized='after-reset@example.test' AND u.telephone='+14155552671' AND u.status='active' AND c.generation=2 AND NOT EXISTS(SELECT 1 FROM email_action_tokens a WHERE a.user_id=u.id AND a.purpose IN ('setup_password','reset_password') AND a.revoked_at IS NULL AND a.consumed_at IS NULL) FROM users u JOIN password_credentials c ON c.user_id=u.id WHERE u.id=$1`, changing.ID).Scan(&retained); err != nil {
			return err
		}
		if !retained {
			return errors.New("BO replacement undid or resurrected committed Core reset state")
		}
		// Race password login/session rotation against the actual BO email change.
		stage = "login-email-replacement-overlap"
		loginRegistration, err := core.Register(ctx, identity.RegisterInput{DisplayName: "Login overlap", Email: "login-overlap@example.test", Telephone: "+14155552673", Password: "A long overlap password 123"}, identity.RequestMeta{Source: "198.51.100.45"})
		if err != nil {
			return err
		}
		loginUser, err := loadCoreAuthCustomer(ctx, pool, "login-overlap@example.test")
		if err != nil {
			return err
		}
		blocker, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = blocker.Exec(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, loginUser.ID); err != nil {
			_ = blocker.Rollback(ctx)
			return err
		}
		var wg sync.WaitGroup
		var loginErr, loginBOErr error
		var loginIssue identity.SessionIssue
		wg.Add(2)
		go func() {
			defer wg.Done()
			loginIssue, loginErr = core.Login(ctx, identity.LoginInput{Email: "login-overlap@example.test", Password: "A long overlap password 123"}, identity.RequestMeta{Source: "198.51.100.45", SessionHash: identity.HashSecret(loginRegistration.SessionSecret)})
		}()
		go func() {
			defer wg.Done()
			_, loginBOErr = commands.ChangeUserEmail(ctx, actor, loginUser.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(loginUser)), Email: "login-overlap-replaced@example.test"})
		}()
		if err = waitForCoreLockWaiters(ctx, pool, 2); err != nil {
			_ = blocker.Rollback(ctx)
			wg.Wait()
			return errors.New("login/email overlap did not reach the actual user-lock barrier")
		}
		if err = blocker.Commit(ctx); err != nil {
			wg.Wait()
			return err
		}
		wg.Wait()
		if loginBOErr != nil {
			return errors.New("actual BO email change failed during customer login")
		}
		if loginErr != nil && !errors.Is(loginErr, identity.ErrUnauthenticated) && !errors.Is(loginErr, identity.ErrConflict) {
			return errors.New("customer login/email race returned unexpected safe result")
		}
		if loginErr == nil {
			if _, err = core.Current(ctx, identity.HashSecret(loginIssue.Secret)); !errors.Is(err, identity.ErrUnauthenticated) {
				return errors.New("BO email change left a concurrently rotated customer session usable")
			}
		}
		var loginRaceSafe bool
		if err = pool.QueryRow(ctx, `SELECT u.email_normalized='login-overlap-replaced@example.test' AND NOT EXISTS(SELECT 1 FROM auth_sessions s WHERE s.user_id=u.id AND s.revoked_at IS NULL) FROM users u WHERE u.id=$1`, loginUser.ID).Scan(&loginRaceSafe); err != nil {
			return err
		}
		if !loginRaceSafe {
			return errors.New("login/email overlap retained a session for the replaced email")
		}
		stage = "login-disable-overlap"
		registration, err := core.Register(ctx, identity.RegisterInput{DisplayName: "Login disable overlap", Email: "login-disable-overlap@example.test", Telephone: "+14155552674", Password: "A long disable overlap password"}, identity.RequestMeta{Source: "198.51.100.46"})
		if err != nil {
			return err
		}
		disablingUser, err := loadCoreAuthCustomer(ctx, pool, "login-disable-overlap@example.test")
		if err != nil {
			return err
		}
		blocker, err = pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = blocker.Exec(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, disablingUser.ID); err != nil {
			_ = blocker.Rollback(ctx)
			return err
		}
		wg = sync.WaitGroup{}
		var disableLoginErr, disableBOErr error
		var disableLogin identity.SessionIssue
		wg.Add(2)
		go func() {
			defer wg.Done()
			disableLogin, disableLoginErr = core.Login(ctx, identity.LoginInput{Email: "login-disable-overlap@example.test", Password: "A long disable overlap password"}, identity.RequestMeta{Source: "198.51.100.46", SessionHash: identity.HashSecret(registration.SessionSecret)})
		}()
		go func() {
			defer wg.Done()
			_, disableBOErr = commands.SetUserStatus(ctx, actor, disablingUser.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(disablingUser)), Status: "disabled"})
		}()
		if err = waitForCoreLockWaiters(ctx, pool, 2); err != nil {
			_ = blocker.Rollback(ctx)
			wg.Wait()
			return errors.New("login/disable overlap did not reach the actual user-lock barrier")
		}
		if err = blocker.Commit(ctx); err != nil {
			wg.Wait()
			return err
		}
		wg.Wait()
		if disableBOErr != nil {
			return errors.New("actual BO disable failed during customer login")
		}
		if disableLoginErr != nil && !errors.Is(disableLoginErr, identity.ErrUnauthenticated) && !errors.Is(disableLoginErr, identity.ErrConflict) {
			return errors.New("customer login/disable race returned unexpected safe outcome")
		}
		for _, cookie := range []string{registration.SessionSecret, disableLogin.Secret} {
			if cookie == "" {
				continue
			}
			if _, err = core.Current(ctx, identity.HashSecret(cookie)); !errors.Is(err, identity.ErrUnauthenticated) {
				return errors.New("BO disable race left a customer session current")
			}
		}
		var disableRaceSafe bool
		if err = pool.QueryRow(ctx, `SELECT u.status='disabled' AND NOT EXISTS(SELECT 1 FROM auth_sessions s WHERE s.user_id=u.id AND s.revoked_at IS NULL) FROM users u WHERE u.id=$1`, disablingUser.ID).Scan(&disableRaceSafe); err != nil {
			return err
		}
		if !disableRaceSafe {
			return errors.New("login/disable overlap left active customer session authority")
		}
		// Actual concurrent Core reset redemption and BO email replacement. Either
		// user-first transaction may win; neither leaves a reset usable for the
		// replacement address or restores the old session authority.
		if err = register("overlap@example.test"); err != nil {
			return err
		}
		overlapUser, err := loadCoreAuthCustomer(ctx, pool, "overlap@example.test")
		if err != nil {
			return err
		}
		overlapToken, err := requestReset("overlap@example.test")
		if err != nil {
			return err
		}
		blocker, err = pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = blocker.Exec(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, overlapUser.ID); err != nil {
			_ = blocker.Rollback(ctx)
			return err
		}
		wg = sync.WaitGroup{}
		var resetErr, boErr error
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, resetErr = core.RedeemAction(ctx, identity.RedeemActionInput{Token: overlapToken, Password: "Concurrent reset password is long"})
		}()
		go func() {
			defer wg.Done()
			_, boErr = commands.ChangeUserEmail(ctx, actor, overlapUser.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(overlapUser)), Email: "overlap-replaced@example.test"})
		}()
		if err = waitForCoreLockWaiters(ctx, pool, 2); err != nil {
			_ = blocker.Rollback(ctx)
			wg.Wait()
			return errors.New("reset/email operations did not reach the actual user-lock barrier")
		}
		if err = blocker.Commit(ctx); err != nil {
			wg.Wait()
			return err
		}
		wg.Wait()
		if boErr != nil {
			return errors.New("actual BO email replacement failed in reset interleaving")
		}
		if resetErr != nil && !errors.Is(resetErr, identity.ErrNotFound) && !errors.Is(resetErr, identity.ErrDenied) && !errors.Is(resetErr, identity.ErrConflict) {
			return errors.New("Core reset race returned an unexpected safe outcome")
		}
		var raceInvariant bool
		if err = pool.QueryRow(ctx, `SELECT u.email_normalized='overlap-replaced@example.test' AND u.status='active' AND NOT EXISTS(SELECT 1 FROM auth_sessions s WHERE s.user_id=u.id AND s.revoked_at IS NULL) AND NOT EXISTS(SELECT 1 FROM email_action_tokens a WHERE a.user_id=u.id AND a.target_email_normalized='overlap@example.test' AND a.purpose='reset_password' AND a.revoked_at IS NULL AND a.consumed_at IS NULL) FROM users u WHERE u.id=$1`, overlapUser.ID).Scan(&raceInvariant); err != nil {
			return err
		}
		if !raceInvariant {
			return errors.New("Core/BO overlapping reset-email state retained usable old authority")
		}
		if err = register("overlap-disable@example.test"); err != nil {
			return err
		}
		disableUser, err := loadCoreAuthCustomer(ctx, pool, "overlap-disable@example.test")
		if err != nil {
			return err
		}
		disableToken, err := requestReset("overlap-disable@example.test")
		if err != nil {
			return err
		}
		blocker, err = pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = blocker.Exec(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, disableUser.ID); err != nil {
			_ = blocker.Rollback(ctx)
			return err
		}
		wg = sync.WaitGroup{}
		resetErr, boErr = nil, nil
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, resetErr = core.RedeemAction(ctx, identity.RedeemActionInput{Token: disableToken, Password: "Concurrent disable reset password"})
		}()
		go func() {
			defer wg.Done()
			_, boErr = commands.SetUserStatus(ctx, actor, disableUser.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(disableUser)), Status: "disabled"})
		}()
		if err = waitForCoreLockWaiters(ctx, pool, 2); err != nil {
			_ = blocker.Rollback(ctx)
			wg.Wait()
			return errors.New("reset/disable operations did not reach the actual user-lock barrier")
		}
		if err = blocker.Commit(ctx); err != nil {
			wg.Wait()
			return err
		}
		wg.Wait()
		if boErr != nil {
			return errors.New("actual BO disable failed in reset interleaving")
		}
		if resetErr != nil && !errors.Is(resetErr, identity.ErrNotFound) && !errors.Is(resetErr, identity.ErrDenied) && !errors.Is(resetErr, identity.ErrConflict) {
			return errors.New("Core reset-disable race returned an unexpected safe outcome")
		}
		if err = pool.QueryRow(ctx, `SELECT u.status='disabled' AND NOT EXISTS(SELECT 1 FROM auth_sessions s WHERE s.user_id=u.id AND s.revoked_at IS NULL) AND NOT EXISTS(SELECT 1 FROM email_action_tokens a WHERE a.user_id=u.id AND a.purpose='reset_password' AND a.revoked_at IS NULL AND a.consumed_at IS NULL) FROM users u WHERE u.id=$1`, disableUser.ID).Scan(&raceInvariant); err != nil {
			return err
		}
		if !raceInvariant {
			return errors.New("Core/BO overlapping reset-disable state reactivated user or left usable authority")
		}
		return nil
	})
}
