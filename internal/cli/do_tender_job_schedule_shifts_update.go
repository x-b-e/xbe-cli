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

type doTenderJobScheduleShiftsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	TruckCount                                     string
	NotifyBeforeShiftStartsHours                   string
	NotifyAfterShiftEndsHours                      string
	NotifyDriverOnLateShiftAssignment              bool
	ExplicitNotifyDriverWhenGPSNotAvailable        bool
	NotifyDriverWhenGPSNotAvailable                bool
	IsAutomatedJobSiteTimeCreationDisabled         bool
	IsTimeCardPayrollCertificationRequiredExplicit string
	IsTimeCardCreatingTimeSheetLineItemExplicit    string
	SkipValidateDriverAssignmentRuleEvaluation     bool
	DriverAssignmentRuleOverrideReason             string
	DisablePreStartNotifications                   bool
	AllTripsEntered                                bool
	HoursAfterWhichOvertimeApplies                 string
	TravelMiles                                    string
	BillableTravelMinutes                          string
	LoadedTonsMax                                  string
	StartAt                                        string
	GrossWeightLegalLimitLbsExplicit               string
	AutoCheckInDriverOnArrivalAtStartSite          bool
	ExplicitIsExpectingTimeCard                    bool
	ExplicitMaterialTransactionTonsMax             string
	IsExpectingMaterialTransactions                bool
	ExpectingMaterialTransactionsMessage           string
	MaterialTransactionStatus                      string
	TruckerCanCreateMaterialTransactions           bool
	ResetHoursAfterWhichOvertimeApplies            bool

	TenderType                         string
	TenderID                           string
	JobScheduleShift                   string
	Trailer                            string
	Tractor                            string
	SellerOperationsContact            string
	SellerOperationsContactDraft       string
	TrailerDraft                       string
	Retainer                           string
	TimeCards                          string
	ServiceEvents                      string
	ExpectedTimeOfArrivals             string
	ShiftFeedbacks                     string
	ProductionIncidents                string
	JobProductionPlanBroadcastMessages string
	Trips                              string
}

func newDoTenderJobScheduleShiftsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tender job schedule shift",
		Long: `Update a tender job schedule shift.

Specify any attribute or relationship flags to update.

Arguments:
  <id>    The tender job schedule shift ID (required).`,
		Example: `  # Update driver assignment
  xbe do tender-job-schedule-shifts update 123 --seller-operations-contact 456

  # Update start time
  xbe do tender-job-schedule-shifts update 123 --start-at "2025-01-01T08:00:00Z"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTenderJobScheduleShiftsUpdate,
	}
	initDoTenderJobScheduleShiftsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftsCmd.AddCommand(newDoTenderJobScheduleShiftsUpdateCmd())
}

func initDoTenderJobScheduleShiftsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")

	cmd.Flags().String("truck-count", "", "Truck count override")
	cmd.Flags().String("notify-before-shift-starts-hours", "", "Notify driver before shift starts (hours)")
	cmd.Flags().String("notify-after-shift-ends-hours", "", "Notify driver after shift ends (hours)")
	cmd.Flags().Bool("notify-driver-on-late-shift-assignment", false, "Notify driver on late shift assignment")
	cmd.Flags().Bool("explicit-notify-driver-when-gps-not-available", false, "Explicitly notify driver when GPS not available")
	cmd.Flags().Bool("notify-driver-when-gps-not-available", false, "Notify driver when GPS not available (legacy)")
	cmd.Flags().Bool("is-automated-job-site-time-creation-disabled", false, "Disable automated job site time creation")
	cmd.Flags().String("is-time-card-payroll-certification-required-explicit", "", "Is time card payroll certification required explicit (true/false/null)")
	cmd.Flags().String("is-time-card-creating-time-sheet-line-item-explicit", "", "Is time card creating time sheet line item explicit (true/false/null)")
	cmd.Flags().Bool("skip-validate-driver-assignment-rule-evaluation", false, "Skip driver assignment rule evaluation")
	cmd.Flags().String("driver-assignment-rule-override-reason", "", "Driver assignment rule override reason")
	cmd.Flags().Bool("disable-pre-start-notifications", false, "Disable pre-start notifications")
	cmd.Flags().Bool("all-trips-entered", false, "All trips entered")
	cmd.Flags().String("hours-after-which-overtime-applies", "", "Hours after which overtime applies")
	cmd.Flags().String("travel-miles", "", "Travel miles")
	cmd.Flags().String("billable-travel-minutes", "", "Billable travel minutes")
	cmd.Flags().String("loaded-tons-max", "", "Loaded tons max")
	cmd.Flags().String("start-at", "", "Shift start time (RFC3339)")
	cmd.Flags().String("gross-weight-legal-limit-lbs-explicit", "", "Gross weight legal limit lbs explicit")
	cmd.Flags().Bool("auto-check-in-driver-on-arrival-at-start-site", false, "Auto check-in driver on arrival at start site")
	cmd.Flags().Bool("explicit-is-expecting-time-card", false, "Explicit is expecting time card")
	cmd.Flags().String("explicit-material-transaction-tons-max", "", "Explicit material transaction tons max")
	cmd.Flags().Bool("is-expecting-material-transactions", false, "Is expecting material transactions")
	cmd.Flags().String("expecting-material-transactions-message", "", "Expecting material transactions message")
	cmd.Flags().String("material-transaction-status", "", "Material transaction status")
	cmd.Flags().Bool("trucker-can-create-material-transactions", false, "Trucker can create material transactions")
	cmd.Flags().Bool("reset-hours-after-which-overtime-applies", false, "Reset hours after which overtime applies")

	cmd.Flags().String("tender-type", "", "Tender resource type (e.g., broker-tenders)")
	cmd.Flags().String("tender-id", "", "Tender ID")
	cmd.Flags().String("job-schedule-shift", "", "Job schedule shift ID")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("seller-operations-contact", "", "Seller operations contact (driver) ID")
	cmd.Flags().String("seller-operations-contact-draft", "", "Seller operations contact draft ID")
	cmd.Flags().String("trailer-draft", "", "Trailer draft ID")
	cmd.Flags().String("retainer", "", "Retainer ID")
	cmd.Flags().String("time-cards", "", "Time card IDs (comma-separated)")
	cmd.Flags().String("service-events", "", "Service event IDs (comma-separated)")
	cmd.Flags().String("expected-time-of-arrivals", "", "Expected time of arrival IDs (comma-separated)")
	cmd.Flags().String("shift-feedbacks", "", "Shift feedback IDs (comma-separated)")
	cmd.Flags().String("production-incidents", "", "Production incident IDs (comma-separated)")
	cmd.Flags().String("job-production-plan-broadcast-messages", "", "Job production plan broadcast message IDs (comma-separated)")
	cmd.Flags().String("trips", "", "Trip IDs (comma-separated)")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTenderJobScheduleShiftsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTenderJobScheduleShiftsUpdateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("tender job schedule shift id is required")
	}

	attributes := map[string]any{}
	relationships := map[string]any{}
	hasChanges := false

	if cmd.Flags().Changed("truck-count") {
		attributes["truck-count"] = opts.TruckCount
		hasChanges = true
	}
	if cmd.Flags().Changed("notify-before-shift-starts-hours") {
		attributes["notify-before-shift-starts-hours"] = opts.NotifyBeforeShiftStartsHours
		hasChanges = true
	}
	if cmd.Flags().Changed("notify-after-shift-ends-hours") {
		attributes["notify-after-shift-ends-hours"] = opts.NotifyAfterShiftEndsHours
		hasChanges = true
	}
	if cmd.Flags().Changed("notify-driver-on-late-shift-assignment") {
		attributes["notify-driver-on-late-shift-assignment"] = opts.NotifyDriverOnLateShiftAssignment
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-notify-driver-when-gps-not-available") {
		attributes["explicit-notify-driver-when-gps-not-available"] = opts.ExplicitNotifyDriverWhenGPSNotAvailable
		hasChanges = true
	}
	if cmd.Flags().Changed("notify-driver-when-gps-not-available") {
		attributes["notify-driver-when-gps-not-available"] = opts.NotifyDriverWhenGPSNotAvailable
		hasChanges = true
	}
	if cmd.Flags().Changed("is-automated-job-site-time-creation-disabled") {
		attributes["is-automated-job-site-time-creation-disabled"] = opts.IsAutomatedJobSiteTimeCreationDisabled
		hasChanges = true
	}
	if cmd.Flags().Changed("is-time-card-payroll-certification-required-explicit") {
		if opts.IsTimeCardPayrollCertificationRequiredExplicit == "null" {
			attributes["is-time-card-payroll-certification-required-explicit"] = nil
		} else {
			attributes["is-time-card-payroll-certification-required-explicit"] = opts.IsTimeCardPayrollCertificationRequiredExplicit == "true"
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("is-time-card-creating-time-sheet-line-item-explicit") {
		if opts.IsTimeCardCreatingTimeSheetLineItemExplicit == "null" {
			attributes["is-time-card-creating-time-sheet-line-item-explicit"] = nil
		} else {
			attributes["is-time-card-creating-time-sheet-line-item-explicit"] = opts.IsTimeCardCreatingTimeSheetLineItemExplicit == "true"
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("skip-validate-driver-assignment-rule-evaluation") {
		attributes["skip-validate-driver-assignment-rule-evaluation"] = opts.SkipValidateDriverAssignmentRuleEvaluation
		hasChanges = true
	}
	if cmd.Flags().Changed("driver-assignment-rule-override-reason") {
		attributes["driver-assignment-rule-override-reason"] = opts.DriverAssignmentRuleOverrideReason
		hasChanges = true
	}
	if cmd.Flags().Changed("disable-pre-start-notifications") {
		attributes["disable-pre-start-notifications"] = opts.DisablePreStartNotifications
		hasChanges = true
	}
	if cmd.Flags().Changed("all-trips-entered") {
		attributes["all-trips-entered"] = opts.AllTripsEntered
		hasChanges = true
	}
	if cmd.Flags().Changed("hours-after-which-overtime-applies") {
		attributes["hours-after-which-overtime-applies"] = opts.HoursAfterWhichOvertimeApplies
		hasChanges = true
	}
	if cmd.Flags().Changed("travel-miles") {
		attributes["travel-miles"] = opts.TravelMiles
		hasChanges = true
	}
	if cmd.Flags().Changed("billable-travel-minutes") {
		attributes["billable-travel-minutes"] = opts.BillableTravelMinutes
		hasChanges = true
	}
	if cmd.Flags().Changed("loaded-tons-max") {
		attributes["loaded-tons-max"] = opts.LoadedTonsMax
		hasChanges = true
	}
	if cmd.Flags().Changed("start-at") {
		attributes["start-at"] = opts.StartAt
		hasChanges = true
	}
	if cmd.Flags().Changed("gross-weight-legal-limit-lbs-explicit") {
		attributes["gross-weight-legal-limit-lbs-explicit"] = opts.GrossWeightLegalLimitLbsExplicit
		hasChanges = true
	}
	if cmd.Flags().Changed("auto-check-in-driver-on-arrival-at-start-site") {
		attributes["auto-check-in-driver-on-arrival-at-start-site"] = opts.AutoCheckInDriverOnArrivalAtStartSite
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-is-expecting-time-card") {
		attributes["explicit-is-expecting-time-card"] = opts.ExplicitIsExpectingTimeCard
		hasChanges = true
	}
	if cmd.Flags().Changed("explicit-material-transaction-tons-max") {
		attributes["explicit-material-transaction-tons-max"] = opts.ExplicitMaterialTransactionTonsMax
		hasChanges = true
	}
	if cmd.Flags().Changed("is-expecting-material-transactions") {
		attributes["is-expecting-material-transactions"] = opts.IsExpectingMaterialTransactions
		hasChanges = true
	}
	if cmd.Flags().Changed("expecting-material-transactions-message") {
		attributes["expecting-material-transactions-message"] = opts.ExpectingMaterialTransactionsMessage
		hasChanges = true
	}
	if cmd.Flags().Changed("material-transaction-status") {
		attributes["material-transaction-status"] = opts.MaterialTransactionStatus
		hasChanges = true
	}
	if cmd.Flags().Changed("trucker-can-create-material-transactions") {
		attributes["trucker-can-create-material-transactions"] = opts.TruckerCanCreateMaterialTransactions
		hasChanges = true
	}
	if cmd.Flags().Changed("reset-hours-after-which-overtime-applies") {
		attributes["reset-hours-after-which-overtime-applies"] = opts.ResetHoursAfterWhichOvertimeApplies
		hasChanges = true
	}

	if cmd.Flags().Changed("tender-type") || cmd.Flags().Changed("tender-id") {
		if opts.TenderType == "" || opts.TenderID == "" {
			return fmt.Errorf("--tender-type and --tender-id are required to update tender")
		}
		relationships["tender"] = map[string]any{
			"data": map[string]any{
				"type": opts.TenderType,
				"id":   opts.TenderID,
			},
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("job-schedule-shift") {
		if opts.JobScheduleShift == "" {
			relationships["job-schedule-shift"] = map[string]any{"data": nil}
		} else {
			relationships["job-schedule-shift"] = map[string]any{
				"data": map[string]any{
					"type": "job-schedule-shifts",
					"id":   opts.JobScheduleShift,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("trailer") {
		if opts.Trailer == "" {
			relationships["trailer"] = map[string]any{"data": nil}
		} else {
			relationships["trailer"] = map[string]any{
				"data": map[string]any{
					"type": "trailers",
					"id":   opts.Trailer,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("tractor") {
		if opts.Tractor == "" {
			relationships["tractor"] = map[string]any{"data": nil}
		} else {
			relationships["tractor"] = map[string]any{
				"data": map[string]any{
					"type": "tractors",
					"id":   opts.Tractor,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("seller-operations-contact") {
		if opts.SellerOperationsContact == "" {
			relationships["seller-operations-contact"] = map[string]any{"data": nil}
		} else {
			relationships["seller-operations-contact"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.SellerOperationsContact,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("seller-operations-contact-draft") {
		if opts.SellerOperationsContactDraft == "" {
			relationships["seller-operations-contact-draft"] = map[string]any{"data": nil}
		} else {
			relationships["seller-operations-contact-draft"] = map[string]any{
				"data": map[string]any{
					"type": "users",
					"id":   opts.SellerOperationsContactDraft,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("trailer-draft") {
		if opts.TrailerDraft == "" {
			relationships["trailer-draft"] = map[string]any{"data": nil}
		} else {
			relationships["trailer-draft"] = map[string]any{
				"data": map[string]any{
					"type": "trailers",
					"id":   opts.TrailerDraft,
				},
			}
		}
		hasChanges = true
	}
	if cmd.Flags().Changed("retainer") {
		if opts.Retainer == "" {
			relationships["retainer"] = map[string]any{"data": nil}
		} else {
			relationships["retainer"] = map[string]any{
				"data": map[string]any{
					"type": "retainers",
					"id":   opts.Retainer,
				},
			}
		}
		hasChanges = true
	}

	setToManyRelationship := func(flagName, key, resourceType, raw string) {
		if !cmd.Flags().Changed(flagName) {
			return
		}
		if strings.TrimSpace(raw) == "" {
			relationships[key] = map[string]any{"data": []any{}}
			hasChanges = true
			return
		}
		ids := splitCommaList(raw)
		data := make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			data = append(data, map[string]any{
				"type": resourceType,
				"id":   id,
			})
		}
		relationships[key] = map[string]any{"data": data}
		hasChanges = true
	}

	setToManyRelationship("time-cards", "time-cards", "time-cards", opts.TimeCards)
	setToManyRelationship("service-events", "service-events", "service-events", opts.ServiceEvents)
	setToManyRelationship("expected-time-of-arrivals", "expected-time-of-arrivals", "expected-time-of-arrivals", opts.ExpectedTimeOfArrivals)
	setToManyRelationship("shift-feedbacks", "shift-feedbacks", "shift-feedbacks", opts.ShiftFeedbacks)
	setToManyRelationship("production-incidents", "production-incidents", "production-incidents", opts.ProductionIncidents)
	setToManyRelationship("job-production-plan-broadcast-messages", "job-production-plan-broadcast-messages", "job-production-plan-broadcast-messages", opts.JobProductionPlanBroadcastMessages)
	setToManyRelationship("trips", "trips", "trips", opts.Trips)

	if !hasChanges {
		return fmt.Errorf("no fields to update")
	}

	data := map[string]any{
		"id":   id,
		"type": "tender-job-schedule-shifts",
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/tender-job-schedule-shifts/"+id, jsonBody)
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

	details := buildTenderJobScheduleShiftDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tender job schedule shift %s\n\n", details.ID)
	return renderTenderJobScheduleShiftDetails(cmd, details)
}

func parseDoTenderJobScheduleShiftsUpdateOptions(cmd *cobra.Command) (doTenderJobScheduleShiftsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	truckCount, _ := cmd.Flags().GetString("truck-count")
	notifyBeforeShiftStartsHours, _ := cmd.Flags().GetString("notify-before-shift-starts-hours")
	notifyAfterShiftEndsHours, _ := cmd.Flags().GetString("notify-after-shift-ends-hours")
	notifyDriverOnLateShiftAssignment, _ := cmd.Flags().GetBool("notify-driver-on-late-shift-assignment")
	explicitNotifyDriverWhenGPSNotAvailable, _ := cmd.Flags().GetBool("explicit-notify-driver-when-gps-not-available")
	notifyDriverWhenGPSNotAvailable, _ := cmd.Flags().GetBool("notify-driver-when-gps-not-available")
	isAutomatedJobSiteTimeCreationDisabled, _ := cmd.Flags().GetBool("is-automated-job-site-time-creation-disabled")
	isTimeCardPayrollCertificationRequiredExplicit, _ := cmd.Flags().GetString("is-time-card-payroll-certification-required-explicit")
	isTimeCardCreatingTimeSheetLineItemExplicit, _ := cmd.Flags().GetString("is-time-card-creating-time-sheet-line-item-explicit")
	skipValidateDriverAssignmentRuleEvaluation, _ := cmd.Flags().GetBool("skip-validate-driver-assignment-rule-evaluation")
	driverAssignmentRuleOverrideReason, _ := cmd.Flags().GetString("driver-assignment-rule-override-reason")
	disablePreStartNotifications, _ := cmd.Flags().GetBool("disable-pre-start-notifications")
	allTripsEntered, _ := cmd.Flags().GetBool("all-trips-entered")
	hoursAfterWhichOvertimeApplies, _ := cmd.Flags().GetString("hours-after-which-overtime-applies")
	travelMiles, _ := cmd.Flags().GetString("travel-miles")
	billableTravelMinutes, _ := cmd.Flags().GetString("billable-travel-minutes")
	loadedTonsMax, _ := cmd.Flags().GetString("loaded-tons-max")
	startAt, _ := cmd.Flags().GetString("start-at")
	grossWeightLegalLimitLbsExplicit, _ := cmd.Flags().GetString("gross-weight-legal-limit-lbs-explicit")
	autoCheckInDriverOnArrivalAtStartSite, _ := cmd.Flags().GetBool("auto-check-in-driver-on-arrival-at-start-site")
	explicitIsExpectingTimeCard, _ := cmd.Flags().GetBool("explicit-is-expecting-time-card")
	explicitMaterialTransactionTonsMax, _ := cmd.Flags().GetString("explicit-material-transaction-tons-max")
	isExpectingMaterialTransactions, _ := cmd.Flags().GetBool("is-expecting-material-transactions")
	expectingMaterialTransactionsMessage, _ := cmd.Flags().GetString("expecting-material-transactions-message")
	materialTransactionStatus, _ := cmd.Flags().GetString("material-transaction-status")
	truckerCanCreateMaterialTransactions, _ := cmd.Flags().GetBool("trucker-can-create-material-transactions")
	resetHoursAfterWhichOvertimeApplies, _ := cmd.Flags().GetBool("reset-hours-after-which-overtime-applies")

	tenderType, _ := cmd.Flags().GetString("tender-type")
	tenderID, _ := cmd.Flags().GetString("tender-id")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractor, _ := cmd.Flags().GetString("tractor")
	sellerOperationsContact, _ := cmd.Flags().GetString("seller-operations-contact")
	sellerOperationsContactDraft, _ := cmd.Flags().GetString("seller-operations-contact-draft")
	trailerDraft, _ := cmd.Flags().GetString("trailer-draft")
	retainer, _ := cmd.Flags().GetString("retainer")
	timeCards, _ := cmd.Flags().GetString("time-cards")
	serviceEvents, _ := cmd.Flags().GetString("service-events")
	expectedTimeOfArrivals, _ := cmd.Flags().GetString("expected-time-of-arrivals")
	shiftFeedbacks, _ := cmd.Flags().GetString("shift-feedbacks")
	productionIncidents, _ := cmd.Flags().GetString("production-incidents")
	jobProductionPlanBroadcastMessages, _ := cmd.Flags().GetString("job-production-plan-broadcast-messages")
	trips, _ := cmd.Flags().GetString("trips")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTenderJobScheduleShiftsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,

		TruckCount:                                     truckCount,
		NotifyBeforeShiftStartsHours:                   notifyBeforeShiftStartsHours,
		NotifyAfterShiftEndsHours:                      notifyAfterShiftEndsHours,
		NotifyDriverOnLateShiftAssignment:              notifyDriverOnLateShiftAssignment,
		ExplicitNotifyDriverWhenGPSNotAvailable:        explicitNotifyDriverWhenGPSNotAvailable,
		NotifyDriverWhenGPSNotAvailable:                notifyDriverWhenGPSNotAvailable,
		IsAutomatedJobSiteTimeCreationDisabled:         isAutomatedJobSiteTimeCreationDisabled,
		IsTimeCardPayrollCertificationRequiredExplicit: isTimeCardPayrollCertificationRequiredExplicit,
		IsTimeCardCreatingTimeSheetLineItemExplicit:    isTimeCardCreatingTimeSheetLineItemExplicit,
		SkipValidateDriverAssignmentRuleEvaluation:     skipValidateDriverAssignmentRuleEvaluation,
		DriverAssignmentRuleOverrideReason:             driverAssignmentRuleOverrideReason,
		DisablePreStartNotifications:                   disablePreStartNotifications,
		AllTripsEntered:                                allTripsEntered,
		HoursAfterWhichOvertimeApplies:                 hoursAfterWhichOvertimeApplies,
		TravelMiles:                                    travelMiles,
		BillableTravelMinutes:                          billableTravelMinutes,
		LoadedTonsMax:                                  loadedTonsMax,
		StartAt:                                        startAt,
		GrossWeightLegalLimitLbsExplicit:               grossWeightLegalLimitLbsExplicit,
		AutoCheckInDriverOnArrivalAtStartSite:          autoCheckInDriverOnArrivalAtStartSite,
		ExplicitIsExpectingTimeCard:                    explicitIsExpectingTimeCard,
		ExplicitMaterialTransactionTonsMax:             explicitMaterialTransactionTonsMax,
		IsExpectingMaterialTransactions:                isExpectingMaterialTransactions,
		ExpectingMaterialTransactionsMessage:           expectingMaterialTransactionsMessage,
		MaterialTransactionStatus:                      materialTransactionStatus,
		TruckerCanCreateMaterialTransactions:           truckerCanCreateMaterialTransactions,
		ResetHoursAfterWhichOvertimeApplies:            resetHoursAfterWhichOvertimeApplies,

		TenderType:                         tenderType,
		TenderID:                           tenderID,
		JobScheduleShift:                   jobScheduleShift,
		Trailer:                            trailer,
		Tractor:                            tractor,
		SellerOperationsContact:            sellerOperationsContact,
		SellerOperationsContactDraft:       sellerOperationsContactDraft,
		TrailerDraft:                       trailerDraft,
		Retainer:                           retainer,
		TimeCards:                          timeCards,
		ServiceEvents:                      serviceEvents,
		ExpectedTimeOfArrivals:             expectedTimeOfArrivals,
		ShiftFeedbacks:                     shiftFeedbacks,
		ProductionIncidents:                productionIncidents,
		JobProductionPlanBroadcastMessages: jobProductionPlanBroadcastMessages,
		Trips:                              trips,
	}, nil
}
