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

type doProjectTransportPlanEventsCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ProjectTransportPlan      string
	ProjectTransportEventType string
	ProjectTransportLocation  string
	ProjectMaterialType       string
	ProjectTransportPlanStop  string
	ExternalTMSEventID        string
	Position                  int
}

func newDoProjectTransportPlanEventsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan event",
		Long: `Create a project transport plan event.

Required flags:
  --project-transport-plan       Project transport plan ID
  --project-transport-event-type Project transport event type ID

Optional flags:
  --external-tms-event-id       External TMS event identifier
  --position                    Event position (0-based index)
  --project-transport-location  Project transport location ID
  --project-material-type       Project material type ID
  --project-transport-plan-stop Project transport plan stop ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project transport plan event
  xbe do project-transport-plan-events create \
    --project-transport-plan 123 \
    --project-transport-event-type 456

  # Create with optional fields
  xbe do project-transport-plan-events create \
    --project-transport-plan 123 \
    --project-transport-event-type 456 \
    --external-tms-event-id "EVT-001" \
    --position 0 \
    --project-transport-location 789`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanEventsCreate,
	}
	initDoProjectTransportPlanEventsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanEventsCmd.AddCommand(newDoProjectTransportPlanEventsCreateCmd())
}

func initDoProjectTransportPlanEventsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID")
	cmd.Flags().String("project-transport-event-type", "", "Project transport event type ID")
	cmd.Flags().String("project-transport-location", "", "Project transport location ID")
	cmd.Flags().String("project-material-type", "", "Project material type ID")
	cmd.Flags().String("project-transport-plan-stop", "", "Project transport plan stop ID")
	cmd.Flags().String("external-tms-event-id", "", "External TMS event identifier")
	cmd.Flags().Int("position", 0, "Event position (0-based index)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-transport-plan")
	_ = cmd.MarkFlagRequired("project-transport-event-type")
}

func runDoProjectTransportPlanEventsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanEventsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.ProjectTransportPlan) == "" {
		err := fmt.Errorf("--project-transport-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ProjectTransportEventType) == "" {
		err := fmt.Errorf("--project-transport-event-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.ExternalTMSEventID != "" {
		attributes["external-tms-event-id"] = opts.ExternalTMSEventID
	}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}

	relationships := map[string]any{
		"project-transport-plan": map[string]any{
			"data": map[string]any{
				"type": "project-transport-plans",
				"id":   opts.ProjectTransportPlan,
			},
		},
		"project-transport-event-type": map[string]any{
			"data": map[string]any{
				"type": "project-transport-event-types",
				"id":   opts.ProjectTransportEventType,
			},
		},
	}

	if strings.TrimSpace(opts.ProjectTransportLocation) != "" {
		relationships["project-transport-location"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-locations",
				"id":   opts.ProjectTransportLocation,
			},
		}
	}
	if strings.TrimSpace(opts.ProjectMaterialType) != "" {
		relationships["project-material-type"] = map[string]any{
			"data": map[string]any{
				"type": "project-material-types",
				"id":   opts.ProjectMaterialType,
			},
		}
	}
	if strings.TrimSpace(opts.ProjectTransportPlanStop) != "" {
		relationships["project-transport-plan-stop"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-stops",
				"id":   opts.ProjectTransportPlanStop,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-plan-events",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-events", jsonBody)
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

	row := buildProjectTransportPlanEventRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan event %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanEventsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanEventsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportEventType, _ := cmd.Flags().GetString("project-transport-event-type")
	projectTransportLocation, _ := cmd.Flags().GetString("project-transport-location")
	projectMaterialType, _ := cmd.Flags().GetString("project-material-type")
	projectTransportPlanStop, _ := cmd.Flags().GetString("project-transport-plan-stop")
	externalTMSEventID, _ := cmd.Flags().GetString("external-tms-event-id")
	position, _ := cmd.Flags().GetInt("position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanEventsCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ProjectTransportPlan:      projectTransportPlan,
		ProjectTransportEventType: projectTransportEventType,
		ProjectTransportLocation:  projectTransportLocation,
		ProjectMaterialType:       projectMaterialType,
		ProjectTransportPlanStop:  projectTransportPlanStop,
		ExternalTMSEventID:        externalTMSEventID,
		Position:                  position,
	}, nil
}

func buildProjectTransportPlanEventRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanEventRow {
	attrs := resp.Data.Attributes

	row := projectTransportPlanEventRow{
		ID:                 resp.Data.ID,
		Position:           intAttr(attrs, "position"),
		ExternalTMSEventID: stringAttr(attrs, "external-tms-event-id"),
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		row.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
		row.ProjectTransportEventTypeID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-location"]; ok && rel.Data != nil {
		row.ProjectTransportLocationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-stop"]; ok && rel.Data != nil {
		row.ProjectTransportPlanStopID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-organization"]; ok && rel.Data != nil {
		row.ProjectTransportOrganizationID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-material-type"]; ok && rel.Data != nil {
		row.ProjectMaterialTypeID = rel.Data.ID
	}

	return row
}
