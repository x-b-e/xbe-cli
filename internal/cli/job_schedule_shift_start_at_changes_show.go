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

type jobScheduleShiftStartAtChangesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobScheduleShiftStartAtChangeDetails struct {
	ID                 string `json:"id"`
	JobScheduleShiftID string `json:"job_schedule_shift_id,omitempty"`
	OldStartAt         string `json:"old_start_at,omitempty"`
	NewStartAt         string `json:"new_start_at,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

func newJobScheduleShiftStartAtChangesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job schedule shift start-at change details",
		Long: `Show the full details of a job schedule shift start-at change.

Output Fields:
  ID
  Job Schedule Shift ID
  Old Start At
  New Start At
  Created At
  Updated At

Arguments:
  <id>    The start-at change ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show start-at change details
  xbe view job-schedule-shift-start-at-changes show 123

  # Get JSON output
  xbe view job-schedule-shift-start-at-changes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobScheduleShiftStartAtChangesShow,
	}
	initJobScheduleShiftStartAtChangesShowFlags(cmd)
	return cmd
}

func init() {
	jobScheduleShiftStartAtChangesCmd.AddCommand(newJobScheduleShiftStartAtChangesShowCmd())
}

func initJobScheduleShiftStartAtChangesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobScheduleShiftStartAtChangesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobScheduleShiftStartAtChangesShowOptions(cmd)
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
		return fmt.Errorf("job schedule shift start-at change id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-schedule-shift-start-at-changes/"+id, nil)
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

	details := buildJobScheduleShiftStartAtChangeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobScheduleShiftStartAtChangeDetails(cmd, details)
}

func parseJobScheduleShiftStartAtChangesShowOptions(cmd *cobra.Command) (jobScheduleShiftStartAtChangesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobScheduleShiftStartAtChangesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobScheduleShiftStartAtChangeDetails(resp jsonAPISingleResponse) jobScheduleShiftStartAtChangeDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobScheduleShiftStartAtChangeDetails{
		ID:         resource.ID,
		OldStartAt: formatDateTime(stringAttr(attrs, "old-start-at")),
		NewStartAt: formatDateTime(stringAttr(attrs, "new-start-at")),
		CreatedAt:  formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:  formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		details.JobScheduleShiftID = rel.Data.ID
	}

	return details
}

func renderJobScheduleShiftStartAtChangeDetails(cmd *cobra.Command, details jobScheduleShiftStartAtChangeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobScheduleShiftID != "" {
		fmt.Fprintf(out, "Job Schedule Shift ID: %s\n", details.JobScheduleShiftID)
	}
	if details.OldStartAt != "" {
		fmt.Fprintf(out, "Old Start At: %s\n", details.OldStartAt)
	}
	if details.NewStartAt != "" {
		fmt.Fprintf(out, "New Start At: %s\n", details.NewStartAt)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
