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

type doInvoiceExportsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	Invoice string
	Comment string
}

type invoiceExportRow struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoice_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newDoInvoiceExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Export an invoice",
		Long: `Export an invoice.

Required flags:
  --invoice  Invoice ID (required)

Optional flags:
  --comment  Export comment`,
		Example: `  # Export an invoice
  xbe do invoice-exports create --invoice 123 --comment "Sent to accounting"`,
		Args: cobra.NoArgs,
		RunE: runDoInvoiceExportsCreate,
	}
	initDoInvoiceExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoiceExportsCmd.AddCommand(newDoInvoiceExportsCreateCmd())
}

func initDoInvoiceExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice", "", "Invoice ID (required)")
	cmd.Flags().String("comment", "", "Export comment")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoInvoiceExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoiceExportsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Invoice) == "" {
		err := fmt.Errorf("--invoice is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
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
			"type":          "invoice-exports",
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

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-exports", jsonBody)
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

	row := invoiceExportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created invoice export %s\n", row.ID)
	return nil
}

func invoiceExportRowFromSingle(resp jsonAPISingleResponse) invoiceExportRow {
	attrs := resp.Data.Attributes
	row := invoiceExportRow{
		ID:      resp.Data.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resp.Data.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}

	return row
}

func parseDoInvoiceExportsCreateOptions(cmd *cobra.Command) (doInvoiceExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoice, _ := cmd.Flags().GetString("invoice")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoiceExportsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		Invoice: invoice,
		Comment: comment,
	}, nil
}
