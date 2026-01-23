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

type doInvoiceAddressesCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	InvoiceID string
	Comment   string
}

type invoiceAddressRow struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoice_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newDoInvoiceAddressesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Address a rejected invoice",
		Long: `Address a rejected invoice.

This action transitions the invoice status from rejected to addressed.
Only rejected invoices can be addressed.

Required flags:
  --invoice   Invoice ID

Optional flags:
  --comment   Comment for the address action`,
		Example: `  # Address a rejected invoice
  xbe do invoice-addresses create --invoice 123 --comment "Resolved dispute"

  # Address a rejected invoice (no comment)
  xbe do invoice-addresses create --invoice 123

  # JSON output
  xbe do invoice-addresses create --invoice 123 --json`,
		RunE: runDoInvoiceAddressesCreate,
	}
	initDoInvoiceAddressesCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoiceAddressesCmd.AddCommand(newDoInvoiceAddressesCreateCmd())
}

func initDoInvoiceAddressesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice", "", "Invoice ID (required)")
	cmd.Flags().String("comment", "", "Comment for the address action")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("invoice")
}

func runDoInvoiceAddressesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoiceAddressesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"invoice": map[string]any{
			"data": map[string]any{
				"type": "invoices",
				"id":   opts.InvoiceID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "invoice-addresses",
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

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-addresses", jsonBody)
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

	row := buildInvoiceAddressRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created invoice address %s\n", row.ID)
	return nil
}

func parseDoInvoiceAddressesCreateOptions(cmd *cobra.Command) (doInvoiceAddressesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoiceID, _ := cmd.Flags().GetString("invoice")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoiceAddressesCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		InvoiceID: invoiceID,
		Comment:   comment,
	}, nil
}

func buildInvoiceAddressRowFromSingle(resp jsonAPISingleResponse) invoiceAddressRow {
	resource := resp.Data
	row := invoiceAddressRow{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}
	return row
}
