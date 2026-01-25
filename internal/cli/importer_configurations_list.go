package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type importerConfigurationsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	IntegrationConfig string
	Broker            string
}

type importerConfigurationRow struct {
	ID                     string `json:"id"`
	ImporterDataSourceType string `json:"importer_data_source_type,omitempty"`
	TicketIdentifierField  string `json:"ticket_identifier_field,omitempty"`
	IntegrationConfigID    string `json:"integration_config_id,omitempty"`
}

func newImporterConfigurationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List importer configurations",
		Long: `List importer configurations with filtering and pagination.

Importer configurations define inbound integrations for free ticketing data.

Output Columns:
  ID                   Importer configuration ID
  DATA SOURCE TYPE     Importer data source type
  TICKET FIELD         Ticket identifier field name
  INTEGRATION CONFIG   Integration config ID

Filters:
  --integration-config   Filter by integration config ID (comma-separated)
  --broker               Filter by broker ID (comma-separated)`,
		Example: `  # List importer configurations
  xbe view importer-configurations list

  # Filter by integration config
  xbe view importer-configurations list --integration-config 123

  # Filter by broker
  xbe view importer-configurations list --broker 456

  # Output as JSON
  xbe view importer-configurations list --json`,
		RunE: runImporterConfigurationsList,
	}
	initImporterConfigurationsListFlags(cmd)
	return cmd
}

func init() {
	importerConfigurationsCmd.AddCommand(newImporterConfigurationsListCmd())
}

func initImporterConfigurationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("integration-config", "", "Filter by integration config ID (comma-separated)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runImporterConfigurationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseImporterConfigurationsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[importer-configurations]", "importer-data-source-type,ticket-identifier-field,integration-config")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[integration-config]", opts.IntegrationConfig)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/free-ticketing/importer-configurations", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildImporterConfigurationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderImporterConfigurationsTable(cmd, rows)
}

func parseImporterConfigurationsListOptions(cmd *cobra.Command) (importerConfigurationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	integrationConfig, _ := cmd.Flags().GetString("integration-config")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return importerConfigurationsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		IntegrationConfig: integrationConfig,
		Broker:            broker,
	}, nil
}

func buildImporterConfigurationRows(resp jsonAPIResponse) []importerConfigurationRow {
	rows := make([]importerConfigurationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, importerConfigurationRow{
			ID:                     resource.ID,
			ImporterDataSourceType: stringAttr(resource.Attributes, "importer-data-source-type"),
			TicketIdentifierField:  stringAttr(resource.Attributes, "ticket-identifier-field"),
			IntegrationConfigID:    relationshipIDFromMap(resource.Relationships, "integration-config"),
		})
	}
	return rows
}

func renderImporterConfigurationsTable(cmd *cobra.Command, rows []importerConfigurationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No importer configurations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDATA SOURCE TYPE\tTICKET FIELD\tINTEGRATION CONFIG")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ImporterDataSourceType, 24),
			truncateString(row.TicketIdentifierField, 24),
			row.IntegrationConfigID,
		)
	}
	return writer.Flush()
}
