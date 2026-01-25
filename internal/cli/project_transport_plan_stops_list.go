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

type projectTransportPlanStopsListOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	NoAuth                     bool
	Limit                      int
	Offset                     int
	Sort                       string
	ProjectTransportPlan       string
	ProjectTransportLocation   string
	PlannedCompletionEventType string
	ExternalTmsStopNumber      string
}

type projectTransportPlanStopRow struct {
	ID                         string `json:"id"`
	ProjectTransportPlanID     string `json:"project_transport_plan_id,omitempty"`
	ProjectTransportLocationID string `json:"project_transport_location_id,omitempty"`
	Status                     string `json:"status,omitempty"`
	Position                   string `json:"position,omitempty"`
	ExternalTmsStopNumber      string `json:"external_tms_stop_number,omitempty"`
}

func newProjectTransportPlanStopsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan stops",
		Long: `List project transport plan stops with filtering and pagination.

Project transport plan stops define ordered locations within a transport plan
and can be associated with planned completion event types.

Output Columns:
  ID        Stop ID
  PLAN      Project transport plan ID
  LOCATION  Project transport location ID
  STATUS    Stop status (planned, started, finished, cancelled)
  POSITION  Stop position
  EXT_TMS   External TMS stop number

Filters:
  --project-transport-plan         Filter by project transport plan ID (comma-separated for multiple)
  --project-transport-location     Filter by project transport location ID (comma-separated for multiple)
  --planned-completion-event-type  Filter by planned completion event type ID (comma-separated for multiple)
  --external-tms-stop-number       Filter by external TMS stop number (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan stops
  xbe view project-transport-plan-stops list

  # Filter by project transport plan
  xbe view project-transport-plan-stops list --project-transport-plan 123

  # Filter by project transport location
  xbe view project-transport-plan-stops list --project-transport-location 456

  # Filter by external TMS stop number
  xbe view project-transport-plan-stops list --external-tms-stop-number 789

  # Output as JSON
  xbe view project-transport-plan-stops list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanStopsList,
	}
	initProjectTransportPlanStopsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStopsCmd.AddCommand(newProjectTransportPlanStopsListCmd())
}

func initProjectTransportPlanStopsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID (comma-separated for multiple)")
	cmd.Flags().String("project-transport-location", "", "Filter by project transport location ID (comma-separated for multiple)")
	cmd.Flags().String("planned-completion-event-type", "", "Filter by planned completion event type ID (comma-separated for multiple)")
	cmd.Flags().String("external-tms-stop-number", "", "Filter by external TMS stop number (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStopsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanStopsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-plan-stops]", "external-tms-stop-number,position,status,project-transport-plan,project-transport-location")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-transport-plan]", opts.ProjectTransportPlan)
	setFilterIfPresent(query, "filter[project-transport-location]", opts.ProjectTransportLocation)
	setFilterIfPresent(query, "filter[planned-completion-event-type]", opts.PlannedCompletionEventType)
	setFilterIfPresent(query, "filter[external-tms-stop-number]", opts.ExternalTmsStopNumber)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-stops", query)
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

	rows := buildProjectTransportPlanStopRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanStopsTable(cmd, rows)
}

func parseProjectTransportPlanStopsListOptions(cmd *cobra.Command) (projectTransportPlanStopsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportLocation, _ := cmd.Flags().GetString("project-transport-location")
	plannedCompletionEventType, _ := cmd.Flags().GetString("planned-completion-event-type")
	externalTmsStopNumber, _ := cmd.Flags().GetString("external-tms-stop-number")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStopsListOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		NoAuth:                     noAuth,
		Limit:                      limit,
		Offset:                     offset,
		Sort:                       sort,
		ProjectTransportPlan:       projectTransportPlan,
		ProjectTransportLocation:   projectTransportLocation,
		PlannedCompletionEventType: plannedCompletionEventType,
		ExternalTmsStopNumber:      externalTmsStopNumber,
	}, nil
}

func buildProjectTransportPlanStopRows(resp jsonAPIResponse) []projectTransportPlanStopRow {
	rows := make([]projectTransportPlanStopRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanStopRow{
			ID:                    resource.ID,
			Status:                stringAttr(attrs, "status"),
			Position:              stringAttr(attrs, "position"),
			ExternalTmsStopNumber: stringAttr(attrs, "external-tms-stop-number"),
		}

		row.ProjectTransportPlanID = relationshipIDFromMap(resource.Relationships, "project-transport-plan")
		row.ProjectTransportLocationID = relationshipIDFromMap(resource.Relationships, "project-transport-location")

		rows = append(rows, row)
	}
	return rows
}

func buildProjectTransportPlanStopRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanStopRow {
	attrs := resp.Data.Attributes
	row := projectTransportPlanStopRow{
		ID:                    resp.Data.ID,
		Status:                stringAttr(attrs, "status"),
		Position:              stringAttr(attrs, "position"),
		ExternalTmsStopNumber: stringAttr(attrs, "external-tms-stop-number"),
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		row.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-location"]; ok && rel.Data != nil {
		row.ProjectTransportLocationID = rel.Data.ID
	}

	return row
}

func renderProjectTransportPlanStopsTable(cmd *cobra.Command, rows []projectTransportPlanStopRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan stops found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tLOCATION\tSTATUS\tPOSITION\tEXT_TMS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectTransportPlanID,
			row.ProjectTransportLocationID,
			row.Status,
			row.Position,
			row.ExternalTmsStopNumber,
		)
	}
	return writer.Flush()
}
