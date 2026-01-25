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

type lineupScenarioTrailersListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	LineupScenarioTruckerID string
	LineupScenarioID        string
	LineupScenarioIDAlt     string
	TruckerID               string
	TruckerIDAlt            string
	TrailerID               string
	LastAssignedOn          string
	LastAssignedOnMin       string
	LastAssignedOnMax       string
	HasLastAssignedOn       string
}

type lineupScenarioTrailerRow struct {
	ID                      string `json:"id"`
	LineupScenarioTruckerID string `json:"lineup_scenario_trucker_id,omitempty"`
	TrailerID               string `json:"trailer_id,omitempty"`
	LastAssignedOn          string `json:"last_assigned_on,omitempty"`
}

func newLineupScenarioTrailersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup scenario trailers",
		Long: `List lineup scenario trailers.

Output Columns:
  ID               Lineup scenario trailer identifier
  SCENARIO_TRUCKER Lineup scenario trucker ID
  TRAILER          Trailer ID
  LAST_ASSIGNED    Last assigned date

Filters:
  --lineup-scenario-trucker  Filter by lineup scenario trucker ID
  --lineup-scenario          Filter by lineup scenario ID
  --lineup-scenario-id       Filter by lineup scenario ID (via trucker)
  --trucker                  Filter by trucker ID
  --trucker-id               Filter by trucker ID (via trucker)
  --trailer                  Filter by trailer ID
  --last-assigned-on         Filter by last assigned date (YYYY-MM-DD)
  --last-assigned-on-min     Filter by minimum last assigned date (YYYY-MM-DD)
  --last-assigned-on-max     Filter by maximum last assigned date (YYYY-MM-DD)
  --has-last-assigned-on     Filter by presence of last assigned date (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup scenario trailers
  xbe view lineup-scenario-trailers list

  # Filter by lineup scenario
  xbe view lineup-scenario-trailers list --lineup-scenario 123

  # Filter by lineup scenario trucker
  xbe view lineup-scenario-trailers list --lineup-scenario-trucker 456

  # Filter by trailer
  xbe view lineup-scenario-trailers list --trailer 789

  # Filter by last assigned date
  xbe view lineup-scenario-trailers list --last-assigned-on-min 2024-01-01 --last-assigned-on-max 2024-12-31

  # Output as JSON
  xbe view lineup-scenario-trailers list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupScenarioTrailersList,
	}
	initLineupScenarioTrailersListFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioTrailersCmd.AddCommand(newLineupScenarioTrailersListCmd())
}

func initLineupScenarioTrailersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup-scenario-trucker", "", "Filter by lineup scenario trucker ID")
	cmd.Flags().String("lineup-scenario", "", "Filter by lineup scenario ID")
	cmd.Flags().String("lineup-scenario-id", "", "Filter by lineup scenario ID (via trucker)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trucker-id", "", "Filter by trucker ID (via trucker)")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("last-assigned-on", "", "Filter by last assigned date (YYYY-MM-DD)")
	cmd.Flags().String("last-assigned-on-min", "", "Filter by minimum last assigned date (YYYY-MM-DD)")
	cmd.Flags().String("last-assigned-on-max", "", "Filter by maximum last assigned date (YYYY-MM-DD)")
	cmd.Flags().String("has-last-assigned-on", "", "Filter by presence of last assigned date (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioTrailersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupScenarioTrailersListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[lineup-scenario-trucker]", opts.LineupScenarioTruckerID)
	setFilterIfPresent(query, "filter[lineup-scenario]", opts.LineupScenarioID)
	setFilterIfPresent(query, "filter[lineup-scenario-id]", opts.LineupScenarioIDAlt)
	setFilterIfPresent(query, "filter[trucker]", opts.TruckerID)
	setFilterIfPresent(query, "filter[trucker-id]", opts.TruckerIDAlt)
	setFilterIfPresent(query, "filter[trailer]", opts.TrailerID)
	setFilterIfPresent(query, "filter[last-assigned-on]", opts.LastAssignedOn)
	setFilterIfPresent(query, "filter[last-assigned-on-min]", opts.LastAssignedOnMin)
	setFilterIfPresent(query, "filter[last-assigned-on-max]", opts.LastAssignedOnMax)
	setFilterIfPresent(query, "filter[has-last-assigned-on]", opts.HasLastAssignedOn)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-trailers", query)
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

	rows := buildLineupScenarioTrailerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupScenarioTrailersTable(cmd, rows)
}

func parseLineupScenarioTrailersListOptions(cmd *cobra.Command) (lineupScenarioTrailersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupScenarioTruckerID, _ := cmd.Flags().GetString("lineup-scenario-trucker")
	lineupScenarioID, _ := cmd.Flags().GetString("lineup-scenario")
	lineupScenarioIDAlt, _ := cmd.Flags().GetString("lineup-scenario-id")
	truckerID, _ := cmd.Flags().GetString("trucker")
	truckerIDAlt, _ := cmd.Flags().GetString("trucker-id")
	trailerID, _ := cmd.Flags().GetString("trailer")
	lastAssignedOn, _ := cmd.Flags().GetString("last-assigned-on")
	lastAssignedOnMin, _ := cmd.Flags().GetString("last-assigned-on-min")
	lastAssignedOnMax, _ := cmd.Flags().GetString("last-assigned-on-max")
	hasLastAssignedOn, _ := cmd.Flags().GetString("has-last-assigned-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioTrailersListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		LineupScenarioTruckerID: lineupScenarioTruckerID,
		LineupScenarioID:        lineupScenarioID,
		LineupScenarioIDAlt:     lineupScenarioIDAlt,
		TruckerID:               truckerID,
		TruckerIDAlt:            truckerIDAlt,
		TrailerID:               trailerID,
		LastAssignedOn:          lastAssignedOn,
		LastAssignedOnMin:       lastAssignedOnMin,
		LastAssignedOnMax:       lastAssignedOnMax,
		HasLastAssignedOn:       hasLastAssignedOn,
	}, nil
}

func buildLineupScenarioTrailerRows(resp jsonAPIResponse) []lineupScenarioTrailerRow {
	rows := make([]lineupScenarioTrailerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := lineupScenarioTrailerRow{
			ID:             resource.ID,
			LastAssignedOn: formatDate(stringAttr(attrs, "last-assigned-on")),
		}

		if rel, ok := resource.Relationships["lineup-scenario-trucker"]; ok && rel.Data != nil {
			row.LineupScenarioTruckerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
			row.TrailerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderLineupScenarioTrailersTable(cmd *cobra.Command, rows []lineupScenarioTrailerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup scenario trailers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSCENARIO_TRUCKER\tTRAILER\tLAST_ASSIGNED")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.LineupScenarioTruckerID,
			row.TrailerID,
			row.LastAssignedOn,
		)
	}
	return writer.Flush()
}
