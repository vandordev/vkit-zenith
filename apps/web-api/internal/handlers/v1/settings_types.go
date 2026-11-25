package v1

import "github.com/vandordev/vkit-zenith/packages/fixtures"

type SettingsListBase struct {
	OrganizationID string `path:"organizationId"`
	Q              string `query:"q" maxLength:"200"`
	Page           int    `query:"page" minimum:"1" default:"1"`
	Size           int    `query:"size" minimum:"1" maximum:"100" default:"20"`
}
type apiKeysListInput struct {
	SettingsListBase
	Status string `query:"status" enum:",active,revoked,expired"`
}
type settingsList[T any] struct {
	CacheControl string `header:"Cache-Control"`
	Body         V1List[T]
}
type settingsWriteOutput[T any] struct {
	CacheControl string `header:"Cache-Control"`
	Body         V1DummyResult[T]
}
type V1CreateApiKeyResult struct {
	Key    fixtures.APIKey `json:"key"`
	Secret string          `json:"secret"`
}
type apiKeyCreateOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         V1DummyResult[V1CreateApiKeyResult]
}
type logoResult struct {
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
}
