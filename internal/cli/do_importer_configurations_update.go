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

type doImporterConfigurationsUpdateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ID                       string
	ImporterDataSourceType   string
	TicketIdentifierField    string
	LatestTicketsQueries     string
	AdditionalConfigurations string
}

func newDoImporterConfigurationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an importer configuration",
		Long: `Update an existing importer configuration.

Provide the importer configuration ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --importer-data-source-type  Importer data source type
  --ticket-identifier-field    Ticket identifier field name
  --latest-tickets-queries     JSON object or array of query definitions
  --additional-configurations  JSON object of additional settings`,
		Example: `  # Update importer data source type
  xbe do importer-configurations update 123 --importer-data-source-type "tms"

  # Update latest tickets queries
  xbe do importer-configurations update 123 --latest-tickets-queries '[{\"name\":\"recent\",\"limit\":50}]'`,
		Args: cobra.ExactArgs(1),
		RunE: runDoImporterConfigurationsUpdate,
	}
	initDoImporterConfigurationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doImporterConfigurationsCmd.AddCommand(newDoImporterConfigurationsUpdateCmd())
}

func initDoImporterConfigurationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("importer-data-source-type", "", "Importer data source type")
	cmd.Flags().String("ticket-identifier-field", "", "Ticket identifier field name")
	cmd.Flags().String("latest-tickets-queries", "", "JSON object or array of query definitions")
	cmd.Flags().String("additional-configurations", "", "JSON object of additional settings")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoImporterConfigurationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoImporterConfigurationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("importer-data-source-type") {
		attributes["importer-data-source-type"] = opts.ImporterDataSourceType
	}
	if cmd.Flags().Changed("ticket-identifier-field") {
		attributes["ticket-identifier-field"] = opts.TicketIdentifierField
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

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "importer-configurations",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/free-ticketing/importer-configurations/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated importer configuration %s\n", details.ID)
	return nil
}

func parseDoImporterConfigurationsUpdateOptions(cmd *cobra.Command, args []string) (doImporterConfigurationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	importerDataSourceType, _ := cmd.Flags().GetString("importer-data-source-type")
	ticketIdentifierField, _ := cmd.Flags().GetString("ticket-identifier-field")
	latestTicketsQueries, _ := cmd.Flags().GetString("latest-tickets-queries")
	additionalConfigurations, _ := cmd.Flags().GetString("additional-configurations")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doImporterConfigurationsUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ID:                       args[0],
		ImporterDataSourceType:   importerDataSourceType,
		TicketIdentifierField:    ticketIdentifierField,
		LatestTicketsQueries:     latestTicketsQueries,
		AdditionalConfigurations: additionalConfigurations,
	}, nil
}
