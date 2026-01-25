package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectTransportPlanEventsListOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	NoAuth                         bool
	Limit                          int
	Offset                         int
	Sort                           string
	ProjectTransportPlan           string
	ProjectTransportEventType      string
	ProjectTransportLocation       string
	ProjectTransportPlanStop       string
	ExternalTMSEventID             string
	ProjectTransportOrganization   string
	ProjectTransportOrganizationID string
	ProjectMaterialType            string
}

type projectTransportPlanEventRow struct {
	ID                             string `json:"id"`
	Position                       int    `json:"position"`
	ExternalTMSEventID             string `json:"external_tms_event_id,omitempty"`
	ProjectTransportPlanID         string `json:"project_transport_plan_id,omitempty"`
	ProjectTransportEventTypeID    string `json:"project_transport_event_type_id,omitempty"`
	ProjectTransportEventTypeName  string `json:"project_transport_event_type_name,omitempty"`
	ProjectTransportLocationID     string `json:"project_transport_location_id,omitempty"`
	ProjectTransportLocationName   string `json:"project_transport_location_name,omitempty"`
	ProjectTransportPlanStopID     string `json:"project_transport_plan_stop_id,omitempty"`
	ProjectTransportOrganizationID string `json:"project_transport_organization_id,omitempty"`
	ProjectMaterialTypeID          string `json:"project_material_type_id,omitempty"`
}

func newProjectTransportPlanEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan events",
		Long: `List project transport plan events with filtering and pagination.

Output Columns:
  ID                         Project transport plan event identifier
  POSITION                   Event position within the plan
  EVENT TYPE                 Project transport event type
  LOCATION                   Project transport location
  PLAN                       Project transport plan ID
  STOP                       Project transport plan stop ID
  EXTERNAL TMS EVENT ID       External TMS event identifier

Filters:
  --project-transport-plan          Filter by project transport plan ID
  --project-transport-event-type    Filter by project transport event type ID
  --project-transport-location      Filter by project transport location ID
  --project-transport-plan-stop     Filter by project transport plan stop ID
  --external-tms-event-id           Filter by external TMS event ID
  --project-transport-organization-id Filter by project transport organization ID (via location)
  --project-transport-organization  Filter by project transport organization ID
  --project-material-type           Filter by project material type ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan events
  xbe view project-transport-plan-events list

  # Filter by plan
  xbe view project-transport-plan-events list --project-transport-plan 123

  # Filter by event type
  xbe view project-transport-plan-events list --project-transport-event-type 456

  # Filter by project transport organization
  xbe view project-transport-plan-events list --project-transport-organization 789

  # Output as JSON
  xbe view project-transport-plan-events list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanEventsList,
	}
	initProjectTransportPlanEventsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanEventsCmd.AddCommand(newProjectTransportPlanEventsListCmd())
}

func initProjectTransportPlanEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("project-transport-event-type", "", "Filter by project transport event type ID")
	cmd.Flags().String("project-transport-location", "", "Filter by project transport location ID")
	cmd.Flags().String("project-transport-plan-stop", "", "Filter by project transport plan stop ID")
	cmd.Flags().String("external-tms-event-id", "", "Filter by external TMS event ID")
	cmd.Flags().String("project-transport-organization-id", "", "Filter by project transport organization ID (via location)")
	cmd.Flags().String("project-transport-organization", "", "Filter by project transport organization ID")
	cmd.Flags().String("project-material-type", "", "Filter by project material type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanEventsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-events]", "external-tms-event-id,position,project-transport-plan,project-transport-event-type,project-transport-location,project-transport-plan-stop,project-material-type,project-transport-organization")
	query.Set("fields[project-transport-event-types]", "name")
	query.Set("fields[project-transport-locations]", "name")
	query.Set("include", "project-transport-event-type,project-transport-location")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[project-transport-plan]", opts.ProjectTransportPlan)
	setFilterIfPresent(query, "filter[project-transport-event-type]", opts.ProjectTransportEventType)
	setFilterIfPresent(query, "filter[project-transport-location]", opts.ProjectTransportLocation)
	setFilterIfPresent(query, "filter[project-transport-plan-stop]", opts.ProjectTransportPlanStop)
	setFilterIfPresent(query, "filter[external-tms-event-id]", opts.ExternalTMSEventID)
	setFilterIfPresent(query, "filter[project-transport-organization-id]", opts.ProjectTransportOrganizationID)
	setFilterIfPresent(query, "filter[project-transport-organization]", opts.ProjectTransportOrganization)
	setFilterIfPresent(query, "filter[project-material-type]", opts.ProjectMaterialType)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-events", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildProjectTransportPlanEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanEventsTable(cmd, rows)
}

func parseProjectTransportPlanEventsListOptions(cmd *cobra.Command) (projectTransportPlanEventsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportEventType, _ := cmd.Flags().GetString("project-transport-event-type")
	projectTransportLocation, _ := cmd.Flags().GetString("project-transport-location")
	projectTransportPlanStop, _ := cmd.Flags().GetString("project-transport-plan-stop")
	externalTMSEventID, _ := cmd.Flags().GetString("external-tms-event-id")
	projectTransportOrganizationID, _ := cmd.Flags().GetString("project-transport-organization-id")
	projectTransportOrganization, _ := cmd.Flags().GetString("project-transport-organization")
	projectMaterialType, _ := cmd.Flags().GetString("project-material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanEventsListOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		NoAuth:                         noAuth,
		Limit:                          limit,
		Offset:                         offset,
		Sort:                           sort,
		ProjectTransportPlan:           projectTransportPlan,
		ProjectTransportEventType:      projectTransportEventType,
		ProjectTransportLocation:       projectTransportLocation,
		ProjectTransportPlanStop:       projectTransportPlanStop,
		ExternalTMSEventID:             externalTMSEventID,
		ProjectTransportOrganizationID: projectTransportOrganizationID,
		ProjectTransportOrganization:   projectTransportOrganization,
		ProjectMaterialType:            projectMaterialType,
	}, nil
}

func buildProjectTransportPlanEventRows(resp jsonAPIResponse) []projectTransportPlanEventRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]projectTransportPlanEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportPlanEventRow{
			ID:                 resource.ID,
			Position:           intAttr(resource.Attributes, "position"),
			ExternalTMSEventID: stringAttr(resource.Attributes, "external-tms-event-id"),
		}

		if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
			row.ProjectTransportPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-transport-event-type"]; ok && rel.Data != nil {
			row.ProjectTransportEventTypeID = rel.Data.ID
			if eventType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectTransportEventTypeName = stringAttr(eventType.Attributes, "name")
			}
		}
		if rel, ok := resource.Relationships["project-transport-location"]; ok && rel.Data != nil {
			row.ProjectTransportLocationID = rel.Data.ID
			if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectTransportLocationName = stringAttr(location.Attributes, "name")
			}
		}
		if rel, ok := resource.Relationships["project-transport-plan-stop"]; ok && rel.Data != nil {
			row.ProjectTransportPlanStopID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-transport-organization"]; ok && rel.Data != nil {
			row.ProjectTransportOrganizationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-material-type"]; ok && rel.Data != nil {
			row.ProjectMaterialTypeID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanEventsTable(cmd *cobra.Command, rows []projectTransportPlanEventRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan events found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPOSITION\tEVENT TYPE\tLOCATION\tPLAN\tSTOP\tEXTERNAL TMS EVENT ID")
	for _, row := range rows {
		eventTypeLabel := firstNonEmpty(row.ProjectTransportEventTypeName, row.ProjectTransportEventTypeID)
		locationLabel := firstNonEmpty(row.ProjectTransportLocationName, row.ProjectTransportLocationID)
		fmt.Fprintf(writer, "%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Position,
			truncateString(eventTypeLabel, 24),
			truncateString(locationLabel, 24),
			truncateString(row.ProjectTransportPlanID, 12),
			truncateString(row.ProjectTransportPlanStopID, 12),
			truncateString(row.ExternalTMSEventID, 20),
		)
	}
	return writer.Flush()
}
