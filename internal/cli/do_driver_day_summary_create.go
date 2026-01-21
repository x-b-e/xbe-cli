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

var driverDaySummaryDefaultMetrics = []string{
	"driver_day_count",
	"duration_hours_sum",
	"shift_count_sum",
}

var driverDaySummaryGroupByAllColumns = map[string][]string{
	"broker":                 {"broker_id", "broker_name"},
	"trucker":                {"trucker_id", "trucker_name"},
	"driver":                 {"driver_id", "driver_name"},
	"is_driver_assigned":     {"is_driver_assigned"},
	"shift_count":            {"shift_count"},
	"is_managed":             {"is_managed"},
	"is_timecarded":          {"is_timecarded"},
	"year":                   {"year"},
	"month":                  {"month"},
	"week":                   {"week"},
	"date":                   {"date"},
	"time_card_cost_band_50": {"time_card_cost_band_50"},
	"trailer_classification": {"trailer_classification"},
}

var driverDaySummaryGroupByDisplayColumns = map[string][]string{
	"broker":                 {"broker_name"},
	"trucker":                {"trucker_name"},
	"driver":                 {"driver_name"},
	"is_driver_assigned":     {"is_driver_assigned"},
	"shift_count":            {"shift_count"},
	"is_managed":             {"is_managed"},
	"is_timecarded":          {"is_timecarded"},
	"year":                   {"year"},
	"month":                  {"month"},
	"week":                   {"week"},
	"date":                   {"date"},
	"time_card_cost_band_50": {"time_card_cost_band_50"},
	"trailer_classification": {"trailer_classification"},
}

type doDriverDaySummaryCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
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

func newDoDriverDaySummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver day summary",
		Long: `Create a driver day summary.

Aggregates daily driver activity including hours worked, time card costs, shift
counts, and miles driven. A "driver day" represents one driver's work on a
single calendar day. Use this to track daily driver utilization and costs.

Group-by attributes:
  broker                   Group by broker
  date                     Group by date
  driver                   Group by driver
  is_driver_assigned       Group by driver assignment status
  is_managed               Group by managed status
  is_timecarded            Group by time card status
  month                    Group by month
  shift_count              Group by shift count
  time_card_cost_band_50   Group by time card cost band
  trailer_classification   Group by trailer classification
  trucker                  Group by trucker
  week                     Group by week
  year                     Group by year

Metrics:
  By default: driver_day_count, duration_hours_sum, shift_count_sum

  Available metrics:
    driver_day_count          Count of driver days
    time_card_cost_sum        Sum of time card costs
    time_card_cost_mean       Mean time card cost
    time_card_cost_median     Median time card cost
    time_card_cost_std_dev    Standard deviation of time card cost
    duration_hours_sum        Sum of duration hours
    duration_hours_avg        Average duration hours
    duration_hours_median     Median duration hours
    shift_count_sum           Sum of shift counts
    shift_count_mean          Mean shift count
    trip_miles_calculated_sum Sum of calculated trip miles
    trip_miles_calculated_mean Mean calculated trip miles
    trip_miles_calculated_median Median calculated trip miles
    unloaded_driving_minutes_sum Sum of unloaded driving minutes
    unloaded_driving_minutes_mean Mean unloaded driving minutes
    unloaded_driving_minutes_median Median unloaded driving minutes

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    broker                   Broker ID
    trucker                  Trucker ID
    driver                   Driver ID
    is_driver_assigned       Driver assignment status (true/false)
    shift_count              Shift count
    is_managed               Managed status (true/false)
    is_timecarded            Time card status (true/false)
    year                     Year
    month                    Month (1-12)
    week                     Week (1-53)
    date                     Specific date (YYYY-MM-DD)
    start_on_min             Minimum start date (YYYY-MM-DD)
    start_on_max             Maximum start date (YYYY-MM-DD)
    trailer_classification   Trailer classification`,
		Example: `  # Summary grouped by driver
  xbe summarize driver-day-summary create --group-by driver --filter broker=123 --filter start_on_min=2025-01-01 --filter start_on_max=2025-01-31

  # Summary by trucker and date
  xbe summarize driver-day-summary create --group-by trucker,date --filter broker=123 --filter start_on_min=2025-01-01 --filter start_on_max=2025-01-31

  # Summary with all metrics
  xbe summarize driver-day-summary create --group-by driver --filter broker=123 --filter start_on_min=2025-01-01 --filter start_on_max=2025-01-31 --all-metrics

  # Total summary (no group-by)
  xbe summarize driver-day-summary create --group-by "" --filter broker=123 --filter start_on_min=2025-01-01 --filter start_on_max=2025-01-31

  # JSON output
  xbe summarize driver-day-summary create --filter broker=123 --filter start_on_min=2025-01-01 --filter start_on_max=2025-01-31 --json`,
		Args: cobra.NoArgs,
		RunE: runDoDriverDaySummaryCreate,
	}
	initDoDriverDaySummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverDaySummaryCmd.AddCommand(newDoDriverDaySummaryCreateCmd())
}

func initDoDriverDaySummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to driver unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to driver_day_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverDaySummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDriverDaySummaryCreateOptions(cmd)
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

	filters, err := parseDriverDaySummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolveDriverDaySummaryGroupBy(opts)
	sort := resolveDriverDaySummarySort(opts)
	metrics := resolveDriverDaySummaryMetrics(opts)

	attributes := map[string]any{
		"filters": filters,
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
			"type":       "driver-day-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/driver-day-summaries", jsonBody)
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
	headers, values = selectDriverDaySummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := driverDaySummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildDriverDaySummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderDriverDaySummaryTable(cmd, headers, values)
}

func parseDoDriverDaySummaryCreateOptions(cmd *cobra.Command) (doDriverDaySummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doDriverDaySummaryCreateOptions{}, err
	}

	return doDriverDaySummaryCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
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

type driverDaySummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildDriverDaySummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderDriverDaySummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver day summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatDriverDaySummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parseDriverDaySummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolveDriverDaySummaryGroupBy(opts doDriverDaySummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"driver"}
}

func resolveDriverDaySummarySort(opts doDriverDaySummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"driver_day_count:desc"}
}

func resolveDriverDaySummaryMetrics(opts doDriverDaySummaryCreateOptions) []string {
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

	return driverDaySummaryDefaultMetrics
}

func selectDriverDaySummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := driverDaySummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := driverDaySummaryGroupByDisplayColumns[group]; ok {
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

func formatDriverDaySummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatDriverDaySummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatDriverDaySummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatDriverDaySummaryNumber(header, typed)
	case float32:
		return formatDriverDaySummaryNumber(header, float64(typed))
	case int:
		return formatDriverDaySummaryNumber(header, float64(typed))
	case int64:
		return formatDriverDaySummaryNumber(header, float64(typed))
	case int32:
		return formatDriverDaySummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatDriverDaySummaryString(header, value string) string {
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

func formatDriverDaySummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_count"):
		return strconv.FormatInt(int64(value+0.5), 10)
	case strings.Contains(lowerHeader, "cost"):
		return fmt.Sprintf("$%.2f", value)
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
