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

type doInvoiceRejectionsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	InvoiceID string
	Comment   string
}

type invoiceRejectionRow struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoice_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newDoInvoiceRejectionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Reject a sent invoice",
		Long: `Reject a sent invoice.

This action transitions the invoice status from sent to rejected.
Only sent invoices can be rejected.

Required flags:
  --invoice   Invoice ID

Optional flags:
  --comment   Comment for the rejection action`,
		Example: `  # Reject a sent invoice
  xbe do invoice-rejections create --invoice 123 --comment "Missing documentation"

  # Reject a sent invoice (no comment)
  xbe do invoice-rejections create --invoice 123

  # JSON output
  xbe do invoice-rejections create --invoice 123 --json`,
		RunE: runDoInvoiceRejectionsCreate,
	}
	initDoInvoiceRejectionsCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoiceRejectionsCmd.AddCommand(newDoInvoiceRejectionsCreateCmd())
}

func initDoInvoiceRejectionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice", "", "Invoice ID (required)")
	cmd.Flags().String("comment", "", "Comment for the rejection action")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("invoice")
}

func runDoInvoiceRejectionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoiceRejectionsCreateOptions(cmd)
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
			"type":          "invoice-rejections",
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

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-rejections", jsonBody)
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

	row := buildInvoiceRejectionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created invoice rejection %s\n", row.ID)
	return nil
}

func parseDoInvoiceRejectionsCreateOptions(cmd *cobra.Command) (doInvoiceRejectionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoiceID, _ := cmd.Flags().GetString("invoice")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoiceRejectionsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		InvoiceID: invoiceID,
		Comment:   comment,
	}, nil
}

func buildInvoiceRejectionRowFromSingle(resp jsonAPISingleResponse) invoiceRejectionRow {
	resource := resp.Data
	row := invoiceRejectionRow{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}
	return row
}
