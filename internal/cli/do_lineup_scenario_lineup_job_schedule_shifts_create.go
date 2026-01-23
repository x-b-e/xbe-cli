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

type doLineupScenarioLineupJobScheduleShiftsCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	LineupScenarioID         string
	LineupJobScheduleShiftID string
}

func newDoLineupScenarioLineupJobScheduleShiftsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup scenario lineup job schedule shift",
		Long: `Create a lineup scenario lineup job schedule shift.

Required:
  --lineup-scenario            Lineup scenario ID
  --lineup-job-schedule-shift  Lineup job schedule shift ID`,
		Example: `  # Create a lineup scenario lineup job schedule shift
  xbe do lineup-scenario-lineup-job-schedule-shifts create --lineup-scenario 123 --lineup-job-schedule-shift 456`,
		RunE: runDoLineupScenarioLineupJobScheduleShiftsCreate,
	}
	initDoLineupScenarioLineupJobScheduleShiftsCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioLineupJobScheduleShiftsCmd.AddCommand(newDoLineupScenarioLineupJobScheduleShiftsCreateCmd())
}

func initDoLineupScenarioLineupJobScheduleShiftsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup-scenario", "", "Lineup scenario ID")
	cmd.Flags().String("lineup-job-schedule-shift", "", "Lineup job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("lineup-scenario")
	_ = cmd.MarkFlagRequired("lineup-job-schedule-shift")
}

func runDoLineupScenarioLineupJobScheduleShiftsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupScenarioLineupJobScheduleShiftsCreateOptions(cmd)
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

	relationships := map[string]any{
		"lineup-scenario": map[string]any{
			"data": map[string]any{
				"type": "lineup-scenarios",
				"id":   opts.LineupScenarioID,
			},
		},
		"lineup-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "lineup-job-schedule-shifts",
				"id":   opts.LineupJobScheduleShiftID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "lineup-scenario-lineup-job-schedule-shifts",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-scenario-lineup-job-schedule-shifts", jsonBody)
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

	if opts.JSON {
		row := lineupScenarioLineupJobScheduleShiftRow{
			ID: resp.Data.ID,
		}
		if rel, ok := resp.Data.Relationships["lineup-scenario"]; ok && rel.Data != nil {
			row.LineupScenarioID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["lineup-job-schedule-shift"]; ok && rel.Data != nil {
			row.LineupJobScheduleShiftID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup scenario lineup job schedule shift %s\n", resp.Data.ID)
	return nil
}

func parseDoLineupScenarioLineupJobScheduleShiftsCreateOptions(cmd *cobra.Command) (doLineupScenarioLineupJobScheduleShiftsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupScenarioID, _ := cmd.Flags().GetString("lineup-scenario")
	lineupJobScheduleShiftID, _ := cmd.Flags().GetString("lineup-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioLineupJobScheduleShiftsCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		LineupScenarioID:         lineupScenarioID,
		LineupJobScheduleShiftID: lineupJobScheduleShiftID,
	}, nil
}
