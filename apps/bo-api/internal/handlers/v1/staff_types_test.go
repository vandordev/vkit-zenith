package v1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestStaffOperationUpdatedAtUTC(t *testing.T) {
	for _, mapping := range []struct {
		name    string
		mapBody func(staff.Staff) any
	}{
		{"list", func(s staff.Staff) any { return v1StaffListItem(s) }},
		{"detail", func(s staff.Staff) any { return v1StaffDetailBody(s) }},
		{"create", func(s staff.Staff) any { return v1StaffCreateBody(s) }},
		{"update", func(s staff.Staff) any { return v1StaffUpdateBody(s) }},
	} {
		for _, offset := range []int{7 * 60 * 60, -5 * 60 * 60, 0} {
			zone := time.FixedZone("fixture", offset)
			t.Run(mapping.name+"/"+(time.Duration(offset)*time.Second).String(), func(t *testing.T) {
				updatedAt := time.Date(2026, 10, 25, 11, 56, 8, 775333000, time.UTC).In(zone)
				source := staff.Staff{
					Email:       "fixture@example.test",
					DisplayName: "Fixture",
					Role:        "superadmin",
					Status:      "active",
					UpdatedAt:   updatedAt,
				}
				got := mapping.mapBody(source)
				encoded, err := json.Marshal(got)
				if err != nil {
					t.Fatal(err)
				}
				var body struct {
					UpdatedAt string `json:"updatedAt"`
					Revision  string `json:"revision"`
				}
				if err := json.Unmarshal(encoded, &body); err != nil {
					t.Fatal(err)
				}
				if body.UpdatedAt != "2026-10-25T11:56:08.775333Z" {
					t.Fatalf("updatedAt = %q, want canonical UTC timestamp", body.UpdatedAt)
				}
				instant, err := time.Parse(time.RFC3339Nano, body.UpdatedAt)
				if err != nil || !instant.Equal(updatedAt) {
					t.Fatal("timestamp normalization changed the instant")
				}
				if body.Revision != staff.Revision(source) {
					t.Fatal("timestamp normalization changed the revision")
				}
				if source.UpdatedAt.Location() != zone {
					t.Fatal("operation mapping changed the source timezone")
				}
			})
		}
	}
}
