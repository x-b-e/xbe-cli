package cli

import (
	"encoding/json"
	"fmt"
)

func parseCommitmentItemIntList(raw string, field string) ([]int, error) {
	var values []int
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, fmt.Errorf("invalid %s JSON array: %w", field, err)
	}
	return values, nil
}

func parseCommitmentItemStringList(raw string, field string) ([]string, error) {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, fmt.Errorf("invalid %s JSON array: %w", field, err)
	}
	return values, nil
}

func parseCommitmentItemJSONObject(raw string, field string) (map[string]any, error) {
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("invalid %s JSON object: %w", field, err)
	}
	return value, nil
}
