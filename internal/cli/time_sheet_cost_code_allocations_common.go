package cli

import (
	"encoding/json"
	"fmt"
	"strings"
)

type timeSheetCostCodeAllocationDetail struct {
	CostCodeID                  string `json:"cost_code_id,omitempty"`
	Percentage                  string `json:"percentage,omitempty"`
	ProjectCostClassificationID string `json:"project_cost_classification_id,omitempty"`
}

func parseTimeSheetCostCodeAllocationDetails(attrs map[string]any) []timeSheetCostCodeAllocationDetail {
	if attrs == nil {
		return nil
	}

	raw, ok := attrs["details"]
	if !ok || raw == nil {
		return nil
	}

	switch typed := raw.(type) {
	case []any:
		return parseTimeSheetCostCodeAllocationDetailsSlice(typed)
	case []map[string]any:
		details := make([]timeSheetCostCodeAllocationDetail, 0, len(typed))
		for _, detail := range typed {
			details = append(details, parseTimeSheetCostCodeAllocationDetailMap(detail))
		}
		return details
	case string:
		return parseTimeSheetCostCodeAllocationDetailsJSON(typed)
	default:
		return nil
	}
}

func parseTimeSheetCostCodeAllocationDetailsSlice(items []any) []timeSheetCostCodeAllocationDetail {
	details := make([]timeSheetCostCodeAllocationDetail, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if detailMap, ok := item.(map[string]any); ok {
			details = append(details, parseTimeSheetCostCodeAllocationDetailMap(detailMap))
			continue
		}
		details = append(details, timeSheetCostCodeAllocationDetail{Percentage: fmt.Sprintf("%v", item)})
	}
	return details
}

func parseTimeSheetCostCodeAllocationDetailsJSON(raw string) []timeSheetCostCodeAllocationDetail {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var decoded []map[string]any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil
	}
	details := make([]timeSheetCostCodeAllocationDetail, 0, len(decoded))
	for _, detail := range decoded {
		details = append(details, parseTimeSheetCostCodeAllocationDetailMap(detail))
	}
	return details
}

func parseTimeSheetCostCodeAllocationDetailMap(detail map[string]any) timeSheetCostCodeAllocationDetail {
	return timeSheetCostCodeAllocationDetail{
		CostCodeID:                  detailValue(detail, "cost_code_id", "cost-code-id"),
		Percentage:                  detailValue(detail, "percentage"),
		ProjectCostClassificationID: detailValue(detail, "project_cost_classification_id", "project-cost-classification-id"),
	}
}

func detailValue(detail map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := detail[key]; ok && value != nil {
			return fmt.Sprintf("%v", value)
		}
	}
	return ""
}

func formatTimeSheetCostCodeAllocationSummary(details []timeSheetCostCodeAllocationDetail) string {
	if len(details) == 0 {
		return ""
	}
	parts := make([]string, 0, len(details))
	for _, detail := range details {
		label := detail.CostCodeID
		if label == "" {
			label = "(unknown)"
		}
		if detail.Percentage != "" {
			label = fmt.Sprintf("%s:%s", label, detail.Percentage)
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, ", ")
}
