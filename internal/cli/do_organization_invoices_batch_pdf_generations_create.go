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

type doOrganizationInvoicesBatchPdfGenerationsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	OrganizationInvoicesBatch string
	OrganizationPdfTemplate   string
}

func newDoOrganizationInvoicesBatchPdfGenerationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an organization invoices batch PDF generation",
		Long: `Create an organization invoices batch PDF generation.

Required flags:
  --organization-invoices-batch  Organization invoices batch ID
  --organization-pdf-template    Organization invoices batch PDF template ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a PDF generation
  xbe do organization-invoices-batch-pdf-generations create \
    --organization-invoices-batch 123 \
    --organization-pdf-template 456

  # JSON output
  xbe do organization-invoices-batch-pdf-generations create \
    --organization-invoices-batch 123 \
    --organization-pdf-template 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationInvoicesBatchPdfGenerationsCreate,
	}
	initDoOrganizationInvoicesBatchPdfGenerationsCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchPdfGenerationsCmd.AddCommand(newDoOrganizationInvoicesBatchPdfGenerationsCreateCmd())
}

func initDoOrganizationInvoicesBatchPdfGenerationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-invoices-batch", "", "Organization invoices batch ID (required)")
	cmd.Flags().String("organization-pdf-template", "", "Organization invoices batch PDF template ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("organization-invoices-batch")
	cmd.MarkFlagRequired("organization-pdf-template")
}

func runDoOrganizationInvoicesBatchPdfGenerationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationInvoicesBatchPdfGenerationsCreateOptions(cmd)
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

	relationships := map[string]any{
		"organization-invoices-batch": map[string]any{
			"data": map[string]any{
				"type": "organization-invoices-batches",
				"id":   opts.OrganizationInvoicesBatch,
			},
		},
		"organization-pdf-template": map[string]any{
			"data": map[string]any{
				"type": "organization-invoices-batch-pdf-templates",
				"id":   opts.OrganizationPdfTemplate,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "organization-invoices-batch-pdf-generations",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/organization-invoices-batch-pdf-generations", jsonBody)
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

	row := buildOrganizationInvoicesBatchPdfGenerationRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch PDF generation %s\n", row.ID)
	return nil
}

func parseDoOrganizationInvoicesBatchPdfGenerationsCreateOptions(cmd *cobra.Command) (doOrganizationInvoicesBatchPdfGenerationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationInvoicesBatch, _ := cmd.Flags().GetString("organization-invoices-batch")
	organizationPdfTemplate, _ := cmd.Flags().GetString("organization-pdf-template")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchPdfGenerationsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		OrganizationInvoicesBatch: organizationInvoicesBatch,
		OrganizationPdfTemplate:   organizationPdfTemplate,
	}, nil
}
