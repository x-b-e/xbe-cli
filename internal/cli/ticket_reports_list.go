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

type ticketReportsListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	Broker           string
	TicketReportType string
}

type ticketReportRow struct {
	ID                 string `json:"id"`
	FileName           string `json:"file_name,omitempty"`
	TransformError     string `json:"transform_error,omitempty"`
	BrokerID           string `json:"broker_id,omitempty"`
	TicketReportTypeID string `json:"ticket_report_type_id,omitempty"`
}

func newTicketReportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ticket reports",
		Long: `List ticket reports with filtering and pagination.

Ticket reports track ticket files uploaded for dispatch, import, or validation.

Output Columns:
  ID             Ticket report identifier
  FILE NAME      Uploaded file name
  BROKER         Broker ID
  REPORT TYPE    Ticket report type ID
  TRANSFORM ERR  Transform error message (if any)

Filters:
  --broker             Filter by broker ID
  --ticket-report-type Filter by ticket report type ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List ticket reports
  xbe view ticket-reports list

  # Filter by broker
  xbe view ticket-reports list --broker 123

  # Filter by report type
  xbe view ticket-reports list --ticket-report-type 456

  # Output as JSON
  xbe view ticket-reports list --json`,
		Args: cobra.NoArgs,
		RunE: runTicketReportsList,
	}
	initTicketReportsListFlags(cmd)
	return cmd
}

func init() {
	ticketReportsCmd.AddCommand(newTicketReportsListCmd())
}

func initTicketReportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("ticket-report-type", "", "Filter by ticket report type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTicketReportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTicketReportsListOptions(cmd)
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
	query.Set("fields[ticket-reports]", "file-name,transform-error,broker,ticket-report-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[ticket-report-type]", opts.TicketReportType)

	body, _, err := client.Get(cmd.Context(), "/v1/ticket-reports", query)
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

	rows := buildTicketReportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTicketReportsTable(cmd, rows)
}

func parseTicketReportsListOptions(cmd *cobra.Command) (ticketReportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	ticketReportType, _ := cmd.Flags().GetString("ticket-report-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return ticketReportsListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		Broker:           broker,
		TicketReportType: ticketReportType,
	}, nil
}

func buildTicketReportRows(resp jsonAPIResponse) []ticketReportRow {
	rows := make([]ticketReportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTicketReportRow(resource))
	}
	return rows
}

func buildTicketReportRow(resource jsonAPIResource) ticketReportRow {
	attrs := resource.Attributes
	return ticketReportRow{
		ID:                 resource.ID,
		FileName:           stringAttr(attrs, "file-name"),
		TransformError:     stringAttr(attrs, "transform-error"),
		BrokerID:           relationshipIDFromMap(resource.Relationships, "broker"),
		TicketReportTypeID: relationshipIDFromMap(resource.Relationships, "ticket-report-type"),
	}
}

func buildTicketReportRowFromSingle(resp jsonAPISingleResponse) ticketReportRow {
	return buildTicketReportRow(resp.Data)
}

func renderTicketReportsTable(cmd *cobra.Command, rows []ticketReportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No ticket reports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tFILE NAME\tBROKER\tREPORT TYPE\tTRANSFORM ERR")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.FileName, 30),
			row.BrokerID,
			row.TicketReportTypeID,
			truncateString(row.TransformError, 40),
		)
	}
	return writer.Flush()
}
