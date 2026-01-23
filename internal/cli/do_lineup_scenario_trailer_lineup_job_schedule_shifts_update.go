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

type doLineupScenarioTrailerLineupJobScheduleShiftsUpdateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ID                       string
	StartSiteDistanceMinutes string
	EndSiteDistanceMinutes   string
}

func newDoLineupScenarioTrailerLineupJobScheduleShiftsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a lineup scenario trailer lineup job schedule shift",
		Long: `Update a lineup scenario trailer lineup job schedule shift.

Optional attributes:
  --start-site-distance-minutes  Start site distance minutes (non-negative integer)
  --end-site-distance-minutes    End site distance minutes (non-negative integer)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update start site distance minutes
  xbe do lineup-scenario-trailer-lineup-job-schedule-shifts update 123 \
    --start-site-distance-minutes 15`,
		Args: cobra.ExactArgs(1),
		RunE: runDoLineupScenarioTrailerLineupJobScheduleShiftsUpdate,
	}
	initDoLineupScenarioTrailerLineupJobScheduleShiftsUpdateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioTrailerLineupJobScheduleShiftsCmd.AddCommand(newDoLineupScenarioTrailerLineupJobScheduleShiftsUpdateCmd())
}

func initDoLineupScenarioTrailerLineupJobScheduleShiftsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-site-distance-minutes", "", "Start site distance minutes (non-negative integer)")
	cmd.Flags().String("end-site-distance-minutes", "", "End site distance minutes (non-negative integer)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLineupScenarioTrailerLineupJobScheduleShiftsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLineupScenarioTrailerLineupJobScheduleShiftsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("start-site-distance-minutes") {
		attributes["start-site-distance-minutes"] = opts.StartSiteDistanceMinutes
	}
	if cmd.Flags().Changed("end-site-distance-minutes") {
		attributes["end-site-distance-minutes"] = opts.EndSiteDistanceMinutes
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "lineup-scenario-trailer-lineup-job-schedule-shifts",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/lineup-scenario-trailer-lineup-job-schedule-shifts/"+opts.ID, jsonBody)
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
		row := buildLineupScenarioTrailerLineupJobScheduleShiftRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated lineup scenario trailer lineup job schedule shift %s\n", resp.Data.ID)
	return nil
}

func parseDoLineupScenarioTrailerLineupJobScheduleShiftsUpdateOptions(cmd *cobra.Command, args []string) (doLineupScenarioTrailerLineupJobScheduleShiftsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	startSiteDistanceMinutes, _ := cmd.Flags().GetString("start-site-distance-minutes")
	endSiteDistanceMinutes, _ := cmd.Flags().GetString("end-site-distance-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioTrailerLineupJobScheduleShiftsUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ID:                       args[0],
		StartSiteDistanceMinutes: startSiteDistanceMinutes,
		EndSiteDistanceMinutes:   endSiteDistanceMinutes,
	}, nil
}
