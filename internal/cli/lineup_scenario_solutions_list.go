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

type lineupScenarioSolutionsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	LineupScenario string
}

type lineupScenarioSolutionRow struct {
	ID               string  `json:"id"`
	LineupScenarioID string  `json:"lineup_scenario_id,omitempty"`
	Status           string  `json:"status,omitempty"`
	Cost             float64 `json:"cost,omitempty"`
	SolvedAt         string  `json:"solved_at,omitempty"`
}

func newLineupScenarioSolutionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup scenario solutions",
		Long: `List lineup scenario solutions.

Output Columns:
  ID         Lineup scenario solution ID
  SCENARIO   Lineup scenario ID
  STATUS     Solver status (unsolved/solved/infeasible)
  COST       Total solution cost (if solved)
  SOLVED AT  Solver completion timestamp

Filters:
  --lineup-scenario   Filter by lineup scenario ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup scenario solutions
  xbe view lineup-scenario-solutions list

  # Filter by lineup scenario
  xbe view lineup-scenario-solutions list --lineup-scenario 123

  # Paginate results
  xbe view lineup-scenario-solutions list --limit 25 --offset 50

  # Output as JSON
  xbe view lineup-scenario-solutions list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupScenarioSolutionsList,
	}
	initLineupScenarioSolutionsListFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioSolutionsCmd.AddCommand(newLineupScenarioSolutionsListCmd())
}

func initLineupScenarioSolutionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup-scenario", "", "Filter by lineup scenario ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioSolutionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupScenarioSolutionsListOptions(cmd)
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
	query.Set("fields[lineup-scenario-solutions]", "status,cost,solved-at,lineup-scenario")
	query.Set("include", "lineup-scenario")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[lineup-scenario]", opts.LineupScenario)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-solutions", query)
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

	rows := buildLineupScenarioSolutionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupScenarioSolutionsTable(cmd, rows)
}

func parseLineupScenarioSolutionsListOptions(cmd *cobra.Command) (lineupScenarioSolutionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupScenario, _ := cmd.Flags().GetString("lineup-scenario")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioSolutionsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		LineupScenario: lineupScenario,
	}, nil
}

func buildLineupScenarioSolutionRows(resp jsonAPIResponse) []lineupScenarioSolutionRow {
	rows := make([]lineupScenarioSolutionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildLineupScenarioSolutionRow(resource))
	}
	return rows
}

func buildLineupScenarioSolutionRow(resource jsonAPIResource) lineupScenarioSolutionRow {
	attrs := resource.Attributes
	row := lineupScenarioSolutionRow{
		ID:       resource.ID,
		Status:   stringAttr(attrs, "status"),
		Cost:     floatAttr(attrs, "cost"),
		SolvedAt: formatDate(stringAttr(attrs, "solved-at")),
	}

	if rel, ok := resource.Relationships["lineup-scenario"]; ok && rel.Data != nil {
		row.LineupScenarioID = rel.Data.ID
	}

	return row
}

func buildLineupScenarioSolutionRowFromSingle(resp jsonAPISingleResponse) lineupScenarioSolutionRow {
	return buildLineupScenarioSolutionRow(resp.Data)
}

func renderLineupScenarioSolutionsTable(cmd *cobra.Command, rows []lineupScenarioSolutionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup scenario solutions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSCENARIO\tSTATUS\tCOST\tSOLVED AT")
	for _, row := range rows {
		cost := formatOptionalFloat(row.Cost)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.LineupScenarioID,
			row.Status,
			cost,
			row.SolvedAt,
		)
	}
	return writer.Flush()
}
