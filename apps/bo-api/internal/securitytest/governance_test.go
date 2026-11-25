//go:build integration

package securitytest

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"github.com/vandordev/vkit-zenith/packages/db/migrate"
	"github.com/vandordev/vkit-zenith/packages/tools/dbtool"
)

const fixturePassword = "isolated fixture password 123"
const replacementPassword = "isolated replacement password 456"

var fixtureNow = time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)

func governanceDatabase(t *testing.T, fn func(context.Context, *pgxpool.Pool, *staff.Service) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	ctx = dbtool.WithOwnershipJournal(ctx, os.Getenv("AUTH_OWNERSHIP_MANIFEST"))
	err := dbtool.WithPostgres(ctx, func(ctx context.Context, dsn string) error {
		settings := db.Settings{DSN: dsn, MaxOpenConnections: 10, ConnectionMaxLifetimeSeconds: 60, ConnectionMaxIdleTimeSeconds: 30, ConnectTimeoutSeconds: 10}
		m, err := migrate.Open(ctx, settings)
		if err != nil {
			return err
		}
		_, runErr := migrate.Run(ctx, m, os.DirFS("../../../../packages/db/migrations"), migrate.Request{Operation: migrate.Up})
		if runErr == nil {
			_, runErr = m.ExecContext(ctx, `ALTER ROLE fixture SET lock_timeout='5s';ALTER ROLE fixture SET statement_timeout='15s'`)
		}
		closeErr := m.Close()
		if runErr != nil || closeErr != nil {
			return errors.Join(runErr, closeErr)
		}
		pool, err := db.Open(ctx, settings)
		if err != nil {
			return err
		}
		defer pool.Close()
		return fn(ctx, pool, staff.NewService(pool, func() time.Time { return fixtureNow }, staff.NewPasswords()))
	})
	if err != nil {
		t.Fatal(err)
	}
}
func bootstrapInput(email string) staff.StaffWrite {
	return staff.StaffWrite{Email: email, DisplayName: "Fixture Staff", Role: "superadmin", Status: "active", Password: fixturePassword}
}

// Sessions are SQL fixtures here; Task 6 alone owns actual login issuance.
func fixtureActor(ctx context.Context, p *pgxpool.Pool, s staff.Staff) (staff.Actor, error) {
	_, hash, err := staff.NewSessionToken()
	if err != nil {
		return staff.Actor{}, err
	}
	_, err = p.Exec(ctx, `INSERT INTO platform_staff_sessions(id,staff_id,token_hash,created_at,expires_at) VALUES(gen_random_uuid(),$1,$2,$3,$4)`, s.ID, hash[:], fixtureNow, fixtureNow.Add(30*24*time.Hour))
	return staff.Actor{StaffID: s.ID, SessionHash: hash, Source: "127.0.0.1"}, err
}

func TestBOBootstrapFirstOnly(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		in := bootstrapInput("first@example.test")
		s, err := svc.Bootstrap(ctx, in)
		if err != nil {
			return err
		}
		if s.Role != "superadmin" || s.Status != "active" {
			return errors.New("invalid bootstrap role")
		}
		if _, err = svc.Bootstrap(ctx, bootstrapInput("second@example.test")); !errors.Is(err, staff.ErrConflict) {
			return errors.New("second bootstrap allowed")
		}
		var count int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action='staff.bootstrap' AND outcome='success'`).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return errors.New("bootstrap audit missing")
		}
		return nil
	})
}

func TestBOBootstrapNoTakeover(t *testing.T) {
	for _, role := range []string{"admin", "superadmin"} {
		t.Run(role, func(t *testing.T) {
			governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
				_, err := p.Exec(ctx, `INSERT INTO platform_staff(id,email_normalized,display_name,role,status,created_at,updated_at) VALUES(gen_random_uuid(),'existing@example.test','Existing',$1,'disabled',$2,$2)`, role, fixtureNow)
				if err != nil {
					return err
				}
				if _, err = svc.Bootstrap(ctx, bootstrapInput("existing@example.test")); !errors.Is(err, staff.ErrConflict) {
					return errors.New("existing identity takeover allowed")
				}
				if role == "superadmin" {
					if _, err = svc.Bootstrap(ctx, bootstrapInput("new@example.test")); !errors.Is(err, staff.ErrConflict) {
						return errors.New("disabled superadmin reopened bootstrap")
					}
				}
				var n int
				if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_credentials`).Scan(&n); err != nil {
					return err
				}
				if n != 0 {
					return errors.New("takeover issued credentials")
				}
				if role == "admin" {
					if _, err = svc.Bootstrap(ctx, bootstrapInput("independent@example.test")); err != nil {
						return err
					}
				}
				return nil
			})
		})
	}
}

func failAudits(ctx context.Context, p *pgxpool.Pool) error {
	_, err := p.Exec(ctx, `CREATE FUNCTION fixture_reject_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'fixture audit unavailable'; END $$;
 CREATE TRIGGER fixture_reject_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fixture_reject_audit()`)
	return err
}

func TestBOBootstrapAuditRollback(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		if err := failAudits(ctx, p); err != nil {
			return err
		}
		if _, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test")); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("bootstrap audit failure not safe")
		}
		var ok bool
		if err := p.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM platform_staff) AND NOT EXISTS(SELECT 1 FROM platform_staff_credentials) AND NOT EXISTS(SELECT 1 FROM audit_events) AND (SELECT bootstrapped_at IS NULL FROM platform_staff_auth_control)`).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("bootstrap partially committed")
		}
		return nil
	})
}

func TestBOGovernanceAuditRollback(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		actor, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		in := bootstrapInput("target@example.test")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		target, err := svc.CreateStaff(ctx, actor, in)
		if err != nil {
			return err
		}
		if _, err = fixtureActor(ctx, p, target); err != nil {
			return err
		}
		var before string
		if err = p.QueryRow(ctx, `SELECT password_hash FROM platform_staff_credentials WHERE staff_id=$1`, target.ID).Scan(&before); err != nil {
			return err
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		for _, kind := range []string{"create", "update", "password"} {
			switch kind {
			case "create":
				in.Email = "failed@example.test"
				_, err = svc.CreateStaff(ctx, actor, in)
			case "update":
				in.Email = target.Email
				in.Role = "superadmin"
				in.Revision = staff.Revision(target)
				_, err = svc.UpdateStaff(ctx, actor, target.ID, inWithoutPassword(in))
			case "password":
				err = svc.ReplacePassword(ctx, actor, target.ID, staff.StaffWrite{Password: replacementPassword, CurrentPassword: fixturePassword, Revision: staff.Revision(target)})
			}
			if !errors.Is(err, staff.ErrUnavailable) {
				return errors.New("audit failure was not safe dependency error")
			}
			var ok bool
			if err = p.QueryRow(ctx, `SELECT s.role='admin' AND s.email_normalized='target@example.test' AND c.password_hash=$2
    AND EXISTS(SELECT 1 FROM platform_staff_sessions WHERE staff_id=s.id AND revoked_at IS NULL)
    AND (SELECT count(*) FROM platform_staff)=2 AND (SELECT count(*) FROM audit_events)=2
    FROM platform_staff s JOIN platform_staff_credentials c ON c.staff_id=s.id WHERE s.id=$1`, target.ID, before).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return errors.New("failed audit partially committed staff/credentials/revocation")
			}
		}
		return nil
	})
}
func inWithoutPassword(in staff.StaffWrite) staff.StaffWrite { in.Password = ""; return in }

func TestBOGovernanceLastSuperadminRace(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		input := bootstrapInput("second@example.test")
		input.CurrentPassword = fixturePassword
		second, err := svc.CreateStaff(ctx, a, input)
		if err != nil {
			return err
		}
		b, err := fixtureActor(ctx, p, second)
		if err != nil {
			return err
		}
		other := staff.NewService(p, func() time.Time { return fixtureNow }, staff.NewPasswords())
		start := make(chan struct{})
		done := make(chan error, 2)
		for i, target := range []staff.Staff{root, second} {
			actor := a
			service := svc
			if i == 1 {
				actor = b
				service = other
			}
			go func(target staff.Staff, actor staff.Actor, service *staff.Service) {
				<-start
				_, e := service.UpdateStaff(ctx, actor, target.ID, staff.StaffWrite{Email: target.Email, DisplayName: target.DisplayName, Role: "admin", Status: "active", CurrentPassword: fixturePassword, Revision: staff.Revision(target)})
				done <- e
			}(target, actor, service)
		}
		close(start)
		wins, conflicts := 0, 0
		for i := 0; i < 2; i++ {
			e := <-done
			if e == nil {
				wins++
			} else if errors.Is(e, staff.ErrConflict) {
				conflicts++
			} else {
				return e
			}
		}
		if wins != 1 || conflicts != 1 {
			return errors.New("last-superadmin race admitted both/neither")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT (SELECT count(*) FROM platform_staff WHERE role='superadmin' AND status='active')=1 AND (SELECT count(*) FROM platform_staff_sessions WHERE revoked_at IS NULL)=1 AND (SELECT count(*) FROM audit_events WHERE action='staff.update')=1`).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("race lost invariant or atomic audit/revocation")
		}
		return nil
	})
}

func TestBOGovernanceDisableReactivateAndSelfReset(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		in := bootstrapInput("target@example.test")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		target, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		if _, err = fixtureActor(ctx, p, target); err != nil {
			return err
		}
		for _, status := range []string{"disabled", "active"} {
			target, err = svc.UpdateStaff(ctx, a, target.ID, staff.StaffWrite{Email: target.Email, DisplayName: target.DisplayName, Role: target.Role, Status: status, Revision: staff.Revision(target), CurrentPassword: fixturePassword})
			if err != nil {
				return err
			}
		}
		var n int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_sessions WHERE staff_id=$1 AND revoked_at IS NULL`, target.ID).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return errors.New("reactivation revived sessions")
		}
		if err = svc.ReplacePassword(ctx, a, root.ID, staff.StaffWrite{Password: replacementPassword, CurrentPassword: fixturePassword, Revision: staff.Revision(root)}); err != nil {
			return err
		}
		in.Email = "after@example.test"
		in.CurrentPassword = replacementPassword
		if _, err = svc.CreateStaff(ctx, a, in); !errors.Is(err, staff.ErrUnauthenticated) {
			return errors.New("self reset retained authority")
		}
		return nil
	})
}

func waitGovernanceBlocked(ctx context.Context, p *pgxpool.Pool, blocker uint32) error {
	bounded, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		var found bool
		if err := p.QueryRow(bounded, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1::integer=ANY(pg_blocking_pids(pid)))`, blocker).Scan(&found); err != nil {
			return err
		}
		if found {
			return nil
		}
		if err := bounded.Err(); err != nil {
			return err
		}
	}
}

func TestBOGovernanceRevalidatesAfterPasswordWork(t *testing.T) {
	for _, kind := range []string{"role", "disable", "password"} {
		t.Run(kind, func(t *testing.T) {
			governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
				root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
				if err != nil {
					return err
				}
				a, err := fixtureActor(ctx, p, root)
				if err != nil {
					return err
				}
				input := bootstrapInput("second@example.test")
				input.CurrentPassword = fixturePassword
				if _, err = svc.CreateStaff(ctx, a, input); err != nil {
					return err
				}
				holder, err := p.Acquire(ctx)
				if err != nil {
					return err
				}
				defer holder.Release()
				tx, err := holder.Begin(ctx)
				if err != nil {
					return err
				}
				defer tx.Rollback(ctx)
				if _, err = tx.Exec(ctx, `UPDATE platform_staff_auth_control SET mutation_epoch=mutation_epoch+1 WHERE id=1`); err != nil {
					return err
				}
				done := make(chan error, 1)
				input.Email = "racing@example.test"
				go func() { _, e := svc.CreateStaff(ctx, a, input); done <- e }()
				// Observed server lock proves confirmation and hashing completed before
				// changing the credential/authority. No test-only KDF or scheduler sleep.
				if err = waitGovernanceBlocked(ctx, p, holder.Conn().PgConn().PID()); err != nil {
					_ = tx.Rollback(ctx)
					<-done
					return err
				}
				switch kind {
				case "role":
					_, err = tx.Exec(ctx, `UPDATE platform_staff SET role='admin' WHERE id=$1`, root.ID)
				case "disable":
					_, err = tx.Exec(ctx, `UPDATE platform_staff SET status='disabled' WHERE id=$1`, root.ID)
				case "password":
					_, err = tx.Exec(ctx, `UPDATE platform_staff_credentials SET changed_at=changed_at+interval '1 second' WHERE staff_id=$1`, root.ID)
				}
				if err != nil {
					_ = tx.Rollback(ctx)
					<-done
					return err
				}
				if err = tx.Commit(ctx); err != nil {
					<-done
					return err
				}
				err = <-done
				if !errors.Is(err, staff.ErrDenied) && !errors.Is(err, staff.ErrUnauthenticated) {
					return errors.New("stale confirmation/authority survived transaction fence")
				}
				var n int
				if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff WHERE email_normalized='racing@example.test'`).Scan(&n); err != nil {
					return err
				}
				if n != 0 {
					return errors.New("stale governance write committed")
				}
				return nil
			})
		})
	}
}

func TestBOGovernanceCustomerIdentityIndependent(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		customerHash, err := staff.NewPasswords().Hash(ctx, replacementPassword)
		if err != nil {
			return err
		}
		_, err = p.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES('11111111-1111-4111-8111-111111111111','shared@example.test','shared@example.test','Customer','active',$1,$1)`, fixtureNow)
		if err != nil {
			return err
		}
		_, err = p.Exec(ctx, `INSERT INTO password_credentials(user_id,password_hash,password_changed_at) VALUES('11111111-1111-4111-8111-111111111111',$2,$1)`, fixtureNow, customerHash)
		if err != nil {
			return err
		}
		root, err := svc.Bootstrap(ctx, bootstrapInput("shared@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		input := bootstrapInput("new@example.test")
		input.Role = "admin"
		input.CurrentPassword = replacementPassword
		if _, err = svc.CreateStaff(ctx, a, input); !errors.Is(err, staff.ErrDenied) {
			return errors.New("customer password authorized staff action")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT (SELECT password_hash FROM password_credentials)=$1 AND (SELECT count(*) FROM platform_staff)=1 AND NOT EXISTS(SELECT 1 FROM audit_events WHERE detail::text LIKE '%isolated%' OR detail::text LIKE '%argon2%')`, customerHash).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("customer credentials mutated or audit secrets leaked")
		}
		return nil
	})
}

func TestBOGovernanceAuthorizeMutationBoundary(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		in := bootstrapInput("admin@example.test")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		admin, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		b, err := fixtureActor(ctx, p, admin)
		if err != nil {
			return err
		}
		return db.InTx(ctx, p, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
			if _, err := staff.AuthorizeMutation(ctx, tx, a, "arbitrary_client_role", fixtureNow); !errors.Is(err, staff.ErrDenied) {
				return errors.New("arbitrary required role accepted")
			}
			result, err := staff.AuthorizeMutation(ctx, tx, a, "any_staff", fixtureNow)
			if err != nil {
				return err
			}
			if result.ID != root.ID {
				return errors.New("wrong current actor")
			}
			if _, err := staff.AuthorizeMutation(ctx, tx, b, "superadmin", fixtureNow); !errors.Is(err, staff.ErrDenied) {
				return errors.New("admin authorized superadmin mutation")
			}
			if _, err := staff.AuthorizeMutation(ctx, tx, b, "any_staff", fixtureNow); err != nil {
				return err
			}
			return nil
		})
	})
}

func TestBOGovernanceProvisionIndependentLegacyCredential(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		var legacy staff.Staff
		err = p.QueryRow(ctx, `INSERT INTO platform_staff(id,email_normalized,display_name,role,status,created_at,updated_at) VALUES(gen_random_uuid(),'legacy@example.test','Legacy','admin','active',$1,$1) RETURNING id,email_normalized,display_name,role,status,updated_at`, fixtureNow).Scan(&legacy.ID, &legacy.Email, &legacy.DisplayName, &legacy.Role, &legacy.Status, &legacy.UpdatedAt)
		if err != nil {
			return err
		}
		if err = svc.ReplacePassword(ctx, a, legacy.ID, staff.StaffWrite{Password: replacementPassword, CurrentPassword: fixturePassword, Revision: staff.Revision(legacy)}); err != nil {
			return err
		}
		var phc string
		if err = p.QueryRow(ctx, `SELECT password_hash FROM platform_staff_credentials WHERE staff_id=$1`, legacy.ID).Scan(&phc); err != nil {
			return err
		}
		valid, err := staff.NewPasswords().Verify(ctx, phc, replacementPassword)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("explicit legacy credential provisioning failed")
		}
		return nil
	})
}

func TestBOBootstrapConcurrent(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		done := make(chan error, 2)
		start := make(chan struct{})
		for _, email := range []string{"a@example.test", "b@example.test"} {
			go func(email string) { <-start; _, err := svc.Bootstrap(ctx, bootstrapInput(email)); done <- err }(email)
		}
		close(start)
		wins, conflicts := 0, 0
		for i := 0; i < 2; i++ {
			err := <-done
			if err == nil {
				wins++
			} else if errors.Is(err, staff.ErrConflict) {
				conflicts++
			} else {
				return err
			}
		}
		if wins != 1 || conflicts != 1 {
			return errors.New("bootstrap race not first-only")
		}
		return nil
	})
}

func TestBOGovernanceConfirmationAndRevocation(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		actor, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		in := bootstrapInput("admin@example.test")
		in.Role = "admin"
		in.CurrentPassword = "wrong fixture password"
		if _, err = svc.CreateStaff(ctx, actor, in); !errors.Is(err, staff.ErrDenied) {
			return errors.New("wrong confirmation allowed")
		}
		in.CurrentPassword = ""
		if _, err = svc.CreateStaff(ctx, actor, in); !errors.Is(err, staff.ErrDenied) {
			return errors.New("missing confirmation allowed")
		}
		in.CurrentPassword = fixturePassword
		target, err := svc.CreateStaff(ctx, actor, in)
		if err != nil {
			return err
		}
		admin, err := fixtureActor(ctx, p, target)
		if err != nil {
			return err
		}
		if _, err = svc.CreateStaff(ctx, admin, in); !errors.Is(err, staff.ErrDenied) {
			return errors.New("admin governance allowed")
		}
		update := staff.StaffWrite{Email: target.Email, DisplayName: "Edited", Role: "superadmin", Status: "active", Revision: staff.Revision(target), CurrentPassword: fixturePassword}
		updated, err := svc.UpdateStaff(ctx, actor, target.ID, update)
		if err != nil {
			return err
		}
		var active int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_sessions WHERE staff_id=$1 AND revoked_at IS NULL`, target.ID).Scan(&active); err != nil {
			return err
		}
		if active != 0 {
			return errors.New("role change failed revocation")
		}
		if _, err = svc.UpdateStaff(ctx, actor, target.ID, update); !errors.Is(err, staff.ErrConflict) {
			return errors.New("stale revision accepted")
		}
		if _, err = fixtureActor(ctx, p, updated); err != nil {
			return err
		}
		reset := staff.StaffWrite{Password: replacementPassword, CurrentPassword: fixturePassword, Revision: staff.Revision(updated)}
		if err = svc.ReplacePassword(ctx, actor, target.ID, reset); err != nil {
			return err
		}
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_sessions WHERE staff_id=$1 AND revoked_at IS NULL`, target.ID).Scan(&active); err != nil {
			return err
		}
		if active != 0 {
			return errors.New("password replacement failed revocation")
		}
		reset.Password = replacementPassword
		if err = svc.ReplacePassword(ctx, actor, target.ID, reset); !errors.Is(err, staff.ErrValidation) {
			return errors.New("same password accepted")
		}
		return nil
	})
}

func TestBOGovernanceProfileAndDuplicateEmail(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		in := bootstrapInput("  TARGET@EXAMPLE.TEST  ")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		target, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		if target.Email != "target@example.test" {
			return errors.New("email not normalized")
		}
		if _, err = svc.CreateStaff(ctx, a, in); !errors.Is(err, staff.ErrConflict) {
			return errors.New("duplicate staff email accepted")
		}
		if _, err = fixtureActor(ctx, p, target); err != nil {
			return err
		}
		if _, err = fixtureActor(ctx, p, target); err != nil {
			return err
		}
		update := staff.StaffWrite{Email: root.Email, DisplayName: "Edited", Role: target.Role, Status: target.Status, CurrentPassword: fixturePassword, Revision: staff.Revision(target)}
		if _, err = svc.UpdateStaff(ctx, a, target.ID, update); !errors.Is(err, staff.ErrConflict) {
			return errors.New("duplicate update accepted")
		}
		update.Email = "edited@example.test"
		edited, err := svc.UpdateStaff(ctx, a, target.ID, update)
		if err != nil {
			return err
		}
		if edited.DisplayName != "Edited" || staff.Revision(edited) == staff.Revision(target) {
			return errors.New("profile edit or revision missing")
		}
		var n int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_sessions WHERE staff_id=$1 AND revoked_at IS NULL`, target.ID).Scan(&n); err != nil {
			return err
		}
		if n != 2 {
			return errors.New("profile-only edit revoked sessions")
		}
		if err = svc.ReplacePassword(ctx, a, target.ID, staff.StaffWrite{Password: replacementPassword, CurrentPassword: fixturePassword, Revision: staff.Revision(edited)}); err != nil {
			return err
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM platform_staff_sessions WHERE staff_id=$1 AND revoked_at IS NULL) AND (SELECT count(*) FROM audit_events)=4 AND NOT EXISTS(SELECT 1 FROM users)`, target.ID).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("failed-action audit, independent identity or all-session revocation mismatch")
		}
		return nil
	})
}

func TestBOGovernanceRejectsInvalidSessions(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		in := bootstrapInput("target@example.test")
		in.CurrentPassword = fixturePassword
		target, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		wrong := a
		wrong.SessionHash[0] ^= 1
		swapped := a
		swapped.StaffID = target.ID
		expired := staff.NewService(p, func() time.Time { return fixtureNow.Add(30 * 24 * time.Hour) }, staff.NewPasswords())
		for _, c := range []struct {
			actor   staff.Actor
			service *staff.Service
		}{{wrong, svc}, {swapped, svc}, {a, expired}} {
			if _, err = c.service.CreateStaff(ctx, c.actor, in); !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("invalid session allowed create")
			}
			update := staff.StaffWrite{Email: target.Email, DisplayName: target.DisplayName, Role: target.Role, Status: target.Status, CurrentPassword: fixturePassword, Revision: staff.Revision(target)}
			if _, err = c.service.UpdateStaff(ctx, c.actor, target.ID, update); !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("invalid session allowed update")
			}
			update.Password = replacementPassword
			if err = c.service.ReplacePassword(ctx, c.actor, target.ID, update); !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("invalid session allowed reset")
			}
		}
		if _, err = p.Exec(ctx, `UPDATE platform_staff_sessions SET revoked_at=$1 WHERE token_hash=$2`, fixtureNow, a.SessionHash[:]); err != nil {
			return err
		}
		return db.InTx(ctx, p, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
			if _, err := staff.AuthorizeMutation(ctx, tx, a, "any_staff", fixtureNow); !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("revoked actor authorized mutation")
			}
			return nil
		})
	})
}

func TestBOGovernanceSamplesTimeAfterFence(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		var now atomic.Int64
		now.Store(fixtureNow.Unix())
		waiting := staff.NewService(p, func() time.Time { return time.Unix(now.Load(), 0).UTC() }, staff.NewPasswords())
		holder, err := p.Acquire(ctx)
		if err != nil {
			return err
		}
		defer holder.Release()
		tx, err := holder.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		if _, err = tx.Exec(ctx, `UPDATE platform_staff_auth_control SET mutation_epoch=mutation_epoch+1 WHERE id=1`); err != nil {
			return err
		}
		in := bootstrapInput("expired@example.test")
		in.CurrentPassword = fixturePassword
		done := make(chan error, 1)
		go func() { _, e := waiting.CreateStaff(ctx, a, in); done <- e }()
		if err = waitGovernanceBlocked(ctx, p, holder.Conn().PgConn().PID()); err != nil {
			_ = tx.Rollback(ctx)
			<-done
			return err
		}
		now.Store(fixtureNow.Add(30 * 24 * time.Hour).Unix())
		if err = tx.Commit(ctx); err != nil {
			<-done
			return err
		}
		if err = <-done; !errors.Is(err, staff.ErrUnauthenticated) {
			return errors.New("pre-fence time accepted expired actor")
		}
		return nil
	})
}
