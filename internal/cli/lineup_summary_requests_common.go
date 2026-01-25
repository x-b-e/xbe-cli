package cli

import (
	"fmt"
	"strings"
)

func normalizeLineupSummaryRequestLevelType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func parseLineupSummaryRequestLevelType(value string) (string, error) {
	normalized := normalizeLineupSummaryRequestLevelType(value)
	switch normalized {
	case "broker", "brokers":
		return "brokers", nil
	case "customer", "customers":
		return "customers", nil
	default:
		return "", fmt.Errorf("invalid level type %q (expected Broker or Customer)", value)
	}
}

func normalizeLineupSummaryRequestEmails(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}
