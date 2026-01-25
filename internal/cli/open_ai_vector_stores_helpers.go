package cli

import (
	"fmt"
	"strings"
)

func parseOpenAiVectorStoreScope(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", nil
	}
	parts := strings.SplitN(value, "|", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("--scope must be in Type|ID format")
	}
	resourceType := strings.TrimSpace(parts[0])
	resourceID := strings.TrimSpace(parts[1])
	if resourceType == "" || resourceID == "" {
		return "", "", fmt.Errorf("--scope must include type and id")
	}
	return resourceType, resourceID, nil
}

func normalizeOpenAiVectorStoreScopeTypeForRelationship(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	normalized := strings.ToLower(strings.ReplaceAll(value, "_", "-"))
	switch normalized {
	case "broker", "brokers":
		return "brokers"
	case "userpostfeed", "user-post-feed", "user-post-feeds", "user-postfeeds":
		return "user-post-feeds"
	default:
		return normalized
	}
}
