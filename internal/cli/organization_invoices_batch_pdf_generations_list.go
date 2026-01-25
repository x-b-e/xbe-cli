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

type organizationInvoicesBatchPdfGenerationsListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	OrganizationInvoicesBatch string
	Status                    string
	CreatedBy                 string
	CreatedAtMin              string
	CreatedAtMax              string
}

type organizationInvoicesBatchPdfGenerationRow struct {
	ID                          string `json:"id"`
	Status                      string `json:"status,omitempty"`
	OrganizationInvoicesBatchID string `json:"organization_invoices_batch_id,omitempty"`
	OrganizationPdfTemplateID   string `json:"organization_pdf_template_id,omitempty"`
	CreatedByID                 string `json:"created_by_id,omitempty"`
}

func newOrganizationInvoicesBatchPdfGenerationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch PDF generations",
		Long: `List organization invoices batch PDF generations with filtering and pagination.

Output Columns:
  ID        PDF generation identifier
  STATUS    Generation status
  BATCH     Organization invoices batch ID
  TEMPLATE  Organization PDF template ID
  CREATED BY Creator user ID

Filters:
  --organization-invoices-batch  Filter by organization invoices batch ID
  --status                       Filter by status
  --created-by                   Filter by creator user ID
  --created-at-min               Filter by created-at on/after (ISO 8601)
  --created-at-max               Filter by created-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List PDF generations
  xbe view organization-invoices-batch-pdf-generations list

  # Filter by batch
  xbe view organization-invoices-batch-pdf-generations list --organization-invoices-batch 123

  # Filter by status
  xbe view organization-invoices-batch-pdf-generations list --status completed

  # Output as JSON
  xbe view organization-invoices-batch-pdf-generations list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchPdfGenerationsList,
	}
	initOrganizationInvoicesBatchPdfGenerationsListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchPdfGenerationsCmd.AddCommand(newOrganizationInvoicesBatchPdfGenerationsListCmd())
}

func initOrganizationInvoicesBatchPdfGenerationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-invoices-batch", "", "Filter by organization invoices batch ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchPdfGenerationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchPdfGenerationsListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-pdf-generations]", "status,organization-invoices-batch,organization-pdf-template,created-by")

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
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-pdf-generations", query)
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

	rows := buildOrganizationInvoicesBatchPdfGenerationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchPdfGenerationsTable(cmd, rows)
}

func parseOrganizationInvoicesBatchPdfGenerationsListOptions(cmd *cobra.Command) (organizationInvoicesBatchPdfGenerationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organizationInvoicesBatch, _ := cmd.Flags().GetString("organization-invoices-batch")
	status, _ := cmd.Flags().GetString("status")
	createdBy, _ := cmd.Flags().GetString("created-by")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchPdfGenerationsListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		OrganizationInvoicesBatch: organizationInvoicesBatch,
		Status:                    status,
		CreatedBy:                 createdBy,
		CreatedAtMin:              createdAtMin,
		CreatedAtMax:              createdAtMax,
	}, nil
}

func buildOrganizationInvoicesBatchPdfGenerationRows(resp jsonAPIResponse) []organizationInvoicesBatchPdfGenerationRow {
	rows := make([]organizationInvoicesBatchPdfGenerationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildOrganizationInvoicesBatchPdfGenerationRow(resource))
	}
	return rows
}

func buildOrganizationInvoicesBatchPdfGenerationRow(resource jsonAPIResource) organizationInvoicesBatchPdfGenerationRow {
	attrs := resource.Attributes
	return organizationInvoicesBatchPdfGenerationRow{
		ID:                          resource.ID,
		Status:                      stringAttr(attrs, "status"),
		OrganizationInvoicesBatchID: relationshipIDFromMap(resource.Relationships, "organization-invoices-batch"),
		OrganizationPdfTemplateID:   relationshipIDFromMap(resource.Relationships, "organization-pdf-template"),
		CreatedByID:                 relationshipIDFromMap(resource.Relationships, "created-by"),
	}
}

func renderOrganizationInvoicesBatchPdfGenerationsTable(cmd *cobra.Command, rows []organizationInvoicesBatchPdfGenerationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch PDF generations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tBATCH\tTEMPLATE\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 12),
			truncateString(row.OrganizationInvoicesBatchID, 14),
			truncateString(row.OrganizationPdfTemplateID, 14),
			truncateString(row.CreatedByID, 14),
		)
	}

	return writer.Flush()
}
