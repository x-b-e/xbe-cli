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

type doProjectTransportPlanTractorsUpdateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	ID                             string
	TractorID                      string
	SegmentStartID                 string
	SegmentEndID                   string
	Status                         string
	AutomaticallyAdjustOverlapping bool
	SkipAssignmentRulesValidation  bool
	AssignmentRuleOverrideReason   string
	ActualizerWindowStartAt        string
	ActualizerWindowEndAt          string
}

func newDoProjectTransportPlanTractorsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan tractor",
		Long: `Update a project transport plan tractor assignment.

Provide the assignment ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Optional attributes:
  --status                                    Assignment status (editing, active)
  --automatically-adjust-overlapping-windows  Auto-adjust overlapping windows
  --actualizer-window-start-at                Actualizer window start timestamp (ISO 8601)
  --actualizer-window-end-at                  Actualizer window end timestamp (ISO 8601)
  --skip-assignment-rules-validation          Skip assignment rules validation
  --assignment-rule-override-reason           Reason for overriding assignment rules

Optional relationships:
  --segment-start  Start segment ID
  --segment-end    End segment ID
  --tractor        Tractor ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update assignment status
  xbe do project-transport-plan-tractors update 123 --status active

  # Update tractor assignment
  xbe do project-transport-plan-tractors update 123 --tractor 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanTractorsUpdate,
	}
	initDoProjectTransportPlanTractorsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanTractorsCmd.AddCommand(newDoProjectTransportPlanTractorsUpdateCmd())
}

func initDoProjectTransportPlanTractorsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("segment-start", "", "Start segment ID")
	cmd.Flags().String("segment-end", "", "End segment ID")
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

func runDoProjectTransportPlanTractorsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanTractorsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("automatically-adjust-overlapping-windows") {
		attributes["automatically-adjust-overlapping-windows"] = opts.AutomaticallyAdjustOverlapping
	}
	if cmd.Flags().Changed("actualizer-window-start-at") {
		attributes["actualizer-window-start-at"] = opts.ActualizerWindowStartAt
	}
	if cmd.Flags().Changed("actualizer-window-end-at") {
		attributes["actualizer-window-end-at"] = opts.ActualizerWindowEndAt
	}
	if cmd.Flags().Changed("skip-assignment-rules-validation") {
		attributes["skip-assignment-rules-validation"] = opts.SkipAssignmentRulesValidation
	}
	if cmd.Flags().Changed("assignment-rule-override-reason") {
		attributes["assignment-rule-override-reason"] = opts.AssignmentRuleOverrideReason
	}

	if cmd.Flags().Changed("segment-start") {
		if opts.SegmentStartID == "" {
			relationships["segment-start"] = map[string]any{"data": nil}
		} else {
			relationships["segment-start"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-plan-segments",
					"id":   opts.SegmentStartID,
				},
			}
		}
	}
	if cmd.Flags().Changed("segment-end") {
		if opts.SegmentEndID == "" {
			relationships["segment-end"] = map[string]any{"data": nil}
		} else {
			relationships["segment-end"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-plan-segments",
					"id":   opts.SegmentEndID,
				},
			}
		}
	}
	if cmd.Flags().Changed("tractor") {
		if opts.TractorID == "" {
			relationships["tractor"] = map[string]any{"data": nil}
		} else {
			relationships["tractor"] = map[string]any{
				"data": map[string]any{
					"type": "tractors",
					"id":   opts.TractorID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-transport-plan-tractors",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-tractors/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan tractor %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanTractorsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanTractorsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
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

	return doProjectTransportPlanTractorsUpdateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		ID:                             args[0],
		TractorID:                      tractorID,
		SegmentStartID:                 segmentStartID,
		SegmentEndID:                   segmentEndID,
		Status:                         status,
		AutomaticallyAdjustOverlapping: automaticallyAdjustOverlapping,
		SkipAssignmentRulesValidation:  skipAssignmentRulesValidation,
		AssignmentRuleOverrideReason:   assignmentRuleOverrideReason,
		ActualizerWindowStartAt:        actualizerWindowStartAt,
		ActualizerWindowEndAt:          actualizerWindowEndAt,
	}, nil
}
