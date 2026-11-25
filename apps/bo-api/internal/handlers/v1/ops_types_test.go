package v1

import "testing"

func TestFailedJobStateGroupMatchesCounterAndList(t *testing.T) {
	for _, state := range []string{"retryable", "discarded"} {
		if !isFailedJobState(state) {
			t.Errorf("state %q should be counted and listed as failed/retrying", state)
		}
	}
	for _, state := range []string{"available", "scheduled", "running", "completed", "cancelled"} {
		if isFailedJobState(state) {
			t.Errorf("state %q should not be counted as failed/retrying", state)
		}
	}
}
