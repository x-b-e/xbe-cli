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

type doOrganizationInvoicesBatchProcessesCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	OrganizationInvoicesBatch string
	Comment                   string
}

func newDoOrganizationInvoicesBatchProcessesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Process an organization invoices batch",
		Long: `Process an organization invoices batch.

Required flags:
  --organization-invoices-batch   Organization invoices batch ID (required)

Optional flags:
  --comment                       Comment explaining the change`,
		Example: `  # Process an organization invoices batch
  xbe do organization-invoices-batch-processes create --organization-invoices-batch 12345

  # Process an organization invoices batch with a comment
  xbe do organization-invoices-batch-processes create \
    --organization-invoices-batch 12345 \
    --comment "Processed after review"

  # JSON output
  xbe do organization-invoices-batch-processes create --organization-invoices-batch 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationInvoicesBatchProcessesCreate,
	}
	initDoOrganizationInvoicesBatchProcessesCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchProcessesCmd.AddCommand(newDoOrganizationInvoicesBatchProcessesCreateCmd())
}

func initDoOrganizationInvoicesBatchProcessesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-invoices-batch", "", "Organization invoices batch ID (required)")
	cmd.Flags().String("comment", "", "Comment explaining the change")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationInvoicesBatchProcessesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationInvoicesBatchProcessesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.OrganizationInvoicesBatch) == "" {
		err := fmt.Errorf("--organization-invoices-batch is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}

	relationships := map[string]any{
		"organization-invoices-batch": map[string]any{
			"data": map[string]any{
				"type": "organization-invoices-batches",
				"id":   opts.OrganizationInvoicesBatch,
			},
		},
	}

	data := map[string]any{
		"type":          "organization-invoices-batch-processes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/organization-invoices-batch-processes", jsonBody)
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

	row := buildOrganizationInvoicesBatchProcessRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.OrganizationInvoicesBatch != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch process %s for batch %s\n", row.ID, row.OrganizationInvoicesBatch)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch process %s\n", row.ID)
	return nil
}

func parseDoOrganizationInvoicesBatchProcessesCreateOptions(cmd *cobra.Command) (doOrganizationInvoicesBatchProcessesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationInvoicesBatch, _ := cmd.Flags().GetString("organization-invoices-batch")
	comment, _ := cmd.Flags().GetString("comment")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchProcessesCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		OrganizationInvoicesBatch: organizationInvoicesBatch,
		Comment:                   comment,
	}, nil
}
