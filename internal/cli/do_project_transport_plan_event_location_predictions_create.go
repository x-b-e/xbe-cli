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

type doProjectTransportPlanEventLocationPredictionsCreateOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	ProjectTransportPlanEvent   string
	TransportOrder              string
	ProjectTransportEventTypeID string
	StrategySetIDExplicit       string
	EventPositionExplicit       int
	BrokerIDExplicit            string
}

func newDoProjectTransportPlanEventLocationPredictionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan event location prediction",
		Long: `Create a project transport plan event location prediction.

Required:
  --transport-order  Transport order ID

Context (one of the following):
  --project-transport-plan-event                     Project transport plan event ID

OR explicit context:
  --project-transport-event-type-id-explicit         Project transport event type ID
  --event-position-explicit                          Event position (integer)
  --broker-id-explicit                               Broker ID

Optional:
  --strategy-set-id-explicit                         Strategy set ID`,
		Example: `  # Create predictions for a project transport plan event
  xbe do project-transport-plan-event-location-predictions create \
    --project-transport-plan-event 123 \
    --transport-order 456

  # Create predictions with explicit context
  xbe do project-transport-plan-event-location-predictions create \
    --transport-order 456 \
    --project-transport-event-type-id-explicit 789 \
    --event-position-explicit 1 \
    --broker-id-explicit 321`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanEventLocationPredictionsCreate,
	}
	initDoProjectTransportPlanEventLocationPredictionsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanEventLocationPredictionsCmd.AddCommand(newDoProjectTransportPlanEventLocationPredictionsCreateCmd())
}

func initDoProjectTransportPlanEventLocationPredictionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan-event", "", "Project transport plan event ID")
	cmd.Flags().String("transport-order", "", "Transport order ID (required)")
	cmd.Flags().String("project-transport-event-type-id-explicit", "", "Project transport event type ID")
	cmd.Flags().String("strategy-set-id-explicit", "", "Strategy set ID")
	cmd.Flags().Int("event-position-explicit", 0, "Event position (integer)")
	cmd.Flags().String("broker-id-explicit", "", "Broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanEventLocationPredictionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanEventLocationPredictionsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.TransportOrder) == "" {
		err := fmt.Errorf("--transport-order is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	hasEvent := strings.TrimSpace(opts.ProjectTransportPlanEvent) != ""
	hasExplicitEventType := strings.TrimSpace(opts.ProjectTransportEventTypeID) != ""
	hasExplicitBroker := strings.TrimSpace(opts.BrokerIDExplicit) != ""
	hasExplicitEventPosition := cmd.Flags().Changed("event-position-explicit")
	if !hasEvent && !(hasExplicitEventType && hasExplicitEventPosition && hasExplicitBroker) {
		err := fmt.Errorf("--project-transport-plan-event or --project-transport-event-type-id-explicit/--event-position-explicit/--broker-id-explicit is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.ProjectTransportEventTypeID != "" {
		attributes["project-transport-event-type-id-explicit"] = opts.ProjectTransportEventTypeID
	}
	if opts.StrategySetIDExplicit != "" {
		attributes["strategy-set-id-explicit"] = opts.StrategySetIDExplicit
	}
	if cmd.Flags().Changed("event-position-explicit") {
		attributes["event-position-explicit"] = opts.EventPositionExplicit
	}
	if opts.BrokerIDExplicit != "" {
		attributes["broker-id-explicit"] = opts.BrokerIDExplicit
	}

	relationships := map[string]any{
		"transport-order": map[string]any{
			"data": map[string]any{
				"type": "transport-orders",
				"id":   opts.TransportOrder,
			},
		},
	}
	if opts.ProjectTransportPlanEvent != "" {
		relationships["project-transport-plan-event"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plan-events",
				"id":   opts.ProjectTransportPlanEvent,
			},
		}
	}

	data := map[string]any{
		"type":          "project-transport-plan-event-location-predictions",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-event-location-predictions", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan event location prediction %s\n", details.ID)
	return nil
}

func parseDoProjectTransportPlanEventLocationPredictionsCreateOptions(cmd *cobra.Command) (doProjectTransportPlanEventLocationPredictionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlanEvent, _ := cmd.Flags().GetString("project-transport-plan-event")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	projectTransportEventTypeID, _ := cmd.Flags().GetString("project-transport-event-type-id-explicit")
	strategySetIDExplicit, _ := cmd.Flags().GetString("strategy-set-id-explicit")
	eventPositionExplicit, _ := cmd.Flags().GetInt("event-position-explicit")
	brokerIDExplicit, _ := cmd.Flags().GetString("broker-id-explicit")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanEventLocationPredictionsCreateOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		ProjectTransportPlanEvent:   projectTransportPlanEvent,
		TransportOrder:              transportOrder,
		ProjectTransportEventTypeID: projectTransportEventTypeID,
		StrategySetIDExplicit:       strategySetIDExplicit,
		EventPositionExplicit:       eventPositionExplicit,
		BrokerIDExplicit:            brokerIDExplicit,
	}, nil
}
