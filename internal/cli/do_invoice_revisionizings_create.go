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

type doInvoiceRevisionizingsCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	InvoiceID string
	Comment   string
	InBulk    bool
}

type invoiceRevisionizingRow struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoice_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newDoInvoiceRevisionizingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Revise an invoice",
		Long: `Revise an invoice.

This action transitions the invoice status to revised.
Only revisionable, exported, or non-exportable invoices can be revised.
Bulk revisionizing is required by the API.

Required flags:
  --invoice   Invoice ID
  --comment   Comment for the revision action
  --in-bulk   Confirm bulk revisionizing`,
		Example: `  # Revise an invoice (bulk required)
  xbe do invoice-revisionizings create --invoice 123 --comment "Bulk revision" --in-bulk

  # JSON output
  xbe do invoice-revisionizings create --invoice 123 --comment "Bulk revision" --in-bulk --json`,
		RunE: runDoInvoiceRevisionizingsCreate,
	}
	initDoInvoiceRevisionizingsCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoiceRevisionizingsCmd.AddCommand(newDoInvoiceRevisionizingsCreateCmd())
}

func initDoInvoiceRevisionizingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice", "", "Invoice ID (required)")
	cmd.Flags().String("comment", "", "Comment for the revision action (required)")
	cmd.Flags().Bool("in-bulk", false, "Confirm bulk revisionizing (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("invoice")
	cmd.MarkFlagRequired("comment")
	cmd.MarkFlagRequired("in-bulk")
}

func runDoInvoiceRevisionizingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoiceRevisionizingsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !opts.InBulk {
		err := fmt.Errorf("in-bulk must be true for invoice revisionizing")
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

	attributes := map[string]any{
		"comment": opts.Comment,
		"in-bulk": true,
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
			"type":          "invoice-revisionizings",
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

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-revisionizings", jsonBody)
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

	row := buildInvoiceRevisionizingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created invoice revisionizing %s\n", row.ID)
	return nil
}

func parseDoInvoiceRevisionizingsCreateOptions(cmd *cobra.Command) (doInvoiceRevisionizingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoiceID, _ := cmd.Flags().GetString("invoice")
	comment, _ := cmd.Flags().GetString("comment")
	inBulk, _ := cmd.Flags().GetBool("in-bulk")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoiceRevisionizingsCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		InvoiceID: invoiceID,
		Comment:   comment,
		InBulk:    inBulk,
	}, nil
}

func buildInvoiceRevisionizingRowFromSingle(resp jsonAPISingleResponse) invoiceRevisionizingRow {
	resource := resp.Data
	row := invoiceRevisionizingRow{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}
	return row
}
