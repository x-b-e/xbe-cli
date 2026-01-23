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

type lineupScenarioTruckersListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	LineupScenario string
	Trucker        string
}

type lineupScenarioTruckerRow struct {
	ID                        string `json:"id"`
	LineupScenarioID          string `json:"lineup_scenario_id,omitempty"`
	LineupScenarioName        string `json:"lineup_scenario_name,omitempty"`
	LineupScenarioDate        string `json:"lineup_scenario_date,omitempty"`
	LineupScenarioWindow      string `json:"lineup_scenario_window,omitempty"`
	TruckerID                 string `json:"trucker_id,omitempty"`
	TruckerName               string `json:"trucker_name,omitempty"`
	MinimumAssignmentCount    string `json:"minimum_assignment_count,omitempty"`
	MaximumAssignmentCount    string `json:"maximum_assignment_count,omitempty"`
	MaximumMinutesToStartSite string `json:"maximum_minutes_to_start_site,omitempty"`
}

func newLineupScenarioTruckersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup scenario truckers",
		Long: `List lineup scenario truckers with filtering and pagination.

Output Columns:
  ID        Lineup scenario trucker identifier
  SCENARIO  Lineup scenario name or ID
  DATE      Lineup scenario date
  WINDOW    Lineup scenario window
  TRUCKER   Trucker name or ID
  MIN       Minimum assignment count
  MAX       Maximum assignment count
  MAX MINS  Maximum minutes to start site

Filters:
  --lineup-scenario  Filter by lineup scenario ID
  --trucker          Filter by trucker ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup scenario truckers
  xbe view lineup-scenario-truckers list

  # Filter by lineup scenario
  xbe view lineup-scenario-truckers list --lineup-scenario 123

  # Filter by trucker
  xbe view lineup-scenario-truckers list --trucker 456

  # Output as JSON
  xbe view lineup-scenario-truckers list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupScenarioTruckersList,
	}
	initLineupScenarioTruckersListFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioTruckersCmd.AddCommand(newLineupScenarioTruckersListCmd())
}

func initLineupScenarioTruckersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup-scenario", "", "Filter by lineup scenario ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioTruckersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupScenarioTruckersListOptions(cmd)
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
	query.Set("fields[lineup-scenario-truckers]", "lineup-scenario,trucker,minimum-assignment-count,maximum-assignment-count,maximum-minutes-to-start-site")
	query.Set("include", "lineup-scenario,trucker")
	query.Set("fields[lineup-scenarios]", "name,date,window")
	query.Set("fields[truckers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[lineup-scenario]", opts.LineupScenario)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-truckers", query)
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

	rows := buildLineupScenarioTruckerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupScenarioTruckersTable(cmd, rows)
}

func parseLineupScenarioTruckersListOptions(cmd *cobra.Command) (lineupScenarioTruckersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupScenario, _ := cmd.Flags().GetString("lineup-scenario")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioTruckersListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		LineupScenario: lineupScenario,
		Trucker:        trucker,
	}, nil
}

func buildLineupScenarioTruckerRows(resp jsonAPIResponse) []lineupScenarioTruckerRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]lineupScenarioTruckerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLineupScenarioTruckerRow(resource, included))
	}
	return rows
}

func lineupScenarioTruckerRowFromSingle(resp jsonAPISingleResponse) lineupScenarioTruckerRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildLineupScenarioTruckerRow(resp.Data, included)
}

func buildLineupScenarioTruckerRow(resource jsonAPIResource, included map[string]jsonAPIResource) lineupScenarioTruckerRow {
	row := lineupScenarioTruckerRow{
		ID:                        resource.ID,
		MinimumAssignmentCount:    stringAttr(resource.Attributes, "minimum-assignment-count"),
		MaximumAssignmentCount:    stringAttr(resource.Attributes, "maximum-assignment-count"),
		MaximumMinutesToStartSite: stringAttr(resource.Attributes, "maximum-minutes-to-start-site"),
	}

	if rel, ok := resource.Relationships["lineup-scenario"]; ok && rel.Data != nil {
		row.LineupScenarioID = rel.Data.ID
		if scenario, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.LineupScenarioName = stringAttr(scenario.Attributes, "name")
			row.LineupScenarioDate = stringAttr(scenario.Attributes, "date")
			row.LineupScenarioWindow = stringAttr(scenario.Attributes, "window")
		}
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = stringAttr(trucker.Attributes, "company-name")
		}
	}

	return row
}

func renderLineupScenarioTruckersTable(cmd *cobra.Command, rows []lineupScenarioTruckerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup scenario truckers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSCENARIO\tDATE\tWINDOW\tTRUCKER\tMIN\tMAX\tMAX MINS")
	for _, row := range rows {
		scenario := firstNonEmpty(row.LineupScenarioName, row.LineupScenarioID)
		trucker := firstNonEmpty(row.TruckerName, row.TruckerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(scenario, 24),
			truncateString(row.LineupScenarioDate, 10),
			truncateString(row.LineupScenarioWindow, 10),
			truncateString(trucker, 24),
			truncateString(row.MinimumAssignmentCount, 6),
			truncateString(row.MaximumAssignmentCount, 6),
			truncateString(row.MaximumMinutesToStartSite, 8),
		)
	}
	return writer.Flush()
}
