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

type lineupScenarioTrailersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupScenarioTrailerDetails struct {
	ID                                        string   `json:"id"`
	LineupScenarioTruckerID                   string   `json:"lineup_scenario_trucker_id,omitempty"`
	TrailerID                                 string   `json:"trailer_id,omitempty"`
	TruckerID                                 string   `json:"trucker_id,omitempty"`
	LineupScenarioID                          string   `json:"lineup_scenario_id,omitempty"`
	LastAssignedOn                            string   `json:"last_assigned_on,omitempty"`
	LineupScenarioTrailerLineupJobScheduleIDs []string `json:"lineup_scenario_trailer_lineup_job_schedule_shift_ids,omitempty"`
}

func newLineupScenarioTrailersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup scenario trailer details",
		Long: `Show the full details of a lineup scenario trailer.

Output Fields:
  ID                 Lineup scenario trailer identifier
  Lineup Scenario Trucker  Associated lineup scenario trucker ID
  Trailer            Associated trailer ID
  Trucker            Associated trucker ID
  Lineup Scenario    Associated lineup scenario ID
  Last Assigned On   Last assigned date
  Trailer Shifts     Lineup scenario trailer lineup job schedule shift IDs

Arguments:
  <id>    The lineup scenario trailer ID (required). You can find IDs using the list command.`,
		Example: `  # Show a lineup scenario trailer
  xbe view lineup-scenario-trailers show 123

  # Get JSON output
  xbe view lineup-scenario-trailers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupScenarioTrailersShow,
	}
	initLineupScenarioTrailersShowFlags(cmd)
	return cmd
}

func init() {
	lineupScenarioTrailersCmd.AddCommand(newLineupScenarioTrailersShowCmd())
}

func initLineupScenarioTrailersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupScenarioTrailersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseLineupScenarioTrailersShowOptions(cmd)
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
		return fmt.Errorf("lineup scenario trailer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-scenario-trailers/"+id, query)
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

	details := buildLineupScenarioTrailerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupScenarioTrailerDetails(cmd, details)
}

func parseLineupScenarioTrailersShowOptions(cmd *cobra.Command) (lineupScenarioTrailersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupScenarioTrailersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupScenarioTrailerDetails(resp jsonAPISingleResponse) lineupScenarioTrailerDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := lineupScenarioTrailerDetails{
		ID:             resource.ID,
		LastAssignedOn: formatDate(stringAttr(attrs, "last-assigned-on")),
	}

	if rel, ok := resource.Relationships["lineup-scenario-trucker"]; ok && rel.Data != nil {
		details.LineupScenarioTruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
		details.TrailerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["lineup-scenario"]; ok && rel.Data != nil {
		details.LineupScenarioID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["lineup-scenario-trailer-lineup-job-schedule-shifts"]; ok && rel.raw != nil {
		details.LineupScenarioTrailerLineupJobScheduleIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderLineupScenarioTrailerDetails(cmd *cobra.Command, details lineupScenarioTrailerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupScenarioTruckerID != "" {
		fmt.Fprintf(out, "Lineup Scenario Trucker: %s\n", details.LineupScenarioTruckerID)
	}
	if details.TrailerID != "" {
		fmt.Fprintf(out, "Trailer: %s\n", details.TrailerID)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerID)
	}
	if details.LineupScenarioID != "" {
		fmt.Fprintf(out, "Lineup Scenario: %s\n", details.LineupScenarioID)
	}
	if details.LastAssignedOn != "" {
		fmt.Fprintf(out, "Last Assigned On: %s\n", details.LastAssignedOn)
	}
	if len(details.LineupScenarioTrailerLineupJobScheduleIDs) > 0 {
		fmt.Fprintf(out, "Trailer Shifts: %s\n", strings.Join(details.LineupScenarioTrailerLineupJobScheduleIDs, ", "))
	}

	return nil
}
