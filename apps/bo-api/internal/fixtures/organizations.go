package fixtures

import (
	"fmt"

	"github.com/google/uuid"
)

func Org(id uuid.UUID) (OrgData, bool) {
	d := OrgData{ID: id}
	if id == EmptyOrgID || id == ErrorOrgID {
		return d, true
	}
	{
		used, expiry := "2026-11-01T00:00:00Z", "2026-12-01T00:00:00Z"
		keyID := uuid.NewSHA1(id, []byte("key")).String()
		d.APIKeys = append(d.APIKeys,
			APIKey{ID: keyID, Name: "Fixture API key", TokenPrefix: "zk_abcd", LastUsedAt: &used, ExpiresAt: &expiry, Status: "active"},
			APIKey{ID: uuid.NewSHA1(id, []byte("expired")).String(), Name: "Expired key", TokenPrefix: "zk_old0", ExpiresAt: strFixture("2026-11-02T00:00:00Z"), Status: "expired"},
			APIKey{ID: uuid.NewSHA1(id, []byte("revoked")).String(), Name: "Revoked key", TokenPrefix: "zk_rev0", RevokedAt: strFixture("2026-11-01T00:00:00Z"), Status: "revoked"},
			APIKey{ID: uuid.NewSHA1(id, []byte("unused")).String(), Name: "Unused key", TokenPrefix: "zk_new0", Status: "active"})
		staffRole, memberRole := "admin", "owner"
		orgID := id.String()
		for i, sys := range []string{"worker.example_job", "scheduler.example_task"} {
			system := sys
			var reason *string
			if i == 0 {
				reason = strFixture("Fixture review reason")
			}
			d.History = append(d.History, History{ID: uuid.NewSHA1(id, []byte(fmt.Sprintf("history-%d", i))).String(), OccurredAt: "2026-11-01T00:00:00Z", ActorKind: "system", ActorName: "Sistem", System: &system, Action: "organization.updated", Outcome: []string{"success", "failure"}[i], Reason: reason, TargetType: strFixture("organization"), TargetID: &orgID})
		}
		d.History = append(d.History,
			History{ID: uuid.NewSHA1(id, []byte("staff-history")).String(), OccurredAt: "2026-11-01T00:00:00Z", ActorKind: "staff", ActorName: "Staf Fixture", Role: &staffRole, TargetType: strFixture("organization"), TargetID: &orgID, Action: "organization.updated", Outcome: "success"},
			History{ID: uuid.NewSHA1(id, []byte("user-history")).String(), OccurredAt: "2026-11-01T00:00:00Z", ActorKind: "user", ActorName: "Pengguna Fixture", Role: &memberRole, TargetType: strFixture("api_key"), TargetID: &keyID, Action: "api_key.created", Outcome: "success"},
			History{ID: uuid.NewSHA1(id, []byte("client-history")).String(), OccurredAt: "2026-11-01T00:00:00Z", ActorKind: "client", ActorName: "API key fixture", APIKeyPrefix: strFixture("zk_abcd"), TargetType: strFixture("api_key"), TargetID: &keyID, Action: "api_key.used", Outcome: "success"},
			History{ID: uuid.NewSHA1(id, []byte("anonymous-history")).String(), OccurredAt: "2026-11-01T00:00:00Z", ActorKind: "anonymous", ActorName: "Publik", Action: "auth.login", Outcome: "denied"})
	}
	return d, true
}

func strFixture(s string) *string { return &s }

func All() []OrgData {
	ids := []uuid.UUID{uuid.MustParse("00000000-0000-4000-8000-000000000001"), uuid.MustParse("00000000-0000-4000-8000-000000000002")}
	all := make([]OrgData, 0, len(ids))
	for _, id := range ids {
		d, _ := Org(id)
		all = append(all, d)
	}
	return all
}
