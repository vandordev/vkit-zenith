package fixtures_test

import (
	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	v1 "github.com/vandordev/vkit-zenith/apps/bo-api/internal/handlers/v1"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var secretField = regexp.MustCompile(`(?i)(hash|ciphertext|secret$|token$|password|^pan$|cardnumber|cvv)`)
var allowedSecretMetadata = map[string]bool{"hasSecret": true, "HasSecret": true, "hasProviderToken": true, "HasProviderToken": true, "tokenPrefix": true, "TokenPrefix": true, "apiKeyPrefix": true, "APIKeyPrefix": true}

func findSecret(t reflect.Type, seen map[reflect.Type]bool) bool {
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		return findSecret(t.Elem(), seen)
	}
	if t.Kind() == reflect.Map {
		return findSecret(t.Key(), seen) || findSecret(t.Elem(), seen)
	}
	if t.Kind() != reflect.Struct || seen[t] {
		return false
	}
	seen[t] = true
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := strings.Split(f.Tag.Get("json"), ",")[0]
		if (secretField.MatchString(f.Name) && !allowedSecretMetadata[f.Name]) || (tag != "-" && secretField.MatchString(tag) && !allowedSecretMetadata[tag]) || findSecret(f.Type, seen) {
			return true
		}
	}
	return false
}

func findSecretValue(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return false
		}
		return findSecretValue(v.Elem())
	case reflect.Map:
		if v.IsNil() {
			return false
		}
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			if k.Kind() == reflect.String && (secretField.MatchString(k.String()) && !allowedSecretMetadata[k.String()]) {
				return true
			}
			if findSecretValue(iter.Value()) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if findSecretValue(v.Index(i)) {
				return true
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			tag := strings.Split(f.Tag.Get("json"), ",")[0]
			if (secretField.MatchString(f.Name) && !allowedSecretMetadata[f.Name]) || (tag != "-" && secretField.MatchString(tag) && !allowedSecretMetadata[tag]) {
				return true
			}
			if f.IsExported() && findSecretValue(v.Field(i)) {
				return true
			}
		}
	}
	return false
}
func TestFixturesAndDummyDTOsContainNoSecrets(t *testing.T) {
	roots := []reflect.Type{reflect.TypeOf(fixtures.APIKey{}), reflect.TypeOf(fixtures.History{}), reflect.TypeOf(fixtures.OrgData{}), reflect.TypeOf(v1.V1Actor{}), reflect.TypeOf(v1.V1HistoryEntry{}), reflect.TypeOf(v1.V1Adjustment{}), reflect.TypeOf(v1.V1DummyResult[fixtures.APIKey]{}), reflect.TypeOf(v1.V1HistoryListBody{}), reflect.TypeOf(v1.V1HistoryListOutput{}), reflect.TypeOf(v1.V1ListAPIKeysBody{}), reflect.TypeOf(v1.V1ListAPIKeysOutput{}), reflect.TypeOf(v1.V1RevokeAPIKeyBody{}), reflect.TypeOf(v1.V1RevokeAPIKeyOutput{})}
	roots = append(roots, reflect.TypeOf(fixtures.Job{}), reflect.TypeOf(v1.V1HomeCounter{}), reflect.TypeOf(v1.V1HomeCountersBody{}), reflect.TypeOf(v1.V1HomeCountersOutput{}), reflect.TypeOf(v1.V1OpsGetInput{}), reflect.TypeOf(v1.V1OpsActionInput{}), reflect.TypeOf(v1.V1JobsListInput{}), reflect.TypeOf(v1.V1OpsDetail[fixtures.Job]{}), reflect.TypeOf(v1.V1OpsDetailOutput[fixtures.Job]{}), reflect.TypeOf(v1.V1OpsActionOutput[fixtures.Job]{}), reflect.TypeOf(v1.V1OpsListBody[fixtures.Job]{}), reflect.TypeOf(v1.V1OpsListOutput[fixtures.Job]{}))
	roots = append(roots, reflect.TypeOf(fixtures.UserSession{}), reflect.TypeOf(fixtures.StaffSession{}), reflect.TypeOf(fixtures.EmailAction{}), reflect.TypeOf(v1.UserSessionsInput{}), reflect.TypeOf(v1.V1UserSession{}), reflect.TypeOf(v1.UserSessionsBody{}), reflect.TypeOf(v1.UserSessionsOutput{}), reflect.TypeOf(v1.UserEmailActionsInput{}), reflect.TypeOf(v1.V1UserEmailAction{}), reflect.TypeOf(v1.UserEmailActionsBody{}), reflect.TypeOf(v1.UserEmailActionsOutput{}), reflect.TypeOf(v1.UserHistoryInput{}), reflect.TypeOf(v1.UserHistoryBody{}), reflect.TypeOf(v1.UserHistoryOutput{}), reflect.TypeOf(v1.StaffSessionsInput{}), reflect.TypeOf(v1.V1StaffSession{}), reflect.TypeOf(v1.StaffSessionsBody{}), reflect.TypeOf(v1.StaffSessionsOutput{}), reflect.TypeOf(v1.StaffHistoryInput{}), reflect.TypeOf(v1.StaffHistoryBody{}), reflect.TypeOf(v1.StaffHistoryOutput{}))
	for _, typ := range roots {
		if findSecret(typ, map[reflect.Type]bool{}) {
			t.Errorf("%s contains sensitive field", typ)
		}
	}
	for _, data := range fixtures.All() {
		if findSecretValue(reflect.ValueOf(data)) {
			t.Fatal("fixture contains a secret key")
		}
	}
	identityID := uuid.MustParse("00000000-0000-4000-8000-000000000101")
	for _, payload := range []any{fixtures.UserSessions(identityID), fixtures.UserEmailActions(identityID), fixtures.StaffSessions(identityID), fixtures.UserHistory(identityID), fixtures.StaffHistory(identityID)} {
		if findSecretValue(reflect.ValueOf(payload)) {
			t.Fatal("identity context fixture contains a secret key")
		}
	}
	if findSecretValue(reflect.ValueOf(fixtures.Jobs())) {
		t.Fatal("job fixture contains a secret key")
	}
	type nestedSecret struct {
		Payload struct {
			Token string `json:"token"`
		}
	}
	if !findSecret(reflect.TypeOf(nestedSecret{}), map[reflect.Type]bool{}) {
		t.Fatal("nested secret was not detected")
	}
	negative := map[string]any{"payload": []any{map[string]any{"token": ""}}}
	if !findSecretValue(reflect.ValueOf(negative)) {
		t.Fatal("nested valueless token key was not detected")
	}
}
