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

type driverMovementSegmentsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	IsMoving                 string
	IsStationary             string
	AtJobSite                string
	AtMaterialSite           string
	AtParkingSite            string
	SiteKind                 string
	DriverMovementSegmentSet string
	Trailer                  string
	Tractor                  string
	StartAtMin               string
	StartAtMax               string
	EndAtMin                 string
	EndAtMax                 string
}

type driverMovementSegmentRow struct {
	ID                         string `json:"id"`
	SequenceIndex              int    `json:"sequence_index"`
	StartAt                    string `json:"start_at,omitempty"`
	EndAt                      string `json:"end_at,omitempty"`
	MetersTravelled            string `json:"meters_travelled,omitempty"`
	IsMoving                   bool   `json:"is_moving"`
	SiteKind                   string `json:"site_kind,omitempty"`
	DriverMovementSegmentSetID string `json:"driver_movement_segment_set_id,omitempty"`
	DriverDayID                string `json:"driver_day_id,omitempty"`
	DriverID                   string `json:"driver_id,omitempty"`
	SiteType                   string `json:"site_type,omitempty"`
	SiteID                     string `json:"site_id,omitempty"`
	TrailerID                  string `json:"trailer_id,omitempty"`
	TractorID                  string `json:"tractor_id,omitempty"`
}

func newDriverMovementSegmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver movement segments",
		Long: `List driver movement segments with filtering and pagination.

Driver movement segments capture contiguous moving or stationary intervals
for a driver day.

Output Columns:
  ID       Movement segment identifier
  INDEX    Sequence index in the segment set
  START    Segment start timestamp
  END      Segment end timestamp
  MOVING   Whether the segment is moving
  METERS   Distance travelled in meters
  SITE     Site kind or site reference

Filters:
  --is-moving                     Filter by moving status (true/false)
  --is-stationary                 Filter by stationary status (true/false)
  --at-job-site                   Filter by job site presence (true/false)
  --at-material-site              Filter by material site presence (true/false)
  --at-parking-site               Filter by parking site presence (true/false)
  --site-kind                     Filter by site kind
  --driver-movement-segment-set   Filter by segment set ID
  --trailer                       Filter by trailer ID
  --tractor                       Filter by tractor ID
  --start-at-min                  Filter by start-at on/after (ISO 8601)
  --start-at-max                  Filter by start-at on/before (ISO 8601)
  --end-at-min                    Filter by end-at on/after (ISO 8601)
  --end-at-max                    Filter by end-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List movement segments
  xbe view driver-movement-segments list

  # Filter moving segments
  xbe view driver-movement-segments list --is-moving true

  # Filter by segment set
  xbe view driver-movement-segments list --driver-movement-segment-set 123

  # Filter by time range
  xbe view driver-movement-segments list --start-at-min 2025-01-01T00:00:00Z --end-at-max 2025-01-02T00:00:00Z

  # Output as JSON
  xbe view driver-movement-segments list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverMovementSegmentsList,
	}
	initDriverMovementSegmentsListFlags(cmd)
	return cmd
}

func init() {
	driverMovementSegmentsCmd.AddCommand(newDriverMovementSegmentsListCmd())
}

func initDriverMovementSegmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("is-moving", "", "Filter by moving status (true/false)")
	cmd.Flags().String("is-stationary", "", "Filter by stationary status (true/false)")
	cmd.Flags().String("at-job-site", "", "Filter by job site presence (true/false)")
	cmd.Flags().String("at-material-site", "", "Filter by material site presence (true/false)")
	cmd.Flags().String("at-parking-site", "", "Filter by parking site presence (true/false)")
	cmd.Flags().String("site-kind", "", "Filter by site kind")
	cmd.Flags().String("driver-movement-segment-set", "", "Filter by segment set ID")
	cmd.Flags().String("trailer", "", "Filter by trailer ID")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("start-at-min", "", "Filter by start-at on/after (ISO 8601)")
	cmd.Flags().String("start-at-max", "", "Filter by start-at on/before (ISO 8601)")
	cmd.Flags().String("end-at-min", "", "Filter by end-at on/after (ISO 8601)")
	cmd.Flags().String("end-at-max", "", "Filter by end-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverMovementSegmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverMovementSegmentsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[driver-movement-segments]", "sequence-index,start-at,end-at,is-moving,meters-travelled,site-kind")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[is_moving]", opts.IsMoving)
	if opts.IsMoving == "" && opts.IsStationary != "" {
		if value, ok := parseBoolish(opts.IsStationary); ok {
			if value {
				query.Set("filter[is_moving]", "false")
			} else {
				query.Set("filter[is_moving]", "true")
			}
		} else {
			setFilterIfPresent(query, "filter[is_stationary]", opts.IsStationary)
		}
	}

	setFilterIfPresent(query, "filter[site_kind]", opts.SiteKind)
	if opts.SiteKind == "" {
		siteKinds := []string{}
		if value, ok := parseBoolish(opts.AtJobSite); ok && value {
			siteKinds = append(siteKinds, "job_site")
		}
		if value, ok := parseBoolish(opts.AtMaterialSite); ok && value {
			siteKinds = append(siteKinds, "material_site")
		}
		if value, ok := parseBoolish(opts.AtParkingSite); ok && value {
			siteKinds = append(siteKinds, "parking_site")
		}
		if len(siteKinds) > 0 {
			query.Set("filter[site_kind]", strings.Join(siteKinds, ","))
		}
	}
	setFilterIfPresent(query, "filter[driver_movement_segment_set]", opts.DriverMovementSegmentSet)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[start-at-min]", opts.StartAtMin)
	setFilterIfPresent(query, "filter[start-at-max]", opts.StartAtMax)
	setFilterIfPresent(query, "filter[end-at-min]", opts.EndAtMin)
	setFilterIfPresent(query, "filter[end-at-max]", opts.EndAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-movement-segments", query)
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

	rows := buildDriverMovementSegmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverMovementSegmentsTable(cmd, rows)
}

func parseDriverMovementSegmentsListOptions(cmd *cobra.Command) (driverMovementSegmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	isMoving, _ := cmd.Flags().GetString("is-moving")
	isStationary, _ := cmd.Flags().GetString("is-stationary")
	atJobSite, _ := cmd.Flags().GetString("at-job-site")
	atMaterialSite, _ := cmd.Flags().GetString("at-material-site")
	atParkingSite, _ := cmd.Flags().GetString("at-parking-site")
	siteKind, _ := cmd.Flags().GetString("site-kind")
	driverMovementSegmentSet, _ := cmd.Flags().GetString("driver-movement-segment-set")
	trailer, _ := cmd.Flags().GetString("trailer")
	tractor, _ := cmd.Flags().GetString("tractor")
	startAtMin, _ := cmd.Flags().GetString("start-at-min")
	startAtMax, _ := cmd.Flags().GetString("start-at-max")
	endAtMin, _ := cmd.Flags().GetString("end-at-min")
	endAtMax, _ := cmd.Flags().GetString("end-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverMovementSegmentsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		IsMoving:                 isMoving,
		IsStationary:             isStationary,
		AtJobSite:                atJobSite,
		AtMaterialSite:           atMaterialSite,
		AtParkingSite:            atParkingSite,
		SiteKind:                 siteKind,
		DriverMovementSegmentSet: driverMovementSegmentSet,
		Trailer:                  trailer,
		Tractor:                  tractor,
		StartAtMin:               startAtMin,
		StartAtMax:               startAtMax,
		EndAtMin:                 endAtMin,
		EndAtMax:                 endAtMax,
	}, nil
}

func buildDriverMovementSegmentRows(resp jsonAPIResponse) []driverMovementSegmentRow {
	rows := make([]driverMovementSegmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := driverMovementSegmentRow{
			ID:              resource.ID,
			SequenceIndex:   intAttr(attrs, "sequence-index"),
			StartAt:         formatDateTime(stringAttr(attrs, "start-at")),
			EndAt:           formatDateTime(stringAttr(attrs, "end-at")),
			MetersTravelled: stringAttr(attrs, "meters-travelled"),
			IsMoving:        boolAttr(attrs, "is-moving"),
			SiteKind:        stringAttr(attrs, "site-kind"),
		}

		if rel, ok := resource.Relationships["driver-movement-segment-set"]; ok && rel.Data != nil {
			row.DriverMovementSegmentSetID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
			row.DriverDayID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.DriverID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["site"]; ok && rel.Data != nil {
			row.SiteType = rel.Data.Type
			row.SiteID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["trailer"]; ok && rel.Data != nil {
			row.TrailerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
			row.TractorID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderDriverMovementSegmentsTable(cmd *cobra.Command, rows []driverMovementSegmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver movement segments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINDEX\tSTART\tEND\tMOVING\tMETERS\tSITE")
	for _, row := range rows {
		site := row.SiteKind
		if site == "" && row.SiteType != "" {
			site = row.SiteType
			if row.SiteID != "" {
				site += "/" + row.SiteID
			}
		}
		fmt.Fprintf(writer, "%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.SequenceIndex,
			row.StartAt,
			row.EndAt,
			formatYesNo(row.IsMoving),
			row.MetersTravelled,
			truncateString(site, 30),
		)
	}
	return writer.Flush()
}

func parseBoolish(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "t", "1", "yes", "y":
		return true, true
	case "false", "f", "0", "no", "n":
		return false, true
	default:
		return false, false
	}
}
