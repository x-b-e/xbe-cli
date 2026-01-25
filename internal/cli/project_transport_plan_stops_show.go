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

type projectTransportPlanStopsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanStopDetails struct {
	ID                           string   `json:"id"`
	ProjectTransportPlanID       string   `json:"project_transport_plan_id,omitempty"`
	ProjectTransportLocationID   string   `json:"project_transport_location_id,omitempty"`
	PlannedCompletionEventTypeID string   `json:"planned_completion_event_type_id,omitempty"`
	CreatedByID                  string   `json:"created_by_id,omitempty"`
	OriginSegmentIDs             []string `json:"origin_segment_ids,omitempty"`
	DestinationSegmentIDs        []string `json:"destination_segment_ids,omitempty"`
	ProjectTransportPlanEventIDs []string `json:"project_transport_plan_event_ids,omitempty"`
	TransportOrderStopIDs        []string `json:"transport_order_stop_ids,omitempty"`
	Status                       string   `json:"status,omitempty"`
	Position                     string   `json:"position,omitempty"`
	ExternalTmsStopNumber        string   `json:"external_tms_stop_number,omitempty"`
}

func newProjectTransportPlanStopsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan stop details",
		Long: `Show the full details of a project transport plan stop.

Output Fields:
  ID
  Project Transport Plan ID
  Project Transport Location ID
  Planned Completion Event Type ID
  Created By ID
  Origin Segment IDs
  Destination Segment IDs
  Project Transport Plan Event IDs
  Transport Order Stop IDs
  Status
  Position
  External TMS Stop Number

Arguments:
  <id>    The project transport plan stop ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan stop
  xbe view project-transport-plan-stops show 123

  # JSON output
  xbe view project-transport-plan-stops show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanStopsShow,
	}
	initProjectTransportPlanStopsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStopsCmd.AddCommand(newProjectTransportPlanStopsShowCmd())
}

func initProjectTransportPlanStopsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStopsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanStopsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan stop id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-stops]", "external-tms-stop-number,position,status,project-transport-plan,project-transport-location,planned-completion-event-type,origin-segments,destination-segments,project-transport-plan-events,transport-order-stops,created-by")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-stops/"+id, query)
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

	details := buildProjectTransportPlanStopDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanStopDetails(cmd, details)
}

func parseProjectTransportPlanStopsShowOptions(cmd *cobra.Command) (projectTransportPlanStopsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStopsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanStopDetails(resp jsonAPISingleResponse) projectTransportPlanStopDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := projectTransportPlanStopDetails{
		ID:                    resource.ID,
		Status:                stringAttr(attrs, "status"),
		Position:              stringAttr(attrs, "position"),
		ExternalTmsStopNumber: stringAttr(attrs, "external-tms-stop-number"),
	}

	details.ProjectTransportPlanID = relationshipIDFromMap(resource.Relationships, "project-transport-plan")
	details.ProjectTransportLocationID = relationshipIDFromMap(resource.Relationships, "project-transport-location")
	details.PlannedCompletionEventTypeID = relationshipIDFromMap(resource.Relationships, "planned-completion-event-type")
	details.CreatedByID = relationshipIDFromMap(resource.Relationships, "created-by")

	if rel, ok := resource.Relationships["origin-segments"]; ok {
		details.OriginSegmentIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["destination-segments"]; ok {
		details.DestinationSegmentIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["project-transport-plan-events"]; ok {
		details.ProjectTransportPlanEventIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["transport-order-stops"]; ok {
		details.TransportOrderStopIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectTransportPlanStopDetails(cmd *cobra.Command, details projectTransportPlanStopDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.ProjectTransportLocationID != "" {
		fmt.Fprintf(out, "Project Transport Location ID: %s\n", details.ProjectTransportLocationID)
	}
	if details.PlannedCompletionEventTypeID != "" {
		fmt.Fprintf(out, "Planned Completion Event Type ID: %s\n", details.PlannedCompletionEventTypeID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if len(details.OriginSegmentIDs) > 0 {
		fmt.Fprintf(out, "Origin Segment IDs: %s\n", strings.Join(details.OriginSegmentIDs, ", "))
	}
	if len(details.DestinationSegmentIDs) > 0 {
		fmt.Fprintf(out, "Destination Segment IDs: %s\n", strings.Join(details.DestinationSegmentIDs, ", "))
	}
	if len(details.ProjectTransportPlanEventIDs) > 0 {
		fmt.Fprintf(out, "Project Transport Plan Event IDs: %s\n", strings.Join(details.ProjectTransportPlanEventIDs, ", "))
	}
	if len(details.TransportOrderStopIDs) > 0 {
		fmt.Fprintf(out, "Transport Order Stop IDs: %s\n", strings.Join(details.TransportOrderStopIDs, ", "))
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Position != "" {
		fmt.Fprintf(out, "Position: %s\n", details.Position)
	}
	if details.ExternalTmsStopNumber != "" {
		fmt.Fprintf(out, "External TMS Stop Number: %s\n", details.ExternalTmsStopNumber)
	}

	return nil
}
