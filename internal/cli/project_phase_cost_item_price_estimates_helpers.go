package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseEstimateInput(raw string) (any, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("estimate JSON is required")
	}
	if value == "null" {
		return nil, nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return nil, fmt.Errorf("invalid estimate JSON: %w", err)
	}
	if parsed == nil {
		return nil, nil
	}
	if _, ok := parsed.(map[string]any); !ok {
		return nil, fmt.Errorf("estimate must be a JSON object")
	}
	return parsed, nil
}
