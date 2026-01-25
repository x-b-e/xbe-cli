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

type doLineupScenarioTrailerLineupJobScheduleShiftsCreateOptions struct {
	BaseURL                              string
	Token                                string
	JSON                                 bool
	LineupScenarioTrailer                string
	LineupScenarioLineupJobScheduleShift string
	StartSiteDistanceMinutes             string
	EndSiteDistanceMinutes               string
}

func newDoLineupScenarioTrailerLineupJobScheduleShiftsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lineup scenario trailer lineup job schedule shift",
		Long: `Create a lineup scenario trailer lineup job schedule shift.

Required flags:
  --lineup-scenario-trailer                 Lineup scenario trailer ID (required)
  --lineup-scenario-lineup-job-schedule-shift Lineup scenario lineup job schedule shift ID (required)

Optional attributes:
  --start-site-distance-minutes             Start site distance minutes (non-negative integer)
  --end-site-distance-minutes               End site distance minutes (non-negative integer)

Notes:
  - The lineup scenario trailer and lineup scenario lineup job schedule shift must belong to the same lineup scenario.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a record
  xbe do lineup-scenario-trailer-lineup-job-schedule-shifts create \
    --lineup-scenario-trailer 123 \
    --lineup-scenario-lineup-job-schedule-shift 456 \
    --start-site-distance-minutes 10 \
    --end-site-distance-minutes 12`,
		Args: cobra.NoArgs,
		RunE: runDoLineupScenarioTrailerLineupJobScheduleShiftsCreate,
	}
	initDoLineupScenarioTrailerLineupJobScheduleShiftsCreateFlags(cmd)
	return cmd
}

func init() {
	doLineupScenarioTrailerLineupJobScheduleShiftsCmd.AddCommand(newDoLineupScenarioTrailerLineupJobScheduleShiftsCreateCmd())
}

func initDoLineupScenarioTrailerLineupJobScheduleShiftsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("lineup-scenario-trailer", "", "Lineup scenario trailer ID (required)")
	cmd.Flags().String("lineup-scenario-lineup-job-schedule-shift", "", "Lineup scenario lineup job schedule shift ID (required)")
	cmd.Flags().String("start-site-distance-minutes", "", "Start site distance minutes (non-negative integer)")
	cmd.Flags().String("end-site-distance-minutes", "", "End site distance minutes (non-negative integer)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("lineup-scenario-trailer")
	_ = cmd.MarkFlagRequired("lineup-scenario-lineup-job-schedule-shift")
}

func runDoLineupScenarioTrailerLineupJobScheduleShiftsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoLineupScenarioTrailerLineupJobScheduleShiftsCreateOptions(cmd)
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

	lineupScenarioTrailerID := strings.TrimSpace(opts.LineupScenarioTrailer)
	if lineupScenarioTrailerID == "" {
		err := fmt.Errorf("--lineup-scenario-trailer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	lineupScenarioLineupJobScheduleShiftID := strings.TrimSpace(opts.LineupScenarioLineupJobScheduleShift)
	if lineupScenarioLineupJobScheduleShiftID == "" {
		err := fmt.Errorf("--lineup-scenario-lineup-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"lineup-scenario-trailer": map[string]any{
			"data": map[string]any{
				"type": "lineup-scenario-trailers",
				"id":   lineupScenarioTrailerID,
			},
		},
		"lineup-scenario-lineup-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "lineup-scenario-lineup-job-schedule-shifts",
				"id":   lineupScenarioLineupJobScheduleShiftID,
			},
		},
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("start-site-distance-minutes") {
		attributes["start-site-distance-minutes"] = opts.StartSiteDistanceMinutes
	}
	if cmd.Flags().Changed("end-site-distance-minutes") {
		attributes["end-site-distance-minutes"] = opts.EndSiteDistanceMinutes
	}

	requestData := map[string]any{
		"type":          "lineup-scenario-trailer-lineup-job-schedule-shifts",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		requestData["attributes"] = attributes
	}

	requestBody := map[string]any{"data": requestData}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/lineup-scenario-trailer-lineup-job-schedule-shifts", jsonBody)
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

	row := buildLineupScenarioTrailerLineupJobScheduleShiftRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created lineup scenario trailer lineup job schedule shift %s\n", row.ID)
	return nil
}

func parseDoLineupScenarioTrailerLineupJobScheduleShiftsCreateOptions(cmd *cobra.Command) (doLineupScenarioTrailerLineupJobScheduleShiftsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	lineupScenarioTrailer, _ := cmd.Flags().GetString("lineup-scenario-trailer")
	lineupScenarioLineupJobScheduleShift, _ := cmd.Flags().GetString("lineup-scenario-lineup-job-schedule-shift")
	startSiteDistanceMinutes, _ := cmd.Flags().GetString("start-site-distance-minutes")
	endSiteDistanceMinutes, _ := cmd.Flags().GetString("end-site-distance-minutes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doLineupScenarioTrailerLineupJobScheduleShiftsCreateOptions{
		BaseURL:                              baseURL,
		Token:                                token,
		JSON:                                 jsonOut,
		LineupScenarioTrailer:                lineupScenarioTrailer,
		LineupScenarioLineupJobScheduleShift: lineupScenarioLineupJobScheduleShift,
		StartSiteDistanceMinutes:             startSiteDistanceMinutes,
		EndSiteDistanceMinutes:               endSiteDistanceMinutes,
	}, nil
}

func buildLineupScenarioTrailerLineupJobScheduleShiftRowFromSingle(resp jsonAPISingleResponse) lineupScenarioTrailerLineupJobScheduleShiftRow {
	resource := resp.Data
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

	return row
}
