package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	LineupJobScheduleShift string
}

func newDoLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate recommendations for a lineup job schedule shift",
		Long: `Generate trucker assignment recommendations for a lineup job schedule shift.

Required flags:
  --lineup-job-schedule-shift  Lineup job schedule shift ID (required)`,
		Example: `  # Generate recommendations for a lineup job schedule shift
  xbe do lineup-job-schedule-shift-trucker-assignment-recommendations create --lineup-job-schedule-shift 123

  # Output as JSON
  xbe do lineup-job-schedule-shift-trucker-assignment-recommendations create --lineup-job-schedule-shift 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoLineupJobScheduleShiftTruckerAssignmentRecommendationsCreate,
	}
	initDoLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupJobScheduleShiftTruckerAssignmentRecommendationsCmd.AddCommand(newDoLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateCmd())
}

func initDoLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup-job-schedule-shift", "", "Lineup job schedule shift ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupJobScheduleShiftTruckerAssignmentRecommendationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.LineupJobScheduleShift) == "" {
		err := fmt.Errorf("--lineup-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"lineup-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "lineup-job-schedule-shifts",
				"id":   opts.LineupJobScheduleShift,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-job-schedule-shift-trucker-assignment-recommendations",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-job-schedule-shift-trucker-assignment-recommendations", jsonBody)
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

	row := lineupJobScheduleShiftTruckerAssignmentRecommendationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup job schedule shift trucker assignment recommendation %s\n", row.ID)
	return nil
}

func parseDoLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateOptions(cmd *cobra.Command) (doLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupJobScheduleShift, _ := cmd.Flags().GetString("lineup-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupJobScheduleShiftTruckerAssignmentRecommendationsCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		LineupJobScheduleShift: lineupJobScheduleShift,
	}, nil
}
