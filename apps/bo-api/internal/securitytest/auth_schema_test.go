//go:build integration

package securitytest

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestAuthSchemaBOCustomerTelephoneOwnership(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, staffService *staff.Service) error {
		svc, actor, err := customerCommands(ctx, p, staffService)
		if err != nil {
			return err
		}
		user, err := svc.CreateUser(ctx, actor, customers.UserCreate{Email: "telephone@example.test", DisplayName: "Telephone"})
		if err != nil {
			return err
		}
		var valid bool
		if err = p.QueryRow(ctx, `SELECT telephone IS NULL AND identity_version IS NULL AND NOT EXISTS(SELECT 1 FROM password_credentials) AND NOT EXISTS(SELECT 1 FROM email_action_tokens) FROM users WHERE id=$1`, user.ID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return errors.New("BO creation provisioned customer identity material")
		}
		// Populated by an independent Auth workflow, never a BO form DTO.
		if _, err = p.Exec(ctx, `UPDATE users SET telephone='+6281234567890' WHERE id=$1`, user.ID); err != nil {
			return err
		}
		user, err = svc.UpdateUser(ctx, actor, user.ID, customers.UserProfile{DisplayName: "Updated", Revision: customers.UserRevision(user)})
		if err != nil {
			return err
		}
		user, err = svc.ChangeUserEmail(ctx, actor, user.ID, customers.UserEmail{Email: "replaced@example.test", Sensitive: customers.Sensitive{Revision: customers.UserRevision(user), Reason: "Fixture email replacement", CurrentPassword: fixturePassword}})
		if err != nil {
			return err
		}
		_, err = svc.SetUserStatus(ctx, actor, user.ID, customers.UserStatus{Status: "disabled", Sensitive: customers.Sensitive{Revision: customers.UserRevision(user), Reason: "Fixture disable", CurrentPassword: fixturePassword}})
		if err != nil {
			return err
		}
		if err = p.QueryRow(ctx, `SELECT telephone='+6281234567890' AND status='disabled' AND NOT EXISTS(SELECT 1 FROM email_action_tokens) FROM users WHERE id=$1`, user.ID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return errors.New("BO write lost telephone or issued email material")
		}
		return nil
	})
}
