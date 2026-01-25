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

type doProjectTransportPlanPlannedEventTimeSchedulesCreateOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	ProjectTransportPlan      string
	TransportOrder            string
	PlanData                  string
	RespectProvidedEventTimes bool
}

func newDoProjectTransportPlanPlannedEventTimeSchedulesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Generate a planned event time schedule",
		Long: `Generate a planned event time schedule.

Required input (choose one):
  --project-transport-plan           Project transport plan ID
  --transport-order + --plan-data    Transport order ID with plan data JSON

Optional flags:
  --respect-provided-event-times     Respect provided event times when computing the schedule
  --plan-data                        Plan data JSON object with events (required with --transport-order)`,
		Example: `  # Generate a schedule for a project transport plan
  xbe do project-transport-plan-planned-event-time-schedules create \
    --project-transport-plan 123

  # Generate a schedule using explicit event data
  xbe do project-transport-plan-planned-event-time-schedules create \
    --transport-order 456 \
    --plan-data '{"events":[{"location_id":1,"event_type_id":2}]}'

  # Respect provided event times
  xbe do project-transport-plan-planned-event-time-schedules create \
    --project-transport-plan 123 \
    --respect-provided-event-times

  # JSON output
  xbe do project-transport-plan-planned-event-time-schedules create \
    --project-transport-plan 123 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanPlannedEventTimeSchedulesCreate,
	}
	initDoProjectTransportPlanPlannedEventTimeSchedulesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanPlannedEventTimeSchedulesCmd.AddCommand(newDoProjectTransportPlanPlannedEventTimeSchedulesCreateCmd())
}

func initDoProjectTransportPlanPlannedEventTimeSchedulesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-transport-plan", "", "Project transport plan ID")
	cmd.Flags().String("transport-order", "", "Transport order ID")
	cmd.Flags().String("plan-data", "", "Plan data JSON object (required with --transport-order)")
	cmd.Flags().Bool("respect-provided-event-times", false, "Respect provided event times when computing schedule")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanPlannedEventTimeSchedulesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanPlannedEventTimeSchedulesCreateOptions(cmd)
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

	planID := strings.TrimSpace(opts.ProjectTransportPlan)
	orderID := strings.TrimSpace(opts.TransportOrder)
	planDataRaw := strings.TrimSpace(opts.PlanData)

	if planID == "" {
		if orderID == "" || planDataRaw == "" {
			err := fmt.Errorf("provide either --project-transport-plan or (--transport-order and --plan-data)")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	} else if orderID != "" || planDataRaw != "" {
		err := fmt.Errorf("use --project-transport-plan by itself or use --transport-order with --plan-data")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("respect-provided-event-times") {
		attributes["respect-provided-event-times"] = opts.RespectProvidedEventTimes
	}

	if planDataRaw != "" {
		var planData map[string]any
		if err := json.Unmarshal([]byte(planDataRaw), &planData); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), fmt.Errorf("invalid --plan-data JSON: %w", err))
			return err
		}
		attributes["plan-data"] = planData
	}

	relationships := map[string]any{}
	if planID != "" {
		relationships["project-transport-plan"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-plans",
				"id":   planID,
			},
		}
	}
	if orderID != "" {
		relationships["transport-order"] = map[string]any{
			"data": map[string]any{
				"type": "transport-orders",
				"id":   orderID,
			},
		}
	}

	data := map[string]any{
		"type":          "project-transport-plan-planned-event-time-schedules",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-planned-event-time-schedules", jsonBody)
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

	details := buildProjectTransportPlanPlannedEventTimeScheduleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanPlannedEventTimeScheduleDetails(cmd, details)
}

func parseDoProjectTransportPlanPlannedEventTimeSchedulesCreateOptions(cmd *cobra.Command) (doProjectTransportPlanPlannedEventTimeSchedulesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	planData, _ := cmd.Flags().GetString("plan-data")
	respectProvidedEventTimes, _ := cmd.Flags().GetBool("respect-provided-event-times")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanPlannedEventTimeSchedulesCreateOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		ProjectTransportPlan:      projectTransportPlan,
		TransportOrder:            transportOrder,
		PlanData:                  planData,
		RespectProvidedEventTimes: respectProvidedEventTimes,
	}, nil
}
