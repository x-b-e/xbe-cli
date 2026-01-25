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

type invoiceStatusChangesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Invoice string
	Status  string
}

type invoiceStatusChangeRow struct {
	ID          string `json:"id"`
	InvoiceID   string `json:"invoice_id,omitempty"`
	Status      string `json:"status,omitempty"`
	ChangedAt   string `json:"changed_at,omitempty"`
	ChangedByID string `json:"changed_by_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newInvoiceStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoice status changes",
		Long: `List invoice status changes.

Output Columns:
  ID          Status change ID
  INVOICE     Invoice ID
  STATUS      Status after the change
  CHANGED AT  Timestamp for the status change
  CHANGED BY  User who made the change (if present)
  COMMENT     Optional comment

Filters:
  --invoice  Filter by invoice ID
  --status   Filter by status

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List invoice status changes
  xbe view invoice-status-changes list

  # Filter by invoice
  xbe view invoice-status-changes list --invoice 123

  # Filter by status
  xbe view invoice-status-changes list --status approved

  # Output as JSON
  xbe view invoice-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runInvoiceStatusChangesList,
	}
	initInvoiceStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	invoiceStatusChangesCmd.AddCommand(newInvoiceStatusChangesListCmd())
}

func initInvoiceStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("invoice", "", "Filter by invoice ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInvoiceStatusChangesListOptions(cmd)
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
	query.Set("fields[invoice-status-changes]", "status,changed-at,comment,invoice,changed-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[invoice]", opts.Invoice)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-status-changes", query)
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

	rows := buildInvoiceStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInvoiceStatusChangesTable(cmd, rows)
}

func parseInvoiceStatusChangesListOptions(cmd *cobra.Command) (invoiceStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	invoice, _ := cmd.Flags().GetString("invoice")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceStatusChangesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Invoice: invoice,
		Status:  status,
	}, nil
}

func buildInvoiceStatusChangeRows(resp jsonAPIResponse) []invoiceStatusChangeRow {
	rows := make([]invoiceStatusChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildInvoiceStatusChangeRow(resource))
	}
	return rows
}

func buildInvoiceStatusChangeRow(resource jsonAPIResource) invoiceStatusChangeRow {
	attrs := resource.Attributes
	row := invoiceStatusChangeRow{
		ID:        resource.ID,
		Status:    stringAttr(attrs, "status"),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		row.ChangedByID = rel.Data.ID
	}

	return row
}

func buildInvoiceStatusChangeRowFromSingle(resp jsonAPISingleResponse) invoiceStatusChangeRow {
	return buildInvoiceStatusChangeRow(resp.Data)
}

func renderInvoiceStatusChangesTable(cmd *cobra.Command, rows []invoiceStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No invoice status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINVOICE\tSTATUS\tCHANGED AT\tCHANGED BY\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.InvoiceID,
			row.Status,
			row.ChangedAt,
			row.ChangedByID,
			truncateString(row.Comment, 50),
		)
	}
	return writer.Flush()
}
