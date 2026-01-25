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

type lineupScenarioTrailerLineupJobScheduleShiftsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupScenarioTrailerLineupJobScheduleShiftDetails struct {
	ID                                     string `json:"id"`
	LineupScenarioTrailerID                string `json:"lineup_scenario_trailer_id,omitempty"`
	LineupScenarioLineupJobScheduleShiftID string `json:"lineup_scenario_lineup_job_schedule_shift_id,omitempty"`
	TrailerID                              string `json:"trailer_id,omitempty"`
	TruckerID                              string `json:"trucker_id,omitempty"`
	LineupJobScheduleShiftID               string `json:"lineup_job_schedule_shift_id,omitempty"`
	StartSiteDistanceMinutes               string `json:"start_site_distance_minutes,omitempty"`
	EndSiteDistanceMinutes                 string `json:"end_site_distance_minutes,omitempty"`
}

func newLineupScenarioTrailerLineupJobScheduleShiftsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup scenario trailer lineup job schedule shift details",
		Long: `Show the full details of a lineup scenario trailer lineup job schedule shift.

Output Fields:
  ID
  Lineup Scenario Trailer ID
  Lineup Scenario Lineup Job Schedule Shift ID
  Trailer ID
  Trucker ID
  Lineup Job Schedule Shift ID
  Start Site Distance Minutes
  End Site Distance Minutes

Arguments:
  <id>    The record ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a record
  xbe view lineup-scenario-trailer-lineup-job-schedule-shifts show 123

  # JSON output
  xbe view lineup-scenario-trailer-lineup-job-schedule-shifts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupScenarioTrailerLineupJobScheduleShiftsShow,
	}
	initLineupScenarioTrailerLineupJobScheduleShiftsShowFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioTrailerLineupJobScheduleShiftsCmd.AddCommand(newLineupScenarioTrailerLineupJobScheduleShiftsShowCmd())
}

func initLineupScenarioTrailerLineupJobScheduleShiftsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioTrailerLineupJobScheduleShiftsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLineupScenarioTrailerLineupJobScheduleShiftsShowOptions(cmd)
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
		return fmt.Errorf("lineup scenario trailer lineup job schedule shift id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lineup-scenario-trailer-lineup-job-schedule-shifts]", "start-site-distance-minutes,end-site-distance-minutes,lineup-scenario-trailer,lineup-scenario-lineup-job-schedule-shift,trailer,trucker,lineup-job-schedule-shift")

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-trailer-lineup-job-schedule-shifts/"+id, query)
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

	details := buildLineupScenarioTrailerLineupJobScheduleShiftDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupScenarioTrailerLineupJobScheduleShiftDetails(cmd, details)
}

func parseLineupScenarioTrailerLineupJobScheduleShiftsShowOptions(cmd *cobra.Command) (lineupScenarioTrailerLineupJobScheduleShiftsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioTrailerLineupJobScheduleShiftsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupScenarioTrailerLineupJobScheduleShiftDetails(resp jsonAPISingleResponse) lineupScenarioTrailerLineupJobScheduleShiftDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := lineupScenarioTrailerLineupJobScheduleShiftDetails{
		ID:                       resource.ID,
		StartSiteDistanceMinutes: stringAttr(attrs, "start-site-distance-minutes"),
		EndSiteDistanceMinutes:   stringAttr(attrs, "end-site-distance-minutes"),
	}

	details.LineupScenarioTrailerID = relationshipIDFromMap(resource.Relationships, "lineup-scenario-trailer")
	details.LineupScenarioLineupJobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "lineup-scenario-lineup-job-schedule-shift")
	details.TrailerID = relationshipIDFromMap(resource.Relationships, "trailer")
	details.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
	details.LineupJobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "lineup-job-schedule-shift")

	return details
}

func renderLineupScenarioTrailerLineupJobScheduleShiftDetails(cmd *cobra.Command, details lineupScenarioTrailerLineupJobScheduleShiftDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupScenarioTrailerID != "" {
		fmt.Fprintf(out, "Lineup Scenario Trailer: %s\n", details.LineupScenarioTrailerID)
	}
	if details.LineupScenarioLineupJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Lineup Scenario Lineup Job Schedule Shift: %s\n", details.LineupScenarioLineupJobScheduleShiftID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer: %s\n", details.TrailerID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
	}
	if details.LineupJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Lineup Job Schedule Shift: %s\n", details.LineupJobScheduleShiftID)
	}
	if details.StartSiteDistanceMinutes != "" {
		fmt.Fprintf(out, "Start Site Distance Minutes: %s\n", details.StartSiteDistanceMinutes)
	}
	if details.EndSiteDistanceMinutes != "" {
		fmt.Fprintf(out, "End Site Distance Minutes: %s\n", details.EndSiteDistanceMinutes)
	}
	return nil
}
