package cli

import (
	"fmt"
	"strings"
)

func allocationDetailsValue(attrs map[string]any) any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs["details"]
	if !ok || value == nil {
		return nil
	}
	return value
}

func allocationDetailsCount(details any) int {
	switch typed := details.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	case []map[string]string:
		return len(typed)
	default:
		return 0
	}
}

func allocationCostCodeIDsFromDetails(details any) []string {
	var ids []string
	switch typed := details.(type) {
	case []any:
		for _, item := range typed {
			switch value := item.(type) {
			case map[string]any:
				if id := normalizeCostCodeID(value["cost_code_id"]); id != "" {
					ids = append(ids, id)
				}
			case map[string]string:
				if id := strings.TrimSpace(value["cost_code_id"]); id != "" {
					ids = append(ids, id)
				}
			}
		}
	case []map[string]any:
		for _, value := range typed {
			if id := normalizeCostCodeID(value["cost_code_id"]); id != "" {
				ids = append(ids, id)
			}
		}
	case []map[string]string:
		for _, value := range typed {
			if id := strings.TrimSpace(value["cost_code_id"]); id != "" {
				ids = append(ids, id)
			}
		}
	}
	return uniqueCostCodeIDs(ids)
}

func normalizeCostCodeID(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", value))
}

func uniqueCostCodeIDs(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}
