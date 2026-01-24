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

type projectTransportPlanStopOrderStopsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	ProjectTransportPlanStop string
	TransportOrderStop       string
	ProjectTransportPlan     string
	TransportOrder           string
	ProjectTransportPlanID   string
	TransportOrderID         string
}

type projectTransportPlanStopOrderStopRow struct {
	ID                         string `json:"id"`
	ProjectTransportPlanStopID string `json:"project_transport_plan_stop_id,omitempty"`
	TransportOrderStopID       string `json:"transport_order_stop_id,omitempty"`
	ProjectTransportPlanID     string `json:"project_transport_plan_id,omitempty"`
	TransportOrderID           string `json:"transport_order_id,omitempty"`
}

func newProjectTransportPlanStopOrderStopsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan stop order stops",
		Long: `List project transport plan stop order stops with filtering and pagination.

Output Columns:
  ID         Link identifier
  PLAN STOP  Project transport plan stop ID
  ORDER STOP Transport order stop ID
  PLAN       Project transport plan ID
  ORDER      Transport order ID

Filters:
  --project-transport-plan-stop  Filter by project transport plan stop ID
  --transport-order-stop         Filter by transport order stop ID
  --project-transport-plan       Filter by project transport plan ID
  --transport-order              Filter by transport order ID
  --project-transport-plan-id    Filter by project transport plan ID (via plan stop join)
  --transport-order-id           Filter by transport order ID (via order stop join)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan stop order stops
  xbe view project-transport-plan-stop-order-stops list

  # Filter by plan stop
  xbe view project-transport-plan-stop-order-stops list --project-transport-plan-stop 123

  # Filter by order stop
  xbe view project-transport-plan-stop-order-stops list --transport-order-stop 456

  # Output as JSON
  xbe view project-transport-plan-stop-order-stops list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanStopOrderStopsList,
	}
	initProjectTransportPlanStopOrderStopsListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanStopOrderStopsCmd.AddCommand(newProjectTransportPlanStopOrderStopsListCmd())
}

func initProjectTransportPlanStopOrderStopsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan-stop", "", "Filter by project transport plan stop ID")
	cmd.Flags().String("transport-order-stop", "", "Filter by transport order stop ID")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("transport-order", "", "Filter by transport order ID")
	cmd.Flags().String("project-transport-plan-id", "", "Filter by project transport plan ID (via plan stop join)")
	cmd.Flags().String("transport-order-id", "", "Filter by transport order ID (via order stop join)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanStopOrderStopsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanStopOrderStopsListOptions(cmd)
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
	query.Set("fields[project-transport-plan-stop-order-stops]", "project-transport-plan-stop,transport-order-stop,project-transport-plan,transport-order")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[project-transport-plan-stop]", opts.ProjectTransportPlanStop)
	setFilterIfPresent(query, "filter[transport-order-stop]", opts.TransportOrderStop)
	setFilterIfPresent(query, "filter[project-transport-plan]", opts.ProjectTransportPlan)
	setFilterIfPresent(query, "filter[transport-order]", opts.TransportOrder)
	setFilterIfPresent(query, "filter[project-transport-plan-id]", opts.ProjectTransportPlanID)
	setFilterIfPresent(query, "filter[transport-order-id]", opts.TransportOrderID)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-stop-order-stops", query)
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

	rows := buildProjectTransportPlanStopOrderStopRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanStopOrderStopsTable(cmd, rows)
}

func parseProjectTransportPlanStopOrderStopsListOptions(cmd *cobra.Command) (projectTransportPlanStopOrderStopsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlanStop, _ := cmd.Flags().GetString("project-transport-plan-stop")
	transportOrderStop, _ := cmd.Flags().GetString("transport-order-stop")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	projectTransportPlanID, _ := cmd.Flags().GetString("project-transport-plan-id")
	transportOrderID, _ := cmd.Flags().GetString("transport-order-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanStopOrderStopsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		ProjectTransportPlanStop: projectTransportPlanStop,
		TransportOrderStop:       transportOrderStop,
		ProjectTransportPlan:     projectTransportPlan,
		TransportOrder:           transportOrder,
		ProjectTransportPlanID:   projectTransportPlanID,
		TransportOrderID:         transportOrderID,
	}, nil
}

func buildProjectTransportPlanStopOrderStopRows(resp jsonAPIResponse) []projectTransportPlanStopOrderStopRow {
	rows := make([]projectTransportPlanStopOrderStopRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectTransportPlanStopOrderStopRow(resource))
	}
	return rows
}

func projectTransportPlanStopOrderStopRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanStopOrderStopRow {
	return buildProjectTransportPlanStopOrderStopRow(resp.Data)
}

func buildProjectTransportPlanStopOrderStopRow(resource jsonAPIResource) projectTransportPlanStopOrderStopRow {
	row := projectTransportPlanStopOrderStopRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["project-transport-plan-stop"]; ok && rel.Data != nil {
		row.ProjectTransportPlanStopID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["transport-order-stop"]; ok && rel.Data != nil {
		row.TransportOrderStopID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		row.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
		row.TransportOrderID = rel.Data.ID
	}

	return row
}

func renderProjectTransportPlanStopOrderStopsTable(cmd *cobra.Command, rows []projectTransportPlanStopOrderStopRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan stop order stops found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN STOP\tORDER STOP\tPLAN\tORDER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ProjectTransportPlanStopID, 16),
			truncateString(row.TransportOrderStopID, 16),
			truncateString(row.ProjectTransportPlanID, 16),
			truncateString(row.TransportOrderID, 16),
		)
	}
	return writer.Flush()
}
