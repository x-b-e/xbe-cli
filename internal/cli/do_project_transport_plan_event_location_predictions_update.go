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

type doProjectTransportPlanEventLocationPredictionsUpdateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ID                          string
	ProjectTransportPlanEvent   string
	TransportOrder              string
	ProjectTransportEventTypeID string
	StrategySetIDExplicit       string
	EventPositionExplicit       int
	BrokerIDExplicit            string
}

func newDoProjectTransportPlanEventLocationPredictionsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan event location prediction",
		Long: `Update a project transport plan event location prediction.

Optional:
  --project-transport-plan-event               Project transport plan event ID
  --transport-order                            Transport order ID
  --project-transport-event-type-id-explicit   Project transport event type ID
  --event-position-explicit                    Event position (integer)
  --broker-id-explicit                         Broker ID
  --strategy-set-id-explicit                   Strategy set ID`,
		Example: `  # Update explicit strategy set
  xbe do project-transport-plan-event-location-predictions update 123 \
    --strategy-set-id-explicit 456

  # Repoint to a different transport order
  xbe do project-transport-plan-event-location-predictions update 123 \
    --transport-order 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanEventLocationPredictionsUpdate,
	}
	initDoProjectTransportPlanEventLocationPredictionsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanEventLocationPredictionsCmd.AddCommand(newDoProjectTransportPlanEventLocationPredictionsUpdateCmd())
}

func initDoProjectTransportPlanEventLocationPredictionsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan-event", "", "Project transport plan event ID")
	cmd.Flags().String("transport-order", "", "Transport order ID")
	cmd.Flags().String("project-transport-event-type-id-explicit", "", "Project transport event type ID")
	cmd.Flags().String("strategy-set-id-explicit", "", "Strategy set ID")
	cmd.Flags().Int("event-position-explicit", 0, "Event position (integer)")
	cmd.Flags().String("broker-id-explicit", "", "Broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanEventLocationPredictionsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanEventLocationPredictionsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("project-transport-event-type-id-explicit") {
		attributes["project-transport-event-type-id-explicit"] = opts.ProjectTransportEventTypeID
	}
	if cmd.Flags().Changed("strategy-set-id-explicit") {
		attributes["strategy-set-id-explicit"] = opts.StrategySetIDExplicit
	}
	if cmd.Flags().Changed("event-position-explicit") {
		attributes["event-position-explicit"] = opts.EventPositionExplicit
	}
	if cmd.Flags().Changed("broker-id-explicit") {
		attributes["broker-id-explicit"] = opts.BrokerIDExplicit
	}

	relationships := map[string]any{}
	if opts.ProjectTransportPlanEvent != "" {
		relationships["project-transport-plan-event"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-events",
				"id":   opts.ProjectTransportPlanEvent,
			},
		}
	}
	if opts.TransportOrder != "" {
		relationships["transport-order"] = map[string]any{
			"data": map[string]any{
				"type": "transport-orders",
				"id":   opts.TransportOrder,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-transport-plan-event-location-predictions",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plan-event-location-predictions/"+opts.ID, jsonBody)
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

	details := buildProjectTransportPlanEventLocationPredictionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan event location prediction %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanEventLocationPredictionsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanEventLocationPredictionsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlanEvent, _ := cmd.Flags().GetString("project-transport-plan-event")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	projectTransportEventTypeID, _ := cmd.Flags().GetString("project-transport-event-type-id-explicit")
	strategySetIDExplicit, _ := cmd.Flags().GetString("strategy-set-id-explicit")
	eventPositionExplicit, _ := cmd.Flags().GetInt("event-position-explicit")
	brokerIDExplicit, _ := cmd.Flags().GetString("broker-id-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanEventLocationPredictionsUpdateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ID:                          args[0],
		ProjectTransportPlanEvent:   projectTransportPlanEvent,
		TransportOrder:              transportOrder,
		ProjectTransportEventTypeID: projectTransportEventTypeID,
		StrategySetIDExplicit:       strategySetIDExplicit,
		EventPositionExplicit:       eventPositionExplicit,
		BrokerIDExplicit:            brokerIDExplicit,
	}, nil
}
