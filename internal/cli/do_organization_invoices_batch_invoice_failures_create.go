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

type doOrganizationInvoicesBatchInvoiceFailuresCreateOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	OrganizationInvoicesBatchInvoice string
	Comment                          string
}

func newDoOrganizationInvoicesBatchInvoiceFailuresCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Fail an organization invoices batch invoice",
		Long: `Fail an organization invoices batch invoice.

Required flags:
  --organization-invoices-batch-invoice   Organization invoices batch invoice ID (required)

Optional flags:
  --comment                               Comment explaining the failure`,
		Example: `  # Fail an organization invoices batch invoice
  xbe do organization-invoices-batch-invoice-failures create --organization-invoices-batch-invoice 12345

  # Fail an organization invoices batch invoice with a comment
  xbe do organization-invoices-batch-invoice-failures create \
    --organization-invoices-batch-invoice 12345 \
    --comment "Marked failed after review"

  # JSON output
  xbe do organization-invoices-batch-invoice-failures create --organization-invoices-batch-invoice 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationInvoicesBatchInvoiceFailuresCreate,
	}
	initDoOrganizationInvoicesBatchInvoiceFailuresCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchInvoiceFailuresCmd.AddCommand(newDoOrganizationInvoicesBatchInvoiceFailuresCreateCmd())
}

func initDoOrganizationInvoicesBatchInvoiceFailuresCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-invoices-batch-invoice", "", "Organization invoices batch invoice ID (required)")
	cmd.Flags().String("comment", "", "Comment explaining the failure")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationInvoicesBatchInvoiceFailuresCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationInvoicesBatchInvoiceFailuresCreateOptions(cmd)
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

	if strings.TrimSpace(opts.OrganizationInvoicesBatchInvoice) == "" {
		err := fmt.Errorf("--organization-invoices-batch-invoice is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"organization-invoices-batch-invoice": map[string]any{
			"data": map[string]any{
				"type": "organization-invoices-batch-invoices",
				"id":   opts.OrganizationInvoicesBatchInvoice,
			},
		},
	}

	data := map[string]any{
		"type":          "organization-invoices-batch-invoice-failures",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/organization-invoices-batch-invoice-failures", jsonBody)
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

	row := buildOrganizationInvoicesBatchInvoiceFailureRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.OrganizationInvoicesBatchInvoice != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch invoice failure %s for batch invoice %s\n", row.ID, row.OrganizationInvoicesBatchInvoice)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch invoice failure %s\n", row.ID)
	return nil
}

func parseDoOrganizationInvoicesBatchInvoiceFailuresCreateOptions(cmd *cobra.Command) (doOrganizationInvoicesBatchInvoiceFailuresCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationInvoicesBatchInvoice, _ := cmd.Flags().GetString("organization-invoices-batch-invoice")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchInvoiceFailuresCreateOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		OrganizationInvoicesBatchInvoice: organizationInvoicesBatchInvoice,
		Comment:                          comment,
	}, nil
}
