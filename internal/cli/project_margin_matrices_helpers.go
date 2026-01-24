package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

func projectMarginMatrixScenarioCount(value any) int {
	if value == nil {
		return 0
	}
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 0
	}
}

func formatProjectMarginMatrixJSONBlock(value any, indent string) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	if indent == "" {
		return string(pretty)
	}
	lines := strings.Split(string(pretty), "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
