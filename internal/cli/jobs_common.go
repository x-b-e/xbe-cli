package cli

import "strings"

func parseCommaSeparatedIDs(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id == "" {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func buildRelationshipDataList(ids []string, typ string) []map[string]any {
	if len(ids) == 0 {
		return nil
	}
	data := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			continue
		}
		data = append(data, map[string]any{
			"type": typ,
			"id":   strings.TrimSpace(id),
		})
	}
	return data
}

func jobProductionPlanLabel(attrs map[string]any) string {
	jobNumber := strings.TrimSpace(stringAttr(attrs, "job-number"))
	jobName := strings.TrimSpace(stringAttr(attrs, "job-name"))
	if jobNumber != "" && jobName != "" {
		return jobNumber + " - " + jobName
	}
	if jobNumber != "" {
		return jobNumber
	}
	return jobName
}
