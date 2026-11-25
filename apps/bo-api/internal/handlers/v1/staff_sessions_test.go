package v1

import (
	"testing"

	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
)

func TestStaffSessionSearchUsesSafeMetadata(t *testing.T) {
	id := uuid.MustParse("00000000-0000-4000-8000-000000000101")
	items := fixtures.StaffSessions(id)
	if got := filterStaffSessions(items, "2026-11-01T"); len(got) != 2 {
		t.Fatalf("created timestamp search returned %d items", len(got))
	}
	if got := filterStaffSessions(items, "no-match"); len(got) != 0 || got == nil {
		t.Fatalf("empty staff session search should return a non-nil empty slice, got %#v", got)
	}
}
