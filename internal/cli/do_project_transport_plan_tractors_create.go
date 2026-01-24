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

type doProjectTransportPlanTractorsCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	ProjectTransportPlanID         string
	SegmentStartID                 string
	SegmentEndID                   string
	TractorID                      string
	Status                         string
	AutomaticallyAdjustOverlapping bool
	SkipAssignmentRulesValidation  bool
	AssignmentRuleOverrideReason   string
	ActualizerWindowStartAt        string
	ActualizerWindowEndAt          string
}

func newDoProjectTransportPlanTractorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan tractor",
		Long: `Create a project transport plan tractor assignment.

Required flags:
  --project-transport-plan  Project transport plan ID
  --segment-start           Start segment ID
  --segment-end             End segment ID

Optional attributes:
  --status                                    Assignment status (editing, active)
  --automatically-adjust-overlapping-windows  Auto-adjust overlapping windows
  --actualizer-window-start-at                Actualizer window start timestamp (ISO 8601)
  --actualizer-window-end-at                  Actualizer window end timestamp (ISO 8601)
  --skip-assignment-rules-validation          Skip assignment rules validation
  --assignment-rule-override-reason           Reason for overriding assignment rules

Optional relationships:
  --tractor  Tractor ID

Notes:
  - Status cannot be active without a tractor.
  - Segment start and end must belong to the same plan.`,
		Example: `  # Create a project transport plan tractor assignment
  xbe do project-transport-plan-tractors create \
    --project-transport-plan 123 \
    --segment-start 456 \
    --segment-end 789

  # Create with tractor and status
  xbe do project-transport-plan-tractors create \
    --project-transport-plan 123 \
    --segment-start 456 \
    --segment-end 789 \
    --tractor 321 \
    --status active`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanTractorsCreate,
	}
	initDoProjectTransportPlanTractorsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanTractorsCmd.AddCommand(newDoProjectTransportPlanTractorsCreateCmd())
}

func initDoProjectTransportPlanTractorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID (required)")
	cmd.Flags().String("segment-start", "", "Start segment ID (required)")
	cmd.Flags().String("segment-end", "", "End segment ID (required)")
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("status", "", "Assignment status (editing, active)")
	cmd.Flags().Bool("automatically-adjust-overlapping-windows", false, "Auto-adjust overlapping windows")
	cmd.Flags().String("actualizer-window-start-at", "", "Actualizer window start timestamp (ISO 8601)")
	cmd.Flags().String("actualizer-window-end-at", "", "Actualizer window end timestamp (ISO 8601)")
	cmd.Flags().Bool("skip-assignment-rules-validation", false, "Skip assignment rules validation")
	cmd.Flags().String("assignment-rule-override-reason", "", "Reason for overriding assignment rules")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanTractorsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanTractorsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectTransportPlanID) == "" {
		err := fmt.Errorf("--project-transport-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.SegmentStartID) == "" {
		err := fmt.Errorf("--segment-start is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.SegmentEndID) == "" {
		err := fmt.Errorf("--segment-end is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("automatically-adjust-overlapping-windows") {
		attributes["automatically-adjust-overlapping-windows"] = opts.AutomaticallyAdjustOverlapping
	}
	if opts.ActualizerWindowStartAt != "" {
		attributes["actualizer-window-start-at"] = opts.ActualizerWindowStartAt
	}
	if opts.ActualizerWindowEndAt != "" {
		attributes["actualizer-window-end-at"] = opts.ActualizerWindowEndAt
	}
	if cmd.Flags().Changed("skip-assignment-rules-validation") {
		attributes["skip-assignment-rules-validation"] = opts.SkipAssignmentRulesValidation
	}
	if opts.AssignmentRuleOverrideReason != "" {
		attributes["assignment-rule-override-reason"] = opts.AssignmentRuleOverrideReason
	}

	relationships := map[string]any{
		"project-transport-plan": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlanID,
			},
		},
		"segment-start": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentStartID,
			},
		},
		"segment-end": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segments",
				"id":   opts.SegmentEndID,
			},
		},
	}
	if strings.TrimSpace(opts.TractorID) != "" {
		relationships["tractor"] = map[string]any{
			"data": map[string]any{
				"type": "tractors",
				"id":   opts.TractorID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-tractors",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-tractors", jsonBody)
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

	details := buildProjectTransportPlanTractorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan tractor %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanTractorsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanTractorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlanID, _ := cmd.Flags().GetString("project-transport-plan")
	segmentStartID, _ := cmd.Flags().GetString("segment-start")
	segmentEndID, _ := cmd.Flags().GetString("segment-end")
	tractorID, _ := cmd.Flags().GetString("tractor")
	status, _ := cmd.Flags().GetString("status")
	automaticallyAdjustOverlapping, _ := cmd.Flags().GetBool("automatically-adjust-overlapping-windows")
	skipAssignmentRulesValidation, _ := cmd.Flags().GetBool("skip-assignment-rules-validation")
	assignmentRuleOverrideReason, _ := cmd.Flags().GetString("assignment-rule-override-reason")
	actualizerWindowStartAt, _ := cmd.Flags().GetString("actualizer-window-start-at")
	actualizerWindowEndAt, _ := cmd.Flags().GetString("actualizer-window-end-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanTractorsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		ProjectTransportPlanID:         projectTransportPlanID,
		SegmentStartID:                 segmentStartID,
		SegmentEndID:                   segmentEndID,
		TractorID:                      tractorID,
		Status:                         status,
		AutomaticallyAdjustOverlapping: automaticallyAdjustOverlapping,
		SkipAssignmentRulesValidation:  skipAssignmentRulesValidation,
		AssignmentRuleOverrideReason:   assignmentRuleOverrideReason,
		ActualizerWindowStartAt:        actualizerWindowStartAt,
		ActualizerWindowEndAt:          actualizerWindowEndAt,
	}, nil
}
