package cli

import (
	"fmt"
	"strings"
)

type incidentSubjectTypeInfo struct {
	ClassName   string
	JSONAPIType string
}

var incidentSubjectTypeLookup = map[string]incidentSubjectTypeInfo{
	"broker":                  {ClassName: "Broker", JSONAPIType: "brokers"},
	"brokers":                 {ClassName: "Broker", JSONAPIType: "brokers"},
	"customer":                {ClassName: "Customer", JSONAPIType: "customers"},
	"customers":               {ClassName: "Customer", JSONAPIType: "customers"},
	"trucker":                 {ClassName: "Trucker", JSONAPIType: "truckers"},
	"truckers":                {ClassName: "Trucker", JSONAPIType: "truckers"},
	"developer":               {ClassName: "Developer", JSONAPIType: "developers"},
	"developers":              {ClassName: "Developer", JSONAPIType: "developers"},
	"contractor":              {ClassName: "Contractor", JSONAPIType: "contractors"},
	"contractors":             {ClassName: "Contractor", JSONAPIType: "contractors"},
	"materialsupplier":        {ClassName: "MaterialSupplier", JSONAPIType: "material-suppliers"},
	"materialsuppliers":       {ClassName: "MaterialSupplier", JSONAPIType: "material-suppliers"},
	"materialsite":            {ClassName: "MaterialSite", JSONAPIType: "material-sites"},
	"materialsites":           {ClassName: "MaterialSite", JSONAPIType: "material-sites"},
	"jobproductionplan":       {ClassName: "JobProductionPlan", JSONAPIType: "job-production-plans"},
	"jobproductionplans":      {ClassName: "JobProductionPlan", JSONAPIType: "job-production-plans"},
	"tenderjobscheduleshift":  {ClassName: "TenderJobScheduleShift", JSONAPIType: "tender-job-schedule-shifts"},
	"tenderjobscheduleshifts": {ClassName: "TenderJobScheduleShift", JSONAPIType: "tender-job-schedule-shifts"},
	"projecttransportplan":    {ClassName: "ProjectTransportPlan", JSONAPIType: "project-transport-plans"},
	"projecttransportplans":   {ClassName: "ProjectTransportPlan", JSONAPIType: "project-transport-plans"},
}

func normalizeIncidentSubjectType(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	lower = strings.ReplaceAll(lower, "_", "")
	lower = strings.ReplaceAll(lower, "-", "")
	lower = strings.ReplaceAll(lower, " ", "")
	return lower
}

func parseIncidentSubjectRef(raw string) (string, string, string, error) {
	parts := strings.SplitN(raw, "|", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid subject format: %q (expected Type|ID, e.g. Broker|123)", raw)
	}
	rawType := strings.TrimSpace(parts[0])
	id := strings.TrimSpace(parts[1])
	if rawType == "" || id == "" {
		return "", "", "", fmt.Errorf("invalid subject format: %q (expected Type|ID, e.g. Broker|123)", raw)
	}

	info, ok := incidentSubjectTypeLookup[normalizeIncidentSubjectType(rawType)]
	if !ok {
		return "", "", "", fmt.Errorf("invalid subject type: %q (allowed: Broker, Customer, Trucker, Developer, Contractor, MaterialSupplier, MaterialSite, JobProductionPlan, TenderJobScheduleShift, ProjectTransportPlan)", rawType)
	}

	return info.ClassName, info.JSONAPIType, id, nil
}

func buildIncidentSubjectFilter(raw string) (string, error) {
	className, _, id, err := parseIncidentSubjectRef(raw)
	if err != nil {
		return "", err
	}
	return className + "|" + id, nil
}
