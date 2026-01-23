package cli

import (
	"fmt"
	"strings"
	"unicode"
)

func normalizeDriverAssignmentRuleLevelType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func parseDriverAssignmentRuleLevelType(value string) (string, error) {
	normalized := normalizeDriverAssignmentRuleLevelType(value)
	switch normalized {
	case "broker", "brokers":
		return "brokers", nil
	case "jobscheduleshift", "jobscheduleshifts":
		return "job-schedule-shifts", nil
	case "project", "projects":
		return "projects", nil
	case "jobproductionplan", "jobproductionplans":
		return "job-production-plans", nil
	case "materialsupplier", "materialsuppliers":
		return "material-suppliers", nil
	case "materialsite", "materialsites":
		return "material-sites", nil
	case "materialtype", "materialtypes":
		return "material-types", nil
	case "trucker", "truckers":
		return "truckers", nil
	case "jobsite", "jobsites":
		return "job-sites", nil
	default:
		return "", fmt.Errorf("invalid level type %q (expected Broker, JobScheduleShift, Project, JobProductionPlan, MaterialSupplier, MaterialSite, MaterialType, Trucker, or JobSite)", value)
	}
}

func normalizeDriverAssignmentRuleLevelFilter(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, r := range value {
		if unicode.IsUpper(r) {
			return value
		}
	}

	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == len(parts)-1 && strings.HasSuffix(part, "s") && len(part) > 1 {
			part = strings.TrimSuffix(part, "s")
		}
		part = strings.ToLower(part)
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}
