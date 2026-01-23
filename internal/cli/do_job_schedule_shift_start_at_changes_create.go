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

type doJobScheduleShiftStartAtChangesCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	JobScheduleShift string
	NewStartAt       string
}

type jobScheduleShiftStartAtChangeCreateRow struct {
	ID                 string `json:"id"`
	JobScheduleShiftID string `json:"job_schedule_shift_id,omitempty"`
	OldStartAt         string `json:"old_start_at,omitempty"`
	NewStartAt         string `json:"new_start_at,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
}

func newDoJobScheduleShiftStartAtChangesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job schedule shift start-at change",
		Long: `Create a job schedule shift start-at change.

Start-at changes reschedule a job schedule shift by moving its start time.

Required flags:
  --job-schedule-shift  Job schedule shift ID (required)
  --new-start-at        New shift start time (ISO 8601, required)

Notes:
  - The new start time must differ from the current start time.
  - Flexible shifts must keep the new start time within allowed bounds.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Move a shift start time
  xbe do job-schedule-shift-start-at-changes create \\
    --job-schedule-shift 123 \\
    --new-start-at 2026-01-23T14:30:00Z`,
		Args: cobra.NoArgs,
		RunE: runDoJobScheduleShiftStartAtChangesCreate,
	}
	initDoJobScheduleShiftStartAtChangesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobScheduleShiftStartAtChangesCmd.AddCommand(newDoJobScheduleShiftStartAtChangesCreateCmd())
}

func initDoJobScheduleShiftStartAtChangesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-schedule-shift", "", "Job schedule shift ID (required)")
	cmd.Flags().String("new-start-at", "", "New shift start time (ISO 8601, required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobScheduleShiftStartAtChangesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobScheduleShiftStartAtChangesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	jobScheduleShiftID := strings.TrimSpace(opts.JobScheduleShift)
	newStartAt := strings.TrimSpace(opts.NewStartAt)

	if jobScheduleShiftID == "" {
		err := fmt.Errorf("--job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if newStartAt == "" {
		err := fmt.Errorf("--new-start-at is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "job-schedule-shifts",
				"id":   jobScheduleShiftID,
			},
		},
	}

	attributes := map[string]any{
		"new-start-at": newStartAt,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-schedule-shift-start-at-changes",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-schedule-shift-start-at-changes", jsonBody)
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

	row := buildJobScheduleShiftStartAtChangeCreateRow(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job schedule shift start-at change %s\n", row.ID)
	return nil
}

func parseDoJobScheduleShiftStartAtChangesCreateOptions(cmd *cobra.Command) (doJobScheduleShiftStartAtChangesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	newStartAt, _ := cmd.Flags().GetString("new-start-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobScheduleShiftStartAtChangesCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		JobScheduleShift: jobScheduleShift,
		NewStartAt:       newStartAt,
	}, nil
}

func buildJobScheduleShiftStartAtChangeCreateRow(resp jsonAPISingleResponse) jobScheduleShiftStartAtChangeCreateRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := jobScheduleShiftStartAtChangeCreateRow{
		ID:         resource.ID,
		OldStartAt: formatDateTime(stringAttr(attrs, "old-start-at")),
		NewStartAt: formatDateTime(stringAttr(attrs, "new-start-at")),
		CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		row.JobScheduleShiftID = rel.Data.ID
	}

	return row
}
