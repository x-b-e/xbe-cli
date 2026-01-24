package cli

import (
	"fmt"
	"sort"
	"strings"
)

func relationshipRefFromMap(relationships map[string]jsonAPIRelationship, key string) (string, string) {
	if relationships == nil {
		return "", ""
	}
	rel, ok := relationships[key]
	if !ok || rel.Data == nil {
		return "", ""
	}
	return rel.Data.ID, rel.Data.Type
}

func formatRelationshipLabel(typ, id string) string {
	typ = strings.TrimSpace(typ)
	id = strings.TrimSpace(id)
	if typ != "" && id != "" {
		return fmt.Sprintf("%s:%s", typ, id)
	}
	if id != "" {
		return id
	}
	return typ
}

func stringMapAttr(attrs map[string]any, key string) map[string]string {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case map[string]string:
		if len(typed) == 0 {
			return nil
		}
		result := make(map[string]string, len(typed))
		for k, v := range typed {
			result[k] = v
		}
		return result
	case map[string]any:
		if len(typed) == 0 {
			return nil
		}
		result := make(map[string]string, len(typed))
		for k, v := range typed {
			if v == nil {
				continue
			}
			result[k] = fmt.Sprintf("%v", v)
		}
		if len(result) == 0 {
			return nil
		}
		return result
	default:
		return nil
	}
}

func formatStringMap(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, values[key]))
	}
	return strings.Join(parts, ", ")
}
