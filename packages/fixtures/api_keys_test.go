package fixtures

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
)

// Mirrors api_keys_prefix_check in packages/db/migrations/00001_baseline.sql.
var apiKeyPrefix = regexp.MustCompile(`^zk_[A-Za-z0-9]{4,12}$`)

func TestAPIKeyFixturesMatchSchemaPrefix(t *testing.T) {
	for _, key := range Org(uuid.MustParse(DemoOrgID)).APIKeys {
		if !apiKeyPrefix.MatchString(key.Prefix) {
			t.Errorf("fixture prefix %q violates api_keys_prefix_check", key.Prefix)
		}
	}
	for _, raw := range []string{DemoAPIKey, SuspendedTestAPIKey} {
		if !regexp.MustCompile(`^zk_`).MatchString(raw) {
			t.Errorf("fixture key %q lacks the zk_ prefix", raw)
		}
	}
}
