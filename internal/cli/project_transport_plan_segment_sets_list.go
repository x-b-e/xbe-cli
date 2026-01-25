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

type projectTransportPlanSegmentSetsListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Sort                 string
	ProjectTransportPlan string
	Trucker              string
	ExternalTmsLegNumber string
	SegmentMilesSum      string
}

type projectTransportPlanSegmentSetRow struct {
	ID                     string `json:"id"`
	Position               string `json:"position,omitempty"`
	ExternalTmsLegNumber   string `json:"external_tms_leg_number,omitempty"`
	SegmentMilesSum        any    `json:"segment_miles_sum,omitempty"`
	ProjectTransportPlanID string `json:"project_transport_plan_id,omitempty"`
	TruckerID              string `json:"trucker_id,omitempty"`
}

func newProjectTransportPlanSegmentSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan segment sets",
		Long: `List project transport plan segment sets with filtering and pagination.

Output Columns:
  ID        Segment set identifier
  POSITION  Sequence position within the plan
  PLAN      Project transport plan ID
  TRUCKER   Trucker ID (if assigned)
  EXT LEG   External TMS leg number
  SEG MI    Cached total segment miles

Filters:
  --project-transport-plan  Filter by project transport plan ID
  --trucker                 Filter by trucker ID
  --external-tms-leg-number Filter by external TMS leg number
  --segment-miles-sum       Filter by cached segment miles sum

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan segment sets
  xbe view project-transport-plan-segment-sets list

  # Filter by project transport plan
  xbe view project-transport-plan-segment-sets list --project-transport-plan 123

  # Filter by trucker
  xbe view project-transport-plan-segment-sets list --trucker 456

  # Output as JSON
  xbe view project-transport-plan-segment-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanSegmentSetsList,
	}
	initProjectTransportPlanSegmentSetsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentSetsCmd.AddCommand(newProjectTransportPlanSegmentSetsListCmd())
}

func initProjectTransportPlanSegmentSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("external-tms-leg-number", "", "Filter by external TMS leg number")
	cmd.Flags().String("segment-miles-sum", "", "Filter by cached segment miles sum")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanSegmentSetsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-segment-sets]", "position,external-tms-leg-number,segment-miles-sum,project-transport-plan,trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project-transport-plan]", opts.ProjectTransportPlan)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[external-tms-leg-number]", opts.ExternalTmsLegNumber)
	setFilterIfPresent(query, "filter[segment-miles-sum]", opts.SegmentMilesSum)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segment-sets", query)
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

	rows := buildProjectTransportPlanSegmentSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanSegmentSetsTable(cmd, rows)
}

func parseProjectTransportPlanSegmentSetsListOptions(cmd *cobra.Command) (projectTransportPlanSegmentSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	trucker, _ := cmd.Flags().GetString("trucker")
	externalTmsLegNumber, _ := cmd.Flags().GetString("external-tms-leg-number")
	segmentMilesSum, _ := cmd.Flags().GetString("segment-miles-sum")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentSetsListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Sort:                 sort,
		ProjectTransportPlan: projectTransportPlan,
		Trucker:              trucker,
		ExternalTmsLegNumber: externalTmsLegNumber,
		SegmentMilesSum:      segmentMilesSum,
	}, nil
}

func buildProjectTransportPlanSegmentSetRows(resp jsonAPIResponse) []projectTransportPlanSegmentSetRow {
	rows := make([]projectTransportPlanSegmentSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectTransportPlanSegmentSetRow(resource))
	}
	return rows
}

func buildProjectTransportPlanSegmentSetRow(resource jsonAPIResource) projectTransportPlanSegmentSetRow {
	attrs := resource.Attributes
	row := projectTransportPlanSegmentSetRow{
		ID:                   resource.ID,
		Position:             stringAttr(attrs, "position"),
		ExternalTmsLegNumber: stringAttr(attrs, "external-tms-leg-number"),
		SegmentMilesSum:      anyAttr(attrs, "segment-miles-sum"),
	}

	if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		row.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}

	return row
}

func renderProjectTransportPlanSegmentSetsTable(cmd *cobra.Command, rows []projectTransportPlanSegmentSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan segment sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPOSITION\tPLAN\tTRUCKER\tEXT LEG\tSEG MI")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Position,
			row.ProjectTransportPlanID,
			row.TruckerID,
			truncateString(row.ExternalTmsLegNumber, 18),
			formatDistanceMiles(row.SegmentMilesSum),
		)
	}
	return writer.Flush()
}
