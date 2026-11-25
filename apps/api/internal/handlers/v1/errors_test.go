package v1

import (
	"encoding/json"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
	"strings"
	"testing"
)

func TestProblemNeverEchoesInput(t *testing.T) {
	body, err := json.Marshal(Problem(http.StatusUnprocessableEntity, "metadata_invalid"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "hostile-input") || !strings.Contains(string(body), "metadata_invalid") {
		t.Fatalf("unsafe problem: %s", body)
	}
}

type detailError struct {
	location string
	value    any
}

func (e detailError) Error() string { return fmt.Sprint(e.value) }
func (e detailError) ErrorDetail() *huma.ErrorDetail {
	return &huma.ErrorDetail{Location: e.location, Message: "invalid value", Value: e.value}
}

func TestValidationErrorOmitsUntrustedFieldNamesAndValues(t *testing.T) {
	for _, field := range []string{"client_supplied_map_key", "client_supplied_extra_property"} {
		t.Run(field, func(t *testing.T) {
			body, err := json.Marshal(safeInputError(nil, http.StatusUnprocessableEntity, detailError{location: "body." + field, value: "hostile-input"}))
			if err != nil {
				t.Fatal(err)
			}
			for _, supplied := range []string{field, "hostile-input", "invalid value"} {
				if strings.Contains(string(body), supplied) {
					t.Fatalf("validation error leaked %q: %s", supplied, body)
				}
			}
		})
	}
}
