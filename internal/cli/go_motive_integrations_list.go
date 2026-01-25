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

type goMotiveIntegrationsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Broker       string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type goMotiveIntegrationRow struct {
	ID                    string `json:"id"`
	IntegrationIdentifier string `json:"integration_identifier,omitempty"`
	FriendlyName          string `json:"friendly_name,omitempty"`
	BrokerID              string `json:"broker_id,omitempty"`
	BrokerName            string `json:"broker_name,omitempty"`
	IntegrationConfigID   string `json:"integration_config_id,omitempty"`
	IntegrationConfigName string `json:"integration_config_name,omitempty"`
}

func newGoMotiveIntegrationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List GoMotive integrations",
		Long: `List GoMotive integrations with filtering and pagination.

Output Columns:
  ID                 GoMotive integration identifier
  NAME               Friendly name
  INTEGRATION ID     GoMotive integration identifier
  INTEGRATION CONFIG Integration config name or ID
  BROKER             Broker name or ID

Filters:
  --broker            Filter by broker ID
  --created-at-min    Filter by created-at on/after (ISO 8601)
  --created-at-max    Filter by created-at on/before (ISO 8601)
  --is-created-at     Filter by has created-at (true/false)
  --updated-at-min    Filter by updated-at on/after (ISO 8601)
  --updated-at-max    Filter by updated-at on/before (ISO 8601)
  --is-updated-at     Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List GoMotive integrations
  xbe view go-motive-integrations list

  # Filter by broker
  xbe view go-motive-integrations list --broker 123

  # Output as JSON
  xbe view go-motive-integrations list --json`,
		Args: cobra.NoArgs,
		RunE: runGoMotiveIntegrationsList,
	}
	initGoMotiveIntegrationsListFlags(cmd)
	return cmd
}

func init() {
	goMotiveIntegrationsCmd.AddCommand(newGoMotiveIntegrationsListCmd())
}

func initGoMotiveIntegrationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGoMotiveIntegrationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseGoMotiveIntegrationsListOptions(cmd)
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
	query.Set("fields[go-motive-integrations]", "integration-identifier,friendly-name,broker,integration-config")
	query.Set("include", "broker,integration-config")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[integration-configs]", "friendly-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/go-motive-integrations", query)
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

	rows := buildGoMotiveIntegrationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderGoMotiveIntegrationsTable(cmd, rows)
}

func parseGoMotiveIntegrationsListOptions(cmd *cobra.Command) (goMotiveIntegrationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return goMotiveIntegrationsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Broker:       broker,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildGoMotiveIntegrationRows(resp jsonAPIResponse) []goMotiveIntegrationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]goMotiveIntegrationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildGoMotiveIntegrationRow(resource, included))
	}
	return rows
}

func goMotiveIntegrationRowFromSingle(resp jsonAPISingleResponse) goMotiveIntegrationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildGoMotiveIntegrationRow(resp.Data, included)
}

func buildGoMotiveIntegrationRow(resource jsonAPIResource, included map[string]jsonAPIResource) goMotiveIntegrationRow {
	attrs := resource.Attributes
	row := goMotiveIntegrationRow{
		ID:                    resource.ID,
		IntegrationIdentifier: stringAttr(attrs, "integration-identifier"),
		FriendlyName:          stringAttr(attrs, "friendly-name"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["integration-config"]; ok && rel.Data != nil {
		row.IntegrationConfigID = rel.Data.ID
		if config, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.IntegrationConfigName = stringAttr(config.Attributes, "friendly-name")
		}
	}

	return row
}

func renderGoMotiveIntegrationsTable(cmd *cobra.Command, rows []goMotiveIntegrationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No GoMotive integrations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tINTEGRATION ID\tINTEGRATION CONFIG\tBROKER")
	for _, row := range rows {
		config := firstNonEmpty(row.IntegrationConfigName, row.IntegrationConfigID)
		broker := firstNonEmpty(row.BrokerName, row.BrokerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.FriendlyName, 24),
			truncateString(row.IntegrationIdentifier, 24),
			truncateString(config, 24),
			truncateString(broker, 24),
		)
	}
	return writer.Flush()
}
