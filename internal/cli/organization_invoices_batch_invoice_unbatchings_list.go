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

type organizationInvoicesBatchInvoiceUnbatchingsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type organizationInvoicesBatchInvoiceUnbatchingRow struct {
	ID                                 string `json:"id"`
	OrganizationInvoicesBatchInvoiceID string `json:"organization_invoices_batch_invoice_id,omitempty"`
	Comment                            string `json:"comment,omitempty"`
}

func newOrganizationInvoicesBatchInvoiceUnbatchingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch invoice unbatchings",
		Long: `List organization invoices batch invoice unbatchings.

Output Columns:
  ID             Unbatching identifier
  BATCH INVOICE  Organization invoices batch invoice ID
  COMMENT        Unbatching comment

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization invoices batch invoice unbatchings
  xbe view organization-invoices-batch-invoice-unbatchings list

  # Output as JSON
  xbe view organization-invoices-batch-invoice-unbatchings list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchInvoiceUnbatchingsList,
	}
	initOrganizationInvoicesBatchInvoiceUnbatchingsListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchInvoiceUnbatchingsCmd.AddCommand(newOrganizationInvoicesBatchInvoiceUnbatchingsListCmd())
}

func initOrganizationInvoicesBatchInvoiceUnbatchingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchInvoiceUnbatchingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchInvoiceUnbatchingsListOptions(cmd)
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
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-invoice-unbatchings", query)
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

	rows := buildOrganizationInvoicesBatchInvoiceUnbatchingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchInvoiceUnbatchingsTable(cmd, rows)
}

func parseOrganizationInvoicesBatchInvoiceUnbatchingsListOptions(cmd *cobra.Command) (organizationInvoicesBatchInvoiceUnbatchingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchInvoiceUnbatchingsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildOrganizationInvoicesBatchInvoiceUnbatchingRows(resp jsonAPIResponse) []organizationInvoicesBatchInvoiceUnbatchingRow {
	rows := make([]organizationInvoicesBatchInvoiceUnbatchingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildOrganizationInvoicesBatchInvoiceUnbatchingRow(resource))
	}
	return rows
}

func buildOrganizationInvoicesBatchInvoiceUnbatchingRow(resource jsonAPIResource) organizationInvoicesBatchInvoiceUnbatchingRow {
	return organizationInvoicesBatchInvoiceUnbatchingRow{
		ID:                                 resource.ID,
		OrganizationInvoicesBatchInvoiceID: relationshipIDFromMap(resource.Relationships, "organization-invoices-batch-invoice"),
		Comment:                            stringAttr(resource.Attributes, "comment"),
	}
}

func renderOrganizationInvoicesBatchInvoiceUnbatchingsTable(cmd *cobra.Command, rows []organizationInvoicesBatchInvoiceUnbatchingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch invoice unbatchings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBATCH INVOICE\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.OrganizationInvoicesBatchInvoiceID,
			truncateString(row.Comment, 40),
		)
	}
	return writer.Flush()
}
