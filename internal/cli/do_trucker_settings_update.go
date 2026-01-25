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

type doTruckerSettingsUpdateOptions struct {
	BaseURL                                               string
	Token                                                 string
	JSON                                                  bool
	BillingPeriodStartOn                                  string
	BillingPeriodDayCount                                 string
	BillingPeriodEndInvoiceOffsetDayCount                 string
	SplitBillingPeriodsSpanningMonths                     string
	GroupDailyInvoiceByJobNumber                          string
	GenerateDailyInvoice                                  string
	DeliverNewInvoices                                    string
	InvoicesBatchProcessingStartOn                        string
	InvoicesGroupedByTimeCardStartDate                    string
	InvoiceDateCalculation                                string
	CreateDetectedProductionIncidents                     string
	PrimaryColor                                          string
	LogoSVG                                               string
	DefaultTenderDispatchInstructions                     string
	DefaultPaymentTermsAndConditions                      string
	DefaultHoursAfterWhichOvertimeApplies                 string
	SetsShiftMaterialTransactionExpectations              string
	ShowPlannerInfoToDrivers                              string
	IsTruckerShiftRejectionPermitted                      string
	NotifyDriverWhenGPSNotAvailable                       string
	DayShiftAssignmentReminderTime                        string
	NightShiftAssignmentReminderTime                      string
	MinimumDriverTrackingMinutes                          string
	AutoGenerateTimeSheetLineItemsPerJob                  string
	RestrictLineItemClassificationEditToTimeSheetApprover string
	DefaultTimeSheetLineItemClassificationID              string
	AutoCombineOverlappingDriverDays                      string
	FlagsSet                                              map[string]bool
}

func newDoTruckerSettingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update trucker settings",
		Long: `Update trucker settings.

Only the fields you specify will be updated. Fields not provided remain unchanged.

Notes:
  Boolean flags expect true/false values.
  Date fields use YYYY-MM-DD. Time fields use HH:MM.`,
		Example: `  # Update GPS notification behavior
  xbe do trucker-settings update 123 --notify-driver-when-gps-not-available true

  # Update shift reminder times
  xbe do trucker-settings update 123 --day-shift-assignment-reminder-time 08:00 --night-shift-assignment-reminder-time 18:00

  # Get JSON output
  xbe do trucker-settings update 123 --notify-driver-when-gps-not-available true --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckerSettingsUpdate,
	}
	initDoTruckerSettingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckerSettingsCmd.AddCommand(newDoTruckerSettingsUpdateCmd())
}

func initDoTruckerSettingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("billing-period-start-on", "", "Billing period start date (YYYY-MM-DD)")
	cmd.Flags().String("billing-period-day-count", "", "Billing period day count")
	cmd.Flags().String("billing-period-end-invoice-offset-day-count", "", "Billing period end invoice offset day count")
	cmd.Flags().String("split-billing-periods-spanning-months", "", "Split billing periods spanning months (true/false)")
	cmd.Flags().String("group-daily-invoice-by-job-number", "", "Group daily invoice by job number (true/false)")
	cmd.Flags().String("generate-daily-invoice", "", "Generate daily invoice (true/false)")
	cmd.Flags().String("deliver-new-invoices", "", "Deliver new invoices (true/false)")
	cmd.Flags().String("invoices-batch-processing-start-on", "", "Invoices batch processing start date (YYYY-MM-DD)")
	cmd.Flags().String("invoices-grouped-by-time-card-start-date", "", "Group invoices by time card start date (true/false)")
	cmd.Flags().String("invoice-date-calculation", "", "Invoice date calculation (average/latest)")
	cmd.Flags().String("create-detected-production-incidents", "", "Create detected production incidents (true/false)")
	cmd.Flags().String("primary-color", "", "Primary color (hex)")
	cmd.Flags().String("logo-svg", "", "Logo SVG content")
	cmd.Flags().String("default-tender-dispatch-instructions", "", "Default tender dispatch instructions")
	cmd.Flags().String("default-payment-terms-and-conditions", "", "Default payment terms and conditions")
	cmd.Flags().String("default-hours-after-which-overtime-applies", "", "Default hours after which overtime applies")
	cmd.Flags().String("sets-shift-material-transaction-expectations", "", "Sets shift material transaction expectations (true/false)")
	cmd.Flags().String("show-planner-info-to-drivers", "", "Show planner info to drivers (true/false)")
	cmd.Flags().String("is-trucker-shift-rejection-permitted", "", "Is trucker shift rejection permitted (true/false)")
	cmd.Flags().String("notify-driver-when-gps-not-available", "", "Notify driver when GPS not available (true/false)")
	cmd.Flags().String("day-shift-assignment-reminder-time", "", "Day shift assignment reminder time (HH:MM)")
	cmd.Flags().String("night-shift-assignment-reminder-time", "", "Night shift assignment reminder time (HH:MM)")
	cmd.Flags().String("minimum-driver-tracking-minutes", "", "Minimum driver tracking minutes")
	cmd.Flags().String("auto-generate-time-sheet-line-items-per-job", "", "Auto-generate time sheet line items per job (true/false)")
	cmd.Flags().String("restrict-line-item-classification-edit-to-time-sheet-approver", "", "Restrict line item edits to time sheet approver (true/false)")
	cmd.Flags().String("default-time-sheet-line-item-classification-id", "", "Default time sheet line item classification ID")
	cmd.Flags().String("auto-combine-overlapping-driver-days", "", "Auto-combine overlapping driver days (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckerSettingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckerSettingsUpdateOptions(cmd)
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
		return fmt.Errorf("trucker settings id is required")
	}

	if !hasAnyFlagSet(opts.FlagsSet) {
		err := fmt.Errorf("at least one field to update is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.FlagsSet["billing-period-start-on"] {
		attributes["billing-period-start-on"] = opts.BillingPeriodStartOn
	}
	if opts.FlagsSet["billing-period-day-count"] {
		var i int
		if _, err := fmt.Sscanf(opts.BillingPeriodDayCount, "%d", &i); err == nil {
			attributes["billing-period-day-count"] = i
		}
	}
	if opts.FlagsSet["billing-period-end-invoice-offset-day-count"] {
		var i int
		if _, err := fmt.Sscanf(opts.BillingPeriodEndInvoiceOffsetDayCount, "%d", &i); err == nil {
			attributes["billing-period-end-invoice-offset-day-count"] = i
		}
	}
	if opts.FlagsSet["split-billing-periods-spanning-months"] {
		attributes["split-billing-periods-spanning-months"] = opts.SplitBillingPeriodsSpanningMonths == "true"
	}
	if opts.FlagsSet["group-daily-invoice-by-job-number"] {
		attributes["group-daily-invoice-by-job-number"] = opts.GroupDailyInvoiceByJobNumber == "true"
	}
	if opts.FlagsSet["generate-daily-invoice"] {
		attributes["generate-daily-invoice"] = opts.GenerateDailyInvoice == "true"
	}
	if opts.FlagsSet["deliver-new-invoices"] {
		attributes["deliver-new-invoices"] = opts.DeliverNewInvoices == "true"
	}
	if opts.FlagsSet["invoices-batch-processing-start-on"] {
		attributes["invoices-batch-processing-start-on"] = opts.InvoicesBatchProcessingStartOn
	}
	if opts.FlagsSet["invoices-grouped-by-time-card-start-date"] {
		attributes["invoices-grouped-by-time-card-start-date"] = opts.InvoicesGroupedByTimeCardStartDate == "true"
	}
	if opts.FlagsSet["invoice-date-calculation"] {
		attributes["invoice-date-calculation"] = opts.InvoiceDateCalculation
	}
	if opts.FlagsSet["create-detected-production-incidents"] {
		attributes["create-detected-production-incidents"] = opts.CreateDetectedProductionIncidents == "true"
	}
	if opts.FlagsSet["primary-color"] {
		attributes["primary-color"] = opts.PrimaryColor
	}
	if opts.FlagsSet["logo-svg"] {
		attributes["logo-svg"] = opts.LogoSVG
	}
	if opts.FlagsSet["default-tender-dispatch-instructions"] {
		attributes["default-tender-dispatch-instructions"] = opts.DefaultTenderDispatchInstructions
	}
	if opts.FlagsSet["default-payment-terms-and-conditions"] {
		attributes["default-payment-terms-and-conditions"] = opts.DefaultPaymentTermsAndConditions
	}
	if opts.FlagsSet["default-hours-after-which-overtime-applies"] {
		attributes["default-hours-after-which-overtime-applies"] = opts.DefaultHoursAfterWhichOvertimeApplies
	}
	if opts.FlagsSet["sets-shift-material-transaction-expectations"] {
		attributes["sets-shift-material-transaction-expectations"] = opts.SetsShiftMaterialTransactionExpectations == "true"
	}
	if opts.FlagsSet["show-planner-info-to-drivers"] {
		attributes["show-planner-info-to-drivers"] = opts.ShowPlannerInfoToDrivers == "true"
	}
	if opts.FlagsSet["is-trucker-shift-rejection-permitted"] {
		attributes["is-trucker-shift-rejection-permitted"] = opts.IsTruckerShiftRejectionPermitted == "true"
	}
	if opts.FlagsSet["notify-driver-when-gps-not-available"] {
		attributes["notify-driver-when-gps-not-available"] = opts.NotifyDriverWhenGPSNotAvailable == "true"
	}
	if opts.FlagsSet["day-shift-assignment-reminder-time"] {
		attributes["day-shift-assignment-reminder-time"] = opts.DayShiftAssignmentReminderTime
	}
	if opts.FlagsSet["night-shift-assignment-reminder-time"] {
		attributes["night-shift-assignment-reminder-time"] = opts.NightShiftAssignmentReminderTime
	}
	if opts.FlagsSet["minimum-driver-tracking-minutes"] {
		var i int
		if _, err := fmt.Sscanf(opts.MinimumDriverTrackingMinutes, "%d", &i); err == nil {
			attributes["minimum-driver-tracking-minutes"] = i
		}
	}
	if opts.FlagsSet["auto-generate-time-sheet-line-items-per-job"] {
		attributes["auto-generate-time-sheet-line-items-per-job"] = opts.AutoGenerateTimeSheetLineItemsPerJob == "true"
	}
	if opts.FlagsSet["restrict-line-item-classification-edit-to-time-sheet-approver"] {
		attributes["restrict-line-item-classification-edit-to-time-sheet-approver"] = opts.RestrictLineItemClassificationEditToTimeSheetApprover == "true"
	}
	if opts.FlagsSet["default-time-sheet-line-item-classification-id"] {
		var i int
		if _, err := fmt.Sscanf(opts.DefaultTimeSheetLineItemClassificationID, "%d", &i); err == nil {
			attributes["default-time-sheet-line-item-classification-id"] = i
		}
	}
	if opts.FlagsSet["auto-combine-overlapping-driver-days"] {
		attributes["auto-combine-overlapping-driver-days"] = opts.AutoCombineOverlappingDriverDays == "true"
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "trucker-settings",
			"id":         id,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/trucker-settings/"+id, jsonBody)
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

	details := buildTruckerSettingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trucker settings %s\n", details.ID)
	return nil
}

func parseDoTruckerSettingsUpdateOptions(cmd *cobra.Command) (doTruckerSettingsUpdateOptions, error) {
	flagsSet := make(map[string]bool)
	flagNames := []string{
		"billing-period-start-on",
		"billing-period-day-count",
		"billing-period-end-invoice-offset-day-count",
		"split-billing-periods-spanning-months",
		"group-daily-invoice-by-job-number",
		"generate-daily-invoice",
		"deliver-new-invoices",
		"invoices-batch-processing-start-on",
		"invoices-grouped-by-time-card-start-date",
		"invoice-date-calculation",
		"create-detected-production-incidents",
		"primary-color",
		"logo-svg",
		"default-tender-dispatch-instructions",
		"default-payment-terms-and-conditions",
		"default-hours-after-which-overtime-applies",
		"sets-shift-material-transaction-expectations",
		"show-planner-info-to-drivers",
		"is-trucker-shift-rejection-permitted",
		"notify-driver-when-gps-not-available",
		"day-shift-assignment-reminder-time",
		"night-shift-assignment-reminder-time",
		"minimum-driver-tracking-minutes",
		"auto-generate-time-sheet-line-items-per-job",
		"restrict-line-item-classification-edit-to-time-sheet-approver",
		"default-time-sheet-line-item-classification-id",
		"auto-combine-overlapping-driver-days",
	}
	for _, name := range flagNames {
		flagsSet[name] = cmd.Flags().Changed(name)
	}

	jsonOut, _ := cmd.Flags().GetBool("json")
	billingPeriodStartOn, _ := cmd.Flags().GetString("billing-period-start-on")
	billingPeriodDayCount, _ := cmd.Flags().GetString("billing-period-day-count")
	billingPeriodEndInvoiceOffsetDayCount, _ := cmd.Flags().GetString("billing-period-end-invoice-offset-day-count")
	splitBillingPeriodsSpanningMonths, _ := cmd.Flags().GetString("split-billing-periods-spanning-months")
	groupDailyInvoiceByJobNumber, _ := cmd.Flags().GetString("group-daily-invoice-by-job-number")
	generateDailyInvoice, _ := cmd.Flags().GetString("generate-daily-invoice")
	deliverNewInvoices, _ := cmd.Flags().GetString("deliver-new-invoices")
	invoicesBatchProcessingStartOn, _ := cmd.Flags().GetString("invoices-batch-processing-start-on")
	invoicesGroupedByTimeCardStartDate, _ := cmd.Flags().GetString("invoices-grouped-by-time-card-start-date")
	invoiceDateCalculation, _ := cmd.Flags().GetString("invoice-date-calculation")
	createDetectedProductionIncidents, _ := cmd.Flags().GetString("create-detected-production-incidents")
	primaryColor, _ := cmd.Flags().GetString("primary-color")
	logoSVG, _ := cmd.Flags().GetString("logo-svg")
	defaultTenderDispatchInstructions, _ := cmd.Flags().GetString("default-tender-dispatch-instructions")
	defaultPaymentTermsAndConditions, _ := cmd.Flags().GetString("default-payment-terms-and-conditions")
	defaultHoursAfterWhichOvertimeApplies, _ := cmd.Flags().GetString("default-hours-after-which-overtime-applies")
	setsShiftMaterialTransactionExpectations, _ := cmd.Flags().GetString("sets-shift-material-transaction-expectations")
	showPlannerInfoToDrivers, _ := cmd.Flags().GetString("show-planner-info-to-drivers")
	isTruckerShiftRejectionPermitted, _ := cmd.Flags().GetString("is-trucker-shift-rejection-permitted")
	notifyDriverWhenGPSNotAvailable, _ := cmd.Flags().GetString("notify-driver-when-gps-not-available")
	dayShiftAssignmentReminderTime, _ := cmd.Flags().GetString("day-shift-assignment-reminder-time")
	nightShiftAssignmentReminderTime, _ := cmd.Flags().GetString("night-shift-assignment-reminder-time")
	minimumDriverTrackingMinutes, _ := cmd.Flags().GetString("minimum-driver-tracking-minutes")
	autoGenerateTimeSheetLineItemsPerJob, _ := cmd.Flags().GetString("auto-generate-time-sheet-line-items-per-job")
	restrictLineItemClassificationEditToTimeSheetApprover, _ := cmd.Flags().GetString("restrict-line-item-classification-edit-to-time-sheet-approver")
	defaultTimeSheetLineItemClassificationID, _ := cmd.Flags().GetString("default-time-sheet-line-item-classification-id")
	autoCombineOverlappingDriverDays, _ := cmd.Flags().GetString("auto-combine-overlapping-driver-days")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckerSettingsUpdateOptions{
		BaseURL:                                               baseURL,
		Token:                                                 token,
		JSON:                                                  jsonOut,
		BillingPeriodStartOn:                                  billingPeriodStartOn,
		BillingPeriodDayCount:                                 billingPeriodDayCount,
		BillingPeriodEndInvoiceOffsetDayCount:                 billingPeriodEndInvoiceOffsetDayCount,
		SplitBillingPeriodsSpanningMonths:                     splitBillingPeriodsSpanningMonths,
		GroupDailyInvoiceByJobNumber:                          groupDailyInvoiceByJobNumber,
		GenerateDailyInvoice:                                  generateDailyInvoice,
		DeliverNewInvoices:                                    deliverNewInvoices,
		InvoicesBatchProcessingStartOn:                        invoicesBatchProcessingStartOn,
		InvoicesGroupedByTimeCardStartDate:                    invoicesGroupedByTimeCardStartDate,
		InvoiceDateCalculation:                                invoiceDateCalculation,
		CreateDetectedProductionIncidents:                     createDetectedProductionIncidents,
		PrimaryColor:                                          primaryColor,
		LogoSVG:                                               logoSVG,
		DefaultTenderDispatchInstructions:                     defaultTenderDispatchInstructions,
		DefaultPaymentTermsAndConditions:                      defaultPaymentTermsAndConditions,
		DefaultHoursAfterWhichOvertimeApplies:                 defaultHoursAfterWhichOvertimeApplies,
		SetsShiftMaterialTransactionExpectations:              setsShiftMaterialTransactionExpectations,
		ShowPlannerInfoToDrivers:                              showPlannerInfoToDrivers,
		IsTruckerShiftRejectionPermitted:                      isTruckerShiftRejectionPermitted,
		NotifyDriverWhenGPSNotAvailable:                       notifyDriverWhenGPSNotAvailable,
		DayShiftAssignmentReminderTime:                        dayShiftAssignmentReminderTime,
		NightShiftAssignmentReminderTime:                      nightShiftAssignmentReminderTime,
		MinimumDriverTrackingMinutes:                          minimumDriverTrackingMinutes,
		AutoGenerateTimeSheetLineItemsPerJob:                  autoGenerateTimeSheetLineItemsPerJob,
		RestrictLineItemClassificationEditToTimeSheetApprover: restrictLineItemClassificationEditToTimeSheetApprover,
		DefaultTimeSheetLineItemClassificationID:              defaultTimeSheetLineItemClassificationID,
		AutoCombineOverlappingDriverDays:                      autoCombineOverlappingDriverDays,
		FlagsSet:                                              flagsSet,
	}, nil
}
