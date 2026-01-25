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

type exporterConfigurationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type exporterConfigurationDetails struct {
	ID                       string `json:"id"`
	Name                     string `json:"name,omitempty"`
	APIURL                   string `json:"api_url,omitempty"`
	TicketIdentifierField    string `json:"ticket_identifier_field,omitempty"`
	Template                 string `json:"template,omitempty"`
	IntegrationConfigID      string `json:"integration_config_id,omitempty"`
	ExporterHeaders          any    `json:"exporter_headers,omitempty"`
	AdditionalConfigurations any    `json:"additional_configurations,omitempty"`
}

func newExporterConfigurationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show exporter configuration details",
		Long: `Show the full details of an exporter configuration.

Output Fields:
  ID                       Exporter configuration ID
  NAME                     Configuration name
  API URL                  Export destination URL
  TICKET IDENTIFIER FIELD  Ticket identifier field name
  TEMPLATE                 Template name (if configured)
  INTEGRATION CONFIG       Integration config ID
  EXPORTER HEADERS          HTTP headers for exporter requests (JSON)
  ADDITIONAL CONFIGURATIONS Additional configuration settings (JSON)

Arguments:
  <id>    The exporter configuration ID (required). You can find IDs using the list command.`,
		Example: `  # Show an exporter configuration
  xbe view exporter-configurations show 123

  # Get JSON output
  xbe view exporter-configurations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runExporterConfigurationsShow,
	}
	initExporterConfigurationsShowFlags(cmd)
	return cmd
}

func init() {
	exporterConfigurationsCmd.AddCommand(newExporterConfigurationsShowCmd())
}

func initExporterConfigurationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runExporterConfigurationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseExporterConfigurationsShowOptions(cmd)
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
		return fmt.Errorf("exporter configuration id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[exporter-configurations]", "name,api-url,exporter-headers,ticket-identifier-field,template,additional-configurations,integration-config")

	body, _, err := client.Get(cmd.Context(), "/v1/free-ticketing/exporter-configurations/"+id, query)
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

	return renderExporterConfigurationDetails(cmd, details)
}

func parseExporterConfigurationsShowOptions(cmd *cobra.Command) (exporterConfigurationsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return exporterConfigurationsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return exporterConfigurationsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return exporterConfigurationsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return exporterConfigurationsShowOptions{}, err
	}

	return exporterConfigurationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildExporterConfigurationDetails(resp jsonAPISingleResponse) exporterConfigurationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := exporterConfigurationDetails{
		ID:                       resource.ID,
		Name:                     stringAttr(attrs, "name"),
		APIURL:                   stringAttr(attrs, "api-url"),
		TicketIdentifierField:    stringAttr(attrs, "ticket-identifier-field"),
		Template:                 stringAttr(attrs, "template"),
		IntegrationConfigID:      relationshipIDFromMap(resource.Relationships, "integration-config"),
		ExporterHeaders:          anyAttr(attrs, "exporter-headers"),
		AdditionalConfigurations: anyAttr(attrs, "additional-configurations"),
	}

	if details.ExporterHeaders == nil {
		details.ExporterHeaders = anyAttr(attrs, "exporter_headers")
	}
	if details.AdditionalConfigurations == nil {
		details.AdditionalConfigurations = anyAttr(attrs, "additional_configurations")
	}

	return details
}

func renderExporterConfigurationDetails(cmd *cobra.Command, details exporterConfigurationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.APIURL != "" {
		fmt.Fprintf(out, "API URL: %s\n", details.APIURL)
	}
	if details.TicketIdentifierField != "" {
		fmt.Fprintf(out, "Ticket Identifier Field: %s\n", details.TicketIdentifierField)
	}
	if details.Template != "" {
		fmt.Fprintf(out, "Template: %s\n", details.Template)
	}
	if details.IntegrationConfigID != "" {
		fmt.Fprintf(out, "Integration Config ID: %s\n", details.IntegrationConfigID)
	}
	if details.ExporterHeaders != nil {
		fmt.Fprintln(out, "Exporter Headers:")
		fmt.Fprintln(out, formatJSONBlock(details.ExporterHeaders, "  "))
	}
	if details.AdditionalConfigurations != nil {
		fmt.Fprintln(out, "Additional Configurations:")
		fmt.Fprintln(out, formatJSONBlock(details.AdditionalConfigurations, "  "))
	}

	return nil
}
