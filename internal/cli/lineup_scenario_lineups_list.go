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

type lineupScenarioLineupsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	LineupScenario string
	Lineup         string
}

type lineupScenarioLineupRow struct {
	ID                   string `json:"id"`
	LineupScenarioID     string `json:"lineup_scenario_id,omitempty"`
	LineupScenarioName   string `json:"lineup_scenario_name,omitempty"`
	LineupScenarioDate   string `json:"lineup_scenario_date,omitempty"`
	LineupScenarioWindow string `json:"lineup_scenario_window,omitempty"`
	LineupID             string `json:"lineup_id,omitempty"`
	LineupName           string `json:"lineup_name,omitempty"`
	LineupStartAtMin     string `json:"lineup_start_at_min,omitempty"`
	LineupStartAtMax     string `json:"lineup_start_at_max,omitempty"`
}

func newLineupScenarioLineupsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup scenario lineups",
		Long: `List lineup scenario lineups with filtering and pagination.

Output Columns:
  ID        Lineup scenario lineup identifier
  SCENARIO  Lineup scenario name or ID
  DATE      Lineup scenario date
  WINDOW    Lineup scenario window
  LINEUP    Lineup name or ID
  START     Lineup start window

Filters:
  --lineup-scenario  Filter by lineup scenario ID
  --lineup           Filter by lineup ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup scenario lineups
  xbe view lineup-scenario-lineups list

  # Filter by lineup scenario
  xbe view lineup-scenario-lineups list --lineup-scenario 123

  # Filter by lineup
  xbe view lineup-scenario-lineups list --lineup 456

  # Output as JSON
  xbe view lineup-scenario-lineups list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupScenarioLineupsList,
	}
	initLineupScenarioLineupsListFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioLineupsCmd.AddCommand(newLineupScenarioLineupsListCmd())
}

func initLineupScenarioLineupsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup-scenario", "", "Filter by lineup scenario ID")
	cmd.Flags().String("lineup", "", "Filter by lineup ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioLineupsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupScenarioLineupsListOptions(cmd)
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
	query.Set("fields[lineup-scenario-lineups]", "lineup-scenario,lineup")
	query.Set("include", "lineup-scenario,lineup")
	query.Set("fields[lineup-scenarios]", "name,date,window")
	query.Set("fields[lineups]", "name,start-at-min,start-at-max")

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
	setFilterIfPresent(query, "filter[lineup]", opts.Lineup)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-lineups", query)
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

	rows := buildLineupScenarioLineupRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupScenarioLineupsTable(cmd, rows)
}

func parseLineupScenarioLineupsListOptions(cmd *cobra.Command) (lineupScenarioLineupsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupScenario, _ := cmd.Flags().GetString("lineup-scenario")
	lineup, _ := cmd.Flags().GetString("lineup")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioLineupsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		LineupScenario: lineupScenario,
		Lineup:         lineup,
	}, nil
}

func buildLineupScenarioLineupRows(resp jsonAPIResponse) []lineupScenarioLineupRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]lineupScenarioLineupRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLineupScenarioLineupRow(resource, included))
	}
	return rows
}

func lineupScenarioLineupRowFromSingle(resp jsonAPISingleResponse) lineupScenarioLineupRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildLineupScenarioLineupRow(resp.Data, included)
}

func buildLineupScenarioLineupRow(resource jsonAPIResource, included map[string]jsonAPIResource) lineupScenarioLineupRow {
	row := lineupScenarioLineupRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["lineup-scenario"]; ok && rel.Data != nil {
		row.LineupScenarioID = rel.Data.ID
		if scenario, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.LineupScenarioName = stringAttr(scenario.Attributes, "name")
			row.LineupScenarioDate = stringAttr(scenario.Attributes, "date")
			row.LineupScenarioWindow = stringAttr(scenario.Attributes, "window")
		}
	}

	if rel, ok := resource.Relationships["lineup"]; ok && rel.Data != nil {
		row.LineupID = rel.Data.ID
		if lineup, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.LineupName = stringAttr(lineup.Attributes, "name")
			row.LineupStartAtMin = stringAttr(lineup.Attributes, "start-at-min")
			row.LineupStartAtMax = stringAttr(lineup.Attributes, "start-at-max")
		}
	}

	return row
}

func renderLineupScenarioLineupsTable(cmd *cobra.Command, rows []lineupScenarioLineupRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup scenario lineups found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSCENARIO\tDATE\tWINDOW\tLINEUP\tSTART")
	for _, row := range rows {
		scenario := firstNonEmpty(row.LineupScenarioName, row.LineupScenarioID)
		lineup := firstNonEmpty(row.LineupName, row.LineupID)
		start := firstNonEmpty(row.LineupStartAtMin, row.LineupStartAtMax)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(scenario, 24),
			truncateString(row.LineupScenarioDate, 10),
			truncateString(row.LineupScenarioWindow, 10),
			truncateString(lineup, 24),
			truncateString(start, 20),
		)
	}
	return writer.Flush()
}
