package v1

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type V1Actor struct {
	Kind         string  `json:"kind" enum:"user,staff,client,system,anonymous"`
	ID           *string `json:"id"`
	Name         *string `json:"name"`
	Role         *string `json:"role"`
	APIKeyPrefix *string `json:"apiKeyPrefix"`
	System       *string `json:"system"`
}
type V1HistoryEntry struct {
	ID         string    `json:"id"`
	OccurredAt time.Time `json:"occurredAt"`
	Actor      V1Actor   `json:"actor"`
	Action     string    `json:"action"`
	Outcome    string    `json:"outcome" enum:"success,failure,denied"`
	TargetType *string   `json:"targetType"`
	TargetID   *string   `json:"targetId"`
	Reason     *string   `json:"reason"`
}
type V1Adjustment struct {
	Reason string `json:"reason" minLength:"10" maxLength:"1024"`
}
type V1DummyResult[T any] struct {
	Result T    `json:"result"`
	Dummy  bool `json:"dummy"`
}

func stateConflict() error {
	return &huma.ErrorModel{Status: 409, Title: "Conflict", Detail: "Operasi tidak dapat dilakukan pada data ini."}
}
