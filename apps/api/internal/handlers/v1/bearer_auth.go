package v1

import (
	"net/http"
	"strings"

	"github.com/vandordev/vkit-zenith/packages/fixtures"
)

// DUMMY (ADR-0001 phase 2): keys resolve from fixtures until API keys are persisted.
func bearerOrg(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", Problem(http.StatusUnauthorized, "unauthorized")
	}
	orgID, ok := fixtures.ResolveAPIKey(parts[1])
	if !ok {
		return "", Problem(http.StatusUnauthorized, "unauthorized")
	}
	return orgID, nil
}
