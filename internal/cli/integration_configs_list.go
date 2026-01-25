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

type integrationConfigsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	FriendlyName string
	Broker       string
	Organization string
}

type integrationConfigRow struct {
	ID               string `json:"id"`
	FriendlyName     string `json:"friendly_name,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	BrokerName       string `json:"broker_name,omitempty"`
	Organization     string `json:"organization,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
}

func newIntegrationConfigsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List integration configs",
		Long: `List integration configs with filtering and pagination.

Output Columns:
  ID            Integration config identifier
  NAME          Friendly name
  BROKER        Broker name or ID
  ORGANIZATION  Organization name or Type/ID

Filters:
  --friendly-name  Filter by friendly name
  --broker         Filter by broker ID
  --organization   Filter by organization (format: Type|ID, e.g., Broker|123)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List integration configs
  xbe view integration-configs list

  # Filter by broker
  xbe view integration-configs list --broker 123

  # Filter by organization
  xbe view integration-configs list --organization "Broker|123"

  # Output as JSON
  xbe view integration-configs list --json`,
		Args: cobra.NoArgs,
		RunE: runIntegrationConfigsList,
	}
	initIntegrationConfigsListFlags(cmd)
	return cmd
}

func init() {
	integrationConfigsCmd.AddCommand(newIntegrationConfigsListCmd())
}

func initIntegrationConfigsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("friendly-name", "", "Filter by friendly name")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g., Broker|123)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIntegrationConfigsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIntegrationConfigsListOptions(cmd)
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
	query.Set("fields[integration-configs]", strings.Join([]string{
		"friendly-name",
		"broker",
		"organization",
	}, ","))
	query.Set("include", "broker,organization")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[developers]", "name")
	query.Set("fields[material-suppliers]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[friendly_name]", opts.FriendlyName)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)

	body, _, err := client.Get(cmd.Context(), "/v1/integration-configs", query)
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

	rows := buildIntegrationConfigRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIntegrationConfigsTable(cmd, rows)
}

func parseIntegrationConfigsListOptions(cmd *cobra.Command) (integrationConfigsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	friendlyName, _ := cmd.Flags().GetString("friendly-name")
	broker, _ := cmd.Flags().GetString("broker")
	organization, _ := cmd.Flags().GetString("organization")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return integrationConfigsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		FriendlyName: friendlyName,
		Broker:       broker,
		Organization: organization,
	}, nil
}

func buildIntegrationConfigRows(resp jsonAPIResponse) []integrationConfigRow {
	rows := make([]integrationConfigRow, 0, len(resp.Data))
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := integrationConfigRow{
			ID:           resource.ID,
			FriendlyName: stringAttr(attrs, "friendly-name"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
			}
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
			if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Organization = organizationNameFromIncluded(org)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderIntegrationConfigsTable(cmd *cobra.Command, rows []integrationConfigRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No integration configs found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER\tORGANIZATION")
	for _, row := range rows {
		broker := formatRelated(row.BrokerName, row.BrokerID)
		organization := formatRelated(row.Organization, formatPolymorphic(row.OrganizationType, row.OrganizationID))
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.FriendlyName, 30),
			truncateString(broker, 24),
			truncateString(organization, 30),
		)
	}
	return writer.Flush()
}
