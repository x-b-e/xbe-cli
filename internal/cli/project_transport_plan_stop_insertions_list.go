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

type projectTransportPlanStopInsertionsListOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	NoAuth                            bool
	Limit                             int
	Offset                            int
	Sort                              string
	ProjectTransportPlan              string
	ProjectTransportPlanID            string
	ReferenceProjectTransportPlanStop string
	ProjectTransportLocation          string
	ProjectTransportPlanSegmentSet    string
	PlannedCompletionEventType        string
	Status                            string
	Mode                              string
	ReuseHalf                         string
	BoundaryChoice                    string
	CreatedBy                         string
	ExistingProjectTransportPlanStop  string
	StopToMove                        string
}

type projectTransportPlanStopInsertionRow struct {
	ID                                string `json:"id"`
	Status                            string `json:"status,omitempty"`
	Mode                              string `json:"mode,omitempty"`
	ProjectTransportPlanID            string `json:"project_transport_plan_id,omitempty"`
	ReferenceProjectTransportPlanStop string `json:"reference_project_transport_plan_stop_id,omitempty"`
	ProjectTransportLocationID        string `json:"project_transport_location_id,omitempty"`
	ErrorMessage                      string `json:"error_message,omitempty"`
}

func newProjectTransportPlanStopInsertionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan stop insertions",
		Long: `List project transport plan stop insertions.

Output Columns:
  ID        Stop insertion identifier
  STATUS    Processing status
  MODE      Insertion mode
  PLAN      Project transport plan ID
  REF STOP  Reference project transport plan stop ID
  LOCATION  Project transport location ID
  ERROR     Error message (if present)

Filters:
  --project-transport-plan               Filter by project transport plan ID
  --project-transport-plan-id            Filter by project transport plan ID (alias)
  --reference-project-transport-plan-stop Filter by reference stop ID
  --project-transport-location           Filter by project transport location ID
  --project-transport-plan-segment-set   Filter by project transport plan segment set ID
  --planned-completion-event-type        Filter by planned completion event type ID
  --status                               Filter by status
  --mode                                 Filter by mode
  --reuse-half                           Filter by reuse half
  --boundary-choice                      Filter by boundary choice
  --created-by                           Filter by creator user ID
  --existing-project-transport-plan-stop Filter by existing stop ID
  --stop-to-move                         Filter by stop-to-move ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List stop insertions
  xbe view project-transport-plan-stop-insertions list --limit 10

  # Filter by plan
  xbe view project-transport-plan-stop-insertions list --project-transport-plan 123

  # Filter by status
  xbe view project-transport-plan-stop-insertions list --status applied

  # Output as JSON
  xbe view project-transport-plan-stop-insertions list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanStopInsertionsList,
	}
	initProjectTransportPlanStopInsertionsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStopInsertionsCmd.AddCommand(newProjectTransportPlanStopInsertionsListCmd())
}

func initProjectTransportPlanStopInsertionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("project-transport-plan-id", "", "Filter by project transport plan ID (alias)")
	cmd.Flags().String("reference-project-transport-plan-stop", "", "Filter by reference stop ID")
	cmd.Flags().String("project-transport-location", "", "Filter by project transport location ID")
	cmd.Flags().String("project-transport-plan-segment-set", "", "Filter by project transport plan segment set ID")
	cmd.Flags().String("planned-completion-event-type", "", "Filter by planned completion event type ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("mode", "", "Filter by mode")
	cmd.Flags().String("reuse-half", "", "Filter by reuse half")
	cmd.Flags().String("boundary-choice", "", "Filter by boundary choice")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("existing-project-transport-plan-stop", "", "Filter by existing stop ID")
	cmd.Flags().String("stop-to-move", "", "Filter by stop-to-move ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStopInsertionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanStopInsertionsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-stop-insertions]", "status,mode,error-message,project-transport-plan,reference-project-transport-plan-stop,project-transport-location")

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
	setFilterIfPresent(query, "filter[project-transport-plan-id]", opts.ProjectTransportPlanID)
	setFilterIfPresent(query, "filter[reference-project-transport-plan-stop]", opts.ReferenceProjectTransportPlanStop)
	setFilterIfPresent(query, "filter[project-transport-location]", opts.ProjectTransportLocation)
	setFilterIfPresent(query, "filter[project-transport-plan-segment-set]", opts.ProjectTransportPlanSegmentSet)
	setFilterIfPresent(query, "filter[planned-completion-event-type]", opts.PlannedCompletionEventType)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[mode]", opts.Mode)
	setFilterIfPresent(query, "filter[reuse-half]", opts.ReuseHalf)
	setFilterIfPresent(query, "filter[boundary-choice]", opts.BoundaryChoice)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[existing-project-transport-plan-stop]", opts.ExistingProjectTransportPlanStop)
	setFilterIfPresent(query, "filter[stop-to-move]", opts.StopToMove)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-stop-insertions", query)
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

	rows := buildProjectTransportPlanStopInsertionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanStopInsertionsTable(cmd, rows)
}

func parseProjectTransportPlanStopInsertionsListOptions(cmd *cobra.Command) (projectTransportPlanStopInsertionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportPlanID, _ := cmd.Flags().GetString("project-transport-plan-id")
	referenceProjectTransportPlanStop, _ := cmd.Flags().GetString("reference-project-transport-plan-stop")
	projectTransportLocation, _ := cmd.Flags().GetString("project-transport-location")
	projectTransportPlanSegmentSet, _ := cmd.Flags().GetString("project-transport-plan-segment-set")
	plannedCompletionEventType, _ := cmd.Flags().GetString("planned-completion-event-type")
	status, _ := cmd.Flags().GetString("status")
	mode, _ := cmd.Flags().GetString("mode")
	reuseHalf, _ := cmd.Flags().GetString("reuse-half")
	boundaryChoice, _ := cmd.Flags().GetString("boundary-choice")
	createdBy, _ := cmd.Flags().GetString("created-by")
	existingProjectTransportPlanStop, _ := cmd.Flags().GetString("existing-project-transport-plan-stop")
	stopToMove, _ := cmd.Flags().GetString("stop-to-move")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStopInsertionsListOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		NoAuth:                            noAuth,
		Limit:                             limit,
		Offset:                            offset,
		Sort:                              sort,
		ProjectTransportPlan:              projectTransportPlan,
		ProjectTransportPlanID:            projectTransportPlanID,
		ReferenceProjectTransportPlanStop: referenceProjectTransportPlanStop,
		ProjectTransportLocation:          projectTransportLocation,
		ProjectTransportPlanSegmentSet:    projectTransportPlanSegmentSet,
		PlannedCompletionEventType:        plannedCompletionEventType,
		Status:                            status,
		Mode:                              mode,
		ReuseHalf:                         reuseHalf,
		BoundaryChoice:                    boundaryChoice,
		CreatedBy:                         createdBy,
		ExistingProjectTransportPlanStop:  existingProjectTransportPlanStop,
		StopToMove:                        stopToMove,
	}, nil
}

func buildProjectTransportPlanStopInsertionRows(resp jsonAPIResponse) []projectTransportPlanStopInsertionRow {
	rows := make([]projectTransportPlanStopInsertionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectTransportPlanStopInsertionRow(resource))
	}
	return rows
}

func buildProjectTransportPlanStopInsertionRow(resource jsonAPIResource) projectTransportPlanStopInsertionRow {
	attrs := resource.Attributes
	row := projectTransportPlanStopInsertionRow{
		ID:           resource.ID,
		Status:       stringAttr(attrs, "status"),
		Mode:         stringAttr(attrs, "mode"),
		ErrorMessage: stringAttr(attrs, "error-message"),
	}

	if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		row.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["reference-project-transport-plan-stop"]; ok && rel.Data != nil {
		row.ReferenceProjectTransportPlanStop = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-transport-location"]; ok && rel.Data != nil {
		row.ProjectTransportLocationID = rel.Data.ID
	}

	return row
}

func buildProjectTransportPlanStopInsertionRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanStopInsertionRow {
	return buildProjectTransportPlanStopInsertionRow(resp.Data)
}

func renderProjectTransportPlanStopInsertionsTable(cmd *cobra.Command, rows []projectTransportPlanStopInsertionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan stop insertions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tMODE\tPLAN\tREF STOP\tLOCATION\tERROR")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Mode,
			row.ProjectTransportPlanID,
			row.ReferenceProjectTransportPlanStop,
			row.ProjectTransportLocationID,
			truncateString(row.ErrorMessage, 60),
		)
	}
	return writer.Flush()
}
