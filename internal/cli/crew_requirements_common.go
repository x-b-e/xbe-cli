package cli

import (
	"fmt"
	"strings"
)

func parseCrewRequirementKind(value string) (string, string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, " ", "-")

	switch normalized {
	case "labor",
		"labor-requirement",
		"labor-requirements",
		"laborrequirement",
		"laborrequirements":
		return "labor-requirements", "/v1/labor-requirements", nil
	case "equipment",
		"equipment-requirement",
		"equipment-requirements",
		"equipmentrequirement",
		"equipmentrequirements":
		return "equipment-requirements", "/v1/equipment-requirements", nil
	default:
		return "", "", fmt.Errorf("invalid requirement type: %s (use labor or equipment)", value)
	}
}

func parseCrewRequirementBool(value string, flagName string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("--%s must be true or false", flagName)
	}
}
