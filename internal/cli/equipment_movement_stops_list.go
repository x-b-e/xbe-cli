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

type equipmentMovementStopsListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	Trip                  string
	Location              string
	ScheduledArrivalAtMin string
	ScheduledArrivalAtMax string
	CreatedAtMin          string
	CreatedAtMax          string
	UpdatedAtMin          string
	UpdatedAtMax          string
}

type equipmentMovementStopRow struct {
	ID                 string `json:"id"`
	TripID             string `json:"trip_id,omitempty"`
	LocationID         string `json:"location_id,omitempty"`
	LocationName       string `json:"location,omitempty"`
	ScheduledArrivalAt string `json:"scheduled_arrival_at,omitempty"`
	SequencePosition   string `json:"sequence_position,omitempty"`
	SequenceIndex      string `json:"sequence_index,omitempty"`
}

func newEquipmentMovementStopsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment movement stops",
		Long: `List equipment movement stops.

Output Columns:
  ID         Stop identifier
  TRIP       Trip ID
  LOCATION   Location name or ID
  ARRIVAL    Scheduled arrival time
  SEQ POS    Sequence position
  SEQ IDX    Sequence index

Filters:
  --trip                      Filter by trip ID
  --location                  Filter by location ID
  --scheduled-arrival-at-min  Filter by scheduled arrival on/after (ISO 8601)
  --scheduled-arrival-at-max  Filter by scheduled arrival on/before (ISO 8601)
  --created-at-min            Filter by created-at on/after (ISO 8601)
  --created-at-max            Filter by created-at on/before (ISO 8601)
  --updated-at-min            Filter by updated-at on/after (ISO 8601)
  --updated-at-max            Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List stops
  xbe view equipment-movement-stops list

  # Filter by trip
  xbe view equipment-movement-stops list --trip 123

  # Filter by scheduled arrival
  xbe view equipment-movement-stops list --scheduled-arrival-at-min 2025-01-01T00:00:00Z

  # Output as JSON
  xbe view equipment-movement-stops list --json`,
		Args: cobra.NoArgs,
		RunE: runEquipmentMovementStopsList,
	}
	initEquipmentMovementStopsListFlags(cmd)
	return cmd
}

func init() {
	equipmentMovementStopsCmd.AddCommand(newEquipmentMovementStopsListCmd())
}

func initEquipmentMovementStopsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("trip", "", "Filter by trip ID")
	cmd.Flags().String("location", "", "Filter by location ID")
	cmd.Flags().String("scheduled-arrival-at-min", "", "Filter by scheduled arrival on/after (ISO 8601)")
	cmd.Flags().String("scheduled-arrival-at-max", "", "Filter by scheduled arrival on/before (ISO 8601)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentMovementStopsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentMovementStopsListOptions(cmd)
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
	query.Set("fields[equipment-movement-stops]", "sequence-position,scheduled-arrival-at,sequence-index,trip,location")
	query.Set("include", "location")
	query.Set("fields[equipment-movement-requirement-locations]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[trip]", opts.Trip)
	setFilterIfPresent(query, "filter[location]", opts.Location)
	setFilterIfPresent(query, "filter[scheduled-arrival-at-min]", opts.ScheduledArrivalAtMin)
	setFilterIfPresent(query, "filter[scheduled-arrival-at-max]", opts.ScheduledArrivalAtMax)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-movement-stops", query)
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

	rows := buildEquipmentMovementStopRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentMovementStopsTable(cmd, rows)
}

func parseEquipmentMovementStopsListOptions(cmd *cobra.Command) (equipmentMovementStopsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	trip, _ := cmd.Flags().GetString("trip")
	location, _ := cmd.Flags().GetString("location")
	scheduledArrivalAtMin, _ := cmd.Flags().GetString("scheduled-arrival-at-min")
	scheduledArrivalAtMax, _ := cmd.Flags().GetString("scheduled-arrival-at-max")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentMovementStopsListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		Trip:                  trip,
		Location:              location,
		ScheduledArrivalAtMin: scheduledArrivalAtMin,
		ScheduledArrivalAtMax: scheduledArrivalAtMax,
		CreatedAtMin:          createdAtMin,
		CreatedAtMax:          createdAtMax,
		UpdatedAtMin:          updatedAtMin,
		UpdatedAtMax:          updatedAtMax,
	}, nil
}

func buildEquipmentMovementStopRows(resp jsonAPIResponse) []equipmentMovementStopRow {
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		if inc.Attributes == nil {
			continue
		}
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	rows := make([]equipmentMovementStopRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentMovementStopRow{
			ID:                 resource.ID,
			ScheduledArrivalAt: formatDateTime(stringAttr(resource.Attributes, "scheduled-arrival-at")),
			SequencePosition:   stringAttr(resource.Attributes, "sequence-position"),
			SequenceIndex:      stringAttr(resource.Attributes, "sequence-index"),
		}

		if rel, ok := resource.Relationships["trip"]; ok && rel.Data != nil {
			row.TripID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["location"]; ok && rel.Data != nil {
			row.LocationID = rel.Data.ID
			if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.LocationName = stringAttr(attrs, "name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func buildEquipmentMovementStopRowFromSingle(resp jsonAPISingleResponse) equipmentMovementStopRow {
	row := equipmentMovementStopRow{
		ID:                 resp.Data.ID,
		ScheduledArrivalAt: formatDateTime(stringAttr(resp.Data.Attributes, "scheduled-arrival-at")),
		SequencePosition:   stringAttr(resp.Data.Attributes, "sequence-position"),
		SequenceIndex:      stringAttr(resp.Data.Attributes, "sequence-index"),
	}

	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		if inc.Attributes == nil {
			continue
		}
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	if rel, ok := resp.Data.Relationships["trip"]; ok && rel.Data != nil {
		row.TripID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["location"]; ok && rel.Data != nil {
		row.LocationID = rel.Data.ID
		if attrs, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.LocationName = stringAttr(attrs, "name")
		}
	}

	return row
}

func renderEquipmentMovementStopsTable(cmd *cobra.Command, rows []equipmentMovementStopRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment movement stops found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRIP\tLOCATION\tARRIVAL\tSEQ POS\tSEQ IDX")
	for _, row := range rows {
		location := firstNonEmpty(row.LocationName, row.LocationID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TripID,
			truncateString(location, 30),
			row.ScheduledArrivalAt,
			row.SequencePosition,
			row.SequenceIndex,
		)
	}
	return writer.Flush()
}
