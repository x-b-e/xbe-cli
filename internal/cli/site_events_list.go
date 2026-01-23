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

type siteEventsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
	MaterialTransaction    string
	Broker                 string
	Trucker                string
	EventType              string
	EventAtMin             string
	EventAtMax             string
	DriverDay              string
	HasShift               string
	MostRecentByShift      string
	MostRecentByDriverDay  string
}

type siteEventRow struct {
	ID                     string `json:"id"`
	EventType              string `json:"event_type,omitempty"`
	EventKind              string `json:"event_kind,omitempty"`
	EventAt                string `json:"event_at,omitempty"`
	EventSiteType          string `json:"event_site_type,omitempty"`
	EventSiteID            string `json:"event_site_id,omitempty"`
	TenderJobScheduleShift string `json:"tender_job_schedule_shift_id,omitempty"`
	MaterialTransaction    string `json:"material_transaction_id,omitempty"`
	TruckerID              string `json:"trucker_id,omitempty"`
	BrokerID               string `json:"broker_id,omitempty"`
}

func newSiteEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List site events",
		Long: `List site events.

Output Columns:
  ID          Site event identifier
  TYPE        Event type
  KIND        Event kind
  EVENT AT    Event timestamp
  SITE        Event site (type/id)
  SHIFT       Tender job schedule shift ID
  TRANSACTION Material transaction ID
  TRUCKER     Trucker ID
  BROKER      Broker ID

Filters:
  --tender-job-schedule-shift   Filter by tender job schedule shift ID
  --material-transaction        Filter by material transaction ID
  --broker                      Filter by broker ID
  --trucker                     Filter by trucker ID
  --event-type                  Filter by event type (arrive_site,start_work,stop_work,depart_site,at_site)
  --event-at-min                Filter by event-at on/after (ISO 8601)
  --event-at-max                Filter by event-at on/before (ISO 8601)
  --driver-day                  Filter by driver day (trucker shift set) ID
  --has-shift                   Filter by whether event has a shift (true/false)
  --most-recent-by-shift        Filter to most recent event per shift (true/false)
  --most-recent-by-driver-day   Filter to most recent event per driver day (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List site events
  xbe view site-events list

  # Filter by shift
  xbe view site-events list --tender-job-schedule-shift 123

  # Filter by event type and date
  xbe view site-events list --event-type start_work --event-at-min 2025-01-01T00:00:00Z

  # JSON output
  xbe view site-events list --json`,
		Args: cobra.NoArgs,
		RunE: runSiteEventsList,
	}
	initSiteEventsListFlags(cmd)
	return cmd
}

func init() {
	siteEventsCmd.AddCommand(newSiteEventsListCmd())
}

func initSiteEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("material-transaction", "", "Filter by material transaction ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("event-type", "", "Filter by event type")
	cmd.Flags().String("event-at-min", "", "Filter by event-at on/after (ISO 8601)")
	cmd.Flags().String("event-at-max", "", "Filter by event-at on/before (ISO 8601)")
	cmd.Flags().String("driver-day", "", "Filter by driver day (trucker shift set) ID")
	cmd.Flags().String("has-shift", "", "Filter by whether event has a shift (true/false)")
	cmd.Flags().String("most-recent-by-shift", "", "Filter to most recent event per shift (true/false)")
	cmd.Flags().String("most-recent-by-driver-day", "", "Filter to most recent event per driver day (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSiteEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseSiteEventsListOptions(cmd)
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
	query.Set("fields[site-events]", "event-type,event-kind,event-at,event-site,tender-job-schedule-shift,material-transaction,trucker,broker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[material_transaction]", opts.MaterialTransaction)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[event_type]", opts.EventType)
	setFilterIfPresent(query, "filter[event-at-min]", opts.EventAtMin)
	setFilterIfPresent(query, "filter[event-at-max]", opts.EventAtMax)
	setFilterIfPresent(query, "filter[driver_day]", opts.DriverDay)
	setFilterIfPresent(query, "filter[has_shift]", opts.HasShift)
	setFilterIfPresent(query, "filter[most_recent_by_shift]", opts.MostRecentByShift)
	setFilterIfPresent(query, "filter[most_recent_by_driver_day]", opts.MostRecentByDriverDay)

	body, _, err := client.Get(cmd.Context(), "/v1/site-events", query)
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

	rows := buildSiteEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderSiteEventsTable(cmd, rows)
}

func parseSiteEventsListOptions(cmd *cobra.Command) (siteEventsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	broker, _ := cmd.Flags().GetString("broker")
	trucker, _ := cmd.Flags().GetString("trucker")
	eventType, _ := cmd.Flags().GetString("event-type")
	eventAtMin, _ := cmd.Flags().GetString("event-at-min")
	eventAtMax, _ := cmd.Flags().GetString("event-at-max")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	hasShift, _ := cmd.Flags().GetString("has-shift")
	mostRecentByShift, _ := cmd.Flags().GetString("most-recent-by-shift")
	mostRecentByDriverDay, _ := cmd.Flags().GetString("most-recent-by-driver-day")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return siteEventsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TenderJobScheduleShift: tenderJobScheduleShift,
		MaterialTransaction:    materialTransaction,
		Broker:                 broker,
		Trucker:                trucker,
		EventType:              eventType,
		EventAtMin:             eventAtMin,
		EventAtMax:             eventAtMax,
		DriverDay:              driverDay,
		HasShift:               hasShift,
		MostRecentByShift:      mostRecentByShift,
		MostRecentByDriverDay:  mostRecentByDriverDay,
	}, nil
}

func buildSiteEventRows(resp jsonAPIResponse) []siteEventRow {
	rows := make([]siteEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildSiteEventRow(resource))
	}
	return rows
}

func buildSiteEventRow(resource jsonAPIResource) siteEventRow {
	row := siteEventRow{
		ID:        resource.ID,
		EventType: stringAttr(resource.Attributes, "event-type"),
		EventKind: stringAttr(resource.Attributes, "event-kind"),
		EventAt:   formatDateTime(stringAttr(resource.Attributes, "event-at")),
	}

	if rel, ok := resource.Relationships["event-site"]; ok && rel.Data != nil {
		row.EventSiteType = rel.Data.Type
		row.EventSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShift = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		row.MaterialTransaction = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}

func renderSiteEventsTable(cmd *cobra.Command, rows []siteEventRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No site events found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tTYPE\tKIND\tEVENT AT\tSITE\tSHIFT\tTRANSACTION\tTRUCKER\tBROKER")
	for _, row := range rows {
		site := truncateString(formatResourceRef(row.EventSiteType, row.EventSiteID), 28)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EventType,
			row.EventKind,
			row.EventAt,
			site,
			row.TenderJobScheduleShift,
			row.MaterialTransaction,
			row.TruckerID,
			row.BrokerID,
		)
	}
	return writer.Flush()
}
