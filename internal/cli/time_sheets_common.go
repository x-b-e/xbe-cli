package cli

import "strings"

func formatTimeSheetSubject(subjectType, subjectID string) string {
	if subjectType == "" && subjectID == "" {
		return ""
	}
	if subjectType == "" {
		return subjectID
	}
	if subjectID == "" {
		return subjectType
	}
	return subjectType + "/" + subjectID
}

func normalizeTimeSheetSubjectFilter(subjectType string) string {
	normalized := strings.TrimSpace(subjectType)
	switch strings.ToLower(normalized) {
	case "workorder", "work-order", "work-orders":
		return "WorkOrder"
	case "crewrequirement", "crew-requirement", "crew-requirements":
		return "CrewRequirement"
	case "truckershiftset", "trucker-shift-set", "trucker-shift-sets":
		return "TruckerShiftSet"
	default:
		return normalized
	}
}

func normalizeTimeSheetSubjectRelationship(subjectType string) string {
	normalized := strings.TrimSpace(subjectType)
	switch strings.ToLower(normalized) {
	case "workorder", "work-order", "work-orders":
		return "work-orders"
	case "crewrequirement", "crew-requirement", "crew-requirements":
		return "crew-requirements"
	case "truckershiftset", "trucker-shift-set", "trucker-shift-sets":
		return "trucker-shift-sets"
	default:
		return normalized
	}
}
