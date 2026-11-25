package fixtures

import "github.com/google/uuid"

const BaseDate = "2026-11-01T00:00:00Z"

func FixtureID(org uuid.UUID, key string) uuid.UUID { return uuid.NewSHA1(org, []byte(key)) }

type OrgData struct {
	APIKeys []APIKey
}

func Org(id uuid.UUID) OrgData {
	add := func(key string) string { return FixtureID(id, key).String() }
	return OrgData{APIKeys: []APIKey{
		{ID: add("api-key-active"), Name: "Kunci integrasi", Prefix: "zk_abcd", Status: "active", LastUsedAt: stringPtr(BaseDate), ExpiresAt: stringPtr("2026-12-01T00:00:00Z"), CreatedAt: "2026-10-01T00:00:00Z"},
		{ID: add("api-key-revoked"), Name: "Kunci lama", Prefix: "zk_old0", Status: "revoked", RevokedAt: stringPtr(BaseDate), CreatedAt: "2026-10-01T00:00:00Z"},
		{ID: add("api-key-expired"), Name: "Kunci kedaluwarsa", Prefix: "zk_old1", Status: "expired", ExpiresAt: stringPtr(BaseDate), CreatedAt: "2026-10-01T00:00:00Z"},
	}}
}

func stringPtr(v string) *string { return &v }
