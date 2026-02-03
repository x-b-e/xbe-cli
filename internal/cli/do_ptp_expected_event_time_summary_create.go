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

var ptpExpectedEventTimeSummaryDefaultMetrics = []string{
	"snapshot_count",
	"error_minutes_avg",
	"error_minutes_p90",
}

var ptpExpectedEventTimeSummaryGroupByAllColumns = map[string][]string{
	"time_type":                      {"time_type"},
	"event_type":                     {"event_type_id", "event_type_name", "event_type_code"},
	"broker":                         {"broker_id", "broker_name"},
	"plan":                           {"project_transport_plan_id"},
	"plan_status":                    {"plan_status"},
	"project":                        {"project_id", "project_name"},
	"transport_order":                {"transport_order_id"},
	"transport_order_is_managed":     {"transport_order_is_managed"},
	"transport_order_project_office": {"transport_order_project_office_id", "transport_order_project_office_name"},
	"transport_order_project_category": {
		"transport_order_project_category_id",
		"transport_order_project_category_name",
	},
	"customer":                {"customer_id", "customer_name"},
	"location":                {"location_id", "location_name", "location_time_zone_id"},
	"driver":                  {"driver_ids", "driver_names"},
	"modeled_and_confident":   {"modeled_and_confident"},
	"actual_start_date_local": {"actual_start_date_local"},
	"actual_start_hour_local": {"actual_start_hour_local"},
	"actual_start_dow_local":  {"actual_start_dow_local"},
	"expected_created_date":   {"expected_created_date"},
	"lead_time_bin":           {"lead_time_bin"},
}

var ptpExpectedEventTimeSummaryGroupByDisplayColumns = map[string][]string{
	"time_type":                      {"time_type"},
	"event_type":                     {"event_type_name"},
	"broker":                         {"broker_name"},
	"plan":                           {"project_transport_plan_id"},
	"plan_status":                    {"plan_status"},
	"project":                        {"project_name"},
	"transport_order":                {"transport_order_id"},
	"transport_order_is_managed":     {"transport_order_is_managed"},
	"transport_order_project_office": {"transport_order_project_office_name"},
	"transport_order_project_category": {
		"transport_order_project_category_name",
	},
	"customer":                {"customer_name"},
	"location":                {"location_name"},
	"driver":                  {"driver_names"},
	"modeled_and_confident":   {"modeled_and_confident"},
	"actual_start_date_local": {"actual_start_date_local"},
	"actual_start_hour_local": {"actual_start_hour_local"},
	"actual_start_dow_local":  {"actual_start_dow_local"},
	"expected_created_date":   {"expected_created_date"},
	"lead_time_bin":           {"lead_time_bin"},
}

type doPTPExpectedEventTimeSummaryCreateOptions struct {
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

func newDoPTPExpectedEventTimeSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan expected event time summary",
		Long: `Create a project transport plan expected event time summary.

Aggregates expected vs actual event time snapshots for project transport plans.
Use this to analyze ETA accuracy, lead times, and modeled confidence for
transport events.

Group-by attributes:
  actual_start_date_local             Group by local actual start date
  actual_start_dow_local              Group by local day of week (0-6)
  actual_start_hour_local             Group by local hour (0-23)
  broker                              Group by broker
  customer                            Group by customer
  driver                              Group by driver
  event_type                          Group by event type
  expected_created_date               Group by expected time creation date
  lead_time_bin                       Group by lead time bin
  location                            Group by location
  modeled_and_confident               Group by modeled + confident flag
  plan                                Group by transport plan
  plan_status                         Group by plan status
  project                             Group by project
  time_type                           Group by time type (start_at/end_at)
  transport_order                     Group by transport order
  transport_order_is_managed          Group by managed status
  transport_order_project_category    Group by project category
  transport_order_project_office      Group by project office

Metrics:
  By default: snapshot_count, error_minutes_avg, error_minutes_p90

  Available metrics:
    snapshot_count                         Snapshot count
    actual_event_time_count                Actual event time count
    error_minutes_avg                      Average error (signed)
    error_minutes_median                   Median error (signed)
    error_minutes_p90                      90th percentile error
    error_minutes_p95                      95th percentile error
    error_minutes_min                      Minimum error
    error_minutes_max                      Maximum error
    error_minutes_stddev                   Error standard deviation
    absolute_error_minutes_avg             Average absolute error
    absolute_error_minutes_median          Median absolute error
    absolute_error_minutes_p90             90th percentile absolute error
    absolute_error_minutes_p95             95th percentile absolute error
    root_mean_squared_error_minutes        RMSE
    early_pct                              Early percentage
    late_pct                               Late percentage
    on_time_pct                            On-time percentage
    within_15_minutes_pct                  Within 15 minutes percentage
    within_30_minutes_pct                  Within 30 minutes percentage
    within_60_minutes_pct                  Within 60 minutes percentage
    within_120_minutes_pct                 Within 120 minutes percentage
    lead_time_minutes_avg                  Average lead time
    lead_time_minutes_median               Median lead time
    lead_time_minutes_p90                  90th percentile lead time
    lead_time_minutes_p95                  95th percentile lead time
    lead_time_minutes_min                  Minimum lead time
    lead_time_minutes_max                  Maximum lead time
    lead_time_minutes_stddev               Lead time standard deviation
    negative_lead_time_pct                 Negative lead time percentage
    expected_snapshots_per_lead_hour_avg   Snapshots per lead hour

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    actual_start_date_local             Actual start date (YYYY-MM-DD)
    actual_start_date_local_min         Minimum actual start date (YYYY-MM-DD)
    actual_start_date_local_max         Maximum actual start date (YYYY-MM-DD)
    actual_start_hour_local             Actual start hour (0-23)
    actual_start_dow_local              Actual start day of week (0-6)
    broker                              Broker ID
    customer                            Customer ID
    driver                              Driver ID
    event_type                          Event type ID
    expected_created_date               Expected created date (YYYY-MM-DD)
    expected_created_date_min           Minimum expected created date (YYYY-MM-DD)
    expected_created_date_max           Maximum expected created date (YYYY-MM-DD)
    lead_time_bin                       Lead time bin
    location                            Location ID
    modeled_and_confident               Modeled and confident (true/false)
    plan                                Transport plan ID
    plan_status                         Plan status
    project                             Project ID
    time_type                           Time type (start_at/end_at)
    transport_order                     Transport order ID
    transport_order_is_managed          Managed status (true/false)
    transport_order_project_category    Project category ID
    transport_order_project_office      Project office ID`,
		Example: `  # Summary grouped by event type
  xbe summarize ptp-expected-event-time-summary create --group-by event_type --filter broker=123 --filter expected_created_date_min=2025-01-01 --filter expected_created_date_max=2025-01-31

  # Summary by lead time bin
  xbe summarize ptp-expected-event-time-summary create --group-by lead_time_bin --filter broker=123

  # Summary by modeled confidence
  xbe summarize ptp-expected-event-time-summary create --group-by modeled_and_confident --filter broker=123

  # Summary with all metrics
  xbe summarize ptp-expected-event-time-summary create --group-by event_type --filter broker=123 --all-metrics

  # JSON output
  xbe summarize ptp-expected-event-time-summary create --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPTPExpectedEventTimeSummaryCreate,
	}
	initDoPTPExpectedEventTimeSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doPTPExpectedEventTimeSummaryCmd.AddCommand(newDoPTPExpectedEventTimeSummaryCreateCmd())
}

func initDoPTPExpectedEventTimeSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to event_type unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to snapshot_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPTPExpectedEventTimeSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPTPExpectedEventTimeSummaryCreateOptions(cmd)
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

	filters, err := parsePTPExpectedEventTimeSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolvePTPExpectedEventTimeSummaryGroupBy(opts)
	sort := resolvePTPExpectedEventTimeSummarySort(opts)
	metrics := resolvePTPExpectedEventTimeSummaryMetrics(opts)

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
			"type":       "project-transport-plan-expected-event-time-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-expected-event-time-summaries", jsonBody)
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
	headers, values = selectPTPExpectedEventTimeSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := ptpExpectedEventTimeSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildPTPExpectedEventTimeSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderPTPExpectedEventTimeSummaryTable(cmd, headers, values)
}

func parseDoPTPExpectedEventTimeSummaryCreateOptions(cmd *cobra.Command) (doPTPExpectedEventTimeSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doPTPExpectedEventTimeSummaryCreateOptions{}, err
	}

	return doPTPExpectedEventTimeSummaryCreateOptions{
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

type ptpExpectedEventTimeSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildPTPExpectedEventTimeSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderPTPExpectedEventTimeSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No PTP expected event time summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatPTPExpectedEventTimeSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parsePTPExpectedEventTimeSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolvePTPExpectedEventTimeSummaryGroupBy(opts doPTPExpectedEventTimeSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"event_type"}
}

func resolvePTPExpectedEventTimeSummarySort(opts doPTPExpectedEventTimeSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"snapshot_count:desc"}
}

func resolvePTPExpectedEventTimeSummaryMetrics(opts doPTPExpectedEventTimeSummaryCreateOptions) []string {
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

	return ptpExpectedEventTimeSummaryDefaultMetrics
}

func selectPTPExpectedEventTimeSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := ptpExpectedEventTimeSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := ptpExpectedEventTimeSummaryGroupByDisplayColumns[group]; ok {
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

func formatPTPExpectedEventTimeSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatPTPExpectedEventTimeSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatPTPExpectedEventTimeSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatPTPExpectedEventTimeSummaryNumber(header, typed)
	case float32:
		return formatPTPExpectedEventTimeSummaryNumber(header, float64(typed))
	case int:
		return formatPTPExpectedEventTimeSummaryNumber(header, float64(typed))
	case int64:
		return formatPTPExpectedEventTimeSummaryNumber(header, float64(typed))
	case int32:
		return formatPTPExpectedEventTimeSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatPTPExpectedEventTimeSummaryString(header, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lowerHeader := strings.ToLower(header)
	switch {
	case strings.HasSuffix(lowerHeader, "_name"), strings.HasSuffix(lowerHeader, "_names"):
		return truncateString(value, 35)
	default:
		return value
	}
}

func formatPTPExpectedEventTimeSummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_pct"):
		percent := value
		if value <= 1.5 {
			percent = value * 100
		}
		return fmt.Sprintf("%.1f%%", percent)
	case lowerHeader == "event_count" || strings.HasSuffix(lowerHeader, "_count"):
		return strconv.FormatInt(int64(value+0.5), 10)
	case strings.Contains(lowerHeader, "minutes"):
		return fmt.Sprintf("%.1f", value)
	default:
		if value == float64(int64(value)) {
			return strconv.FormatInt(int64(value), 10)
		}
		return fmt.Sprintf("%.2f", value)
	}
}
