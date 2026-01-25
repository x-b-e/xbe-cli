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

type tractorFuelConsumptionReadingsListOptions struct {
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
	ReadingOn     string
	ReadingOnMin  string
	ReadingOnMax  string
	HasReadingOn  string
}

type tractorFuelConsumptionReadingRow struct {
	ID                        string  `json:"id"`
	ReadingOn                 string  `json:"reading_on,omitempty"`
	ReadingTime               string  `json:"reading_time,omitempty"`
	DateSequence              int     `json:"date_sequence,omitempty"`
	StateCode                 string  `json:"state_code,omitempty"`
	Value                     float64 `json:"value,omitempty"`
	TractorID                 string  `json:"tractor_id,omitempty"`
	TractorNumber             string  `json:"tractor,omitempty"`
	DriverDayID               string  `json:"driver_day_id,omitempty"`
	DriverDayStartOn          string  `json:"driver_day_start_on,omitempty"`
	UnitOfMeasureID           string  `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasureName         string  `json:"unit_of_measure,omitempty"`
	UnitOfMeasureAbbreviation string  `json:"unit_of_measure_abbreviation,omitempty"`
}

func newTractorFuelConsumptionReadingsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tractor fuel consumption readings",
		Long: `List tractor fuel consumption readings with filtering and pagination.

Output Columns:
  ID           Reading identifier
  READING ON   Reading date
  TIME         Reading time
  SEQ          Sequence number (per day)
  VALUE        Fuel consumption value
  UNIT         Unit of measure
  TRACTOR      Tractor number

Filters:
  --tractor         Filter by tractor ID (comma-separated for multiple)
  --trucker         Filter by trucker ID (comma-separated for multiple)
  --driver-day      Filter by driver day ID (comma-separated for multiple)
  --unit-of-measure Filter by unit of measure ID (comma-separated for multiple)
  --reading-on      Filter by reading date (YYYY-MM-DD)
  --reading-on-min  Filter by minimum reading date (YYYY-MM-DD)
  --reading-on-max  Filter by maximum reading date (YYYY-MM-DD)
  --has-reading-on  Filter by presence of reading date (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List tractor fuel consumption readings
  xbe view tractor-fuel-consumption-readings list

  # Filter by tractor
  xbe view tractor-fuel-consumption-readings list --tractor 123

  # Filter by reading date range
  xbe view tractor-fuel-consumption-readings list --reading-on-min 2025-01-01 --reading-on-max 2025-01-31

  # Output as JSON
  xbe view tractor-fuel-consumption-readings list --json`,
		Args: cobra.NoArgs,
		RunE: runTractorFuelConsumptionReadingsList,
	}
	initTractorFuelConsumptionReadingsListFlags(cmd)
	return cmd
}

func init() {
	tractorFuelConsumptionReadingsCmd.AddCommand(newTractorFuelConsumptionReadingsListCmd())
}

func initTractorFuelConsumptionReadingsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tractor", "", "Filter by tractor ID (comma-separated for multiple)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (comma-separated for multiple)")
	cmd.Flags().String("driver-day", "", "Filter by driver day ID (comma-separated for multiple)")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID (comma-separated for multiple)")
	cmd.Flags().String("reading-on", "", "Filter by reading date (YYYY-MM-DD)")
	cmd.Flags().String("reading-on-min", "", "Filter by minimum reading date (YYYY-MM-DD)")
	cmd.Flags().String("reading-on-max", "", "Filter by maximum reading date (YYYY-MM-DD)")
	cmd.Flags().String("has-reading-on", "", "Filter by presence of reading date (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTractorFuelConsumptionReadingsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTractorFuelConsumptionReadingsListOptions(cmd)
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
	query.Set("fields[tractor-fuel-consumption-readings]", "reading-on,reading-time,date-sequence,state-code,value,tractor,driver-day,unit-of-measure")
	query.Set("fields[tractors]", "number")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[trucker-shift-sets]", "start-on")
	query.Set("include", "tractor,unit-of-measure,driver-day")

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
	setFilterIfPresent(query, "filter[reading-on]", opts.ReadingOn)
	setFilterIfPresent(query, "filter[reading-on-min]", opts.ReadingOnMin)
	setFilterIfPresent(query, "filter[reading-on-max]", opts.ReadingOnMax)
	setFilterIfPresent(query, "filter[has-reading-on]", opts.HasReadingOn)

	body, _, err := client.Get(cmd.Context(), "/v1/tractor-fuel-consumption-readings", query)
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

	rows := buildTractorFuelConsumptionReadingRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTractorFuelConsumptionReadingsTable(cmd, rows)
}

func parseTractorFuelConsumptionReadingsListOptions(cmd *cobra.Command) (tractorFuelConsumptionReadingsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tractor, _ := cmd.Flags().GetString("tractor")
	trucker, _ := cmd.Flags().GetString("trucker")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	readingOn, _ := cmd.Flags().GetString("reading-on")
	readingOnMin, _ := cmd.Flags().GetString("reading-on-min")
	readingOnMax, _ := cmd.Flags().GetString("reading-on-max")
	hasReadingOn, _ := cmd.Flags().GetString("has-reading-on")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tractorFuelConsumptionReadingsListOptions{
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
		ReadingOn:     readingOn,
		ReadingOnMin:  readingOnMin,
		ReadingOnMax:  readingOnMax,
		HasReadingOn:  hasReadingOn,
	}, nil
}

func buildTractorFuelConsumptionReadingRows(resp jsonAPIResponse) []tractorFuelConsumptionReadingRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]tractorFuelConsumptionReadingRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := tractorFuelConsumptionReadingRow{
			ID:           resource.ID,
			ReadingOn:    formatDate(stringAttr(attrs, "reading-on")),
			ReadingTime:  formatTime(stringAttr(attrs, "reading-time")),
			DateSequence: intAttr(attrs, "date-sequence"),
			StateCode:    stringAttr(attrs, "state-code"),
			Value:        floatAttr(attrs, "value"),
		}

		row.TractorID = relationshipIDFromMap(resource.Relationships, "tractor")
		if row.TractorID != "" {
			if tractor, ok := included[resourceKey("tractors", row.TractorID)]; ok {
				row.TractorNumber = stringAttr(tractor.Attributes, "number")
			}
		}

		row.DriverDayID = relationshipIDFromMap(resource.Relationships, "driver-day")
		if row.DriverDayID != "" {
			if driverDay, ok := included[resourceKey("trucker-shift-sets", row.DriverDayID)]; ok {
				row.DriverDayStartOn = formatDate(stringAttr(driverDay.Attributes, "start-on"))
			}
		}

		row.UnitOfMeasureID = relationshipIDFromMap(resource.Relationships, "unit-of-measure")
		if row.UnitOfMeasureID != "" {
			if uom, ok := included[resourceKey("unit-of-measures", row.UnitOfMeasureID)]; ok {
				row.UnitOfMeasureName = stringAttr(uom.Attributes, "name")
				row.UnitOfMeasureAbbreviation = stringAttr(uom.Attributes, "abbreviation")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderTractorFuelConsumptionReadingsTable(cmd *cobra.Command, rows []tractorFuelConsumptionReadingRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tractor fuel consumption readings found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREADING ON\tTIME\tSEQ\tVALUE\tUNIT\tTRACTOR")
	for _, row := range rows {
		sequence := ""
		if row.DateSequence != 0 {
			sequence = strconv.Itoa(row.DateSequence)
		}
		unit := firstNonEmpty(row.UnitOfMeasureAbbreviation, row.UnitOfMeasureName, row.UnitOfMeasureID)
		tractor := firstNonEmpty(row.TractorNumber, row.TractorID)
		value := fmt.Sprintf("%.2f", row.Value)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ReadingOn,
			row.ReadingTime,
			sequence,
			value,
			unit,
			tractor,
		)
	}

	return writer.Flush()
}
