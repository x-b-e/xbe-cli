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

type doOrganizationInvoicesBatchInvoiceUnbatchingsCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	OrganizationInvoicesBatchInvoiceID string
	Comment                            string
}

type organizationInvoicesBatchInvoiceUnbatchingRowCreate struct {
	ID                                 string `json:"id"`
	OrganizationInvoicesBatchInvoiceID string `json:"organization_invoices_batch_invoice_id,omitempty"`
	Comment                            string `json:"comment,omitempty"`
}

func newDoOrganizationInvoicesBatchInvoiceUnbatchingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unbatch an organization invoices batch invoice",
		Long: `Unbatch an organization invoices batch invoice.

This action transitions the batch invoice status to skipped. Only batch invoices
in successful or failed status can be unbatched.

Required flags:
  --organization-invoices-batch-invoice   Organization invoices batch invoice ID

Optional flags:
  --comment   Comment for the unbatching

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Unbatch an organization invoices batch invoice
  xbe do organization-invoices-batch-invoice-unbatchings create \\
    --organization-invoices-batch-invoice 123 \\
    --comment "Unbatching for correction"

  # JSON output
  xbe do organization-invoices-batch-invoice-unbatchings create \\
    --organization-invoices-batch-invoice 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationInvoicesBatchInvoiceUnbatchingsCreate,
	}
	initDoOrganizationInvoicesBatchInvoiceUnbatchingsCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchInvoiceUnbatchingsCmd.AddCommand(newDoOrganizationInvoicesBatchInvoiceUnbatchingsCreateCmd())
}

func initDoOrganizationInvoicesBatchInvoiceUnbatchingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-invoices-batch-invoice", "", "Organization invoices batch invoice ID (required)")
	cmd.Flags().String("comment", "", "Comment for the unbatching")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("organization-invoices-batch-invoice")
}

func runDoOrganizationInvoicesBatchInvoiceUnbatchingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationInvoicesBatchInvoiceUnbatchingsCreateOptions(cmd)
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

	attributes := map[string]any{}
	setStringAttrIfPresent(attributes, "comment", opts.Comment)

	relationships := map[string]any{
		"organization-invoices-batch-invoice": map[string]any{
			"data": map[string]any{
				"type": "organization-invoices-batch-invoices",
				"id":   opts.OrganizationInvoicesBatchInvoiceID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "organization-invoices-batch-invoice-unbatchings",
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

	body, _, err := client.Post(cmd.Context(), "/v1/organization-invoices-batch-invoice-unbatchings", jsonBody)
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

	row := buildOrganizationInvoicesBatchInvoiceUnbatchingRow(resp.Data)
	createRow := organizationInvoicesBatchInvoiceUnbatchingRowCreate{
		ID:                                 row.ID,
		OrganizationInvoicesBatchInvoiceID: row.OrganizationInvoicesBatchInvoiceID,
		Comment:                            row.Comment,
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), createRow)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch invoice unbatching %s\n", row.ID)
	return nil
}

func parseDoOrganizationInvoicesBatchInvoiceUnbatchingsCreateOptions(cmd *cobra.Command) (doOrganizationInvoicesBatchInvoiceUnbatchingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	batchInvoiceID, _ := cmd.Flags().GetString("organization-invoices-batch-invoice")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchInvoiceUnbatchingsCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		OrganizationInvoicesBatchInvoiceID: batchInvoiceID,
		Comment:                            comment,
	}, nil
}
