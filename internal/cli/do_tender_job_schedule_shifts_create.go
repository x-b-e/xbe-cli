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

type doTenderJobScheduleShiftsCreateOptions struct {
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

func newDoTenderJobScheduleShiftsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tender job schedule shift",
		Long: `Create a tender job schedule shift.

Required flags:
  --tender-type                  Tender resource type (required, e.g., broker-tenders)
  --tender-id                    Tender ID (required)
  --job-schedule-shift           Job schedule shift ID (required)
  --material-transaction-status  Material transaction status (required, e.g., open)

Optional attributes and relationships are available for driver, equipment, and scheduling fields.`,
		Example: `  # Create a tender job schedule shift
  xbe do tender-job-schedule-shifts create \
    --tender-type broker-tenders \
    --tender-id 123 \
    --job-schedule-shift 456 \
    --material-transaction-status open

  # Create with driver and trailer
  xbe do tender-job-schedule-shifts create \
    --tender-type broker-tenders \
    --tender-id 123 \
    --job-schedule-shift 456 \
    --material-transaction-status open \
    --seller-operations-contact 789 \
    --trailer 321`,
		Args: cobra.NoArgs,
		RunE: runDoTenderJobScheduleShiftsCreate,
	}
	initDoTenderJobScheduleShiftsCreateFlags(cmd)
	return cmd
}

func init() {
	doTenderJobScheduleShiftsCmd.AddCommand(newDoTenderJobScheduleShiftsCreateCmd())
}

func initDoTenderJobScheduleShiftsCreateFlags(cmd *cobra.Command) {
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
	cmd.Flags().String("material-transaction-status", "", "Material transaction status (required)")
	cmd.Flags().Bool("trucker-can-create-material-transactions", false, "Trucker can create material transactions")
	cmd.Flags().Bool("reset-hours-after-which-overtime-applies", false, "Reset hours after which overtime applies")

	cmd.Flags().String("tender-type", "", "Tender resource type (required, e.g., broker-tenders)")
	cmd.Flags().String("tender-id", "", "Tender ID (required)")
	cmd.Flags().String("job-schedule-shift", "", "Job schedule shift ID (required)")
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

func runDoTenderJobScheduleShiftsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTenderJobScheduleShiftsCreateOptions(cmd)
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

	if opts.TenderType == "" {
		err := fmt.Errorf("--tender-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.TenderID == "" {
		err := fmt.Errorf("--tender-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.JobScheduleShift == "" {
		err := fmt.Errorf("--job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.MaterialTransactionStatus == "" {
		err := fmt.Errorf("--material-transaction-status is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.TruckCount != "" {
		attributes["truck-count"] = opts.TruckCount
	}
	if opts.NotifyBeforeShiftStartsHours != "" {
		attributes["notify-before-shift-starts-hours"] = opts.NotifyBeforeShiftStartsHours
	}
	if opts.NotifyAfterShiftEndsHours != "" {
		attributes["notify-after-shift-ends-hours"] = opts.NotifyAfterShiftEndsHours
	}
	if cmd.Flags().Changed("notify-driver-on-late-shift-assignment") {
		attributes["notify-driver-on-late-shift-assignment"] = opts.NotifyDriverOnLateShiftAssignment
	}
	if cmd.Flags().Changed("explicit-notify-driver-when-gps-not-available") {
		attributes["explicit-notify-driver-when-gps-not-available"] = opts.ExplicitNotifyDriverWhenGPSNotAvailable
	}
	if cmd.Flags().Changed("notify-driver-when-gps-not-available") {
		attributes["notify-driver-when-gps-not-available"] = opts.NotifyDriverWhenGPSNotAvailable
	}
	if cmd.Flags().Changed("is-automated-job-site-time-creation-disabled") {
		attributes["is-automated-job-site-time-creation-disabled"] = opts.IsAutomatedJobSiteTimeCreationDisabled
	}
	if cmd.Flags().Changed("is-time-card-payroll-certification-required-explicit") {
		if opts.IsTimeCardPayrollCertificationRequiredExplicit == "null" {
			attributes["is-time-card-payroll-certification-required-explicit"] = nil
		} else {
			attributes["is-time-card-payroll-certification-required-explicit"] = opts.IsTimeCardPayrollCertificationRequiredExplicit == "true"
		}
	}
	if cmd.Flags().Changed("is-time-card-creating-time-sheet-line-item-explicit") {
		if opts.IsTimeCardCreatingTimeSheetLineItemExplicit == "null" {
			attributes["is-time-card-creating-time-sheet-line-item-explicit"] = nil
		} else {
			attributes["is-time-card-creating-time-sheet-line-item-explicit"] = opts.IsTimeCardCreatingTimeSheetLineItemExplicit == "true"
		}
	}
	if cmd.Flags().Changed("skip-validate-driver-assignment-rule-evaluation") {
		attributes["skip-validate-driver-assignment-rule-evaluation"] = opts.SkipValidateDriverAssignmentRuleEvaluation
	}
	if opts.DriverAssignmentRuleOverrideReason != "" {
		attributes["driver-assignment-rule-override-reason"] = opts.DriverAssignmentRuleOverrideReason
	}
	if cmd.Flags().Changed("disable-pre-start-notifications") {
		attributes["disable-pre-start-notifications"] = opts.DisablePreStartNotifications
	}
	if cmd.Flags().Changed("all-trips-entered") {
		attributes["all-trips-entered"] = opts.AllTripsEntered
	}
	if opts.HoursAfterWhichOvertimeApplies != "" {
		attributes["hours-after-which-overtime-applies"] = opts.HoursAfterWhichOvertimeApplies
	}
	if opts.TravelMiles != "" {
		attributes["travel-miles"] = opts.TravelMiles
	}
	if opts.BillableTravelMinutes != "" {
		attributes["billable-travel-minutes"] = opts.BillableTravelMinutes
	}
	if opts.LoadedTonsMax != "" {
		attributes["loaded-tons-max"] = opts.LoadedTonsMax
	}
	if opts.StartAt != "" {
		attributes["start-at"] = opts.StartAt
	}
	if opts.GrossWeightLegalLimitLbsExplicit != "" {
		attributes["gross-weight-legal-limit-lbs-explicit"] = opts.GrossWeightLegalLimitLbsExplicit
	}
	if cmd.Flags().Changed("auto-check-in-driver-on-arrival-at-start-site") {
		attributes["auto-check-in-driver-on-arrival-at-start-site"] = opts.AutoCheckInDriverOnArrivalAtStartSite
	}
	if cmd.Flags().Changed("explicit-is-expecting-time-card") {
		attributes["explicit-is-expecting-time-card"] = opts.ExplicitIsExpectingTimeCard
	}
	if opts.ExplicitMaterialTransactionTonsMax != "" {
		attributes["explicit-material-transaction-tons-max"] = opts.ExplicitMaterialTransactionTonsMax
	}
	if cmd.Flags().Changed("is-expecting-material-transactions") {
		attributes["is-expecting-material-transactions"] = opts.IsExpectingMaterialTransactions
	}
	if opts.ExpectingMaterialTransactionsMessage != "" {
		attributes["expecting-material-transactions-message"] = opts.ExpectingMaterialTransactionsMessage
	}
	attributes["material-transaction-status"] = opts.MaterialTransactionStatus
	if cmd.Flags().Changed("trucker-can-create-material-transactions") {
		attributes["trucker-can-create-material-transactions"] = opts.TruckerCanCreateMaterialTransactions
	}
	if cmd.Flags().Changed("reset-hours-after-which-overtime-applies") {
		attributes["reset-hours-after-which-overtime-applies"] = opts.ResetHoursAfterWhichOvertimeApplies
	}

	relationships := map[string]any{
		"tender": map[string]any{
			"data": map[string]any{
				"type": opts.TenderType,
				"id":   opts.TenderID,
			},
		},
		"job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "job-schedule-shifts",
				"id":   opts.JobScheduleShift,
			},
		},
	}

	if opts.Trailer != "" {
		relationships["trailer"] = map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.Trailer,
			},
		}
	}
	if opts.Tractor != "" {
		relationships["tractor"] = map[string]any{
			"data": map[string]any{
				"type": "tractors",
				"id":   opts.Tractor,
			},
		}
	}
	if opts.SellerOperationsContact != "" {
		relationships["seller-operations-contact"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.SellerOperationsContact,
			},
		}
	}
	if opts.SellerOperationsContactDraft != "" {
		relationships["seller-operations-contact-draft"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.SellerOperationsContactDraft,
			},
		}
	}
	if opts.TrailerDraft != "" {
		relationships["trailer-draft"] = map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.TrailerDraft,
			},
		}
	}
	if opts.Retainer != "" {
		relationships["retainer"] = map[string]any{
			"data": map[string]any{
				"type": "retainers",
				"id":   opts.Retainer,
			},
		}
	}

	addRelationshipIDs := func(key, resourceType, raw string) {
		ids := splitCommaList(raw)
		if len(ids) == 0 {
			return
		}
		data := make([]map[string]any, 0, len(ids))
		for _, id := range ids {
			data = append(data, map[string]any{
				"type": resourceType,
				"id":   id,
			})
		}
		relationships[key] = map[string]any{"data": data}
	}

	addRelationshipIDs("time-cards", "time-cards", opts.TimeCards)
	addRelationshipIDs("service-events", "service-events", opts.ServiceEvents)
	addRelationshipIDs("expected-time-of-arrivals", "expected-time-of-arrivals", opts.ExpectedTimeOfArrivals)
	addRelationshipIDs("shift-feedbacks", "shift-feedbacks", opts.ShiftFeedbacks)
	addRelationshipIDs("production-incidents", "production-incidents", opts.ProductionIncidents)
	addRelationshipIDs("job-production-plan-broadcast-messages", "job-production-plan-broadcast-messages", opts.JobProductionPlanBroadcastMessages)
	addRelationshipIDs("trips", "trips", opts.Trips)

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tender-job-schedule-shifts",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tender-job-schedule-shifts", jsonBody)
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

	row := buildTenderJobScheduleShiftRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tender job schedule shift %s\n", row.ID)
	return nil
}

func parseDoTenderJobScheduleShiftsCreateOptions(cmd *cobra.Command) (doTenderJobScheduleShiftsCreateOptions, error) {
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

	return doTenderJobScheduleShiftsCreateOptions{
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
