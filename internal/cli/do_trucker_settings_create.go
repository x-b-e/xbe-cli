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

type doTruckerSettingsCreateOptions struct {
	BaseURL                                               string
	Token                                                 string
	JSON                                                  bool
	Trucker                                               string
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
}

func newDoTruckerSettingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create trucker settings",
		Long: `Create trucker settings for a trucking company.

Required flags:
  --trucker    Trucker ID (required)

Notes:
  Boolean flags expect true/false values.
  Date fields use YYYY-MM-DD. Time fields use HH:MM.`,
		Example: `  # Create trucker settings with basic flags
  xbe do trucker-settings create --trucker 123 --notify-driver-when-gps-not-available true

  # Create with billing period configuration
  xbe do trucker-settings create --trucker 123 --billing-period-start-on 2025-01-01 --billing-period-day-count 30

  # Get JSON output
  xbe do trucker-settings create --trucker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTruckerSettingsCreate,
	}
	initDoTruckerSettingsCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckerSettingsCmd.AddCommand(newDoTruckerSettingsCreateCmd())
}

func initDoTruckerSettingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")
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

func runDoTruckerSettingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTruckerSettingsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	if strings.TrimSpace(opts.Trucker) == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	setStringAttrIfPresent(attributes, "billing-period-start-on", opts.BillingPeriodStartOn)
	setIntAttrIfPresent(attributes, "billing-period-day-count", opts.BillingPeriodDayCount)
	setIntAttrIfPresent(attributes, "billing-period-end-invoice-offset-day-count", opts.BillingPeriodEndInvoiceOffsetDayCount)
	setBoolAttrIfPresent(attributes, "split-billing-periods-spanning-months", opts.SplitBillingPeriodsSpanningMonths)
	setBoolAttrIfPresent(attributes, "group-daily-invoice-by-job-number", opts.GroupDailyInvoiceByJobNumber)
	setBoolAttrIfPresent(attributes, "generate-daily-invoice", opts.GenerateDailyInvoice)
	setBoolAttrIfPresent(attributes, "deliver-new-invoices", opts.DeliverNewInvoices)
	setStringAttrIfPresent(attributes, "invoices-batch-processing-start-on", opts.InvoicesBatchProcessingStartOn)
	setBoolAttrIfPresent(attributes, "invoices-grouped-by-time-card-start-date", opts.InvoicesGroupedByTimeCardStartDate)
	setStringAttrIfPresent(attributes, "invoice-date-calculation", opts.InvoiceDateCalculation)
	setBoolAttrIfPresent(attributes, "create-detected-production-incidents", opts.CreateDetectedProductionIncidents)
	setStringAttrIfPresent(attributes, "primary-color", opts.PrimaryColor)
	setStringAttrIfPresent(attributes, "logo-svg", opts.LogoSVG)
	setStringAttrIfPresent(attributes, "default-tender-dispatch-instructions", opts.DefaultTenderDispatchInstructions)
	setStringAttrIfPresent(attributes, "default-payment-terms-and-conditions", opts.DefaultPaymentTermsAndConditions)
	setStringAttrIfPresent(attributes, "default-hours-after-which-overtime-applies", opts.DefaultHoursAfterWhichOvertimeApplies)
	setBoolAttrIfPresent(attributes, "sets-shift-material-transaction-expectations", opts.SetsShiftMaterialTransactionExpectations)
	setBoolAttrIfPresent(attributes, "show-planner-info-to-drivers", opts.ShowPlannerInfoToDrivers)
	setBoolAttrIfPresent(attributes, "is-trucker-shift-rejection-permitted", opts.IsTruckerShiftRejectionPermitted)
	setBoolAttrIfPresent(attributes, "notify-driver-when-gps-not-available", opts.NotifyDriverWhenGPSNotAvailable)
	setStringAttrIfPresent(attributes, "day-shift-assignment-reminder-time", opts.DayShiftAssignmentReminderTime)
	setStringAttrIfPresent(attributes, "night-shift-assignment-reminder-time", opts.NightShiftAssignmentReminderTime)
	setIntAttrIfPresent(attributes, "minimum-driver-tracking-minutes", opts.MinimumDriverTrackingMinutes)
	setBoolAttrIfPresent(attributes, "auto-generate-time-sheet-line-items-per-job", opts.AutoGenerateTimeSheetLineItemsPerJob)
	setBoolAttrIfPresent(attributes, "restrict-line-item-classification-edit-to-time-sheet-approver", opts.RestrictLineItemClassificationEditToTimeSheetApprover)
	setIntAttrIfPresent(attributes, "default-time-sheet-line-item-classification-id", opts.DefaultTimeSheetLineItemClassificationID)
	setBoolAttrIfPresent(attributes, "auto-combine-overlapping-driver-days", opts.AutoCombineOverlappingDriverDays)

	relationships := map[string]any{
		"trucker": map[string]any{
			"data": map[string]string{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trucker-settings",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trucker-settings", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created trucker settings %s\n", details.ID)
	return nil
}

func parseDoTruckerSettingsCreateOptions(cmd *cobra.Command) (doTruckerSettingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trucker, _ := cmd.Flags().GetString("trucker")
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

	return doTruckerSettingsCreateOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
		Trucker:                                  trucker,
		BillingPeriodStartOn:                     billingPeriodStartOn,
		BillingPeriodDayCount:                    billingPeriodDayCount,
		BillingPeriodEndInvoiceOffsetDayCount:    billingPeriodEndInvoiceOffsetDayCount,
		SplitBillingPeriodsSpanningMonths:        splitBillingPeriodsSpanningMonths,
		GroupDailyInvoiceByJobNumber:             groupDailyInvoiceByJobNumber,
		GenerateDailyInvoice:                     generateDailyInvoice,
		DeliverNewInvoices:                       deliverNewInvoices,
		InvoicesBatchProcessingStartOn:           invoicesBatchProcessingStartOn,
		InvoicesGroupedByTimeCardStartDate:       invoicesGroupedByTimeCardStartDate,
		InvoiceDateCalculation:                   invoiceDateCalculation,
		CreateDetectedProductionIncidents:        createDetectedProductionIncidents,
		PrimaryColor:                             primaryColor,
		LogoSVG:                                  logoSVG,
		DefaultTenderDispatchInstructions:        defaultTenderDispatchInstructions,
		DefaultPaymentTermsAndConditions:         defaultPaymentTermsAndConditions,
		DefaultHoursAfterWhichOvertimeApplies:    defaultHoursAfterWhichOvertimeApplies,
		SetsShiftMaterialTransactionExpectations: setsShiftMaterialTransactionExpectations,
		ShowPlannerInfoToDrivers:                 showPlannerInfoToDrivers,
		IsTruckerShiftRejectionPermitted:         isTruckerShiftRejectionPermitted,
		NotifyDriverWhenGPSNotAvailable:          notifyDriverWhenGPSNotAvailable,
		DayShiftAssignmentReminderTime:           dayShiftAssignmentReminderTime,
		NightShiftAssignmentReminderTime:         nightShiftAssignmentReminderTime,
		MinimumDriverTrackingMinutes:             minimumDriverTrackingMinutes,
		AutoGenerateTimeSheetLineItemsPerJob:     autoGenerateTimeSheetLineItemsPerJob,
		RestrictLineItemClassificationEditToTimeSheetApprover: restrictLineItemClassificationEditToTimeSheetApprover,
		DefaultTimeSheetLineItemClassificationID:              defaultTimeSheetLineItemClassificationID,
		AutoCombineOverlappingDriverDays:                      autoCombineOverlappingDriverDays,
	}, nil
}
