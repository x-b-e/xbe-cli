package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type tenderJobScheduleShiftsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderJobScheduleShiftDetails struct {
	ID                                                      string `json:"id"`
	TruckCount                                              string `json:"truck_count,omitempty"`
	NotifyBeforeShiftStartsHours                            string `json:"notify_before_shift_starts_hours,omitempty"`
	NotifyAfterShiftEndsHours                               string `json:"notify_after_shift_ends_hours,omitempty"`
	NotifyDriverOnLateShiftAssignment                       bool   `json:"notify_driver_on_late_shift_assignment"`
	ExplicitNotifyDriverWhenGPSNotAvailable                 bool   `json:"explicit_notify_driver_when_gps_not_available"`
	NotifyDriverWhenGPSNotAvailable                         bool   `json:"notify_driver_when_gps_not_available"`
	IsAutomatedJobSiteTimeCreationDisabled                  bool   `json:"is_automated_job_site_time_creation_disabled"`
	IsTimeCardPayrollCertificationRequiredExplicit          string `json:"is_time_card_payroll_certification_required_explicit,omitempty"`
	IsTimeCardCreatingTimeSheetLineItemExplicit             string `json:"is_time_card_creating_time_sheet_line_item_explicit,omitempty"`
	SkipValidateDriverAssignmentRuleEvaluation              bool   `json:"skip_validate_driver_assignment_rule_evaluation"`
	DriverAssignmentRuleOverrideReason                      string `json:"driver_assignment_rule_override_reason,omitempty"`
	SkipMaterialTransactionImageExtraction                  bool   `json:"skip_material_transaction_image_extraction"`
	DisablePreStartNotifications                            bool   `json:"disable_pre_start_notifications"`
	CancelledAt                                             string `json:"cancelled_at,omitempty"`
	RejectedAt                                              string `json:"rejected_at,omitempty"`
	ReturnedAt                                              string `json:"returned_at,omitempty"`
	StatusChangeComment                                     string `json:"status_change_comment,omitempty"`
	AllTripsEntered                                         bool   `json:"all_trips_entered"`
	HoursAfterWhichOvertimeApplies                          string `json:"hours_after_which_overtime_applies,omitempty"`
	TravelMiles                                             string `json:"travel_miles,omitempty"`
	TravelMinutes                                           string `json:"travel_minutes,omitempty"`
	BillableTravelMinutes                                   string `json:"billable_travel_minutes,omitempty"`
	LoadedTonsMax                                           string `json:"loaded_tons_max,omitempty"`
	LoadedTonsMaxEffective                                  string `json:"loaded_tons_max_effective,omitempty"`
	StartAt                                                 string `json:"start_at,omitempty"`
	GrossWeightLegalLimitLbsExplicit                        string `json:"gross_weight_legal_limit_lbs_explicit,omitempty"`
	AutoCheckInDriverOnArrivalAtStartSite                   bool   `json:"auto_check_in_driver_on_arrival_at_start_site"`
	ExplicitIsExpectingTimeCard                             bool   `json:"explicit_is_expecting_time_card"`
	IsExpectingTimeCard                                     bool   `json:"is_expecting_time_card"`
	MaterialTransactionTonsMax                              string `json:"material_transaction_tons_max,omitempty"`
	ExplicitMaterialTransactionTonsMax                      string `json:"explicit_material_transaction_tons_max,omitempty"`
	TimeZoneID                                              string `json:"time_zone_id,omitempty"`
	SuppressAutomatedShiftFeedback                          bool   `json:"suppress_automated_shift_feedback"`
	GrossWeightLegalLimitLbs                                string `json:"gross_weight_legal_limit_lbs,omitempty"`
	FirstMaterialTransactionLoadedAt                        string `json:"first_material_transaction_loaded_at,omitempty"`
	ImpliedTimeCardStartAt                                  string `json:"implied_time_card_start_at,omitempty"`
	TenderType                                              string `json:"tender_type,omitempty"`
	TrackedPct                                              string `json:"tracked_pct,omitempty"`
	IsExpectingMaterialTransactions                         bool   `json:"is_expecting_material_transactions"`
	ExpectingMaterialTransactionsMessage                    string `json:"expecting_material_transactions_message,omitempty"`
	Rates                                                   any    `json:"rates,omitempty"`
	CurrentUserCanCreateDriverAssignmentRefusal             bool   `json:"current_user_can_create_driver_assignment_refusal"`
	CanDriverAssignmentBeRefused                            bool   `json:"can_driver_assignment_be_refused"`
	DriverDaySequenceIndex                                  string `json:"driver_day_sequence_index,omitempty"`
	MaterialTransactionStatus                               string `json:"material_transaction_status,omitempty"`
	CurrentUserCanCreateMaterialTransactions                bool   `json:"current_user_can_create_material_transactions"`
	CurrentUserCanChangeTruck                               bool   `json:"current_user_can_change_truck"`
	IsNonDriverPermittedToCheckIn                           bool   `json:"is_non_driver_permitted_to_check_in"`
	TruckerCanCreateMaterialTransactions                    bool   `json:"trucker_can_create_material_transactions"`
	TruckerCanCreateMaterialTransactionsDisabledByMaxTimeAt string `json:"trucker_can_create_material_transactions_disabled_by_max_time_at,omitempty"`
	Managed                                                 bool   `json:"managed"`
	IsManaged                                               bool   `json:"is_managed"`
	ResetHoursAfterWhichOvertimeApplies                     bool   `json:"reset_hours_after_which_overtime_applies"`
	SellerOperationsContactAssignedAt                       string `json:"seller_operations_contact_assigned_at,omitempty"`

	StatusChangedByID                         string   `json:"status_changed_by_id,omitempty"`
	TenderJobScheduleShiftTimeCardReviewIDs   []string `json:"tender_job_schedule_shift_time_card_review_ids,omitempty"`
	AcceptedTruckerID                         string   `json:"accepted_trucker_id,omitempty"`
	BrokerTenderID                            string   `json:"broker_tender_id,omitempty"`
	LineupTrailerClassificationID             string   `json:"lineup_trailer_classification_id,omitempty"`
	DriverAssignmentRefusalIDs                []string `json:"driver_assignment_refusal_ids,omitempty"`
	CurrentDriverAssignmentAcknowledgementIDs []string `json:"current_driver_assignment_acknowledgement_ids,omitempty"`
	ShiftTimeCardRequisitionIDs               []string `json:"shift_time_card_requisition_ids,omitempty"`
	MaterialPurchaseOrderReleaseIDs           []string `json:"material_purchase_order_release_ids,omitempty"`
	SiteEventIDs                              []string `json:"site_event_ids,omitempty"`
	JobProductionPlanMaterialTypeIDs          []string `json:"job_production_plan_material_type_ids,omitempty"`
	TenderID                                  string   `json:"tender_id,omitempty"`
	TenderRelationshipType                    string   `json:"tender_relationship_type,omitempty"`
	JobScheduleShiftID                        string   `json:"job_schedule_shift_id,omitempty"`
	TrailerID                                 string   `json:"trailer_id,omitempty"`
	TractorID                                 string   `json:"tractor_id,omitempty"`
	SellerOperationsContactID                 string   `json:"seller_operations_contact_id,omitempty"`
	SellerOperationsContactDraftID            string   `json:"seller_operations_contact_draft_id,omitempty"`
	TrailerDraftID                            string   `json:"trailer_draft_id,omitempty"`
	RetainerID                                string   `json:"retainer_id,omitempty"`
	TimeCardIDs                               []string `json:"time_card_ids,omitempty"`
	TimeCardPreApprovalID                     string   `json:"time_card_pre_approval_id,omitempty"`
	ServiceEventIDs                           []string `json:"service_event_ids,omitempty"`
	ReadyToWorkID                             string   `json:"ready_to_work_id,omitempty"`
	AcceptedBrokerTenderJobScheduleShiftID    string   `json:"accepted_broker_tender_job_schedule_shift_id,omitempty"`
	ExpectedTimeOfArrivalIDs                  []string `json:"expected_time_of_arrival_ids,omitempty"`
	ShiftFeedbackIDs                          []string `json:"shift_feedback_ids,omitempty"`
	ProductionIncidentIDs                     []string `json:"production_incident_ids,omitempty"`
	JobProductionPlanBroadcastMessageIDs      []string `json:"job_production_plan_broadcast_message_ids,omitempty"`
	TripIDs                                   []string `json:"trip_ids,omitempty"`
	MaterialTransactionIDs                    []string `json:"material_transaction_ids,omitempty"`
	TruckerShiftSetID                         string   `json:"trucker_shift_set_id,omitempty"`
	ShiftDriverIDs                            []string `json:"shift_driver_ids,omitempty"`
	PrimaryDriverID                           string   `json:"primary_driver_id,omitempty"`
}

func newTenderJobScheduleShiftsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender job schedule shift details",
		Long: `Show the full details of a tender job schedule shift.

Output Fields include:
  Core attributes, status, timestamps, assignment flags
  Driver, trailer, tractor, tender, and job schedule shift relationships
  Associated time cards, trips, events, and related resources

Arguments:
  <id>    The tender job schedule shift ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show tender job schedule shift details
  xbe view tender-job-schedule-shifts show 123

  # JSON output
  xbe view tender-job-schedule-shifts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderJobScheduleShiftsShow,
	}
	initTenderJobScheduleShiftsShowFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftsCmd.AddCommand(newTenderJobScheduleShiftsShowCmd())
}

func initTenderJobScheduleShiftsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTenderJobScheduleShiftsShowOptions(cmd)
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
		return fmt.Errorf("tender job schedule shift id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shifts/"+id, nil)
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

	details := buildTenderJobScheduleShiftDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderJobScheduleShiftDetails(cmd, details)
}

func parseTenderJobScheduleShiftsShowOptions(cmd *cobra.Command) (tenderJobScheduleShiftsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderJobScheduleShiftDetails(resp jsonAPISingleResponse) tenderJobScheduleShiftDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := tenderJobScheduleShiftDetails{
		ID:                                             resource.ID,
		TruckCount:                                     stringAttr(attrs, "truck-count"),
		NotifyBeforeShiftStartsHours:                   stringAttr(attrs, "notify-before-shift-starts-hours"),
		NotifyAfterShiftEndsHours:                      stringAttr(attrs, "notify-after-shift-ends-hours"),
		NotifyDriverOnLateShiftAssignment:              boolAttr(attrs, "notify-driver-on-late-shift-assignment"),
		ExplicitNotifyDriverWhenGPSNotAvailable:        boolAttr(attrs, "explicit-notify-driver-when-gps-not-available"),
		NotifyDriverWhenGPSNotAvailable:                boolAttr(attrs, "notify-driver-when-gps-not-available"),
		IsAutomatedJobSiteTimeCreationDisabled:         boolAttr(attrs, "is-automated-job-site-time-creation-disabled"),
		IsTimeCardPayrollCertificationRequiredExplicit: stringAttr(attrs, "is-time-card-payroll-certification-required-explicit"),
		IsTimeCardCreatingTimeSheetLineItemExplicit:    stringAttr(attrs, "is-time-card-creating-time-sheet-line-item-explicit"),
		SkipValidateDriverAssignmentRuleEvaluation:     boolAttr(attrs, "skip-validate-driver-assignment-rule-evaluation"),
		DriverAssignmentRuleOverrideReason:             stringAttr(attrs, "driver-assignment-rule-override-reason"),
		SkipMaterialTransactionImageExtraction:         boolAttr(attrs, "skip-material-transaction-image-extraction"),
		DisablePreStartNotifications:                   boolAttr(attrs, "disable-pre-start-notifications"),
		CancelledAt:                                    formatDateTime(stringAttr(attrs, "cancelled-at")),
		RejectedAt:                                     formatDateTime(stringAttr(attrs, "rejected-at")),
		ReturnedAt:                                     formatDateTime(stringAttr(attrs, "returned-at")),
		StatusChangeComment:                            stringAttr(attrs, "status-change-comment"),
		AllTripsEntered:                                boolAttr(attrs, "all-trips-entered"),
		HoursAfterWhichOvertimeApplies:                 stringAttr(attrs, "hours-after-which-overtime-applies"),
		TravelMiles:                                    stringAttr(attrs, "travel-miles"),
		TravelMinutes:                                  stringAttr(attrs, "travel-minutes"),
		BillableTravelMinutes:                          stringAttr(attrs, "billable-travel-minutes"),
		LoadedTonsMax:                                  stringAttr(attrs, "loaded-tons-max"),
		LoadedTonsMaxEffective:                         stringAttr(attrs, "loaded-tons-max-effective"),
		StartAt:                                        formatDateTime(stringAttr(attrs, "start-at")),
		GrossWeightLegalLimitLbsExplicit:               stringAttr(attrs, "gross-weight-legal-limit-lbs-explicit"),
		AutoCheckInDriverOnArrivalAtStartSite:          boolAttr(attrs, "auto-check-in-driver-on-arrival-at-start-site"),
		ExplicitIsExpectingTimeCard:                    boolAttr(attrs, "explicit-is-expecting-time-card"),
		IsExpectingTimeCard:                            boolAttr(attrs, "is-expecting-time-card"),
		MaterialTransactionTonsMax:                     stringAttr(attrs, "material-transaction-tons-max"),
		ExplicitMaterialTransactionTonsMax:             stringAttr(attrs, "explicit-material-transaction-tons-max"),
		TimeZoneID:                                     stringAttr(attrs, "time-zone-id"),
		SuppressAutomatedShiftFeedback:                 boolAttr(attrs, "suppress-automated-shift-feedback"),
		GrossWeightLegalLimitLbs:                       stringAttr(attrs, "gross-weight-legal-limit-lbs"),
		FirstMaterialTransactionLoadedAt:               formatDateTime(stringAttr(attrs, "first-material-transaction-loaded-at")),
		ImpliedTimeCardStartAt:                         formatDateTime(stringAttr(attrs, "implied-time-card-start-at")),
		TenderType:                                     stringAttr(attrs, "tender-type"),
		TrackedPct:                                     stringAttr(attrs, "tracked-pct"),
		IsExpectingMaterialTransactions:                boolAttr(attrs, "is-expecting-material-transactions"),
		ExpectingMaterialTransactionsMessage:           stringAttr(attrs, "expecting-material-transactions-message"),
		Rates:                                          anyAttr(attrs, "rates"),
		CurrentUserCanCreateDriverAssignmentRefusal:    boolAttr(attrs, "current-user-can-create-driver-assignment-refusal"),
		CanDriverAssignmentBeRefused:                   boolAttr(attrs, "can-driver-assignment-be-refused"),
		DriverDaySequenceIndex:                         stringAttr(attrs, "driver-day-sequence-index"),
		MaterialTransactionStatus:                      stringAttr(attrs, "material-transaction-status"),
		CurrentUserCanCreateMaterialTransactions:       boolAttr(attrs, "current-user-can-create-material-transactions"),
		CurrentUserCanChangeTruck:                      boolAttr(attrs, "current-user-can-change-truck"),
		IsNonDriverPermittedToCheckIn:                  boolAttr(attrs, "is-non-driver-permitted-to-check-in"),
		TruckerCanCreateMaterialTransactions:           boolAttr(attrs, "trucker-can-create-material-transactions"),
		TruckerCanCreateMaterialTransactionsDisabledByMaxTimeAt: stringAttr(attrs, "trucker-can-create-material-transactions-disabled-by-max-time-at"),
		Managed:                             boolAttr(attrs, "managed"),
		IsManaged:                           boolAttr(attrs, "is-managed"),
		ResetHoursAfterWhichOvertimeApplies: boolAttr(attrs, "reset-hours-after-which-overtime-applies"),
		SellerOperationsContactAssignedAt:   formatDateTime(stringAttr(attrs, "seller-operations-contact-assigned-at")),

		StatusChangedByID:                         relationshipIDFromMap(resource.Relationships, "status-changed-by"),
		TenderJobScheduleShiftTimeCardReviewIDs:   relationshipIDsFromMap(resource.Relationships, "tender-job-schedule-shift-time-card-reviews"),
		AcceptedTruckerID:                         relationshipIDFromMap(resource.Relationships, "accepted-trucker"),
		BrokerTenderID:                            relationshipIDFromMap(resource.Relationships, "broker-tender"),
		LineupTrailerClassificationID:             relationshipIDFromMap(resource.Relationships, "lineup-trailer-classification"),
		DriverAssignmentRefusalIDs:                relationshipIDsFromMap(resource.Relationships, "driver-assignment-refusals"),
		CurrentDriverAssignmentAcknowledgementIDs: relationshipIDsFromMap(resource.Relationships, "current-driver-assignment-acknowledgements"),
		ShiftTimeCardRequisitionIDs:               relationshipIDsFromMap(resource.Relationships, "shift-time-card-requisitions"),
		MaterialPurchaseOrderReleaseIDs:           relationshipIDsFromMap(resource.Relationships, "material-purchase-order-releases"),
		SiteEventIDs:                              relationshipIDsFromMap(resource.Relationships, "site-events"),
		JobProductionPlanMaterialTypeIDs:          relationshipIDsFromMap(resource.Relationships, "job-production-plan-material-types"),
		TenderID:                                  relationshipIDFromMap(resource.Relationships, "tender"),
		JobScheduleShiftID:                        relationshipIDFromMap(resource.Relationships, "job-schedule-shift"),
		TrailerID:                                 relationshipIDFromMap(resource.Relationships, "trailer"),
		TractorID:                                 relationshipIDFromMap(resource.Relationships, "tractor"),
		SellerOperationsContactID:                 relationshipIDFromMap(resource.Relationships, "seller-operations-contact"),
		SellerOperationsContactDraftID:            relationshipIDFromMap(resource.Relationships, "seller-operations-contact-draft"),
		TrailerDraftID:                            relationshipIDFromMap(resource.Relationships, "trailer-draft"),
		RetainerID:                                relationshipIDFromMap(resource.Relationships, "retainer"),
		TimeCardIDs:                               relationshipIDsFromMap(resource.Relationships, "time-cards"),
		TimeCardPreApprovalID:                     relationshipIDFromMap(resource.Relationships, "time-card-pre-approval"),
		ServiceEventIDs:                           relationshipIDsFromMap(resource.Relationships, "service-events"),
		ReadyToWorkID:                             relationshipIDFromMap(resource.Relationships, "ready-to-work"),
		AcceptedBrokerTenderJobScheduleShiftID:    relationshipIDFromMap(resource.Relationships, "accepted-broker-tender-job-schedule-shift"),
		ExpectedTimeOfArrivalIDs:                  relationshipIDsFromMap(resource.Relationships, "expected-time-of-arrivals"),
		ShiftFeedbackIDs:                          relationshipIDsFromMap(resource.Relationships, "shift-feedbacks"),
		ProductionIncidentIDs:                     relationshipIDsFromMap(resource.Relationships, "production-incidents"),
		JobProductionPlanBroadcastMessageIDs:      relationshipIDsFromMap(resource.Relationships, "job-production-plan-broadcast-messages"),
		TripIDs:                                   relationshipIDsFromMap(resource.Relationships, "trips"),
		MaterialTransactionIDs:                    relationshipIDsFromMap(resource.Relationships, "material-transactions"),
		TruckerShiftSetID:                         relationshipIDFromMap(resource.Relationships, "trucker-shift-set"),
		ShiftDriverIDs:                            relationshipIDsFromMap(resource.Relationships, "shift-drivers"),
		PrimaryDriverID:                           relationshipIDFromMap(resource.Relationships, "primary-driver"),
	}

	if details.TenderType == "" {
		if rel, ok := resource.Relationships["tender"]; ok && rel.Data != nil {
			details.TenderType = rel.Data.Type
			details.TenderRelationshipType = rel.Data.Type
		}
	} else {
		details.TenderRelationshipType = details.TenderType
	}

	return details
}

func renderTenderJobScheduleShiftDetails(cmd *cobra.Command, details tenderJobScheduleShiftDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckCount != "" {
		fmt.Fprintf(out, "Truck Count: %s\n", details.TruckCount)
	}
	if details.NotifyBeforeShiftStartsHours != "" {
		fmt.Fprintf(out, "Notify Before Shift Starts Hours: %s\n", details.NotifyBeforeShiftStartsHours)
	}
	if details.NotifyAfterShiftEndsHours != "" {
		fmt.Fprintf(out, "Notify After Shift Ends Hours: %s\n", details.NotifyAfterShiftEndsHours)
	}
	fmt.Fprintf(out, "Notify Driver On Late Shift Assignment: %t\n", details.NotifyDriverOnLateShiftAssignment)
	fmt.Fprintf(out, "Explicit Notify Driver When GPS Not Available: %t\n", details.ExplicitNotifyDriverWhenGPSNotAvailable)
	fmt.Fprintf(out, "Notify Driver When GPS Not Available: %t\n", details.NotifyDriverWhenGPSNotAvailable)
	fmt.Fprintf(out, "Automated Job Site Time Creation Disabled: %t\n", details.IsAutomatedJobSiteTimeCreationDisabled)
	if details.IsTimeCardPayrollCertificationRequiredExplicit != "" {
		fmt.Fprintf(out, "Is Time Card Payroll Certification Required Explicit: %s\n", details.IsTimeCardPayrollCertificationRequiredExplicit)
	}
	if details.IsTimeCardCreatingTimeSheetLineItemExplicit != "" {
		fmt.Fprintf(out, "Is Time Card Creating Time Sheet Line Item Explicit: %s\n", details.IsTimeCardCreatingTimeSheetLineItemExplicit)
	}
	fmt.Fprintf(out, "Skip Validate Driver Assignment Rule Evaluation: %t\n", details.SkipValidateDriverAssignmentRuleEvaluation)
	if details.DriverAssignmentRuleOverrideReason != "" {
		fmt.Fprintf(out, "Driver Assignment Rule Override Reason: %s\n", details.DriverAssignmentRuleOverrideReason)
	}
	fmt.Fprintf(out, "Skip Material Transaction Image Extraction: %t\n", details.SkipMaterialTransactionImageExtraction)
	fmt.Fprintf(out, "Disable Pre-Start Notifications: %t\n", details.DisablePreStartNotifications)
	if details.CancelledAt != "" {
		fmt.Fprintf(out, "Cancelled At: %s\n", details.CancelledAt)
	}
	if details.RejectedAt != "" {
		fmt.Fprintf(out, "Rejected At: %s\n", details.RejectedAt)
	}
	if details.ReturnedAt != "" {
		fmt.Fprintf(out, "Returned At: %s\n", details.ReturnedAt)
	}
	if details.StatusChangeComment != "" {
		fmt.Fprintf(out, "Status Change Comment: %s\n", details.StatusChangeComment)
	}
	fmt.Fprintf(out, "All Trips Entered: %t\n", details.AllTripsEntered)
	if details.HoursAfterWhichOvertimeApplies != "" {
		fmt.Fprintf(out, "Hours After Which Overtime Applies: %s\n", details.HoursAfterWhichOvertimeApplies)
	}
	if details.TravelMiles != "" {
		fmt.Fprintf(out, "Travel Miles: %s\n", details.TravelMiles)
	}
	if details.TravelMinutes != "" {
		fmt.Fprintf(out, "Travel Minutes: %s\n", details.TravelMinutes)
	}
	if details.BillableTravelMinutes != "" {
		fmt.Fprintf(out, "Billable Travel Minutes: %s\n", details.BillableTravelMinutes)
	}
	if details.LoadedTonsMax != "" {
		fmt.Fprintf(out, "Loaded Tons Max: %s\n", details.LoadedTonsMax)
	}
	if details.LoadedTonsMaxEffective != "" {
		fmt.Fprintf(out, "Loaded Tons Max Effective: %s\n", details.LoadedTonsMaxEffective)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.GrossWeightLegalLimitLbsExplicit != "" {
		fmt.Fprintf(out, "Gross Weight Legal Limit Lbs Explicit: %s\n", details.GrossWeightLegalLimitLbsExplicit)
	}
	fmt.Fprintf(out, "Auto Check In Driver On Arrival At Start Site: %t\n", details.AutoCheckInDriverOnArrivalAtStartSite)
	fmt.Fprintf(out, "Explicit Is Expecting Time Card: %t\n", details.ExplicitIsExpectingTimeCard)
	fmt.Fprintf(out, "Is Expecting Time Card: %t\n", details.IsExpectingTimeCard)
	if details.MaterialTransactionTonsMax != "" {
		fmt.Fprintf(out, "Material Transaction Tons Max: %s\n", details.MaterialTransactionTonsMax)
	}
	if details.ExplicitMaterialTransactionTonsMax != "" {
		fmt.Fprintf(out, "Explicit Material Transaction Tons Max: %s\n", details.ExplicitMaterialTransactionTonsMax)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone ID: %s\n", details.TimeZoneID)
	}
	fmt.Fprintf(out, "Suppress Automated Shift Feedback: %t\n", details.SuppressAutomatedShiftFeedback)
	if details.GrossWeightLegalLimitLbs != "" {
		fmt.Fprintf(out, "Gross Weight Legal Limit Lbs: %s\n", details.GrossWeightLegalLimitLbs)
	}
	if details.FirstMaterialTransactionLoadedAt != "" {
		fmt.Fprintf(out, "First Material Transaction Loaded At: %s\n", details.FirstMaterialTransactionLoadedAt)
	}
	if details.ImpliedTimeCardStartAt != "" {
		fmt.Fprintf(out, "Implied Time Card Start At: %s\n", details.ImpliedTimeCardStartAt)
	}
	if details.TenderType != "" {
		fmt.Fprintf(out, "Tender Type: %s\n", details.TenderType)
	}
	if details.TrackedPct != "" {
		fmt.Fprintf(out, "Tracked Percent: %s\n", details.TrackedPct)
	}
	fmt.Fprintf(out, "Is Expecting Material Transactions: %t\n", details.IsExpectingMaterialTransactions)
	if details.ExpectingMaterialTransactionsMessage != "" {
		fmt.Fprintf(out, "Expecting Material Transactions Message: %s\n", details.ExpectingMaterialTransactionsMessage)
	}
	fmt.Fprintf(out, "Current User Can Create Driver Assignment Refusal: %t\n", details.CurrentUserCanCreateDriverAssignmentRefusal)
	fmt.Fprintf(out, "Can Driver Assignment Be Refused: %t\n", details.CanDriverAssignmentBeRefused)
	if details.DriverDaySequenceIndex != "" {
		fmt.Fprintf(out, "Driver Day Sequence Index: %s\n", details.DriverDaySequenceIndex)
	}
	if details.MaterialTransactionStatus != "" {
		fmt.Fprintf(out, "Material Transaction Status: %s\n", details.MaterialTransactionStatus)
	}
	fmt.Fprintf(out, "Current User Can Create Material Transactions: %t\n", details.CurrentUserCanCreateMaterialTransactions)
	fmt.Fprintf(out, "Current User Can Change Truck: %t\n", details.CurrentUserCanChangeTruck)
	fmt.Fprintf(out, "Is Non-Driver Permitted To Check In: %t\n", details.IsNonDriverPermittedToCheckIn)
	fmt.Fprintf(out, "Trucker Can Create Material Transactions: %t\n", details.TruckerCanCreateMaterialTransactions)
	if details.TruckerCanCreateMaterialTransactionsDisabledByMaxTimeAt != "" {
		fmt.Fprintf(out, "Trucker Can Create Material Transactions Disabled By Max Time At: %s\n", details.TruckerCanCreateMaterialTransactionsDisabledByMaxTimeAt)
	}
	fmt.Fprintf(out, "Managed: %t\n", details.Managed)
	fmt.Fprintf(out, "Is Managed: %t\n", details.IsManaged)
	fmt.Fprintf(out, "Reset Hours After Which Overtime Applies: %t\n", details.ResetHoursAfterWhichOvertimeApplies)
	if details.SellerOperationsContactAssignedAt != "" {
		fmt.Fprintf(out, "Seller Operations Contact Assigned At: %s\n", details.SellerOperationsContactAssignedAt)
	}

	if details.StatusChangedByID != "" {
		fmt.Fprintf(out, "Status Changed By ID: %s\n", details.StatusChangedByID)
	}
	if len(details.TenderJobScheduleShiftTimeCardReviewIDs) > 0 {
		fmt.Fprintf(out, "Tender Job Schedule Shift Time Card Review IDs: %s\n", strings.Join(details.TenderJobScheduleShiftTimeCardReviewIDs, ", "))
	}
	if details.AcceptedTruckerID != "" {
		fmt.Fprintf(out, "Accepted Trucker ID: %s\n", details.AcceptedTruckerID)
	}
	if details.BrokerTenderID != "" {
		fmt.Fprintf(out, "Broker Tender ID: %s\n", details.BrokerTenderID)
	}
	if details.LineupTrailerClassificationID != "" {
		fmt.Fprintf(out, "Lineup Trailer Classification ID: %s\n", details.LineupTrailerClassificationID)
	}
	if len(details.DriverAssignmentRefusalIDs) > 0 {
		fmt.Fprintf(out, "Driver Assignment Refusal IDs: %s\n", strings.Join(details.DriverAssignmentRefusalIDs, ", "))
	}
	if len(details.CurrentDriverAssignmentAcknowledgementIDs) > 0 {
		fmt.Fprintf(out, "Current Driver Assignment Acknowledgement IDs: %s\n", strings.Join(details.CurrentDriverAssignmentAcknowledgementIDs, ", "))
	}
	if len(details.ShiftTimeCardRequisitionIDs) > 0 {
		fmt.Fprintf(out, "Shift Time Card Requisition IDs: %s\n", strings.Join(details.ShiftTimeCardRequisitionIDs, ", "))
	}
	if len(details.MaterialPurchaseOrderReleaseIDs) > 0 {
		fmt.Fprintf(out, "Material Purchase Order Release IDs: %s\n", strings.Join(details.MaterialPurchaseOrderReleaseIDs, ", "))
	}
	if len(details.SiteEventIDs) > 0 {
		fmt.Fprintf(out, "Site Event IDs: %s\n", strings.Join(details.SiteEventIDs, ", "))
	}
	if len(details.JobProductionPlanMaterialTypeIDs) > 0 {
		fmt.Fprintf(out, "Job Production Plan Material Type IDs: %s\n", strings.Join(details.JobProductionPlanMaterialTypeIDs, ", "))
	}
	if details.TenderID != "" {
		fmt.Fprintf(out, "Tender ID: %s\n", details.TenderID)
	}
	if details.TenderRelationshipType != "" {
		fmt.Fprintf(out, "Tender Relationship Type: %s\n", details.TenderRelationshipType)
	}
	if details.JobScheduleShiftID != "" {
		fmt.Fprintf(out, "Job Schedule Shift ID: %s\n", details.JobScheduleShiftID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer ID: %s\n", details.TrailerID)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor ID: %s\n", details.TractorID)
	}
	if details.SellerOperationsContactID != "" {
		fmt.Fprintf(out, "Seller Operations Contact ID: %s\n", details.SellerOperationsContactID)
	}
	if details.SellerOperationsContactDraftID != "" {
		fmt.Fprintf(out, "Seller Operations Contact Draft ID: %s\n", details.SellerOperationsContactDraftID)
	}
	if details.TrailerDraftID != "" {
		fmt.Fprintf(out, "Trailer Draft ID: %s\n", details.TrailerDraftID)
	}
	if details.RetainerID != "" {
		fmt.Fprintf(out, "Retainer ID: %s\n", details.RetainerID)
	}
	if len(details.TimeCardIDs) > 0 {
		fmt.Fprintf(out, "Time Card IDs: %s\n", strings.Join(details.TimeCardIDs, ", "))
	}
	if details.TimeCardPreApprovalID != "" {
		fmt.Fprintf(out, "Time Card Pre-Approval ID: %s\n", details.TimeCardPreApprovalID)
	}
	if len(details.ServiceEventIDs) > 0 {
		fmt.Fprintf(out, "Service Event IDs: %s\n", strings.Join(details.ServiceEventIDs, ", "))
	}
	if details.ReadyToWorkID != "" {
		fmt.Fprintf(out, "Ready To Work ID: %s\n", details.ReadyToWorkID)
	}
	if details.AcceptedBrokerTenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Accepted Broker Tender Job Schedule Shift ID: %s\n", details.AcceptedBrokerTenderJobScheduleShiftID)
	}
	if len(details.ExpectedTimeOfArrivalIDs) > 0 {
		fmt.Fprintf(out, "Expected Time Of Arrival IDs: %s\n", strings.Join(details.ExpectedTimeOfArrivalIDs, ", "))
	}
	if len(details.ShiftFeedbackIDs) > 0 {
		fmt.Fprintf(out, "Shift Feedback IDs: %s\n", strings.Join(details.ShiftFeedbackIDs, ", "))
	}
	if len(details.ProductionIncidentIDs) > 0 {
		fmt.Fprintf(out, "Production Incident IDs: %s\n", strings.Join(details.ProductionIncidentIDs, ", "))
	}
	if len(details.JobProductionPlanBroadcastMessageIDs) > 0 {
		fmt.Fprintf(out, "Job Production Plan Broadcast Message IDs: %s\n", strings.Join(details.JobProductionPlanBroadcastMessageIDs, ", "))
	}
	if len(details.TripIDs) > 0 {
		fmt.Fprintf(out, "Trip IDs: %s\n", strings.Join(details.TripIDs, ", "))
	}
	if len(details.MaterialTransactionIDs) > 0 {
		fmt.Fprintf(out, "Material Transaction IDs: %s\n", strings.Join(details.MaterialTransactionIDs, ", "))
	}
	if details.TruckerShiftSetID != "" {
		fmt.Fprintf(out, "Trucker Shift Set ID: %s\n", details.TruckerShiftSetID)
	}
	if len(details.ShiftDriverIDs) > 0 {
		fmt.Fprintf(out, "Shift Driver IDs: %s\n", strings.Join(details.ShiftDriverIDs, ", "))
	}
	if details.PrimaryDriverID != "" {
		fmt.Fprintf(out, "Primary Driver ID: %s\n", details.PrimaryDriverID)
	}
	if details.Rates != nil {
		fmt.Fprintln(out, "Rates:")
		if err := writeJSON(out, details.Rates); err != nil {
			return err
		}
	}

	return nil
}
