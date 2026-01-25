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

type equipmentLocationEventsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Equipment    string
	UpdatedBy    string
	EventAtMin   string
	EventAtMax   string
	IsEventAt    string
	Provenance   string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type equipmentLocationEventRow struct {
	ID             string `json:"id"`
	EquipmentID    string `json:"equipment_id,omitempty"`
	EventAt        string `json:"event_at,omitempty"`
	EventLatitude  string `json:"event_latitude,omitempty"`
	EventLongitude string `json:"event_longitude,omitempty"`
	Provenance     string `json:"provenance,omitempty"`
	UpdatedByID    string `json:"updated_by_id,omitempty"`
}

func newEquipmentLocationEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment location events",
		Long: `List equipment location events.

Output Columns:
  ID          Equipment location event identifier
  EQUIPMENT   Equipment ID
  EVENT AT    Event timestamp
  LAT         Event latitude
  LON         Event longitude
  PROVENANCE  Event provenance (gps, map)
  UPDATED BY  Updated by user ID

Filters:
  --equipment        Filter by equipment ID
  --updated-by       Filter by updated by user ID
  --event-at-min     Filter by minimum event timestamp (ISO 8601)
  --event-at-max     Filter by maximum event timestamp (ISO 8601)
  --is-event-at      Filter by has event timestamp (true/false)
  --provenance       Filter by provenance (gps, map)
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)
  --is-created-at    Filter by has created-at (true/false)
  --updated-at-min   Filter by updated-at on/after (ISO 8601)
  --updated-at-max   Filter by updated-at on/before (ISO 8601)
  --is-updated-at    Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List equipment location events
  xbe view equipment-location-events list

  # Filter by equipment
  xbe view equipment-location-events list --equipment 123

  # Filter by event time range
  xbe view equipment-location-events list --event-at-min 2025-01-01T00:00:00Z --event-at-max 2025-01-31T23:59:59Z

  # Filter by provenance
  xbe view equipment-location-events list --provenance gps

  # Output as JSON
  xbe view equipment-location-events list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentLocationEventsList,
	}
	initEquipmentLocationEventsListFlags(cmd)
	return cmd
}

func init() {
	equipmentLocationEventsCmd.AddCommand(newEquipmentLocationEventsListCmd())
}

func initEquipmentLocationEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("equipment", "", "Filter by equipment ID")
	cmd.Flags().String("updated-by", "", "Filter by updated by user ID")
	cmd.Flags().String("event-at-min", "", "Filter by minimum event timestamp (ISO 8601)")
	cmd.Flags().String("event-at-max", "", "Filter by maximum event timestamp (ISO 8601)")
	cmd.Flags().String("is-event-at", "", "Filter by has event timestamp (true/false)")
	cmd.Flags().String("provenance", "", "Filter by provenance (gps, map)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentLocationEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentLocationEventsListOptions(cmd)
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
	query.Set("fields[equipment-location-events]", "event-latitude,event-longitude,event-at,provenance,equipment,updated-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[equipment]", opts.Equipment)
	setFilterIfPresent(query, "filter[updated_by]", opts.UpdatedBy)
	setFilterIfPresent(query, "filter[event_at_min]", opts.EventAtMin)
	setFilterIfPresent(query, "filter[event_at_max]", opts.EventAtMax)
	setFilterIfPresent(query, "filter[is_event_at]", opts.IsEventAt)
	setFilterIfPresent(query, "filter[provenance]", opts.Provenance)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-location-events", query)
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

	rows := buildEquipmentLocationEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentLocationEventsTable(cmd, rows)
}

func parseEquipmentLocationEventsListOptions(cmd *cobra.Command) (equipmentLocationEventsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	equipment, _ := cmd.Flags().GetString("equipment")
	updatedBy, _ := cmd.Flags().GetString("updated-by")
	eventAtMin, _ := cmd.Flags().GetString("event-at-min")
	eventAtMax, _ := cmd.Flags().GetString("event-at-max")
	isEventAt, _ := cmd.Flags().GetString("is-event-at")
	provenance, _ := cmd.Flags().GetString("provenance")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentLocationEventsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Equipment:    equipment,
		UpdatedBy:    updatedBy,
		EventAtMin:   eventAtMin,
		EventAtMax:   eventAtMax,
		IsEventAt:    isEventAt,
		Provenance:   provenance,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildEquipmentLocationEventRows(resp jsonAPIResponse) []equipmentLocationEventRow {
	rows := make([]equipmentLocationEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildEquipmentLocationEventRow(resource))
	}
	return rows
}

func buildEquipmentLocationEventRow(resource jsonAPIResource) equipmentLocationEventRow {
	row := equipmentLocationEventRow{
		ID:             resource.ID,
		EventAt:        formatDateTime(stringAttr(resource.Attributes, "event-at")),
		EventLatitude:  stringAttr(resource.Attributes, "event-latitude"),
		EventLongitude: stringAttr(resource.Attributes, "event-longitude"),
		Provenance:     stringAttr(resource.Attributes, "provenance"),
	}

	if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
		row.EquipmentID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["updated-by"]; ok && rel.Data != nil {
		row.UpdatedByID = rel.Data.ID
	}

	return row
}

func renderEquipmentLocationEventsTable(cmd *cobra.Command, rows []equipmentLocationEventRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment location events found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEQUIPMENT\tEVENT AT\tLAT\tLON\tPROVENANCE\tUPDATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EquipmentID,
			row.EventAt,
			row.EventLatitude,
			row.EventLongitude,
			row.Provenance,
			row.UpdatedByID,
		)
	}
	return writer.Flush()
}
