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

type organizationInvoicesBatchInvoiceStatusChangesListOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	NoAuth                           bool
	Limit                            int
	Offset                           int
	Sort                             string
	OrganizationInvoicesBatchInvoice string
	Status                           string
}

type organizationInvoicesBatchInvoiceStatusChangeRow struct {
	ID                                 string `json:"id"`
	OrganizationInvoicesBatchInvoiceID string `json:"organization_invoices_batch_invoice_id,omitempty"`
	Status                             string `json:"status,omitempty"`
	ChangedAt                          string `json:"changed_at,omitempty"`
	ChangedByID                        string `json:"changed_by_id,omitempty"`
	Comment                            string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchInvoiceStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch invoice status changes",
		Long: `List organization invoices batch invoice status changes with filtering and pagination.

Output Columns:
  ID            Status change identifier
  BATCH INVOICE Organization invoices batch invoice ID
  STATUS        Batch invoice status
  CHANGED AT    Status change timestamp
  CHANGED BY    User who changed the status
  COMMENT       Status change comment

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  --organization-invoices-batch-invoice  Filter by organization invoices batch invoice ID
  --status                               Filter by status (successful, failed, skipped)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List status changes
  xbe view organization-invoices-batch-invoice-status-changes list

  # Filter by batch invoice
  xbe view organization-invoices-batch-invoice-status-changes list --organization-invoices-batch-invoice 123

  # Filter by status
  xbe view organization-invoices-batch-invoice-status-changes list --status successful

  # Output as JSON
  xbe view organization-invoices-batch-invoice-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchInvoiceStatusChangesList,
	}
	initOrganizationInvoicesBatchInvoiceStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchInvoiceStatusChangesCmd.AddCommand(newOrganizationInvoicesBatchInvoiceStatusChangesListCmd())
}

func initOrganizationInvoicesBatchInvoiceStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-invoices-batch-invoice", "", "Filter by organization invoices batch invoice ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchInvoiceStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchInvoiceStatusChangesListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-invoice-status-changes]", "organization-invoices-batch-invoice,status,changed-at,comment,changed-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization-invoices-batch-invoice]", opts.OrganizationInvoicesBatchInvoice)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-invoice-status-changes", query)
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

	rows := buildOrganizationInvoicesBatchInvoiceStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchInvoiceStatusChangesTable(cmd, rows)
}

func parseOrganizationInvoicesBatchInvoiceStatusChangesListOptions(cmd *cobra.Command) (organizationInvoicesBatchInvoiceStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organizationInvoicesBatchInvoice, _ := cmd.Flags().GetString("organization-invoices-batch-invoice")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchInvoiceStatusChangesListOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		NoAuth:                           noAuth,
		Limit:                            limit,
		Offset:                           offset,
		Sort:                             sort,
		OrganizationInvoicesBatchInvoice: organizationInvoicesBatchInvoice,
		Status:                           status,
	}, nil
}

func buildOrganizationInvoicesBatchInvoiceStatusChangeRows(resp jsonAPIResponse) []organizationInvoicesBatchInvoiceStatusChangeRow {
	rows := make([]organizationInvoicesBatchInvoiceStatusChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := organizationInvoicesBatchInvoiceStatusChangeRow{
			ID:        resource.ID,
			Status:    stringAttr(resource.Attributes, "status"),
			ChangedAt: formatDateTime(stringAttr(resource.Attributes, "changed-at")),
			Comment:   stringAttr(resource.Attributes, "comment"),
		}
		if rel, ok := resource.Relationships["organization-invoices-batch-invoice"]; ok && rel.Data != nil {
			row.OrganizationInvoicesBatchInvoiceID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
			row.ChangedByID = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderOrganizationInvoicesBatchInvoiceStatusChangesTable(cmd *cobra.Command, rows []organizationInvoicesBatchInvoiceStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch invoice status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBATCH INVOICE\tSTATUS\tCHANGED AT\tCHANGED BY\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.OrganizationInvoicesBatchInvoiceID,
			row.Status,
			row.ChangedAt,
			row.ChangedByID,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
