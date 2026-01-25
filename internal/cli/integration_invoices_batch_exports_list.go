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

type integrationInvoicesBatchExportsListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	OrganizationInvoicesBatch     string
	OrganizationInvoicesBatchFile string
	IntegrationExport             string
}

type integrationInvoicesBatchExportRow struct {
	ID                              string `json:"id"`
	OrganizationInvoicesBatchID     string `json:"organization_invoices_batch_id,omitempty"`
	OrganizationInvoicesBatchFileID string `json:"organization_invoices_batch_file_id,omitempty"`
	IntegrationExportID             string `json:"integration_export_id,omitempty"`
}

func newIntegrationInvoicesBatchExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List integration invoices batch exports",
		Long: `List integration invoices batch exports with filtering and pagination.

Output Columns:
  ID           Integration invoices batch export identifier
  BATCH        Organization invoices batch ID
  BATCH FILE   Organization invoices batch file ID
  EXPORT       Integration export ID

Filters:
  --organization-invoices-batch       Filter by organization invoices batch ID
  --organization-invoices-batch-file  Filter by organization invoices batch file ID
  --integration-export                Filter by integration export ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List integration invoices batch exports
  xbe view integration-invoices-batch-exports list

  # Filter by batch
  xbe view integration-invoices-batch-exports list --organization-invoices-batch 123

  # Filter by integration export
  xbe view integration-invoices-batch-exports list --integration-export 456

  # Output as JSON
  xbe view integration-invoices-batch-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runIntegrationInvoicesBatchExportsList,
	}
	initIntegrationInvoicesBatchExportsListFlags(cmd)
	return cmd
}

func init() {
	integrationInvoicesBatchExportsCmd.AddCommand(newIntegrationInvoicesBatchExportsListCmd())
}

func initIntegrationInvoicesBatchExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-invoices-batch", "", "Filter by organization invoices batch ID")
	cmd.Flags().String("organization-invoices-batch-file", "", "Filter by organization invoices batch file ID")
	cmd.Flags().String("integration-export", "", "Filter by integration export ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIntegrationInvoicesBatchExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIntegrationInvoicesBatchExportsListOptions(cmd)
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
	query.Set("fields[integration-invoices-batch-exports]", "organization-invoices-batch,organization-invoices-batch-file,integration-export")

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
	setFilterIfPresent(query, "filter[organization-invoices-batch-file]", opts.OrganizationInvoicesBatchFile)
	setFilterIfPresent(query, "filter[integration-export]", opts.IntegrationExport)

	body, _, err := client.Get(cmd.Context(), "/v1/integration-invoices-batch-exports", query)
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

	rows := buildIntegrationInvoicesBatchExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIntegrationInvoicesBatchExportsTable(cmd, rows)
}

func parseIntegrationInvoicesBatchExportsListOptions(cmd *cobra.Command) (integrationInvoicesBatchExportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organizationInvoicesBatch, _ := cmd.Flags().GetString("organization-invoices-batch")
	organizationInvoicesBatchFile, _ := cmd.Flags().GetString("organization-invoices-batch-file")
	integrationExport, _ := cmd.Flags().GetString("integration-export")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return integrationInvoicesBatchExportsListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		OrganizationInvoicesBatch:     organizationInvoicesBatch,
		OrganizationInvoicesBatchFile: organizationInvoicesBatchFile,
		IntegrationExport:             integrationExport,
	}, nil
}

func buildIntegrationInvoicesBatchExportRows(resp jsonAPIResponse) []integrationInvoicesBatchExportRow {
	rows := make([]integrationInvoicesBatchExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildIntegrationInvoicesBatchExportRow(resource))
	}
	return rows
}

func buildIntegrationInvoicesBatchExportRow(resource jsonAPIResource) integrationInvoicesBatchExportRow {
	return integrationInvoicesBatchExportRow{
		ID:                              resource.ID,
		OrganizationInvoicesBatchID:     relationshipIDFromMap(resource.Relationships, "organization-invoices-batch"),
		OrganizationInvoicesBatchFileID: relationshipIDFromMap(resource.Relationships, "organization-invoices-batch-file"),
		IntegrationExportID:             relationshipIDFromMap(resource.Relationships, "integration-export"),
	}
}

func renderIntegrationInvoicesBatchExportsTable(cmd *cobra.Command, rows []integrationInvoicesBatchExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No integration invoices batch exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tBATCH\tBATCH FILE\tEXPORT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.OrganizationInvoicesBatchID, 14),
			truncateString(row.OrganizationInvoicesBatchFileID, 14),
			truncateString(row.IntegrationExportID, 14),
		)
	}

	return writer.Flush()
}
