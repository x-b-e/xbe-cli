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

type exporterConfigurationsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	IntegrationConfig string
	Broker            string
}

type exporterConfigurationRow struct {
	ID                  string `json:"id"`
	Name                string `json:"name,omitempty"`
	APIURL              string `json:"api_url,omitempty"`
	TicketIdentifier    string `json:"ticket_identifier_field,omitempty"`
	Template            string `json:"template,omitempty"`
	IntegrationConfigID string `json:"integration_config_id,omitempty"`
}

func newExporterConfigurationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List exporter configurations",
		Long: `List exporter configurations with filtering and pagination.

Exporter configurations define outbound integrations for free ticketing data.

Output Columns:
  ID                   Exporter configuration ID
  NAME                 Configuration name
  API URL              Export destination URL
  TICKET FIELD         Ticket identifier field name
  TEMPLATE             Template name (if configured)
  INTEGRATION CONFIG   Integration config ID

Filters:
  --integration-config   Filter by integration config ID (comma-separated)
  --broker               Filter by broker ID (comma-separated)`,
		Example: `  # List exporter configurations
  xbe view exporter-configurations list

  # Filter by integration config
  xbe view exporter-configurations list --integration-config 123

  # Filter by broker
  xbe view exporter-configurations list --broker 456

  # Output as JSON
  xbe view exporter-configurations list --json`,
		RunE: runExporterConfigurationsList,
	}
	initExporterConfigurationsListFlags(cmd)
	return cmd
}

func init() {
	exporterConfigurationsCmd.AddCommand(newExporterConfigurationsListCmd())
}

func initExporterConfigurationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("integration-config", "", "Filter by integration config ID (comma-separated)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runExporterConfigurationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseExporterConfigurationsListOptions(cmd)
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
	query.Set("fields[exporter-configurations]", "name,api-url,ticket-identifier-field,template,integration-config")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[integration-config]", opts.IntegrationConfig)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/free-ticketing/exporter-configurations", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildExporterConfigurationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderExporterConfigurationsTable(cmd, rows)
}

func parseExporterConfigurationsListOptions(cmd *cobra.Command) (exporterConfigurationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	integrationConfig, _ := cmd.Flags().GetString("integration-config")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return exporterConfigurationsListOptions{
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

func buildExporterConfigurationRows(resp jsonAPIResponse) []exporterConfigurationRow {
	rows := make([]exporterConfigurationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, exporterConfigurationRow{
			ID:                  resource.ID,
			Name:                stringAttr(resource.Attributes, "name"),
			APIURL:              stringAttr(resource.Attributes, "api-url"),
			TicketIdentifier:    stringAttr(resource.Attributes, "ticket-identifier-field"),
			Template:            stringAttr(resource.Attributes, "template"),
			IntegrationConfigID: relationshipIDFromMap(resource.Relationships, "integration-config"),
		})
	}
	return rows
}

func renderExporterConfigurationsTable(cmd *cobra.Command, rows []exporterConfigurationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No exporter configurations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tAPI URL\tTICKET FIELD\tTEMPLATE\tINTEGRATION CONFIG")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(row.APIURL, 40),
			truncateString(row.TicketIdentifier, 24),
			truncateString(row.Template, 20),
			row.IntegrationConfigID,
		)
	}
	return writer.Flush()
}
