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

type projectTransportPlanSegmentTractorsListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	ProjectTransportPlanSegment string
	Tractor                     string
}

type projectTransportPlanSegmentTractorRow struct {
	ID                        string `json:"id"`
	ProjectTransportPlanSegID string `json:"project_transport_plan_segment_id,omitempty"`
	TractorID                 string `json:"tractor_id,omitempty"`
	ActualMilesCached         string `json:"actual_miles_cached,omitempty"`
	ActualMilesSource         string `json:"actual_miles_source,omitempty"`
	ActualMilesComputedAt     string `json:"actual_miles_computed_at,omitempty"`
}

func newProjectTransportPlanSegmentTractorsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan segment tractors",
		Long: `List project transport plan segment tractors with filtering and pagination.

Project transport plan segment tractors link tractors to individual transport
plan segments and track cached actual miles.

Output Columns:
  ID            Segment tractor identifier
  SEGMENT       Project transport plan segment ID
  TRACTOR       Tractor ID
  ACTUAL MILES  Cached actual miles
  MILES SOURCE  Actual miles source
  COMPUTED AT   Actual miles computed at timestamp

Filters:
  --project-transport-plan-segment  Filter by project transport plan segment ID (comma-separated for multiple)
  --tractor                         Filter by tractor ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan segment tractors
  xbe view project-transport-plan-segment-tractors list

  # Filter by segment
  xbe view project-transport-plan-segment-tractors list --project-transport-plan-segment 123

  # Filter by tractor
  xbe view project-transport-plan-segment-tractors list --tractor 456

  # Output as JSON
  xbe view project-transport-plan-segment-tractors list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanSegmentTractorsList,
	}
	initProjectTransportPlanSegmentTractorsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentTractorsCmd.AddCommand(newProjectTransportPlanSegmentTractorsListCmd())
}

func initProjectTransportPlanSegmentTractorsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan-segment", "", "Filter by project transport plan segment ID (comma-separated for multiple)")
	cmd.Flags().String("tractor", "", "Filter by tractor ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentTractorsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanSegmentTractorsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-segment-tractors]", "actual-miles-cached,actual-miles-source,actual-miles-computed-at,project-transport-plan-segment,tractor")
	query.Set("include", "project-transport-plan-segment,tractor")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-transport-plan-segment]", opts.ProjectTransportPlanSegment)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segment-tractors", query)
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

	rows := buildProjectTransportPlanSegmentTractorRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanSegmentTractorsTable(cmd, rows)
}

func parseProjectTransportPlanSegmentTractorsListOptions(cmd *cobra.Command) (projectTransportPlanSegmentTractorsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	segment, _ := cmd.Flags().GetString("project-transport-plan-segment")
	tractor, _ := cmd.Flags().GetString("tractor")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentTractorsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		ProjectTransportPlanSegment: segment,
		Tractor:                     tractor,
	}, nil
}

func buildProjectTransportPlanSegmentTractorRows(resp jsonAPIResponse) []projectTransportPlanSegmentTractorRow {
	rows := make([]projectTransportPlanSegmentTractorRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanSegmentTractorRow{
			ID:                        resource.ID,
			ProjectTransportPlanSegID: relationshipIDFromMap(resource.Relationships, "project-transport-plan-segment"),
			TractorID:                 relationshipIDFromMap(resource.Relationships, "tractor"),
			ActualMilesCached:         stringAttr(attrs, "actual-miles-cached"),
			ActualMilesSource:         stringAttr(attrs, "actual-miles-source"),
			ActualMilesComputedAt:     formatDateTime(stringAttr(attrs, "actual-miles-computed-at")),
		}
		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanSegmentTractorsTable(cmd *cobra.Command, rows []projectTransportPlanSegmentTractorRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan segment tractors found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSEGMENT\tTRACTOR\tACTUAL_MILES\tMILES_SOURCE\tCOMPUTED_AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectTransportPlanSegID,
			row.TractorID,
			row.ActualMilesCached,
			row.ActualMilesSource,
			row.ActualMilesComputedAt,
		)
	}
	return writer.Flush()
}
