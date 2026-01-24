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

type projectTransportPlanTrailersListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	ProjectTransportPlan   string
	Trailer                string
	SegmentStart           string
	SegmentEnd             string
	Status                 string
	WindowStartAtCachedMin string
	WindowStartAtCachedMax string
	WindowEndAtCachedMin   string
	WindowEndAtCachedMax   string
	MostRecent             string
}

type projectTransportPlanTrailerRow struct {
	ID                     string `json:"id"`
	ProjectTransportPlanID string `json:"project_transport_plan_id,omitempty"`
	TrailerID              string `json:"trailer_id,omitempty"`
	SegmentStartID         string `json:"segment_start_id,omitempty"`
	SegmentEndID           string `json:"segment_end_id,omitempty"`
	Status                 string `json:"status,omitempty"`
	WindowStartAtCached    string `json:"window_start_at_cached,omitempty"`
	WindowEndAtCached      string `json:"window_end_at_cached,omitempty"`
}

func newProjectTransportPlanTrailersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan trailers",
		Long: `List project transport plan trailers.

Output Columns:
  ID           Trailer assignment identifier
  PLAN         Project transport plan ID
  TRAILER      Trailer ID
  SEG_START    Segment start ID
  SEG_END      Segment end ID
  STATUS       Assignment status (editing, active)
  WINDOW_START Cached window start time
  WINDOW_END   Cached window end time

Filters:
  --project-transport-plan         Filter by project transport plan ID
  --trailer                        Filter by trailer ID
  --segment-start                  Filter by segment start ID
  --segment-end                    Filter by segment end ID
  --status                         Filter by status (editing, active)
  --window-start-at-cached-min     Filter by minimum cached window start date (YYYY-MM-DD)
  --window-start-at-cached-max     Filter by maximum cached window start date (YYYY-MM-DD)
  --window-end-at-cached-min       Filter by minimum cached window end date (YYYY-MM-DD)
  --window-end-at-cached-max       Filter by maximum cached window end date (YYYY-MM-DD)
  --most-recent                    Filter to most recent assignment per trailer (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List trailer assignments
  xbe view project-transport-plan-trailers list

  # Filter by project transport plan
  xbe view project-transport-plan-trailers list --project-transport-plan 123

  # Filter by status
  xbe view project-transport-plan-trailers list --status active

  # Only most recent assignments
  xbe view project-transport-plan-trailers list --most-recent true

  # JSON output
  xbe view project-transport-plan-trailers list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanTrailersList,
	}
	initProjectTransportPlanTrailersListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanTrailersCmd.AddCommand(newProjectTransportPlanTrailersListCmd())
}

func initProjectTransportPlanTrailersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("segment-start", "", "Filter by segment start ID")
	cmd.Flags().String("segment-end", "", "Filter by segment end ID")
	cmd.Flags().String("status", "", "Filter by status (editing, active)")
	cmd.Flags().String("window-start-at-cached-min", "", "Filter by minimum cached window start date (YYYY-MM-DD)")
	cmd.Flags().String("window-start-at-cached-max", "", "Filter by maximum cached window start date (YYYY-MM-DD)")
	cmd.Flags().String("window-end-at-cached-min", "", "Filter by minimum cached window end date (YYYY-MM-DD)")
	cmd.Flags().String("window-end-at-cached-max", "", "Filter by maximum cached window end date (YYYY-MM-DD)")
	cmd.Flags().String("most-recent", "", "Filter to most recent assignment per trailer (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanTrailersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanTrailersListOptions(cmd)
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
	query.Set("fields[project-transport-plan-trailers]", strings.Join([]string{
		"status",
		"window-start-at-cached",
		"window-end-at-cached",
		"project-transport-plan",
		"segment-start",
		"segment-end",
		"trailer",
	}, ","))

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
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[segment-start]", opts.SegmentStart)
	setFilterIfPresent(query, "filter[segment-end]", opts.SegmentEnd)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[window_start_at_cached_min]", opts.WindowStartAtCachedMin)
	setFilterIfPresent(query, "filter[window_start_at_cached_max]", opts.WindowStartAtCachedMax)
	setFilterIfPresent(query, "filter[window_end_at_cached_min]", opts.WindowEndAtCachedMin)
	setFilterIfPresent(query, "filter[window_end_at_cached_max]", opts.WindowEndAtCachedMax)
	setFilterIfPresent(query, "filter[most_recent]", opts.MostRecent)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-trailers", query)
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

	rows := buildProjectTransportPlanTrailerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanTrailersTable(cmd, rows)
}

func parseProjectTransportPlanTrailersListOptions(cmd *cobra.Command) (projectTransportPlanTrailersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	trailer, _ := cmd.Flags().GetString("trailer")
	segmentStart, _ := cmd.Flags().GetString("segment-start")
	segmentEnd, _ := cmd.Flags().GetString("segment-end")
	status, _ := cmd.Flags().GetString("status")
	windowStartAtCachedMin, _ := cmd.Flags().GetString("window-start-at-cached-min")
	windowStartAtCachedMax, _ := cmd.Flags().GetString("window-start-at-cached-max")
	windowEndAtCachedMin, _ := cmd.Flags().GetString("window-end-at-cached-min")
	windowEndAtCachedMax, _ := cmd.Flags().GetString("window-end-at-cached-max")
	mostRecent, _ := cmd.Flags().GetString("most-recent")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanTrailersListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		ProjectTransportPlan:   projectTransportPlan,
		Trailer:                trailer,
		SegmentStart:           segmentStart,
		SegmentEnd:             segmentEnd,
		Status:                 status,
		WindowStartAtCachedMin: windowStartAtCachedMin,
		WindowStartAtCachedMax: windowStartAtCachedMax,
		WindowEndAtCachedMin:   windowEndAtCachedMin,
		WindowEndAtCachedMax:   windowEndAtCachedMax,
		MostRecent:             mostRecent,
	}, nil
}

func buildProjectTransportPlanTrailerRows(resp jsonAPIResponse) []projectTransportPlanTrailerRow {
	rows := make([]projectTransportPlanTrailerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanTrailerRow{
			ID:                  resource.ID,
			Status:              stringAttr(attrs, "status"),
			WindowStartAtCached: formatDateTime(stringAttr(attrs, "window-start-at-cached")),
			WindowEndAtCached:   formatDateTime(stringAttr(attrs, "window-end-at-cached")),
		}

		if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
			row.ProjectTransportPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["segment-start"]; ok && rel.Data != nil {
			row.SegmentStartID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["segment-end"]; ok && rel.Data != nil {
			row.SegmentEndID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
			row.TrailerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanTrailersTable(cmd *cobra.Command, rows []projectTransportPlanTrailerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan trailers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tTRAILER\tSEG_START\tSEG_END\tSTATUS\tWINDOW_START\tWINDOW_END")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectTransportPlanID,
			row.TrailerID,
			row.SegmentStartID,
			row.SegmentEndID,
			row.Status,
			row.WindowStartAtCached,
			row.WindowEndAtCached,
		)
	}
	return writer.Flush()
}
