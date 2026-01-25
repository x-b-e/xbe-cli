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

type serviceSitesListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Name    string
	Broker  string
}

type serviceSiteRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Broker   string `json:"broker,omitempty"`
	BrokerID string `json:"broker_id,omitempty"`
}

func newServiceSitesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service sites",
		Long: `List service sites with filtering and pagination.

Service sites are locations used for service work orders. Use this command
for looking up service site IDs.

Output Columns:
  ID       Service site identifier
  NAME     Service site name
  ADDRESS  Service site address
  BROKER   Broker name`,
		Example: `  # List service sites
  xbe view service-sites list

  # Filter by name
  xbe view service-sites list --name "North Yard"

  # Filter by broker
  xbe view service-sites list --broker 123

  # Output as JSON
  xbe view service-sites list --json`,
		RunE: runServiceSitesList,
	}
	initServiceSitesListFlags(cmd)
	return cmd
}

func init() {
	serviceSitesCmd.AddCommand(newServiceSitesListCmd())
}

func initServiceSitesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by name (exact match)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceSitesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseServiceSitesListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[service-sites]", "name,address,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/service-sites", query)
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

	rows := buildServiceSiteRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderServiceSitesTable(cmd, rows)
}

func parseServiceSitesListOptions(cmd *cobra.Command) (serviceSitesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return serviceSitesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return serviceSitesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return serviceSitesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return serviceSitesListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return serviceSitesListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return serviceSitesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return serviceSitesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return serviceSitesListOptions{}, err
	}

	return serviceSitesListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Name:    name,
		Broker:  broker,
	}, nil
}

func buildServiceSiteRows(resp jsonAPIResponse) []serviceSiteRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]serviceSiteRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := serviceSiteRow{
			ID:      resource.ID,
			Name:    stringAttr(resource.Attributes, "name"),
			Address: stringAttr(resource.Attributes, "address"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
			if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Broker = stringAttr(broker.Attributes, "company-name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderServiceSitesTable(cmd *cobra.Command, rows []serviceSiteRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No service sites found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tADDRESS\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 40),
			truncateString(row.Address, 40),
			truncateString(row.Broker, 25),
		)
	}
	return writer.Flush()
}
