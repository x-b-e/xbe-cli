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

type truckerSettingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type truckerSettingDetails struct {
	ID                                                    string `json:"id"`
	TruckerID                                             string `json:"trucker_id,omitempty"`
	TruckerName                                           string `json:"trucker_name,omitempty"`
	BillingPeriodStartOn                                  string `json:"billing_period_start_on,omitempty"`
	BillingPeriodDayCount                                 int    `json:"billing_period_day_count"`
	BillingPeriodEndInvoiceOffsetDayCount                 int    `json:"billing_period_end_invoice_offset_day_count"`
	SplitBillingPeriodsSpanningMonths                     bool   `json:"split_billing_periods_spanning_months"`
	GroupDailyInvoiceByJobNumber                          bool   `json:"group_daily_invoice_by_job_number"`
	GenerateDailyInvoice                                  bool   `json:"generate_daily_invoice"`
	DeliverNewInvoices                                    bool   `json:"deliver_new_invoices"`
	InvoicesBatchProcessingStartOn                        string `json:"invoices_batch_processing_start_on,omitempty"`
	InvoicesGroupedByTimeCardStartDate                    bool   `json:"invoices_grouped_by_time_card_start_date"`
	InvoiceDateCalculation                                string `json:"invoice_date_calculation,omitempty"`
	CreateDetectedProductionIncidents                     bool   `json:"create_detected_production_incidents"`
	PrimaryColor                                          string `json:"primary_color,omitempty"`
	LogoSVG                                               string `json:"logo_svg,omitempty"`
	DefaultTenderDispatchInstructions                     string `json:"default_tender_dispatch_instructions,omitempty"`
	DefaultPaymentTermsAndConditions                      string `json:"default_payment_terms_and_conditions,omitempty"`
	DefaultHoursAfterWhichOvertimeApplies                 string `json:"default_hours_after_which_overtime_applies,omitempty"`
	SetsShiftMaterialTransactionExpectations              bool   `json:"sets_shift_material_transaction_expectations"`
	ShowPlannerInfoToDrivers                              bool   `json:"show_planner_info_to_drivers"`
	IsTruckerShiftRejectionPermitted                      bool   `json:"is_trucker_shift_rejection_permitted"`
	NotifyDriverWhenGPSNotAvailable                       bool   `json:"notify_driver_when_gps_not_available"`
	DayShiftAssignmentReminderTime                        string `json:"day_shift_assignment_reminder_time,omitempty"`
	NightShiftAssignmentReminderTime                      string `json:"night_shift_assignment_reminder_time,omitempty"`
	MinimumDriverTrackingMinutes                          int    `json:"minimum_driver_tracking_minutes"`
	AutoGenerateTimeSheetLineItemsPerJob                  bool   `json:"auto_generate_time_sheet_line_items_per_job"`
	RestrictLineItemClassificationEditToTimeSheetApprover bool   `json:"restrict_line_item_classification_edit_to_time_sheet_approver"`
	DefaultTimeSheetLineItemClassificationID              string `json:"default_time_sheet_line_item_classification_id,omitempty"`
	AutoCombineOverlappingDriverDays                      bool   `json:"auto_combine_overlapping_driver_days"`
}

func newTruckerSettingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show trucker setting details",
		Long: `Show the full details of a specific trucker setting.

Trucker settings configure invoicing, shift, and time sheet behaviors for a
trucking company.

Arguments:
  <id>    The trucker setting ID (required). You can find IDs using the list command.`,
		Example: `  # View a trucker setting
  xbe view trucker-settings show 123

  # Get JSON output
  xbe view trucker-settings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTruckerSettingsShow,
	}
	initTruckerSettingsShowFlags(cmd)
	return cmd
}

func init() {
	truckerSettingsCmd.AddCommand(newTruckerSettingsShowCmd())
}

func initTruckerSettingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerSettingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTruckerSettingsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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
		return fmt.Errorf("trucker setting id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "trucker")
	query.Set("fields[truckers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-settings/"+id, query)
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

	return renderTruckerSettingDetails(cmd, details)
}

func parseTruckerSettingsShowOptions(cmd *cobra.Command) (truckerSettingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerSettingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTruckerSettingDetails(resp jsonAPISingleResponse) truckerSettingDetails {
	attrs := resp.Data.Attributes
	details := truckerSettingDetails{
		ID:                                                    resp.Data.ID,
		BillingPeriodStartOn:                                  stringAttr(attrs, "billing-period-start-on"),
		BillingPeriodDayCount:                                 intAttr(attrs, "billing-period-day-count"),
		BillingPeriodEndInvoiceOffsetDayCount:                 intAttr(attrs, "billing-period-end-invoice-offset-day-count"),
		SplitBillingPeriodsSpanningMonths:                     boolAttr(attrs, "split-billing-periods-spanning-months"),
		GroupDailyInvoiceByJobNumber:                          boolAttr(attrs, "group-daily-invoice-by-job-number"),
		GenerateDailyInvoice:                                  boolAttr(attrs, "generate-daily-invoice"),
		DeliverNewInvoices:                                    boolAttr(attrs, "deliver-new-invoices"),
		InvoicesBatchProcessingStartOn:                        stringAttr(attrs, "invoices-batch-processing-start-on"),
		InvoicesGroupedByTimeCardStartDate:                    boolAttr(attrs, "invoices-grouped-by-time-card-start-date"),
		InvoiceDateCalculation:                                stringAttr(attrs, "invoice-date-calculation"),
		CreateDetectedProductionIncidents:                     boolAttr(attrs, "create-detected-production-incidents"),
		PrimaryColor:                                          stringAttr(attrs, "primary-color"),
		LogoSVG:                                               stringAttr(attrs, "logo-svg"),
		DefaultTenderDispatchInstructions:                     stringAttr(attrs, "default-tender-dispatch-instructions"),
		DefaultPaymentTermsAndConditions:                      stringAttr(attrs, "default-payment-terms-and-conditions"),
		DefaultHoursAfterWhichOvertimeApplies:                 stringAttr(attrs, "default-hours-after-which-overtime-applies"),
		SetsShiftMaterialTransactionExpectations:              boolAttr(attrs, "sets-shift-material-transaction-expectations"),
		ShowPlannerInfoToDrivers:                              boolAttr(attrs, "show-planner-info-to-drivers"),
		IsTruckerShiftRejectionPermitted:                      boolAttr(attrs, "is-trucker-shift-rejection-permitted"),
		NotifyDriverWhenGPSNotAvailable:                       boolAttr(attrs, "notify-driver-when-gps-not-available"),
		DayShiftAssignmentReminderTime:                        stringAttr(attrs, "day-shift-assignment-reminder-time"),
		NightShiftAssignmentReminderTime:                      stringAttr(attrs, "night-shift-assignment-reminder-time"),
		MinimumDriverTrackingMinutes:                          intAttr(attrs, "minimum-driver-tracking-minutes"),
		AutoGenerateTimeSheetLineItemsPerJob:                  boolAttr(attrs, "auto-generate-time-sheet-line-items-per-job"),
		RestrictLineItemClassificationEditToTimeSheetApprover: boolAttr(attrs, "restrict-line-item-classification-edit-to-time-sheet-approver"),
		DefaultTimeSheetLineItemClassificationID:              stringAttr(attrs, "default-time-sheet-line-item-classification-id"),
		AutoCombineOverlappingDriverDays:                      boolAttr(attrs, "auto-combine-overlapping-driver-days"),
	}

	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}

	if details.TruckerID != "" {
		included := make(map[string]jsonAPIResource)
		for _, inc := range resp.Included {
			included[resourceKey(inc.Type, inc.ID)] = inc
		}
		if trucker, ok := included[resourceKey("truckers", details.TruckerID)]; ok {
			details.TruckerName = strings.TrimSpace(stringAttr(trucker.Attributes, "company-name"))
		}
	}

	return details
}

func renderTruckerSettingDetails(cmd *cobra.Command, details truckerSettingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TruckerName != "" {
		if details.TruckerID != "" {
			fmt.Fprintf(out, "Trucker: %s (%s)\n", details.TruckerName, details.TruckerID)
		} else {
			fmt.Fprintf(out, "Trucker: %s\n", details.TruckerName)
		}
	} else if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}

	fmt.Fprintf(out, "Billing Period Start On: %s\n", formatOptional(details.BillingPeriodStartOn))
	fmt.Fprintf(out, "Billing Period Day Count: %d\n", details.BillingPeriodDayCount)
	fmt.Fprintf(out, "Billing Period End Invoice Offset Day Count: %d\n", details.BillingPeriodEndInvoiceOffsetDayCount)
	fmt.Fprintf(out, "Split Billing Periods Spanning Months: %s\n", formatBool(details.SplitBillingPeriodsSpanningMonths))
	fmt.Fprintf(out, "Group Daily Invoice By Job Number: %s\n", formatBool(details.GroupDailyInvoiceByJobNumber))
	fmt.Fprintf(out, "Generate Daily Invoice: %s\n", formatBool(details.GenerateDailyInvoice))
	fmt.Fprintf(out, "Deliver New Invoices: %s\n", formatBool(details.DeliverNewInvoices))
	fmt.Fprintf(out, "Invoices Batch Processing Start On: %s\n", formatOptional(details.InvoicesBatchProcessingStartOn))
	fmt.Fprintf(out, "Invoices Grouped By Time Card Start Date: %s\n", formatBool(details.InvoicesGroupedByTimeCardStartDate))
	fmt.Fprintf(out, "Invoice Date Calculation: %s\n", formatOptional(details.InvoiceDateCalculation))
	fmt.Fprintf(out, "Create Detected Production Incidents: %s\n", formatBool(details.CreateDetectedProductionIncidents))
	fmt.Fprintf(out, "Primary Color: %s\n", formatOptional(details.PrimaryColor))
	fmt.Fprintf(out, "Logo SVG: %s\n", formatOptional(details.LogoSVG))
	fmt.Fprintf(out, "Default Tender Dispatch Instructions: %s\n", formatOptional(details.DefaultTenderDispatchInstructions))
	fmt.Fprintf(out, "Default Payment Terms And Conditions: %s\n", formatOptional(details.DefaultPaymentTermsAndConditions))
	fmt.Fprintf(out, "Default Hours After Which Overtime Applies: %s\n", formatOptional(details.DefaultHoursAfterWhichOvertimeApplies))
	fmt.Fprintf(out, "Sets Shift Material Transaction Expectations: %s\n", formatBool(details.SetsShiftMaterialTransactionExpectations))
	fmt.Fprintf(out, "Show Planner Info To Drivers: %s\n", formatBool(details.ShowPlannerInfoToDrivers))
	fmt.Fprintf(out, "Is Trucker Shift Rejection Permitted: %s\n", formatBool(details.IsTruckerShiftRejectionPermitted))
	fmt.Fprintf(out, "Notify Driver When GPS Not Available: %s\n", formatBool(details.NotifyDriverWhenGPSNotAvailable))
	fmt.Fprintf(out, "Day Shift Assignment Reminder Time: %s\n", formatOptional(details.DayShiftAssignmentReminderTime))
	fmt.Fprintf(out, "Night Shift Assignment Reminder Time: %s\n", formatOptional(details.NightShiftAssignmentReminderTime))
	fmt.Fprintf(out, "Minimum Driver Tracking Minutes: %d\n", details.MinimumDriverTrackingMinutes)
	fmt.Fprintf(out, "Auto Generate Time Sheet Line Items Per Job: %s\n", formatBool(details.AutoGenerateTimeSheetLineItemsPerJob))
	fmt.Fprintf(out, "Restrict Line Item Classification Edit To Time Sheet Approver: %s\n", formatBool(details.RestrictLineItemClassificationEditToTimeSheetApprover))
	fmt.Fprintf(out, "Default Time Sheet Line Item Classification ID: %s\n", formatOptional(details.DefaultTimeSheetLineItemClassificationID))
	fmt.Fprintf(out, "Auto Combine Overlapping Driver Days: %s\n", formatBool(details.AutoCombineOverlappingDriverDays))

	return nil
}
