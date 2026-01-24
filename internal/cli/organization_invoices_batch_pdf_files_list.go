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

type organizationInvoicesBatchPdfFilesListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Sort            string
	PDFGeneration   string
	InvoiceRevision string
	Status          string
	CreatedAtMin    string
	CreatedAtMax    string
}

type organizationInvoicesBatchPdfFileRow struct {
	ID                string `json:"id"`
	Status            string `json:"status,omitempty"`
	FileName          string `json:"file_name,omitempty"`
	MimeType          string `json:"mime_type,omitempty"`
	PDFGenerationID   string `json:"pdf_generation_id,omitempty"`
	InvoiceRevisionID string `json:"invoice_revision_id,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
}

func newOrganizationInvoicesBatchPdfFilesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch PDF files",
		Long: `List organization invoices batch PDF files.

Output Columns:
  ID               PDF file identifier
  STATUS           Generation status
  FILE NAME        File name (truncated)
  PDF GENERATION   Organization invoices batch PDF generation ID
  INVOICE REVISION Invoice revision ID

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization invoices batch PDF files
  xbe view organization-invoices-batch-pdf-files list

  # Filter by PDF generation
  xbe view organization-invoices-batch-pdf-files list --pdf-generation 123

  # Filter by invoice revision
  xbe view organization-invoices-batch-pdf-files list --invoice-revision 456

  # Filter by status
  xbe view organization-invoices-batch-pdf-files list --status completed

  # Filter by created-at range
  xbe view organization-invoices-batch-pdf-files list --created-at-min 2025-01-01T00:00:00Z --created-at-max 2025-01-31T23:59:59Z

  # JSON output
  xbe view organization-invoices-batch-pdf-files list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchPdfFilesList,
	}
	initOrganizationInvoicesBatchPdfFilesListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchPdfFilesCmd.AddCommand(newOrganizationInvoicesBatchPdfFilesListCmd())
}

func initOrganizationInvoicesBatchPdfFilesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("pdf-generation", "", "Filter by PDF generation ID")
	cmd.Flags().String("invoice-revision", "", "Filter by invoice revision ID")
	cmd.Flags().String("status", "", "Filter by status (pending/completed/failed)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchPdfFilesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchPdfFilesListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-pdf-files]", "file-name,mime-type,status,created-at,pdf-generation,invoice-revision")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[pdf-generation]", opts.PDFGeneration)
	setFilterIfPresent(query, "filter[invoice-revision]", opts.InvoiceRevision)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-pdf-files", query)
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

	rows := buildOrganizationInvoicesBatchPdfFileRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchPdfFilesTable(cmd, rows)
}

func parseOrganizationInvoicesBatchPdfFilesListOptions(cmd *cobra.Command) (organizationInvoicesBatchPdfFilesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	pdfGeneration, _ := cmd.Flags().GetString("pdf-generation")
	invoiceRevision, _ := cmd.Flags().GetString("invoice-revision")
	status, _ := cmd.Flags().GetString("status")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchPdfFilesListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Sort:            sort,
		PDFGeneration:   pdfGeneration,
		InvoiceRevision: invoiceRevision,
		Status:          status,
		CreatedAtMin:    createdAtMin,
		CreatedAtMax:    createdAtMax,
	}, nil
}

func buildOrganizationInvoicesBatchPdfFileRows(resp jsonAPIResponse) []organizationInvoicesBatchPdfFileRow {
	rows := make([]organizationInvoicesBatchPdfFileRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildOrganizationInvoicesBatchPdfFileRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildOrganizationInvoicesBatchPdfFileRow(resource jsonAPIResource) organizationInvoicesBatchPdfFileRow {
	attrs := resource.Attributes
	row := organizationInvoicesBatchPdfFileRow{
		ID:        resource.ID,
		Status:    stringAttr(attrs, "status"),
		FileName:  strings.TrimSpace(stringAttr(attrs, "file-name")),
		MimeType:  strings.TrimSpace(stringAttr(attrs, "mime-type")),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resource.Relationships["pdf-generation"]; ok && rel.Data != nil {
		row.PDFGenerationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["invoice-revision"]; ok && rel.Data != nil {
		row.InvoiceRevisionID = rel.Data.ID
	}

	return row
}

func renderOrganizationInvoicesBatchPdfFilesTable(cmd *cobra.Command, rows []organizationInvoicesBatchPdfFileRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch PDF files found.")
		return nil
	}

	const fileNameMax = 50

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tFILE NAME\tPDF GENERATION\tINVOICE REVISION")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.FileName, fileNameMax),
			row.PDFGenerationID,
			row.InvoiceRevisionID,
		)
	}

	return writer.Flush()
}
