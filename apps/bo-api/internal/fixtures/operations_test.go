package fixtures_test

import (
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"testing"
)

func TestOperationsFixturesCoverCanonicalStates(t *testing.T) {
	jobs := map[string]bool{}
	for _, job := range fixtures.Jobs() {
		jobs[job.State] = true
		if job.Errors == nil || job.Args == nil {
			t.Errorf("job %s has a nil collection/object", job.ID)
		}
		if job.Attempt == "" || job.MaxAttempts != "10" {
			t.Errorf("job %s has invalid decimal count fields", job.ID)
		}
	}
	for _, state := range []string{"available", "scheduled", "running", "retryable", "completed", "cancelled", "discarded"} {
		if !jobs[state] {
			t.Errorf("missing River state %q", state)
		}
	}
}
