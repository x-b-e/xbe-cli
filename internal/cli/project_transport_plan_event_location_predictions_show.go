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

type projectTransportPlanEventLocationPredictionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanEventLocationPredictionDetails struct {
	ID                                string `json:"id"`
	Predictions                       any    `json:"predictions,omitempty"`
	ProjectTransportPlanEventID       string `json:"project_transport_plan_event_id,omitempty"`
	TransportOrderID                  string `json:"transport_order_id,omitempty"`
	ProjectTransportEventTypeExplicit string `json:"project_transport_event_type_id_explicit,omitempty"`
	StrategySetIDExplicit             string `json:"strategy_set_id_explicit,omitempty"`
	EventPositionExplicit             string `json:"event_position_explicit,omitempty"`
	BrokerIDExplicit                  string `json:"broker_id_explicit,omitempty"`
}

func newProjectTransportPlanEventLocationPredictionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan event location prediction details",
		Long: `Show the full details of a project transport plan event location prediction.

Output Fields:
  ID
  Predictions
  Project Transport Plan Event ID
  Transport Order ID
  Project Transport Event Type ID Explicit
  Strategy Set ID Explicit
  Event Position Explicit
  Broker ID Explicit

Arguments:
  <id>    The prediction ID (required). You can find IDs using the list command.`,
		Example: `  # Show prediction details
  xbe view project-transport-plan-event-location-predictions show 123

  # JSON output
  xbe view project-transport-plan-event-location-predictions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanEventLocationPredictionsShow,
	}
	initProjectTransportPlanEventLocationPredictionsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanEventLocationPredictionsCmd.AddCommand(newProjectTransportPlanEventLocationPredictionsShowCmd())
}

func initProjectTransportPlanEventLocationPredictionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanEventLocationPredictionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanEventLocationPredictionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan event location prediction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-event-location-predictions/"+id, nil)
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

	return renderProjectTransportPlanEventLocationPredictionDetails(cmd, details)
}

func parseProjectTransportPlanEventLocationPredictionsShowOptions(cmd *cobra.Command) (projectTransportPlanEventLocationPredictionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanEventLocationPredictionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanEventLocationPredictionDetails(resp jsonAPISingleResponse) projectTransportPlanEventLocationPredictionDetails {
	attrs := resp.Data.Attributes
	details := projectTransportPlanEventLocationPredictionDetails{
		ID:                                resp.Data.ID,
		ProjectTransportEventTypeExplicit: stringAttr(attrs, "project-transport-event-type-id-explicit"),
		StrategySetIDExplicit:             stringAttr(attrs, "strategy-set-id-explicit"),
		EventPositionExplicit:             stringAttr(attrs, "event-position-explicit"),
		BrokerIDExplicit:                  stringAttr(attrs, "broker-id-explicit"),
	}

	if predictions, ok := attrs["predictions"]; ok {
		details.Predictions = predictions
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan-event"]; ok && rel.Data != nil {
		details.ProjectTransportPlanEventID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["transport-order"]; ok && rel.Data != nil {
		details.TransportOrderID = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanEventLocationPredictionDetails(cmd *cobra.Command, details projectTransportPlanEventLocationPredictionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if formatted := formatJSONValue(details.Predictions); formatted != "" {
		fmt.Fprintf(out, "Predictions: %s\n", formatted)
	}
	if details.ProjectTransportPlanEventID != "" {
		fmt.Fprintf(out, "Project Transport Plan Event ID: %s\n", details.ProjectTransportPlanEventID)
	}
	if details.TransportOrderID != "" {
		fmt.Fprintf(out, "Transport Order ID: %s\n", details.TransportOrderID)
	}
	if details.ProjectTransportEventTypeExplicit != "" {
		fmt.Fprintf(out, "Project Transport Event Type ID Explicit: %s\n", details.ProjectTransportEventTypeExplicit)
	}
	if details.StrategySetIDExplicit != "" {
		fmt.Fprintf(out, "Strategy Set ID Explicit: %s\n", details.StrategySetIDExplicit)
	}
	if details.EventPositionExplicit != "" {
		fmt.Fprintf(out, "Event Position Explicit: %s\n", details.EventPositionExplicit)
	}
	if details.BrokerIDExplicit != "" {
		fmt.Fprintf(out, "Broker ID Explicit: %s\n", details.BrokerIDExplicit)
	}

	return nil
}
