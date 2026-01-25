package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doInvoicePdfEmailsCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Invoice      string
	EmailAddress string
}

type invoicePdfEmailRow struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email_address,omitempty"`
	InvoiceID    string `json:"invoice_id,omitempty"`
}

func newDoInvoicePdfEmailsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Email an invoice PDF",
		Long: `Email an invoice PDF.

The invoice PDF is sent asynchronously to the specified email address.

Required flags:
  --invoice        Invoice ID
  --email-address  Recipient email address

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Email an invoice PDF
  xbe do invoice-pdf-emails create \
    --invoice 123 \
    --email-address "ap@example.com"

  # JSON output
  xbe do invoice-pdf-emails create \
    --invoice 123 \
    --email-address "ap@example.com" \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoInvoicePdfEmailsCreate,
	}
	initDoInvoicePdfEmailsCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoicePdfEmailsCmd.AddCommand(newDoInvoicePdfEmailsCreateCmd())
}

func initDoInvoicePdfEmailsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice", "", "Invoice ID")
	cmd.Flags().String("email-address", "", "Recipient email address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoInvoicePdfEmailsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoicePdfEmailsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.Invoice) == "" {
		err := fmt.Errorf("--invoice is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.EmailAddress) == "" {
		err := fmt.Errorf("--email-address is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"email-address": opts.EmailAddress,
	}

	relationships := map[string]any{
		"invoice": map[string]any{
			"data": map[string]any{
				"type": "invoices",
				"id":   opts.Invoice,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "invoice-pdf-emails",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-pdf-emails", jsonBody)
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

	row := buildInvoicePdfEmailRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	email := row.EmailAddress
	if email == "" {
		email = opts.EmailAddress
	}
	invoiceID := row.InvoiceID
	if invoiceID == "" {
		invoiceID = opts.Invoice
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Queued invoice PDF email to %s for invoice %s\n", email, invoiceID)
	return nil
}

func parseDoInvoicePdfEmailsCreateOptions(cmd *cobra.Command) (doInvoicePdfEmailsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoice, _ := cmd.Flags().GetString("invoice")
	emailAddress, _ := cmd.Flags().GetString("email-address")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoicePdfEmailsCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Invoice:      invoice,
		EmailAddress: emailAddress,
	}, nil
}

func buildInvoicePdfEmailRow(resource jsonAPIResource) invoicePdfEmailRow {
	row := invoicePdfEmailRow{ID: resource.ID}
	row.EmailAddress = stringAttr(resource.Attributes, "email-address")
	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}
	return row
}
