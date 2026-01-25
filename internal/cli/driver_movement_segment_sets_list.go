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

type driverMovementSegmentSetsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	DriverDay string
	Driver    string
}

type driverMovementSegmentSetRow struct {
	ID                   string `json:"id"`
	Date                 string `json:"date,omitempty"`
	DriverName           string `json:"driver_name,omitempty"`
	DriverID             string `json:"driver_id,omitempty"`
	DriverDayID          string `json:"driver_day_id,omitempty"`
	SegmentsCount        int    `json:"segments_count"`
	MovingSegmentsCount  int    `json:"moving_segments_count"`
	TotalMetersTravelled int    `json:"total_meters_travelled"`
}

func newDriverMovementSegmentSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver movement segment sets",
		Long: `List driver movement segment sets.

Output Columns:
  ID         Driver movement segment set identifier
  DATE       Driver day date
  DRIVER     Driver name (or ID if name unavailable)
  DRIVER DAY Driver day ID
  SEGMENTS   Total segments count
  MOVING     Moving segments count
  METERS     Total meters travelled

Filters:
  --driver-day  Filter by driver day ID
  --driver      Filter by driver ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List driver movement segment sets
  xbe view driver-movement-segment-sets list

  # Filter by driver day
  xbe view driver-movement-segment-sets list --driver-day 123

  # Filter by driver
  xbe view driver-movement-segment-sets list --driver 456

  # Output as JSON
  xbe view driver-movement-segment-sets list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverMovementSegmentSetsList,
	}
	initDriverMovementSegmentSetsListFlags(cmd)
	return cmd
}

func init() {
	driverMovementSegmentSetsCmd.AddCommand(newDriverMovementSegmentSetsListCmd())
}

func initDriverMovementSegmentSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("driver-day", "", "Filter by driver day ID")
	cmd.Flags().String("driver", "", "Filter by driver ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverMovementSegmentSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverMovementSegmentSetsListOptions(cmd)
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
	query.Set("fields[driver-movement-segment-sets]", "date,driver-name,segments-count,moving-segments-count,total-meters-travelled,driver-day,driver")
	query.Set("include", "driver-day,driver")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[driver_day]", opts.DriverDay)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-movement-segment-sets", query)
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

	rows := buildDriverMovementSegmentSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverMovementSegmentSetsTable(cmd, rows)
}

func parseDriverMovementSegmentSetsListOptions(cmd *cobra.Command) (driverMovementSegmentSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	driver, _ := cmd.Flags().GetString("driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverMovementSegmentSetsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		DriverDay: driverDay,
		Driver:    driver,
	}, nil
}

func buildDriverMovementSegmentSetRows(resp jsonAPIResponse) []driverMovementSegmentSetRow {
	rows := make([]driverMovementSegmentSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := driverMovementSegmentSetRow{
			ID:                   resource.ID,
			Date:                 formatDate(stringAttr(attrs, "date")),
			DriverName:           stringAttr(attrs, "driver-name"),
			SegmentsCount:        intAttr(attrs, "segments-count"),
			MovingSegmentsCount:  intAttr(attrs, "moving-segments-count"),
			TotalMetersTravelled: intAttr(attrs, "total-meters-travelled"),
		}
		if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
			row.DriverDayID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
			row.DriverID = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func buildDriverMovementSegmentSetRowFromSingle(resp jsonAPISingleResponse) driverMovementSegmentSetRow {
	return buildDriverMovementSegmentSetRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}})[0]
}

func renderDriverMovementSegmentSetsTable(cmd *cobra.Command, rows []driverMovementSegmentSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver movement segment sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDATE\tDRIVER\tDRIVER DAY\tSEGMENTS\tMOVING\tMETERS")
	for _, row := range rows {
		driverLabel := strings.TrimSpace(row.DriverName)
		if driverLabel == "" {
			driverLabel = row.DriverID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%d\t%d\t%d\n",
			row.ID,
			row.Date,
			truncateString(driverLabel, 24),
			row.DriverDayID,
			row.SegmentsCount,
			row.MovingSegmentsCount,
			row.TotalMetersTravelled,
		)
	}
	writer.Flush()
	return nil
}
