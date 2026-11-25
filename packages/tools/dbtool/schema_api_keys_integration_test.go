//go:build integration

package dbtool

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func apiKeyInsert(id int, prefix, hashByte string) string {
	return fmt.Sprintf(`INSERT INTO api_keys(id,organization_id,name,token_prefix,token_hash,created_at) VALUES (@%d@,@30@,'Integration','%s',decode(repeat('%s',32),'hex'),'2026-01-01')`, id, prefix, hashByte)
}

// API key contracts: zk_ prefix, 32-byte unique hash, organization ownership.
func TestSchemaAPIKeys(t *testing.T) {
	schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) error {
		if err := identityFixture(ctx, p); err != nil {
			return err
		}
		if err := fixtureExec(ctx, p, apiKeyInsert(100, "zk_abcd", "01")); err != nil {
			return err
		}
		return checkNegatives(ctx, p, []negativeSQL{
			{"legacy live prefix", apiKeyInsert(101, "vc_live_abcd", "02"), "23514", "api_keys_prefix_check"},
			{"short prefix", apiKeyInsert(101, "zk_abc", "02"), "23514", "api_keys_prefix_check"},
			{"duplicate hash", apiKeyInsert(101, "zk_efgh", "01"), "23505", "api_keys_hash_key"},
			{"hash length", `INSERT INTO api_keys(id,organization_id,name,token_prefix,token_hash,created_at) VALUES (@101@,@30@,'Integration','zk_efgh','\x01','2026-01-01')`, "23514", "api_keys_hash_check"},
			{"organization FK", `INSERT INTO api_keys(id,organization_id,name,token_prefix,token_hash,created_at) VALUES (@101@,@199@,'Integration','zk_efgh',decode(repeat('03',32),'hex'),'2026-01-01')`, "23503", "api_keys_organization_fkey"},
			{"blank name", `INSERT INTO api_keys(id,organization_id,name,token_prefix,token_hash,created_at) VALUES (@101@,@30@,' ','zk_efgh',decode(repeat('03',32),'hex'),'2026-01-01')`, "23514", "api_keys_name_check"},
		})
	})
}
