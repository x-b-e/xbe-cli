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

type doOrganizationInvoicesBatchesCreateOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	Organization string
	Invoices     string
}

func newDoOrganizationInvoicesBatchesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an organization invoices batch",
		Long: `Create an organization invoices batch.

Required flags:
  --organization  Organization in Type|ID format (e.g. Broker|123)
  --invoices      Invoice IDs (comma-separated)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an organization invoices batch
  xbe do organization-invoices-batches create \
    --organization "Broker|123" \
    --invoices 111,222

  # JSON output
  xbe do organization-invoices-batches create \
    --organization "Broker|123" \
    --invoices 111,222 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationInvoicesBatchesCreate,
	}
	initDoOrganizationInvoicesBatchesCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchesCmd.AddCommand(newDoOrganizationInvoicesBatchesCreateCmd())
}

func initDoOrganizationInvoicesBatchesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (required)")
	cmd.Flags().String("invoices", "", "Invoice IDs (comma-separated, required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("organization")
	cmd.MarkFlagRequired("invoices")
}

func runDoOrganizationInvoicesBatchesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationInvoicesBatchesCreateOptions(cmd)
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

	orgType, orgID, err := parseOrganization(opts.Organization)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	invoiceIDs := splitCommaList(opts.Invoices)
	if len(invoiceIDs) == 0 {
		err := fmt.Errorf("--invoices must include at least one invoice ID")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	invoiceData := make([]map[string]any, 0, len(invoiceIDs))
	for _, id := range invoiceIDs {
		invoiceData = append(invoiceData, map[string]any{
			"type": "invoices",
			"id":   id,
		})
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": orgType,
				"id":   orgID,
			},
		},
		"invoices": map[string]any{
			"data": invoiceData,
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "organization-invoices-batches",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/organization-invoices-batches", jsonBody)
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

	row := buildOrganizationInvoicesBatchRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch %s\n", row.ID)
	return nil
}

func parseDoOrganizationInvoicesBatchesCreateOptions(cmd *cobra.Command) (doOrganizationInvoicesBatchesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organization, _ := cmd.Flags().GetString("organization")
	invoices, _ := cmd.Flags().GetString("invoices")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchesCreateOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		Organization: organization,
		Invoices:     invoices,
	}, nil
}
