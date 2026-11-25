//go:build integration

package authtest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type countingIdentitySender struct{ count int }

func (s *countingIdentitySender) Send(_ context.Context, _ identity.ActionDelivery) error {
	s.count++
	return nil
}

func TestAuthAccountRegistrationAdmissionCap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
		clock := fixedIdentityClock{at: &at}
		sender := &countingIdentitySender{}
		svc := identity.NewService(f.Pool, clock, sender, [32]byte{3}, [32]byte{4}, "https://auth.example.test", false)
		input := identity.RegisterInput{DisplayName: "Admission", Email: "cap@example.test", Telephone: "+14155552671", Password: "A sufficiently long fake password"}
		for i := 0; i < 7; i++ {
			result, err := svc.Register(ctx, input, identity.RequestMeta{Source: "198.51.100.25"})
			if err != nil {
				return err
			}
			if result.Accepted != (identity.PublicAccepted{}) || len(result.EnrollmentSecret) != 43 || len(result.SessionSecret) != 43 {
				return errors.New("admission changed generic registration response")
			}
			if i > 0 {
				at = at.Add(61 * time.Second)
			}
		}
		var users, credentials, sessions, actions int
		if err := f.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM password_credentials),(SELECT count(*) FROM auth_sessions),(SELECT count(*) FROM email_action_tokens)`).Scan(&users, &credentials, &sessions, &actions); err != nil {
			return err
		}
		if users != 1 || credentials != 1 || sessions != 1 || actions != 1 || sender.count != 1 {
			return errors.New("account-purpose admission cap permitted repeated identity/email issuance")
		}
		return nil
	})
	if err != nil {
		t.Fatal("owned account admission cap failed; details withheld")
	}
}
