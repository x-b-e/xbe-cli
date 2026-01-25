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

type truckerShiftSetsListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Sort                         string
	Trucker                      string
	Trailer                      string
	TrailerClassification        string
	Tractor                      string
	Driver                       string
	DriverName                   string
	DriverID                     string
	HasDriver                    string
	Broker                       string
	BrokerID                     string
	BusinessUnit                 string
	Shifts                       string
	Customer                     string
	NumberOfShiftsEq             string
	NumberOfShiftsGte            string
	IsExpectingTimeSheet         string
	WithoutTimeCard              string
	WithoutApprovedTimeCard      string
	WithMissingTimeCardApprovals string
	WithoutApprovedTimeSheet     string
	WithoutSubmittedTimeSheet    string
	StartAt                      string
	StartOn                      string
	StartOnMin                   string
	StartOnMax                   string
	HasStartOn                   string
	DriverDayOn                  string
	HasConstraint                string
	OdometerStartValue           string
	OdometerStartValueMin        string
	OdometerStartValueMax        string
	OdometerEndValue             string
	OdometerEndValueMin          string
	OdometerEndValueMax          string
}

type truckerShiftSetRow struct {
	ID                 string `json:"id"`
	StartOn            string `json:"start_on,omitempty"`
	EarliestStartAt    string `json:"earliest_start_at,omitempty"`
	Trucker            string `json:"trucker,omitempty"`
	TruckerID          string `json:"trucker_id,omitempty"`
	Driver             string `json:"driver,omitempty"`
	DriverID           string `json:"driver_id,omitempty"`
	ShiftCount         int    `json:"shift_count,omitempty"`
	IsTimeSheetEnabled bool   `json:"is_time_sheet_enabled"`
	IsManaged          bool   `json:"is_managed"`
}

func newTruckerShiftSetsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker shift sets",
		Long: `List trucker shift sets (driver days) with filtering and pagination.

Output Columns:
  ID        Trucker shift set identifier
  START ON  Shift set date
  EARLIEST  Earliest shift start time
  TRUCKER   Trucker company name
  DRIVER    Driver name
  SHIFTS    Number of shifts in the set
  TIME SHEET  Time sheet enabled
  MANAGED   Managed status

Filters:
  --trucker                      Filter by trucker ID (comma-separated for multiple)
  --trailer                      Filter by trailer ID (comma-separated for multiple)
  --trailer-classification       Filter by trailer classification ID (comma-separated for multiple)
  --tractor                      Filter by tractor ID (comma-separated for multiple)
  --driver                       Filter by driver user ID (comma-separated for multiple)
  --driver-name                  Filter by driver name
  --driver-id                    Filter by driver user ID (comma-separated for multiple)
  --has-driver                   Filter by presence of a driver (true/false)
  --broker                       Filter by broker ID (comma-separated for multiple)
  --broker-id                    Filter by broker ID (comma-separated for multiple)
  --business-unit                Filter by business unit ID (comma-separated for multiple)
  --shifts                       Filter by tender job schedule shift IDs (comma-separated for multiple)
  --customer                     Filter by customer ID (comma-separated for multiple)
  --number-of-shifts-eq           Filter by exact number of shifts
  --number-of-shifts-gte          Filter by minimum number of shifts
  --is-expecting-time-sheet       Filter by time sheet expectation (true/false)
  --without-time-card             Filter by absence of time cards (true/false)
  --without-approved-time-card    Filter by missing approved time cards (true/false)
  --with-missing-time-card-approvals  Filter by missing time card approvals (true/false)
  --without-approved-time-sheet   Filter by missing approved time sheets (true/false)
  --without-submitted-time-sheet  Filter by missing submitted time sheets (true/false)
  --start-at                      Filter by shift set start-at timestamp (ISO 8601)
  --start-on                      Filter by shift set date (YYYY-MM-DD)
  --start-on-min                  Filter by minimum shift set date (YYYY-MM-DD)
  --start-on-max                  Filter by maximum shift set date (YYYY-MM-DD)
  --has-start-on                  Filter by presence of shift set date (true/false)
  --driver-day-on                 Filter by driver day date (YYYY-MM-DD)
  --has-constraint                Filter by presence of constraints (true/false)
  --odometer-start-value          Filter by odometer start value (exact)
  --odometer-start-value-min      Filter by minimum odometer start value
  --odometer-start-value-max      Filter by maximum odometer start value
  --odometer-end-value            Filter by odometer end value (exact)
  --odometer-end-value-min        Filter by minimum odometer end value
  --odometer-end-value-max        Filter by maximum odometer end value

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List trucker shift sets
  xbe view trucker-shift-sets list

  # Filter by trucker
  xbe view trucker-shift-sets list --trucker 123

  # Filter by date range
  xbe view trucker-shift-sets list --start-on-min 2025-01-01 --start-on-max 2025-01-31

  # Output as JSON
  xbe view trucker-shift-sets list --json`,
		RunE: runTruckerShiftSetsList,
	}
	initTruckerShiftSetsListFlags(cmd)
	return cmd
}

func init() {
	truckerShiftSetsCmd.AddCommand(newTruckerShiftSetsListCmd())
}

func initTruckerShiftSetsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("trucker", "", "Filter by trucker ID (comma-separated for multiple)")
	cmd.Flags().String("trailer", "", "Filter by trailer ID (comma-separated for multiple)")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification ID (comma-separated for multiple)")
	cmd.Flags().String("tractor", "", "Filter by tractor ID (comma-separated for multiple)")
	cmd.Flags().String("driver", "", "Filter by driver user ID (comma-separated for multiple)")
	cmd.Flags().String("driver-name", "", "Filter by driver name")
	cmd.Flags().String("driver-id", "", "Filter by driver user ID (comma-separated for multiple)")
	cmd.Flags().String("has-driver", "", "Filter by presence of a driver (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID (comma-separated for multiple)")
	cmd.Flags().String("shifts", "", "Filter by tender job schedule shift IDs (comma-separated for multiple)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("number-of-shifts-eq", "", "Filter by exact number of shifts")
	cmd.Flags().String("number-of-shifts-gte", "", "Filter by minimum number of shifts")
	cmd.Flags().String("is-expecting-time-sheet", "", "Filter by time sheet expectation (true/false)")
	cmd.Flags().String("without-time-card", "", "Filter by absence of time cards (true/false)")
	cmd.Flags().String("without-approved-time-card", "", "Filter by missing approved time cards (true/false)")
	cmd.Flags().String("with-missing-time-card-approvals", "", "Filter by missing time card approvals (true/false)")
	cmd.Flags().String("without-approved-time-sheet", "", "Filter by missing approved time sheets (true/false)")
	cmd.Flags().String("without-submitted-time-sheet", "", "Filter by missing submitted time sheets (true/false)")
	cmd.Flags().String("start-at", "", "Filter by shift set start-at timestamp (ISO 8601)")
	cmd.Flags().String("start-on", "", "Filter by shift set date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-min", "", "Filter by minimum shift set date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by maximum shift set date (YYYY-MM-DD)")
	cmd.Flags().String("has-start-on", "", "Filter by presence of shift set date (true/false)")
	cmd.Flags().String("driver-day-on", "", "Filter by driver day date (YYYY-MM-DD)")
	cmd.Flags().String("has-constraint", "", "Filter by presence of constraints (true/false)")
	cmd.Flags().String("odometer-start-value", "", "Filter by odometer start value (exact)")
	cmd.Flags().String("odometer-start-value-min", "", "Filter by minimum odometer start value")
	cmd.Flags().String("odometer-start-value-max", "", "Filter by maximum odometer start value")
	cmd.Flags().String("odometer-end-value", "", "Filter by odometer end value (exact)")
	cmd.Flags().String("odometer-end-value-min", "", "Filter by minimum odometer end value")
	cmd.Flags().String("odometer-end-value-max", "", "Filter by maximum odometer end value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerShiftSetsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerShiftSetsListOptions(cmd)
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
	query.Set("fields[trucker-shift-sets]", "start-on,earliest-start-at,ordered-shift-ids,is-time-sheet-enabled,is-managed,trucker,driver")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("include", "trucker,driver")

	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trailer]", opts.Trailer)
	setFilterIfPresent(query, "filter[trailer-classification]", opts.TrailerClassification)
	setFilterIfPresent(query, "filter[tractor]", opts.Tractor)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[driver-name]", opts.DriverName)
	setFilterIfPresent(query, "filter[driver-id]", opts.DriverID)
	setFilterIfPresent(query, "filter[has-driver]", opts.HasDriver)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[shifts]", opts.Shifts)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[number-of-shifts-eq]", opts.NumberOfShiftsEq)
	setFilterIfPresent(query, "filter[number-of-shifts-gte]", opts.NumberOfShiftsGte)
	setFilterIfPresent(query, "filter[is-expecting-time-sheet]", opts.IsExpectingTimeSheet)
	setFilterIfPresent(query, "filter[without-time-card]", opts.WithoutTimeCard)
	setFilterIfPresent(query, "filter[without-approved-time-card]", opts.WithoutApprovedTimeCard)
	setFilterIfPresent(query, "filter[with-missing-time-card-approvals]", opts.WithMissingTimeCardApprovals)
	setFilterIfPresent(query, "filter[without-approved-time-sheet]", opts.WithoutApprovedTimeSheet)
	setFilterIfPresent(query, "filter[without-submitted-time-sheet]", opts.WithoutSubmittedTimeSheet)
	setFilterIfPresent(query, "filter[start-at]", opts.StartAt)
	setFilterIfPresent(query, "filter[start-on]", opts.StartOn)
	setFilterIfPresent(query, "filter[start-on-min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start-on-max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[has-start-on]", opts.HasStartOn)
	setFilterIfPresent(query, "filter[driver-day-on]", opts.DriverDayOn)
	setFilterIfPresent(query, "filter[has-constraint]", opts.HasConstraint)
	setFilterIfPresent(query, "filter[odometer-start-value]", opts.OdometerStartValue)
	setFilterIfPresent(query, "filter[odometer-start-value-min]", opts.OdometerStartValueMin)
	setFilterIfPresent(query, "filter[odometer-start-value-max]", opts.OdometerStartValueMax)
	setFilterIfPresent(query, "filter[odometer-end-value]", opts.OdometerEndValue)
	setFilterIfPresent(query, "filter[odometer-end-value-min]", opts.OdometerEndValueMin)
	setFilterIfPresent(query, "filter[odometer-end-value-max]", opts.OdometerEndValueMax)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-shift-sets", query)
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

	rows := buildTruckerShiftSetRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckerShiftSetsTable(cmd, rows)
}

func parseTruckerShiftSetsListOptions(cmd *cobra.Command) (truckerShiftSetsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	trucker, _ := cmd.Flags().GetString("trucker")
	trailer, _ := cmd.Flags().GetString("trailer")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	tractor, _ := cmd.Flags().GetString("tractor")
	driver, _ := cmd.Flags().GetString("driver")
	driverName, _ := cmd.Flags().GetString("driver-name")
	driverID, _ := cmd.Flags().GetString("driver-id")
	hasDriver, _ := cmd.Flags().GetString("has-driver")
	broker, _ := cmd.Flags().GetString("broker")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	shifts, _ := cmd.Flags().GetString("shifts")
	customer, _ := cmd.Flags().GetString("customer")
	numberOfShiftsEq, _ := cmd.Flags().GetString("number-of-shifts-eq")
	numberOfShiftsGte, _ := cmd.Flags().GetString("number-of-shifts-gte")
	isExpectingTimeSheet, _ := cmd.Flags().GetString("is-expecting-time-sheet")
	withoutTimeCard, _ := cmd.Flags().GetString("without-time-card")
	withoutApprovedTimeCard, _ := cmd.Flags().GetString("without-approved-time-card")
	withMissingTimeCardApprovals, _ := cmd.Flags().GetString("with-missing-time-card-approvals")
	withoutApprovedTimeSheet, _ := cmd.Flags().GetString("without-approved-time-sheet")
	withoutSubmittedTimeSheet, _ := cmd.Flags().GetString("without-submitted-time-sheet")
	startAt, _ := cmd.Flags().GetString("start-at")
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	hasStartOn, _ := cmd.Flags().GetString("has-start-on")
	driverDayOn, _ := cmd.Flags().GetString("driver-day-on")
	hasConstraint, _ := cmd.Flags().GetString("has-constraint")
	odometerStartValue, _ := cmd.Flags().GetString("odometer-start-value")
	odometerStartValueMin, _ := cmd.Flags().GetString("odometer-start-value-min")
	odometerStartValueMax, _ := cmd.Flags().GetString("odometer-start-value-max")
	odometerEndValue, _ := cmd.Flags().GetString("odometer-end-value")
	odometerEndValueMin, _ := cmd.Flags().GetString("odometer-end-value-min")
	odometerEndValueMax, _ := cmd.Flags().GetString("odometer-end-value-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerShiftSetsListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Sort:                         sort,
		Trucker:                      trucker,
		Trailer:                      trailer,
		TrailerClassification:        trailerClassification,
		Tractor:                      tractor,
		Driver:                       driver,
		DriverName:                   driverName,
		DriverID:                     driverID,
		HasDriver:                    hasDriver,
		Broker:                       broker,
		BrokerID:                     brokerID,
		BusinessUnit:                 businessUnit,
		Shifts:                       shifts,
		Customer:                     customer,
		NumberOfShiftsEq:             numberOfShiftsEq,
		NumberOfShiftsGte:            numberOfShiftsGte,
		IsExpectingTimeSheet:         isExpectingTimeSheet,
		WithoutTimeCard:              withoutTimeCard,
		WithoutApprovedTimeCard:      withoutApprovedTimeCard,
		WithMissingTimeCardApprovals: withMissingTimeCardApprovals,
		WithoutApprovedTimeSheet:     withoutApprovedTimeSheet,
		WithoutSubmittedTimeSheet:    withoutSubmittedTimeSheet,
		StartAt:                      startAt,
		StartOn:                      startOn,
		StartOnMin:                   startOnMin,
		StartOnMax:                   startOnMax,
		HasStartOn:                   hasStartOn,
		DriverDayOn:                  driverDayOn,
		HasConstraint:                hasConstraint,
		OdometerStartValue:           odometerStartValue,
		OdometerStartValueMin:        odometerStartValueMin,
		OdometerStartValueMax:        odometerStartValueMax,
		OdometerEndValue:             odometerEndValue,
		OdometerEndValueMin:          odometerEndValueMin,
		OdometerEndValueMax:          odometerEndValueMax,
	}, nil
}

func buildTruckerShiftSetRows(resp jsonAPIResponse) []truckerShiftSetRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]truckerShiftSetRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := truckerShiftSetRow{
			ID:                 resource.ID,
			StartOn:            formatDate(stringAttr(attrs, "start-on")),
			EarliestStartAt:    formatDateTime(stringAttr(attrs, "earliest-start-at")),
			IsTimeSheetEnabled: boolAttr(attrs, "is-time-sheet-enabled"),
			IsManaged:          boolAttr(attrs, "is-managed"),
		}
		orderedShiftIDs := stringSliceAttr(attrs, "ordered-shift-ids")
		row.ShiftCount = len(orderedShiftIDs)

		row.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
		row.DriverID = relationshipIDFromMap(resource.Relationships, "driver")

		if row.TruckerID != "" {
			if trucker, ok := included[resourceKey("truckers", row.TruckerID)]; ok {
				row.Trucker = firstNonEmpty(
					stringAttr(trucker.Attributes, "company-name"),
					stringAttr(trucker.Attributes, "name"),
				)
			}
		}

		if row.DriverID != "" {
			if driver, ok := included[resourceKey("users", row.DriverID)]; ok {
				row.Driver = stringAttr(driver.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderTruckerShiftSetsTable(cmd *cobra.Command, rows []truckerShiftSetRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trucker shift sets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTART ON\tEARLIEST\tTRUCKER\tDRIVER\tSHIFTS\tTIME SHEET\tMANAGED")
	for _, row := range rows {
		trucker := row.Trucker
		if trucker == "" {
			trucker = row.TruckerID
		}
		driver := row.Driver
		if driver == "" {
			driver = row.DriverID
		}
		timeSheet := "no"
		if row.IsTimeSheetEnabled {
			timeSheet = "yes"
		}
		managed := "no"
		if row.IsManaged {
			managed = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			row.ID,
			row.StartOn,
			truncateString(row.EarliestStartAt, 16),
			truncateString(trucker, 20),
			truncateString(driver, 20),
			row.ShiftCount,
			timeSheet,
			managed,
		)
	}
	return writer.Flush()
}
