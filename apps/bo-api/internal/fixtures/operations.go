package fixtures

import "fmt"

func Jobs() []Job {
	states := []string{"available", "scheduled", "running", "retryable", "completed", "cancelled", "discarded"}
	out := make([]Job, 0, len(states))
	for i, state := range states {
		errs := make([]string, 0)
		if state == "retryable" || state == "discarded" {
			errs = []string{"Upstream request timed out"}
		}
		out = append(out, Job{ID: fmt.Sprint(i + 1), Kind: []string{"email.send", "report.generate", "email.send", "cleanup.sessions", "report.generate", "email.send", "cleanup.sessions"}[i], Queue: "default", State: state, Attempt: fmt.Sprint(i % 3), MaxAttempts: "10", Errors: errs, Args: map[string]any{"resourceId": "fixture-resource"}, CreatedAt: "2026-11-01T00:00:00Z"})
	}
	return out
}
