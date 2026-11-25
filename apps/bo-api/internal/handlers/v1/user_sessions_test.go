package v1

import (
	"testing"

	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
)

func TestUserSessionAndEmailActionSearchUsesSafeMetadata(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000101")
	if got := filterUserSessions(fixtures.UserSessions(id), "2026-11-02"); len(got) != 2 {
		t.Fatalf("session timestamp search returned %d items", len(got))
	}
	if got := filterUserEmailActions(fixtures.UserEmailActions(id), "RESET_PASSWORD"); len(got) != 1 || got[0].Purpose != "reset_password" {
		t.Fatalf("purpose search returned %#v", got)
	}
	if got := filterUserEmailActions(fixtures.UserEmailActions(id), "PENDING"); len(got) != 2 {
		t.Fatalf("lifecycle status search returned %d items", len(got))
	}
	if got := filterUserSessions(fixtures.UserSessions(id), "no-match"); len(got) != 0 || got == nil {
		t.Fatalf("empty session search should return a non-nil empty slice, got %#v", got)
	}
}
