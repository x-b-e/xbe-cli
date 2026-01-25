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

type doJobScheduleShiftIsManagedTogglesCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	JobScheduleShiftID string
}

type jobScheduleShiftIsManagedToggleRow struct {
	ID                 string `json:"id"`
	JobScheduleShiftID string `json:"job_schedule_shift_id,omitempty"`
}

func newDoJobScheduleShiftIsManagedTogglesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Toggle a job schedule shift managed status",
		Long: `Toggle a job schedule shift managed status.

Required:
  --job-schedule-shift  Job schedule shift ID

Notes:
  The shift must not be cancelled, tied to an unmanaged tender, or have a time
  card.`,
		Example: `  # Toggle managed status for a job schedule shift
  xbe do job-schedule-shift-is-managed-toggles create --job-schedule-shift 123`,
		Args: cobra.NoArgs,
		RunE: runDoJobScheduleShiftIsManagedTogglesCreate,
	}
	initDoJobScheduleShiftIsManagedTogglesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobScheduleShiftIsManagedTogglesCmd.AddCommand(newDoJobScheduleShiftIsManagedTogglesCreateCmd())
}

func initDoJobScheduleShiftIsManagedTogglesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-schedule-shift", "", "Job schedule shift ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobScheduleShiftIsManagedTogglesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobScheduleShiftIsManagedTogglesCreateOptions(cmd)
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

	if opts.JobScheduleShiftID == "" {
		err := fmt.Errorf("--job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "job-schedule-shifts",
				"id":   opts.JobScheduleShiftID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-schedule-shift-is-managed-toggles",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-schedule-shift-is-managed-toggles", jsonBody)
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

	row := buildJobScheduleShiftIsManagedToggleRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job schedule shift managed toggle %s\n", row.ID)
	return nil
}

func parseDoJobScheduleShiftIsManagedTogglesCreateOptions(cmd *cobra.Command) (doJobScheduleShiftIsManagedTogglesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobScheduleShiftID, _ := cmd.Flags().GetString("job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobScheduleShiftIsManagedTogglesCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		JobScheduleShiftID: jobScheduleShiftID,
	}, nil
}

func buildJobScheduleShiftIsManagedToggleRowFromSingle(resp jsonAPISingleResponse) jobScheduleShiftIsManagedToggleRow {
	resource := resp.Data
	row := jobScheduleShiftIsManagedToggleRow{
		ID: resource.ID,
	}
	if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		row.JobScheduleShiftID = rel.Data.ID
	}
	return row
}
