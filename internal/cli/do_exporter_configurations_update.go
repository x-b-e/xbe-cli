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

type doExporterConfigurationsUpdateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ID                       string
	Name                     string
	APIURL                   string
	TicketIdentifierField    string
	Template                 string
	ExporterHeaders          string
	AdditionalConfigurations string
}

func newDoExporterConfigurationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an exporter configuration",
		Long: `Update an existing exporter configuration.

Provide the exporter configuration ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name                     Configuration name
  --api-url                  Export destination URL
  --ticket-identifier-field  Ticket identifier field name
  --template                 Template name
  --exporter-headers         JSON object of HTTP headers
  --additional-configurations JSON object of additional settings`,
		Example: `  # Update the name
  xbe do exporter-configurations update 123 --name "Updated Exporter"

  # Update headers
  xbe do exporter-configurations update 123 --exporter-headers '{\"Authorization\":\"Bearer token\"}'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoExporterConfigurationsUpdate,
	}
	initDoExporterConfigurationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doExporterConfigurationsCmd.AddCommand(newDoExporterConfigurationsUpdateCmd())
}

func initDoExporterConfigurationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Configuration name")
	cmd.Flags().String("api-url", "", "Export destination URL")
	cmd.Flags().String("ticket-identifier-field", "", "Ticket identifier field name")
	cmd.Flags().String("template", "", "Template name")
	cmd.Flags().String("exporter-headers", "", "JSON object of HTTP headers")
	cmd.Flags().String("additional-configurations", "", "JSON object of additional settings")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoExporterConfigurationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoExporterConfigurationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("api-url") {
		attributes["api-url"] = opts.APIURL
	}
	if cmd.Flags().Changed("ticket-identifier-field") {
		attributes["ticket-identifier-field"] = opts.TicketIdentifierField
	}
	if cmd.Flags().Changed("template") {
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

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "exporter-configurations",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/free-ticketing/exporter-configurations/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated exporter configuration %s (%s)\n", details.ID, details.Name)
	return nil
}

func parseDoExporterConfigurationsUpdateOptions(cmd *cobra.Command, args []string) (doExporterConfigurationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	apiURL, _ := cmd.Flags().GetString("api-url")
	ticketIdentifierField, _ := cmd.Flags().GetString("ticket-identifier-field")
	template, _ := cmd.Flags().GetString("template")
	exporterHeaders, _ := cmd.Flags().GetString("exporter-headers")
	additionalConfigurations, _ := cmd.Flags().GetString("additional-configurations")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doExporterConfigurationsUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ID:                       args[0],
		Name:                     name,
		APIURL:                   apiURL,
		TicketIdentifierField:    ticketIdentifierField,
		Template:                 template,
		ExporterHeaders:          exporterHeaders,
		AdditionalConfigurations: additionalConfigurations,
	}, nil
}
