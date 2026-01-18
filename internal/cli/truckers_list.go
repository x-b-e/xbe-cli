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

type truckersListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Name     string
	IsActive bool
}

func newTruckersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List truckers",
		Long: `List truckers with filtering and pagination.

Returns a list of truckers matching the specified criteria.

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Use Case:
  Find trucker IDs for filtering posts by creator:
    xbe view posts list --creator "Trucker|<id>"`,
		Example: `  # List truckers
  xbe view truckers list

  # Search by company name
  xbe view truckers list --name "Acme"

  # Filter by active status
  xbe view truckers list --active

  # Paginate results
  xbe view truckers list --limit 20 --offset 40

  # Output as JSON
  xbe view truckers list --json`,
		RunE: runTruckersList,
	}
	initTruckersListFlags(cmd)
	return cmd
}

func init() {
	truckersCmd.AddCommand(newTruckersListCmd())
}

func initTruckersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by company name (partial match)")
	cmd.Flags().Bool("active", false, "Filter to only active truckers")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckersListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[truckers]", "company-name,is-active,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[company_name]", opts.Name)
	if opts.IsActive {
		query.Set("filter[is_active]", "true")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/truckers", query)
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

	if opts.JSON {
		rows := buildTruckerRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckersList(cmd, resp)
}

func parseTruckersListOptions(cmd *cobra.Command) (truckersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return truckersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return truckersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return truckersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return truckersListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return truckersListOptions{}, err
	}
	isActive, err := cmd.Flags().GetBool("active")
	if err != nil {
		return truckersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return truckersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return truckersListOptions{}, err
	}

	return truckersListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Name:     name,
		IsActive: isActive,
	}, nil
}

type truckerRow struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Broker   string `json:"broker"`
	IsActive bool   `json:"is_active"`
}

func buildTruckerRows(resp jsonAPIResponse) []truckerRow {
	// Build included map for broker lookup
	included := make(map[string]map[string]any)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc.Attributes
	}

	rows := make([]truckerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		brokerName := ""
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if attrs, ok := included[key]; ok {
				brokerName = strings.TrimSpace(stringAttr(attrs, "company-name"))
			}
		}

		rows = append(rows, truckerRow{
			ID:       resource.ID,
			Name:     strings.TrimSpace(stringAttr(resource.Attributes, "company-name")),
			Broker:   brokerName,
			IsActive: boolAttr(resource.Attributes, "is-active"),
		})
	}
	return rows
}

func renderTruckersList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildTruckerRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No truckers found.")
		return nil
	}

	const nameMax = 50
	const brokerMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(row.Broker, brokerMax),
		)
	}
	return writer.Flush()
}
