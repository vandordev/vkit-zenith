//go:build integration

package dbtool

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertStaffFixture(ctx context.Context, p *pgxpool.Pool, id, user int) error {
	return fixtureExec(ctx, p, fmt.Sprintf(`INSERT INTO platform_staff(id,email_normalized,display_name,status,role,created_at,updated_at) SELECT @%d@,email_normalized,display_name,'active','admin','2026-01-01','2026-01-01' FROM users WHERE id=@%d@`, id, user))
}
