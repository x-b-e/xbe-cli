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

type organizationInvoicesBatchFilesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchFileDetails struct {
	ID                                  string   `json:"id"`
	Status                              string   `json:"status,omitempty"`
	FileName                            string   `json:"file_name,omitempty"`
	Body                                string   `json:"body,omitempty"`
	MimeType                            string   `json:"mime_type,omitempty"`
	RefreshInvoiceRevisions             bool     `json:"refresh_invoice_revisions,omitempty"`
	FormatterErrorsDetails              any      `json:"formatter_errors_details,omitempty"`
	InvoiceRevisionNumbers              any      `json:"invoice_revision_numbers,omitempty"`
	OrganizationInvoicesBatchVersion    string   `json:"organization_invoices_batch_version_number,omitempty"`
	OrganizationFormatterVersion        string   `json:"organization_formatter_version_number,omitempty"`
	CanBeRevised                        bool     `json:"can_be_revised,omitempty"`
	InvoicesRevised                     bool     `json:"invoices_revised,omitempty"`
	FormatterRevised                    bool     `json:"formatter_revised,omitempty"`
	BatchRevised                        bool     `json:"batch_revised,omitempty"`
	OrganizationType                    string   `json:"organization_type,omitempty"`
	OrganizationID                      string   `json:"organization_id,omitempty"`
	OrganizationInvoicesBatchID         string   `json:"organization_invoices_batch_id,omitempty"`
	OrganizationFormatterID             string   `json:"organization_formatter_id,omitempty"`
	CreatedByID                         string   `json:"created_by_id,omitempty"`
	OrganizationInvoicesBatchInvoiceIDs []string `json:"organization_invoices_batch_invoice_ids,omitempty"`
	InvoiceRevisionIDs                  []string `json:"invoice_revision_ids,omitempty"`
}

func newOrganizationInvoicesBatchFilesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch file details",
		Long: `Show the full details of an organization invoices batch file.

Output Fields:
  ID
  Status
  File Name
  Body
  Mime Type
  Refresh Invoice Revisions
  Formatter Errors Details
  Invoice Revision Numbers
  Organization Invoices Batch Version Number
  Organization Formatter Version Number
  Can Be Revised
  Invoices Revised
  Formatter Revised
  Batch Revised
  Organization (type + ID)
  Organization Invoices Batch ID
  Organization Formatter ID
  Created By
  Organization Invoices Batch Invoice IDs
  Invoice Revision IDs

Arguments:
  <id>    The organization invoices batch file ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an organization invoices batch file
  xbe view organization-invoices-batch-files show 123

  # JSON output
  xbe view organization-invoices-batch-files show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchFilesShow,
	}
	initOrganizationInvoicesBatchFilesShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchFilesCmd.AddCommand(newOrganizationInvoicesBatchFilesShowCmd())
}

func initOrganizationInvoicesBatchFilesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchFilesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOrganizationInvoicesBatchFilesShowOptions(cmd)
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
		return fmt.Errorf("organization invoices batch file id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batch-files]", "body,mime-type,refresh-invoice-revisions,status,file-name,formatter-errors-details,invoice-revision-numbers,organization-invoices-batch-version-number,organization-formatter-version-number,can-be-revised,invoices-revised,formatter-revised,batch-revised,organization,organization-invoices-batch,organization-formatter,created-by,organization-invoices-batch-invoices,invoice-revisions")

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-files/"+id, query)
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

	details := buildOrganizationInvoicesBatchFileDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchFileDetails(cmd, details)
}

func parseOrganizationInvoicesBatchFilesShowOptions(cmd *cobra.Command) (organizationInvoicesBatchFilesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return organizationInvoicesBatchFilesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return organizationInvoicesBatchFilesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return organizationInvoicesBatchFilesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return organizationInvoicesBatchFilesShowOptions{}, err
	}

	return organizationInvoicesBatchFilesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchFileDetails(resp jsonAPISingleResponse) organizationInvoicesBatchFileDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := organizationInvoicesBatchFileDetails{
		ID:                                  resource.ID,
		Status:                              stringAttr(attrs, "status"),
		FileName:                            stringAttr(attrs, "file-name"),
		Body:                                stringAttr(attrs, "body"),
		MimeType:                            stringAttr(attrs, "mime-type"),
		RefreshInvoiceRevisions:             boolAttr(attrs, "refresh-invoice-revisions"),
		FormatterErrorsDetails:              attrs["formatter-errors-details"],
		InvoiceRevisionNumbers:              attrs["invoice-revision-numbers"],
		OrganizationInvoicesBatchVersion:    stringAttr(attrs, "organization-invoices-batch-version-number"),
		OrganizationFormatterVersion:        stringAttr(attrs, "organization-formatter-version-number"),
		CanBeRevised:                        boolAttr(attrs, "can-be-revised"),
		InvoicesRevised:                     boolAttr(attrs, "invoices-revised"),
		FormatterRevised:                    boolAttr(attrs, "formatter-revised"),
		BatchRevised:                        boolAttr(attrs, "batch-revised"),
		OrganizationInvoicesBatchID:         relationshipIDFromMap(resource.Relationships, "organization-invoices-batch"),
		OrganizationFormatterID:             relationshipIDFromMap(resource.Relationships, "organization-formatter"),
		CreatedByID:                         relationshipIDFromMap(resource.Relationships, "created-by"),
		OrganizationInvoicesBatchInvoiceIDs: relationshipIDsFromMap(resource.Relationships, "organization-invoices-batch-invoices"),
		InvoiceRevisionIDs:                  relationshipIDsFromMap(resource.Relationships, "invoice-revisions"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}

	return details
}

func renderOrganizationInvoicesBatchFileDetails(cmd *cobra.Command, details organizationInvoicesBatchFileDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.FileName != "" {
		fmt.Fprintf(out, "File Name: %s\n", details.FileName)
	}
	if details.MimeType != "" {
		fmt.Fprintf(out, "Mime Type: %s\n", details.MimeType)
	}
	fmt.Fprintf(out, "Refresh Invoice Revisions: %t\n", details.RefreshInvoiceRevisions)
	fmt.Fprintf(out, "Can Be Revised: %t\n", details.CanBeRevised)
	fmt.Fprintf(out, "Invoices Revised: %t\n", details.InvoicesRevised)
	fmt.Fprintf(out, "Formatter Revised: %t\n", details.FormatterRevised)
	fmt.Fprintf(out, "Batch Revised: %t\n", details.BatchRevised)

	if details.OrganizationInvoicesBatchVersion != "" {
		fmt.Fprintf(out, "Organization Invoices Batch Version: %s\n", details.OrganizationInvoicesBatchVersion)
	}
	if details.OrganizationFormatterVersion != "" {
		fmt.Fprintf(out, "Organization Formatter Version: %s\n", details.OrganizationFormatterVersion)
	}

	if details.OrganizationType != "" || details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization: %s:%s\n", details.OrganizationType, details.OrganizationID)
	}
	if details.OrganizationInvoicesBatchID != "" {
		fmt.Fprintf(out, "Organization Invoices Batch: %s\n", details.OrganizationInvoicesBatchID)
	}
	if details.OrganizationFormatterID != "" {
		fmt.Fprintf(out, "Organization Formatter: %s\n", details.OrganizationFormatterID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}

	if details.Body != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Body:")
		fmt.Fprintln(out, details.Body)
	}
	if details.FormatterErrorsDetails != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Formatter Errors Details:")
		fmt.Fprintln(out, formatJSONBlock(details.FormatterErrorsDetails, "  "))
	}
	if details.InvoiceRevisionNumbers != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Invoice Revision Numbers:")
		fmt.Fprintln(out, formatJSONBlock(details.InvoiceRevisionNumbers, "  "))
	}
	if len(details.OrganizationInvoicesBatchInvoiceIDs) > 0 {
		fmt.Fprintf(out, "Organization Invoices Batch Invoice IDs: %s\n", strings.Join(details.OrganizationInvoicesBatchInvoiceIDs, ", "))
	}
	if len(details.InvoiceRevisionIDs) > 0 {
		fmt.Fprintf(out, "Invoice Revision IDs: %s\n", strings.Join(details.InvoiceRevisionIDs, ", "))
	}

	return nil
}
