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

type importerConfigurationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type importerConfigurationDetails struct {
	ID                       string `json:"id"`
	ImporterDataSourceType   string `json:"importer_data_source_type,omitempty"`
	TicketIdentifierField    string `json:"ticket_identifier_field,omitempty"`
	LatestTicketsQueries     any    `json:"latest_tickets_queries,omitempty"`
	AdditionalConfigurations any    `json:"additional_configurations,omitempty"`
	IntegrationConfigID      string `json:"integration_config_id,omitempty"`
}

func newImporterConfigurationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show importer configuration details",
		Long: `Show the full details of an importer configuration.

Output Fields:
  ID                       Importer configuration ID
  IMPORTER DATA SOURCE     Importer data source type
  TICKET IDENTIFIER FIELD  Ticket identifier field name
  LATEST TICKETS QUERIES   Latest tickets query definitions (JSON)
  ADDITIONAL CONFIGURATIONS Additional configuration settings (JSON)
  INTEGRATION CONFIG       Integration config ID

Arguments:
  <id>    The importer configuration ID (required). You can find IDs using the list command.`,
		Example: `  # Show an importer configuration
  xbe view importer-configurations show 123

  # Get JSON output
  xbe view importer-configurations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runImporterConfigurationsShow,
	}
	initImporterConfigurationsShowFlags(cmd)
	return cmd
}

func init() {
	importerConfigurationsCmd.AddCommand(newImporterConfigurationsShowCmd())
}

func initImporterConfigurationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runImporterConfigurationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseImporterConfigurationsShowOptions(cmd)
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
		return fmt.Errorf("importer configuration id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[importer-configurations]", "importer-data-source-type,ticket-identifier-field,latest-tickets-queries,additional-configurations,integration-config")

	body, _, err := client.Get(cmd.Context(), "/v1/free-ticketing/importer-configurations/"+id, query)
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

	details := buildImporterConfigurationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderImporterConfigurationDetails(cmd, details)
}

func parseImporterConfigurationsShowOptions(cmd *cobra.Command) (importerConfigurationsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return importerConfigurationsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return importerConfigurationsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return importerConfigurationsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return importerConfigurationsShowOptions{}, err
	}

	return importerConfigurationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildImporterConfigurationDetails(resp jsonAPISingleResponse) importerConfigurationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := importerConfigurationDetails{
		ID:                       resource.ID,
		ImporterDataSourceType:   stringAttr(attrs, "importer-data-source-type"),
		TicketIdentifierField:    stringAttr(attrs, "ticket-identifier-field"),
		LatestTicketsQueries:     anyAttr(attrs, "latest-tickets-queries"),
		AdditionalConfigurations: anyAttr(attrs, "additional-configurations"),
		IntegrationConfigID:      relationshipIDFromMap(resource.Relationships, "integration-config"),
	}

	if details.LatestTicketsQueries == nil {
		details.LatestTicketsQueries = anyAttr(attrs, "latest_tickets_queries")
	}
	if details.AdditionalConfigurations == nil {
		details.AdditionalConfigurations = anyAttr(attrs, "additional_configurations")
	}

	return details
}

func renderImporterConfigurationDetails(cmd *cobra.Command, details importerConfigurationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ImporterDataSourceType != "" {
		fmt.Fprintf(out, "Importer Data Source Type: %s\n", details.ImporterDataSourceType)
	}
	if details.TicketIdentifierField != "" {
		fmt.Fprintf(out, "Ticket Identifier Field: %s\n", details.TicketIdentifierField)
	}
	if details.IntegrationConfigID != "" {
		fmt.Fprintf(out, "Integration Config ID: %s\n", details.IntegrationConfigID)
	}
	if details.LatestTicketsQueries != nil {
		fmt.Fprintln(out, "Latest Tickets Queries:")
		fmt.Fprintln(out, formatJSONBlock(details.LatestTicketsQueries, "  "))
	}
	if details.AdditionalConfigurations != nil {
		fmt.Fprintln(out, "Additional Configurations:")
		fmt.Fprintln(out, formatJSONBlock(details.AdditionalConfigurations, "  "))
	}

	return nil
}
