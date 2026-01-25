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

type jobProductionPlanScheduleChangeWorksShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanScheduleChangeWorkDetails struct {
	ID                         string `json:"id"`
	JobProductionPlanID        string `json:"job_production_plan_id,omitempty"`
	CreatedByID                string `json:"created_by_id,omitempty"`
	OffsetSeconds              string `json:"offset_seconds,omitempty"`
	TimeKind                   string `json:"time_kind,omitempty"`
	SkipUpdateShifts           bool   `json:"skip_update_shifts"`
	SkipUpdateCrewRequirements bool   `json:"skip_update_crew_requirements"`
	SkipUpdateSafetyMeeting    bool   `json:"skip_update_safety_meeting"`
	ScheduledAt                string `json:"scheduled_at,omitempty"`
	ProcessedAt                string `json:"processed_at,omitempty"`
	WorkResults                any    `json:"work_results,omitempty"`
	WorkErrors                 any    `json:"work_errors,omitempty"`
	CreatedAt                  string `json:"created_at,omitempty"`
	UpdatedAt                  string `json:"updated_at,omitempty"`
}

func newJobProductionPlanScheduleChangeWorksShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan schedule change work details",
		Long: `Show the full details of a job production plan schedule change work item.

Output Fields:
  ID
  Job Production Plan ID
  Created By (user ID)
  Offset Seconds
  Time Kind
  Skip Update Shifts
  Skip Update Crew Requirements
  Skip Update Safety Meeting
  Scheduled At
  Processed At
  Work Results
  Work Errors
  Created At
  Updated At

Arguments:
  <id>    The schedule change work ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show schedule change work details
  xbe view job-production-plan-schedule-change-works show 123

  # Get JSON output
  xbe view job-production-plan-schedule-change-works show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanScheduleChangeWorksShow,
	}
	initJobProductionPlanScheduleChangeWorksShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanScheduleChangeWorksCmd.AddCommand(newJobProductionPlanScheduleChangeWorksShowCmd())
}

func initJobProductionPlanScheduleChangeWorksShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanScheduleChangeWorksShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanScheduleChangeWorksShowOptions(cmd)
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
		return fmt.Errorf("job production plan schedule change work id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-schedule-change-works/"+id, nil)
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

	details := buildJobProductionPlanScheduleChangeWorkDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanScheduleChangeWorkDetails(cmd, details)
}

func parseJobProductionPlanScheduleChangeWorksShowOptions(cmd *cobra.Command) (jobProductionPlanScheduleChangeWorksShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanScheduleChangeWorksShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanScheduleChangeWorkDetails(resp jsonAPISingleResponse) jobProductionPlanScheduleChangeWorkDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanScheduleChangeWorkDetails{
		ID:                         resource.ID,
		OffsetSeconds:              stringAttr(attrs, "offset-seconds"),
		TimeKind:                   stringAttr(attrs, "time-kind"),
		SkipUpdateShifts:           boolAttr(attrs, "skip-update-shifts"),
		SkipUpdateCrewRequirements: boolAttr(attrs, "skip-update-crew-requirements"),
		SkipUpdateSafetyMeeting:    boolAttr(attrs, "skip-update-safety-meeting"),
		ScheduledAt:                formatDateTime(stringAttr(attrs, "scheduled-at")),
		ProcessedAt:                formatDateTime(stringAttr(attrs, "processed-at")),
		WorkResults:                anyAttr(attrs, "work-results"),
		WorkErrors:                 anyAttr(attrs, "work-errors"),
		CreatedAt:                  formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                  formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}

	return details
}

func renderJobProductionPlanScheduleChangeWorkDetails(cmd *cobra.Command, details jobProductionPlanScheduleChangeWorkDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.OffsetSeconds != "" {
		fmt.Fprintf(out, "Offset Seconds: %s\n", details.OffsetSeconds)
	}
	if details.TimeKind != "" {
		fmt.Fprintf(out, "Time Kind: %s\n", details.TimeKind)
	}
	fmt.Fprintf(out, "Skip Update Shifts: %s\n", formatYesNo(details.SkipUpdateShifts))
	fmt.Fprintf(out, "Skip Update Crew Requirements: %s\n", formatYesNo(details.SkipUpdateCrewRequirements))
	fmt.Fprintf(out, "Skip Update Safety Meeting: %s\n", formatYesNo(details.SkipUpdateSafetyMeeting))
	if details.ScheduledAt != "" {
		fmt.Fprintf(out, "Scheduled At: %s\n", details.ScheduledAt)
	}
	if details.ProcessedAt != "" {
		fmt.Fprintf(out, "Processed At: %s\n", details.ProcessedAt)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.WorkResults != nil {
		fmt.Fprintln(out, "Work Results:")
		if err := writeJSON(out, details.WorkResults); err != nil {
			return err
		}
	}
	if details.WorkErrors != nil {
		fmt.Fprintln(out, "Work Errors:")
		if err := writeJSON(out, details.WorkErrors); err != nil {
			return err
		}
	}

	return nil
}
