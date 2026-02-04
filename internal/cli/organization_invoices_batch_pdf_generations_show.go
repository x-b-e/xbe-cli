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

type organizationInvoicesBatchPdfGenerationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchPdfGenerationDetails struct {
	ID                          string   `json:"id"`
	Status                      string   `json:"status,omitempty"`
	GenerationMetadata          any      `json:"generation_metadata,omitempty"`
	CreatedAt                   string   `json:"created_at,omitempty"`
	OrganizationInvoicesBatchID string   `json:"organization_invoices_batch_id,omitempty"`
	OrganizationPdfTemplateID   string   `json:"organization_pdf_template_id,omitempty"`
	CreatedByID                 string   `json:"created_by_id,omitempty"`
	PdfFileIDs                  []string `json:"pdf_file_ids,omitempty"`
}

func newOrganizationInvoicesBatchPdfGenerationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch PDF generation details",
		Long: `Show the full details of an organization invoices batch PDF generation.

Output Fields:
  ID
  Status
  Generation Metadata
  Created At
  Organization Invoices Batch ID
  Organization PDF Template ID
  Created By
  PDF File IDs

Arguments:
  <id>    The PDF generation ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a PDF generation
  xbe view organization-invoices-batch-pdf-generations show 123

  # JSON output
  xbe view organization-invoices-batch-pdf-generations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchPdfGenerationsShow,
	}
	initOrganizationInvoicesBatchPdfGenerationsShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchPdfGenerationsCmd.AddCommand(newOrganizationInvoicesBatchPdfGenerationsShowCmd())
}

func initOrganizationInvoicesBatchPdfGenerationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchPdfGenerationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseOrganizationInvoicesBatchPdfGenerationsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("organization invoices batch PDF generation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batch-pdf-generations]", "status,generation-metadata,created-at,organization-invoices-batch,organization-pdf-template,created-by,pdf-files")

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-pdf-generations/"+id, query)
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

	details := buildOrganizationInvoicesBatchPdfGenerationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchPdfGenerationDetails(cmd, details)
}

func parseOrganizationInvoicesBatchPdfGenerationsShowOptions(cmd *cobra.Command) (organizationInvoicesBatchPdfGenerationsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return organizationInvoicesBatchPdfGenerationsShowOptions{}, err
	}

	return organizationInvoicesBatchPdfGenerationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchPdfGenerationDetails(resp jsonAPISingleResponse) organizationInvoicesBatchPdfGenerationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return organizationInvoicesBatchPdfGenerationDetails{
		ID:                          resource.ID,
		Status:                      stringAttr(attrs, "status"),
		GenerationMetadata:          attrs["generation-metadata"],
		CreatedAt:                   formatDateTime(stringAttr(attrs, "created-at")),
		OrganizationInvoicesBatchID: relationshipIDFromMap(resource.Relationships, "organization-invoices-batch"),
		OrganizationPdfTemplateID:   relationshipIDFromMap(resource.Relationships, "organization-pdf-template"),
		CreatedByID:                 relationshipIDFromMap(resource.Relationships, "created-by"),
		PdfFileIDs:                  relationshipIDsFromMap(resource.Relationships, "pdf-files"),
	}
}

func renderOrganizationInvoicesBatchPdfGenerationDetails(cmd *cobra.Command, details organizationInvoicesBatchPdfGenerationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.OrganizationInvoicesBatchID != "" {
		fmt.Fprintf(out, "Organization Invoices Batch: %s\n", details.OrganizationInvoicesBatchID)
	}
	if details.OrganizationPdfTemplateID != "" {
		fmt.Fprintf(out, "Organization PDF Template: %s\n", details.OrganizationPdfTemplateID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if len(details.PdfFileIDs) > 0 {
		fmt.Fprintf(out, "PDF File IDs: %s\n", strings.Join(details.PdfFileIDs, ", "))
	}
	if details.GenerationMetadata != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Generation Metadata:")
		fmt.Fprintln(out, formatJSONBlock(details.GenerationMetadata, "  "))
	}

	return nil
}
