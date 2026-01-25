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

type projectTransportPlansListOptions struct {
	BaseURL                                     string
	Token                                       string
	JSON                                        bool
	NoAuth                                      bool
	Limit                                       int
	Offset                                      int
	Sort                                        string
	Project                                     string
	Broker                                      string
	ProjectTransportOrganization                string
	BrokerID                                    string
	Q                                           string
	SegmentMilesMin                             string
	SegmentMilesMax                             string
	EventTimesAtMinMin                          string
	EventTimesAtMinMax                          string
	IsEventTimesAtMin                           string
	EventTimesAtMaxMin                          string
	EventTimesAtMaxMax                          string
	IsEventTimesAtMax                           string
	EventTimesOnMin                             string
	EventTimesOnMinMin                          string
	EventTimesOnMinMax                          string
	HasEventTimesOnMin                          string
	EventTimesOnMax                             string
	EventTimesOnMaxMin                          string
	EventTimesOnMaxMax                          string
	HasEventTimesOnMax                          string
	MaybeActive                                 string
	ProjectOffice                               string
	ProjectCategory                             string
	IsManaged                                   string
	NearestProjectOfficeIDs                     string
	ExternalOrderNumber                         string
	TransportOrderProjectOfficeOrNearestOffices string
	PickupAddressStateCodes                     string
	DeliveryAddressStateCodes                   string
}

type projectTransportPlanRow struct {
	ID                                string `json:"id"`
	Status                            string `json:"status,omitempty"`
	ProjectID                         string `json:"project_id,omitempty"`
	BrokerID                          string `json:"broker_id,omitempty"`
	ProjectTransportPlanStrategySetID string `json:"project_transport_plan_strategy_set_id,omitempty"`
	SegmentMiles                      any    `json:"segment_miles,omitempty"`
	EventTimesAtMin                   string `json:"event_times_at_min,omitempty"`
	EventTimesAtMax                   string `json:"event_times_at_max,omitempty"`
	EventTimesOnMin                   string `json:"event_times_on_min,omitempty"`
	EventTimesOnMax                   string `json:"event_times_on_max,omitempty"`
	StrategySetPredictionPosition     string `json:"strategy_set_prediction_position,omitempty"`
}

func newProjectTransportPlansListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project transport plans",
		Long: `List project transport plans with filtering and pagination.

Output Columns:
  ID        Plan identifier
  STATUS    Plan status
  PROJECT   Project ID
  BROKER    Broker ID
  STRATEGY  Strategy set ID
  SEG MI    Cached total segment miles
  AT MIN    Earliest event timestamp (UTC)
  AT MAX    Latest event timestamp (UTC)
  ON MIN    Earliest event date
  ON MAX    Latest event date

Filters:
  --project                                        Filter by project ID
  --broker                                         Filter by broker ID
  --project-transport-organization                 Filter by project transport organization ID
  --broker-id                                      Filter by broker ID via project developer
  --q                                              Search query (federated)
  --segment-miles-min                              Filter by minimum segment miles
  --segment-miles-max                              Filter by maximum segment miles
  --event-times-at-min-min                         Filter by minimum earliest event time (RFC3339)
  --event-times-at-min-max                         Filter by maximum earliest event time (RFC3339)
  --is-event-times-at-min                          Filter by presence of earliest event time (true/false)
  --event-times-at-max-min                         Filter by minimum latest event time (RFC3339)
  --event-times-at-max-max                         Filter by maximum latest event time (RFC3339)
  --is-event-times-at-max                          Filter by presence of latest event time (true/false)
  --event-times-on-min                             Filter by earliest event date (YYYY-MM-DD)
  --event-times-on-min-min                         Filter by minimum earliest event date (YYYY-MM-DD)
  --event-times-on-min-max                         Filter by maximum earliest event date (YYYY-MM-DD)
  --has-event-times-on-min                         Filter by presence of earliest event date (true/false)
  --event-times-on-max                             Filter by latest event date (YYYY-MM-DD)
  --event-times-on-max-min                         Filter by minimum latest event date (YYYY-MM-DD)
  --event-times-on-max-max                         Filter by maximum latest event date (YYYY-MM-DD)
  --has-event-times-on-max                         Filter by presence of latest event date (true/false)
  --maybe-active                                   Filter by maybe active status (true/false)
  --project-office                                 Filter by project office ID(s)
  --project-category                               Filter by project category ID(s)
  --is-managed                                     Filter by managed status (true/false)
  --nearest-project-office-ids                     Filter by nearest project office ID(s)
  --external-order-number                          Filter by transport order external order number
  --transport-order-project-office-or-nearest-project-offices  Filter by project office ID(s)
  --pickup-address-state-codes                     Filter by pickup state code(s)
  --delivery-address-state-codes                   Filter by delivery state code(s)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project transport plans
  xbe view project-transport-plans list

  # Filter by project and managed status
  xbe view project-transport-plans list --project 123 --is-managed true

  # Filter by event time range
  xbe view project-transport-plans list --event-times-at-min-min 2026-01-15T00:00:00Z --event-times-at-max-max 2026-01-16T00:00:00Z

  # Output as JSON
  xbe view project-transport-plans list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectTransportPlansList,
	}
	initProjectTransportPlansListFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlansCmd.AddCommand(newProjectTransportPlansListCmd())
}

func initProjectTransportPlansListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("project-transport-organization", "", "Filter by project transport organization ID")
	cmd.Flags().String("broker-id", "", "Filter by broker ID via project developer")
	cmd.Flags().String("q", "", "Search query")
	cmd.Flags().String("segment-miles-min", "", "Filter by minimum segment miles")
	cmd.Flags().String("segment-miles-max", "", "Filter by maximum segment miles")
	cmd.Flags().String("event-times-at-min-min", "", "Filter by minimum earliest event time (RFC3339)")
	cmd.Flags().String("event-times-at-min-max", "", "Filter by maximum earliest event time (RFC3339)")
	cmd.Flags().String("is-event-times-at-min", "", "Filter by presence of earliest event time (true/false)")
	cmd.Flags().String("event-times-at-max-min", "", "Filter by minimum latest event time (RFC3339)")
	cmd.Flags().String("event-times-at-max-max", "", "Filter by maximum latest event time (RFC3339)")
	cmd.Flags().String("is-event-times-at-max", "", "Filter by presence of latest event time (true/false)")
	cmd.Flags().String("event-times-on-min", "", "Filter by earliest event date (YYYY-MM-DD)")
	cmd.Flags().String("event-times-on-min-min", "", "Filter by minimum earliest event date (YYYY-MM-DD)")
	cmd.Flags().String("event-times-on-min-max", "", "Filter by maximum earliest event date (YYYY-MM-DD)")
	cmd.Flags().String("has-event-times-on-min", "", "Filter by presence of earliest event date (true/false)")
	cmd.Flags().String("event-times-on-max", "", "Filter by latest event date (YYYY-MM-DD)")
	cmd.Flags().String("event-times-on-max-min", "", "Filter by minimum latest event date (YYYY-MM-DD)")
	cmd.Flags().String("event-times-on-max-max", "", "Filter by maximum latest event date (YYYY-MM-DD)")
	cmd.Flags().String("has-event-times-on-max", "", "Filter by presence of latest event date (true/false)")
	cmd.Flags().String("maybe-active", "", "Filter by maybe active status (true/false)")
	cmd.Flags().String("project-office", "", "Filter by project office ID(s)")
	cmd.Flags().String("project-category", "", "Filter by project category ID(s)")
	cmd.Flags().String("is-managed", "", "Filter by managed status (true/false)")
	cmd.Flags().String("nearest-project-office-ids", "", "Filter by nearest project office ID(s)")
	cmd.Flags().String("external-order-number", "", "Filter by transport order external order number")
	cmd.Flags().String("transport-order-project-office-or-nearest-project-offices", "", "Filter by project office ID(s)")
	cmd.Flags().String("pickup-address-state-codes", "", "Filter by pickup state code(s)")
	cmd.Flags().String("delivery-address-state-codes", "", "Filter by delivery state code(s)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlansList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectTransportPlansListOptions(cmd)
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
	query.Set("fields[project-transport-plans]", "status,segment-miles,event-times-at-min,event-times-at-max,event-times-on-min,event-times-on-max,strategy-set-prediction-position,project,broker,project-transport-plan-strategy-set")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[project-transport-organization]", opts.ProjectTransportOrganization)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[segment-miles-min]", opts.SegmentMilesMin)
	setFilterIfPresent(query, "filter[segment-miles-max]", opts.SegmentMilesMax)
	setFilterIfPresent(query, "filter[event-times-at-min-min]", opts.EventTimesAtMinMin)
	setFilterIfPresent(query, "filter[event-times-at-min-max]", opts.EventTimesAtMinMax)
	setFilterIfPresent(query, "filter[is-event-times-at-min]", opts.IsEventTimesAtMin)
	setFilterIfPresent(query, "filter[event-times-at-max-min]", opts.EventTimesAtMaxMin)
	setFilterIfPresent(query, "filter[event-times-at-max-max]", opts.EventTimesAtMaxMax)
	setFilterIfPresent(query, "filter[is-event-times-at-max]", opts.IsEventTimesAtMax)
	setFilterIfPresent(query, "filter[event-times-on-min]", opts.EventTimesOnMin)
	setFilterIfPresent(query, "filter[event-times-on-min-min]", opts.EventTimesOnMinMin)
	setFilterIfPresent(query, "filter[event-times-on-min-max]", opts.EventTimesOnMinMax)
	setFilterIfPresent(query, "filter[has-event-times-on-min]", opts.HasEventTimesOnMin)
	setFilterIfPresent(query, "filter[event-times-on-max]", opts.EventTimesOnMax)
	setFilterIfPresent(query, "filter[event-times-on-max-min]", opts.EventTimesOnMaxMin)
	setFilterIfPresent(query, "filter[event-times-on-max-max]", opts.EventTimesOnMaxMax)
	setFilterIfPresent(query, "filter[has-event-times-on-max]", opts.HasEventTimesOnMax)
	setFilterIfPresent(query, "filter[maybe-active]", opts.MaybeActive)
	setFilterIfPresent(query, "filter[project-office]", opts.ProjectOffice)
	setFilterIfPresent(query, "filter[project-category]", opts.ProjectCategory)
	setFilterIfPresent(query, "filter[is-managed]", opts.IsManaged)
	setFilterIfPresent(query, "filter[nearest-project-office-ids]", opts.NearestProjectOfficeIDs)
	setFilterIfPresent(query, "filter[external-order-number]", opts.ExternalOrderNumber)
	setFilterIfPresent(query, "filter[transport-order-project-office-or-nearest-project-offices]", opts.TransportOrderProjectOfficeOrNearestOffices)
	setFilterIfPresent(query, "filter[pickup-address-state-codes]", opts.PickupAddressStateCodes)
	setFilterIfPresent(query, "filter[delivery-address-state-codes]", opts.DeliveryAddressStateCodes)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plans", query)
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

	rows := buildProjectTransportPlanRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectTransportPlansTable(cmd, rows)
}

func parseProjectTransportPlansListOptions(cmd *cobra.Command) (projectTransportPlansListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	project, _ := cmd.Flags().GetString("project")
	broker, _ := cmd.Flags().GetString("broker")
	projectTransportOrganization, _ := cmd.Flags().GetString("project-transport-organization")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	q, _ := cmd.Flags().GetString("q")
	segmentMilesMin, _ := cmd.Flags().GetString("segment-miles-min")
	segmentMilesMax, _ := cmd.Flags().GetString("segment-miles-max")
	eventTimesAtMinMin, _ := cmd.Flags().GetString("event-times-at-min-min")
	eventTimesAtMinMax, _ := cmd.Flags().GetString("event-times-at-min-max")
	isEventTimesAtMin, _ := cmd.Flags().GetString("is-event-times-at-min")
	eventTimesAtMaxMin, _ := cmd.Flags().GetString("event-times-at-max-min")
	eventTimesAtMaxMax, _ := cmd.Flags().GetString("event-times-at-max-max")
	isEventTimesAtMax, _ := cmd.Flags().GetString("is-event-times-at-max")
	eventTimesOnMin, _ := cmd.Flags().GetString("event-times-on-min")
	eventTimesOnMinMin, _ := cmd.Flags().GetString("event-times-on-min-min")
	eventTimesOnMinMax, _ := cmd.Flags().GetString("event-times-on-min-max")
	hasEventTimesOnMin, _ := cmd.Flags().GetString("has-event-times-on-min")
	eventTimesOnMax, _ := cmd.Flags().GetString("event-times-on-max")
	eventTimesOnMaxMin, _ := cmd.Flags().GetString("event-times-on-max-min")
	eventTimesOnMaxMax, _ := cmd.Flags().GetString("event-times-on-max-max")
	hasEventTimesOnMax, _ := cmd.Flags().GetString("has-event-times-on-max")
	maybeActive, _ := cmd.Flags().GetString("maybe-active")
	projectOffice, _ := cmd.Flags().GetString("project-office")
	projectCategory, _ := cmd.Flags().GetString("project-category")
	isManaged, _ := cmd.Flags().GetString("is-managed")
	nearestProjectOfficeIDs, _ := cmd.Flags().GetString("nearest-project-office-ids")
	externalOrderNumber, _ := cmd.Flags().GetString("external-order-number")
	transportOrderProjectOfficeOrNearestOffices, _ := cmd.Flags().GetString("transport-order-project-office-or-nearest-project-offices")
	pickupAddressStateCodes, _ := cmd.Flags().GetString("pickup-address-state-codes")
	deliveryAddressStateCodes, _ := cmd.Flags().GetString("delivery-address-state-codes")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlansListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Sort:                         sort,
		Project:                      project,
		Broker:                       broker,
		ProjectTransportOrganization: projectTransportOrganization,
		BrokerID:                     brokerID,
		Q:                            q,
		SegmentMilesMin:              segmentMilesMin,
		SegmentMilesMax:              segmentMilesMax,
		EventTimesAtMinMin:           eventTimesAtMinMin,
		EventTimesAtMinMax:           eventTimesAtMinMax,
		IsEventTimesAtMin:            isEventTimesAtMin,
		EventTimesAtMaxMin:           eventTimesAtMaxMin,
		EventTimesAtMaxMax:           eventTimesAtMaxMax,
		IsEventTimesAtMax:            isEventTimesAtMax,
		EventTimesOnMin:              eventTimesOnMin,
		EventTimesOnMinMin:           eventTimesOnMinMin,
		EventTimesOnMinMax:           eventTimesOnMinMax,
		HasEventTimesOnMin:           hasEventTimesOnMin,
		EventTimesOnMax:              eventTimesOnMax,
		EventTimesOnMaxMin:           eventTimesOnMaxMin,
		EventTimesOnMaxMax:           eventTimesOnMaxMax,
		HasEventTimesOnMax:           hasEventTimesOnMax,
		MaybeActive:                  maybeActive,
		ProjectOffice:                projectOffice,
		ProjectCategory:              projectCategory,
		IsManaged:                    isManaged,
		NearestProjectOfficeIDs:      nearestProjectOfficeIDs,
		ExternalOrderNumber:          externalOrderNumber,
		TransportOrderProjectOfficeOrNearestOffices: transportOrderProjectOfficeOrNearestOffices,
		PickupAddressStateCodes:                     pickupAddressStateCodes,
		DeliveryAddressStateCodes:                   deliveryAddressStateCodes,
	}, nil
}

func buildProjectTransportPlanRows(resp jsonAPIResponse) []projectTransportPlanRow {
	rows := make([]projectTransportPlanRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectTransportPlanRow(resource))
	}
	return rows
}

func buildProjectTransportPlanRow(resource jsonAPIResource) projectTransportPlanRow {
	attrs := resource.Attributes
	row := projectTransportPlanRow{
		ID:                            resource.ID,
		Status:                        stringAttr(attrs, "status"),
		SegmentMiles:                  anyAttr(attrs, "segment-miles"),
		EventTimesAtMin:               formatDateTime(stringAttr(attrs, "event-times-at-min")),
		EventTimesAtMax:               formatDateTime(stringAttr(attrs, "event-times-at-max")),
		EventTimesOnMin:               formatDate(stringAttr(attrs, "event-times-on-min")),
		EventTimesOnMax:               formatDate(stringAttr(attrs, "event-times-on-max")),
		StrategySetPredictionPosition: stringAttr(attrs, "strategy-set-prediction-position"),
	}

	if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
		row.ProjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-transport-plan-strategy-set"]; ok && rel.Data != nil {
		row.ProjectTransportPlanStrategySetID = rel.Data.ID
	}

	return row
}

func projectTransportPlanRowFromSingle(resp jsonAPISingleResponse) projectTransportPlanRow {
	return buildProjectTransportPlanRow(resp.Data)
}

func renderProjectTransportPlansTable(cmd *cobra.Command, rows []projectTransportPlanRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project transport plans found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tPROJECT\tBROKER\tSTRATEGY\tSEG MI\tAT MIN\tAT MAX\tON MIN\tON MAX")
	for _, row := range rows {
		segmentMiles := formatDistanceMiles(row.SegmentMiles)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 12),
			truncateString(row.ProjectID, 12),
			truncateString(row.BrokerID, 12),
			truncateString(row.ProjectTransportPlanStrategySetID, 12),
			segmentMiles,
			truncateString(row.EventTimesAtMin, 16),
			truncateString(row.EventTimesAtMax, 16),
			truncateString(row.EventTimesOnMin, 10),
			truncateString(row.EventTimesOnMax, 10),
		)
	}
	return writer.Flush()
}
