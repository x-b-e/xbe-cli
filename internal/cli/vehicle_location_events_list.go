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

type vehicleLocationEventsListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Sort                        string
	Tractor                     string
	Trailer                     string
	EventAtMin                  string
	EventAtMax                  string
	IsEventAt                   string
	IncludeDeviceLocationEvents string
	TimeSliceMin                string
	TotalCountMax               string
	CreatedAtMin                string
	CreatedAtMax                string
	IsCreatedAt                 string
	UpdatedAtMin                string
	UpdatedAtMax                string
	IsUpdatedAt                 string
}

type vehicleLocationEventRow struct {
	ID             string `json:"id"`
	TractorID      string `json:"tractor_id,omitempty"`
	TrailerID      string `json:"trailer_id,omitempty"`
	EventAt        string `json:"event_at,omitempty"`
	EventLatitude  string `json:"event_latitude,omitempty"`
	EventLongitude string `json:"event_longitude,omitempty"`
}

func newVehicleLocationEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List vehicle location events",
		Long: `List vehicle location events.

Output Columns:
  ID       Vehicle location event identifier
  TRACTOR  Tractor ID
  TRAILER  Trailer ID
  EVENT AT Event timestamp
  LAT      Event latitude
  LON      Event longitude

Filters:
  --tractor                          Filter by tractor ID
  --trailer                          Filter by trailer ID
  --event-at-min                     Filter by minimum event timestamp (ISO 8601)
  --event-at-max                     Filter by maximum event timestamp (ISO 8601)
  --is-event-at                      Filter by has event timestamp (true/false)
  --include-device-location-events   Include blended device location events (true/false). Requires --tractor or --trailer
  --time-slice-min                   Resample events into time buckets (integer minutes, >= 1)
  --total-count-max                  Limit total events returned (integer, >= 1)
  --created-at-min                   Filter by created-at on/after (ISO 8601)
  --created-at-max                   Filter by created-at on/before (ISO 8601)
  --is-created-at                    Filter by has created-at (true/false)
  --updated-at-min                   Filter by updated-at on/after (ISO 8601)
  --updated-at-max                   Filter by updated-at on/before (ISO 8601)
  --is-updated-at                    Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List vehicle location events
  xbe view vehicle-location-events list

  # Filter by tractor
  xbe view vehicle-location-events list --tractor 123

  # Filter by event time range
  xbe view vehicle-location-events list --event-at-min 2025-01-01T00:00:00Z --event-at-max 2025-01-31T23:59:59Z

  # Include blended device location events
  xbe view vehicle-location-events list --tractor 123 --include-device-location-events true

  # Resample or cap events
  xbe view vehicle-location-events list --time-slice-min 5 --total-count-max 100

  # Output as JSON
  xbe view vehicle-location-events list --json`,
		Args: cobra.NoArgs,
		RunE: runVehicleLocationEventsList,
	}
	initVehicleLocationEventsListFlags(cmd)
	return cmd
}

func init() {
	vehicleLocationEventsCmd.AddCommand(newVehicleLocationEventsListCmd())
}

func initVehicleLocationEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("event-at-min", "", "Filter by minimum event timestamp (ISO 8601)")
	cmd.Flags().String("event-at-max", "", "Filter by maximum event timestamp (ISO 8601)")
	cmd.Flags().String("is-event-at", "", "Filter by has event timestamp (true/false)")
	cmd.Flags().String("include-device-location-events", "", "Include blended device location events (true/false)")
	cmd.Flags().String("time-slice-min", "", "Resample events into time buckets (integer minutes, >= 1)")
	cmd.Flags().String("total-count-max", "", "Limit total events returned (integer, >= 1)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runVehicleLocationEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseVehicleLocationEventsListOptions(cmd)
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
	query.Set("fields[vehicle-location-events]", "event-latitude,event-longitude,event-at,tractor,trailer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[event_at_min]", opts.EventAtMin)
	setFilterIfPresent(query, "filter[event_at_max]", opts.EventAtMax)
	setFilterIfPresent(query, "filter[is_event_at]", opts.IsEventAt)
	setFilterIfPresent(query, "filter[include_device_location_events]", opts.IncludeDeviceLocationEvents)
	setFilterIfPresent(query, "filter[time_slice_min]", opts.TimeSliceMin)
	setFilterIfPresent(query, "filter[total_count_max]", opts.TotalCountMax)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/vehicle-location-events", query)
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

	rows := buildVehicleLocationEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderVehicleLocationEventsTable(cmd, rows)
}

func parseVehicleLocationEventsListOptions(cmd *cobra.Command) (vehicleLocationEventsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tractor, _ := cmd.Flags().GetString("tractor")
	trailer, _ := cmd.Flags().GetString("trailer")
	eventAtMin, _ := cmd.Flags().GetString("event-at-min")
	eventAtMax, _ := cmd.Flags().GetString("event-at-max")
	isEventAt, _ := cmd.Flags().GetString("is-event-at")
	includeDeviceLocationEvents, _ := cmd.Flags().GetString("include-device-location-events")
	timeSliceMin, _ := cmd.Flags().GetString("time-slice-min")
	totalCountMax, _ := cmd.Flags().GetString("total-count-max")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return vehicleLocationEventsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Sort:                        sort,
		Tractor:                     tractor,
		Trailer:                     trailer,
		EventAtMin:                  eventAtMin,
		EventAtMax:                  eventAtMax,
		IsEventAt:                   isEventAt,
		IncludeDeviceLocationEvents: includeDeviceLocationEvents,
		TimeSliceMin:                timeSliceMin,
		TotalCountMax:               totalCountMax,
		CreatedAtMin:                createdAtMin,
		CreatedAtMax:                createdAtMax,
		IsCreatedAt:                 isCreatedAt,
		UpdatedAtMin:                updatedAtMin,
		UpdatedAtMax:                updatedAtMax,
		IsUpdatedAt:                 isUpdatedAt,
	}, nil
}

func buildVehicleLocationEventRows(resp jsonAPIResponse) []vehicleLocationEventRow {
	rows := make([]vehicleLocationEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildVehicleLocationEventRow(resource))
	}
	return rows
}

func buildVehicleLocationEventRow(resource jsonAPIResource) vehicleLocationEventRow {
	row := vehicleLocationEventRow{
		ID:             resource.ID,
		EventAt:        formatDateTime(stringAttr(resource.Attributes, "event-at")),
		EventLatitude:  stringAttr(resource.Attributes, "event-latitude"),
		EventLongitude: stringAttr(resource.Attributes, "event-longitude"),
	}

	if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
		row.TractorID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
		row.TrailerID = rel.Data.ID
	}

	return row
}

func renderVehicleLocationEventsTable(cmd *cobra.Command, rows []vehicleLocationEventRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No vehicle location events found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRACTOR\tTRAILER\tEVENT AT\tLAT\tLON")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TractorID,
			row.TrailerID,
			row.EventAt,
			row.EventLatitude,
			row.EventLongitude,
		)
	}
	return writer.Flush()
}
