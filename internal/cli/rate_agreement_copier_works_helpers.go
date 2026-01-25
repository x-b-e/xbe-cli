package cli

import "strings"

func normalizeTargetOrganizationTypeForFilter(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	switch strings.ToLower(value) {
	case "customer", "customers":
		return "Customer"
	case "trucker", "truckers":
		return "Trucker"
	default:
		return value
	}
}

func normalizeTargetOrganizationTypeForRelationship(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	switch strings.ToLower(value) {
	case "customer", "customers":
		return "customers"
	case "trucker", "truckers":
		return "truckers"
	default:
		return value
	}
}
