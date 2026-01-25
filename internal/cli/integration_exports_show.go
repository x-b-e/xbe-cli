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

type integrationExportsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type integrationExportDetails struct {
	ID                                string   `json:"id"`
	Summary                           string   `json:"summary,omitempty"`
	Description                       string   `json:"description,omitempty"`
	JID                               string   `json:"jid,omitempty"`
	ExportedAt                        string   `json:"exported_at,omitempty"`
	ExportResults                     any      `json:"export_results,omitempty"`
	ExportErrors                      any      `json:"export_errors,omitempty"`
	IntegrationConfigID               string   `json:"integration_config_id,omitempty"`
	IntegrationConfigName             string   `json:"integration_config_name,omitempty"`
	BrokerID                          string   `json:"broker_id,omitempty"`
	BrokerName                        string   `json:"broker_name,omitempty"`
	CreatedByID                       string   `json:"created_by_id,omitempty"`
	CreatedByName                     string   `json:"created_by_name,omitempty"`
	IntegrationInvoicesBatchExportIDs []string `json:"integration_invoices_batch_export_ids,omitempty"`
	CreatedAt                         string   `json:"created_at,omitempty"`
	UpdatedAt                         string   `json:"updated_at,omitempty"`
}

func newIntegrationExportsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show integration export details",
		Long: `Show the full details of a specific integration export.

Output Fields:
  ID           Integration export identifier
  SUMMARY      Summary
  DESCRIPTION  Description
  JID          Job ID
  EXPORTED AT  Export timestamp
  CONFIG       Integration config
  BROKER       Broker
  CREATED BY   Creator
  CREATED AT   Creation timestamp
  UPDATED AT   Update timestamp
  BATCH EXPORTS Associated invoices batch export IDs
  EXPORT RESULTS Export result payload
  EXPORT ERRORS  Export error payload

Arguments:
  <id>  Integration export ID (required). Find IDs using the list command.`,
		Example: `  # View an integration export by ID
  xbe view integration-exports show 123

  # Get JSON output
  xbe view integration-exports show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIntegrationExportsShow,
	}
	initIntegrationExportsShowFlags(cmd)
	return cmd
}

func init() {
	integrationExportsCmd.AddCommand(newIntegrationExportsShowCmd())
}

func initIntegrationExportsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIntegrationExportsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseIntegrationExportsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("integration export id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[integration-exports]", strings.Join([]string{
		"summary",
		"description",
		"jid",
		"exported-at",
		"export-results",
		"export-errors",
		"created-at",
		"updated-at",
		"integration-config",
		"broker",
		"created-by",
		"integration-invoices-batch-exports",
	}, ","))
	query.Set("include", "integration-config,broker,created-by")
	query.Set("fields[integration-configs]", "friendly-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/integration-exports/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildIntegrationExportDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIntegrationExportDetails(cmd, details)
}

func parseIntegrationExportsShowOptions(cmd *cobra.Command) (integrationExportsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return integrationExportsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIntegrationExportDetails(resp jsonAPISingleResponse) integrationExportDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := integrationExportDetails{
		ID:            resp.Data.ID,
		Summary:       stringAttr(attrs, "summary"),
		Description:   stringAttr(attrs, "description"),
		JID:           stringAttr(attrs, "jid"),
		ExportedAt:    formatDateTime(stringAttr(attrs, "exported-at")),
		ExportResults: anyAttr(attrs, "export-results"),
		ExportErrors:  anyAttr(attrs, "export-errors"),
		CreatedAt:     formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:     formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["integration-config"]; ok && rel.Data != nil {
		details.IntegrationConfigID = rel.Data.ID
		if config, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.IntegrationConfigName = strings.TrimSpace(stringAttr(config.Attributes, "friendly-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["integration-invoices-batch-exports"]; ok {
		details.IntegrationInvoicesBatchExportIDs = extractRelationshipIDs(rel)
	}

	return details
}

func renderIntegrationExportDetails(cmd *cobra.Command, details integrationExportDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Summary != "" {
		fmt.Fprintf(out, "Summary: %s\n", details.Summary)
	}
	if details.Description != "" {
		fmt.Fprintf(out, "Description: %s\n", details.Description)
	}
	if details.JID != "" {
		fmt.Fprintf(out, "JID: %s\n", details.JID)
	}
	if details.ExportedAt != "" {
		fmt.Fprintf(out, "Exported At: %s\n", details.ExportedAt)
	}
	if details.IntegrationConfigID != "" || details.IntegrationConfigName != "" {
		fmt.Fprintf(out, "Integration Config: %s\n", formatRelated(details.IntegrationConfigName, details.IntegrationConfigID))
	}
	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}
	if details.CreatedByID != "" || details.CreatedByName != "" {
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(details.CreatedByName, details.CreatedByID))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if len(details.IntegrationInvoicesBatchExportIDs) > 0 {
		fmt.Fprintf(out, "Integration Invoices Batch Exports: %s\n", strings.Join(details.IntegrationInvoicesBatchExportIDs, ", "))
	}

	if formatted := formatAnyJSON(details.ExportResults); formatted != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Export Results:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatted)
	}

	if formatted := formatAnyJSON(details.ExportErrors); formatted != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Export Errors:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, formatted)
	}

	return nil
}
