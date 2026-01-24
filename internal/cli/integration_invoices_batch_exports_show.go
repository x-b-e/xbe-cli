package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type integrationInvoicesBatchExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type integrationInvoicesBatchExportDetails struct {
	ID                              string `json:"id"`
	Details                         any    `json:"details,omitempty"`
	OrganizationInvoicesBatchID     string `json:"organization_invoices_batch_id,omitempty"`
	OrganizationInvoicesBatchFileID string `json:"organization_invoices_batch_file_id,omitempty"`
	IntegrationExportID             string `json:"integration_export_id,omitempty"`
	IntegrationConfigID             string `json:"integration_config_id,omitempty"`
	BrokerID                        string `json:"broker_id,omitempty"`
}

func newIntegrationInvoicesBatchExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show integration invoices batch export details",
		Long: `Show the full details of an integration invoices batch export.

Output Fields:
  ID
  Details
  Organization Invoices Batch ID
  Organization Invoices Batch File ID
  Integration Export ID
  Integration Config ID
  Broker ID

Arguments:
  <id>    The integration invoices batch export ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an integration invoices batch export
  xbe view integration-invoices-batch-exports show 123

  # JSON output
  xbe view integration-invoices-batch-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIntegrationInvoicesBatchExportsShow,
	}
	initIntegrationInvoicesBatchExportsShowFlags(cmd)
	return cmd
}

func init() {
	integrationInvoicesBatchExportsCmd.AddCommand(newIntegrationInvoicesBatchExportsShowCmd())
}

func initIntegrationInvoicesBatchExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIntegrationInvoicesBatchExportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseIntegrationInvoicesBatchExportsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("integration invoices batch export id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[integration-invoices-batch-exports]", "details,organization-invoices-batch,organization-invoices-batch-file,integration-export,integration-config,broker")

	body, _, err := client.Get(cmd.Context(), "/v1/integration-invoices-batch-exports/"+id, query)
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

	details := buildIntegrationInvoicesBatchExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIntegrationInvoicesBatchExportDetails(cmd, details)
}

func parseIntegrationInvoicesBatchExportsShowOptions(cmd *cobra.Command) (integrationInvoicesBatchExportsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return integrationInvoicesBatchExportsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return integrationInvoicesBatchExportsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return integrationInvoicesBatchExportsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return integrationInvoicesBatchExportsShowOptions{}, err
	}

	return integrationInvoicesBatchExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIntegrationInvoicesBatchExportDetails(resp jsonAPISingleResponse) integrationInvoicesBatchExportDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return integrationInvoicesBatchExportDetails{
		ID:                              resource.ID,
		Details:                         attrs["details"],
		OrganizationInvoicesBatchID:     relationshipIDFromMap(resource.Relationships, "organization-invoices-batch"),
		OrganizationInvoicesBatchFileID: relationshipIDFromMap(resource.Relationships, "organization-invoices-batch-file"),
		IntegrationExportID:             relationshipIDFromMap(resource.Relationships, "integration-export"),
		IntegrationConfigID:             relationshipIDFromMap(resource.Relationships, "integration-config"),
		BrokerID:                        relationshipIDFromMap(resource.Relationships, "broker"),
	}
}

func renderIntegrationInvoicesBatchExportDetails(cmd *cobra.Command, details integrationInvoicesBatchExportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.OrganizationInvoicesBatchID != "" {
		fmt.Fprintf(out, "Organization Invoices Batch: %s\n", details.OrganizationInvoicesBatchID)
	}
	if details.OrganizationInvoicesBatchFileID != "" {
		fmt.Fprintf(out, "Organization Invoices Batch File: %s\n", details.OrganizationInvoicesBatchFileID)
	}
	if details.IntegrationExportID != "" {
		fmt.Fprintf(out, "Integration Export: %s\n", details.IntegrationExportID)
	}
	if details.IntegrationConfigID != "" {
		fmt.Fprintf(out, "Integration Config: %s\n", details.IntegrationConfigID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}

	if details.Details != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Details:")
		fmt.Fprintln(out, formatJSONBlock(details.Details, "  "))
	}

	return nil
}
