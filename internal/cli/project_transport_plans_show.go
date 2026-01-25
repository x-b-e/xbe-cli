package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectTransportPlansShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanDetails struct {
	ID                                string   `json:"id"`
	Status                            string   `json:"status,omitempty"`
	SegmentMiles                      any      `json:"segment_miles,omitempty"`
	EventTimesAtMin                   string   `json:"event_times_at_min,omitempty"`
	EventTimesAtMax                   string   `json:"event_times_at_max,omitempty"`
	EventTimesOnMin                   string   `json:"event_times_on_min,omitempty"`
	EventTimesOnMax                   string   `json:"event_times_on_max,omitempty"`
	StrategySetPredictionPosition     string   `json:"strategy_set_prediction_position,omitempty"`
	SkipActualization                 bool     `json:"skip_actualization"`
	SkipAssignmentRulesValidation     bool     `json:"skip_assignment_rules_validation"`
	AssignmentRuleOverrideReason      string   `json:"assignment_rule_override_reason,omitempty"`
	CreatedAt                         string   `json:"created_at,omitempty"`
	UpdatedAt                         string   `json:"updated_at,omitempty"`
	ProjectID                         string   `json:"project_id,omitempty"`
	BrokerID                          string   `json:"broker_id,omitempty"`
	CreatedByID                       string   `json:"created_by_id,omitempty"`
	ProjectTransportPlanStrategySetID string   `json:"project_transport_plan_strategy_set_id,omitempty"`
	ProjectTransportPlanStopIDs       []string `json:"project_transport_plan_stop_ids,omitempty"`
	ProjectTransportPlanSegmentIDs    []string `json:"project_transport_plan_segment_ids,omitempty"`
	ProjectTransportPlanSegmentSetIDs []string `json:"project_transport_plan_segment_set_ids,omitempty"`
	ProjectTransportPlanEventIDs      []string `json:"project_transport_plan_event_ids,omitempty"`
	ProjectTransportPlanDriverIDs     []string `json:"project_transport_plan_driver_ids,omitempty"`
	ProjectTransportPlanTractorIDs    []string `json:"project_transport_plan_tractor_ids,omitempty"`
	ProjectTransportPlanTrailerIDs    []string `json:"project_transport_plan_trailer_ids,omitempty"`
	TransportOrderIDs                 []string `json:"transport_order_ids,omitempty"`
	NearestProjectOfficeIDs           []string `json:"nearest_project_office_ids,omitempty"`
	IncidentIDs                       []string `json:"incident_ids,omitempty"`
}

func newProjectTransportPlansShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan details",
		Long: `Show the full details of a project transport plan.

Output Fields:
  ID        Plan identifier
  STATUS    Plan status
  SEG MI    Cached total segment miles
  AT MIN    Earliest event timestamp (UTC)
  AT MAX    Latest event timestamp (UTC)
  ON MIN    Earliest event date
  ON MAX    Latest event date
  STRAT POS Strategy set prediction position
  PROJECT   Project ID
  BROKER    Broker ID

Arguments:
  <id>  Project transport plan ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan
  xbe view project-transport-plans show 123

  # Output as JSON
  xbe view project-transport-plans show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlansShow,
	}
	initProjectTransportPlansShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlansCmd.AddCommand(newProjectTransportPlansShowCmd())
}

func initProjectTransportPlansShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlansShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlansShowOptions(cmd)
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
		return fmt.Errorf("project transport plan id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plans]", "status,segment-miles,event-times-at-min,event-times-at-max,event-times-on-min,event-times-on-max,strategy-set-prediction-position,skip-actualization,assignment-rule-override-reason,skip-assignment-rules-validation,created-at,updated-at,project,broker,created-by,project-transport-plan-strategy-set,project-transport-plan-stops,project-transport-plan-segments,project-transport-plan-segment-sets,project-transport-plan-events,project-transport-plan-drivers,project-transport-plan-tractors,project-transport-plan-trailers,transport-orders,nearest-project-offices,incidents")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plans/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectTransportPlanDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanDetails(cmd, details)
}

func parseProjectTransportPlansShowOptions(cmd *cobra.Command) (projectTransportPlansShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlansShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanDetails(resp jsonAPISingleResponse) projectTransportPlanDetails {
	attrs := resp.Data.Attributes
	details := projectTransportPlanDetails{
		ID:                            resp.Data.ID,
		Status:                        stringAttr(attrs, "status"),
		SegmentMiles:                  anyAttr(attrs, "segment-miles"),
		EventTimesAtMin:               formatDateTime(stringAttr(attrs, "event-times-at-min")),
		EventTimesAtMax:               formatDateTime(stringAttr(attrs, "event-times-at-max")),
		EventTimesOnMin:               formatDate(stringAttr(attrs, "event-times-on-min")),
		EventTimesOnMax:               formatDate(stringAttr(attrs, "event-times-on-max")),
		StrategySetPredictionPosition: stringAttr(attrs, "strategy-set-prediction-position"),
		SkipActualization:             boolAttr(attrs, "skip-actualization"),
		SkipAssignmentRulesValidation: boolAttr(attrs, "skip-assignment-rules-validation"),
		AssignmentRuleOverrideReason:  stringAttr(attrs, "assignment-rule-override-reason"),
		CreatedAt:                     formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                     formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-strategy-set"]; ok && rel.Data != nil {
		details.ProjectTransportPlanStrategySetID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-stops"]; ok {
		details.ProjectTransportPlanStopIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-segments"]; ok {
		details.ProjectTransportPlanSegmentIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-segment-sets"]; ok {
		details.ProjectTransportPlanSegmentSetIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-events"]; ok {
		details.ProjectTransportPlanEventIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-drivers"]; ok {
		details.ProjectTransportPlanDriverIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-tractors"]; ok {
		details.ProjectTransportPlanTractorIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-trailers"]; ok {
		details.ProjectTransportPlanTrailerIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["transport-orders"]; ok {
		details.TransportOrderIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["nearest-project-offices"]; ok {
		details.NearestProjectOfficeIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["incidents"]; ok {
		details.IncidentIDs = relationshipIDList(rel)
	}

	return details
}

func renderProjectTransportPlanDetails(cmd *cobra.Command, details projectTransportPlanDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.SegmentMiles != nil {
		fmt.Fprintf(out, "Segment Miles: %s\n", formatDistanceMiles(details.SegmentMiles))
	}
	if details.EventTimesAtMin != "" {
		fmt.Fprintf(out, "Event Times At Min: %s\n", details.EventTimesAtMin)
	}
	if details.EventTimesAtMax != "" {
		fmt.Fprintf(out, "Event Times At Max: %s\n", details.EventTimesAtMax)
	}
	if details.EventTimesOnMin != "" {
		fmt.Fprintf(out, "Event Times On Min: %s\n", details.EventTimesOnMin)
	}
	if details.EventTimesOnMax != "" {
		fmt.Fprintf(out, "Event Times On Max: %s\n", details.EventTimesOnMax)
	}
	if details.StrategySetPredictionPosition != "" {
		fmt.Fprintf(out, "Strategy Set Prediction Position: %s\n", details.StrategySetPredictionPosition)
	}
	fmt.Fprintf(out, "Skip Actualization: %t\n", details.SkipActualization)
	fmt.Fprintf(out, "Skip Assignment Rules Validation: %t\n", details.SkipAssignmentRulesValidation)
	if details.AssignmentRuleOverrideReason != "" {
		fmt.Fprintf(out, "Assignment Rule Override Reason: %s\n", details.AssignmentRuleOverrideReason)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project: %s\n", details.ProjectID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By: %s\n", details.CreatedByID)
	}
	if details.ProjectTransportPlanStrategySetID != "" {
		fmt.Fprintf(out, "Strategy Set: %s\n", details.ProjectTransportPlanStrategySetID)
	}
	if len(details.ProjectTransportPlanStopIDs) > 0 {
		fmt.Fprintf(out, "Plan Stops: %s\n", strings.Join(details.ProjectTransportPlanStopIDs, ", "))
	}
	if len(details.ProjectTransportPlanSegmentIDs) > 0 {
		fmt.Fprintf(out, "Segments: %s\n", strings.Join(details.ProjectTransportPlanSegmentIDs, ", "))
	}
	if len(details.ProjectTransportPlanSegmentSetIDs) > 0 {
		fmt.Fprintf(out, "Segment Sets: %s\n", strings.Join(details.ProjectTransportPlanSegmentSetIDs, ", "))
	}
	if len(details.ProjectTransportPlanEventIDs) > 0 {
		fmt.Fprintf(out, "Events: %s\n", strings.Join(details.ProjectTransportPlanEventIDs, ", "))
	}
	if len(details.ProjectTransportPlanDriverIDs) > 0 {
		fmt.Fprintf(out, "Drivers: %s\n", strings.Join(details.ProjectTransportPlanDriverIDs, ", "))
	}
	if len(details.ProjectTransportPlanTractorIDs) > 0 {
		fmt.Fprintf(out, "Tractors: %s\n", strings.Join(details.ProjectTransportPlanTractorIDs, ", "))
	}
	if len(details.ProjectTransportPlanTrailerIDs) > 0 {
		fmt.Fprintf(out, "Trailers: %s\n", strings.Join(details.ProjectTransportPlanTrailerIDs, ", "))
	}
	if len(details.TransportOrderIDs) > 0 {
		fmt.Fprintf(out, "Transport Orders: %s\n", strings.Join(details.TransportOrderIDs, ", "))
	}
	if len(details.NearestProjectOfficeIDs) > 0 {
		fmt.Fprintf(out, "Nearest Project Offices: %s\n", strings.Join(details.NearestProjectOfficeIDs, ", "))
	}
	if len(details.IncidentIDs) > 0 {
		fmt.Fprintf(out, "Incidents: %s\n", strings.Join(details.IncidentIDs, ", "))
	}

	return nil
}
