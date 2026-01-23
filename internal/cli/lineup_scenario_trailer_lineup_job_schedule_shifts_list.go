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

type lineupScenarioTrailerLineupJobScheduleShiftsListOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	NoAuth                               bool
	Limit                                int
	Offset                               int
	Sort                                 string
	LineupScenarioTrailer                string
	LineupScenarioLineupJobScheduleShift string
}

type lineupScenarioTrailerLineupJobScheduleShiftRow struct {
	ID                                     string `json:"id"`
	LineupScenarioTrailerID                string `json:"lineup_scenario_trailer_id,omitempty"`
	LineupScenarioLineupJobScheduleShiftID string `json:"lineup_scenario_lineup_job_schedule_shift_id,omitempty"`
	TrailerID                              string `json:"trailer_id,omitempty"`
	TruckerID                              string `json:"trucker_id,omitempty"`
	LineupJobScheduleShiftID               string `json:"lineup_job_schedule_shift_id,omitempty"`
	StartSiteDistanceMinutes               string `json:"start_site_distance_minutes,omitempty"`
	EndSiteDistanceMinutes                 string `json:"end_site_distance_minutes,omitempty"`
}

func newLineupScenarioTrailerLineupJobScheduleShiftsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup scenario trailer lineup job schedule shifts",
		Long: `List lineup scenario trailer lineup job schedule shifts with filtering and pagination.

Output Columns:
  ID                      Record identifier
  SCENARIO_TRAILER        Lineup scenario trailer ID
  SCENARIO_SHIFT          Lineup scenario lineup job schedule shift ID
  TRAILER                 Trailer ID
  TRUCKER                 Trucker ID
  JOB_SHIFT               Lineup job schedule shift ID
  START_SITE_MIN          Start site distance minutes
  END_SITE_MIN            End site distance minutes

Filters:
  --lineup-scenario-trailer                 Filter by lineup scenario trailer ID
  --lineup-scenario-lineup-job-schedule-shift Filter by lineup scenario lineup job schedule shift ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List records
  xbe view lineup-scenario-trailer-lineup-job-schedule-shifts list

  # Filter by lineup scenario trailer
  xbe view lineup-scenario-trailer-lineup-job-schedule-shifts list --lineup-scenario-trailer 123

  # Filter by lineup scenario lineup job schedule shift
  xbe view lineup-scenario-trailer-lineup-job-schedule-shifts list --lineup-scenario-lineup-job-schedule-shift 456

  # Output as JSON
  xbe view lineup-scenario-trailer-lineup-job-schedule-shifts list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupScenarioTrailerLineupJobScheduleShiftsList,
	}
	initLineupScenarioTrailerLineupJobScheduleShiftsListFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioTrailerLineupJobScheduleShiftsCmd.AddCommand(newLineupScenarioTrailerLineupJobScheduleShiftsListCmd())
}

func initLineupScenarioTrailerLineupJobScheduleShiftsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup-scenario-trailer", "", "Filter by lineup scenario trailer ID")
	cmd.Flags().String("lineup-scenario-lineup-job-schedule-shift", "", "Filter by lineup scenario lineup job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioTrailerLineupJobScheduleShiftsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupScenarioTrailerLineupJobScheduleShiftsListOptions(cmd)
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
	query.Set("fields[lineup-scenario-trailer-lineup-job-schedule-shifts]", "start-site-distance-minutes,end-site-distance-minutes,lineup-scenario-trailer,lineup-scenario-lineup-job-schedule-shift,trailer,trucker,lineup-job-schedule-shift")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[lineup-scenario-trailer]", opts.LineupScenarioTrailer)
	setFilterIfPresent(query, "filter[lineup-scenario-lineup-job-schedule-shift]", opts.LineupScenarioLineupJobScheduleShift)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-trailer-lineup-job-schedule-shifts", query)
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

	rows := buildLineupScenarioTrailerLineupJobScheduleShiftRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupScenarioTrailerLineupJobScheduleShiftsTable(cmd, rows)
}

func parseLineupScenarioTrailerLineupJobScheduleShiftsListOptions(cmd *cobra.Command) (lineupScenarioTrailerLineupJobScheduleShiftsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupScenarioTrailer, _ := cmd.Flags().GetString("lineup-scenario-trailer")
	lineupScenarioLineupJobScheduleShift, _ := cmd.Flags().GetString("lineup-scenario-lineup-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioTrailerLineupJobScheduleShiftsListOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		NoAuth:                               noAuth,
		Limit:                                limit,
		Offset:                               offset,
		Sort:                                 sort,
		LineupScenarioTrailer:                lineupScenarioTrailer,
		LineupScenarioLineupJobScheduleShift: lineupScenarioLineupJobScheduleShift,
	}, nil
}

func buildLineupScenarioTrailerLineupJobScheduleShiftRows(resp jsonAPIResponse) []lineupScenarioTrailerLineupJobScheduleShiftRow {
	rows := make([]lineupScenarioTrailerLineupJobScheduleShiftRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := lineupScenarioTrailerLineupJobScheduleShiftRow{
			ID:                       resource.ID,
			StartSiteDistanceMinutes: stringAttr(resource.Attributes, "start-site-distance-minutes"),
			EndSiteDistanceMinutes:   stringAttr(resource.Attributes, "end-site-distance-minutes"),
		}

		row.LineupScenarioTrailerID = relationshipIDFromMap(resource.Relationships, "lineup-scenario-trailer")
		row.LineupScenarioLineupJobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "lineup-scenario-lineup-job-schedule-shift")
		row.TrailerID = relationshipIDFromMap(resource.Relationships, "trailer")
		row.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
		row.LineupJobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "lineup-job-schedule-shift")

		rows = append(rows, row)
	}
	return rows
}

func renderLineupScenarioTrailerLineupJobScheduleShiftsTable(cmd *cobra.Command, rows []lineupScenarioTrailerLineupJobScheduleShiftRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup scenario trailer lineup job schedule shifts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSCENARIO_TRAILER\tSCENARIO_SHIFT\tTRAILER\tTRUCKER\tJOB_SHIFT\tSTART_SITE_MIN\tEND_SITE_MIN")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.LineupScenarioTrailerID,
			row.LineupScenarioLineupJobScheduleShiftID,
			row.TrailerID,
			row.TruckerID,
			row.LineupJobScheduleShiftID,
			row.StartSiteDistanceMinutes,
			row.EndSiteDistanceMinutes,
		)
	}
	return writer.Flush()
}
