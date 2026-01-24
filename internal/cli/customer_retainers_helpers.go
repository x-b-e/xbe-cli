package cli

import "strings"

func normalizeIDList(ids []string) []string {
	cleaned := make([]string, 0, len(ids))
	for _, raw := range ids {
		for _, part := range strings.Split(raw, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func buildRelationshipData(resourceType string, ids []string) []map[string]any {
	cleaned := normalizeIDList(ids)
	data := make([]map[string]any, 0, len(cleaned))
	for _, id := range cleaned {
		data = append(data, map[string]any{
			"type": resourceType,
			"id":   id,
		})
	}
	return data
}
