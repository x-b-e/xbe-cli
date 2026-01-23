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

type doInvoiceRevisionablesCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	InvoiceID string
	Comment   string
}

type invoiceRevisionableRow struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoice_id,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

func newDoInvoiceRevisionablesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Mark an invoice as revisionable",
		Long: `Mark an invoice as revisionable.

This action transitions the invoice status to revisionable.
Only exported, non-exportable, or revised invoices can be marked revisionable.

Required flags:
  --invoice   Invoice ID
  --comment   Comment for the revisionable action`,
		Example: `  # Mark an invoice as revisionable
  xbe do invoice-revisionables create --invoice 123 --comment "Needs updates"

  # JSON output
  xbe do invoice-revisionables create --invoice 123 --comment "Needs updates" --json`,
		RunE: runDoInvoiceRevisionablesCreate,
	}
	initDoInvoiceRevisionablesCreateFlags(cmd)
	return cmd
}

func init() {
	doInvoiceRevisionablesCmd.AddCommand(newDoInvoiceRevisionablesCreateCmd())
}

func initDoInvoiceRevisionablesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("invoice", "", "Invoice ID (required)")
	cmd.Flags().String("comment", "", "Comment for the revisionable action (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("invoice")
	cmd.MarkFlagRequired("comment")
}

func runDoInvoiceRevisionablesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoInvoiceRevisionablesCreateOptions(cmd)
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

	attributes := map[string]any{
		"comment": opts.Comment,
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
			"type":          "invoice-revisionables",
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

	body, _, err := client.Post(cmd.Context(), "/v1/invoice-revisionables", jsonBody)
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

	row := buildInvoiceRevisionableRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created invoice revisionable %s\n", row.ID)
	return nil
}

func parseDoInvoiceRevisionablesCreateOptions(cmd *cobra.Command) (doInvoiceRevisionablesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	invoiceID, _ := cmd.Flags().GetString("invoice")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doInvoiceRevisionablesCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		InvoiceID: invoiceID,
		Comment:   comment,
	}, nil
}

func buildInvoiceRevisionableRowFromSingle(resp jsonAPISingleResponse) invoiceRevisionableRow {
	resource := resp.Data
	row := invoiceRevisionableRow{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}
	return row
}
