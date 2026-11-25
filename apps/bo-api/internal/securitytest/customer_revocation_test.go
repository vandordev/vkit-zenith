//go:build integration

package securitytest

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

// Valid canonical FKs/timestamp/scope rows, no disabled guards, fake secrets only.
// Future-created unrevoked rows exercise creation/issuance lower bounds; historical
// consumption/revocation instants are deliberately distinct from operation time.
func seedCustomerAuthorization(ctx context.Context, p *pgxpool.Pool, userID pgtype.UUID) error {
	statements := []string{
		`INSERT INTO password_credentials(user_id,password_hash,password_changed_at) VALUES($1,'fake-customer-credential-never-used',$2)`,
		`INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,revoked_at,created_at) VALUES
		('10000000-0000-4000-8000-000000000001',$1,'\x0101',$2::timestamptz+interval '1 hour',$2::timestamptz+interval '1 hour',$2::timestamptz+interval '2 days',NULL,$2::timestamptz+interval '1 hour'),
		('10000000-0000-4000-8000-000000000002',$1,'\x0102',$2::timestamptz-interval '1 hour',$2::timestamptz-interval '1 hour',$2::timestamptz+interval '2 days',$2::timestamptz-interval '30 minutes',$2::timestamptz-interval '1 hour')`,
		`INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,consumed_at,revoked_at) VALUES
		('20000000-0000-4000-8000-000000000001',$1,'verify_email','person@example.test','\x0201',$2::timestamptz+interval '1 hour',$2::timestamptz+interval '2 days',NULL,NULL),
		('20000000-0000-4000-8000-000000000002',$1,'reset_password','person@example.test','\x0202',$2::timestamptz-interval '1 hour',$2::timestamptz+interval '2 days',$2::timestamptz-interval '20 minutes',NULL),
		('20000000-0000-4000-8000-000000000003',$1,'verify_email','person@example.test','\x0203',$2::timestamptz-interval '1 hour',$2::timestamptz+interval '2 days',NULL,$2::timestamptz-interval '30 minutes')`,
		`INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at) VALUES('20000000-0000-4000-8000-000000000004',$1,'reset_password','person@example.test','\x0204',$2::timestamptz-interval '1 hour',$2::timestamptz+interval '2 days')`,
	}
	for _, sql := range statements {
		if _, err := p.Exec(ctx, sql, userID, fixtureNow); err != nil {
			return err
		}
	}
	return nil
}

const customerAuthorizationSnapshotSQL = `SELECT jsonb_build_array(
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY id) FROM users x),
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY id) FROM organization_memberships x),
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY id) FROM auth_sessions x),
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY id) FROM email_action_tokens x),
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY id) FROM platform_staff_sessions x),
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY user_id) FROM password_credentials x),
	(SELECT count(*) FROM audit_events))::text`

func customerAuthorizationSnapshot(ctx context.Context, p *pgxpool.Pool) (string, error) {
	var state string
	err := p.QueryRow(ctx, customerAuthorizationSnapshotSQL).Scan(&state)
	return state, err
}

func assertCustomerRevocation(ctx context.Context, p *pgxpool.Pool, userID pgtype.UUID) error {
	var ok bool
	err := p.QueryRow(ctx, `SELECT
	(SELECT count(*)=2 AND bool_and(revoked_at IS NOT NULL) FROM auth_sessions WHERE user_id=$1)
	AND (SELECT revoked_at=created_at FROM auth_sessions WHERE id='10000000-0000-4000-8000-000000000001')
	AND (SELECT revoked_at=$2::timestamptz-interval '30 minutes' FROM auth_sessions WHERE id='10000000-0000-4000-8000-000000000002')
	AND (SELECT revoked_at=created_at AND consumed_at IS NULL FROM email_action_tokens WHERE id='20000000-0000-4000-8000-000000000001')
	AND (SELECT revoked_at IS NULL AND consumed_at=$2::timestamptz-interval '20 minutes' FROM email_action_tokens WHERE id='20000000-0000-4000-8000-000000000002')
	AND (SELECT revoked_at=$2::timestamptz-interval '30 minutes' FROM email_action_tokens WHERE id='20000000-0000-4000-8000-000000000003')
	AND (SELECT revoked_at=$2::timestamptz AND consumed_at IS NULL FROM email_action_tokens WHERE id='20000000-0000-4000-8000-000000000004')
	AND NOT EXISTS(SELECT 1 FROM platform_staff_sessions WHERE revoked_at IS NOT NULL)
	AND (SELECT password_hash='fake-customer-credential-never-used' AND password_changed_at=$2 FROM password_credentials WHERE user_id=$1)`, userID, fixtureNow).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("persisted revocation/history/credential invariant failed")
	}
	return nil
}

func TestBOCustomerRevocationDisableAndEmail(t *testing.T) {
	for _, action := range []string{"disable", "email"} {
		t.Run(action, func(t *testing.T) {
			governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
				c, a, err := customerCommands(ctx, p, svc)
				if err != nil {
					return err
				}
				u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "person@example.test", DisplayName: "Person"})
				if err != nil {
					return err
				}
				// A historical removed relation must not be restored by reactivation.
				o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Other", Slug: "other", NewOwner: &customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"}})
				if err != nil {
					return err
				}
				m, err := c.AddMember(ctx, a, o.Organization.ID, u.ID, customers.RoleMember)
				if err != nil {
					return err
				}
				m, err = c.RemoveMember(ctx, a, o.Organization.ID, m.ID, customers.MembershipRevision(m))
				if err != nil {
					return err
				}
				if err = seedCustomerAuthorization(ctx, p, u.ID); err != nil {
					return err
				}
				if _, err = p.Exec(ctx, `INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,created_at) VALUES('10000000-0000-4000-8000-000000000003',$1,'\x0103',$2,$2,$2::timestamptz+interval '1 day',$2)`, o.User.ID, fixtureNow); err != nil {
					return err
				}
				if _, err = c.CreateUser(ctx, a, customers.UserCreate{Email: "taken@example.test", DisplayName: "Taken"}); err != nil {
					return err
				}
				before, err := customerAuthorizationSnapshot(ctx, p)
				if err != nil {
					return err
				}
				if _, err = c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(u)), Email: "TAKEN@example.test"}); !errors.Is(err, staff.ErrConflict) {
					return errors.New("duplicate replacement not conflict")
				}
				after, err := customerAuthorizationSnapshot(ctx, p)
				if err != nil {
					return err
				}
				if before != after {
					return errors.New("duplicate email committed authorization revocation")
				}
				if action == "disable" {
					u, err = c.SetUserStatus(ctx, a, u.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(u)), Status: "disabled"})
				} else {
					u, err = c.VerifyUserEmail(ctx, a, u.ID, sensitive(customers.UserRevision(u)))
					if err != nil {
						return err
					}
					u, err = c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(u)), Email: "replacement@example.test"})
				}
				if err != nil {
					return err
				}
				if err = assertCustomerRevocation(ctx, p, u.ID); err != nil {
					return err
				}
				var otherUnaffected bool
				if err = p.QueryRow(ctx, `SELECT revoked_at IS NULL FROM auth_sessions WHERE id='10000000-0000-4000-8000-000000000003'`).Scan(&otherUnaffected); err != nil {
					return err
				}
				if !otherUnaffected {
					return errors.New("other customer's session revoked")
				}
				if action == "disable" {
					u, err = c.SetUserStatus(ctx, a, u.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(u)), Status: "active"})
					if err != nil {
						return err
					}
				}
				u, err = c.VerifyUserEmail(ctx, a, u.ID, sensitive(customers.UserRevision(u)))
				if err != nil {
					return err
				}
				if err = assertCustomerRevocation(ctx, p, u.ID); err != nil {
					return err
				}
				var ok bool
				if err = p.QueryRow(ctx, `SELECT status='removed' AND role <> 'owner' AND updated_at=$2 FROM organization_memberships WHERE id=$1`, m.ID, m.UpdatedAt).Scan(&ok); err != nil {
					return err
				}
				if !ok {
					return errors.New("reactivation restored removed membership")
				}
				return nil
			})
		})
	}
}

func TestBOCustomerRevocationAuditFailureRollsBackEverything(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "person@example.test", DisplayName: "Person"})
		if err != nil {
			return err
		}
		u, err = c.VerifyUserEmail(ctx, a, u.ID, sensitive(customers.UserRevision(u)))
		if err != nil {
			return err
		}
		if err = seedCustomerAuthorization(ctx, p, u.ID); err != nil {
			return err
		}
		before, err := customerAuthorizationSnapshot(ctx, p)
		if err != nil {
			return err
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		for _, action := range []string{"disable", "email"} {
			if action == "disable" {
				_, err = c.SetUserStatus(ctx, a, u.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(u)), Status: "disabled"})
			} else {
				_, err = c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(u)), Email: "rollback@example.test"})
			}
			if !errors.Is(err, staff.ErrUnavailable) {
				return errors.New("revocation audit error unsafe")
			}
			after, err := customerAuthorizationSnapshot(ctx, p)
			if err != nil {
				return err
			}
			if before != after {
				return errors.New("audit failure committed entity/authorization/history effects")
			}
		}
		return nil
	})
}
