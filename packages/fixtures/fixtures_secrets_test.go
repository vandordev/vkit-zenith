package fixtures

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var secretField = regexp.MustCompile(`(?i)(hash|ciphertext|secret|token|password|objectkey|^pan$|cardnumber|cvv)`)

func TestFixturesContainNoSecretFields(t *testing.T) {
	for _, root := range []any{OrgData{}, APIKey{}} {
		checkFixtureType(t, reflect.TypeOf(root))
	}
}

func TestFixtureSecretScannerParsesJSONTagOptions(t *testing.T) {
	field, _ := reflect.TypeOf(struct {
		PAN string `json:"pan,omitempty"`
	}{}).FieldByName("PAN")
	name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
	if !secretField.MatchString(name) {
		t.Fatal("pan,omitempty was not recognized as a secret key")
	}
}

func TestFixtureSecretScannerRejectsSecretGoNamesWithSafeJSONAliases(t *testing.T) {
	for _, alias := range []string{"prefix", "tokenPrefix", "hasCredentials"} {
		field, _ := reflect.TypeOf(struct {
			Secret string `json:"prefix"`
		}{}).FieldByName("Secret")
		field.Tag = reflect.StructTag(`json:"` + alias + `"`)
		if !fixtureFieldIsSecret(reflect.TypeOf(struct{}{}), field) {
			t.Errorf("secret Go field was allowed by JSON alias %q", alias)
		}
	}
}

func TestFixtureSecretScannerFindsSnakeCaseJSONNames(t *testing.T) {
	typ := reflect.TypeOf(struct {
		LogoKey string `json:"logo_object_key"`
		Card    string `json:"card_number"`
	}{})
	for _, name := range []string{"LogoKey", "Card"} {
		field, _ := typ.FieldByName(name)
		if !fixtureFieldIsSecret(typ, field) {
			t.Errorf("snake-case secret JSON name on %s was not detected", name)
		}
	}
}

func checkFixtureType(t *testing.T, typ reflect.Type) {
	t.Helper()
	for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		if fixtureFieldIsSecret(typ, field) {
			t.Errorf("%s contains secret-like field %q", typ.Name(), field.Name)
		}
		checkFixtureType(t, field.Type)
	}
}

func fixtureFieldIsSecret(typ reflect.Type, field reflect.StructField) bool {
	name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
	if name == "" {
		name = field.Name
	}
	return secretField.MatchString(normalizeSecretIdentifier(field.Name)) || secretField.MatchString(normalizeSecretIdentifier(name))
}

func normalizeSecretIdentifier(value string) string {
	return strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(value))
}
