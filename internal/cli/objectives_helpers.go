package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func resolveObjectiveOrganization(cmd *cobra.Command, organization, orgType, orgID string) (string, string, error) {
	if cmd.Flags().Changed("organization") {
		return parseOrganization(organization)
	}
	if cmd.Flags().Changed("organization-type") || cmd.Flags().Changed("organization-id") {
		if strings.TrimSpace(orgType) == "" || strings.TrimSpace(orgID) == "" {
			return "", "", fmt.Errorf("--organization-type and --organization-id must be provided together")
		}
		return parseOrganization(fmt.Sprintf("%s|%s", orgType, orgID))
	}
	return "", "", nil
}

func resolveObjectiveParent(cmd *cobra.Command, parent, parentType, parentID string) (string, string, error) {
	if cmd.Flags().Changed("parent") {
		return parseObjectiveParent(parent)
	}
	if cmd.Flags().Changed("parent-type") || cmd.Flags().Changed("parent-id") {
		if strings.TrimSpace(parentType) == "" || strings.TrimSpace(parentID) == "" {
			return "", "", fmt.Errorf("--parent-type and --parent-id must be provided together")
		}
		return parseObjectiveParent(fmt.Sprintf("%s|%s", parentType, parentID))
	}
	return "", "", nil
}

func parseObjectiveParent(value string) (string, string, error) {
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid parent format: %q (expected Objective|ID or KeyResult|ID)", value)
	}
	parentType, err := normalizeObjectiveParentType(parts[0])
	if err != nil {
		return "", "", err
	}
	parentID := strings.TrimSpace(parts[1])
	if parentID == "" {
		return "", "", fmt.Errorf("invalid parent format: %q (expected Objective|ID or KeyResult|ID)", value)
	}
	return parentType, parentID, nil
}

func normalizeObjectiveParentType(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, " ", "-")
	switch normalized {
	case "objective", "objectives":
		return "objectives", nil
	case "key-result", "key-results", "keyresult", "keyresults":
		return "key-results", nil
	default:
		return "", fmt.Errorf("invalid parent type: %q (expected Objective or KeyResult)", value)
	}
}
