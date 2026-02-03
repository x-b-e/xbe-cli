package cli

import (
	"encoding/json"
	"testing"
)

func mustRelationship(t *testing.T, raw string) jsonAPIRelationship {
	t.Helper()
	var rel jsonAPIRelationship
	if err := json.Unmarshal([]byte(raw), &rel); err != nil {
		t.Fatalf("unmarshal relationship: %v", err)
	}
	return rel
}

func oneRelationship(id, typ string) jsonAPIRelationship {
	return jsonAPIRelationship{
		Data: &jsonAPIResourceIdentifier{
			ID:   id,
			Type: typ,
		},
	}
}

func TestBuildTimeCardRowsRelationships(t *testing.T) {
	resp := jsonAPIResponse{
		Data: []jsonAPIResource{
			{
				ID: "tc-1",
				Relationships: map[string]jsonAPIRelationship{
					"job":                       oneRelationship("job-1", "jobs"),
					"job-production-plan":       oneRelationship("jpp-1", "job-production-plans"),
					"driver":                    oneRelationship("driver-1", "users"),
					"trucker":                   oneRelationship("trucker-1", "truckers"),
					"tender-job-schedule-shift": oneRelationship("tjss-1", "tender-job-schedule-shifts"),
					"job-schedule-shift":        oneRelationship("jss-1", "job-schedule-shifts"),
				},
			},
		},
	}

	rows := buildTimeCardRows(resp)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row.JobID != "job-1" {
		t.Fatalf("expected job id job-1, got %q", row.JobID)
	}
	if row.JobProductionPlanID != "jpp-1" {
		t.Fatalf("expected job production plan id jpp-1, got %q", row.JobProductionPlanID)
	}
	if row.DriverID != "driver-1" {
		t.Fatalf("expected driver id driver-1, got %q", row.DriverID)
	}
	if row.TruckerID != "trucker-1" {
		t.Fatalf("expected trucker id trucker-1, got %q", row.TruckerID)
	}
	if row.TenderJobScheduleShift != "tjss-1" {
		t.Fatalf("expected tender job schedule shift id tjss-1, got %q", row.TenderJobScheduleShift)
	}
	if row.JobScheduleShiftID != "jss-1" {
		t.Fatalf("expected job schedule shift id jss-1, got %q", row.JobScheduleShiftID)
	}
}

func TestTimeCardRowFromSingleRelationships(t *testing.T) {
	resp := jsonAPISingleResponse{
		Data: jsonAPIResource{
			ID: "tc-2",
			Relationships: map[string]jsonAPIRelationship{
				"job":                 oneRelationship("job-2", "jobs"),
				"job-production-plan": oneRelationship("jpp-2", "job-production-plans"),
			},
		},
	}

	row := timeCardRowFromSingle(resp)
	if row.JobID != "job-2" {
		t.Fatalf("expected job id job-2, got %q", row.JobID)
	}
	if row.JobProductionPlanID != "jpp-2" {
		t.Fatalf("expected job production plan id jpp-2, got %q", row.JobProductionPlanID)
	}
}

func TestBuildTimeCardDetailsRelationships(t *testing.T) {
	resp := jsonAPISingleResponse{
		Data: jsonAPIResource{
			ID: "tc-3",
			Relationships: map[string]jsonAPIRelationship{
				"broker-tender":                  oneRelationship("bt-1", "broker-tenders"),
				"submitted-by":                   oneRelationship("user-1", "users"),
				"tender-job-schedule-shift":      oneRelationship("tjss-2", "tender-job-schedule-shifts"),
				"trucker":                        oneRelationship("trucker-2", "truckers"),
				"time-card-cost-code-allocation": oneRelationship("tcca-1", "time-card-cost-code-allocations"),
				"customer":                       oneRelationship("customer-1", "customers"),
				"broker":                         oneRelationship("broker-1", "brokers"),
				"driver":                         oneRelationship("driver-2", "users"),
				"trailer":                        oneRelationship("trailer-1", "trailers"),
				"tractor":                        oneRelationship("tractor-1", "tractors"),
				"job":                            oneRelationship("job-3", "jobs"),
				"job-site":                       oneRelationship("job-site-1", "job-sites"),
				"job-schedule-shift":             oneRelationship("jss-2", "job-schedule-shifts"),
				"job-production-plan":            oneRelationship("jpp-3", "job-production-plans"),
				"contractor":                     oneRelationship("contractor-1", "contractors"),
				"accepted-customer-tender-job-schedule-shift": oneRelationship("ctjss-1", "tender-job-schedule-shifts"),
				"time-card-payroll-certification":             oneRelationship("tcpc-1", "time-card-payroll-certifications"),
				"time-card-approval-audit":                    oneRelationship("tcaa-1", "time-card-approval-audits"),
				"time-card-status-changes":                    mustRelationship(t, `{"data":[{"type":"time-card-status-changes","id":"sc-1"},{"type":"time-card-status-changes","id":"sc-2"}]}`),
				"service-type-unit-of-measure-quantities":     mustRelationship(t, `{"data":[{"type":"service-type-unit-of-measure-quantities","id":"stuom-1"}]}`),
				"file-attachments":                            mustRelationship(t, `{"data":[{"type":"file-attachments","id":"fa-1"}]}`),
				"invoices":                                    mustRelationship(t, `{"data":[{"type":"invoices","id":"inv-1"},{"type":"invoices","id":"inv-2"}]}`),
				"job-production-plan-time-card-approvers":     mustRelationship(t, `{"data":[{"type":"job-production-plan-time-card-approvers","id":"jpp-tca-1"}]}`),
				"job-production-plan-material-types":          mustRelationship(t, `{"data":[{"type":"job-production-plan-material-types","id":"jpp-mt-1"}]}`),
			},
		},
	}

	details := buildTimeCardDetails(resp)
	if details.BrokerTenderID != "bt-1" {
		t.Fatalf("expected broker tender id bt-1, got %q", details.BrokerTenderID)
	}
	if details.SubmittedByID != "user-1" {
		t.Fatalf("expected submitted by id user-1, got %q", details.SubmittedByID)
	}
	if details.TenderJobScheduleShiftID != "tjss-2" {
		t.Fatalf("expected tender job schedule shift id tjss-2, got %q", details.TenderJobScheduleShiftID)
	}
	if details.TruckerID != "trucker-2" {
		t.Fatalf("expected trucker id trucker-2, got %q", details.TruckerID)
	}
	if details.TimeCardCostCodeAllocationID != "tcca-1" {
		t.Fatalf("expected cost code allocation id tcca-1, got %q", details.TimeCardCostCodeAllocationID)
	}
	if details.CustomerID != "customer-1" {
		t.Fatalf("expected customer id customer-1, got %q", details.CustomerID)
	}
	if details.BrokerID != "broker-1" {
		t.Fatalf("expected broker id broker-1, got %q", details.BrokerID)
	}
	if details.DriverID != "driver-2" {
		t.Fatalf("expected driver id driver-2, got %q", details.DriverID)
	}
	if details.TrailerID != "trailer-1" {
		t.Fatalf("expected trailer id trailer-1, got %q", details.TrailerID)
	}
	if details.TractorID != "tractor-1" {
		t.Fatalf("expected tractor id tractor-1, got %q", details.TractorID)
	}
	if details.JobID != "job-3" {
		t.Fatalf("expected job id job-3, got %q", details.JobID)
	}
	if details.JobSiteID != "job-site-1" {
		t.Fatalf("expected job site id job-site-1, got %q", details.JobSiteID)
	}
	if details.JobScheduleShiftID != "jss-2" {
		t.Fatalf("expected job schedule shift id jss-2, got %q", details.JobScheduleShiftID)
	}
	if details.JobProductionPlanID != "jpp-3" {
		t.Fatalf("expected job production plan id jpp-3, got %q", details.JobProductionPlanID)
	}
	if details.ContractorID != "contractor-1" {
		t.Fatalf("expected contractor id contractor-1, got %q", details.ContractorID)
	}
	if details.AcceptedCustomerTenderJobScheduleShift != "ctjss-1" {
		t.Fatalf("expected accepted customer tender job schedule shift id ctjss-1, got %q", details.AcceptedCustomerTenderJobScheduleShift)
	}
	if details.TimeCardPayrollCertificationID != "tcpc-1" {
		t.Fatalf("expected payroll certification id tcpc-1, got %q", details.TimeCardPayrollCertificationID)
	}
	if details.TimeCardApprovalAuditID != "tcaa-1" {
		t.Fatalf("expected approval audit id tcaa-1, got %q", details.TimeCardApprovalAuditID)
	}
	if len(details.TimeCardStatusChangeIDs) != 2 || details.TimeCardStatusChangeIDs[0] != "sc-1" {
		t.Fatalf("unexpected status change ids: %v", details.TimeCardStatusChangeIDs)
	}
	if len(details.ServiceTypeUnitOfMeasureQuantityIDs) != 1 || details.ServiceTypeUnitOfMeasureQuantityIDs[0] != "stuom-1" {
		t.Fatalf("unexpected stuom quantity ids: %v", details.ServiceTypeUnitOfMeasureQuantityIDs)
	}
	if len(details.FileAttachmentIDs) != 1 || details.FileAttachmentIDs[0] != "fa-1" {
		t.Fatalf("unexpected file attachment ids: %v", details.FileAttachmentIDs)
	}
	if len(details.InvoiceIDs) != 2 || details.InvoiceIDs[0] != "invoices:inv-1" {
		t.Fatalf("unexpected invoice ids: %v", details.InvoiceIDs)
	}
	if len(details.JobProductionPlanTimeCardApproverIDs) != 1 || details.JobProductionPlanTimeCardApproverIDs[0] != "jpp-tca-1" {
		t.Fatalf("unexpected jpp time card approver ids: %v", details.JobProductionPlanTimeCardApproverIDs)
	}
	if len(details.JobProductionPlanMaterialTypeIDs) != 1 || details.JobProductionPlanMaterialTypeIDs[0] != "jpp-mt-1" {
		t.Fatalf("unexpected jpp material type ids: %v", details.JobProductionPlanMaterialTypeIDs)
	}
}
