package cli

import (
	"fmt"
	"strings"
)

func parsePublicOrganizationScope(scope, scopeType, scopeID string) (string, string, error) {
	scope = strings.TrimSpace(scope)
	scopeType = strings.TrimSpace(scopeType)
	scopeID = strings.TrimSpace(scopeID)

	if scope != "" {
		if scopeType != "" || scopeID != "" {
			return "", "", fmt.Errorf("--public-organization-scope cannot be combined with --public-organization-scope-type or --public-organization-scope-id")
		}
		parts := strings.SplitN(scope, "|", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return "", "", fmt.Errorf("--public-organization-scope must be in the format type|id")
		}
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
	}

	if scopeType == "" && scopeID == "" {
		return "", "", nil
	}
	if scopeType == "" || scopeID == "" {
		return "", "", fmt.Errorf("--public-organization-scope-type and --public-organization-scope-id are required together")
	}

	return scopeType, scopeID, nil
}
