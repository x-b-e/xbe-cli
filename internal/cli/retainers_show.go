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

type retainersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type retainerDetails struct {
	ID                               string   `json:"id"`
	PolymorphicType                  string   `json:"polymorphic_type,omitempty"`
	Status                           string   `json:"status,omitempty"`
	TerminatedOn                     string   `json:"terminated_on,omitempty"`
	ExpectedEarnings                 string   `json:"expected_earnings,omitempty"`
	ActualEarnings                   string   `json:"actual_earnings,omitempty"`
	ConsumptionPct                   string   `json:"consumption_pct,omitempty"`
	TermStartOn                      string   `json:"term_start_on,omitempty"`
	TermEndOn                        string   `json:"term_end_on,omitempty"`
	MaximumExpectedDailyHours        string   `json:"maximum_expected_daily_hours,omitempty"`
	MaximumTravelMinutes             string   `json:"maximum_travel_minutes,omitempty"`
	BillableTravelMinutesPerTravelMi string   `json:"billable_travel_minutes_per_travel_mile,omitempty"`
	BuyerType                        string   `json:"buyer_type,omitempty"`
	BuyerID                          string   `json:"buyer_id,omitempty"`
	SellerType                       string   `json:"seller_type,omitempty"`
	SellerID                         string   `json:"seller_id,omitempty"`
	RateIDs                          []string `json:"rate_ids,omitempty"`
	FileAttachmentIDs                []string `json:"file_attachment_ids,omitempty"`
	RetainerPeriodIDs                []string `json:"retainer_period_ids,omitempty"`
	RetainerPaymentIDs               []string `json:"retainer_payment_ids,omitempty"`
	RetainerDeductionIDs             []string `json:"retainer_deduction_ids,omitempty"`
	TenderJobScheduleShiftIDs        []string `json:"tender_job_schedule_shift_ids,omitempty"`
}

func newRetainersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show retainer details",
		Long: `Show the full details of a retainer.

Output Fields:
  ID
  Type
  Status
  Terminated On
  Expected/Actual Earnings
  Consumption Percentage
  Term Start/End
  Maximum Expected Daily Hours
  Maximum Travel Minutes
  Billable Travel Minutes per Travel Mile
  Buyer (type/id)
  Seller (type/id)
  Rates
  File Attachments
  Retainer Periods
  Retainer Payments
  Retainer Deductions
  Tender Job Schedule Shifts

Arguments:
  <id>    The retainer ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a retainer
  xbe view retainers show 123

  # Output as JSON
  xbe view retainers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runRetainersShow,
	}
	initRetainersShowFlags(cmd)
	return cmd
}

func init() {
	retainersCmd.AddCommand(newRetainersShowCmd())
}

func initRetainersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseRetainersShowOptions(cmd)
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
		return fmt.Errorf("retainer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[retainers]", "polymorphic-type,status,terminated-on,expected-earnings,actual-earnings,consumption-pct,term-start-on,term-end-on,maximum-expected-daily-hours,maximum-travel-minutes,billable-travel-minutes-per-travel-mile,buyer,seller,rates,file-attachments,retainer-periods,retainer-payments,retainer-deductions,tender-job-schedule-shifts")

	body, _, err := client.Get(cmd.Context(), "/v1/retainers/"+id, query)
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

	details := buildRetainerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderRetainerDetails(cmd, details)
}

func parseRetainersShowOptions(cmd *cobra.Command) (retainersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildRetainerDetails(resp jsonAPISingleResponse) retainerDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := retainerDetails{
		ID:                               resource.ID,
		PolymorphicType:                  stringAttr(attrs, "polymorphic-type"),
		Status:                           stringAttr(attrs, "status"),
		TerminatedOn:                     stringAttr(attrs, "terminated-on"),
		ExpectedEarnings:                 stringAttr(attrs, "expected-earnings"),
		ActualEarnings:                   stringAttr(attrs, "actual-earnings"),
		ConsumptionPct:                   stringAttr(attrs, "consumption-pct"),
		TermStartOn:                      stringAttr(attrs, "term-start-on"),
		TermEndOn:                        stringAttr(attrs, "term-end-on"),
		MaximumExpectedDailyHours:        stringAttr(attrs, "maximum-expected-daily-hours"),
		MaximumTravelMinutes:             stringAttr(attrs, "maximum-travel-minutes"),
		BillableTravelMinutesPerTravelMi: stringAttr(attrs, "billable-travel-minutes-per-travel-mile"),
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerType = rel.Data.Type
		details.BuyerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerType = rel.Data.Type
		details.SellerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["rates"]; ok {
		details.RateIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["file-attachments"]; ok {
		details.FileAttachmentIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["retainer-periods"]; ok {
		details.RetainerPeriodIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["retainer-payments"]; ok {
		details.RetainerPaymentIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["retainer-deductions"]; ok {
		details.RetainerDeductionIDs = relationshipIDList(rel)
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shifts"]; ok {
		details.TenderJobScheduleShiftIDs = relationshipIDList(rel)
	}

	return details
}

func renderRetainerDetails(cmd *cobra.Command, details retainerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PolymorphicType != "" {
		fmt.Fprintf(out, "Type: %s\n", details.PolymorphicType)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.TerminatedOn != "" {
		fmt.Fprintf(out, "Terminated On: %s\n", details.TerminatedOn)
	}
	if details.TermStartOn != "" {
		fmt.Fprintf(out, "Term Start On: %s\n", details.TermStartOn)
	}
	if details.TermEndOn != "" {
		fmt.Fprintf(out, "Term End On: %s\n", details.TermEndOn)
	}
	if details.ExpectedEarnings != "" {
		fmt.Fprintf(out, "Expected Earnings: %s\n", details.ExpectedEarnings)
	}
	if details.ActualEarnings != "" {
		fmt.Fprintf(out, "Actual Earnings: %s\n", details.ActualEarnings)
	}
	if details.ConsumptionPct != "" {
		fmt.Fprintf(out, "Consumption Pct: %s\n", details.ConsumptionPct)
	}
	if details.MaximumExpectedDailyHours != "" {
		fmt.Fprintf(out, "Maximum Expected Daily Hours: %s\n", details.MaximumExpectedDailyHours)
	}
	if details.MaximumTravelMinutes != "" {
		fmt.Fprintf(out, "Maximum Travel Minutes: %s\n", details.MaximumTravelMinutes)
	}
	if details.BillableTravelMinutesPerTravelMi != "" {
		fmt.Fprintf(out, "Billable Travel Minutes per Travel Mile: %s\n", details.BillableTravelMinutesPerTravelMi)
	}
	if details.BuyerType != "" || details.BuyerID != "" {
		fmt.Fprintf(out, "Buyer: %s\n", formatTypeID(details.BuyerType, details.BuyerID))
	}
	if details.SellerType != "" || details.SellerID != "" {
		fmt.Fprintf(out, "Seller: %s\n", formatTypeID(details.SellerType, details.SellerID))
	}

	if len(details.RateIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Rates (%d):\n", len(details.RateIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.RateIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "File Attachments (%d):\n", len(details.FileAttachmentIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.FileAttachmentIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	if len(details.RetainerPeriodIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Retainer Periods (%d):\n", len(details.RetainerPeriodIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.RetainerPeriodIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	if len(details.RetainerPaymentIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Retainer Payments (%d):\n", len(details.RetainerPaymentIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.RetainerPaymentIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	if len(details.RetainerDeductionIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Retainer Deductions (%d):\n", len(details.RetainerDeductionIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.RetainerDeductionIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	if len(details.TenderJobScheduleShiftIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Tender Job Schedule Shifts (%d):\n", len(details.TenderJobScheduleShiftIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.TenderJobScheduleShiftIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	return nil
}
