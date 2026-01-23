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

type equipmentLocationEstimatesListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Equipment            string
	AsOf                 string
	EarliestEventAt      string
	LatestEventAt        string
	MaxAbsLatencySeconds string
	MaxLatestSeconds     string
}

type equipmentLocationEstimateRow struct {
	ID                      string `json:"id"`
	EquipmentID             string `json:"equipment_id"`
	AsOf                    string `json:"as_of,omitempty"`
	LastKnownLatitude       string `json:"last_known_latitude,omitempty"`
	LastKnownLongitude      string `json:"last_known_longitude,omitempty"`
	LastKnownTimeZoneID     string `json:"last_known_time_zone_id,omitempty"`
	LastKnownAt             string `json:"last_known_at,omitempty"`
	LastKnownLatencySeconds string `json:"last_known_latency_seconds,omitempty"`
	LastKnownSourceClass    string `json:"last_known_source_class_name,omitempty"`
}

func newEquipmentLocationEstimatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List equipment location estimates",
		Long: `List equipment location estimates.

Equipment location estimates calculate the most recent known location for
specified equipment based on location events and movement stop completions.

Output Columns:
  ID            Estimate identifier
  EQUIPMENT ID  Equipment identifier
  AS OF         Timestamp used for the estimate
  LAST KNOWN AT Timestamp of the last known location
  LAT           Last known latitude
  LNG           Last known longitude
  LATENCY       Latency in seconds between as-of and last known event
  SOURCE        Source class name for the last known location
  TZ            Time zone ID for the last known location

Filters:
  --equipment                 Equipment ID (required)
  --as-of                     Estimate time (RFC3339)
  --earliest-event-at         Earliest event time (RFC3339)
  --latest-event-at           Latest event time (RFC3339)
  --max-abs-latency-seconds   Maximum absolute latency in seconds
  --max-latest-seconds        Maximum seconds added for latest event window`,
		Example: `  # Estimate location for an equipment ID
  xbe view equipment-location-estimates list --equipment 123

  # Estimate location as of a specific time
  xbe view equipment-location-estimates list --equipment 123 --as-of 2026-01-23T12:00:00Z

  # Constrain the event window
  xbe view equipment-location-estimates list --equipment 123 \\
    --earliest-event-at 2026-01-22T00:00:00Z \\
    --latest-event-at 2026-01-23T00:00:00Z

  # Output as JSON
  xbe view equipment-location-estimates list --equipment 123 --json`,
		RunE: runEquipmentLocationEstimatesList,
	}
	initEquipmentLocationEstimatesListFlags(cmd)
	return cmd
}

func init() {
	equipmentLocationEstimatesCmd.AddCommand(newEquipmentLocationEstimatesListCmd())
}

func initEquipmentLocationEstimatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("equipment", "", "Equipment ID (required)")
	cmd.Flags().String("as-of", "", "Estimate time (RFC3339)")
	cmd.Flags().String("earliest-event-at", "", "Earliest event time (RFC3339)")
	cmd.Flags().String("latest-event-at", "", "Latest event time (RFC3339)")
	cmd.Flags().String("max-abs-latency-seconds", "", "Maximum absolute latency in seconds")
	cmd.Flags().String("max-latest-seconds", "", "Maximum seconds added for latest event window")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runEquipmentLocationEstimatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseEquipmentLocationEstimatesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Equipment) == "" {
		err := fmt.Errorf("--equipment is required")
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
	query.Set("fields[equipment-location-estimates]", "as-of,last-known-latitude,last-known-longitude,last-known-time-zone-id,last-known-at,last-known-latency-seconds,last-known-source-class-name,equipment")
	query.Set("filter[equipment]", opts.Equipment)

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[as_of]", opts.AsOf)
	setFilterIfPresent(query, "filter[earliest_event_at]", opts.EarliestEventAt)
	setFilterIfPresent(query, "filter[latest_event_at]", opts.LatestEventAt)
	setFilterIfPresent(query, "filter[max_abs_latency_seconds]", opts.MaxAbsLatencySeconds)
	setFilterIfPresent(query, "filter[max_latest_seconds]", opts.MaxLatestSeconds)

	body, _, err := client.Get(cmd.Context(), "/v1/equipment-location-estimates", query)
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

	rows := buildEquipmentLocationEstimateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderEquipmentLocationEstimatesTable(cmd, rows)
}

func parseEquipmentLocationEstimatesListOptions(cmd *cobra.Command) (equipmentLocationEstimatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	equipment, _ := cmd.Flags().GetString("equipment")
	asOf, _ := cmd.Flags().GetString("as-of")
	earliestEventAt, _ := cmd.Flags().GetString("earliest-event-at")
	latestEventAt, _ := cmd.Flags().GetString("latest-event-at")
	maxAbsLatencySeconds, _ := cmd.Flags().GetString("max-abs-latency-seconds")
	maxLatestSeconds, _ := cmd.Flags().GetString("max-latest-seconds")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return equipmentLocationEstimatesListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Equipment:            equipment,
		AsOf:                 asOf,
		EarliestEventAt:      earliestEventAt,
		LatestEventAt:        latestEventAt,
		MaxAbsLatencySeconds: maxAbsLatencySeconds,
		MaxLatestSeconds:     maxLatestSeconds,
	}, nil
}

func buildEquipmentLocationEstimateRows(resp jsonAPIResponse) []equipmentLocationEstimateRow {
	rows := make([]equipmentLocationEstimateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := equipmentLocationEstimateRow{
			ID:                      resource.ID,
			AsOf:                    stringAttr(resource.Attributes, "as-of"),
			LastKnownLatitude:       stringAttr(resource.Attributes, "last-known-latitude"),
			LastKnownLongitude:      stringAttr(resource.Attributes, "last-known-longitude"),
			LastKnownTimeZoneID:     stringAttr(resource.Attributes, "last-known-time-zone-id"),
			LastKnownAt:             stringAttr(resource.Attributes, "last-known-at"),
			LastKnownLatencySeconds: stringAttr(resource.Attributes, "last-known-latency-seconds"),
			LastKnownSourceClass:    stringAttr(resource.Attributes, "last-known-source-class-name"),
		}

		if rel, ok := resource.Relationships["equipment"]; ok && rel.Data != nil {
			row.EquipmentID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderEquipmentLocationEstimatesTable(cmd *cobra.Command, rows []equipmentLocationEstimateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No equipment location estimates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEQUIPMENT ID\tAS OF\tLAST KNOWN AT\tLAT\tLNG\tLATENCY\tSOURCE\tTZ")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EquipmentID,
			truncateString(row.AsOf, 19),
			truncateString(row.LastKnownAt, 19),
			truncateString(row.LastKnownLatitude, 12),
			truncateString(row.LastKnownLongitude, 12),
			truncateString(row.LastKnownLatencySeconds, 8),
			truncateString(row.LastKnownSourceClass, 18),
			truncateString(row.LastKnownTimeZoneID, 12),
		)
	}

	return writer.Flush()
}
