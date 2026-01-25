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

type doOrganizationInvoicesBatchFileExportsCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	OrganizationInvoicesBatchFile string
	DryRun                        bool
}

type organizationInvoicesBatchFileExportRow struct {
	ID                              string `json:"id"`
	OrganizationInvoicesBatchFileID string `json:"organization_invoices_batch_file_id,omitempty"`
	DryRun                          bool   `json:"dry_run"`
	ExportResults                   any    `json:"export_results,omitempty"`
	ExportErrors                    any    `json:"export_errors,omitempty"`
}

func newDoOrganizationInvoicesBatchFileExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Export an organization invoices batch file",
		Long: `Export an organization invoices batch file.

Required flags:
  --organization-invoices-batch-file  Organization invoices batch file ID (required)

Optional flags:
  --dry-run                           Validate export without sending

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Export a batch file
  xbe do organization-invoices-batch-file-exports create --organization-invoices-batch-file 123

  # Run export as a dry run
  xbe do organization-invoices-batch-file-exports create --organization-invoices-batch-file 123 --dry-run

  # Output as JSON
  xbe do organization-invoices-batch-file-exports create --organization-invoices-batch-file 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOrganizationInvoicesBatchFileExportsCreate,
	}
	initDoOrganizationInvoicesBatchFileExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doOrganizationInvoicesBatchFileExportsCmd.AddCommand(newDoOrganizationInvoicesBatchFileExportsCreateCmd())
}

func initDoOrganizationInvoicesBatchFileExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-invoices-batch-file", "", "Organization invoices batch file ID (required)")
	cmd.Flags().Bool("dry-run", false, "Validate export without sending")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOrganizationInvoicesBatchFileExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOrganizationInvoicesBatchFileExportsCreateOptions(cmd)
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

	opts.OrganizationInvoicesBatchFile = strings.TrimSpace(opts.OrganizationInvoicesBatchFile)
	if opts.OrganizationInvoicesBatchFile == "" {
		err := fmt.Errorf("--organization-invoices-batch-file is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("dry-run") {
		attributes["dry-run"] = opts.DryRun
	}

	relationships := map[string]any{
		"organization-invoices-batch-file": map[string]any{
			"data": map[string]any{
				"type": "organization-invoices-batch-files",
				"id":   opts.OrganizationInvoicesBatchFile,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "organization-invoices-batch-file-exports",
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

	body, _, err := client.Post(cmd.Context(), "/v1/organization-invoices-batch-file-exports", jsonBody)
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

	row := organizationInvoicesBatchFileExportRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created organization invoices batch file export %s\n", row.ID)
	return nil
}

func organizationInvoicesBatchFileExportRowFromSingle(resp jsonAPISingleResponse) organizationInvoicesBatchFileExportRow {
	attrs := resp.Data.Attributes
	row := organizationInvoicesBatchFileExportRow{
		ID:            resp.Data.ID,
		DryRun:        boolAttr(attrs, "dry-run"),
		ExportResults: anyAttr(attrs, "export-results"),
		ExportErrors:  anyAttr(attrs, "export-errors"),
	}

	if rel, ok := resp.Data.Relationships["organization-invoices-batch-file"]; ok && rel.Data != nil {
		row.OrganizationInvoicesBatchFileID = rel.Data.ID
	}

	return row
}

func parseDoOrganizationInvoicesBatchFileExportsCreateOptions(cmd *cobra.Command) (doOrganizationInvoicesBatchFileExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationInvoicesBatchFile, _ := cmd.Flags().GetString("organization-invoices-batch-file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOrganizationInvoicesBatchFileExportsCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		OrganizationInvoicesBatchFile: organizationInvoicesBatchFile,
		DryRun:                        dryRun,
	}, nil
}
