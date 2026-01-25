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

type lineupScenarioLineupJobScheduleShiftsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	LineupScenarioID         string
	LineupJobScheduleShiftID string
}

type lineupScenarioLineupJobScheduleShiftRow struct {
	ID                       string `json:"id"`
	LineupScenarioID         string `json:"lineup_scenario_id,omitempty"`
	LineupJobScheduleShiftID string `json:"lineup_job_schedule_shift_id,omitempty"`
}

func newLineupScenarioLineupJobScheduleShiftsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup scenario lineup job schedule shifts",
		Long: `List lineup scenario lineup job schedule shifts.

Output Columns:
  ID        Lineup scenario lineup job schedule shift identifier
  SCENARIO  Lineup scenario ID
  SHIFT     Lineup job schedule shift ID

Filters:
  --lineup-scenario            Filter by lineup scenario ID
  --lineup-job-schedule-shift  Filter by lineup job schedule shift ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup scenario lineup job schedule shifts
  xbe view lineup-scenario-lineup-job-schedule-shifts list

  # Filter by lineup scenario
  xbe view lineup-scenario-lineup-job-schedule-shifts list --lineup-scenario 123

  # Filter by lineup job schedule shift
  xbe view lineup-scenario-lineup-job-schedule-shifts list --lineup-job-schedule-shift 456

  # Output as JSON
  xbe view lineup-scenario-lineup-job-schedule-shifts list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupScenarioLineupJobScheduleShiftsList,
	}
	initLineupScenarioLineupJobScheduleShiftsListFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioLineupJobScheduleShiftsCmd.AddCommand(newLineupScenarioLineupJobScheduleShiftsListCmd())
}

func initLineupScenarioLineupJobScheduleShiftsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup-scenario", "", "Filter by lineup scenario ID")
	cmd.Flags().String("lineup-job-schedule-shift", "", "Filter by lineup job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioLineupJobScheduleShiftsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupScenarioLineupJobScheduleShiftsListOptions(cmd)
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

	setFilterIfPresent(query, "filter[lineup-scenario]", opts.LineupScenarioID)
	setFilterIfPresent(query, "filter[lineup-job-schedule-shift]", opts.LineupJobScheduleShiftID)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-lineup-job-schedule-shifts", query)
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

	rows := buildLineupScenarioLineupJobScheduleShiftRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupScenarioLineupJobScheduleShiftsTable(cmd, rows)
}

func parseLineupScenarioLineupJobScheduleShiftsListOptions(cmd *cobra.Command) (lineupScenarioLineupJobScheduleShiftsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupScenarioID, _ := cmd.Flags().GetString("lineup-scenario")
	lineupJobScheduleShiftID, _ := cmd.Flags().GetString("lineup-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioLineupJobScheduleShiftsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		LineupScenarioID:         lineupScenarioID,
		LineupJobScheduleShiftID: lineupJobScheduleShiftID,
	}, nil
}

func buildLineupScenarioLineupJobScheduleShiftRows(resp jsonAPIResponse) []lineupScenarioLineupJobScheduleShiftRow {
	rows := make([]lineupScenarioLineupJobScheduleShiftRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := lineupScenarioLineupJobScheduleShiftRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["lineup-scenario"]; ok && rel.Data != nil {
			row.LineupScenarioID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["lineup-job-schedule-shift"]; ok && rel.Data != nil {
			row.LineupJobScheduleShiftID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderLineupScenarioLineupJobScheduleShiftsTable(cmd *cobra.Command, rows []lineupScenarioLineupJobScheduleShiftRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup scenario lineup job schedule shifts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSCENARIO\tSHIFT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.LineupScenarioID,
			row.LineupJobScheduleShiftID,
		)
	}
	return writer.Flush()
}
