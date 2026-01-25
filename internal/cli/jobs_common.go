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
