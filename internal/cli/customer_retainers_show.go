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

type customerRetainersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type customerRetainerDetails struct {
	ID                                 string   `json:"id"`
	Status                             string   `json:"status,omitempty"`
	TerminatedOn                       string   `json:"terminated_on,omitempty"`
	ExpectedEarnings                   string   `json:"expected_earnings,omitempty"`
	ActualEarnings                     string   `json:"actual_earnings,omitempty"`
	ConsumptionPct                     string   `json:"consumption_pct,omitempty"`
	TermStartOn                        string   `json:"term_start_on,omitempty"`
	TermEndOn                          string   `json:"term_end_on,omitempty"`
	MaximumExpectedDailyHours          string   `json:"maximum_expected_daily_hours,omitempty"`
	MaximumTravelMinutes               string   `json:"maximum_travel_minutes,omitempty"`
	BillableTravelMinutesPerTravelMile string   `json:"billable_travel_minutes_per_travel_mile,omitempty"`
	BuyerType                          string   `json:"buyer_type,omitempty"`
	BuyerID                            string   `json:"buyer_id,omitempty"`
	SellerType                         string   `json:"seller_type,omitempty"`
	SellerID                           string   `json:"seller_id,omitempty"`
	CustomerID                         string   `json:"customer_id,omitempty"`
	BrokerID                           string   `json:"broker_id,omitempty"`
	RateIDs                            []string `json:"rate_ids,omitempty"`
	FileAttachmentIDs                  []string `json:"file_attachment_ids,omitempty"`
	RetainerPeriodIDs                  []string `json:"retainer_period_ids,omitempty"`
	RetainerPaymentIDs                 []string `json:"retainer_payment_ids,omitempty"`
	RetainerDeductionIDs               []string `json:"retainer_deduction_ids,omitempty"`
	TenderJobScheduleShiftIDs          []string `json:"tender_job_schedule_shift_ids,omitempty"`
}

func newCustomerRetainersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show customer retainer details",
		Long: `Show the full details of a customer retainer.

Output Fields:
  ID
  Status
  Terminated On
  Expected Earnings
  Actual Earnings
  Consumption Pct
  Term Start On
  Term End On
  Maximum Expected Daily Hours
  Maximum Travel Minutes
  Billable Travel Minutes Per Travel Mile
  Buyer (type and ID)
  Seller (type and ID)
  Customer ID
  Broker ID
  Rate IDs
  File Attachment IDs
  Retainer Period IDs
  Retainer Payment IDs
  Retainer Deduction IDs
  Tender Job Schedule Shift IDs

Arguments:
  <id>    Customer retainer ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a customer retainer
  xbe view customer-retainers show 123

  # JSON output
  xbe view customer-retainers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCustomerRetainersShow,
	}
	initCustomerRetainersShowFlags(cmd)
	return cmd
}

func init() {
	customerRetainersCmd.AddCommand(newCustomerRetainersShowCmd())
}

func initCustomerRetainersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerRetainersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCustomerRetainersShowOptions(cmd)
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
		return fmt.Errorf("customer retainer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-retainers]", "status,terminated-on,expected-earnings,actual-earnings,consumption-pct,term-start-on,term-end-on,maximum-expected-daily-hours,maximum-travel-minutes,billable-travel-minutes-per-travel-mile,buyer,seller,rates,file-attachments,retainer-periods,retainer-payments,retainer-deductions,tender-job-schedule-shifts")

	body, _, err := client.Get(cmd.Context(), "/v1/customer-retainers/"+id, query)
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

	details := buildCustomerRetainerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCustomerRetainerDetails(cmd, details)
}

func parseCustomerRetainersShowOptions(cmd *cobra.Command) (customerRetainersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerRetainersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCustomerRetainerDetails(resp jsonAPISingleResponse) customerRetainerDetails {
	attrs := resp.Data.Attributes
	details := customerRetainerDetails{
		ID:                                 resp.Data.ID,
		Status:                             stringAttr(attrs, "status"),
		TerminatedOn:                       formatDate(stringAttr(attrs, "terminated-on")),
		ExpectedEarnings:                   stringAttr(attrs, "expected-earnings"),
		ActualEarnings:                     stringAttr(attrs, "actual-earnings"),
		ConsumptionPct:                     stringAttr(attrs, "consumption-pct"),
		TermStartOn:                        formatDate(stringAttr(attrs, "term-start-on")),
		TermEndOn:                          formatDate(stringAttr(attrs, "term-end-on")),
		MaximumExpectedDailyHours:          stringAttr(attrs, "maximum-expected-daily-hours"),
		MaximumTravelMinutes:               stringAttr(attrs, "maximum-travel-minutes"),
		BillableTravelMinutesPerTravelMile: stringAttr(attrs, "billable-travel-minutes-per-travel-mile"),
	}

	if rel, ok := resp.Data.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerType = rel.Data.Type
		details.BuyerID = rel.Data.ID
		if rel.Data.Type == "customers" {
			details.CustomerID = rel.Data.ID
		}
	}
	if rel, ok := resp.Data.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerType = rel.Data.Type
		details.SellerID = rel.Data.ID
		if rel.Data.Type == "brokers" {
			details.BrokerID = rel.Data.ID
		}
	}
	if rel, ok := resp.Data.Relationships["rates"]; ok {
		details.RateIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["file-attachments"]; ok {
		details.FileAttachmentIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["retainer-periods"]; ok {
		details.RetainerPeriodIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["retainer-payments"]; ok {
		details.RetainerPaymentIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["retainer-deductions"]; ok {
		details.RetainerDeductionIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["tender-job-schedule-shifts"]; ok {
		details.TenderJobScheduleShiftIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderCustomerRetainerDetails(cmd *cobra.Command, details customerRetainerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.TerminatedOn != "" {
		fmt.Fprintf(out, "Terminated On: %s\n", details.TerminatedOn)
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
	if details.TermStartOn != "" {
		fmt.Fprintf(out, "Term Start On: %s\n", details.TermStartOn)
	}
	if details.TermEndOn != "" {
		fmt.Fprintf(out, "Term End On: %s\n", details.TermEndOn)
	}
	if details.MaximumExpectedDailyHours != "" {
		fmt.Fprintf(out, "Maximum Expected Daily Hours: %s\n", details.MaximumExpectedDailyHours)
	}
	if details.MaximumTravelMinutes != "" {
		fmt.Fprintf(out, "Maximum Travel Minutes: %s\n", details.MaximumTravelMinutes)
	}
	if details.BillableTravelMinutesPerTravelMile != "" {
		fmt.Fprintf(out, "Billable Travel Minutes Per Travel Mile: %s\n", details.BillableTravelMinutesPerTravelMile)
	}
	if details.BuyerType != "" && details.BuyerID != "" {
		fmt.Fprintf(out, "Buyer: %s/%s\n", details.BuyerType, details.BuyerID)
	}
	if details.SellerType != "" && details.SellerID != "" {
		fmt.Fprintf(out, "Seller: %s/%s\n", details.SellerType, details.SellerID)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer ID: %s\n", details.CustomerID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if len(details.RateIDs) > 0 {
		fmt.Fprintf(out, "Rate IDs: %s\n", strings.Join(details.RateIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachment IDs: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}
	if len(details.RetainerPeriodIDs) > 0 {
		fmt.Fprintf(out, "Retainer Period IDs: %s\n", strings.Join(details.RetainerPeriodIDs, ", "))
	}
	if len(details.RetainerPaymentIDs) > 0 {
		fmt.Fprintf(out, "Retainer Payment IDs: %s\n", strings.Join(details.RetainerPaymentIDs, ", "))
	}
	if len(details.RetainerDeductionIDs) > 0 {
		fmt.Fprintf(out, "Retainer Deduction IDs: %s\n", strings.Join(details.RetainerDeductionIDs, ", "))
	}
	if len(details.TenderJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Tender Job Schedule Shift IDs: %s\n", strings.Join(details.TenderJobScheduleShiftIDs, ", "))
	}

	return nil
}
