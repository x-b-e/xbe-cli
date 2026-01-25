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

type doExporterConfigurationsCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	Name                     string
	APIURL                   string
	TicketIdentifierField    string
	Template                 string
	ExporterHeaders          string
	AdditionalConfigurations string
	IntegrationConfig        string
}

func newDoExporterConfigurationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an exporter configuration",
		Long: `Create an exporter configuration for free ticketing exports.

Required flags:
  --name                     Configuration name (required)
  --api-url                  Export destination URL (required)
  --ticket-identifier-field  Ticket identifier field name (required)
  --integration-config       Integration config ID (required)

Optional flags:
  --template                 Template name
  --exporter-headers         JSON object of HTTP headers
  --additional-configurations JSON object of additional settings`,
		Example: `  # Create an exporter configuration
  xbe do exporter-configurations create \\
    --name "Default Exporter" \\
    --api-url "https://example.com/api/export" \\
    --ticket-identifier-field "ticket_id" \\
    --integration-config 123

  # Provide headers and extra configuration
  xbe do exporter-configurations create \\
    --name "Header Exporter" \\
    --api-url "https://example.com/api/export" \\
    --ticket-identifier-field "ticket_id" \\
    --integration-config 123 \\
    --exporter-headers '{\"Authorization\":\"Bearer token\"}' \\
    --additional-configurations '{\"mode\":\"full\"}'

  # Get JSON output
  xbe do exporter-configurations create --name \"Default Exporter\" --api-url \"https://example.com/api/export\" --ticket-identifier-field \"ticket_id\" --integration-config 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoExporterConfigurationsCreate,
	}
	initDoExporterConfigurationsCreateFlags(cmd)
	return cmd
}

func init() {
	doExporterConfigurationsCmd.AddCommand(newDoExporterConfigurationsCreateCmd())
}

func initDoExporterConfigurationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Configuration name (required)")
	cmd.Flags().String("api-url", "", "Export destination URL (required)")
	cmd.Flags().String("ticket-identifier-field", "", "Ticket identifier field name (required)")
	cmd.Flags().String("template", "", "Template name")
	cmd.Flags().String("exporter-headers", "", "JSON object of HTTP headers")
	cmd.Flags().String("additional-configurations", "", "JSON object of additional settings")
	cmd.Flags().String("integration-config", "", "Integration config ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExporterConfigurationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoExporterConfigurationsCreateOptions(cmd)
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

	if opts.Name == "" {
		err := fmt.Errorf("--name is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.APIURL == "" {
		err := fmt.Errorf("--api-url is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.TicketIdentifierField == "" {
		err := fmt.Errorf("--ticket-identifier-field is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.IntegrationConfig == "" {
		err := fmt.Errorf("--integration-config is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"name":                    opts.Name,
		"api-url":                 opts.APIURL,
		"ticket-identifier-field": opts.TicketIdentifierField,
	}
	if opts.Template != "" {
		attributes["template"] = opts.Template
	}
	if cmd.Flags().Changed("exporter-headers") {
		if strings.TrimSpace(opts.ExporterHeaders) == "" {
			return fmt.Errorf("--exporter-headers requires valid JSON")
		}
		var parsed any
		if err := json.Unmarshal([]byte(opts.ExporterHeaders), &parsed); err != nil {
			return fmt.Errorf("invalid exporter-headers JSON: %w", err)
		}
		attributes["exporter-headers"] = parsed
	}
	if cmd.Flags().Changed("additional-configurations") {
		if strings.TrimSpace(opts.AdditionalConfigurations) == "" {
			return fmt.Errorf("--additional-configurations requires valid JSON")
		}
		var parsed any
		if err := json.Unmarshal([]byte(opts.AdditionalConfigurations), &parsed); err != nil {
			return fmt.Errorf("invalid additional-configurations JSON: %w", err)
		}
		attributes["additional-configurations"] = parsed
	}

	relationships := map[string]any{
		"integration-config": map[string]any{
			"data": map[string]any{
				"type": "integration-configs",
				"id":   opts.IntegrationConfig,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "exporter-configurations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/free-ticketing/exporter-configurations", jsonBody)
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

	details := buildExporterConfigurationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created exporter configuration %s (%s)\n", details.ID, details.Name)
	return renderExporterConfigurationDetails(cmd, details)
}

func parseDoExporterConfigurationsCreateOptions(cmd *cobra.Command) (doExporterConfigurationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	apiURL, _ := cmd.Flags().GetString("api-url")
	ticketIdentifierField, _ := cmd.Flags().GetString("ticket-identifier-field")
	template, _ := cmd.Flags().GetString("template")
	exporterHeaders, _ := cmd.Flags().GetString("exporter-headers")
	additionalConfigurations, _ := cmd.Flags().GetString("additional-configurations")
	integrationConfig, _ := cmd.Flags().GetString("integration-config")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExporterConfigurationsCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		Name:                     name,
		APIURL:                   apiURL,
		TicketIdentifierField:    ticketIdentifierField,
		Template:                 template,
		ExporterHeaders:          exporterHeaders,
		AdditionalConfigurations: additionalConfigurations,
		IntegrationConfig:        integrationConfig,
	}, nil
}
