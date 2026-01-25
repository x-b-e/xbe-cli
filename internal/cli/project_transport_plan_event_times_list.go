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

type projectTransportPlanEventTimesListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	ProjectTransportPlanEvent string
	ChangedBy                 string
	ProjectTransportPlan      string
	ProjectTransportPlanID    string
	AtMin                     string
	AtMax                     string
	IsAt                      string
	StartAtMin                string
	StartAtMax                string
	EndAtMin                  string
	EndAtMax                  string
	Kind                      string
}

type projectTransportPlanEventTimeRow struct {
	ID                          string `json:"id"`
	ProjectTransportPlanEventID string `json:"project_transport_plan_event_id,omitempty"`
	Kind                        string `json:"kind,omitempty"`
	StartAt                     string `json:"start_at,omitempty"`
	EndAt                       string `json:"end_at,omitempty"`
}

func newProjectTransportPlanEventTimesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan event times",
		Long: `List project transport plan event times with filtering and pagination.

Project transport plan event times capture planned, expected, actual, and modeled
timestamps for project transport plan events.

Output Columns:
  ID        Event time ID
  EVENT     Project transport plan event ID
  KIND      Time kind (planned, expected, actual, modeled)
  START     Start timestamp
  END       End timestamp

Filters:
  --project-transport-plan-event  Filter by project transport plan event ID (comma-separated for multiple)
  --changed-by                    Filter by changed-by user ID (comma-separated for multiple)
  --project-transport-plan        Filter by project transport plan ID (comma-separated for multiple)
  --project-transport-plan-id     Filter by project transport plan ID (legacy filter)
  --at-min                        Filter by minimum at timestamp (ISO 8601, legacy)
  --at-max                        Filter by maximum at timestamp (ISO 8601, legacy)
  --is-at                         Filter by presence of at timestamp (true/false, legacy)
  --start-at-min                  Filter by minimum start timestamp (ISO 8601)
  --start-at-max                  Filter by maximum start timestamp (ISO 8601)
  --end-at-min                    Filter by minimum end timestamp (ISO 8601)
  --end-at-max                    Filter by maximum end timestamp (ISO 8601)
  --kind                          Filter by time kind (planned, expected, actual, modeled)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan event times
  xbe view project-transport-plan-event-times list

  # Filter by project transport plan event
  xbe view project-transport-plan-event-times list --project-transport-plan-event 123

  # Filter by kind and time range
  xbe view project-transport-plan-event-times list --kind expected --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view project-transport-plan-event-times list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanEventTimesList,
	}
	initProjectTransportPlanEventTimesListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanEventTimesCmd.AddCommand(newProjectTransportPlanEventTimesListCmd())
}

func initProjectTransportPlanEventTimesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan-event", "", "Filter by project transport plan event ID (comma-separated for multiple)")
	cmd.Flags().String("changed-by", "", "Filter by changed-by user ID (comma-separated for multiple)")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID (comma-separated for multiple)")
	cmd.Flags().String("project-transport-plan-id", "", "Filter by project transport plan ID (legacy filter)")
	cmd.Flags().String("at-min", "", "Filter by minimum at timestamp (ISO 8601, legacy)")
	cmd.Flags().String("at-max", "", "Filter by maximum at timestamp (ISO 8601, legacy)")
	cmd.Flags().String("is-at", "", "Filter by presence of at timestamp (true/false, legacy)")
	cmd.Flags().String("start-at-min", "", "Filter by minimum start timestamp (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by maximum start timestamp (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by minimum end timestamp (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by maximum end timestamp (ISO 8601)")
	cmd.Flags().String("kind", "", "Filter by time kind (planned, expected, actual, modeled)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanEventTimesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanEventTimesListOptions(cmd)
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
	query.Set("fields[project-transport-plan-event-times]", "start-at,end-at,kind,project-transport-plan-event")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-transport-plan-event]", opts.ProjectTransportPlanEvent)
	setFilterIfPresent(query, "filter[changed-by]", opts.ChangedBy)
	setFilterIfPresent(query, "filter[project-transport-plan]", opts.ProjectTransportPlan)
	setFilterIfPresent(query, "filter[project-transport-plan-id]", opts.ProjectTransportPlanID)
	setFilterIfPresent(query, "filter[at-min]", opts.AtMin)
	setFilterIfPresent(query, "filter[at-max]", opts.AtMax)
	setFilterIfPresent(query, "filter[is-at]", opts.IsAt)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end-at-min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end-at-max]", opts.EndAtMax)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-event-times", query)
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

	rows := buildProjectTransportPlanEventTimeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanEventTimesTable(cmd, rows)
}

func parseProjectTransportPlanEventTimesListOptions(cmd *cobra.Command) (projectTransportPlanEventTimesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlanEvent, _ := cmd.Flags().GetString("project-transport-plan-event")
	changedBy, _ := cmd.Flags().GetString("changed-by")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportPlanID, _ := cmd.Flags().GetString("project-transport-plan-id")
	atMin, _ := cmd.Flags().GetString("at-min")
	atMax, _ := cmd.Flags().GetString("at-max")
	isAt, _ := cmd.Flags().GetString("is-at")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	kind, _ := cmd.Flags().GetString("kind")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanEventTimesListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		ProjectTransportPlanEvent: projectTransportPlanEvent,
		ChangedBy:                 changedBy,
		ProjectTransportPlan:      projectTransportPlan,
		ProjectTransportPlanID:    projectTransportPlanID,
		AtMin:                     atMin,
		AtMax:                     atMax,
		IsAt:                      isAt,
		StartAtMin:                startAtMin,
		StartAtMax:                startAtMax,
		EndAtMin:                  endAtMin,
		EndAtMax:                  endAtMax,
		Kind:                      kind,
	}, nil
}

func buildProjectTransportPlanEventTimeRows(resp jsonAPIResponse) []projectTransportPlanEventTimeRow {
	rows := make([]projectTransportPlanEventTimeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanEventTimeRow{
			ID:      resource.ID,
			Kind:    stringAttr(attrs, "kind"),
			StartAt: formatDateTime(stringAttr(attrs, "start-at")),
			EndAt:   formatDateTime(stringAttr(attrs, "end-at")),
		}

		row.ProjectTransportPlanEventID = relationshipIDFromMap(resource.Relationships, "project-transport-plan-event")

		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanEventTimesTable(cmd *cobra.Command, rows []projectTransportPlanEventTimeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan event times found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEVENT\tKIND\tSTART\tEND")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectTransportPlanEventID,
			row.Kind,
			row.StartAt,
			row.EndAt,
		)
	}
	return writer.Flush()
}
