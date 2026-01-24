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

type organizationInvoicesBatchInvoicesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type organizationInvoicesBatchInvoiceDetails struct {
	ID                          string `json:"id"`
	Status                      string `json:"status,omitempty"`
	BatchStatus                 string `json:"batch_status,omitempty"`
	InvoiceAmount               string `json:"invoice_amount,omitempty"`
	IsRevised                   bool   `json:"is_revised"`
	BatchInvoiceRevision        any    `json:"batch_invoice_revision,omitempty"`
	InvoiceID                   string `json:"invoice_id,omitempty"`
	InvoiceType                 string `json:"invoice_type,omitempty"`
	OrganizationInvoicesBatchID string `json:"organization_invoices_batch_id,omitempty"`
	OrganizationID              string `json:"organization_id,omitempty"`
	OrganizationType            string `json:"organization_type,omitempty"`
	CreatedByID                 string `json:"created_by_id,omitempty"`
	UpdatedByID                 string `json:"updated_by_id,omitempty"`
	ChangedByID                 string `json:"changed_by_id,omitempty"`
}

func newOrganizationInvoicesBatchInvoicesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show organization invoices batch invoice details",
		Long: `Show the full details of an organization invoices batch invoice.

Output Fields:
  ID                     Batch invoice identifier
  STATUS                 Batch invoice status
  BATCH STATUS           Batch status for the invoice
  INVOICE AMOUNT         Invoice amount from the revision
  IS REVISED             Whether the invoice revision differs
  BATCH INVOICE REVISION Invoice revision snapshot
  INVOICE                Invoice (Type/ID)
  BATCH                  Organization invoices batch ID
  ORGANIZATION           Organization (Type/ID)
  CREATED BY             User who created the batch invoice
  UPDATED BY             User who last updated the batch invoice

Arguments:
  <id>  Batch invoice ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a batch invoice
  xbe view organization-invoices-batch-invoices show 123

  # Output as JSON
  xbe view organization-invoices-batch-invoices show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOrganizationInvoicesBatchInvoicesShow,
	}
	initOrganizationInvoicesBatchInvoicesShowFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchInvoicesCmd.AddCommand(newOrganizationInvoicesBatchInvoicesShowCmd())
}

func initOrganizationInvoicesBatchInvoicesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchInvoicesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOrganizationInvoicesBatchInvoicesShowOptions(cmd)
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
		return fmt.Errorf("organization invoices batch invoice id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[organization-invoices-batch-invoices]", "status,batch-status,invoice-amount,is-revised,batch-invoice-revision,invoice,organization-invoices-batch,organization,created-by,updated-by")

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-invoices/"+id, query)
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

	details := buildOrganizationInvoicesBatchInvoiceDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOrganizationInvoicesBatchInvoiceDetails(cmd, details)
}

func parseOrganizationInvoicesBatchInvoicesShowOptions(cmd *cobra.Command) (organizationInvoicesBatchInvoicesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchInvoicesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOrganizationInvoicesBatchInvoiceDetails(resp jsonAPISingleResponse) organizationInvoicesBatchInvoiceDetails {
	attrs := resp.Data.Attributes
	details := organizationInvoicesBatchInvoiceDetails{
		ID:                   resp.Data.ID,
		Status:               stringAttr(attrs, "status"),
		BatchStatus:          stringAttr(attrs, "batch-status"),
		InvoiceAmount:        stringAttr(attrs, "invoice-amount"),
		IsRevised:            boolAttr(attrs, "is-revised"),
		BatchInvoiceRevision: anyAttr(attrs, "batch-invoice-revision"),
	}

	if rel, ok := resp.Data.Relationships["invoice"]; ok && rel.Data != nil {
		details.InvoiceID = rel.Data.ID
		details.InvoiceType = rel.Data.Type
	}
	if rel, ok := resp.Data.Relationships["organization-invoices-batch"]; ok && rel.Data != nil {
		details.OrganizationInvoicesBatchID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationID = rel.Data.ID
		details.OrganizationType = rel.Data.Type
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["updated-by"]; ok && rel.Data != nil {
		details.UpdatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["changed-by"]; ok && rel.Data != nil {
		details.ChangedByID = rel.Data.ID
	}

	return details
}

func renderOrganizationInvoicesBatchInvoiceDetails(cmd *cobra.Command, details organizationInvoicesBatchInvoiceDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.BatchStatus != "" {
		fmt.Fprintf(out, "Batch Status: %s\n", details.BatchStatus)
	}
	if details.InvoiceAmount != "" {
		fmt.Fprintf(out, "Invoice Amount: %s\n", formatAnyValue(details.InvoiceAmount))
	}
	fmt.Fprintf(out, "Is Revised: %t\n", details.IsRevised)
	if details.BatchInvoiceRevision != nil {
		if formatted := formatAnyJSON(details.BatchInvoiceRevision); formatted != "" {
			fmt.Fprintln(out, "Batch Invoice Revision:")
			fmt.Fprintln(out, formatted)
		}
	}
	if details.InvoiceType != "" || details.InvoiceID != "" {
		fmt.Fprintf(out, "Invoice: %s\n", formatPolymorphic(details.InvoiceType, details.InvoiceID))
	}
	if details.OrganizationInvoicesBatchID != "" {
		fmt.Fprintf(out, "Batch: %s\n", details.OrganizationInvoicesBatchID)
	}
	if details.OrganizationType != "" || details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization: %s\n", formatPolymorphic(details.OrganizationType, details.OrganizationID))
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.UpdatedByID != "" {
		fmt.Fprintf(out, "Updated By: %s\n", details.UpdatedByID)
	}
	if details.ChangedByID != "" {
		fmt.Fprintf(out, "Changed By: %s\n", details.ChangedByID)
	}

	return nil
}
