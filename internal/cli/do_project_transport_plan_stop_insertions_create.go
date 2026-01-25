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

type doProjectTransportPlanStopInsertionsCreateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	Mode                              string
	ReuseHalf                         string
	BoundaryChoice                    string
	PlannedEventTimeStartAt           string
	PlannedEventTimeEndAt             string
	PreserveStopOnDelete              bool
	ReferenceProjectTransportPlanStop string
	ProjectTransportPlan              string
	ProjectTransportLocation          string
	ProjectTransportPlanSegmentSet    string
	PlannedCompletionEventType        string
	ExistingProjectTransportPlanStop  string
	StopToMove                        string
}

func newDoProjectTransportPlanStopInsertionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan stop insertion",
		Long: `Create a project transport plan stop insertion.

Stop insertions insert, move, replace, or delete stops in a project transport
plan. The required relationships depend on the mode.

Required flags:
  --mode        Insertion mode (replace, insert_before, insert_after,
                insert_at_head, insert_at_tail, seed_plan, move_before,
                move_after, delete)
  --reuse-half  Which half to reuse when splitting (upstream, downstream)

Optional flags:
  --boundary-choice                       Boundary choice (join_upstream, join_downstream)
  --planned-event-time-start-at           Planned event time start (RFC3339)
  --planned-event-time-end-at             Planned event time end (RFC3339)
  --preserve-stop-on-delete               Preserve stop record in delete mode
  --reference-project-transport-plan-stop Reference stop ID
  --project-transport-plan                Project transport plan ID
  --project-transport-location            Project transport location ID
  --project-transport-plan-segment-set    Project transport plan segment set ID
  --planned-completion-event-type         Planned completion event type ID
  --existing-project-transport-plan-stop  Existing stop to reuse
  --stop-to-move                          Stop to move (move modes)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Insert a new stop before a reference stop
  xbe do project-transport-plan-stop-insertions create \
    --mode insert_before \
    --reuse-half upstream \
    --reference-project-transport-plan-stop 123 \
    --project-transport-location 456

  # Delete a stop while preserving the record
  xbe do project-transport-plan-stop-insertions create \
    --mode delete \
    --reuse-half downstream \
    --reference-project-transport-plan-stop 123 \
    --preserve-stop-on-delete`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanStopInsertionsCreate,
	}
	initDoProjectTransportPlanStopInsertionsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanStopInsertionsCmd.AddCommand(newDoProjectTransportPlanStopInsertionsCreateCmd())
}

func initDoProjectTransportPlanStopInsertionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("mode", "", "Insertion mode")
	cmd.Flags().String("reuse-half", "", "Reuse half (upstream or downstream)")
	cmd.Flags().String("boundary-choice", "", "Boundary choice (join_upstream or join_downstream)")
	cmd.Flags().String("planned-event-time-start-at", "", "Planned event time start (RFC3339)")
	cmd.Flags().String("planned-event-time-end-at", "", "Planned event time end (RFC3339)")
	cmd.Flags().Bool("preserve-stop-on-delete", false, "Preserve stop record in delete mode")
	cmd.Flags().String("reference-project-transport-plan-stop", "", "Reference project transport plan stop ID")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID")
	cmd.Flags().String("project-transport-location", "", "Project transport location ID")
	cmd.Flags().String("project-transport-plan-segment-set", "", "Project transport plan segment set ID")
	cmd.Flags().String("planned-completion-event-type", "", "Planned completion event type ID")
	cmd.Flags().String("existing-project-transport-plan-stop", "", "Existing project transport plan stop ID")
	cmd.Flags().String("stop-to-move", "", "Stop to move ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("mode")
	_ = cmd.MarkFlagRequired("reuse-half")
}

func runDoProjectTransportPlanStopInsertionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanStopInsertionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Mode) == "" {
		err := fmt.Errorf("--mode is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ReuseHalf) == "" {
		err := fmt.Errorf("--reuse-half is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"mode":       opts.Mode,
		"reuse-half": opts.ReuseHalf,
	}

	if strings.TrimSpace(opts.BoundaryChoice) != "" {
		attributes["boundary-choice"] = opts.BoundaryChoice
	}
	if strings.TrimSpace(opts.PlannedEventTimeStartAt) != "" {
		attributes["planned-event-time-start-at"] = opts.PlannedEventTimeStartAt
	}
	if strings.TrimSpace(opts.PlannedEventTimeEndAt) != "" {
		attributes["planned-event-time-end-at"] = opts.PlannedEventTimeEndAt
	}
	if cmd.Flags().Changed("preserve-stop-on-delete") {
		attributes["preserve-stop-on-delete"] = opts.PreserveStopOnDelete
	}

	relationships := map[string]any{}

	if strings.TrimSpace(opts.ReferenceProjectTransportPlanStop) != "" {
		relationships["reference-project-transport-plan-stop"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-stops",
				"id":   opts.ReferenceProjectTransportPlanStop,
			},
		}
	}
	if strings.TrimSpace(opts.ProjectTransportPlan) != "" {
		relationships["project-transport-plan"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlan,
			},
		}
	}
	if strings.TrimSpace(opts.ProjectTransportLocation) != "" {
		relationships["project-transport-location"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-locations",
				"id":   opts.ProjectTransportLocation,
			},
		}
	}
	if strings.TrimSpace(opts.ProjectTransportPlanSegmentSet) != "" {
		relationships["project-transport-plan-segment-set"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-segment-sets",
				"id":   opts.ProjectTransportPlanSegmentSet,
			},
		}
	}
	if strings.TrimSpace(opts.PlannedCompletionEventType) != "" {
		relationships["planned-completion-event-type"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-event-types",
				"id":   opts.PlannedCompletionEventType,
			},
		}
	}
	if strings.TrimSpace(opts.ExistingProjectTransportPlanStop) != "" {
		relationships["existing-project-transport-plan-stop"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-stops",
				"id":   opts.ExistingProjectTransportPlanStop,
			},
		}
	}
	if strings.TrimSpace(opts.StopToMove) != "" {
		relationships["stop-to-move"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-stops",
				"id":   opts.StopToMove,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-stop-insertions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-stop-insertions", jsonBody)
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

	row := buildProjectTransportPlanStopInsertionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan stop insertion %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanStopInsertionsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanStopInsertionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	mode, _ := cmd.Flags().GetString("mode")
	reuseHalf, _ := cmd.Flags().GetString("reuse-half")
	boundaryChoice, _ := cmd.Flags().GetString("boundary-choice")
	plannedEventTimeStartAt, _ := cmd.Flags().GetString("planned-event-time-start-at")
	plannedEventTimeEndAt, _ := cmd.Flags().GetString("planned-event-time-end-at")
	preserveStopOnDelete, _ := cmd.Flags().GetBool("preserve-stop-on-delete")
	referenceProjectTransportPlanStop, _ := cmd.Flags().GetString("reference-project-transport-plan-stop")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportLocation, _ := cmd.Flags().GetString("project-transport-location")
	projectTransportPlanSegmentSet, _ := cmd.Flags().GetString("project-transport-plan-segment-set")
	plannedCompletionEventType, _ := cmd.Flags().GetString("planned-completion-event-type")
	existingProjectTransportPlanStop, _ := cmd.Flags().GetString("existing-project-transport-plan-stop")
	stopToMove, _ := cmd.Flags().GetString("stop-to-move")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanStopInsertionsCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		Mode:                              mode,
		ReuseHalf:                         reuseHalf,
		BoundaryChoice:                    boundaryChoice,
		PlannedEventTimeStartAt:           plannedEventTimeStartAt,
		PlannedEventTimeEndAt:             plannedEventTimeEndAt,
		PreserveStopOnDelete:              preserveStopOnDelete,
		ReferenceProjectTransportPlanStop: referenceProjectTransportPlanStop,
		ProjectTransportPlan:              projectTransportPlan,
		ProjectTransportLocation:          projectTransportLocation,
		ProjectTransportPlanSegmentSet:    projectTransportPlanSegmentSet,
		PlannedCompletionEventType:        plannedCompletionEventType,
		ExistingProjectTransportPlanStop:  existingProjectTransportPlanStop,
		StopToMove:                        stopToMove,
	}, nil
}
