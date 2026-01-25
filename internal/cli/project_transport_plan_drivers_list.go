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

type projectTransportPlanDriversListOptions struct {
	BaseURL                               string
	Token                                 string
	JSON                                  bool
	NoAuth                                bool
	Limit                                 int
	Offset                                int
	Sort                                  string
	ProjectTransportPlan                  string
	ProjectTransportPlanStatus            string
	Driver                                string
	HasDriver                             string
	SegmentStart                          string
	SegmentEnd                            string
	Status                                string
	WindowStartAtCached                   string
	WindowStartAtCachedMin                string
	WindowStartAtCachedMax                string
	HasWindowStartAtCached                string
	WindowEndAtCached                     string
	WindowEndAtCachedMin                  string
	WindowEndAtCachedMax                  string
	HasWindowEndAtCached                  string
	Trucker                               string
	InboundProjectOffice                  string
	MostRecent                            string
	HasManagedTransportOrderOrNoTransport string
	Broker                                string
}

type projectTransportPlanDriverRow struct {
	ID                               string `json:"id"`
	Status                           string `json:"status,omitempty"`
	ProjectTransportPlanID           string `json:"project_transport_plan_id,omitempty"`
	ProjectTransportPlanStatus       string `json:"project_transport_plan_status,omitempty"`
	DriverID                         string `json:"driver_id,omitempty"`
	DriverName                       string `json:"driver_name,omitempty"`
	SegmentStartID                   string `json:"segment_start_id,omitempty"`
	SegmentStartPosition             string `json:"segment_start_position,omitempty"`
	SegmentStartTruckerID            string `json:"segment_start_trucker_id,omitempty"`
	SegmentEndID                     string `json:"segment_end_id,omitempty"`
	SegmentEndPosition               string `json:"segment_end_position,omitempty"`
	SegmentEndTruckerID              string `json:"segment_end_trucker_id,omitempty"`
	WindowStartAtCached              string `json:"window_start_at_cached,omitempty"`
	WindowEndAtCached                string `json:"window_end_at_cached,omitempty"`
	InboundProjectOfficeID           string `json:"inbound_project_office_id,omitempty"`
	InboundProjectOfficeName         string `json:"inbound_project_office_name,omitempty"`
	InboundProjectOfficeAbbreviation string `json:"inbound_project_office_abbreviation,omitempty"`
	BrokerID                         string `json:"broker_id,omitempty"`
	BrokerName                       string `json:"broker_name,omitempty"`
}

func newProjectTransportPlanDriversListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plan drivers",
		Long: `List project transport plan drivers with filtering and pagination.

Output Columns:
  ID           Driver assignment identifier
  STATUS       Assignment status
  PLAN         Project transport plan ID
  PLAN STATUS  Project transport plan status
  DRIVER       Driver name or ID
  SEG START    Segment start position or ID
  SEG END      Segment end position or ID
  WINDOW START Assignment window start (cached)
  WINDOW END   Assignment window end (cached)
  OFFICE       Inbound project office

Filters:
  --project-transport-plan                              Filter by project transport plan ID
  --project-transport-plan-status                       Filter by project transport plan status
  --driver                                              Filter by driver (user) ID
  --has-driver                                          Filter by presence of driver (true/false)
  --segment-start                                       Filter by segment start ID
  --segment-end                                         Filter by segment end ID
  --status                                              Filter by assignment status (editing/pending/active)
  --window-start-at-cached                              Filter by window start date (YYYY-MM-DD)
  --window-start-at-cached-min                          Filter by minimum window start date (YYYY-MM-DD)
  --window-start-at-cached-max                          Filter by maximum window start date (YYYY-MM-DD)
  --has-window-start-at-cached                          Filter by presence of window start date (true/false)
  --window-end-at-cached                                Filter by window end date (YYYY-MM-DD)
  --window-end-at-cached-min                            Filter by minimum window end date (YYYY-MM-DD)
  --window-end-at-cached-max                            Filter by maximum window end date (YYYY-MM-DD)
  --has-window-end-at-cached                            Filter by presence of window end date (true/false)
  --trucker                                             Filter by trucker ID (segment range)
  --inbound-project-office                              Filter by inbound project office ID
  --most-recent                                         Filter to most recent assignment per driver (true/false)
  --has-managed-transport-order-or-no-transport-order   Filter by managed transport orders or none (true/false)
  --broker                                              Filter by broker ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plan drivers
  xbe view project-transport-plan-drivers list

  # Filter by plan and status
  xbe view project-transport-plan-drivers list --project-transport-plan 123 --status active

  # Filter by driver
  xbe view project-transport-plan-drivers list --driver 456

  # Output as JSON
  xbe view project-transport-plan-drivers list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlanDriversList,
	}
	initProjectTransportPlanDriversListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanDriversCmd.AddCommand(newProjectTransportPlanDriversListCmd())
}

func initProjectTransportPlanDriversListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID")
	cmd.Flags().String("project-transport-plan-status", "", "Filter by project transport plan status")
	cmd.Flags().String("driver", "", "Filter by driver (user) ID")
	cmd.Flags().String("has-driver", "", "Filter by presence of driver (true/false)")
	cmd.Flags().String("segment-start", "", "Filter by segment start ID")
	cmd.Flags().String("segment-end", "", "Filter by segment end ID")
	cmd.Flags().String("status", "", "Filter by assignment status (editing/pending/active)")
	cmd.Flags().String("window-start-at-cached", "", "Filter by window start date (YYYY-MM-DD)")
	cmd.Flags().String("window-start-at-cached-min", "", "Filter by minimum window start date (YYYY-MM-DD)")
	cmd.Flags().String("window-start-at-cached-max", "", "Filter by maximum window start date (YYYY-MM-DD)")
	cmd.Flags().String("has-window-start-at-cached", "", "Filter by presence of window start date (true/false)")
	cmd.Flags().String("window-end-at-cached", "", "Filter by window end date (YYYY-MM-DD)")
	cmd.Flags().String("window-end-at-cached-min", "", "Filter by minimum window end date (YYYY-MM-DD)")
	cmd.Flags().String("window-end-at-cached-max", "", "Filter by maximum window end date (YYYY-MM-DD)")
	cmd.Flags().String("has-window-end-at-cached", "", "Filter by presence of window end date (true/false)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (segment range)")
	cmd.Flags().String("inbound-project-office", "", "Filter by inbound project office ID")
	cmd.Flags().String("most-recent", "", "Filter to most recent assignment per driver (true/false)")
	cmd.Flags().String("has-managed-transport-order-or-no-transport-order", "", "Filter by managed transport orders or none (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanDriversList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlanDriversListOptions(cmd)
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
	query.Set("fields[project-transport-plan-drivers]", "status,window-start-at-cached,window-end-at-cached,project-transport-plan,segment-start,segment-end,driver,inbound-project-office")
	query.Set("include", "project-transport-plan,project-transport-plan.broker,segment-start,segment-end,segment-start.trucker,segment-end.trucker,driver,inbound-project-office")
	query.Set("fields[project-transport-plans]", "status,broker")
	query.Set("fields[project-transport-plan-segments]", "position,trucker")
	query.Set("fields[users]", "name")
	query.Set("fields[project-offices]", "name,abbreviation")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")

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
	setFilterIfPresent(query, "filter[project-transport-plan-status]", opts.ProjectTransportPlanStatus)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[has-driver]", opts.HasDriver)
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
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[inbound-project-office]", opts.InboundProjectOffice)
	setFilterIfPresent(query, "filter[most-recent]", opts.MostRecent)
	setFilterIfPresent(query, "filter[has-managed-transport-order-or-no-transport-order]", opts.HasManagedTransportOrderOrNoTransport)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-drivers", query)
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

	rows := buildProjectTransportPlanDriverRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlanDriversTable(cmd, rows)
}

func parseProjectTransportPlanDriversListOptions(cmd *cobra.Command) (projectTransportPlanDriversListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectTransportPlan, _ := cmd.Flags().GetString("project-transport-plan")
	projectTransportPlanStatus, _ := cmd.Flags().GetString("project-transport-plan-status")
	driver, _ := cmd.Flags().GetString("driver")
	hasDriver, _ := cmd.Flags().GetString("has-driver")
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
	trucker, _ := cmd.Flags().GetString("trucker")
	inboundProjectOffice, _ := cmd.Flags().GetString("inbound-project-office")
	mostRecent, _ := cmd.Flags().GetString("most-recent")
	hasManagedTransportOrderOrNoTransport, _ := cmd.Flags().GetString("has-managed-transport-order-or-no-transport-order")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanDriversListOptions{
		BaseURL:                               baseURL,
		Token:                                 token,
		JSON:                                  jsonOut,
		NoAuth:                                noAuth,
		Limit:                                 limit,
		Offset:                                offset,
		Sort:                                  sort,
		ProjectTransportPlan:                  projectTransportPlan,
		ProjectTransportPlanStatus:            projectTransportPlanStatus,
		Driver:                                driver,
		HasDriver:                             hasDriver,
		SegmentStart:                          segmentStart,
		SegmentEnd:                            segmentEnd,
		Status:                                status,
		WindowStartAtCached:                   windowStartAtCached,
		WindowStartAtCachedMin:                windowStartAtCachedMin,
		WindowStartAtCachedMax:                windowStartAtCachedMax,
		HasWindowStartAtCached:                hasWindowStartAtCached,
		WindowEndAtCached:                     windowEndAtCached,
		WindowEndAtCachedMin:                  windowEndAtCachedMin,
		WindowEndAtCachedMax:                  windowEndAtCachedMax,
		HasWindowEndAtCached:                  hasWindowEndAtCached,
		Trucker:                               trucker,
		InboundProjectOffice:                  inboundProjectOffice,
		MostRecent:                            mostRecent,
		HasManagedTransportOrderOrNoTransport: hasManagedTransportOrderOrNoTransport,
		Broker:                                broker,
	}, nil
}

func buildProjectTransportPlanDriverRows(resp jsonAPIResponse) []projectTransportPlanDriverRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]projectTransportPlanDriverRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectTransportPlanDriverRow(resource, included))
	}
	return rows
}

func projectTransportPlanDriverRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanDriverRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	return buildProjectTransportPlanDriverRow(resp.Data, included)
}

func buildProjectTransportPlanDriverRow(resource jsonAPIResource, included map[string]jsonAPIResource) projectTransportPlanDriverRow {
	attrs := resource.Attributes
	row := projectTransportPlanDriverRow{
		ID:                  resource.ID,
		Status:              stringAttr(attrs, "status"),
		WindowStartAtCached: formatDateTime(stringAttr(attrs, "window-start-at-cached")),
		WindowEndAtCached:   formatDateTime(stringAttr(attrs, "window-end-at-cached")),
	}

	if rel, ok := resource.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		row.ProjectTransportPlanID = rel.Data.ID
		if plan, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ProjectTransportPlanStatus = stringAttr(plan.Attributes, "status")
			if brokerRel, ok := plan.Relationships["broker"]; ok && brokerRel.Data != nil {
				row.BrokerID = brokerRel.Data.ID
				if broker, ok := included[resourceKey(brokerRel.Data.Type, brokerRel.Data.ID)]; ok {
					row.BrokerName = stringAttr(broker.Attributes, "company-name")
				}
			}
		}
	}

	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
		if driver, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.DriverName = stringAttr(driver.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["segment-start"]; ok && rel.Data != nil {
		row.SegmentStartID = rel.Data.ID
		if segment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.SegmentStartPosition = stringAttr(segment.Attributes, "position")
			if truckerRel, ok := segment.Relationships["trucker"]; ok && truckerRel.Data != nil {
				row.SegmentStartTruckerID = truckerRel.Data.ID
			}
		}
	}

	if rel, ok := resource.Relationships["segment-end"]; ok && rel.Data != nil {
		row.SegmentEndID = rel.Data.ID
		if segment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.SegmentEndPosition = stringAttr(segment.Attributes, "position")
			if truckerRel, ok := segment.Relationships["trucker"]; ok && truckerRel.Data != nil {
				row.SegmentEndTruckerID = truckerRel.Data.ID
			}
		}
	}

	if rel, ok := resource.Relationships["inbound-project-office"]; ok && rel.Data != nil {
		row.InboundProjectOfficeID = rel.Data.ID
		if office, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.InboundProjectOfficeName = stringAttr(office.Attributes, "name")
			row.InboundProjectOfficeAbbreviation = stringAttr(office.Attributes, "abbreviation")
		}
	}

	return row
}

func renderProjectTransportPlanDriversTable(cmd *cobra.Command, rows []projectTransportPlanDriverRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plan drivers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tPLAN\tPLAN STATUS\tDRIVER\tSEG START\tSEG END\tWINDOW START\tWINDOW END\tOFFICE")
	for _, row := range rows {
		driver := firstNonEmpty(row.DriverName, row.DriverID)
		segmentStart := firstNonEmpty(row.SegmentStartPosition, row.SegmentStartID)
		segmentEnd := firstNonEmpty(row.SegmentEndPosition, row.SegmentEndID)
		office := firstNonEmpty(row.InboundProjectOfficeAbbreviation, row.InboundProjectOfficeName, row.InboundProjectOfficeID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 10),
			truncateString(row.ProjectTransportPlanID, 12),
			truncateString(row.ProjectTransportPlanStatus, 12),
			truncateString(driver, 20),
			truncateString(segmentStart, 10),
			truncateString(segmentEnd, 10),
			truncateString(row.WindowStartAtCached, 16),
			truncateString(row.WindowEndAtCached, 16),
			truncateString(office, 16),
		)
	}
	return writer.Flush()
}
