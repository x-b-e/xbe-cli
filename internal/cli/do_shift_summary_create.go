package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

var shiftSummaryDefaultMetrics = []string{
	"shift_count",
	"hours_sum",
	"tons_sum",
	"trip_sum",
}

var shiftSummaryGroupByAllColumns = map[string][]string{
	"tender_job_schedule_shift": {"tender_job_schedule_shift_id"},
	"date":                      {"date"},
	"broker":                    {"broker_id", "broker_name"},
	"business_unit":             {"business_unit_id", "business_unit_name"},
	"month":                     {"month"},
	"dow":                       {"dow"},
	"start_hour":                {"start_hour"},
	"day_or_night":              {"day_or_night"},
	"week":                      {"week"},
	"year":                      {"year"},
	"customer":                  {"customer_id", "customer_name"},
	"contractor":                {"contractor_id", "contractor_name"},
	"trucker":                   {"trucker_id", "trucker_name"},
	"driver":                    {"driver_id", "driver_name"},
	"job_number":                {"job_number"},
	"raw_job_number":            {"raw_job_number"},
	"managed":                   {"managed"},
	"expects_time_cards":        {"expects_time_cards"},
	"planner":                   {"planner_id", "planner_name"},
	"is_stockpiling":            {"is_stockpiling"},
	"job_site_state_code":       {"job_site_state_code"},
	"trailer_classification":    {"trailer_classification"},
	"trailer":                   {"trailer_id", "trailer_number"},
	"has_trips":                 {"has_trips"},
	"driver_day":                {"driver_day_id"},
}

var shiftSummaryGroupByDisplayColumns = map[string][]string{
	"tender_job_schedule_shift": {"tender_job_schedule_shift_id"},
	"date":                      {"date"},
	"broker":                    {"broker_name"},
	"business_unit":             {"business_unit_name"},
	"month":                     {"month"},
	"dow":                       {"dow"},
	"start_hour":                {"start_hour"},
	"day_or_night":              {"day_or_night"},
	"week":                      {"week"},
	"year":                      {"year"},
	"customer":                  {"customer_name"},
	"contractor":                {"contractor_name"},
	"trucker":                   {"trucker_name"},
	"driver":                    {"driver_name"},
	"job_number":                {"job_number"},
	"raw_job_number":            {"raw_job_number"},
	"managed":                   {"managed"},
	"expects_time_cards":        {"expects_time_cards"},
	"planner":                   {"planner_name"},
	"is_stockpiling":            {"is_stockpiling"},
	"job_site_state_code":       {"job_site_state_code"},
	"trailer_classification":    {"trailer_classification"},
	"trailer":                   {"trailer_number"},
	"has_trips":                 {"has_trips"},
	"driver_day":                {"driver_day_id"},
}

type doShiftSummaryCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	StartOn     string
	EndOn       string
	GroupByRaw  string
	GroupBySet  bool
	SortRaw     string
	SortSet     bool
	Limit       int
	MetricsRaw  string
	MetricsSet  bool
	Metrics     []string
	AllMetrics  bool
	FiltersJSON string
	FilterPairs []string
}

func newDoShiftSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a shift summary",
		Long: `Create a shift summary.

Aggregates driver work shift data including hours worked, trips completed, tons
hauled, revenue, and costs. Use this to analyze driver productivity, time card
metrics, and operational efficiency across your fleet.

REQUIRED: You must provide --start-on and --end-on date flags.

Group-by attributes:
  broker                    Group by broker
  business_unit             Group by business unit
  contractor                Group by contractor
  customer                  Group by customer
  date                      Group by shift date
  day_or_night              Group by day or night shift
  dow                       Group by day of week (0=Sunday)
  driver                    Group by driver
  driver_day                Group by driver day
  expects_time_cards        Group by time card expectation
  has_trips                 Group by whether shift has trips
  is_stockpiling            Group by stockpiling status
  job_number                Group by job number
  job_site_state_code       Group by job site state
  managed                   Group by managed status
  month                     Group by month
  planner                   Group by planner
  raw_job_number            Group by raw job number
  start_hour                Group by start hour
  tender_job_schedule_shift Group by tender job schedule shift
  trailer                   Group by trailer
  trailer_classification    Group by trailer classification
  trucker                   Group by trucker
  week                      Group by week
  year                      Group by year

Metrics:
  By default: shift_count, hours_sum, tons_sum, trip_sum

  Available metrics include:
    shift_count, hours_sum, gps_tracked_hours_pct, driver_days, shift_day_count,
    hours_per_driver_day, time_card_submission_latency_median,
    time_card_submission_latency_exceeds_twelve_hours_pct,
    time_card_approval_latency_median, time_card_creation_pct,
    time_card_submission_pct, time_card_pre_approval_latency_median,
    time_card_pre_approval_pct, approved_via_pre_approval_pct,
    time_card_overall_approval_pct, time_card_audit_pct, check_in_pct,
    driver_check_in_pct, timely_check_in_pct, check_in_timeliness_score_avg,
    check_in_directness_score_avg, check_in_distance_miles_avg,
    check_in_nearby_score_avg, on_time_score_avg,
    check_in_minutes_from_start_at_avg, revenue_sum, revenue_shortfall_sum,
    cost_sum, cost_shortfall_sum, margin_sum, margin_pct, tons_sum, trip_sum,
    tons_per_trip_avg, trip_minutes_sum, trip_minutes_avg, trip_miles_sum,
    trip_miles_avg, cost_per_ton, cost_per_hour, trip_minutes_pct,
    tons_per_truck_hour, production_incident_count,
    production_incident_duration_hours, production_incident_duration_hours_avg,
    production_incident_impact_hours, production_incident_duration_pct,
    negative_rating_avg, trips_per_hour_relative_efficiency_pct_avg

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    broker, business_unit, contractor, customer, date, driver, job_number,
    managed, planner, time_card_invoiced, time_card_status, timely_dispatched,
    trucker, trailer_classification, and all group-by attributes`,
		Example: `  # Shift summary grouped by driver
  xbe summarize shift-summary create --start-on 2025-01-01 --end-on 2025-01-31 --group-by driver --filter broker=123

  # Summary by trucker and date
  xbe summarize shift-summary create --start-on 2025-01-01 --end-on 2025-01-31 --group-by trucker,date --filter broker=123

  # Summary with revenue and cost metrics
  xbe summarize shift-summary create --start-on 2025-01-01 --end-on 2025-01-31 --group-by driver --filter broker=123 --metrics shift_count,revenue_sum,cost_sum,margin_sum

  # Total summary (no group-by)
  xbe summarize shift-summary create --start-on 2025-01-01 --end-on 2025-01-31 --group-by "" --filter broker=123

  # JSON output
  xbe summarize shift-summary create --start-on 2025-01-01 --end-on 2025-01-31 --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoShiftSummaryCreate,
	}
	initDoShiftSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doShiftSummaryCmd.AddCommand(newDoShiftSummaryCreateCmd())
}

func initDoShiftSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD) [required]")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD) [required]")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to driver unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to shift_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoShiftSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoShiftSummaryCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			err := errors.New("authentication required. Run 'xbe auth login' first")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	// Validate required fields
	if strings.TrimSpace(opts.StartOn) == "" {
		err := errors.New("--start-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.EndOn) == "" {
		err := errors.New("--end-on is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	filters, err := parseShiftSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolveShiftSummaryGroupBy(opts)
	sort := resolveShiftSummarySort(opts)
	metrics := resolveShiftSummaryMetrics(opts)

	attributes := map[string]any{
		"start-on": opts.StartOn,
		"end-on":   opts.EndOn,
		"filters":  filters,
	}
	if opts.GroupBySet || len(groupBy) > 0 {
		attributes["group-by"] = groupBy
	}
	if opts.SortSet || len(sort) > 0 {
		attributes["sort"] = sort
	}
	if opts.Limit > 0 {
		attributes["limit"] = opts.Limit
	}
	if len(metrics) > 0 {
		attributes["included-metrics"] = metrics
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "shift-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/shift-summaries", jsonBody)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	headers := stringSliceAttr(resp.Data.Attributes, "headers")
	values := valuesAttr(resp.Data.Attributes, "values")
	headers, values = selectShiftSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := shiftSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildShiftSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderShiftSummaryTable(cmd, headers, values)
}

func parseDoShiftSummaryCreateOptions(cmd *cobra.Command) (doShiftSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	startOn, err := cmd.Flags().GetString("start-on")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	endOn, err := cmd.Flags().GetString("end-on")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doShiftSummaryCreateOptions{}, err
	}

	return doShiftSummaryCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		StartOn:     startOn,
		EndOn:       endOn,
		GroupByRaw:  groupByRaw,
		GroupBySet:  cmd.Flags().Changed("group-by"),
		SortRaw:     sortRaw,
		SortSet:     cmd.Flags().Changed("sort"),
		Limit:       limit,
		MetricsRaw:  metricsRaw,
		MetricsSet:  cmd.Flags().Changed("metrics"),
		Metrics:     metrics,
		AllMetrics:  allMetrics,
		FiltersJSON: filtersJSON,
		FilterPairs: filterPairs,
	}, nil
}

type shiftSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildShiftSummaryRows(headers []string, values [][]any) []map[string]any {
	if len(headers) == 0 || len(values) == 0 {
		return nil
	}

	rows := make([]map[string]any, 0, len(values))
	for _, row := range values {
		mapped := make(map[string]any, len(headers))
		for i, header := range headers {
			if i < len(row) {
				mapped[header] = row[i]
			} else {
				mapped[header] = nil
			}
		}
		rows = append(rows, mapped)
	}
	return rows
}

func renderShiftSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No shift summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatShiftSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parseShiftSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
	filters := map[string]any{}

	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON != "" {
		if err := json.Unmarshal([]byte(rawJSON), &filters); err != nil {
			return nil, fmt.Errorf("invalid --filters JSON: %w", err)
		}
	}

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --filter %q (expected key=value)", pair)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("invalid --filter %q (missing key)", pair)
		}
		filters[key] = value
	}

	return filters, nil
}

func resolveShiftSummaryGroupBy(opts doShiftSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"driver"}
}

func resolveShiftSummarySort(opts doShiftSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"shift_count:desc"}
}

func resolveShiftSummaryMetrics(opts doShiftSummaryCreateOptions) []string {
	if opts.AllMetrics {
		return nil
	}

	metrics := []string{}
	if opts.MetricsSet {
		metrics = append(metrics, splitCommaList(opts.MetricsRaw)...)
	}
	if len(opts.Metrics) > 0 {
		metrics = append(metrics, opts.Metrics...)
	}
	metrics = uniqueStrings(metrics)

	if opts.MetricsSet || len(opts.Metrics) > 0 {
		return metrics
	}

	return shiftSummaryDefaultMetrics
}

func selectShiftSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
	if len(headers) == 0 {
		return headers, values
	}

	headerIndex := make(map[string]int, len(headers))
	for i, header := range headers {
		headerIndex[header] = i
	}

	displayColumns := []string{}
	fullGroupByColumns := map[string]struct{}{}

	for _, group := range groupBy {
		if cols, ok := shiftSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := shiftSummaryGroupByDisplayColumns[group]; ok {
			displayColumns = append(displayColumns, cols...)
		} else if group != "" {
			displayColumns = append(displayColumns, group)
		}
	}

	seen := make(map[string]struct{}, len(headers))
	selectedHeaders := []string{}
	selectedIndexes := []int{}

	addHeader := func(header string) {
		if header == "" {
			return
		}
		if _, ok := seen[header]; ok {
			return
		}
		idx, ok := headerIndex[header]
		if !ok {
			return
		}
		seen[header] = struct{}{}
		selectedHeaders = append(selectedHeaders, header)
		selectedIndexes = append(selectedIndexes, idx)
	}

	for _, header := range displayColumns {
		addHeader(header)
	}

	if allMetrics {
		for _, header := range headers {
			if _, isGroupBy := fullGroupByColumns[header]; isGroupBy {
				continue
			}
			addHeader(header)
		}
	} else {
		metricSet := make(map[string]struct{}, len(metrics))
		for _, metric := range metrics {
			metricSet[metric] = struct{}{}
		}
		for _, header := range headers {
			if _, isGroupBy := fullGroupByColumns[header]; isGroupBy {
				continue
			}
			if _, ok := metricSet[header]; ok {
				addHeader(header)
			}
		}
	}

	if len(selectedHeaders) == 0 {
		return headers, values
	}

	filteredValues := make([][]any, 0, len(values))
	for _, row := range values {
		filteredRow := make([]any, len(selectedIndexes))
		for i, idx := range selectedIndexes {
			if idx < len(row) {
				filteredRow[i] = row[idx]
			}
		}
		filteredValues = append(filteredValues, filteredRow)
	}

	return selectedHeaders, filteredValues
}

func formatShiftSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatShiftSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatShiftSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatShiftSummaryNumber(header, typed)
	case float32:
		return formatShiftSummaryNumber(header, float64(typed))
	case int:
		return formatShiftSummaryNumber(header, float64(typed))
	case int64:
		return formatShiftSummaryNumber(header, float64(typed))
	case int32:
		return formatShiftSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatShiftSummaryString(header, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lowerHeader := strings.ToLower(header)
	switch {
	case strings.HasSuffix(lowerHeader, "_name"):
		return truncateString(value, 35)
	default:
		return value
	}
}

func formatShiftSummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_pct"):
		percent := value
		if value <= 1.5 {
			percent = value * 100
		}
		return fmt.Sprintf("%.1f%%", percent)
	case strings.HasSuffix(lowerHeader, "_count") || strings.HasSuffix(lowerHeader, "_sum") && strings.Contains(lowerHeader, "trip"):
		return strconv.FormatInt(int64(value+0.5), 10)
	case strings.Contains(lowerHeader, "revenue") || strings.Contains(lowerHeader, "cost") || strings.Contains(lowerHeader, "margin") && !strings.HasSuffix(lowerHeader, "_pct"):
		return fmt.Sprintf("$%.2f", value)
	case strings.Contains(lowerHeader, "tons"):
		return fmt.Sprintf("%.2f", value)
	case strings.Contains(lowerHeader, "hours"):
		return fmt.Sprintf("%.1f", value)
	case strings.Contains(lowerHeader, "miles"):
		return fmt.Sprintf("%.1f", value)
	case strings.Contains(lowerHeader, "minutes"):
		return fmt.Sprintf("%.1f", value)
	default:
		if value == float64(int64(value)) {
			return strconv.FormatInt(int64(value), 10)
		}
		return fmt.Sprintf("%.2f", value)
	}
}
