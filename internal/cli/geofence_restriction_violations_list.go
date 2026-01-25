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

type geofenceRestrictionViolationsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	Geofence               string
	Trailer                string
	Tractor                string
	Driver                 string
	TenderJobScheduleShift string
	EventAtMin             string
	EventAtMax             string
	NotificationSentAtMin  string
	NotificationSentAtMax  string
	ShouldNotify           string
	CreatedAtMin           string
	CreatedAtMax           string
	UpdatedAtMin           string
	UpdatedAtMax           string
}

type geofenceRestrictionViolationRow struct {
	ID                     string `json:"id"`
	GeofenceID             string `json:"geofence_id,omitempty"`
	TrailerID              string `json:"trailer_id,omitempty"`
	TractorID              string `json:"tractor_id,omitempty"`
	DriverID               string `json:"driver_id,omitempty"`
	TenderJobScheduleShift string `json:"tender_job_schedule_shift_id,omitempty"`
	EventAt                string `json:"event_at,omitempty"`
	ShouldNotify           bool   `json:"should_notify"`
	NotificationSentAt     string `json:"notification_sent_at,omitempty"`
}

func newGeofenceRestrictionViolationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List geofence restriction violations",
		Long: `List geofence restriction violations with filtering and pagination.

Geofence restriction violations are generated when a trailer, tractor, or driver
triggers a restricted geofence.

Output Columns:
  ID                 Violation identifier
  GEOFENCE           Geofence ID
  TRAILER            Trailer ID (if present)
  TRACTOR            Tractor ID (if present)
  DRIVER             Driver ID (if present)
  SHIFT              Tender job schedule shift ID (if present)
  EVENT AT           Event timestamp
  SHOULD NOTIFY      Whether notifications should be sent
  NOTIFICATION AT    Notification timestamp (if sent)

Filters:
  --geofence                     Filter by geofence ID
  --trailer                      Filter by trailer ID
  --tractor                      Filter by tractor ID
  --driver                       Filter by driver (user) ID
  --tender-job-schedule-shift    Filter by tender job schedule shift ID
  --event-at-min                 Filter by minimum event time (ISO 8601)
  --event-at-max                 Filter by maximum event time (ISO 8601)
  --notification-sent-at-min     Filter by minimum notification time (ISO 8601)
  --notification-sent-at-max     Filter by maximum notification time (ISO 8601)
  --should-notify                Filter by should-notify status (true/false)
  --created-at-min               Filter by minimum created-at time (ISO 8601)
  --created-at-max               Filter by maximum created-at time (ISO 8601)
  --updated-at-min               Filter by minimum updated-at time (ISO 8601)
  --updated-at-max               Filter by maximum updated-at time (ISO 8601)`,
		Example: `  # List violations
  xbe view geofence-restriction-violations list

  # Filter by geofence
  xbe view geofence-restriction-violations list --geofence 123

  # Filter by event time range
  xbe view geofence-restriction-violations list \
    --event-at-min 2025-01-01T00:00:00Z \
    --event-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view geofence-restriction-violations list --json`,
		RunE: runGeofenceRestrictionViolationsList,
	}
	initGeofenceRestrictionViolationsListFlags(cmd)
	return cmd
}

func init() {
	geofenceRestrictionViolationsCmd.AddCommand(newGeofenceRestrictionViolationsListCmd())
}

func initGeofenceRestrictionViolationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order (prefix with - for descending)")
	cmd.Flags().String("geofence", "", "Filter by geofence ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("driver", "", "Filter by driver (user) ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("event-at-min", "", "Filter by minimum event time (ISO 8601)")
	cmd.Flags().String("event-at-max", "", "Filter by maximum event time (ISO 8601)")
	cmd.Flags().String("notification-sent-at-min", "", "Filter by minimum notification time (ISO 8601)")
	cmd.Flags().String("notification-sent-at-max", "", "Filter by maximum notification time (ISO 8601)")
	cmd.Flags().String("should-notify", "", "Filter by should-notify status (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by minimum created-at time (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by maximum created-at time (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by minimum updated-at time (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by maximum updated-at time (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGeofenceRestrictionViolationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseGeofenceRestrictionViolationsListOptions(cmd)
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
	query.Set("fields[geofence-restriction-violations]", "event-at,should-notify,notification-sent-at,geofence,trailer,tractor,driver,tender-job-schedule-shift")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[geofence]", opts.Geofence)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[tender-job-schedule-shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[event-at-min]", opts.EventAtMin)
	setFilterIfPresent(query, "filter[event-at-max]", opts.EventAtMax)
	setFilterIfPresent(query, "filter[notification-sent-at-min]", opts.NotificationSentAtMin)
	setFilterIfPresent(query, "filter[notification-sent-at-max]", opts.NotificationSentAtMax)
	setFilterIfPresent(query, "filter[should-notify]", opts.ShouldNotify)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/geofence-restriction-violations", query)
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

	rows := buildGeofenceRestrictionViolationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderGeofenceRestrictionViolationsTable(cmd, rows)
}

func parseGeofenceRestrictionViolationsListOptions(cmd *cobra.Command) (geofenceRestrictionViolationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	geofence, _ := cmd.Flags().GetString("geofence")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractor, _ := cmd.Flags().GetString("tractor")
	driver, _ := cmd.Flags().GetString("driver")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	eventAtMin, _ := cmd.Flags().GetString("event-at-min")
	eventAtMax, _ := cmd.Flags().GetString("event-at-max")
	notificationSentAtMin, _ := cmd.Flags().GetString("notification-sent-at-min")
	notificationSentAtMax, _ := cmd.Flags().GetString("notification-sent-at-max")
	shouldNotify, _ := cmd.Flags().GetString("should-notify")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return geofenceRestrictionViolationsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		Geofence:               geofence,
		Trailer:                trailer,
		Tractor:                tractor,
		Driver:                 driver,
		TenderJobScheduleShift: tenderJobScheduleShift,
		EventAtMin:             eventAtMin,
		EventAtMax:             eventAtMax,
		NotificationSentAtMin:  notificationSentAtMin,
		NotificationSentAtMax:  notificationSentAtMax,
		ShouldNotify:           shouldNotify,
		CreatedAtMin:           createdAtMin,
		CreatedAtMax:           createdAtMax,
		UpdatedAtMin:           updatedAtMin,
		UpdatedAtMax:           updatedAtMax,
	}, nil
}

func buildGeofenceRestrictionViolationRows(resp jsonAPIResponse) []geofenceRestrictionViolationRow {
	rows := make([]geofenceRestrictionViolationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := geofenceRestrictionViolationRow{
			ID:                 resource.ID,
			EventAt:            formatDateTime(stringAttr(attrs, "event-at")),
			ShouldNotify:       boolAttr(attrs, "should-notify"),
			NotificationSentAt: formatDateTime(stringAttr(attrs, "notification-sent-at")),
		}

		if rel, ok := resource.Relationships["geofence"]; ok && rel.Data != nil {
			row.GeofenceID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
			row.TrailerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
			row.TractorID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.DriverID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShift = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildGeofenceRestrictionViolationRowFromSingle(resp jsonAPISingleResponse) geofenceRestrictionViolationRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := geofenceRestrictionViolationRow{
		ID:                 resource.ID,
		EventAt:            formatDateTime(stringAttr(attrs, "event-at")),
		ShouldNotify:       boolAttr(attrs, "should-notify"),
		NotificationSentAt: formatDateTime(stringAttr(attrs, "notification-sent-at")),
	}

	if rel, ok := resource.Relationships["geofence"]; ok && rel.Data != nil {
		row.GeofenceID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
		row.TrailerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
		row.TractorID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShift = rel.Data.ID
	}

	return row
}

func renderGeofenceRestrictionViolationsTable(cmd *cobra.Command, rows []geofenceRestrictionViolationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No geofence restriction violations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tGEOFENCE\tTRAILER\tTRACTOR\tDRIVER\tSHIFT\tEVENT AT\tSHOULD NOTIFY\tNOTIFICATION AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\t%s\n",
			row.ID,
			row.GeofenceID,
			row.TrailerID,
			row.TractorID,
			row.DriverID,
			row.TenderJobScheduleShift,
			row.EventAt,
			row.ShouldNotify,
			row.NotificationSentAt,
		)
	}
	return writer.Flush()
}
