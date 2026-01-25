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

type doJobProductionPlanScheduleChangesCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	JobProductionPlan          string
	Work                       string
	OffsetSeconds              int
	TimeKind                   string
	SkipUpdateShifts           bool
	SkipUpdateCrewRequirements bool
	SkipUpdateSafetyMeeting    bool
	SkipPersistence            bool
}

type jobProductionPlanScheduleChangeRow struct {
	ID                                  string   `json:"id"`
	JobProductionPlanID                 string   `json:"job_production_plan_id,omitempty"`
	WorkID                              string   `json:"work_id,omitempty"`
	OffsetSeconds                       int      `json:"offset_seconds"`
	TimeKind                            string   `json:"time_kind,omitempty"`
	SkipUpdateShifts                    bool     `json:"skip_update_shifts"`
	SkipUpdateCrewRequirements          bool     `json:"skip_update_crew_requirements"`
	SkipUpdateSafetyMeeting             bool     `json:"skip_update_safety_meeting"`
	SkipPersistence                     bool     `json:"skip_persistence"`
	NullifiedResourceCrewRequirementIDs []string `json:"nullified_resource_crew_requirement_ids,omitempty"`
	ScheduleChanges                     []any    `json:"schedule_changes,omitempty"`
}

func newDoJobProductionPlanScheduleChangesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Apply a job production plan schedule change",
		Long: `Apply a schedule change to a job production plan.

Schedule changes shift plan times by an offset in seconds. The job production
plan cannot be schedule locked, have time cards, time sheets, or equipment
movement trips.

Required flags:
  --job-production-plan   Job production plan ID
  --offset-seconds        Offset in seconds (positive or negative)

Optional flags:
  --time-kind                     Which times to shift (both, material_site, job_site)
  --skip-update-shifts            Skip updating shifts
  --skip-update-crew-requirements Skip updating crew requirements
  --skip-update-safety-meeting    Skip updating safety meeting
  --skip-persistence              Run without persisting the work record
  --work                          Existing schedule change work ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Push a job production plan schedule by 30 minutes
  xbe do job-production-plan-schedule-changes create \
    --job-production-plan 123 \
    --offset-seconds 1800

  # Move job site times earlier and skip shifts updates
  xbe do job-production-plan-schedule-changes create \
    --job-production-plan 123 \
    --offset-seconds -900 \
    --time-kind job_site \
    --skip-update-shifts`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanScheduleChangesCreate,
	}
	initDoJobProductionPlanScheduleChangesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanScheduleChangesCmd.AddCommand(newDoJobProductionPlanScheduleChangesCreateCmd())
}

func initDoJobProductionPlanScheduleChangesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().Int("offset-seconds", 0, "Offset in seconds (positive or negative)")
	cmd.Flags().String("time-kind", "", "Which times to shift (both, material_site, job_site)")
	cmd.Flags().Bool("skip-update-shifts", false, "Skip updating shifts")
	cmd.Flags().Bool("skip-update-crew-requirements", false, "Skip updating crew requirements")
	cmd.Flags().Bool("skip-update-safety-meeting", false, "Skip updating safety meeting")
	cmd.Flags().Bool("skip-persistence", false, "Run without persisting the work record")
	cmd.Flags().String("work", "", "Existing schedule change work ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanScheduleChangesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanScheduleChangesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.JobProductionPlan) == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if !cmd.Flags().Changed("offset-seconds") {
		err := fmt.Errorf("--offset-seconds is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if cmd.Flags().Changed("time-kind") {
		if !isJobProductionPlanScheduleChangeTimeKindValid(opts.TimeKind) {
			err := fmt.Errorf("--time-kind must be one of: both, material_site, job_site")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("offset-seconds") {
		attributes["offset-seconds"] = opts.OffsetSeconds
	}
	if cmd.Flags().Changed("time-kind") {
		attributes["time-kind"] = opts.TimeKind
	}
	if cmd.Flags().Changed("skip-update-shifts") {
		attributes["skip-update-shifts"] = opts.SkipUpdateShifts
	}
	if cmd.Flags().Changed("skip-update-crew-requirements") {
		attributes["skip-update-crew-requirements"] = opts.SkipUpdateCrewRequirements
	}
	if cmd.Flags().Changed("skip-update-safety-meeting") {
		attributes["skip-update-safety-meeting"] = opts.SkipUpdateSafetyMeeting
	}
	if cmd.Flags().Changed("skip-persistence") {
		attributes["skip-persistence"] = opts.SkipPersistence
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
	}
	if strings.TrimSpace(opts.Work) != "" {
		relationships["work"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-schedule-change-works",
				"id":   opts.Work,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-schedule-changes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-schedule-changes", jsonBody)
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

	row := buildJobProductionPlanScheduleChangeRow(resp.Data)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.ID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan schedule change %s\n", row.ID)
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan schedule change for job production plan %s\n", opts.JobProductionPlan)
	return nil
}

func parseDoJobProductionPlanScheduleChangesCreateOptions(cmd *cobra.Command) (doJobProductionPlanScheduleChangesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	work, _ := cmd.Flags().GetString("work")
	offsetSeconds, _ := cmd.Flags().GetInt("offset-seconds")
	timeKind, _ := cmd.Flags().GetString("time-kind")
	skipUpdateShifts, _ := cmd.Flags().GetBool("skip-update-shifts")
	skipUpdateCrewRequirements, _ := cmd.Flags().GetBool("skip-update-crew-requirements")
	skipUpdateSafetyMeeting, _ := cmd.Flags().GetBool("skip-update-safety-meeting")
	skipPersistence, _ := cmd.Flags().GetBool("skip-persistence")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanScheduleChangesCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		JobProductionPlan:          jobProductionPlan,
		Work:                       work,
		OffsetSeconds:              offsetSeconds,
		TimeKind:                   timeKind,
		SkipUpdateShifts:           skipUpdateShifts,
		SkipUpdateCrewRequirements: skipUpdateCrewRequirements,
		SkipUpdateSafetyMeeting:    skipUpdateSafetyMeeting,
		SkipPersistence:            skipPersistence,
	}, nil
}

func buildJobProductionPlanScheduleChangeRow(resource jsonAPIResource) jobProductionPlanScheduleChangeRow {
	attrs := resource.Attributes
	row := jobProductionPlanScheduleChangeRow{
		ID:                                  resource.ID,
		OffsetSeconds:                       intAttr(attrs, "offset-seconds"),
		TimeKind:                            stringAttr(attrs, "time-kind"),
		SkipUpdateShifts:                    boolAttr(attrs, "skip-update-shifts"),
		SkipUpdateCrewRequirements:          boolAttr(attrs, "skip-update-crew-requirements"),
		SkipUpdateSafetyMeeting:             boolAttr(attrs, "skip-update-safety-meeting"),
		SkipPersistence:                     boolAttr(attrs, "skip-persistence"),
		NullifiedResourceCrewRequirementIDs: stringSliceAttr(attrs, "nullified-resource-crew-requirement-ids"),
		ScheduleChanges:                     anySliceAttr(attrs, "schedule-changes"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["work"]; ok && rel.Data != nil {
		row.WorkID = rel.Data.ID
	}

	return row
}

func anySliceAttr(attrs map[string]any, key string) []any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		return typed
	default:
		return []any{typed}
	}
}

func isJobProductionPlanScheduleChangeTimeKindValid(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	switch value {
	case "both", "material_site", "job_site":
		return true
	default:
		return false
	}
}
