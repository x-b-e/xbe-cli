package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type organizationInvoicesBatchInvoiceBatchingsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type organizationInvoicesBatchInvoiceBatchingRow struct {
	ID                               string `json:"id"`
	OrganizationInvoicesBatchInvoice string `json:"organization_invoices_batch_invoice_id,omitempty"`
	Comment                          string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchInvoiceBatchingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch invoice batchings",
		Long: `List organization invoices batch invoice batchings.

Output Columns:
  ID             Batching identifier
  BATCH INVOICE  Organization invoices batch invoice ID
  COMMENT        Comment (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization invoices batch invoice batchings
  xbe view organization-invoices-batch-invoice-batchings list

  # JSON output
  xbe view organization-invoices-batch-invoice-batchings list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchInvoiceBatchingsList,
	}
	initOrganizationInvoicesBatchInvoiceBatchingsListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchInvoiceBatchingsCmd.AddCommand(newOrganizationInvoicesBatchInvoiceBatchingsListCmd())
}

func initOrganizationInvoicesBatchInvoiceBatchingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchInvoiceBatchingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchInvoiceBatchingsListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-invoice-batchings]", "organization-invoices-batch-invoice,comment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-invoice-batchings", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderOrganizationInvoicesBatchInvoiceBatchingsUnavailable(cmd, opts.JSON)
		}
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

	rows := buildOrganizationInvoicesBatchInvoiceBatchingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchInvoiceBatchingsTable(cmd, rows)
}

func renderOrganizationInvoicesBatchInvoiceBatchingsUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []organizationInvoicesBatchInvoiceBatchingRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Organization invoices batch invoice batchings are write-only; list is not available.")
	return nil
}

func parseOrganizationInvoicesBatchInvoiceBatchingsListOptions(cmd *cobra.Command) (organizationInvoicesBatchInvoiceBatchingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchInvoiceBatchingsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildOrganizationInvoicesBatchInvoiceBatchingRows(resp jsonAPIResponse) []organizationInvoicesBatchInvoiceBatchingRow {
	rows := make([]organizationInvoicesBatchInvoiceBatchingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildOrganizationInvoicesBatchInvoiceBatchingRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildOrganizationInvoicesBatchInvoiceBatchingRow(resource jsonAPIResource) organizationInvoicesBatchInvoiceBatchingRow {
	attrs := resource.Attributes
	row := organizationInvoicesBatchInvoiceBatchingRow{
		ID:      resource.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["organization-invoices-batch-invoice"]; ok && rel.Data != nil {
		row.OrganizationInvoicesBatchInvoice = rel.Data.ID
	}

	return row
}

func renderOrganizationInvoicesBatchInvoiceBatchingsTable(cmd *cobra.Command, rows []organizationInvoicesBatchInvoiceBatchingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch invoice batchings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBATCH INVOICE\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.OrganizationInvoicesBatchInvoice,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
