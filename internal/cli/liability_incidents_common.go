package cli

import "strings"

func normalizeIncidentSubjectRelationshipType(value string) string {
	normalized := strings.TrimSpace(value)
	switch strings.ToLower(normalized) {
	case "jobproductionplan", "job-production-plan", "job-production-plans":
		return "job-production-plans"
	case "tenderjobscheduleshift", "tender-job-schedule-shift", "tender-job-schedule-shifts":
		return "tender-job-schedule-shifts"
	case "broker", "brokers":
		return "brokers"
	case "customer", "customers":
		return "customers"
	case "trucker", "truckers":
		return "truckers"
	case "developer", "developers":
		return "developers"
	case "contractor", "contractors":
		return "contractors"
	case "materialsupplier", "material-supplier", "material-suppliers":
		return "material-suppliers"
	case "materialsite", "material-site", "material-sites":
		return "material-sites"
	case "projecttransportplan", "project-transport-plan", "project-transport-plans":
		return "project-transport-plans"
	default:
		return normalized
	}
}

func normalizeIncidentRelationshipType(value string) string {
	normalized := strings.TrimSpace(value)
	switch strings.ToLower(normalized) {
	case "incident", "incidents":
		return "incidents"
	case "liabilityincident", "liability-incident", "liability-incidents":
		return "liability-incidents"
	case "safetyincident", "safety-incident", "safety-incidents":
		return "safety-incidents"
	case "productionincident", "production-incident", "production-incidents":
		return "production-incidents"
	case "efficiencyincident", "efficiency-incident", "efficiency-incidents":
		return "efficiency-incidents"
	case "administrativeincident", "administrative-incident", "administrative-incidents":
		return "administrative-incidents"
	default:
		return normalized
	}
}
