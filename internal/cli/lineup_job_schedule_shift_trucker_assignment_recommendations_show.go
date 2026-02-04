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

type lineupJobScheduleShiftTruckerAssignmentRecommendationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupJobScheduleShiftTruckerAssignmentRecommendationDetails struct {
	ID                       string `json:"id"`
	LineupJobScheduleShiftID string `json:"lineup_job_schedule_shift_id,omitempty"`
	Candidates               any    `json:"candidates,omitempty"`
}

func newLineupJobScheduleShiftTruckerAssignmentRecommendationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup job schedule shift trucker assignment recommendation details",
		Long: `Show the full details of a lineup job schedule shift trucker assignment recommendation.

Output Fields:
  ID
  Lineup Job Schedule Shift ID
  Candidates (ranked truckers with scores and probabilities)

Arguments:
  <id>    The recommendation ID (required). Use the list command to find IDs.`,
		Example: `  # Show a recommendation
  xbe view lineup-job-schedule-shift-trucker-assignment-recommendations show 123

  # Show as JSON
  xbe view lineup-job-schedule-shift-trucker-assignment-recommendations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupJobScheduleShiftTruckerAssignmentRecommendationsShow,
	}
	initLineupJobScheduleShiftTruckerAssignmentRecommendationsShowFlags(cmd)
	return cmd
}

func init() {
	lineupJobScheduleShiftTruckerAssignmentRecommendationsCmd.AddCommand(newLineupJobScheduleShiftTruckerAssignmentRecommendationsShowCmd())
}

func initLineupJobScheduleShiftTruckerAssignmentRecommendationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupJobScheduleShiftTruckerAssignmentRecommendationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseLineupJobScheduleShiftTruckerAssignmentRecommendationsShowOptions(cmd)
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
		return fmt.Errorf("lineup job schedule shift trucker assignment recommendation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[lineup-job-schedule-shift-trucker-assignment-recommendations]", "lineup-job-schedule-shift,candidates")

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-job-schedule-shift-trucker-assignment-recommendations/"+id, query)
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

	details := buildLineupJobScheduleShiftTruckerAssignmentRecommendationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupJobScheduleShiftTruckerAssignmentRecommendationDetails(cmd, details)
}

func parseLineupJobScheduleShiftTruckerAssignmentRecommendationsShowOptions(cmd *cobra.Command) (lineupJobScheduleShiftTruckerAssignmentRecommendationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupJobScheduleShiftTruckerAssignmentRecommendationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupJobScheduleShiftTruckerAssignmentRecommendationDetails(resp jsonAPISingleResponse) lineupJobScheduleShiftTruckerAssignmentRecommendationDetails {
	resource := resp.Data
	details := lineupJobScheduleShiftTruckerAssignmentRecommendationDetails{
		ID:         resource.ID,
		Candidates: resource.Attributes["candidates"],
	}

	if rel, ok := resource.Relationships["lineup-job-schedule-shift"]; ok && rel.Data != nil {
		details.LineupJobScheduleShiftID = rel.Data.ID
	}

	return details
}

func renderLineupJobScheduleShiftTruckerAssignmentRecommendationDetails(cmd *cobra.Command, details lineupJobScheduleShiftTruckerAssignmentRecommendationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Lineup Job Schedule Shift ID: %s\n", details.LineupJobScheduleShiftID)
	}

	if details.Candidates != nil {
		fmt.Fprintf(out, "Candidates: %d\n", candidateCountFromAny(details.Candidates))
		formatted := formatCandidates(details.Candidates)
		if formatted != "" {
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Candidate Details:")
			fmt.Fprintln(out, strings.Repeat("-", 40))
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}

func formatCandidates(value any) string {
	if value == nil {
		return ""
	}
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(pretty)
}
