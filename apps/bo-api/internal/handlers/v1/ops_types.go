package v1

import (
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"time"
)

type V1OpsListBody[T any] struct {
	Items      []T                          `json:"items" nullable:"false"`
	Pagination V1OrganizationListPagination `json:"pagination"`
}
type V1OpsListOutput[T any] struct{ Body V1OpsListBody[T] }
type V1OpsDetail[T any] struct {
	Item    T                `json:"item"`
	History []V1HistoryEntry `json:"history" nullable:"false"`
}
type V1OpsDetailOutput[T any] struct{ Body V1OpsDetail[T] }
type V1HomeCounter struct {
	Count string `json:"count"`
	Href  string `json:"href"`
}
type V1HomeCountersBody struct {
	FailedJobs V1HomeCounter `json:"failedJobs"`
}
type V1HomeCountersOutput struct{ Body V1HomeCountersBody }
type V1OpsActionInput struct {
	ID   string `path:"id"`
	Body V1Adjustment
}
type V1OpsActionOutput[T any] struct{ Body V1DummyResult[T] }
type V1OpsGetInput struct {
	ID string `path:"id"`
}
type V1JobsListInput struct {
	CustomerPage
	Status     string `query:"status" enum:",available,scheduled,running,retryable,completed,cancelled,discarded"`
	StateGroup string `query:"stateGroup" enum:",failed"`
}

func isFailedJobState(state string) bool { return state == "retryable" || state == "discarded" }

func opsHistory(id, target string) []V1HistoryEntry {
	when, _ := time.Parse(time.RFC3339, "2026-11-01T00:02:00Z")
	system := "worker.example_job"
	return []V1HistoryEntry{{ID: "history-" + id, OccurredAt: when, Actor: V1Actor{Kind: "system", Name: strPtr("Sistem"), System: &system}, Action: "operation.received", Outcome: "success", TargetType: strPtr(target), TargetID: &id}, {ID: "history-staff-" + id, OccurredAt: when.Add(-time.Minute), Actor: V1Actor{Kind: "staff", Name: strPtr("Staf Fixture"), Role: strPtr("admin")}, Action: "operation.reviewed", Outcome: "success", TargetType: strPtr(target), TargetID: &id}}
}
func opsFindJob(id string) (fixtures.Job, []V1HistoryEntry, bool) {
	for _, x := range fixtures.Jobs() {
		if x.ID == id {
			return x, opsHistory(id, "job"), true
		}
	}
	return fixtures.Job{}, []V1HistoryEntry{}, false
}
