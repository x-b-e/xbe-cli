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

type organizationInvoicesBatchProcessesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type organizationInvoicesBatchProcessRow struct {
	ID                        string `json:"id"`
	OrganizationInvoicesBatch string `json:"organization_invoices_batch_id,omitempty"`
	Comment                   string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchProcessesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch processes",
		Long: `List organization invoices batch processes.

Output Columns:
  ID       Process identifier
  BATCH    Organization invoices batch ID
  COMMENT  Comment (truncated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization invoices batch processes
  xbe view organization-invoices-batch-processes list

  # JSON output
  xbe view organization-invoices-batch-processes list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchProcessesList,
	}
	initOrganizationInvoicesBatchProcessesListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchProcessesCmd.AddCommand(newOrganizationInvoicesBatchProcessesListCmd())
}

func initOrganizationInvoicesBatchProcessesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchProcessesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchProcessesListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-processes]", "organization-invoices-batch,comment")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, status, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-processes", query)
	if err != nil {
		if status == http.StatusNotFound {
			return renderOrganizationInvoicesBatchProcessesUnavailable(cmd, opts.JSON)
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

	rows := buildOrganizationInvoicesBatchProcessRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchProcessesTable(cmd, rows)
}

func renderOrganizationInvoicesBatchProcessesUnavailable(cmd *cobra.Command, jsonOut bool) error {
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), []organizationInvoicesBatchProcessRow{})
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Organization invoices batch processes are write-only; list is not available.")
	return nil
}

func parseOrganizationInvoicesBatchProcessesListOptions(cmd *cobra.Command) (organizationInvoicesBatchProcessesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchProcessesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildOrganizationInvoicesBatchProcessRows(resp jsonAPIResponse) []organizationInvoicesBatchProcessRow {
	rows := make([]organizationInvoicesBatchProcessRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildOrganizationInvoicesBatchProcessRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildOrganizationInvoicesBatchProcessRow(resource jsonAPIResource) organizationInvoicesBatchProcessRow {
	attrs := resource.Attributes
	row := organizationInvoicesBatchProcessRow{
		ID:      resource.ID,
		Comment: strings.TrimSpace(stringAttr(attrs, "comment")),
	}

	if rel, ok := resource.Relationships["organization-invoices-batch"]; ok && rel.Data != nil {
		row.OrganizationInvoicesBatch = rel.Data.ID
	}

	return row
}

func renderOrganizationInvoicesBatchProcessesTable(cmd *cobra.Command, rows []organizationInvoicesBatchProcessRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch processes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBATCH\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.OrganizationInvoicesBatch,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
