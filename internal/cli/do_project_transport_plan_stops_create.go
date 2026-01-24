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

type doProjectTransportPlanStopsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ProjectTransportPlanID       string
	ProjectTransportLocationID   string
	PlannedCompletionEventTypeID string
	Status                       string
	Position                     string
}

func newDoProjectTransportPlanStopsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan stop",
		Long: `Create a project transport plan stop.

Required flags:
  --project-transport-plan      Project transport plan ID (required)
  --project-transport-location  Project transport location ID (required)

Optional attributes:
  --status                      Stop status (planned, started, finished, cancelled)
  --position                    Stop position

Optional relationships:
  --planned-completion-event-type  Planned completion event type ID`,
		Example: `  # Create a stop
  xbe do project-transport-plan-stops create \
    --project-transport-plan 123 \
    --project-transport-location 456

  # Create a stop with status and position
  xbe do project-transport-plan-stops create \
    --project-transport-plan 123 \
    --project-transport-location 456 \
    --status planned \
    --position 1

  # JSON output
  xbe do project-transport-plan-stops create \
    --project-transport-plan 123 \
    --project-transport-location 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanStopsCreate,
	}
	initDoProjectTransportPlanStopsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanStopsCmd.AddCommand(newDoProjectTransportPlanStopsCreateCmd())
}

func initDoProjectTransportPlanStopsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID (required)")
	cmd.Flags().String("project-transport-location", "", "Project transport location ID (required)")
	cmd.Flags().String("planned-completion-event-type", "", "Planned completion event type ID")
	cmd.Flags().String("status", "", "Stop status (planned, started, finished, cancelled)")
	cmd.Flags().String("position", "", "Stop position")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanStopsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanStopsCreateOptions(cmd)
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
	if strings.TrimSpace(opts.ProjectTransportLocationID) == "" {
		err := fmt.Errorf("--project-transport-location is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.Position != "" {
		attributes["position"] = opts.Position
	}

	relationships := map[string]any{
		"project-transport-plan": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlanID,
			},
		},
		"project-transport-location": map[string]any{
			"data": map[string]any{
				"type": "project-transport-locations",
				"id":   opts.ProjectTransportLocationID,
			},
		},
	}

	if opts.PlannedCompletionEventTypeID != "" {
		relationships["planned-completion-event-type"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-event-types",
				"id":   opts.PlannedCompletionEventTypeID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-stops",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-stops", jsonBody)
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

	row := buildProjectTransportPlanStopRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan stop %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanStopsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanStopsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportLocation, _ := cmd.Flags().GetString("project-transport-location")
	plannedCompletionEventType, _ := cmd.Flags().GetString("planned-completion-event-type")
	status, _ := cmd.Flags().GetString("status")
	position, _ := cmd.Flags().GetString("position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanStopsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ProjectTransportPlanID:       projectTransportPlan,
		ProjectTransportLocationID:   projectTransportLocation,
		PlannedCompletionEventTypeID: plannedCompletionEventType,
		Status:                       status,
		Position:                     position,
	}, nil
}
