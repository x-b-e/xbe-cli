package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type organizationInvoicesBatchPdfFilesShowOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	IncludePDFBody bool
}

type organizationInvoicesBatchPdfFileDetails struct {
	ID                string `json:"id"`
	Status            string `json:"status,omitempty"`
	FileName          string `json:"file_name,omitempty"`
	MimeType          string `json:"mime_type,omitempty"`
	PDFGenerationID   string `json:"pdf_generation_id,omitempty"`
	InvoiceRevisionID string `json:"invoice_revision_id,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	GenerationErrors  any    `json:"generation_errors,omitempty"`
	PDFBody           string `json:"pdf_body,omitempty"`
}

func newOrganizationInvoicesBatchPdfFilesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch PDF file details",
		Long: `Show full details of an organization invoices batch PDF file.

Output Fields:
  ID                 PDF file identifier
  Status             Generation status
  File Name          File name
  MIME Type          MIME type
  PDF Generation     Organization invoices batch PDF generation ID
  Invoice Revision   Invoice revision ID
  Created At         Creation timestamp
  Generation Errors  Error details (if any)
  PDF Body           Base64-encoded PDF (optional)

Arguments:
  <id>    Organization invoices batch PDF file ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an organization invoices batch PDF file
  xbe view organization-invoices-batch-pdf-files show 123

  # Include base64 PDF body in the response
  xbe view organization-invoices-batch-pdf-files show 123 --include-pdf-body

  # JSON output
  xbe view organization-invoices-batch-pdf-files show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchPdfFilesShow,
	}
	initOrganizationInvoicesBatchPdfFilesShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchPdfFilesCmd.AddCommand(newOrganizationInvoicesBatchPdfFilesShowCmd())
}

func initOrganizationInvoicesBatchPdfFilesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("include-pdf-body", false, "Include base64 PDF body in the response")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchPdfFilesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseOrganizationInvoicesBatchPdfFilesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("organization invoices batch PDF file id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batch-pdf-files]", "file-name,mime-type,status,generation-errors,created-at,pdf-generation,invoice-revision")
	if opts.IncludePDFBody {
		query.Set("meta[organization_invoices_batch_pdf_file]", "pdf_body")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-pdf-files/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildOrganizationInvoicesBatchPdfFileDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchPdfFileDetails(cmd, details)
}

func parseOrganizationInvoicesBatchPdfFilesShowOptions(cmd *cobra.Command) (organizationInvoicesBatchPdfFilesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	includePDFBody, _ := cmd.Flags().GetBool("include-pdf-body")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchPdfFilesShowOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		IncludePDFBody: includePDFBody,
	}, nil
}

func buildOrganizationInvoicesBatchPdfFileDetails(resp jsonAPISingleResponse) organizationInvoicesBatchPdfFileDetails {
	attrs := resp.Data.Attributes
	meta := resp.Data.Meta
	details := organizationInvoicesBatchPdfFileDetails{
		ID:               resp.Data.ID,
		Status:           stringAttr(attrs, "status"),
		FileName:         strings.TrimSpace(stringAttr(attrs, "file-name")),
		MimeType:         strings.TrimSpace(stringAttr(attrs, "mime-type")),
		CreatedAt:        formatDateTime(stringAttr(attrs, "created-at")),
		GenerationErrors: anyAttr(attrs, "generation-errors"),
		PDFBody:          stringAttr(meta, "pdf_body"),
	}

	if rel, ok := resp.Data.Relationships["pdf-generation"]; ok && rel.Data != nil {
		details.PDFGenerationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["invoice-revision"]; ok && rel.Data != nil {
		details.InvoiceRevisionID = rel.Data.ID
	}

	return details
}

func renderOrganizationInvoicesBatchPdfFileDetails(cmd *cobra.Command, details organizationInvoicesBatchPdfFileDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.FileName != "" {
		fmt.Fprintf(out, "File Name: %s\n", details.FileName)
	}
	if details.MimeType != "" {
		fmt.Fprintf(out, "MIME Type: %s\n", details.MimeType)
	}
	if details.PDFGenerationID != "" {
		fmt.Fprintf(out, "PDF Generation: %s\n", details.PDFGenerationID)
	}
	if details.InvoiceRevisionID != "" {
		fmt.Fprintf(out, "Invoice Revision: %s\n", details.InvoiceRevisionID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}

	formattedErrors := formatAny(details.GenerationErrors)
	if formattedErrors == "" {
		fmt.Fprintln(out, "Generation Errors: (none)")
	} else {
		fmt.Fprintln(out, "Generation Errors:")
		fmt.Fprintln(out, indentLines(formattedErrors, "  "))
	}

	if details.PDFBody != "" {
		fmt.Fprintf(out, "PDF Body (base64): %s\n", truncateString(details.PDFBody, 80))
		fmt.Fprintf(out, "PDF Body Size: %d\n", len(details.PDFBody))
	}

	return nil
}
