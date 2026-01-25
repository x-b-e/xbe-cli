package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type ticketReportDispatchesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	TicketReport string
}

type ticketReportDispatchRow struct {
	ID                              string `json:"id"`
	TicketReportID                  string `json:"ticket_report_id,omitempty"`
	TicketReportFileName            string `json:"ticket_report_file_name,omitempty"`
	IsFulfilled                     bool   `json:"is_fulfilled,omitempty"`
	FulfillmentApprovedTimeCardTons string `json:"fulfillment_approved_time_card_tons,omitempty"`
	FulfillmentBillableTons         string `json:"fulfillment_billable_tons,omitempty"`
	FulfillmentNonBillableTons      string `json:"fulfillment_non_billable_tons,omitempty"`
}

func newTicketReportDispatchesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ticket report dispatches",
		Long: `List ticket report dispatches.

Output Columns:
  ID         Ticket report dispatch identifier
  REPORT     Ticket report file name or ID
  FULFILLED  Fulfillment status
  APPROVED   Approved time card tons
  BILLABLE   Billable tons
  NON-BILL   Non-billable tons

Filters:
  --ticket-report  Filter by ticket report ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List ticket report dispatches
  xbe view ticket-report-dispatches list

  # Filter by ticket report
  xbe view ticket-report-dispatches list --ticket-report 123

  # Output as JSON
  xbe view ticket-report-dispatches list --json`,
		Args: cobra.NoArgs,
		RunE: runTicketReportDispatchesList,
	}
	initTicketReportDispatchesListFlags(cmd)
	return cmd
}

func init() {
	ticketReportDispatchesCmd.AddCommand(newTicketReportDispatchesListCmd())
}

func initTicketReportDispatchesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("ticket-report", "", "Filter by ticket report ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTicketReportDispatchesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTicketReportDispatchesListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[ticket-report-dispatches]", "ticket-report,is-fulfilled,fulfillment-approved-time-card-tons,fulfillment-billable-tons,fulfillment-non-billable-tons")
	query.Set("include", "ticket-report")
	query.Set("fields[ticket-reports]", "file-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[ticket-report]", opts.TicketReport)

	body, _, err := client.Get(cmd.Context(), "/v1/ticket-report-dispatches", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildTicketReportDispatchRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTicketReportDispatchesTable(cmd, rows)
}

func parseTicketReportDispatchesListOptions(cmd *cobra.Command) (ticketReportDispatchesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	ticketReport, _ := cmd.Flags().GetString("ticket-report")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return ticketReportDispatchesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		TicketReport: ticketReport,
	}, nil
}

func buildTicketReportDispatchRows(resp jsonAPIResponse) []ticketReportDispatchRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]ticketReportDispatchRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTicketReportDispatchRow(resource, included))
	}
	return rows
}

func ticketReportDispatchRowFromSingle(resp jsonAPISingleResponse) ticketReportDispatchRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildTicketReportDispatchRow(resp.Data, included)
}

func buildTicketReportDispatchRow(resource jsonAPIResource, included map[string]jsonAPIResource) ticketReportDispatchRow {
	attrs := resource.Attributes
	row := ticketReportDispatchRow{
		ID:                              resource.ID,
		IsFulfilled:                     boolAttr(attrs, "is-fulfilled"),
		FulfillmentApprovedTimeCardTons: stringAttr(attrs, "fulfillment-approved-time-card-tons"),
		FulfillmentBillableTons:         stringAttr(attrs, "fulfillment-billable-tons"),
		FulfillmentNonBillableTons:      stringAttr(attrs, "fulfillment-non-billable-tons"),
	}

	if rel, ok := resource.Relationships["ticket-report"]; ok && rel.Data != nil {
		row.TicketReportID = rel.Data.ID
		if report, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TicketReportFileName = stringAttr(report.Attributes, "file-name")
		}
	}

	return row
}

func renderTicketReportDispatchesTable(cmd *cobra.Command, rows []ticketReportDispatchRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No ticket report dispatches found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREPORT\tFULFILLED\tAPPROVED\tBILLABLE\tNON-BILL")
	for _, row := range rows {
		report := firstNonEmpty(row.TicketReportFileName, row.TicketReportID)
		fmt.Fprintf(writer, "%s\t%s\t%t\t%s\t%s\t%s\n",
			row.ID,
			truncateString(report, 32),
			row.IsFulfilled,
			row.FulfillmentApprovedTimeCardTons,
			row.FulfillmentBillableTons,
			row.FulfillmentNonBillableTons,
		)
	}
	return writer.Flush()
}
