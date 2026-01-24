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

type organizationInvoicesBatchStatusChangesListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	OrganizationInvoicesBatch string
	Status                    string
}

type organizationInvoicesBatchStatusChangeRow struct {
	ID                        string `json:"id"`
	OrganizationInvoicesBatch string `json:"organization_invoices_batch_id,omitempty"`
	Status                    string `json:"status,omitempty"`
	ChangedAt                 string `json:"changed_at,omitempty"`
	ChangedBy                 string `json:"changed_by_id,omitempty"`
	Comment                   string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch status changes",
		Long: `List organization invoices batch status changes.

Output Columns:
  ID         Status change identifier
  STATUS     Batch status
  CHANGED AT When the status changed
  CHANGED BY User who made the change
  BATCH      Organization invoices batch ID
  COMMENT    Comment (truncated)

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization invoices batch status changes
  xbe view organization-invoices-batch-status-changes list

  # Filter by batch
  xbe view organization-invoices-batch-status-changes list --organization-invoices-batch 123

  # Filter by status
  xbe view organization-invoices-batch-status-changes list --status processed

  # JSON output
  xbe view organization-invoices-batch-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchStatusChangesList,
	}
	initOrganizationInvoicesBatchStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchStatusChangesCmd.AddCommand(newOrganizationInvoicesBatchStatusChangesListCmd())
}

func initOrganizationInvoicesBatchStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-invoices-batch", "", "Filter by organization invoices batch ID")
	cmd.Flags().String("status", "", "Filter by status (not_processed/processed)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchStatusChangesListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-status-changes]", "status,changed-at,comment,changed-by,organization-invoices-batch")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization-invoices-batch]", opts.OrganizationInvoicesBatch)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-status-changes", query)
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

	rows := buildOrganizationInvoicesBatchStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchStatusChangesTable(cmd, rows)
}

func parseOrganizationInvoicesBatchStatusChangesListOptions(cmd *cobra.Command) (organizationInvoicesBatchStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organizationInvoicesBatch, _ := cmd.Flags().GetString("organization-invoices-batch")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchStatusChangesListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		OrganizationInvoicesBatch: organizationInvoicesBatch,
		Status:                    status,
	}, nil
}

func buildOrganizationInvoicesBatchStatusChangeRows(resp jsonAPIResponse) []organizationInvoicesBatchStatusChangeRow {
	rows := make([]organizationInvoicesBatchStatusChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildOrganizationInvoicesBatchStatusChangeRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildOrganizationInvoicesBatchStatusChangeRow(resource jsonAPIResource) organizationInvoicesBatchStatusChangeRow {
	attrs := resource.Attributes
	row := organizationInvoicesBatchStatusChangeRow{
		ID:        resource.ID,
		Status:    stringAttr(attrs, "status"),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["organization-invoices-batch"]; ok && rel.Data != nil {
		row.OrganizationInvoicesBatch = rel.Data.ID
	}
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		row.ChangedBy = rel.Data.ID
	}

	return row
}

func renderOrganizationInvoicesBatchStatusChangesTable(cmd *cobra.Command, rows []organizationInvoicesBatchStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tCHANGED AT\tCHANGED BY\tBATCH\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.ChangedAt,
			row.ChangedBy,
			row.OrganizationInvoicesBatch,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
