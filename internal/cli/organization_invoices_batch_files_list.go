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

type organizationInvoicesBatchFilesListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	OrganizationInvoicesBatch string
}

type organizationInvoicesBatchFileRow struct {
	ID                          string `json:"id"`
	Status                      string `json:"status,omitempty"`
	FileName                    string `json:"file_name,omitempty"`
	OrganizationInvoicesBatchID string `json:"organization_invoices_batch_id,omitempty"`
	OrganizationFormatterID     string `json:"organization_formatter_id,omitempty"`
	CreatedByID                 string `json:"created_by_id,omitempty"`
}

func newOrganizationInvoicesBatchFilesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch files",
		Long: `List organization invoices batch files with filtering and pagination.

Output Columns:
  ID             Organization invoices batch file identifier
  STATUS         Processing status
  FILE NAME      Generated file name
  BATCH          Organization invoices batch ID
  FORMATTER      Organization formatter ID
  CREATED BY     Creator user ID

Filters:
  --organization-invoices-batch  Filter by organization invoices batch ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization invoices batch files
  xbe view organization-invoices-batch-files list

  # Filter by batch
  xbe view organization-invoices-batch-files list --organization-invoices-batch 123

  # Output as JSON
  xbe view organization-invoices-batch-files list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchFilesList,
	}
	initOrganizationInvoicesBatchFilesListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchFilesCmd.AddCommand(newOrganizationInvoicesBatchFilesListCmd())
}

func initOrganizationInvoicesBatchFilesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-invoices-batch", "", "Filter by organization invoices batch ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchFilesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchFilesListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-files]", "status,file-name,organization-invoices-batch,organization-formatter,created-by")

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

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-files", query)
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

	rows := buildOrganizationInvoicesBatchFileRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchFilesTable(cmd, rows)
}

func parseOrganizationInvoicesBatchFilesListOptions(cmd *cobra.Command) (organizationInvoicesBatchFilesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organizationInvoicesBatch, _ := cmd.Flags().GetString("organization-invoices-batch")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchFilesListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		OrganizationInvoicesBatch: organizationInvoicesBatch,
	}, nil
}

func buildOrganizationInvoicesBatchFileRows(resp jsonAPIResponse) []organizationInvoicesBatchFileRow {
	rows := make([]organizationInvoicesBatchFileRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildOrganizationInvoicesBatchFileRow(resource))
	}
	return rows
}

func buildOrganizationInvoicesBatchFileRow(resource jsonAPIResource) organizationInvoicesBatchFileRow {
	attrs := resource.Attributes
	return organizationInvoicesBatchFileRow{
		ID:                          resource.ID,
		Status:                      stringAttr(attrs, "status"),
		FileName:                    stringAttr(attrs, "file-name"),
		OrganizationInvoicesBatchID: relationshipIDFromMap(resource.Relationships, "organization-invoices-batch"),
		OrganizationFormatterID:     relationshipIDFromMap(resource.Relationships, "organization-formatter"),
		CreatedByID:                 relationshipIDFromMap(resource.Relationships, "created-by"),
	}
}

func renderOrganizationInvoicesBatchFilesTable(cmd *cobra.Command, rows []organizationInvoicesBatchFileRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch files found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tFILE NAME\tBATCH\tFORMATTER\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 12),
			truncateString(row.FileName, 30),
			truncateString(row.OrganizationInvoicesBatchID, 14),
			truncateString(row.OrganizationFormatterID, 14),
			truncateString(row.CreatedByID, 14),
		)
	}

	return writer.Flush()
}
