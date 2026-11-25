//go:build integration

package securitytest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func authSchemaUserLock() (string, error) {
	b, err := os.ReadFile("../../../../packages/application/identity/locking.sql")
	if err != nil {
		return "", errors.New("identity lock source unavailable")
	}
	for _, part := range strings.Split(string(b), "-- name: ")[1:] {
		header, query, _ := strings.Cut(part, "\n")
		if header == "IdentityLockUser :one" {
			return strings.TrimSpace(query), nil
		}
	}
	return "", errors.New("identity lock source missing")
}

func authSchemaWaitPID(ctx context.Context, p *pgxpool.Pool, pid uint32) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		var blocked bool
		if err := p.QueryRow(ctx, `SELECT cardinality(pg_blocking_pids($1))>0`, pid).Scan(&blocked); err != nil {
			return err
		}
		if blocked {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.New("expected schema writer barrier unavailable")
		case <-tick.C:
		}
	}
}

// Schema-stage candidate writer only: it executes the actual application user
// fence and rechecks live eligibility before storing new-generation material.
// This is NOT the future registration/session implementation. Tasks 12–13
// must repeat the interleavings through those actual commands/provider hooks.
func authSchemaWriteAction(ctx context.Context, tx pgx.Tx, id pgtype.UUID) (bool, error) {
	lock, err := authSchemaUserLock()
	if err != nil {
		return false, err
	}
	if _, err = tx.Exec(ctx, lock, id); err != nil {
		return false, err
	}
	var eligible bool
	if err = tx.QueryRow(ctx, `SELECT u.status='active' AND u.email_normalized='person@example.test' AND u.identity_version=1 AND c.generation=1 FROM users u JOIN password_credentials c ON c.user_id=u.id WHERE u.id=$1`, id).Scan(&eligible); err != nil {
		return false, err
	}
	if !eligible {
		return false, nil
	}
	_, err = tx.Exec(ctx, `INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation) VALUES(gen_random_uuid(),$1,'reset_password','person@example.test',sha256('candidate'::bytea),$2,$2::timestamptz+interval '30 minutes',1,1)`, id, fixtureNow)
	return err == nil, err
}

func TestAuthSchemaBOUserWriterRaces(t *testing.T) {
	for _, kind := range []string{"disable", "email"} {
		for _, writerFirst := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/writer-first=%t", kind, writerFirst), func(t *testing.T) {
				governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
					ctx, cancel := context.WithTimeout(ctx, 40*time.Second)
					defer cancel()
					commands, actor, err := customerCommands(ctx, p, svc)
					if err != nil {
						return err
					}
					user, err := commands.CreateUser(ctx, actor, customers.UserCreate{Email: "person@example.test", DisplayName: "Person"})
					if err != nil {
						return err
					}
					if _, err = p.Exec(ctx, `UPDATE users SET identity_version=1 WHERE id=$1`, user.ID); err != nil {
						return err
					}
					if _, err = p.Exec(ctx, `INSERT INTO password_credentials(user_id,password_hash,password_changed_at,generation) VALUES($1,'fake-schema-credential',$2,1)`, user.ID, fixtureNow); err != nil {
						return err
					}
					if _, err = p.Exec(ctx, `INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation) VALUES(gen_random_uuid(),$1,'reset_password','person@example.test',sha256('initial'::bytea),$2,$2::timestamptz+interval '30 minutes',1,1)`, user.ID, fixtureNow); err != nil {
						return err
					}
					mutate := func() error {
						if kind == "disable" {
							_, e := commands.SetUserStatus(ctx, actor, user.ID, customers.UserStatus{Status: "disabled", Sensitive: sensitive(customers.UserRevision(user))})
							return e
						}
						_, e := commands.ChangeUserEmail(ctx, actor, user.ID, customers.UserEmail{Email: "replacement@example.test", Sensitive: sensitive(customers.UserRevision(user))})
						return e
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
					if writerFirst {
						issued, err := authSchemaWriteAction(ctx, tx, user.ID)
						if err != nil {
							return err
						}
						if !issued {
							return errors.New("candidate positive control rejected")
						}
						done := make(chan error, 1)
						go func() { done <- mutate() }()
						barrier := waitGovernanceBlocked(ctx, p, holder.Conn().PgConn().PID())
						if err = tx.Commit(ctx); err != nil {
							cancel()
							<-done
							return err
						}
						if err = <-done; err != nil {
							return err
						}
						if barrier != nil {
							return barrier
						}
					} else {
						// Pause the real BO command at its last downstream fence, after
						// it acquired the user. This holder never later acquires user.
						if _, err = tx.Exec(ctx, `SELECT id FROM email_action_tokens WHERE user_id=$1 FOR UPDATE`, user.ID); err != nil {
							return err
						}
						boDone := make(chan error, 1)
						go func() { boDone <- mutate() }()
						if err = waitGovernanceBlocked(ctx, p, holder.Conn().PgConn().PID()); err != nil {
							cancel()
							tx.Rollback(ctx)
							<-boDone
							return err
						}
						writer, err := p.Acquire(ctx)
						if err != nil {
							cancel()
							tx.Rollback(ctx)
							<-boDone
							return err
						}
						defer writer.Release()
						candidate, err := writer.Begin(ctx)
						if err != nil {
							cancel()
							tx.Rollback(ctx)
							<-boDone
							return err
						}
						defer candidate.Rollback(ctx)
						type result struct {
							issued bool
							err    error
						}
						writeDone := make(chan result, 1)
						go func() { issued, e := authSchemaWriteAction(ctx, candidate, user.ID); writeDone <- result{issued, e} }()
						barrier := authSchemaWaitPID(ctx, p, writer.Conn().PgConn().PID())
						if err = tx.Rollback(ctx); err != nil {
							cancel()
							<-boDone
							<-writeDone
							return err
						}
						boErr := <-boDone
						out := <-writeDone
						if boErr != nil {
							return boErr
						}
						if out.err != nil {
							return out.err
						}
						if barrier != nil {
							return barrier
						}
						if out.issued {
							return errors.New("candidate issued after committed BO revocation")
						}
						if err = candidate.Commit(ctx); err != nil {
							return err
						}
					}
					var total, live int
					if err = p.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE revoked_at IS NULL) FROM email_action_tokens WHERE user_id=$1`, user.ID).Scan(&total, &live); err != nil {
						return err
					}
					want := 1
					if writerFirst {
						want = 2
					}
					if total != want || live != 0 {
						return errors.New("BO race left usable candidate material or lost history")
					}
					return nil
				})
			})
		}
	}
}
