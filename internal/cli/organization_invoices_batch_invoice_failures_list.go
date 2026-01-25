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

type organizationInvoicesBatchInvoiceFailuresListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type organizationInvoicesBatchInvoiceFailureRow struct {
	ID                               string `json:"id"`
	OrganizationInvoicesBatchInvoice string `json:"organization_invoices_batch_invoice_id,omitempty"`
	Comment                          string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchInvoiceFailuresListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch invoice failures",
		Long: `List organization invoices batch invoice failures.

Output Columns:
  ID             Failure identifier
  BATCH INVOICE  Organization invoices batch invoice ID
  COMMENT        Comment (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization invoices batch invoice failures
  xbe view organization-invoices-batch-invoice-failures list

  # JSON output
  xbe view organization-invoices-batch-invoice-failures list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchInvoiceFailuresList,
	}
	initOrganizationInvoicesBatchInvoiceFailuresListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchInvoiceFailuresCmd.AddCommand(newOrganizationInvoicesBatchInvoiceFailuresListCmd())
}

func initOrganizationInvoicesBatchInvoiceFailuresListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchInvoiceFailuresList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchInvoiceFailuresListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-invoice-failures]", "organization-invoices-batch-invoice,comment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-invoice-failures", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderOrganizationInvoicesBatchInvoiceFailuresUnavailable(cmd, opts.JSON)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildOrganizationInvoicesBatchInvoiceFailureRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchInvoiceFailuresTable(cmd, rows)
}

func renderOrganizationInvoicesBatchInvoiceFailuresUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []organizationInvoicesBatchInvoiceFailureRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Organization invoices batch invoice failures are write-only; list is not available.")
	return nil
}

func parseOrganizationInvoicesBatchInvoiceFailuresListOptions(cmd *cobra.Command) (organizationInvoicesBatchInvoiceFailuresListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchInvoiceFailuresListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildOrganizationInvoicesBatchInvoiceFailureRows(resp jsonAPIResponse) []organizationInvoicesBatchInvoiceFailureRow {
	rows := make([]organizationInvoicesBatchInvoiceFailureRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildOrganizationInvoicesBatchInvoiceFailureRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildOrganizationInvoicesBatchInvoiceFailureRow(resource jsonAPIResource) organizationInvoicesBatchInvoiceFailureRow {
	attrs := resource.Attributes
	row := organizationInvoicesBatchInvoiceFailureRow{
		ID:      resource.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["organization-invoices-batch-invoice"]; ok && rel.Data != nil {
		row.OrganizationInvoicesBatchInvoice = rel.Data.ID
	}

	return row
}

func renderOrganizationInvoicesBatchInvoiceFailuresTable(cmd *cobra.Command, rows []organizationInvoicesBatchInvoiceFailureRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch invoice failures found.")
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
