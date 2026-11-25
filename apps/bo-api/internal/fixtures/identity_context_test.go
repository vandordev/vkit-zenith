package fixtures_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
)

func TestIdentityContextFixturesAreDeterministicAndTargetRequestedIdentity(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000101")
	if got, want := fixtures.UserSessions(id), fixtures.UserSessions(id); len(got) != len(want) || got[0].ID != want[0].ID {
		t.Fatal("user session fixtures are not deterministic")
	}
	if got := fixtures.UserHistory(id); len(got) == 0 || got[0].TargetID == nil || *got[0].TargetID != id.String() {
		t.Fatal("user history does not target the requested user")
	}
	if got := fixtures.StaffHistory(id); len(got) == 0 || got[0].TargetID == nil || *got[0].TargetID != id.String() {
		t.Fatal("staff history does not target the requested staff record")
	}
	if got := fixtures.UserEmailActions(id); len(got) == 0 || got[0].Purpose == "" {
		t.Fatal("email action fixture is missing purpose metadata")
	}
}
