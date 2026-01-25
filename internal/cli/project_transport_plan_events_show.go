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

type projectTransportPlanEventsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanEventDetails struct {
	ID                               string   `json:"id"`
	ExternalTMSEventID               string   `json:"external_tms_event_id,omitempty"`
	Position                         int      `json:"position"`
	DwellMinutesMin                  any      `json:"dwell_minutes_min,omitempty"`
	LocationPredictionPosition       any      `json:"location_prediction_position,omitempty"`
	ProjectTransportPlanID           string   `json:"project_transport_plan_id,omitempty"`
	ProjectTransportEventTypeID      string   `json:"project_transport_event_type_id,omitempty"`
	ProjectTransportEventTypeName    string   `json:"project_transport_event_type_name,omitempty"`
	ProjectTransportLocationID       string   `json:"project_transport_location_id,omitempty"`
	ProjectTransportLocationName     string   `json:"project_transport_location_name,omitempty"`
	ProjectMaterialTypeID            string   `json:"project_material_type_id,omitempty"`
	ProjectMaterialTypeName          string   `json:"project_material_type_name,omitempty"`
	ProjectTransportPlanStopID       string   `json:"project_transport_plan_stop_id,omitempty"`
	ProjectTransportPlanStopNumber   string   `json:"project_transport_plan_stop_external_tms_stop_number,omitempty"`
	ProjectTransportPlanStopStatus   string   `json:"project_transport_plan_stop_status,omitempty"`
	ProjectTransportPlanStopPosition int      `json:"project_transport_plan_stop_position,omitempty"`
	ProjectID                        string   `json:"project_id,omitempty"`
	ProjectName                      string   `json:"project_name,omitempty"`
	ProjectNumber                    string   `json:"project_number,omitempty"`
	BrokerID                         string   `json:"broker_id,omitempty"`
	BrokerName                       string   `json:"broker_name,omitempty"`
	ProjectTransportOrganizationID   string   `json:"project_transport_organization_id,omitempty"`
	ProjectTransportOrganizationName string   `json:"project_transport_organization_name,omitempty"`
	ProjectTransportPlanEventTimeIDs []string `json:"project_transport_plan_event_time_ids,omitempty"`
}

func newProjectTransportPlanEventsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan event details",
		Long: `Show the full details of a project transport plan event.

Output Fields:
  ID
  External TMS Event ID
  Position
  Dwell Minutes Min
  Location Prediction Position
  Project Transport Plan
  Project Transport Event Type
  Project Transport Location
  Project Material Type
  Project Transport Plan Stop
  Project
  Broker
  Project Transport Organization
  Project Transport Plan Event Times

Arguments:
  <id>    The project transport plan event ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project transport plan event
  xbe view project-transport-plan-events show 123

  # Output as JSON
  xbe view project-transport-plan-events show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanEventsShow,
	}
	initProjectTransportPlanEventsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanEventsCmd.AddCommand(newProjectTransportPlanEventsShowCmd())
}

func initProjectTransportPlanEventsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanEventsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportPlanEventsShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan event id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-events]", "external-tms-event-id,position,dwell-minutes-min,location-prediction-position,project-transport-plan,project-transport-event-type,project-transport-location,project-material-type,project-transport-plan-stop,project,broker,project-transport-organization,project-transport-plan-event-times")
	query.Set("fields[project-transport-event-types]", "name")
	query.Set("fields[project-transport-locations]", "name")
	query.Set("fields[project-material-types]", "display-name")
	query.Set("fields[project-transport-plan-stops]", "external-tms-stop-number,position,status")
	query.Set("fields[projects]", "name,number")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[project-transport-organizations]", "name")
	query.Set("include", "project-transport-event-type,project-transport-location,project-material-type,project-transport-plan-stop,project,broker,project-transport-organization")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-events/"+id, query)
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

	details := buildProjectTransportPlanEventDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanEventDetails(cmd, details)
}

func parseProjectTransportPlanEventsShowOptions(cmd *cobra.Command) (projectTransportPlanEventsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanEventsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanEventDetails(resp jsonAPISingleResponse) projectTransportPlanEventDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := projectTransportPlanEventDetails{
		ID:                         resource.ID,
		ExternalTMSEventID:         stringAttr(attrs, "external-tms-event-id"),
		Position:                   intAttr(attrs, "position"),
		DwellMinutesMin:            attrs["dwell-minutes-min"],
		LocationPredictionPosition: attrs["location-prediction-position"],
	}

	if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
		details.ProjectTransportEventTypeID = rel.Data.ID
		if eventType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportEventTypeName = stringAttr(eventType.Attributes, "name")
		}
	}
	if rel, ok := resource.Relationships["project-transport-location"]; ok && rel.Data != nil {
		details.ProjectTransportLocationID = rel.Data.ID
		if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportLocationName = stringAttr(location.Attributes, "name")
		}
	}
	if rel, ok := resource.Relationships["project-material-type"]; ok && rel.Data != nil {
		details.ProjectMaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectMaterialTypeName = stringAttr(materialType.Attributes, "display-name")
		}
	}
	if rel, ok := resource.Relationships["project-transport-plan-stop"]; ok && rel.Data != nil {
		details.ProjectTransportPlanStopID = rel.Data.ID
		if stop, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportPlanStopNumber = stringAttr(stop.Attributes, "external-tms-stop-number")
			details.ProjectTransportPlanStopStatus = stringAttr(stop.Attributes, "status")
			details.ProjectTransportPlanStopPosition = intAttr(stop.Attributes, "position")
		}
	}
	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = stringAttr(project.Attributes, "name")
			details.ProjectNumber = stringAttr(project.Attributes, "number")
		}
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}
	if rel, ok := resource.Relationships["project-transport-organization"]; ok && rel.Data != nil {
		details.ProjectTransportOrganizationID = rel.Data.ID
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportOrganizationName = stringAttr(org.Attributes, "name")
		}
	}
	if rel, ok := resource.Relationships["project-transport-plan-event-times"]; ok {
		details.ProjectTransportPlanEventTimeIDs = relationshipIDList(rel)
	}

	return details
}

func renderProjectTransportPlanEventDetails(cmd *cobra.Command, details projectTransportPlanEventDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ExternalTMSEventID != "" {
		fmt.Fprintf(out, "External TMS Event ID: %s\n", details.ExternalTMSEventID)
	}
	fmt.Fprintf(out, "Position: %d\n", details.Position)
	if details.DwellMinutesMin != nil {
		fmt.Fprintf(out, "Dwell Minutes Min: %v\n", details.DwellMinutesMin)
	}
	if details.LocationPredictionPosition != nil {
		fmt.Fprintf(out, "Location Prediction Position: %v\n", details.LocationPredictionPosition)
	}

	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.ProjectTransportEventTypeID != "" {
		if details.ProjectTransportEventTypeName != "" {
			fmt.Fprintf(out, "Project Transport Event Type: %s (%s)\n", details.ProjectTransportEventTypeName, details.ProjectTransportEventTypeID)
		} else {
			fmt.Fprintf(out, "Project Transport Event Type ID: %s\n", details.ProjectTransportEventTypeID)
		}
	}
	if details.ProjectTransportLocationID != "" {
		if details.ProjectTransportLocationName != "" {
			fmt.Fprintf(out, "Project Transport Location: %s (%s)\n", details.ProjectTransportLocationName, details.ProjectTransportLocationID)
		} else {
			fmt.Fprintf(out, "Project Transport Location ID: %s\n", details.ProjectTransportLocationID)
		}
	}
	if details.ProjectMaterialTypeID != "" {
		if details.ProjectMaterialTypeName != "" {
			fmt.Fprintf(out, "Project Material Type: %s (%s)\n", details.ProjectMaterialTypeName, details.ProjectMaterialTypeID)
		} else {
			fmt.Fprintf(out, "Project Material Type ID: %s\n", details.ProjectMaterialTypeID)
		}
	}
	if details.ProjectTransportPlanStopID != "" {
		label := details.ProjectTransportPlanStopID
		meta := []string{}
		if details.ProjectTransportPlanStopNumber != "" {
			meta = append(meta, "TMS stop #"+details.ProjectTransportPlanStopNumber)
		}
		if details.ProjectTransportPlanStopStatus != "" {
			meta = append(meta, "status "+details.ProjectTransportPlanStopStatus)
		}
		meta = append(meta, fmt.Sprintf("position %d", details.ProjectTransportPlanStopPosition))
		if len(meta) > 0 {
			fmt.Fprintf(out, "Project Transport Plan Stop: %s (%s)\n", label, strings.Join(meta, ", "))
		} else {
			fmt.Fprintf(out, "Project Transport Plan Stop ID: %s\n", label)
		}
	}
	if details.ProjectID != "" {
		projectLabel := firstNonEmpty(details.ProjectNumber, details.ProjectName)
		if projectLabel != "" {
			fmt.Fprintf(out, "Project: %s (%s)\n", projectLabel, details.ProjectID)
		} else {
			fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
		}
	}
	if details.BrokerID != "" {
		if details.BrokerName != "" {
			fmt.Fprintf(out, "Broker: %s (%s)\n", details.BrokerName, details.BrokerID)
		} else {
			fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
		}
	}
	if details.ProjectTransportOrganizationID != "" {
		if details.ProjectTransportOrganizationName != "" {
			fmt.Fprintf(out, "Project Transport Organization: %s (%s)\n", details.ProjectTransportOrganizationName, details.ProjectTransportOrganizationID)
		} else {
			fmt.Fprintf(out, "Project Transport Organization ID: %s\n", details.ProjectTransportOrganizationID)
		}
	}

	if len(details.ProjectTransportPlanEventTimeIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Project Transport Plan Event Times (%d):\n", len(details.ProjectTransportPlanEventTimeIDs))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, id := range details.ProjectTransportPlanEventTimeIDs {
			fmt.Fprintf(out, "  - %s\n", id)
		}
	}

	return nil
}
