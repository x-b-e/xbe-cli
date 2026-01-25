package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type jsonAPIResponse struct {
	Data     []jsonAPIResource `json:"data"`
	Included []jsonAPIResource `json:"included"`
}

type jsonAPISingleResponse struct {
	Data     jsonAPIResource   `json:"data"`
	Included []jsonAPIResource `json:"included"`
}

type jsonAPIResource struct {
	ID            string                         `json:"id"`
	Type          string                         `json:"type"`
	Attributes    map[string]any                 `json:"attributes"`
	Relationships map[string]jsonAPIRelationship `json:"relationships"`
	Meta          map[string]any                 `json:"meta"`
}

type jsonAPIRelationship struct {
	Data *jsonAPIResourceIdentifier `json:"-"`
	raw  json.RawMessage
}

func (r *jsonAPIRelationship) UnmarshalJSON(data []byte) error {
	// Parse the wrapper object
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	r.raw = wrapper.Data

	// Try to parse as single object (has_one)
	if len(wrapper.Data) > 0 && wrapper.Data[0] == '{' {
		var single jsonAPIResourceIdentifier
		if err := json.Unmarshal(wrapper.Data, &single); err == nil {
			r.Data = &single
		}
	}
	// If it's an array (has_many) or null, Data stays nil
	return nil
}

type jsonAPIResourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func resolveOrganization(resource jsonAPIResource, included map[string]map[string]any) string {
	rel, ok := resource.Relationships["organization"]
	if !ok || rel.Data == nil {
		return "XBE Horizon"
	}

	key := resourceKey(rel.Data.Type, rel.Data.ID)
	if attrs, ok := included[key]; ok {
		name := firstNonEmpty(
			stringAttr(attrs, "company-name"),
			stringAttr(attrs, "name"),
			stringAttr(attrs, "title"),
		)
		if name != "" {
			return name
		}
	}

	return fmt.Sprintf("%s:%s", rel.Data.Type, rel.Data.ID)
}

func resourceKey(typ, id string) string {
	return typ + "|" + id
}

func stringAttr(attrs map[string]any, key string) string {
	if attrs == nil {
		return ""
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func stringSliceAttr(attrs map[string]any, key string) []string {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if item == nil {
				continue
			}
			values = append(values, fmt.Sprintf("%v", item))
		}
		return values
	default:
		return []string{fmt.Sprintf("%v", typed)}
	}
}

func boolAttr(attrs map[string]any, key string) bool {
	if attrs == nil {
		return false
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return fmt.Sprintf("%v", typed) == "true"
	}
}

func intAttr(attrs map[string]any, key string) int {
	if attrs == nil {
		return 0
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		var result int
		fmt.Sscanf(typed, "%d", &result)
		return result
	default:
		return 0
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func truncateString(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	if max < 4 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func formatDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", value); err == nil {
		return value
	}
	return value
}

func writeJSON(out io.Writer, value any) error {
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if _, err := out.Write(pretty); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out)
	return err
}

func defaultBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("XBE_BASE_URL")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("XBE_API_BASE_URL")); value != "" {
		return value
	}
	return "https://server.x-b-e.com"
}

// decodeHTMLEntities replaces common HTML entities with their character equivalents
func decodeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&#39;", "'")
	return s
}
