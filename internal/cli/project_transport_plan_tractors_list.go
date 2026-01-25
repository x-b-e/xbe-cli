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

type projectTransportPlanTractorsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	ProjectTransportPlan   string
	Tractor                string
	SegmentStart           string
	SegmentEnd             string
	Status                 string
	WindowStartAtCached    string
	WindowStartAtCachedMin string
	WindowStartAtCachedMax string
	HasWindowStartAtCached string
	WindowEndAtCached      string
	WindowEndAtCachedMin   string
	WindowEndAtCachedMax   string
	HasWindowEndAtCached   string
	Actualizing            string
	MostRecent             string
}

type projectTransportPlanTractorRow struct {
	ID                   string `json:"id"`
	ProjectTransportPlan string `json:"project_transport_plan_id,omitempty"`
	Tractor              string `json:"tractor_id,omitempty"`
	SegmentStart         string `json:"segment_start_id,omitempty"`
	SegmentEnd           string `json:"segment_end_id,omitempty"`
	Status               string `json:"status,omitempty"`
	WindowStartAtCached  string `json:"window_start_at_cached,omitempty"`
	WindowEndAtCached    string `json:"window_end_at_cached,omitempty"`
}

func newProjectTransportPlanTractorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan tractors",
		Long: `List project transport plan tractors with filtering and pagination.

Project transport plan tractors assign tractors to segment ranges within
transport plans.

Output Columns:
  ID            Project transport plan tractor identifier
  STATUS        Assignment status (editing, active)
  PLAN          Project transport plan ID
  SEGMENT_START Start segment ID
  SEGMENT_END   End segment ID
  TRACTOR       Tractor ID
  WINDOW_START  Cached window start timestamp
  WINDOW_END    Cached window end timestamp

Filters:
  --project-transport-plan        Filter by project transport plan ID (comma-separated for multiple)
  --tractor                       Filter by tractor ID (comma-separated for multiple)
  --segment-start                 Filter by start segment ID (comma-separated for multiple)
  --segment-end                   Filter by end segment ID (comma-separated for multiple)
  --status                        Filter by status (editing, active)
  --window-start-at-cached        Filter by window start date (YYYY-MM-DD)
  --window-start-at-cached-min    Filter by minimum window start date (YYYY-MM-DD)
  --window-start-at-cached-max    Filter by maximum window start date (YYYY-MM-DD)
  --has-window-start-at-cached    Filter by presence of window start date (true/false)
  --window-end-at-cached          Filter by window end date (YYYY-MM-DD)
  --window-end-at-cached-min      Filter by minimum window end date (YYYY-MM-DD)
  --window-end-at-cached-max      Filter by maximum window end date (YYYY-MM-DD)
  --has-window-end-at-cached      Filter by presence of window end date (true/false)
  --actualizing                   Filter by whether actualizer window is open (true/false)
  --most-recent                   Filter to most recent assignment per tractor (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan tractors
  xbe view project-transport-plan-tractors list

  # Filter by project transport plan
  xbe view project-transport-plan-tractors list --project-transport-plan 123

  # Filter by status and most recent assignments
  xbe view project-transport-plan-tractors list --status active --most-recent true

  # Output as JSON
  xbe view project-transport-plan-tractors list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanTractorsList,
	}
	initProjectTransportPlanTractorsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanTractorsCmd.AddCommand(newProjectTransportPlanTractorsListCmd())
}

func initProjectTransportPlanTractorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID (comma-separated for multiple)")
	cmd.Flags().String("tractor", "", "Filter by tractor ID (comma-separated for multiple)")
	cmd.Flags().String("segment-start", "", "Filter by start segment ID (comma-separated for multiple)")
	cmd.Flags().String("segment-end", "", "Filter by end segment ID (comma-separated for multiple)")
	cmd.Flags().String("status", "", "Filter by status (editing, active)")
	cmd.Flags().String("window-start-at-cached", "", "Filter by window start date (YYYY-MM-DD)")
	cmd.Flags().String("window-start-at-cached-min", "", "Filter by minimum window start date (YYYY-MM-DD)")
	cmd.Flags().String("window-start-at-cached-max", "", "Filter by maximum window start date (YYYY-MM-DD)")
	cmd.Flags().String("has-window-start-at-cached", "", "Filter by presence of window start date (true/false)")
	cmd.Flags().String("window-end-at-cached", "", "Filter by window end date (YYYY-MM-DD)")
	cmd.Flags().String("window-end-at-cached-min", "", "Filter by minimum window end date (YYYY-MM-DD)")
	cmd.Flags().String("window-end-at-cached-max", "", "Filter by maximum window end date (YYYY-MM-DD)")
	cmd.Flags().String("has-window-end-at-cached", "", "Filter by presence of window end date (true/false)")
	cmd.Flags().String("actualizing", "", "Filter by whether actualizer window is open (true/false)")
	cmd.Flags().String("most-recent", "", "Filter to most recent assignment per tractor (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanTractorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanTractorsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-tractors]", "status,automatically-adjust-overlapping-windows,window-start-at-cached,window-end-at-cached,project-transport-plan,segment-start,segment-end,tractor")
	query.Set("include", "project-transport-plan,segment-start,segment-end,tractor")

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
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[segment-start]", opts.SegmentStart)
	setFilterIfPresent(query, "filter[segment-end]", opts.SegmentEnd)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[window-start-at-cached]", opts.WindowStartAtCached)
	setFilterIfPresent(query, "filter[window-start-at-cached-min]", opts.WindowStartAtCachedMin)
	setFilterIfPresent(query, "filter[window-start-at-cached-max]", opts.WindowStartAtCachedMax)
	setFilterIfPresent(query, "filter[has-window-start-at-cached]", opts.HasWindowStartAtCached)
	setFilterIfPresent(query, "filter[window-end-at-cached]", opts.WindowEndAtCached)
	setFilterIfPresent(query, "filter[window-end-at-cached-min]", opts.WindowEndAtCachedMin)
	setFilterIfPresent(query, "filter[window-end-at-cached-max]", opts.WindowEndAtCachedMax)
	setFilterIfPresent(query, "filter[has-window-end-at-cached]", opts.HasWindowEndAtCached)
	setFilterIfPresent(query, "filter[actualizing]", opts.Actualizing)
	setFilterIfPresent(query, "filter[most-recent]", opts.MostRecent)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-tractors", query)
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

	rows := buildProjectTransportPlanTractorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanTractorsTable(cmd, rows)
}

func parseProjectTransportPlanTractorsListOptions(cmd *cobra.Command) (projectTransportPlanTractorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	tractor, _ := cmd.Flags().GetString("tractor")
	segmentStart, _ := cmd.Flags().GetString("segment-start")
	segmentEnd, _ := cmd.Flags().GetString("segment-end")
	status, _ := cmd.Flags().GetString("status")
	windowStartAtCached, _ := cmd.Flags().GetString("window-start-at-cached")
	windowStartAtCachedMin, _ := cmd.Flags().GetString("window-start-at-cached-min")
	windowStartAtCachedMax, _ := cmd.Flags().GetString("window-start-at-cached-max")
	hasWindowStartAtCached, _ := cmd.Flags().GetString("has-window-start-at-cached")
	windowEndAtCached, _ := cmd.Flags().GetString("window-end-at-cached")
	windowEndAtCachedMin, _ := cmd.Flags().GetString("window-end-at-cached-min")
	windowEndAtCachedMax, _ := cmd.Flags().GetString("window-end-at-cached-max")
	hasWindowEndAtCached, _ := cmd.Flags().GetString("has-window-end-at-cached")
	actualizing, _ := cmd.Flags().GetString("actualizing")
	mostRecent, _ := cmd.Flags().GetString("most-recent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanTractorsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		ProjectTransportPlan:   projectTransportPlan,
		Tractor:                tractor,
		SegmentStart:           segmentStart,
		SegmentEnd:             segmentEnd,
		Status:                 status,
		WindowStartAtCached:    windowStartAtCached,
		WindowStartAtCachedMin: windowStartAtCachedMin,
		WindowStartAtCachedMax: windowStartAtCachedMax,
		HasWindowStartAtCached: hasWindowStartAtCached,
		WindowEndAtCached:      windowEndAtCached,
		WindowEndAtCachedMin:   windowEndAtCachedMin,
		WindowEndAtCachedMax:   windowEndAtCachedMax,
		HasWindowEndAtCached:   hasWindowEndAtCached,
		Actualizing:            actualizing,
		MostRecent:             mostRecent,
	}, nil
}

func buildProjectTransportPlanTractorRows(resp jsonAPIResponse) []projectTransportPlanTractorRow {
	rows := make([]projectTransportPlanTractorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanTractorRow{
			ID:                   resource.ID,
			ProjectTransportPlan: relationshipIDFromMap(resource.Relationships, "project-transport-plan"),
			Tractor:              relationshipIDFromMap(resource.Relationships, "tractor"),
			SegmentStart:         relationshipIDFromMap(resource.Relationships, "segment-start"),
			SegmentEnd:           relationshipIDFromMap(resource.Relationships, "segment-end"),
			Status:               stringAttr(attrs, "status"),
			WindowStartAtCached:  formatDateTime(stringAttr(attrs, "window-start-at-cached")),
			WindowEndAtCached:    formatDateTime(stringAttr(attrs, "window-end-at-cached")),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanTractorsTable(cmd *cobra.Command, rows []projectTransportPlanTractorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan tractors found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tPLAN\tSEGMENT_START\tSEGMENT_END\tTRACTOR\tWINDOW_START\tWINDOW_END")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.ProjectTransportPlan,
			row.SegmentStart,
			row.SegmentEnd,
			row.Tractor,
			row.WindowStartAtCached,
			row.WindowEndAtCached,
		)
	}
	return writer.Flush()
}
