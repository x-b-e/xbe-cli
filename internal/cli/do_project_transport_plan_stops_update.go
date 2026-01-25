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

type doProjectTransportPlanStopsUpdateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ID                           string
	ProjectTransportLocationID   string
	PlannedCompletionEventTypeID string
	Status                       string
	Position                     string
}

func newDoProjectTransportPlanStopsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan stop",
		Long: `Update a project transport plan stop.

Provide the stop ID as an argument, then use flags to specify which
fields to update. Only specified fields will be modified.

Optional attributes:
  --status    Stop status (planned, started, finished, cancelled)
  --position  Stop position

Optional relationships:
  --project-transport-location     Project transport location ID
  --planned-completion-event-type  Planned completion event type ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update stop status
  xbe do project-transport-plan-stops update 123 --status started

  # Update stop location
  xbe do project-transport-plan-stops update 123 --project-transport-location 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanStopsUpdate,
	}
	initDoProjectTransportPlanStopsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanStopsCmd.AddCommand(newDoProjectTransportPlanStopsUpdateCmd())
}

func initDoProjectTransportPlanStopsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-location", "", "Project transport location ID")
	cmd.Flags().String("planned-completion-event-type", "", "Planned completion event type ID")
	cmd.Flags().String("status", "", "Stop status (planned, started, finished, cancelled)")
	cmd.Flags().String("position", "", "Stop position")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanStopsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanStopsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("position") {
		attributes["position"] = opts.Position
	}

	if cmd.Flags().Changed("project-transport-location") {
		if opts.ProjectTransportLocationID == "" {
			relationships["project-transport-location"] = map[string]any{"data": nil}
		} else {
			relationships["project-transport-location"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-locations",
					"id":   opts.ProjectTransportLocationID,
				},
			}
		}
	}
	if cmd.Flags().Changed("planned-completion-event-type") {
		if opts.PlannedCompletionEventTypeID == "" {
			relationships["planned-completion-event-type"] = map[string]any{"data": nil}
		} else {
			relationships["planned-completion-event-type"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-event-types",
					"id":   opts.PlannedCompletionEventTypeID,
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
		"type": "project-transport-plan-stops",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-stops/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan stop %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanStopsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanStopsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportLocation, _ := cmd.Flags().GetString("project-transport-location")
	plannedCompletionEventType, _ := cmd.Flags().GetString("planned-completion-event-type")
	status, _ := cmd.Flags().GetString("status")
	position, _ := cmd.Flags().GetString("position")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanStopsUpdateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ID:                           args[0],
		ProjectTransportLocationID:   projectTransportLocation,
		PlannedCompletionEventTypeID: plannedCompletionEventType,
		Status:                       status,
		Position:                     position,
	}, nil
}
