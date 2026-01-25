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

type projectTransportPlanTractorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanTractorDetails struct {
	ID                                    string   `json:"id"`
	ProjectTransportPlanID                string   `json:"project_transport_plan_id,omitempty"`
	TractorID                             string   `json:"tractor_id,omitempty"`
	SegmentStartID                        string   `json:"segment_start_id,omitempty"`
	SegmentEndID                          string   `json:"segment_end_id,omitempty"`
	Status                                string   `json:"status,omitempty"`
	AutomaticallyAdjustOverlapping        bool     `json:"automatically_adjust_overlapping_windows"`
	SkipAssignmentRulesValidation         bool     `json:"skip_assignment_rules_validation"`
	AssignmentRuleOverrideReason          string   `json:"assignment_rule_override_reason,omitempty"`
	WindowStartAtCached                   string   `json:"window_start_at_cached,omitempty"`
	WindowEndAtCached                     string   `json:"window_end_at_cached,omitempty"`
	PlannedWindowStartAtCached            string   `json:"planned_window_start_at_cached,omitempty"`
	ExpectedWindowStartAtCached           string   `json:"expected_window_start_at_cached,omitempty"`
	ActualizerWindowStartAt               string   `json:"actualizer_window_start_at,omitempty"`
	ActualizerWindowEndAt                 string   `json:"actualizer_window_end_at,omitempty"`
	ProjectTransportPlanSegmentTractorIDs []string `json:"project_transport_plan_segment_tractor_ids,omitempty"`
	SegmentIDs                            []string `json:"segment_ids,omitempty"`
}

func newProjectTransportPlanTractorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan tractor details",
		Long: `Show the full details of a project transport plan tractor.

Output Fields:
  ID
  Project Transport Plan ID
  Tractor ID
  Segment Start ID
  Segment End ID
  Status
  Automatically Adjust Overlapping Windows
  Skip Assignment Rules Validation
  Assignment Rule Override Reason
  Window Start At Cached
  Window End At Cached
  Planned Window Start At Cached
  Expected Window Start At Cached
  Actualizer Window Start At
  Actualizer Window End At
  Project Transport Plan Segment Tractor IDs
  Segment IDs

Arguments:
  <id>    The project transport plan tractor ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan tractor
  xbe view project-transport-plan-tractors show 123

  # JSON output
  xbe view project-transport-plan-tractors show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanTractorsShow,
	}
	initProjectTransportPlanTractorsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanTractorsCmd.AddCommand(newProjectTransportPlanTractorsShowCmd())
}

func initProjectTransportPlanTractorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanTractorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanTractorsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan tractor id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-tractors]", "status,automatically-adjust-overlapping-windows,window-start-at-cached,window-end-at-cached,planned-window-start-at-cached,expected-window-start-at-cached,actualizer-window-start-at,actualizer-window-end-at,skip-assignment-rules-validation,assignment-rule-override-reason,project-transport-plan,segment-start,segment-end,tractor,project-transport-plan-segment-tractors,segments")
	query.Set("include", "project-transport-plan,segment-start,segment-end,tractor,project-transport-plan-segment-tractors,segments")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-tractors/"+id, query)
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

	details := buildProjectTransportPlanTractorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanTractorDetails(cmd, details)
}

func parseProjectTransportPlanTractorsShowOptions(cmd *cobra.Command) (projectTransportPlanTractorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanTractorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanTractorDetails(resp jsonAPISingleResponse) projectTransportPlanTractorDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := projectTransportPlanTractorDetails{
		ID:                             resource.ID,
		ProjectTransportPlanID:         relationshipIDFromMap(resource.Relationships, "project-transport-plan"),
		TractorID:                      relationshipIDFromMap(resource.Relationships, "tractor"),
		SegmentStartID:                 relationshipIDFromMap(resource.Relationships, "segment-start"),
		SegmentEndID:                   relationshipIDFromMap(resource.Relationships, "segment-end"),
		Status:                         stringAttr(attrs, "status"),
		AutomaticallyAdjustOverlapping: boolAttr(attrs, "automatically-adjust-overlapping-windows"),
		SkipAssignmentRulesValidation:  boolAttr(attrs, "skip-assignment-rules-validation"),
		AssignmentRuleOverrideReason:   stringAttr(attrs, "assignment-rule-override-reason"),
		WindowStartAtCached:            formatDateTime(stringAttr(attrs, "window-start-at-cached")),
		WindowEndAtCached:              formatDateTime(stringAttr(attrs, "window-end-at-cached")),
		PlannedWindowStartAtCached:     formatDateTime(stringAttr(attrs, "planned-window-start-at-cached")),
		ExpectedWindowStartAtCached:    formatDateTime(stringAttr(attrs, "expected-window-start-at-cached")),
		ActualizerWindowStartAt:        formatDateTime(stringAttr(attrs, "actualizer-window-start-at")),
		ActualizerWindowEndAt:          formatDateTime(stringAttr(attrs, "actualizer-window-end-at")),
	}

	if rel, ok := resource.Relationships["project-transport-plan-segment-tractors"]; ok {
		details.ProjectTransportPlanSegmentTractorIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["segments"]; ok {
		details.SegmentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectTransportPlanTractorDetails(cmd *cobra.Command, details projectTransportPlanTractorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.TractorID != "" {
		fmt.Fprintf(out, "Tractor ID: %s\n", details.TractorID)
	}
	if details.SegmentStartID != "" {
		fmt.Fprintf(out, "Segment Start ID: %s\n", details.SegmentStartID)
	}
	if details.SegmentEndID != "" {
		fmt.Fprintf(out, "Segment End ID: %s\n", details.SegmentEndID)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	fmt.Fprintf(out, "Automatically Adjust Overlapping Windows: %t\n", details.AutomaticallyAdjustOverlapping)
	fmt.Fprintf(out, "Skip Assignment Rules Validation: %t\n", details.SkipAssignmentRulesValidation)
	if details.AssignmentRuleOverrideReason != "" {
		fmt.Fprintf(out, "Assignment Rule Override Reason: %s\n", details.AssignmentRuleOverrideReason)
	}
	if details.WindowStartAtCached != "" {
		fmt.Fprintf(out, "Window Start At Cached: %s\n", details.WindowStartAtCached)
	}
	if details.WindowEndAtCached != "" {
		fmt.Fprintf(out, "Window End At Cached: %s\n", details.WindowEndAtCached)
	}
	if details.PlannedWindowStartAtCached != "" {
		fmt.Fprintf(out, "Planned Window Start At Cached: %s\n", details.PlannedWindowStartAtCached)
	}
	if details.ExpectedWindowStartAtCached != "" {
		fmt.Fprintf(out, "Expected Window Start At Cached: %s\n", details.ExpectedWindowStartAtCached)
	}
	if details.ActualizerWindowStartAt != "" {
		fmt.Fprintf(out, "Actualizer Window Start At: %s\n", details.ActualizerWindowStartAt)
	}
	if details.ActualizerWindowEndAt != "" {
		fmt.Fprintf(out, "Actualizer Window End At: %s\n", details.ActualizerWindowEndAt)
	}
	if len(details.ProjectTransportPlanSegmentTractorIDs) > 0 {
		fmt.Fprintf(out, "Project Transport Plan Segment Tractor IDs: %s\n", strings.Join(details.ProjectTransportPlanSegmentTractorIDs, ", "))
	}
	if len(details.SegmentIDs) > 0 {
		fmt.Fprintf(out, "Segment IDs: %s\n", strings.Join(details.SegmentIDs, ", "))
	}

	return nil
}
