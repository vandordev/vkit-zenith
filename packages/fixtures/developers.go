package fixtures

const (
	DemoOrgID           = "00000000-0000-4000-8000-000000000001"
	DemoAPIKey          = "zk_demofixturekey"
	SuspendedTestOrgID  = "00000000-0000-4000-8000-000000000002"
	SuspendedTestAPIKey = "zk_suspendedfixturekey"
)

type APIKey struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Prefix     string  `json:"prefix"`
	Status     string  `json:"status"`
	LastUsedAt *string `json:"lastUsedAt,omitempty"`
	ExpiresAt  *string `json:"expiresAt,omitempty"`
	RevokedAt  *string `json:"revokedAt,omitempty"`
	CreatedAt  string  `json:"createdAt"`
}

func ResolveAPIKey(raw string) (string, bool) {
	switch raw {
	case DemoAPIKey:
		return DemoOrgID, true
	case SuspendedTestAPIKey:
		return SuspendedTestOrgID, true
	default:
		return "", false
	}
}

type Organization struct {
	ID, Name, Slug, Status, CreatedAt string
}

// OrganizationFor returns the dummy organization an API key resolves to.
func OrganizationFor(id string) Organization {
	if id == SuspendedTestOrgID {
		return Organization{ID: id, Name: "Suspended Example", Slug: "suspended-example", Status: "suspended", CreatedAt: BaseDate}
	}
	return Organization{ID: id, Name: "Demo Organization", Slug: "demo", Status: "active", CreatedAt: BaseDate}
}
