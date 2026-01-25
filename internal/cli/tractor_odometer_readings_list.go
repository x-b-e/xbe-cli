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

type tractorOdometerReadingsListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	Tractor       string
	Trucker       string
	DriverDay     string
	UnitOfMeasure string
	ReadingOnMin  string
	ReadingOnMax  string
	CreatedBy     string
}

type tractorOdometerReadingRow struct {
	ID              string `json:"id"`
	TractorID       string `json:"tractor_id,omitempty"`
	DriverDayID     string `json:"driver_day_id,omitempty"`
	UnitOfMeasureID string `json:"unit_of_measure_id,omitempty"`
	CreatedByID     string `json:"created_by_id,omitempty"`
	DateSequence    string `json:"date_sequence,omitempty"`
	ReadingOn       string `json:"reading_on,omitempty"`
	ReadingTime     string `json:"reading_time,omitempty"`
	StateCode       string `json:"state_code,omitempty"`
	Value           string `json:"value,omitempty"`
}

func newTractorOdometerReadingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tractor odometer readings",
		Long: `List tractor odometer readings with filtering and pagination.

Output Columns:
  ID          Odometer reading ID
  TRACTOR     Tractor ID
  DRIVER DAY  Driver day ID
  READING ON  Reading date
  TIME        Reading time
  VALUE       Odometer reading value
  UOM         Unit of measure ID
  STATE       State code

Filters:
  --tractor          Filter by tractor ID
  --trucker          Filter by trucker ID
  --driver-day       Filter by driver day ID
  --unit-of-measure  Filter by unit of measure ID
  --reading-on-min   Filter by minimum reading date
  --reading-on-max   Filter by maximum reading date
  --created-by       Filter by created-by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List tractor odometer readings
  xbe view tractor-odometer-readings list

  # Filter by tractor
  xbe view tractor-odometer-readings list --tractor 123

  # Filter by trucker
  xbe view tractor-odometer-readings list --trucker 456

  # Filter by reading date range
  xbe view tractor-odometer-readings list --reading-on-min 2025-01-01 --reading-on-max 2025-01-31

  # Output as JSON
  xbe view tractor-odometer-readings list --json`,
		Args: cobra.NoArgs,
		RunE: runTractorOdometerReadingsList,
	}
	initTractorOdometerReadingsListFlags(cmd)
	return cmd
}

func init() {
	tractorOdometerReadingsCmd.AddCommand(newTractorOdometerReadingsListCmd())
}

func initTractorOdometerReadingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tractor", "", "Filter by tractor ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("driver-day", "", "Filter by driver day ID")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID")
	cmd.Flags().String("reading-on-min", "", "Filter by minimum reading date (YYYY-MM-DD)")
	cmd.Flags().String("reading-on-max", "", "Filter by maximum reading date (YYYY-MM-DD)")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTractorOdometerReadingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTractorOdometerReadingsListOptions(cmd)
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
	query.Set("fields[tractor-odometer-readings]", "date-sequence,reading-on,reading-time,state-code,value,tractor,unit-of-measure,driver-day,created-by")
	query.Set("include", "tractor,unit-of-measure,driver-day,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[driver_day]", opts.DriverDay)
	setFilterIfPresent(query, "filter[unit_of_measure]", opts.UnitOfMeasure)
	setFilterIfPresent(query, "filter[reading_on_min]", opts.ReadingOnMin)
	setFilterIfPresent(query, "filter[reading_on_max]", opts.ReadingOnMax)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/tractor-odometer-readings", query)
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

	rows := buildTractorOdometerReadingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTractorOdometerReadingsTable(cmd, rows)
}

func parseTractorOdometerReadingsListOptions(cmd *cobra.Command) (tractorOdometerReadingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tractor, _ := cmd.Flags().GetString("tractor")
	trucker, _ := cmd.Flags().GetString("trucker")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	readingOnMin, _ := cmd.Flags().GetString("reading-on-min")
	readingOnMax, _ := cmd.Flags().GetString("reading-on-max")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tractorOdometerReadingsListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		Tractor:       tractor,
		Trucker:       trucker,
		DriverDay:     driverDay,
		UnitOfMeasure: unitOfMeasure,
		ReadingOnMin:  readingOnMin,
		ReadingOnMax:  readingOnMax,
		CreatedBy:     createdBy,
	}, nil
}

func buildTractorOdometerReadingRows(resp jsonAPIResponse) []tractorOdometerReadingRow {
	rows := make([]tractorOdometerReadingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := tractorOdometerReadingRow{
			ID:           resource.ID,
			DateSequence: stringAttr(attrs, "date-sequence"),
			ReadingOn:    formatDate(stringAttr(attrs, "reading-on")),
			ReadingTime:  formatTime(stringAttr(attrs, "reading-time")),
			StateCode:    stringAttr(attrs, "state-code"),
			Value:        stringAttr(attrs, "value"),
		}

		if rel, ok := resource.Relationships["tractor"]; ok && rel.Data != nil {
			row.TractorID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
			row.DriverDayID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
			row.UnitOfMeasureID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderTractorOdometerReadingsTable(cmd *cobra.Command, rows []tractorOdometerReadingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tractor odometer readings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRACTOR\tDRIVER DAY\tREADING ON\tTIME\tVALUE\tUOM\tSTATE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TractorID,
			row.DriverDayID,
			row.ReadingOn,
			row.ReadingTime,
			row.Value,
			row.UnitOfMeasureID,
			row.StateCode,
		)
	}
	return writer.Flush()
}
