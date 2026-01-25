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

type deereIntegrationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Broker  string
}

type deereIntegrationRow struct {
	ID                  string `json:"id"`
	IntegrationID       string `json:"integration_identifier,omitempty"`
	FriendlyName        string `json:"friendly_name,omitempty"`
	BrokerID            string `json:"broker_id,omitempty"`
	IntegrationConfigID string `json:"integration_config_id,omitempty"`
}

func newDeereIntegrationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Deere integrations",
		Long: `List Deere integrations with filtering and pagination.

Deere integrations connect John Deere accounts to XBE and track OAuth status.

Output Columns:
  ID                   Deere integration identifier
  INTEGRATION ID       Integration identifier from Deere
  FRIENDLY NAME        Friendly name for the integration
  BROKER ID            Broker ID
  INTEGRATION CONFIG   Integration config ID

Filters:
  --broker   Filter by broker ID (comma-separated)`,
		Example: `  # List Deere integrations
  xbe view deere-integrations list

  # Filter by broker
  xbe view deere-integrations list --broker 123

  # Output as JSON
  xbe view deere-integrations list --json`,
		RunE: runDeereIntegrationsList,
	}
	initDeereIntegrationsListFlags(cmd)
	return cmd
}

func init() {
	deereIntegrationsCmd.AddCommand(newDeereIntegrationsListCmd())
}

func initDeereIntegrationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDeereIntegrationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDeereIntegrationsListOptions(cmd)
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
	query.Set("fields[deere-integrations]", "integration-identifier,friendly-name,broker,integration-config")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/deere-integrations", query)
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

	rows := buildDeereIntegrationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDeereIntegrationsTable(cmd, rows)
}

func parseDeereIntegrationsListOptions(cmd *cobra.Command) (deereIntegrationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return deereIntegrationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Broker:  broker,
	}, nil
}

func buildDeereIntegrationRows(resp jsonAPIResponse) []deereIntegrationRow {
	rows := make([]deereIntegrationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, deereIntegrationRow{
			ID:                  resource.ID,
			IntegrationID:       stringAttr(resource.Attributes, "integration-identifier"),
			FriendlyName:        stringAttr(resource.Attributes, "friendly-name"),
			BrokerID:            relationshipIDFromMap(resource.Relationships, "broker"),
			IntegrationConfigID: relationshipIDFromMap(resource.Relationships, "integration-config"),
		})
	}
	return rows
}

func renderDeereIntegrationsTable(cmd *cobra.Command, rows []deereIntegrationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No Deere integrations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINTEGRATION ID\tFRIENDLY NAME\tBROKER ID\tINTEGRATION CONFIG")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.IntegrationID, 24),
			truncateString(row.FriendlyName, 30),
			row.BrokerID,
			row.IntegrationConfigID,
		)
	}
	return writer.Flush()
}
