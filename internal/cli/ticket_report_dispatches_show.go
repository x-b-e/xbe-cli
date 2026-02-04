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

type ticketReportDispatchesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type ticketReportDispatchDetails struct {
	ID                              string `json:"id"`
	TicketReportID                  string `json:"ticket_report_id,omitempty"`
	TicketReportFileName            string `json:"ticket_report_file_name,omitempty"`
	IsFulfilled                     bool   `json:"is_fulfilled,omitempty"`
	FulfillmentApprovedTimeCardTons string `json:"fulfillment_approved_time_card_tons,omitempty"`
	FulfillmentBillableTons         string `json:"fulfillment_billable_tons,omitempty"`
	FulfillmentNonBillableTons      string `json:"fulfillment_non_billable_tons,omitempty"`
	FulfillmentErrors               any    `json:"fulfillment_errors,omitempty"`
	FulfillmentLineupResults        any    `json:"fulfillment_lineup_results,omitempty"`
	CreatedAt                       string `json:"created_at,omitempty"`
	UpdatedAt                       string `json:"updated_at,omitempty"`
}

func newTicketReportDispatchesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show ticket report dispatch details",
		Long: `Show the full details of a ticket report dispatch.

Output Fields:
  ID                               Ticket report dispatch identifier
  Ticket Report                    Ticket report file name and ID
  Is Fulfilled                      Fulfillment status
  Fulfillment Approved Time Card Tons  Approved time card tons
  Fulfillment Billable Tons         Billable tons
  Fulfillment Non-Billable Tons     Non-billable tons
  Fulfillment Errors                Errors from fulfillment (if any)
  Fulfillment Lineup Results        Lineup results from fulfillment
  Created At                        Created timestamp
  Updated At                        Updated timestamp

Arguments:
  <id>  The ticket report dispatch ID (required).`,
		Example: `  # Show a ticket report dispatch
  xbe view ticket-report-dispatches show 123

  # Output as JSON
  xbe view ticket-report-dispatches show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTicketReportDispatchesShow,
	}
	initTicketReportDispatchesShowFlags(cmd)
	return cmd
}

func init() {
	ticketReportDispatchesCmd.AddCommand(newTicketReportDispatchesShowCmd())
}

func initTicketReportDispatchesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTicketReportDispatchesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTicketReportDispatchesShowOptions(cmd)
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
		return fmt.Errorf("ticket report dispatch id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[ticket-report-dispatches]", "ticket-report,is-fulfilled,fulfillment-approved-time-card-tons,fulfillment-billable-tons,fulfillment-non-billable-tons,fulfillment-errors,fulfillment-lineup-results,created-at,updated-at")
	query.Set("include", "ticket-report")
	query.Set("fields[ticket-reports]", "file-name")

	body, _, err := client.Get(cmd.Context(), "/v1/ticket-report-dispatches/"+id, query)
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

	details := buildTicketReportDispatchDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTicketReportDispatchDetails(cmd, details)
}

func parseTicketReportDispatchesShowOptions(cmd *cobra.Command) (ticketReportDispatchesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return ticketReportDispatchesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTicketReportDispatchDetails(resp jsonAPISingleResponse) ticketReportDispatchDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := ticketReportDispatchDetails{
		ID:                              resp.Data.ID,
		IsFulfilled:                     boolAttr(attrs, "is-fulfilled"),
		FulfillmentApprovedTimeCardTons: stringAttr(attrs, "fulfillment-approved-time-card-tons"),
		FulfillmentBillableTons:         stringAttr(attrs, "fulfillment-billable-tons"),
		FulfillmentNonBillableTons:      stringAttr(attrs, "fulfillment-non-billable-tons"),
		FulfillmentErrors:               anyAttr(attrs, "fulfillment-errors"),
		FulfillmentLineupResults:        anyAttr(attrs, "fulfillment-lineup-results"),
		CreatedAt:                       formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                       formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["ticket-report"]; ok && rel.Data != nil {
		details.TicketReportID = rel.Data.ID
		if report, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TicketReportFileName = stringAttr(report.Attributes, "file-name")
		}
	}

	return details
}

func renderTicketReportDispatchDetails(cmd *cobra.Command, details ticketReportDispatchDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TicketReportID != "" || details.TicketReportFileName != "" {
		fmt.Fprintf(out, "Ticket Report: %s\n", formatRelated(details.TicketReportFileName, details.TicketReportID))
	}
	fmt.Fprintf(out, "Is Fulfilled: %t\n", details.IsFulfilled)
	if details.FulfillmentApprovedTimeCardTons != "" {
		fmt.Fprintf(out, "Fulfillment Approved Time Card Tons: %s\n", details.FulfillmentApprovedTimeCardTons)
	}
	if details.FulfillmentBillableTons != "" {
		fmt.Fprintf(out, "Fulfillment Billable Tons: %s\n", details.FulfillmentBillableTons)
	}
	if details.FulfillmentNonBillableTons != "" {
		fmt.Fprintf(out, "Fulfillment Non-Billable Tons: %s\n", details.FulfillmentNonBillableTons)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	if formatted := formatAnyJSON(details.FulfillmentErrors); formatted != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Fulfillment Errors:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatted)
	}

	if formatted := formatAnyJSON(details.FulfillmentLineupResults); formatted != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Fulfillment Lineup Results:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatted)
	}

	return nil
}
