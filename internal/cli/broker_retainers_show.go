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

type brokerRetainersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerRetainerDetails struct {
	ID                                 string   `json:"id"`
	Status                             string   `json:"status,omitempty"`
	TerminatedOn                       string   `json:"terminated_on,omitempty"`
	ExpectedEarnings                   any      `json:"expected_earnings,omitempty"`
	ActualEarnings                     any      `json:"actual_earnings,omitempty"`
	ConsumptionPct                     any      `json:"consumption_pct,omitempty"`
	TermStartOn                        string   `json:"term_start_on,omitempty"`
	TermEndOn                          string   `json:"term_end_on,omitempty"`
	MaximumExpectedDailyHours          any      `json:"maximum_expected_daily_hours,omitempty"`
	MaximumTravelMinutes               any      `json:"maximum_travel_minutes,omitempty"`
	BillableTravelMinutesPerTravelMile any      `json:"billable_travel_minutes_per_travel_mile,omitempty"`
	BuyerType                          string   `json:"buyer_type,omitempty"`
	BuyerID                            string   `json:"buyer_id,omitempty"`
	BuyerName                          string   `json:"buyer,omitempty"`
	SellerType                         string   `json:"seller_type,omitempty"`
	SellerID                           string   `json:"seller_id,omitempty"`
	SellerName                         string   `json:"seller,omitempty"`
	RateIDs                            []string `json:"rate_ids,omitempty"`
	FileAttachmentIDs                  []string `json:"file_attachment_ids,omitempty"`
	RetainerPeriodIDs                  []string `json:"retainer_period_ids,omitempty"`
	RetainerPaymentIDs                 []string `json:"retainer_payment_ids,omitempty"`
	RetainerDeductionIDs               []string `json:"retainer_deduction_ids,omitempty"`
	TenderJobScheduleShiftIDs          []string `json:"tender_job_schedule_shift_ids,omitempty"`
}

func newBrokerRetainersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker retainer details",
		Long: `Show the full details of a broker retainer.

Arguments:
  <id>  The broker retainer ID (required).`,
		Example: `  # Show a broker retainer
  xbe view broker-retainers show 123

  # Output as JSON
  xbe view broker-retainers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerRetainersShow,
	}
	initBrokerRetainersShowFlags(cmd)
	return cmd
}

func init() {
	brokerRetainersCmd.AddCommand(newBrokerRetainersShowCmd())
}

func initBrokerRetainersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerRetainersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBrokerRetainersShowOptions(cmd)
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
		return fmt.Errorf("broker retainer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-retainers]", strings.Join([]string{
		"status",
		"terminated-on",
		"expected-earnings",
		"actual-earnings",
		"consumption-pct",
		"term-start-on",
		"term-end-on",
		"maximum-expected-daily-hours",
		"maximum-travel-minutes",
		"billable-travel-minutes-per-travel-mile",
		"buyer",
		"seller",
		"rates",
		"file-attachments",
		"retainer-periods",
		"retainer-payments",
		"retainer-deductions",
		"tender-job-schedule-shifts",
	}, ","))
	query.Set("include", "buyer,seller")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-retainers/"+id, query)
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

	details := buildBrokerRetainerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerRetainerDetails(cmd, details)
}

func parseBrokerRetainersShowOptions(cmd *cobra.Command) (brokerRetainersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerRetainersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerRetainerDetails(resp jsonAPISingleResponse) brokerRetainerDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := brokerRetainerDetails{
		ID:                                 resp.Data.ID,
		Status:                             stringAttr(attrs, "status"),
		TerminatedOn:                       formatDate(stringAttr(attrs, "terminated-on")),
		ExpectedEarnings:                   anyAttr(attrs, "expected-earnings"),
		ActualEarnings:                     anyAttr(attrs, "actual-earnings"),
		ConsumptionPct:                     anyAttr(attrs, "consumption-pct"),
		TermStartOn:                        formatDate(stringAttr(attrs, "term-start-on")),
		TermEndOn:                          formatDate(stringAttr(attrs, "term-end-on")),
		MaximumExpectedDailyHours:          anyAttr(attrs, "maximum-expected-daily-hours"),
		MaximumTravelMinutes:               anyAttr(attrs, "maximum-travel-minutes"),
		BillableTravelMinutesPerTravelMile: anyAttr(attrs, "billable-travel-minutes-per-travel-mile"),
	}

	if rel, ok := resp.Data.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerType = rel.Data.Type
		details.BuyerID = rel.Data.ID
		if buyer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BuyerName = firstNonEmpty(
				stringAttr(buyer.Attributes, "company-name"),
				stringAttr(buyer.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerType = rel.Data.Type
		details.SellerID = rel.Data.ID
		if seller, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.SellerName = firstNonEmpty(
				stringAttr(seller.Attributes, "company-name"),
				stringAttr(seller.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["rates"]; ok {
		details.RateIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["file-attachments"]; ok {
		details.FileAttachmentIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["retainer-periods"]; ok {
		details.RetainerPeriodIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["retainer-payments"]; ok {
		details.RetainerPaymentIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["retainer-deductions"]; ok {
		details.RetainerDeductionIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["tender-job-schedule-shifts"]; ok {
		details.TenderJobScheduleShiftIDs = relationshipIDList(rel)
	}

	return details
}

func renderBrokerRetainerDetails(cmd *cobra.Command, details brokerRetainerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.TerminatedOn != "" {
		fmt.Fprintf(out, "Terminated On: %s\n", details.TerminatedOn)
	}
	if details.ExpectedEarnings != nil {
		fmt.Fprintf(out, "Expected Earnings: %s\n", formatAnyValue(details.ExpectedEarnings))
	}
	if details.ActualEarnings != nil {
		fmt.Fprintf(out, "Actual Earnings: %s\n", formatAnyValue(details.ActualEarnings))
	}
	if details.ConsumptionPct != nil {
		fmt.Fprintf(out, "Consumption Pct: %s\n", formatAnyValue(details.ConsumptionPct))
	}
	if details.TermStartOn != "" {
		fmt.Fprintf(out, "Term Start On: %s\n", details.TermStartOn)
	}
	if details.TermEndOn != "" {
		fmt.Fprintf(out, "Term End On: %s\n", details.TermEndOn)
	}
	if details.MaximumExpectedDailyHours != nil {
		fmt.Fprintf(out, "Max Expected Daily Hours: %s\n", formatAnyValue(details.MaximumExpectedDailyHours))
	}
	if details.MaximumTravelMinutes != nil {
		fmt.Fprintf(out, "Max Travel Minutes: %s\n", formatAnyValue(details.MaximumTravelMinutes))
	}
	if details.BillableTravelMinutesPerTravelMile != nil {
		fmt.Fprintf(out, "Billable Travel Minutes per Mile: %s\n", formatAnyValue(details.BillableTravelMinutesPerTravelMile))
	}

	buyerLabel := formatRelated(details.BuyerName, formatPolymorphic(details.BuyerType, details.BuyerID))
	if buyerLabel != "" {
		fmt.Fprintf(out, "Buyer: %s\n", buyerLabel)
	}
	sellerLabel := formatRelated(details.SellerName, formatPolymorphic(details.SellerType, details.SellerID))
	if sellerLabel != "" {
		fmt.Fprintf(out, "Seller: %s\n", sellerLabel)
	}

	if len(details.RateIDs) > 0 {
		fmt.Fprintf(out, "Rates: %s\n", strings.Join(details.RateIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachments: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}
	if len(details.RetainerPeriodIDs) > 0 {
		fmt.Fprintf(out, "Retainer Periods: %s\n", strings.Join(details.RetainerPeriodIDs, ", "))
	}
	if len(details.RetainerPaymentIDs) > 0 {
		fmt.Fprintf(out, "Retainer Payments: %s\n", strings.Join(details.RetainerPaymentIDs, ", "))
	}
	if len(details.RetainerDeductionIDs) > 0 {
		fmt.Fprintf(out, "Retainer Deductions: %s\n", strings.Join(details.RetainerDeductionIDs, ", "))
	}
	if len(details.TenderJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Tender Job Schedule Shifts: %s\n", strings.Join(details.TenderJobScheduleShiftIDs, ", "))
	}

	return nil
}
