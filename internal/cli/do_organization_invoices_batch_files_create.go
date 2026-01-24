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

type doOrganizationInvoicesBatchFilesCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	OrganizationInvoicesBatch string
	OrganizationFormatter     string
	Organization              string
	Body                      string
	MimeType                  string
	RefreshInvoiceRevisions   string
}

func newDoOrganizationInvoicesBatchFilesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an organization invoices batch file",
		Long: `Create an organization invoices batch file.

Required flags:
  --organization-invoices-batch  Organization invoices batch ID
  --organization-formatter       Organization formatter ID

Optional flags:
  --organization               Organization in Type|ID format (e.g. Broker|123)
  --body                       File body content (string)
  --mime-type                  File MIME type (e.g. text/plain)
  --refresh-invoice-revisions  Refresh invoice revisions before formatting (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an organization invoices batch file
  xbe do organization-invoices-batch-files create \
    --organization-invoices-batch 123 \
    --organization-formatter 456

  # Create with body and mime type
  xbe do organization-invoices-batch-files create \
    --organization-invoices-batch 123 \
    --organization-formatter 456 \
    --body "Invoice export" \
    --mime-type text/plain

  # Create with explicit organization
  xbe do organization-invoices-batch-files create \
    --organization-invoices-batch 123 \
    --organization-formatter 456 \
    --organization "Broker|789"

  # JSON output
  xbe do organization-invoices-batch-files create \
    --organization-invoices-batch 123 \
    --organization-formatter 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationInvoicesBatchFilesCreate,
	}
	initDoOrganizationInvoicesBatchFilesCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchFilesCmd.AddCommand(newDoOrganizationInvoicesBatchFilesCreateCmd())
}

func initDoOrganizationInvoicesBatchFilesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-invoices-batch", "", "Organization invoices batch ID (required)")
	cmd.Flags().String("organization-formatter", "", "Organization formatter ID (required)")
	cmd.Flags().String("organization", "", "Organization in Type|ID format (optional)")
	cmd.Flags().String("body", "", "File body content")
	cmd.Flags().String("mime-type", "", "File MIME type")
	cmd.Flags().String("refresh-invoice-revisions", "", "Refresh invoice revisions (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("organization-invoices-batch")
	cmd.MarkFlagRequired("organization-formatter")
}

func runDoOrganizationInvoicesBatchFilesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationInvoicesBatchFilesCreateOptions(cmd)
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
	setStringAttrIfPresent(attributes, "body", opts.Body)
	setStringAttrIfPresent(attributes, "mime-type", opts.MimeType)
	setBoolAttrIfPresent(attributes, "refresh-invoice-revisions", opts.RefreshInvoiceRevisions)

	relationships := map[string]any{
		"organization-invoices-batch": map[string]any{
			"data": map[string]any{
				"type": "organization-invoices-batches",
				"id":   opts.OrganizationInvoicesBatch,
			},
		},
		"organization-formatter": map[string]any{
			"data": map[string]any{
				"type": "organization-formatters",
				"id":   opts.OrganizationFormatter,
			},
		},
	}

	if opts.Organization != "" {
		orgType, orgID, err := parseOrganization(opts.Organization)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["organization"] = map[string]any{
			"data": map[string]any{
				"type": orgType,
				"id":   orgID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "organization-invoices-batch-files",
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

	body, _, err := client.Post(cmd.Context(), "/v1/organization-invoices-batch-files", jsonBody)
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

	row := buildOrganizationInvoicesBatchFileRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch file %s\n", row.ID)
	return nil
}

func parseDoOrganizationInvoicesBatchFilesCreateOptions(cmd *cobra.Command) (doOrganizationInvoicesBatchFilesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationInvoicesBatch, _ := cmd.Flags().GetString("organization-invoices-batch")
	organizationFormatter, _ := cmd.Flags().GetString("organization-formatter")
	organization, _ := cmd.Flags().GetString("organization")
	body, _ := cmd.Flags().GetString("body")
	mimeType, _ := cmd.Flags().GetString("mime-type")
	refreshInvoiceRevisions, _ := cmd.Flags().GetString("refresh-invoice-revisions")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchFilesCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		OrganizationInvoicesBatch: organizationInvoicesBatch,
		OrganizationFormatter:     organizationFormatter,
		Organization:              organization,
		Body:                      body,
		MimeType:                  mimeType,
		RefreshInvoiceRevisions:   refreshInvoiceRevisions,
	}, nil
}
