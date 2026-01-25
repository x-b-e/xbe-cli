package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobScheduleShiftsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobScheduleShiftDetails struct {
	ID                                         string   `json:"id"`
	StartAt                                    string   `json:"start_at,omitempty"`
	EndAt                                      string   `json:"end_at,omitempty"`
	StartDate                                  string   `json:"start_date,omitempty"`
	TimeZoneID                                 string   `json:"time_zone_id,omitempty"`
	TruckCount                                 string   `json:"truck_count,omitempty"`
	DispatchInstructions                       string   `json:"dispatch_instructions,omitempty"`
	IsPlannedProductive                        bool     `json:"is_planned_productive"`
	CancelledAt                                string   `json:"cancelled_at,omitempty"`
	SuppressAutomatedShiftFeedback             bool     `json:"suppress_automated_shift_feedback"`
	ExpectedMaterialTransactionCount           string   `json:"expected_material_transaction_count,omitempty"`
	ExpectedMaterialTransactionTons            string   `json:"expected_material_transaction_tons,omitempty"`
	IsFlexible                                 bool     `json:"is_flexible"`
	StartAtMin                                 string   `json:"start_at_min,omitempty"`
	StartAtMax                                 string   `json:"start_at_max,omitempty"`
	CurrentUserCanSplit                        bool     `json:"current_user_can_split"`
	CurrentUserCanReschedule                   bool     `json:"current_user_can_reschedule"`
	ColorHex                                   string   `json:"color_hex,omitempty"`
	IsManaged                                  bool     `json:"is_managed"`
	IsSubsequentShiftInDriverDay               bool     `json:"is_subsequent_shift_in_driver_day"`
	IsTruckerIncidentCreationAutomated         bool     `json:"is_trucker_incident_creation_automated"`
	IsTruckerIncidentCreationAutomatedExplicit bool     `json:"is_trucker_incident_creation_automated_explicit"`
	StartSiteType                              string   `json:"start_site_type,omitempty"`
	StartSiteID                                string   `json:"start_site_id,omitempty"`
	StartLocationID                            string   `json:"start_location_id,omitempty"`
	ShowPlannerInfoToDrivers                   bool     `json:"show_planner_info_to_drivers"`
	ExcludeSamePlanDriverMoves                 bool     `json:"exclude_same_plan_driver_moves"`
	PlannerName                                string   `json:"planner_name,omitempty"`
	PlannerMobileNumber                        string   `json:"planner_mobile_number,omitempty"`
	PlannerMobileNumberFormatted               string   `json:"planner_mobile_number_formatted,omitempty"`
	DriverAssignmentRuleTextCached             string   `json:"driver_assignment_rule_text_cached,omitempty"`
	TicketReportID                             string   `json:"ticket_report_id,omitempty"`
	JobID                                      string   `json:"job_id,omitempty"`
	TrailerClassificationID                    string   `json:"trailer_classification_id,omitempty"`
	ProjectLaborClassificationID               string   `json:"project_labor_classification_id,omitempty"`
	JobSiteID                                  string   `json:"job_site_id,omitempty"`
	CustomerID                                 string   `json:"customer_id,omitempty"`
	BrokerID                                   string   `json:"broker_id,omitempty"`
	LineupJobScheduleShiftID                   string   `json:"lineup_job_schedule_shift_id,omitempty"`
	JobProductionPlanIDs                       []string `json:"job_production_plan_ids,omitempty"`
	AcceptedBrokerTenderJobScheduleShiftID     string   `json:"accepted_broker_tender_job_schedule_shift_id,omitempty"`
	TimeCardID                                 string   `json:"time_card_id,omitempty"`
	AcceptedCustomerTenderJobScheduleShiftID   string   `json:"accepted_customer_tender_job_schedule_shift_id,omitempty"`
	ExpectedTimeOfArrivalIDs                   []string `json:"expected_time_of_arrival_ids,omitempty"`
	TenderJobScheduleShiftIDs                  []string `json:"tender_job_schedule_shift_ids,omitempty"`
	MaterialPurchaseOrderReleaseIDs            []string `json:"material_purchase_order_release_ids,omitempty"`
	JobScheduleShiftSplitIDs                   []string `json:"job_schedule_shift_split_ids,omitempty"`
	OfferedBrokerTenderJobScheduleShiftIDs     []string `json:"offered_broker_tender_job_schedule_shift_ids,omitempty"`
	AcceptedBrokerTenderID                     string   `json:"accepted_broker_tender_id,omitempty"`
	AcceptedDriverID                           string   `json:"accepted_driver_id,omitempty"`
	AcceptedTrailerID                          string   `json:"accepted_trailer_id,omitempty"`
	AcceptedTruckerID                          string   `json:"accepted_trucker_id,omitempty"`
	AcceptedServiceEventIDs                    []string `json:"accepted_service_event_ids,omitempty"`
	ReadyToWorkID                              string   `json:"ready_to_work_id,omitempty"`
	DriverAssignmentRuleIDs                    []string `json:"driver_assignment_rule_ids,omitempty"`
}

func newJobScheduleShiftsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job schedule shift details",
		Long: `Show the full details of a job schedule shift.

Output Fields:
  ID
  Start At / End At / Start Date
  Time Zone ID / Truck Count
  Dispatch Instructions
  Is Planned Productive / Is Flexible / Is Managed
  Cancelled At / Suppress Automated Shift Feedback
  Expected Material Transaction Count / Tons
  Start At Min / Start At Max
  Current User Can Split / Reschedule
  Color Hex
  Is Subsequent Shift In Driver Day
  Trucker Incident Creation Automated (computed + explicit)
  Start Site (type + ID) / Start Location
  Planner Info
  Driver Assignment Rule Text
  Job / Trailer Classification / Project Labor Classification
  Job Site / Customer / Broker
  Ticket Report
  Lineup Job Schedule Shift
  Job Production Plans (via job)
  Accepted Broker Tender Job Schedule Shift / Accepted Customer Tender Job Schedule Shift
  Time Card
  Expected Time Of Arrivals
  Tender Job Schedule Shifts
  Material Purchase Order Releases
  Job Schedule Shift Splits
  Offered Broker Tender Job Schedule Shifts
  Accepted Broker Tender / Driver / Trailer / Trucker
  Accepted Service Events / Ready To Work
  Driver Assignment Rules

Arguments:
  <id>    The job schedule shift ID (required). Use the list command to find IDs.`,
		Example: `  # Show a job schedule shift
  xbe view job-schedule-shifts show 123

  # Show as JSON
  xbe view job-schedule-shifts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobScheduleShiftsShow,
	}
	initJobScheduleShiftsShowFlags(cmd)
	return cmd
}

func init() {
	jobScheduleShiftsCmd.AddCommand(newJobScheduleShiftsShowCmd())
}

func initJobScheduleShiftsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobScheduleShiftsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobScheduleShiftsShowOptions(cmd)
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
		return fmt.Errorf("job schedule shift id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-schedule-shifts]", strings.Join([]string{
		"start-at",
		"end-at",
		"start-date",
		"time-zone-id",
		"truck-count",
		"dispatch-instructions",
		"is-planned-productive",
		"cancelled-at",
		"suppress-automated-shift-feedback",
		"expected-material-transaction-count",
		"expected-material-transaction-tons",
		"is-flexible",
		"start-at-min",
		"start-at-max",
		"current-user-can-split",
		"current-user-can-reschedule",
		"color-hex",
		"is-managed",
		"is-subsequent-shift-in-driver-day",
		"is-trucker-incident-creation-automated",
		"is-trucker-incident-creation-automated-explicit",
		"start-site-type",
		"show-planner-info-to-drivers",
		"exclude-same-plan-driver-moves",
		"planner-name",
		"planner-mobile-number",
		"planner-mobile-number-formatted",
		"driver-assignment-rule-text-cached",
		"ticket-report",
		"job",
		"trailer-classification",
		"project-labor-classification",
		"job-site",
		"customer",
		"broker",
		"lineup-job-schedule-shift",
		"job-production-plans-via-job",
		"accepted-broker-tender-job-schedule-shift",
		"time-card",
		"accepted-customer-tender-job-schedule-shift",
		"expected-time-of-arrivals",
		"tender-job-schedule-shifts",
		"material-purchase-order-releases",
		"job-schedule-shift-splits",
		"offered-broker-tender-job-schedule-shifts",
		"start-site",
		"start-location",
		"accepted-broker-tender",
		"accepted-driver",
		"accepted-trailer",
		"accepted-trucker",
		"accepted-service-events",
		"ready-to-work",
		"driver-assignment-rules",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/job-schedule-shifts/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildJobScheduleShiftDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobScheduleShiftDetails(cmd, details)
}

func parseJobScheduleShiftsShowOptions(cmd *cobra.Command) (jobScheduleShiftsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return jobScheduleShiftsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return jobScheduleShiftsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return jobScheduleShiftsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return jobScheduleShiftsShowOptions{}, err
	}

	return jobScheduleShiftsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobScheduleShiftDetails(resp jsonAPISingleResponse) jobScheduleShiftDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobScheduleShiftDetails{
		ID:                                 resource.ID,
		StartAt:                            formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                              formatDateTime(stringAttr(attrs, "end-at")),
		StartDate:                          stringAttr(attrs, "start-date"),
		TimeZoneID:                         stringAttr(attrs, "time-zone-id"),
		TruckCount:                         stringAttr(attrs, "truck-count"),
		DispatchInstructions:               stringAttr(attrs, "dispatch-instructions"),
		IsPlannedProductive:                boolAttr(attrs, "is-planned-productive"),
		CancelledAt:                        formatDateTime(stringAttr(attrs, "cancelled-at")),
		SuppressAutomatedShiftFeedback:     boolAttr(attrs, "suppress-automated-shift-feedback"),
		ExpectedMaterialTransactionCount:   stringAttr(attrs, "expected-material-transaction-count"),
		ExpectedMaterialTransactionTons:    stringAttr(attrs, "expected-material-transaction-tons"),
		IsFlexible:                         boolAttr(attrs, "is-flexible"),
		StartAtMin:                         formatDateTime(stringAttr(attrs, "start-at-min")),
		StartAtMax:                         formatDateTime(stringAttr(attrs, "start-at-max")),
		CurrentUserCanSplit:                boolAttr(attrs, "current-user-can-split"),
		CurrentUserCanReschedule:           boolAttr(attrs, "current-user-can-reschedule"),
		ColorHex:                           stringAttr(attrs, "color-hex"),
		IsManaged:                          boolAttr(attrs, "is-managed"),
		IsSubsequentShiftInDriverDay:       boolAttr(attrs, "is-subsequent-shift-in-driver-day"),
		IsTruckerIncidentCreationAutomated: boolAttr(attrs, "is-trucker-incident-creation-automated"),
		IsTruckerIncidentCreationAutomatedExplicit: boolAttr(attrs, "is-trucker-incident-creation-automated-explicit"),
		StartSiteType:                  stringAttr(attrs, "start-site-type"),
		ShowPlannerInfoToDrivers:       boolAttr(attrs, "show-planner-info-to-drivers"),
		ExcludeSamePlanDriverMoves:     boolAttr(attrs, "exclude-same-plan-driver-moves"),
		PlannerName:                    stringAttr(attrs, "planner-name"),
		PlannerMobileNumber:            stringAttr(attrs, "planner-mobile-number"),
		PlannerMobileNumberFormatted:   stringAttr(attrs, "planner-mobile-number-formatted"),
		DriverAssignmentRuleTextCached: stringAttr(attrs, "driver-assignment-rule-text-cached"),
	}

	if rel, ok := resource.Relationships["ticket-report"]; ok && rel.Data != nil {
		details.TicketReportID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job"]; ok && rel.Data != nil {
		details.JobID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer-classification"]; ok && rel.Data != nil {
		details.TrailerClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-labor-classification"]; ok && rel.Data != nil {
		details.ProjectLaborClassificationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
		details.JobSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["lineup-job-schedule-shift"]; ok && rel.Data != nil {
		details.LineupJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["accepted-broker-tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.AcceptedBrokerTenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		details.TimeCardID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["accepted-customer-tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.AcceptedCustomerTenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["start-site"]; ok && rel.Data != nil {
		details.StartSiteID = rel.Data.ID
		if details.StartSiteType == "" {
			details.StartSiteType = rel.Data.Type
		}
	}
	if rel, ok := resource.Relationships["start-location"]; ok && rel.Data != nil {
		details.StartLocationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["accepted-broker-tender"]; ok && rel.Data != nil {
		details.AcceptedBrokerTenderID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["accepted-driver"]; ok && rel.Data != nil {
		details.AcceptedDriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["accepted-trailer"]; ok && rel.Data != nil {
		details.AcceptedTrailerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["accepted-trucker"]; ok && rel.Data != nil {
		details.AcceptedTruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["ready-to-work"]; ok && rel.Data != nil {
		details.ReadyToWorkID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["job-production-plans-via-job"]; ok {
		details.JobProductionPlanIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["expected-time-of-arrivals"]; ok {
		details.ExpectedTimeOfArrivalIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shifts"]; ok {
		details.TenderJobScheduleShiftIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["material-purchase-order-releases"]; ok {
		details.MaterialPurchaseOrderReleaseIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["job-schedule-shift-splits"]; ok {
		details.JobScheduleShiftSplitIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["offered-broker-tender-job-schedule-shifts"]; ok {
		details.OfferedBrokerTenderJobScheduleShiftIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["accepted-service-events"]; ok {
		details.AcceptedServiceEventIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["driver-assignment-rules"]; ok {
		details.DriverAssignmentRuleIDs = relationshipIDList(rel)
	}

	return details
}

func renderJobScheduleShiftDetails(cmd *cobra.Command, details jobScheduleShiftDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	if details.StartDate != "" {
		fmt.Fprintf(out, "Start Date: %s\n", details.StartDate)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone ID: %s\n", details.TimeZoneID)
	}
	if details.TruckCount != "" {
		fmt.Fprintf(out, "Truck Count: %s\n", details.TruckCount)
	}
	if details.DispatchInstructions != "" {
		fmt.Fprintf(out, "Dispatch Instructions: %s\n", details.DispatchInstructions)
	}
	fmt.Fprintf(out, "Is Planned Productive: %v\n", details.IsPlannedProductive)
	fmt.Fprintf(out, "Is Flexible: %v\n", details.IsFlexible)
	fmt.Fprintf(out, "Is Managed: %v\n", details.IsManaged)
	fmt.Fprintf(out, "Is Subsequent Shift In Driver Day: %v\n", details.IsSubsequentShiftInDriverDay)
	fmt.Fprintf(out, "Suppress Automated Shift Feedback: %v\n", details.SuppressAutomatedShiftFeedback)
	if details.CancelledAt != "" {
		fmt.Fprintf(out, "Cancelled At: %s\n", details.CancelledAt)
	}
	if details.ExpectedMaterialTransactionCount != "" {
		fmt.Fprintf(out, "Expected Material Transaction Count: %s\n", details.ExpectedMaterialTransactionCount)
	}
	if details.ExpectedMaterialTransactionTons != "" {
		fmt.Fprintf(out, "Expected Material Transaction Tons: %s\n", details.ExpectedMaterialTransactionTons)
	}
	if details.StartAtMin != "" {
		fmt.Fprintf(out, "Start At Min: %s\n", details.StartAtMin)
	}
	if details.StartAtMax != "" {
		fmt.Fprintf(out, "Start At Max: %s\n", details.StartAtMax)
	}
	fmt.Fprintf(out, "Current User Can Split: %v\n", details.CurrentUserCanSplit)
	fmt.Fprintf(out, "Current User Can Reschedule: %v\n", details.CurrentUserCanReschedule)
	if details.ColorHex != "" {
		fmt.Fprintf(out, "Color Hex: %s\n", details.ColorHex)
	}
	fmt.Fprintf(out, "Trucker Incident Creation Automated: %v\n", details.IsTruckerIncidentCreationAutomated)
	fmt.Fprintf(out, "Trucker Incident Creation Automated Explicit: %v\n", details.IsTruckerIncidentCreationAutomatedExplicit)
	if details.StartSiteID != "" {
		fmt.Fprintf(out, "Start Site: %s/%s\n", details.StartSiteType, details.StartSiteID)
	}
	if details.StartLocationID != "" {
		fmt.Fprintf(out, "Start Location: %s\n", details.StartLocationID)
	}
	if details.ShowPlannerInfoToDrivers {
		fmt.Fprintln(out, "Show Planner Info To Drivers: true")
	} else {
		fmt.Fprintln(out, "Show Planner Info To Drivers: false")
	}
	if details.ExcludeSamePlanDriverMoves {
		fmt.Fprintln(out, "Exclude Same Plan Driver Moves: true")
	} else {
		fmt.Fprintln(out, "Exclude Same Plan Driver Moves: false")
	}
	if details.PlannerName != "" {
		fmt.Fprintf(out, "Planner Name: %s\n", details.PlannerName)
	}
	if details.PlannerMobileNumber != "" {
		fmt.Fprintf(out, "Planner Mobile Number: %s\n", details.PlannerMobileNumber)
	}
	if details.PlannerMobileNumberFormatted != "" {
		fmt.Fprintf(out, "Planner Mobile Number (Formatted): %s\n", details.PlannerMobileNumberFormatted)
	}
	if details.DriverAssignmentRuleTextCached != "" {
		fmt.Fprintf(out, "Driver Assignment Rule Text: %s\n", details.DriverAssignmentRuleTextCached)
	}

	if details.JobID != "" {
		fmt.Fprintf(out, "Job: %s\n", details.JobID)
	}
	if details.TrailerClassificationID != "" {
		fmt.Fprintf(out, "Trailer Classification: %s\n", details.TrailerClassificationID)
	}
	if details.ProjectLaborClassificationID != "" {
		fmt.Fprintf(out, "Project Labor Classification: %s\n", details.ProjectLaborClassificationID)
	}
	if details.JobSiteID != "" {
		fmt.Fprintf(out, "Job Site: %s\n", details.JobSiteID)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer: %s\n", details.CustomerID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.TicketReportID != "" {
		fmt.Fprintf(out, "Ticket Report: %s\n", details.TicketReportID)
	}
	if details.LineupJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Lineup Job Schedule Shift: %s\n", details.LineupJobScheduleShiftID)
	}
	if len(details.JobProductionPlanIDs) > 0 {
		fmt.Fprintf(out, "Job Production Plans: %s\n", strings.Join(details.JobProductionPlanIDs, ", "))
	}
	if details.AcceptedBrokerTenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Accepted Broker Tender Job Schedule Shift: %s\n", details.AcceptedBrokerTenderJobScheduleShiftID)
	}
	if details.AcceptedCustomerTenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Accepted Customer Tender Job Schedule Shift: %s\n", details.AcceptedCustomerTenderJobScheduleShiftID)
	}
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card: %s\n", details.TimeCardID)
	}
	if len(details.ExpectedTimeOfArrivalIDs) > 0 {
		fmt.Fprintf(out, "Expected Time Of Arrivals: %s\n", strings.Join(details.ExpectedTimeOfArrivalIDs, ", "))
	}
	if len(details.TenderJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Tender Job Schedule Shifts: %s\n", strings.Join(details.TenderJobScheduleShiftIDs, ", "))
	}
	if len(details.MaterialPurchaseOrderReleaseIDs) > 0 {
		fmt.Fprintf(out, "Material Purchase Order Releases: %s\n", strings.Join(details.MaterialPurchaseOrderReleaseIDs, ", "))
	}
	if len(details.JobScheduleShiftSplitIDs) > 0 {
		fmt.Fprintf(out, "Job Schedule Shift Splits: %s\n", strings.Join(details.JobScheduleShiftSplitIDs, ", "))
	}
	if len(details.OfferedBrokerTenderJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Offered Broker Tender Job Schedule Shifts: %s\n", strings.Join(details.OfferedBrokerTenderJobScheduleShiftIDs, ", "))
	}
	if details.AcceptedBrokerTenderID != "" {
		fmt.Fprintf(out, "Accepted Broker Tender: %s\n", details.AcceptedBrokerTenderID)
	}
	if details.AcceptedDriverID != "" {
		fmt.Fprintf(out, "Accepted Driver: %s\n", details.AcceptedDriverID)
	}
	if details.AcceptedTrailerID != "" {
		fmt.Fprintf(out, "Accepted Trailer: %s\n", details.AcceptedTrailerID)
	}
	if details.AcceptedTruckerID != "" {
		fmt.Fprintf(out, "Accepted Trucker: %s\n", details.AcceptedTruckerID)
	}
	if len(details.AcceptedServiceEventIDs) > 0 {
		fmt.Fprintf(out, "Accepted Service Events: %s\n", strings.Join(details.AcceptedServiceEventIDs, ", "))
	}
	if details.ReadyToWorkID != "" {
		fmt.Fprintf(out, "Ready To Work: %s\n", details.ReadyToWorkID)
	}
	if len(details.DriverAssignmentRuleIDs) > 0 {
		fmt.Fprintf(out, "Driver Assignment Rules: %s\n", strings.Join(details.DriverAssignmentRuleIDs, ", "))
	}

	return nil
}
