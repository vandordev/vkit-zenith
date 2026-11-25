package fixtures

import "github.com/google/uuid"

var ErrorOrgID = uuid.MustParse("00000000-0000-4000-8000-0000000000e5")
var EmptyOrgID = uuid.MustParse("00000000-0000-4000-8000-0000000000e0")

func FixtureID(org uuid.UUID, key string) uuid.UUID { return uuid.NewSHA1(org, []byte(key)) }

type APIKey struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	TokenPrefix string  `json:"tokenPrefix"`
	LastUsedAt  *string `json:"lastUsedAt"`
	ExpiresAt   *string `json:"expiresAt"`
	RevokedAt   *string `json:"revokedAt"`
	Status      string  `json:"status"`
}
type History struct {
	ID           string  `json:"id"`
	OccurredAt   string  `json:"occurredAt"`
	ActorKind    string  `json:"actorKind"`
	ActorID      *string `json:"actorId"`
	ActorName    string  `json:"actorName"`
	Role         *string `json:"role"`
	APIKeyPrefix *string `json:"apiKeyPrefix"`
	System       *string `json:"system"`
	Action       string  `json:"action"`
	Outcome      string  `json:"outcome"`
	TargetType   *string `json:"targetType"`
	TargetID     *string `json:"targetId"`
	Reason       *string `json:"reason"`
}
type UserSession struct {
	ID              string  `json:"id"`
	AuthenticatedAt string  `json:"authenticatedAt"`
	LastSeenAt      string  `json:"lastSeenAt"`
	ExpiresAt       string  `json:"expiresAt"`
	RevokedAt       *string `json:"revokedAt"`
	CreatedAt       string  `json:"createdAt"`
}
type StaffSession struct {
	ID        string  `json:"id"`
	CreatedAt string  `json:"createdAt"`
	ExpiresAt string  `json:"expiresAt"`
	RevokedAt *string `json:"revokedAt"`
}
type EmailAction struct {
	ID         string  `json:"id"`
	Purpose    string  `json:"purpose"`
	CreatedAt  string  `json:"createdAt"`
	ExpiresAt  string  `json:"expiresAt"`
	ConsumedAt *string `json:"consumedAt"`
	RevokedAt  *string `json:"revokedAt"`
}
type Job struct {
	ID          string         `json:"id"`
	Kind        string         `json:"kind"`
	Queue       string         `json:"queue"`
	State       string         `json:"state"`
	Attempt     string         `json:"attempt"`
	MaxAttempts string         `json:"maxAttempts"`
	Errors      []string       `json:"errors" nullable:"false"`
	Args        map[string]any `json:"args"`
	CreatedAt   string         `json:"createdAt"`
}
type OrgData struct {
	History []History
	ID      uuid.UUID
	APIKeys []APIKey
}
