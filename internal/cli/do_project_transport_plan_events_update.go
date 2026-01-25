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

type doProjectTransportPlanEventsUpdateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ID                        string
	ProjectTransportEventType string
	ProjectTransportLocation  string
	ProjectMaterialType       string
	ProjectTransportPlanStop  string
	ExternalTMSEventID        string
	Position                  int
}

func newDoProjectTransportPlanEventsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan event",
		Long: `Update a project transport plan event.

Note: project transport plan cannot be changed after creation.

Optional flags:
  --external-tms-event-id        External TMS event identifier (use empty to clear)
  --position                     Event position (0-based index)
  --project-transport-event-type Project transport event type ID
  --project-transport-location   Project transport location ID (use empty to clear)
  --project-material-type        Project material type ID (use empty to clear)
  --project-transport-plan-stop  Project transport plan stop ID (use empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update external TMS event ID
  xbe do project-transport-plan-events update 123 --external-tms-event-id "EVT-002"

  # Update event type
  xbe do project-transport-plan-events update 123 --project-transport-event-type 456

  # Clear location
  xbe do project-transport-plan-events update 123 --project-transport-location ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanEventsUpdate,
	}
	initDoProjectTransportPlanEventsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanEventsCmd.AddCommand(newDoProjectTransportPlanEventsUpdateCmd())
}

func initDoProjectTransportPlanEventsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("external-tms-event-id", "", "External TMS event identifier (use empty to clear)")
	cmd.Flags().Int("position", 0, "Event position (0-based index)")
	cmd.Flags().String("project-transport-event-type", "", "Project transport event type ID")
	cmd.Flags().String("project-transport-location", "", "Project transport location ID (use empty to clear)")
	cmd.Flags().String("project-material-type", "", "Project material type ID (use empty to clear)")
	cmd.Flags().String("project-transport-plan-stop", "", "Project transport plan stop ID (use empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanEventsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanEventsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("external-tms-event-id") {
		attributes["external-tms-event-id"] = opts.ExternalTMSEventID
	}
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}

	if cmd.Flags().Changed("project-transport-event-type") {
		if strings.TrimSpace(opts.ProjectTransportEventType) == "" {
			err := fmt.Errorf("--project-transport-event-type cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["project-transport-event-type"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-event-types",
				"id":   opts.ProjectTransportEventType,
			},
		}
	}

	if cmd.Flags().Changed("project-transport-location") {
		if strings.TrimSpace(opts.ProjectTransportLocation) == "" {
			relationships["project-transport-location"] = map[string]any{"data": nil}
		} else {
			relationships["project-transport-location"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-locations",
					"id":   opts.ProjectTransportLocation,
				},
			}
		}
	}

	if cmd.Flags().Changed("project-material-type") {
		if strings.TrimSpace(opts.ProjectMaterialType) == "" {
			relationships["project-material-type"] = map[string]any{"data": nil}
		} else {
			relationships["project-material-type"] = map[string]any{
				"data": map[string]any{
					"type": "project-material-types",
					"id":   opts.ProjectMaterialType,
				},
			}
		}
	}

	if cmd.Flags().Changed("project-transport-plan-stop") {
		if strings.TrimSpace(opts.ProjectTransportPlanStop) == "" {
			relationships["project-transport-plan-stop"] = map[string]any{"data": nil}
		} else {
			relationships["project-transport-plan-stop"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-plan-stops",
					"id":   opts.ProjectTransportPlanStop,
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
		"type": "project-transport-plan-events",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-events/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan event %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanEventsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanEventsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	externalTMSEventID, _ := cmd.Flags().GetString("external-tms-event-id")
	position, _ := cmd.Flags().GetInt("position")
	projectTransportEventType, _ := cmd.Flags().GetString("project-transport-event-type")
	projectTransportLocation, _ := cmd.Flags().GetString("project-transport-location")
	projectMaterialType, _ := cmd.Flags().GetString("project-material-type")
	projectTransportPlanStop, _ := cmd.Flags().GetString("project-transport-plan-stop")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanEventsUpdateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ID:                        strings.TrimSpace(args[0]),
		ProjectTransportEventType: projectTransportEventType,
		ProjectTransportLocation:  projectTransportLocation,
		ProjectMaterialType:       projectMaterialType,
		ProjectTransportPlanStop:  projectTransportPlanStop,
		ExternalTMSEventID:        externalTMSEventID,
		Position:                  position,
	}, nil
}
