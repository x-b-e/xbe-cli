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

type doImporterConfigurationsCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ImporterDataSourceType   string
	TicketIdentifierField    string
	LatestTicketsQueries     string
	AdditionalConfigurations string
	IntegrationConfig        string
}

func newDoImporterConfigurationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an importer configuration",
		Long: `Create an importer configuration for free ticketing imports.

Required flags:
  --importer-data-source-type  Importer data source type (required)
  --ticket-identifier-field    Ticket identifier field name (required)
  --integration-config         Integration config ID (required)

Optional flags:
  --latest-tickets-queries     JSON object or array of query definitions
  --additional-configurations  JSON object of additional settings`,
		Example: `  # Create an importer configuration
  xbe do importer-configurations create \\
    --importer-data-source-type "tms" \\
    --ticket-identifier-field "ticket_id" \\
    --integration-config 123

  # Provide latest tickets queries and extra configuration
  xbe do importer-configurations create \\
    --importer-data-source-type "tms" \\
    --ticket-identifier-field "ticket_id" \\
    --integration-config 123 \\
    --latest-tickets-queries '[{\"name\":\"recent\",\"limit\":50}]' \\
    --additional-configurations '{\"mode\":\"full\"}'

  # Get JSON output
  xbe do importer-configurations create --importer-data-source-type \"tms\" --ticket-identifier-field \"ticket_id\" --integration-config 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoImporterConfigurationsCreate,
	}
	initDoImporterConfigurationsCreateFlags(cmd)
	return cmd
}

func init() {
	doImporterConfigurationsCmd.AddCommand(newDoImporterConfigurationsCreateCmd())
}

func initDoImporterConfigurationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("importer-data-source-type", "", "Importer data source type (required)")
	cmd.Flags().String("ticket-identifier-field", "", "Ticket identifier field name (required)")
	cmd.Flags().String("latest-tickets-queries", "", "JSON object or array of query definitions")
	cmd.Flags().String("additional-configurations", "", "JSON object of additional settings")
	cmd.Flags().String("integration-config", "", "Integration config ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoImporterConfigurationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoImporterConfigurationsCreateOptions(cmd)
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

	if opts.ImporterDataSourceType == "" {
		err := fmt.Errorf("--importer-data-source-type is required")
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
		"importer-data-source-type": opts.ImporterDataSourceType,
		"ticket-identifier-field":   opts.TicketIdentifierField,
	}
	if cmd.Flags().Changed("latest-tickets-queries") {
		if strings.TrimSpace(opts.LatestTicketsQueries) == "" {
			return fmt.Errorf("--latest-tickets-queries requires valid JSON")
		}
		var parsed any
		if err := json.Unmarshal([]byte(opts.LatestTicketsQueries), &parsed); err != nil {
			return fmt.Errorf("invalid latest-tickets-queries JSON: %w", err)
		}
		attributes["latest-tickets-queries"] = parsed
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
			"type":          "importer-configurations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/free-ticketing/importer-configurations", jsonBody)
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

	details := buildImporterConfigurationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created importer configuration %s\n", details.ID)
	return renderImporterConfigurationDetails(cmd, details)
}

func parseDoImporterConfigurationsCreateOptions(cmd *cobra.Command) (doImporterConfigurationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	importerDataSourceType, _ := cmd.Flags().GetString("importer-data-source-type")
	ticketIdentifierField, _ := cmd.Flags().GetString("ticket-identifier-field")
	latestTicketsQueries, _ := cmd.Flags().GetString("latest-tickets-queries")
	additionalConfigurations, _ := cmd.Flags().GetString("additional-configurations")
	integrationConfig, _ := cmd.Flags().GetString("integration-config")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doImporterConfigurationsCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ImporterDataSourceType:   importerDataSourceType,
		TicketIdentifierField:    ticketIdentifierField,
		LatestTicketsQueries:     latestTicketsQueries,
		AdditionalConfigurations: additionalConfigurations,
		IntegrationConfig:        integrationConfig,
	}, nil
}
