package v1

import (
	"github.com/vandordev/vkit-zenith/packages/fixtures"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/handlers/v1/testtypes"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var dummySecretField = regexp.MustCompile(`(?i)(hash|ciphertext|secret|token|password|objectkey|^pan$|cardnumber|cvv)`)

func TestDummyResponseDTOsContainNoSecretFields(t *testing.T) {
	for _, root := range []any{V1Page{}, V1List[fixtures.APIKey]{}, settingsList[fixtures.APIKey]{}, settingsWriteOutput[fixtures.APIKey]{}, apiKeyCreateOutput{}, logoUploadOutput{}, V1DummyResult[fixtures.APIKey]{}, webOrganizationProfileOutput{}, webOrganizationMembersOutput{}, webOrganizationsOutput{}} {
		checkDummyDTO(t, reflect.TypeOf(root))
	}
}

func TestDummySecretScannerParsesJSONTagOptions(t *testing.T) {
	f, _ := reflect.TypeOf(struct {
		PAN string `json:"pan,omitempty"`
	}{}).FieldByName("PAN")
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if !dummySecretField.MatchString(name) {
		t.Fatal("pan,omitempty was not recognized as a secret key")
	}
}

func TestDummySecretScannerRejectsSecretGoNamesWithSafeJSONAliases(t *testing.T) {
	for _, alias := range []string{"prefix", "tokenPrefix", "hasCredentials"} {
		f, _ := reflect.TypeOf(struct {
			Secret string `json:"prefix"`
		}{}).FieldByName("Secret")
		f.Tag = reflect.StructTag(`json:"` + alias + `"`)
		if !dummyFieldIsSecret(reflect.TypeOf(struct{}{}), f) {
			t.Errorf("secret Go field was allowed by JSON alias %q", alias)
		}
	}
}

func TestDummySecretScannerFindsSnakeCaseJSONNames(t *testing.T) {
	typ := reflect.TypeOf(struct {
		LogoKey string `json:"logo_object_key"`
		Card    string `json:"card_number"`
	}{})
	for _, name := range []string{"LogoKey", "Card"} {
		field, _ := typ.FieldByName(name)
		if !dummyFieldIsSecret(typ, field) {
			t.Errorf("snake-case secret JSON name on %s was not detected", name)
		}
	}
}

func TestDummySecretScannerAllowsOnlyExactCreationResultTypes(t *testing.T) {
	for _, tc := range []struct {
		typ   reflect.Type
		field string
	}{{reflect.TypeOf(V1CreateApiKeyResult{}), "Secret"}} {
		field, ok := tc.typ.FieldByName(tc.field)
		if !ok || dummyFieldIsSecret(tc.typ, field) {
			t.Errorf("expected exact creation field %s.%s to be allowed", tc.typ, tc.field)
		}
	}
	for _, tc := range []struct {
		typ   reflect.Type
		field string
	}{{reflect.TypeOf(testtypes.V1CreateApiKeyResult{}), "Secret"}} {
		field, ok := tc.typ.FieldByName(tc.field)
		if !ok || !dummyFieldIsSecret(tc.typ, field) {
			t.Errorf("same-named type from another package was allowed: %s.%s", tc.typ, tc.field)
		}
	}
}

func checkDummyDTO(t *testing.T, typ reflect.Type) {
	t.Helper()
	for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		if dummyFieldIsSecret(typ, f) {
			t.Errorf("%s contains secret-like field %q", typ.Name(), f.Name)
		}
		checkDummyDTO(t, f.Type)
	}
}

func dummyFieldIsSecret(typ reflect.Type, field reflect.StructField) bool {
	name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
	if name == "" {
		name = field.Name
	}
	if typ == reflect.TypeOf(V1CreateApiKeyResult{}) && field.Name == "Secret" {
		return false
	}
	return dummySecretField.MatchString(normalizeSecretIdentifier(field.Name)) || dummySecretField.MatchString(normalizeSecretIdentifier(name))
}

func normalizeSecretIdentifier(value string) string {
	return strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(value))
}
