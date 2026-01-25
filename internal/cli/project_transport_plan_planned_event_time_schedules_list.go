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

type projectTransportPlanPlannedEventTimeSchedulesListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Sort                 string
	ProjectTransportPlan string
	TransportOrder       string
	Success              string
}

type projectTransportPlanPlannedEventTimeScheduleRow struct {
	ID                     string `json:"id"`
	ProjectTransportPlanID string `json:"project_transport_plan_id,omitempty"`
	TransportOrderID       string `json:"transport_order_id,omitempty"`
	Success                bool   `json:"success"`
}

func newProjectTransportPlanPlannedEventTimeSchedulesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan planned event time schedules",
		Long: `List project transport plan planned event time schedules.

Output Columns:
  ID       Schedule identifier
  PLAN     Project transport plan ID
  ORDER    Transport order ID
  SUCCESS  Whether schedule generation succeeded

Filters:
  --project-transport-plan  Filter by project transport plan ID
  --transport-order         Filter by transport order ID
  --success                 Filter by success flag (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List schedules
  xbe view project-transport-plan-planned-event-time-schedules list

  # Filter by project transport plan
  xbe view project-transport-plan-planned-event-time-schedules list --project-transport-plan 123

  # Filter by transport order
  xbe view project-transport-plan-planned-event-time-schedules list --transport-order 456

  # Filter by success
  xbe view project-transport-plan-planned-event-time-schedules list --success true

  # JSON output
  xbe view project-transport-plan-planned-event-time-schedules list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanPlannedEventTimeSchedulesList,
	}
	initProjectTransportPlanPlannedEventTimeSchedulesListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanPlannedEventTimeSchedulesCmd.AddCommand(newProjectTransportPlanPlannedEventTimeSchedulesListCmd())
}

func initProjectTransportPlanPlannedEventTimeSchedulesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("transport-order", "", "Filter by transport order ID")
	cmd.Flags().String("success", "", "Filter by success flag (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanPlannedEventTimeSchedulesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanPlannedEventTimeSchedulesListOptions(cmd)
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
	query.Set("fields[project-transport-plan-planned-event-time-schedules]", "success,project-transport-plan,transport-order")

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
	setFilterIfPresent(query, "filter[transport-order]", opts.TransportOrder)
	setFilterIfPresent(query, "filter[success]", opts.Success)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-planned-event-time-schedules", query)
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

	rows := buildProjectTransportPlanPlannedEventTimeScheduleRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanPlannedEventTimeSchedulesTable(cmd, rows)
}

func parseProjectTransportPlanPlannedEventTimeSchedulesListOptions(cmd *cobra.Command) (projectTransportPlanPlannedEventTimeSchedulesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	transportOrder, _ := cmd.Flags().GetString("transport-order")
	success, _ := cmd.Flags().GetString("success")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanPlannedEventTimeSchedulesListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Sort:                 sort,
		ProjectTransportPlan: projectTransportPlan,
		TransportOrder:       transportOrder,
		Success:              success,
	}, nil
}

func buildProjectTransportPlanPlannedEventTimeScheduleRows(resp jsonAPIResponse) []projectTransportPlanPlannedEventTimeScheduleRow {
	rows := make([]projectTransportPlanPlannedEventTimeScheduleRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectTransportPlanPlannedEventTimeScheduleRow{
			ID:      resource.ID,
			Success: boolAttr(resource.Attributes, "success"),
		}

		if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
			row.ProjectTransportPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["transport-order"]; ok && rel.Data != nil {
			row.TransportOrderID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectTransportPlanPlannedEventTimeSchedulesTable(cmd *cobra.Command, rows []projectTransportPlanPlannedEventTimeScheduleRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan planned event time schedules found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPLAN\tORDER\tSUCCESS")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%t\n",
			row.ID,
			row.ProjectTransportPlanID,
			row.TransportOrderID,
			row.Success,
		)
	}
	return writer.Flush()
}
