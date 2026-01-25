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

type projectTransportPlanSegmentsListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	ProjectTransportPlan        string
	Origin                      string
	Destination                 string
	ProjectTransportPlanSet     string
	Trucker                     string
	ExternalTmsOrderNumber      string
	ExternalTmsMovementNumber   string
	ExternalIdentificationValue string
}

type projectTransportPlanSegmentRow struct {
	ID                               string  `json:"id"`
	ProjectTransportPlanID           string  `json:"project_transport_plan_id,omitempty"`
	OriginID                         string  `json:"origin_id,omitempty"`
	DestinationID                    string  `json:"destination_id,omitempty"`
	ProjectTransportPlanSegmentSetID string  `json:"project_transport_plan_segment_set_id,omitempty"`
	TruckerID                        string  `json:"trucker_id,omitempty"`
	Position                         int     `json:"position,omitempty"`
	Miles                            float64 `json:"miles,omitempty"`
	MilesSource                      string  `json:"miles_source,omitempty"`
	ExternalTmsOrderNumber           string  `json:"external_tms_order_number,omitempty"`
	ExternalTmsMovementNumber        string  `json:"external_tms_movement_number,omitempty"`
}

func newProjectTransportPlanSegmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan segments",
		Long: `List project transport plan segments.

Output Columns:
  ID          Segment identifier
  PLAN        Project transport plan ID
  ORIGIN      Origin stop ID
  DEST        Destination stop ID
  POSITION    Segment position within plan
  MILES       Segment distance in miles
  SOURCE      Miles source (unknown, transport_route)
  SET         Project transport plan segment set ID
  TRUCKER     Trucker ID
  TMS_ORDER   External TMS order number
  TMS_MOVE    External TMS movement number

Filters:
  --project-transport-plan         Filter by project transport plan ID
  --origin                         Filter by origin stop ID
  --destination                    Filter by destination stop ID
  --project-transport-plan-segment-set  Filter by segment set ID
  --trucker                        Filter by trucker ID
  --external-tms-order-number      Filter by external TMS order number
  --external-tms-movement-number   Filter by external TMS movement number
  --external-identification-value  Filter by external identification value

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List segments
  xbe view project-transport-plan-segments list

  # Filter by plan
  xbe view project-transport-plan-segments list --project-transport-plan 123

  # Filter by origin stop
  xbe view project-transport-plan-segments list --origin 456

  # JSON output
  xbe view project-transport-plan-segments list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanSegmentsList,
	}
	initProjectTransportPlanSegmentsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanSegmentsCmd.AddCommand(newProjectTransportPlanSegmentsListCmd())
}

func initProjectTransportPlanSegmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("origin", "", "Filter by origin stop ID")
	cmd.Flags().String("destination", "", "Filter by destination stop ID")
	cmd.Flags().String("project-transport-plan-segment-set", "", "Filter by segment set ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("external-tms-order-number", "", "Filter by external TMS order number")
	cmd.Flags().String("external-tms-movement-number", "", "Filter by external TMS movement number")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanSegmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanSegmentsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-segments]", strings.Join([]string{
		"position",
		"miles",
		"miles-source",
		"external-tms-order-number",
		"external-tms-movement-number",
		"project-transport-plan",
		"origin",
		"destination",
		"project-transport-plan-segment-set",
		"trucker",
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
	setFilterIfPresent(query, "filter[origin]", opts.Origin)
	setFilterIfPresent(query, "filter[destination]", opts.Destination)
	setFilterIfPresent(query, "filter[project-transport-plan-segment-set]", opts.ProjectTransportPlanSet)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[external-tms-order-number]", opts.ExternalTmsOrderNumber)
	setFilterIfPresent(query, "filter[external-tms-movement-number]", opts.ExternalTmsMovementNumber)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationValue)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-segments", query)
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

	rows := buildProjectTransportPlanSegmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanSegmentsTable(cmd, rows)
}

func parseProjectTransportPlanSegmentsListOptions(cmd *cobra.Command) (projectTransportPlanSegmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	origin, _ := cmd.Flags().GetString("origin")
	destination, _ := cmd.Flags().GetString("destination")
	segmentSet, _ := cmd.Flags().GetString("project-transport-plan-segment-set")
	trucker, _ := cmd.Flags().GetString("trucker")
	externalTmsOrderNumber, _ := cmd.Flags().GetString("external-tms-order-number")
	externalTmsMovementNumber, _ := cmd.Flags().GetString("external-tms-movement-number")
	externalIdentificationValue, _ := cmd.Flags().GetString("external-identification-value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanSegmentsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		ProjectTransportPlan:        projectTransportPlan,
		Origin:                      origin,
		Destination:                 destination,
		ProjectTransportPlanSet:     segmentSet,
		Trucker:                     trucker,
		ExternalTmsOrderNumber:      externalTmsOrderNumber,
		ExternalTmsMovementNumber:   externalTmsMovementNumber,
		ExternalIdentificationValue: externalIdentificationValue,
	}, nil
}

func buildProjectTransportPlanSegmentRows(resp jsonAPIResponse) []projectTransportPlanSegmentRow {
	rows := make([]projectTransportPlanSegmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := projectTransportPlanSegmentRow{
			ID:                        resource.ID,
			Position:                  intAttr(attrs, "position"),
			Miles:                     floatAttr(attrs, "miles"),
			MilesSource:               strings.TrimSpace(stringAttr(attrs, "miles-source")),
			ExternalTmsOrderNumber:    strings.TrimSpace(stringAttr(attrs, "external-tms-order-number")),
			ExternalTmsMovementNumber: strings.TrimSpace(stringAttr(attrs, "external-tms-movement-number")),
		}

		if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
			row.ProjectTransportPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["origin"]; ok && rel.Data != nil {
			row.OriginID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["destination"]; ok && rel.Data != nil {
			row.DestinationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-transport-plan-segment-set"]; ok && rel.Data != nil {
			row.ProjectTransportPlanSegmentSetID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
			row.TruckerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanSegmentsTable(cmd *cobra.Command, rows []projectTransportPlanSegmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan segments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tORIGIN\tDEST\tPOSITION\tMILES\tSOURCE\tSET\tTRUCKER\tTMS_ORDER\tTMS_MOVE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectTransportPlanID,
			row.OriginID,
			row.DestinationID,
			formatOptionalInt(row.Position),
			formatMiles(row.Miles),
			row.MilesSource,
			row.ProjectTransportPlanSegmentSetID,
			row.TruckerID,
			row.ExternalTmsOrderNumber,
			row.ExternalTmsMovementNumber,
		)
	}
	return writer.Flush()
}
