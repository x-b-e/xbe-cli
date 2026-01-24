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

type doOzingaTkBatchFileExportsCreateOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	OrganizationInvoicesBatchFileID string
	DryRun                          bool
}

type ozingaTkBatchFileExportDetails struct {
	ID                            string `json:"id"`
	OrganizationInvoicesBatchFile string `json:"organization_invoices_batch_file_id,omitempty"`
	DryRun                        bool   `json:"dry_run"`
	ExportResults                 any    `json:"export_results,omitempty"`
	ExportErrors                  any    `json:"export_errors,omitempty"`
}

func newDoOzingaTkBatchFileExportsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an Ozinga TK batch file export",
		Long: `Create an Ozinga TK batch file export.

Required flags:
  --organization-invoices-batch-file   Organization invoices batch file ID

Optional flags:
  --dry-run                             Validate and preview export without sending

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an export
  xbe do ozinga-tk-batch-file-exports create --organization-invoices-batch-file 123

  # Dry-run export
  xbe do ozinga-tk-batch-file-exports create --organization-invoices-batch-file 123 --dry-run

  # JSON output
  xbe do ozinga-tk-batch-file-exports create --organization-invoices-batch-file 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoOzingaTkBatchFileExportsCreate,
	}
	initDoOzingaTkBatchFileExportsCreateFlags(cmd)
	return cmd
}

func init() {
	doOzingaTkBatchFileExportsCmd.AddCommand(newDoOzingaTkBatchFileExportsCreateCmd())
}

func initDoOzingaTkBatchFileExportsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("organization-invoices-batch-file", "", "Organization invoices batch file ID (required)")
	cmd.Flags().Bool("dry-run", false, "Validate and preview export without sending")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("organization-invoices-batch-file")
}

func runDoOzingaTkBatchFileExportsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOzingaTkBatchFileExportsCreateOptions(cmd)
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

	organizationInvoicesBatchFileID := strings.TrimSpace(opts.OrganizationInvoicesBatchFileID)
	if organizationInvoicesBatchFileID == "" {
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
				"id":   organizationInvoicesBatchFileID,
			},
		},
	}

	payload := map[string]any{
		"data": map[string]any{
			"type":          "ozinga-tk-batch-file-exports",
			"relationships": relationships,
		},
	}
	if len(attributes) > 0 {
		payload["data"].(map[string]any)["attributes"] = attributes
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/ozinga-tk-batch-file-exports", jsonBody)
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

	details := buildOzingaTkBatchFileExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderOzingaTkBatchFileExportDetails(cmd, details)
}

func parseDoOzingaTkBatchFileExportsCreateOptions(cmd *cobra.Command) (doOzingaTkBatchFileExportsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	organizationInvoicesBatchFileID, _ := cmd.Flags().GetString("organization-invoices-batch-file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOzingaTkBatchFileExportsCreateOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		OrganizationInvoicesBatchFileID: organizationInvoicesBatchFileID,
		DryRun:                          dryRun,
	}, nil
}

func buildOzingaTkBatchFileExportDetails(resp jsonAPISingleResponse) ozingaTkBatchFileExportDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return ozingaTkBatchFileExportDetails{
		ID:                            resource.ID,
		OrganizationInvoicesBatchFile: relationshipIDFromMap(resource.Relationships, "organization-invoices-batch-file"),
		DryRun:                        boolAttr(attrs, "dry-run"),
		ExportResults:                 anyAttr(attrs, "export-results"),
		ExportErrors:                  anyAttr(attrs, "export-errors"),
	}
}

func renderOzingaTkBatchFileExportDetails(cmd *cobra.Command, details ozingaTkBatchFileExportDetails) error {
	out := cmd.OutOrStdout()

	if details.ID != "" {
		fmt.Fprintf(out, "ID: %s\n", details.ID)
	}
	if details.OrganizationInvoicesBatchFile != "" {
		fmt.Fprintf(out, "Organization Invoices Batch File: %s\n", details.OrganizationInvoicesBatchFile)
	}
	fmt.Fprintf(out, "Dry Run: %t\n", details.DryRun)

	if details.ExportResults != nil {
		fmt.Fprintln(out, "Export Results:")
		if err := writeJSON(out, details.ExportResults); err != nil {
			return err
		}
	}
	if details.ExportErrors != nil {
		fmt.Fprintln(out, "Export Errors:")
		if err := writeJSON(out, details.ExportErrors); err != nil {
			return err
		}
	}

	if details.ID != "" {
		fmt.Fprintf(out, "Created Ozinga TK batch file export %s\n", details.ID)
		return nil
	}

	fmt.Fprintln(out, "Created Ozinga TK batch file export")
	return nil
}
