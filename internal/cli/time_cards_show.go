package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type timeCardsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeCardDetails struct {
	ID                                      string   `json:"id"`
	Status                                  string   `json:"status,omitempty"`
	TicketNumber                            string   `json:"ticket_number,omitempty"`
	StartAt                                 string   `json:"start_at,omitempty"`
	EndAt                                   string   `json:"end_at,omitempty"`
	DownMinutes                             string   `json:"down_minutes,omitempty"`
	CreditedMinutes                         string   `json:"credited_minutes,omitempty"`
	CreditedHours                           string   `json:"credited_hours,omitempty"`
	SkipAutoSubmissionUponMaterialTxnAccept bool     `json:"skip_auto_submission_upon_material_transaction_acceptance"`
	MaximumTravelMinutes                    string   `json:"maximum_travel_minutes,omitempty"`
	SubmittedTravelMinutes                  string   `json:"submitted_travel_minutes,omitempty"`
	SubmittedCustomerTravelMinutes          string   `json:"submitted_customer_travel_minutes,omitempty"`
	CustomerTenderMaximumTravelMinutes      string   `json:"customer_tender_maximum_travel_minutes,omitempty"`
	CustomerTravelMinutes                   string   `json:"customer_travel_minutes,omitempty"`
	TimeZoneID                              string   `json:"time_zone_id,omitempty"`
	TotalHours                              string   `json:"total_hours,omitempty"`
	TotalCustomerHours                      string   `json:"total_customer_hours,omitempty"`
	DurationMinutes                         string   `json:"duration_minutes,omitempty"`
	TravelBeforeMinutes                     string   `json:"travel_before_minutes,omitempty"`
	TravelDuringMinutes                     string   `json:"travel_during_minutes,omitempty"`
	TravelAfterMinutes                      string   `json:"travel_after_minutes,omitempty"`
	ApprovalProcess                         string   `json:"approval_process,omitempty"`
	ApprovalCount                           string   `json:"approval_count,omitempty"`
	SubmitAt                                string   `json:"submit_at,omitempty"`
	IsCustomerInvoiceableWithoutOverride    bool     `json:"is_customer_invoiceable_without_override"`
	CustomerBillingOffsetDayCount           string   `json:"customer_billing_offset_day_count,omitempty"`
	CustomerBillingPeriodMax                string   `json:"customer_billing_period_max,omitempty"`
	CustomerBillingPeriodMin                string   `json:"customer_billing_period_min,omitempty"`
	CustomerInvoiceDateWithoutOverride      string   `json:"customer_invoice_date_without_override,omitempty"`
	IsTruckerInvoiceableWithoutOverride     bool     `json:"is_trucker_invoiceable_without_override"`
	TruckerBillingOffsetDayCount            string   `json:"trucker_billing_offset_day_count,omitempty"`
	TruckerBillingPeriodMax                 string   `json:"trucker_billing_period_max,omitempty"`
	TruckerBillingPeriodMin                 string   `json:"trucker_billing_period_min,omitempty"`
	TruckerInvoiceDateWithoutOverride       string   `json:"trucker_invoice_date_without_override,omitempty"`
	BrokerAmount                            string   `json:"broker_amount,omitempty"`
	CostAmount                              string   `json:"cost_amount,omitempty"`
	CustomerAmount                          string   `json:"customer_amount,omitempty"`
	RevenueAmount                           string   `json:"revenue_amount,omitempty"`
	SquanderedHours                         string   `json:"squandered_hours,omitempty"`
	CustomerSquanderedAmount                string   `json:"customer_squandered_amount,omitempty"`
	CustomerBaseSquanderedAmount            string   `json:"customer_base_squandered_amount,omitempty"`
	BrokerSquanderedAmount                  string   `json:"broker_squandered_amount,omitempty"`
	BrokerBaseSquanderedAmount              string   `json:"broker_base_squandered_amount,omitempty"`
	JobSiteHours                            string   `json:"job_site_hours,omitempty"`
	CanDelete                               bool     `json:"can_delete"`
	ExplicitIsTrailerRequiredForApproval    bool     `json:"explicit_is_trailer_required_for_approval"`
	IsTrailerRequiredForApproval            bool     `json:"is_trailer_required_for_approval"`
	ExplicitEnforceTicketNumberUniqueness   bool     `json:"explicit_enforce_ticket_number_uniqueness"`
	EnforceTicketNumberUniqueness           bool     `json:"enforce_ticket_number_uniqueness"`
	IsTimeCardCreatingTimeSheetLineItem     bool     `json:"is_time_card_creating_time_sheet_line_item_explicit"`
	ExplicitIsInvoiceableWhenApproved       bool     `json:"explicit_is_invoiceable_when_approved"`
	CurrentUserCanApprove                   bool     `json:"current_user_can_approve"`
	IsManagementService                     bool     `json:"is_management_service"`
	IsTimeCardStartAtEvidenceRequired       bool     `json:"is_time_card_start_at_evidence_required"`
	GenerateBrokerInvoice                   bool     `json:"generate_broker_invoice"`
	GenerateTruckerInvoice                  bool     `json:"generate_trucker_invoice"`
	BrokerTenderID                          string   `json:"broker_tender_id,omitempty"`
	SubmittedByID                           string   `json:"submitted_by_id,omitempty"`
	TenderJobScheduleShiftID                string   `json:"tender_job_schedule_shift_id,omitempty"`
	TruckerID                               string   `json:"trucker_id,omitempty"`
	TimeCardCostCodeAllocationID            string   `json:"time_card_cost_code_allocation_id,omitempty"`
	CustomerID                              string   `json:"customer_id,omitempty"`
	BrokerID                                string   `json:"broker_id,omitempty"`
	DriverID                                string   `json:"driver_id,omitempty"`
	TrailerID                               string   `json:"trailer_id,omitempty"`
	TractorID                               string   `json:"tractor_id,omitempty"`
	JobID                                   string   `json:"job_id,omitempty"`
	JobSiteID                               string   `json:"job_site_id,omitempty"`
	JobScheduleShiftID                      string   `json:"job_schedule_shift_id,omitempty"`
	JobProductionPlanID                     string   `json:"job_production_plan_id,omitempty"`
	ContractorID                            string   `json:"contractor_id,omitempty"`
	AcceptedCustomerTenderJobScheduleShift  string   `json:"accepted_customer_tender_job_schedule_shift_id,omitempty"`
	TimeCardPayrollCertificationID          string   `json:"time_card_payroll_certification_id,omitempty"`
	TimeCardApprovalAuditID                 string   `json:"time_card_approval_audit_id,omitempty"`
	TimeCardStatusChangeIDs                 []string `json:"time_card_status_change_ids,omitempty"`
	ServiceTypeUnitOfMeasureQuantityIDs     []string `json:"service_type_unit_of_measure_quantity_ids,omitempty"`
	FileAttachmentIDs                       []string `json:"file_attachment_ids,omitempty"`
	InvoiceIDs                              []string `json:"invoice_ids,omitempty"`
	JobProductionPlanTimeCardApproverIDs    []string `json:"job_production_plan_time_card_approver_ids,omitempty"`
	JobProductionPlanMaterialTypeIDs        []string `json:"job_production_plan_material_type_ids,omitempty"`
}

func newTimeCardsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card details",
		Long: `Show the full details of a time card.

Output Fields:
  ID / Status / Ticket Number
  Start / End / Down / Duration / Total Hours
  Travel minutes (submitted + breakdown)
  Approval process / Approval count / Submit at
  Invoicing flags and billing windows
  Financial amounts and squandered amounts
  Key flags (trailer required, ticket uniqueness, invoiceable)
  Relationships (tender, shifts, trucker, broker, customer, attachments, invoices)

Arguments:
  <id>    The time card ID (required). Use the list command to find IDs.`,
		Example: `  # Show a time card
  xbe view time-cards show 123

  # Show as JSON
  xbe view time-cards show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardsShow,
	}
	initTimeCardsShowFlags(cmd)
	return cmd
}

func init() {
	timeCardsCmd.AddCommand(newTimeCardsShowCmd())
}

func initTimeCardsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeCardsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("time card id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[time-cards]", strings.Join([]string{
		"status",
		"ticket-number",
		"start-at",
		"end-at",
		"down-minutes",
		"credited-minutes",
		"credited-hours",
		"skip-auto-submission-upon-material-transaction-acceptance",
		"maximum-travel-minutes",
		"submitted-travel-minutes",
		"explicit-is-trailer-required-for-approval",
		"is-trailer-required-for-approval",
		"explicit-enforce-ticket-number-uniqueness",
		"is-time-card-creating-time-sheet-line-item-explicit",
		"explicit-is-invoiceable-when-approved",
		"submitted-customer-travel-minutes",
		"customer-tender-maximum-travel-minutes",
		"customer-travel-minutes",
		"time-zone-id",
		"total-hours",
		"total-customer-hours",
		"duration-minutes",
		"travel-before-minutes",
		"travel-during-minutes",
		"travel-after-minutes",
		"approval-process",
		"approval-count",
		"submit-at",
		"is-customer-invoiceable-without-override",
		"customer-billing-offset-day-count",
		"customer-billing-period-max",
		"customer-billing-period-min",
		"customer-invoice-date-without-override",
		"is-trucker-invoiceable-without-override",
		"trucker-billing-offset-day-count",
		"trucker-billing-period-max",
		"trucker-billing-period-min",
		"trucker-invoice-date-without-override",
		"broker-amount",
		"cost-amount",
		"customer-amount",
		"revenue-amount",
		"squandered-hours",
		"customer-squandered-amount",
		"customer-base-squandered-amount",
		"broker-squandered-amount",
		"broker-base-squandered-amount",
		"job-site-hours",
		"can-delete",
		"enforce-ticket-number-uniqueness",
		"current-user-can-approve",
		"is-management-service",
		"is-time-card-start-at-evidence-required",
		"generate-broker-invoice",
		"generate-trucker-invoice",
		"broker-tender",
		"submitted-by",
		"tender-job-schedule-shift",
		"trucker",
		"time-card-cost-code-allocation",
		"customer",
		"broker",
		"driver",
		"trailer",
		"tractor",
		"job",
		"job-site",
		"job-schedule-shift",
		"job-production-plan",
		"contractor",
		"accepted-customer-tender-job-schedule-shift",
		"time-card-status-changes",
		"service-type-unit-of-measure-quantities",
		"time-card-payroll-certification",
		"file-attachments",
		"invoices",
		"job-production-plan-time-card-approvers",
		"time-card-approval-audit",
		"job-production-plan-material-types",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/time-cards/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	details := buildTimeCardDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardDetails(cmd, details)
}

func parseTimeCardsShowOptions(cmd *cobra.Command) (timeCardsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeCardDetails(resp jsonAPISingleResponse) timeCardDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := timeCardDetails{
		ID:                                      resource.ID,
		Status:                                  stringAttr(attrs, "status"),
		TicketNumber:                            stringAttr(attrs, "ticket-number"),
		StartAt:                                 formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                                   formatDateTime(stringAttr(attrs, "end-at")),
		DownMinutes:                             numberAttrAsString(attrs, "down-minutes"),
		CreditedMinutes:                         numberAttrAsString(attrs, "credited-minutes"),
		CreditedHours:                           numberAttrAsString(attrs, "credited-hours"),
		SkipAutoSubmissionUponMaterialTxnAccept: boolAttr(attrs, "skip-auto-submission-upon-material-transaction-acceptance"),
		MaximumTravelMinutes:                    numberAttrAsString(attrs, "maximum-travel-minutes"),
		SubmittedTravelMinutes:                  numberAttrAsString(attrs, "submitted-travel-minutes"),
		SubmittedCustomerTravelMinutes:          numberAttrAsString(attrs, "submitted-customer-travel-minutes"),
		CustomerTenderMaximumTravelMinutes:      numberAttrAsString(attrs, "customer-tender-maximum-travel-minutes"),
		CustomerTravelMinutes:                   numberAttrAsString(attrs, "customer-travel-minutes"),
		TimeZoneID:                              stringAttr(attrs, "time-zone-id"),
		TotalHours:                              numberAttrAsString(attrs, "total-hours"),
		TotalCustomerHours:                      numberAttrAsString(attrs, "total-customer-hours"),
		DurationMinutes:                         numberAttrAsString(attrs, "duration-minutes"),
		TravelBeforeMinutes:                     numberAttrAsString(attrs, "travel-before-minutes"),
		TravelDuringMinutes:                     numberAttrAsString(attrs, "travel-during-minutes"),
		TravelAfterMinutes:                      numberAttrAsString(attrs, "travel-after-minutes"),
		ApprovalProcess:                         stringAttr(attrs, "approval-process"),
		ApprovalCount:                           numberAttrAsString(attrs, "approval-count"),
		SubmitAt:                                formatDateTime(stringAttr(attrs, "submit-at")),
		IsCustomerInvoiceableWithoutOverride:    boolAttr(attrs, "is-customer-invoiceable-without-override"),
		CustomerBillingOffsetDayCount:           numberAttrAsString(attrs, "customer-billing-offset-day-count"),
		CustomerBillingPeriodMax:                numberAttrAsString(attrs, "customer-billing-period-max"),
		CustomerBillingPeriodMin:                numberAttrAsString(attrs, "customer-billing-period-min"),
		CustomerInvoiceDateWithoutOverride:      stringAttr(attrs, "customer-invoice-date-without-override"),
		IsTruckerInvoiceableWithoutOverride:     boolAttr(attrs, "is-trucker-invoiceable-without-override"),
		TruckerBillingOffsetDayCount:            numberAttrAsString(attrs, "trucker-billing-offset-day-count"),
		TruckerBillingPeriodMax:                 numberAttrAsString(attrs, "trucker-billing-period-max"),
		TruckerBillingPeriodMin:                 numberAttrAsString(attrs, "trucker-billing-period-min"),
		TruckerInvoiceDateWithoutOverride:       stringAttr(attrs, "trucker-invoice-date-without-override"),
		BrokerAmount:                            numberAttrAsString(attrs, "broker-amount"),
		CostAmount:                              numberAttrAsString(attrs, "cost-amount"),
		CustomerAmount:                          numberAttrAsString(attrs, "customer-amount"),
		RevenueAmount:                           numberAttrAsString(attrs, "revenue-amount"),
		SquanderedHours:                         numberAttrAsString(attrs, "squandered-hours"),
		CustomerSquanderedAmount:                numberAttrAsString(attrs, "customer-squandered-amount"),
		CustomerBaseSquanderedAmount:            numberAttrAsString(attrs, "customer-base-squandered-amount"),
		BrokerSquanderedAmount:                  numberAttrAsString(attrs, "broker-squandered-amount"),
		BrokerBaseSquanderedAmount:              numberAttrAsString(attrs, "broker-base-squandered-amount"),
		JobSiteHours:                            numberAttrAsString(attrs, "job-site-hours"),
		CanDelete:                               boolAttr(attrs, "can-delete"),
		ExplicitIsTrailerRequiredForApproval:    boolAttr(attrs, "explicit-is-trailer-required-for-approval"),
		IsTrailerRequiredForApproval:            boolAttr(attrs, "is-trailer-required-for-approval"),
		ExplicitEnforceTicketNumberUniqueness:   boolAttr(attrs, "explicit-enforce-ticket-number-uniqueness"),
		EnforceTicketNumberUniqueness:           boolAttr(attrs, "enforce-ticket-number-uniqueness"),
		IsTimeCardCreatingTimeSheetLineItem:     boolAttr(attrs, "is-time-card-creating-time-sheet-line-item-explicit"),
		ExplicitIsInvoiceableWhenApproved:       boolAttr(attrs, "explicit-is-invoiceable-when-approved"),
		CurrentUserCanApprove:                   boolAttr(attrs, "current-user-can-approve"),
		IsManagementService:                     boolAttr(attrs, "is-management-service"),
		IsTimeCardStartAtEvidenceRequired:       boolAttr(attrs, "is-time-card-start-at-evidence-required"),
		GenerateBrokerInvoice:                   boolAttr(attrs, "generate-broker-invoice"),
		GenerateTruckerInvoice:                  boolAttr(attrs, "generate-trucker-invoice"),
	}

	if rel, ok := resource.Relationships["broker-tender"]; ok && rel.Data != nil {
		details.BrokerTenderID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["submitted-by"]; ok && rel.Data != nil {
		details.SubmittedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["time-card-cost-code-allocation"]; ok && rel.Data != nil {
		details.TimeCardCostCodeAllocationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
		details.TrailerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
		details.TractorID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job"]; ok && rel.Data != nil {
		details.JobID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
		details.JobSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		details.JobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["contractor"]; ok && rel.Data != nil {
		details.ContractorID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["accepted-customer-tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.AcceptedCustomerTenderJobScheduleShift = rel.Data.ID
	}
	if rel, ok := resource.Relationships["time-card-payroll-certification"]; ok && rel.Data != nil {
		details.TimeCardPayrollCertificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["time-card-approval-audit"]; ok && rel.Data != nil {
		details.TimeCardApprovalAuditID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["time-card-status-changes"]; ok {
		details.TimeCardStatusChangeIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resource.Relationships["service-type-unit-of-measure-quantities"]; ok {
		details.ServiceTypeUnitOfMeasureQuantityIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resource.Relationships["file-attachments"]; ok {
		details.FileAttachmentIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resource.Relationships["invoices"]; ok {
		details.InvoiceIDs = relationshipIDsToTypedStrings(rel)
	}
	if rel, ok := resource.Relationships["job-production-plan-time-card-approvers"]; ok {
		details.JobProductionPlanTimeCardApproverIDs = relationshipIDsToStrings(rel)
	}
	if rel, ok := resource.Relationships["job-production-plan-material-types"]; ok {
		details.JobProductionPlanMaterialTypeIDs = relationshipIDsToStrings(rel)
	}

	return details
}

func renderTimeCardDetails(cmd *cobra.Command, details timeCardDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.TicketNumber != "" {
		fmt.Fprintf(out, "Ticket Number: %s\n", details.TicketNumber)
	}
	if details.ApprovalProcess != "" {
		fmt.Fprintf(out, "Approval Process: %s\n", details.ApprovalProcess)
	}
	if details.ApprovalCount != "" {
		fmt.Fprintf(out, "Approval Count: %s\n", details.ApprovalCount)
	}
	if details.SubmitAt != "" {
		fmt.Fprintf(out, "Submit At: %s\n", details.SubmitAt)
	}

	if details.StartAt != "" || details.EndAt != "" || details.DurationMinutes != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Timing:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if details.StartAt != "" {
			fmt.Fprintf(out, "  Start: %s\n", details.StartAt)
		}
		if details.EndAt != "" {
			fmt.Fprintf(out, "  End: %s\n", details.EndAt)
		}
		if details.DownMinutes != "" {
			fmt.Fprintf(out, "  Down Minutes: %s\n", details.DownMinutes)
		}
		if details.DurationMinutes != "" {
			fmt.Fprintf(out, "  Duration Minutes: %s\n", details.DurationMinutes)
		}
		if details.TotalHours != "" {
			fmt.Fprintf(out, "  Total Hours: %s\n", details.TotalHours)
		}
		if details.TotalCustomerHours != "" {
			fmt.Fprintf(out, "  Total Customer Hours: %s\n", details.TotalCustomerHours)
		}
		if details.CreditedMinutes != "" {
			fmt.Fprintf(out, "  Credited Minutes: %s\n", details.CreditedMinutes)
		}
		if details.CreditedHours != "" {
			fmt.Fprintf(out, "  Credited Hours: %s\n", details.CreditedHours)
		}
	}

	if details.SubmittedTravelMinutes != "" || details.TravelBeforeMinutes != "" || details.TravelDuringMinutes != "" || details.TravelAfterMinutes != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Travel:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if details.SubmittedTravelMinutes != "" {
			fmt.Fprintf(out, "  Submitted Travel Minutes: %s\n", details.SubmittedTravelMinutes)
		}
		if details.MaximumTravelMinutes != "" {
			fmt.Fprintf(out, "  Maximum Travel Minutes: %s\n", details.MaximumTravelMinutes)
		}
		if details.TravelBeforeMinutes != "" {
			fmt.Fprintf(out, "  Travel Before Minutes: %s\n", details.TravelBeforeMinutes)
		}
		if details.TravelDuringMinutes != "" {
			fmt.Fprintf(out, "  Travel During Minutes: %s\n", details.TravelDuringMinutes)
		}
		if details.TravelAfterMinutes != "" {
			fmt.Fprintf(out, "  Travel After Minutes: %s\n", details.TravelAfterMinutes)
		}
		if details.SubmittedCustomerTravelMinutes != "" {
			fmt.Fprintf(out, "  Submitted Customer Travel Minutes: %s\n", details.SubmittedCustomerTravelMinutes)
		}
		if details.CustomerTenderMaximumTravelMinutes != "" {
			fmt.Fprintf(out, "  Customer Tender Max Travel Minutes: %s\n", details.CustomerTenderMaximumTravelMinutes)
		}
		if details.CustomerTravelMinutes != "" {
			fmt.Fprintf(out, "  Customer Travel Minutes: %s\n", details.CustomerTravelMinutes)
		}
	}

	if details.BrokerAmount != "" || details.CustomerAmount != "" || details.CostAmount != "" || details.RevenueAmount != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Financials:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if details.BrokerAmount != "" {
			fmt.Fprintf(out, "  Broker Amount: %s\n", details.BrokerAmount)
		}
		if details.CustomerAmount != "" {
			fmt.Fprintf(out, "  Customer Amount: %s\n", details.CustomerAmount)
		}
		if details.CostAmount != "" {
			fmt.Fprintf(out, "  Cost Amount: %s\n", details.CostAmount)
		}
		if details.RevenueAmount != "" {
			fmt.Fprintf(out, "  Revenue Amount: %s\n", details.RevenueAmount)
		}
		if details.SquanderedHours != "" {
			fmt.Fprintf(out, "  Squandered Hours: %s\n", details.SquanderedHours)
		}
		if details.CustomerSquanderedAmount != "" {
			fmt.Fprintf(out, "  Customer Squandered Amount: %s\n", details.CustomerSquanderedAmount)
		}
		if details.CustomerBaseSquanderedAmount != "" {
			fmt.Fprintf(out, "  Customer Base Squandered Amount: %s\n", details.CustomerBaseSquanderedAmount)
		}
		if details.BrokerSquanderedAmount != "" {
			fmt.Fprintf(out, "  Broker Squandered Amount: %s\n", details.BrokerSquanderedAmount)
		}
		if details.BrokerBaseSquanderedAmount != "" {
			fmt.Fprintf(out, "  Broker Base Squandered Amount: %s\n", details.BrokerBaseSquanderedAmount)
		}
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintf(out, "  Can Delete: %s\n", formatBool(details.CanDelete))
	fmt.Fprintf(out, "  Current User Can Approve: %s\n", formatBool(details.CurrentUserCanApprove))
	fmt.Fprintf(out, "  Management Service: %s\n", formatBool(details.IsManagementService))
	fmt.Fprintf(out, "  Time Card Start Evidence Required: %s\n", formatBool(details.IsTimeCardStartAtEvidenceRequired))
	fmt.Fprintf(out, "  Skip Auto Submission Upon MTXN Acceptance: %s\n", formatBool(details.SkipAutoSubmissionUponMaterialTxnAccept))
	fmt.Fprintf(out, "  Explicit Trailer Required: %s\n", formatBool(details.ExplicitIsTrailerRequiredForApproval))
	fmt.Fprintf(out, "  Trailer Required: %s\n", formatBool(details.IsTrailerRequiredForApproval))
	fmt.Fprintf(out, "  Explicit Ticket Number Uniqueness: %s\n", formatBool(details.ExplicitEnforceTicketNumberUniqueness))
	fmt.Fprintf(out, "  Enforce Ticket Number Uniqueness: %s\n", formatBool(details.EnforceTicketNumberUniqueness))
	fmt.Fprintf(out, "  Create Time Sheet Line Item Explicit: %s\n", formatBool(details.IsTimeCardCreatingTimeSheetLineItem))
	fmt.Fprintf(out, "  Explicit Invoiceable When Approved: %s\n", formatBool(details.ExplicitIsInvoiceableWhenApproved))
	fmt.Fprintf(out, "  Customer Invoiceable Without Override: %s\n", formatBool(details.IsCustomerInvoiceableWithoutOverride))
	fmt.Fprintf(out, "  Trucker Invoiceable Without Override: %s\n", formatBool(details.IsTruckerInvoiceableWithoutOverride))
	fmt.Fprintf(out, "  Generate Broker Invoice: %s\n", formatBool(details.GenerateBrokerInvoice))
	fmt.Fprintf(out, "  Generate Trucker Invoice: %s\n", formatBool(details.GenerateTruckerInvoice))

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Relationships:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	if details.BrokerTenderID != "" {
		fmt.Fprintf(out, "  Broker Tender: %s\n", details.BrokerTenderID)
	}
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "  Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.JobScheduleShiftID != "" {
		fmt.Fprintf(out, "  Job Schedule Shift: %s\n", details.JobScheduleShiftID)
	}
	if details.JobID != "" {
		fmt.Fprintf(out, "  Job: %s\n", details.JobID)
	}
	if details.JobSiteID != "" {
		fmt.Fprintf(out, "  Job Site: %s\n", details.JobSiteID)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "  Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "  Customer: %s\n", details.CustomerID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "  Broker: %s\n", details.BrokerID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "  Trucker: %s\n", details.TruckerID)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "  Driver: %s\n", details.DriverID)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "  Tractor: %s\n", details.TractorID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "  Trailer: %s\n", details.TrailerID)
	}
	if details.ContractorID != "" {
		fmt.Fprintf(out, "  Contractor: %s\n", details.ContractorID)
	}
	if details.SubmittedByID != "" {
		fmt.Fprintf(out, "  Submitted By: %s\n", details.SubmittedByID)
	}
	if details.TimeCardCostCodeAllocationID != "" {
		fmt.Fprintf(out, "  Cost Code Allocation: %s\n", details.TimeCardCostCodeAllocationID)
	}
	if details.TimeCardPayrollCertificationID != "" {
		fmt.Fprintf(out, "  Payroll Certification: %s\n", details.TimeCardPayrollCertificationID)
	}
	if details.TimeCardApprovalAuditID != "" {
		fmt.Fprintf(out, "  Approval Audit: %s\n", details.TimeCardApprovalAuditID)
	}
	if len(details.TimeCardStatusChangeIDs) > 0 {
		fmt.Fprintf(out, "  Status Changes: %s\n", strings.Join(details.TimeCardStatusChangeIDs, ", "))
	}
	if len(details.ServiceTypeUnitOfMeasureQuantityIDs) > 0 {
		fmt.Fprintf(out, "  STUOM Quantities: %s\n", strings.Join(details.ServiceTypeUnitOfMeasureQuantityIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "  File Attachments: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}
	if len(details.InvoiceIDs) > 0 {
		fmt.Fprintf(out, "  Invoices: %s\n", strings.Join(details.InvoiceIDs, ", "))
	}
	if len(details.JobProductionPlanTimeCardApproverIDs) > 0 {
		fmt.Fprintf(out, "  JPP Time Card Approvers: %s\n", strings.Join(details.JobProductionPlanTimeCardApproverIDs, ", "))
	}
	if len(details.JobProductionPlanMaterialTypeIDs) > 0 {
		fmt.Fprintf(out, "  JPP Material Types: %s\n", strings.Join(details.JobProductionPlanMaterialTypeIDs, ", "))
	}

	return nil
}

func numberAttrAsString(attrs map[string]any, key string) string {
	if attrs == nil {
		return ""
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func relationshipIDsToTypedStrings(rel jsonAPIRelationship) []string {
	ids := relationshipIDs(rel)
	if len(ids) == 0 {
		return nil
	}
	entries := make([]string, 0, len(ids))
	for _, id := range ids {
		if id.Type != "" {
			entries = append(entries, fmt.Sprintf("%s:%s", id.Type, id.ID))
		} else {
			entries = append(entries, id.ID)
		}
	}
	return entries
}
