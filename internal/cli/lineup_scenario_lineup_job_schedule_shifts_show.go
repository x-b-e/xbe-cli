package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type lineupScenarioLineupJobScheduleShiftsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupScenarioLineupJobScheduleShiftDetails struct {
	ID                                           string   `json:"id"`
	LineupScenarioID                             string   `json:"lineup_scenario_id,omitempty"`
	LineupJobScheduleShiftID                     string   `json:"lineup_job_schedule_shift_id,omitempty"`
	LineupScenarioTrailerLineupJobScheduleShifts []string `json:"lineup_scenario_trailer_lineup_job_schedule_shift_ids,omitempty"`
}

func newLineupScenarioLineupJobScheduleShiftsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup scenario lineup job schedule shift details",
		Long: `Show the full details of a lineup scenario lineup job schedule shift.

Output Fields:
  ID                         Lineup scenario lineup job schedule shift identifier
  Lineup Scenario            Associated lineup scenario ID
  Lineup Job Schedule Shift  Associated lineup job schedule shift ID
  Trailer Shifts             Lineup scenario trailer lineup job schedule shift IDs

Arguments:
  <id>    The lineup scenario lineup job schedule shift ID (required). You can find IDs using the list command.`,
		Example: `  # Show a lineup scenario lineup job schedule shift
  xbe view lineup-scenario-lineup-job-schedule-shifts show 123

  # Get JSON output
  xbe view lineup-scenario-lineup-job-schedule-shifts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupScenarioLineupJobScheduleShiftsShow,
	}
	initLineupScenarioLineupJobScheduleShiftsShowFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioLineupJobScheduleShiftsCmd.AddCommand(newLineupScenarioLineupJobScheduleShiftsShowCmd())
}

func initLineupScenarioLineupJobScheduleShiftsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioLineupJobScheduleShiftsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseLineupScenarioLineupJobScheduleShiftsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("lineup scenario lineup job schedule shift id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-lineup-job-schedule-shifts/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildLineupScenarioLineupJobScheduleShiftDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupScenarioLineupJobScheduleShiftDetails(cmd, details)
}

func parseLineupScenarioLineupJobScheduleShiftsShowOptions(cmd *cobra.Command) (lineupScenarioLineupJobScheduleShiftsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioLineupJobScheduleShiftsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupScenarioLineupJobScheduleShiftDetails(resp jsonAPISingleResponse) lineupScenarioLineupJobScheduleShiftDetails {
	resource := resp.Data
	details := lineupScenarioLineupJobScheduleShiftDetails{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["lineup-scenario"]; ok && rel.Data != nil {
		details.LineupScenarioID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["lineup-job-schedule-shift"]; ok && rel.Data != nil {
		details.LineupJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["lineup-scenario-trailer-lineup-job-schedule-shifts"]; ok && rel.raw != nil {
		details.LineupScenarioTrailerLineupJobScheduleShifts = relationshipIDStrings(rel)
	}

	return details
}

func renderLineupScenarioLineupJobScheduleShiftDetails(cmd *cobra.Command, details lineupScenarioLineupJobScheduleShiftDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupScenarioID != "" {
		fmt.Fprintf(out, "Lineup Scenario: %s\n", details.LineupScenarioID)
	}
	if details.LineupJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Lineup Job Schedule Shift: %s\n", details.LineupJobScheduleShiftID)
	}
	if len(details.LineupScenarioTrailerLineupJobScheduleShifts) > 0 {
		fmt.Fprintf(out, "Trailer Shifts: %s\n", strings.Join(details.LineupScenarioTrailerLineupJobScheduleShifts, ", "))
	}

	return nil
}
