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

var ptpEventTimeSummaryDefaultMetrics = []string{
	"event_count",
	"duration_minutes_avg",
	"duration_minutes_p90",
}

var ptpEventTimeSummaryGroupByAllColumns = map[string][]string{
	"event_kind":                     {"event_kind"},
	"event_type":                     {"event_type_id", "event_type_name"},
	"location":                       {"location_id", "location_name"},
	"project_transport_organization": {"project_transport_organization_id", "project_transport_organization_name"},
	"material_type":                  {"material_type_id", "material_type_name"},
	"driver":                         {"driver_id", "driver_name"},
	"project":                        {"project_id", "project_name"},
	"broker":                         {"broker_id", "broker_name"},
	"plan":                           {"plan_id"},
	"plan_status":                    {"plan_status"},
	"event_date_local":               {"event_date_local"},
	"event_hour_local":               {"event_hour_local"},
	"event_dow_local":                {"event_dow_local"},
	"is_app_owned":                   {"is_app_owned"},
	"is_manual_override":             {"is_manual_override"},
	"transport_order_is_managed":     {"transport_order_is_managed"},
}

var ptpEventTimeSummaryGroupByDisplayColumns = map[string][]string{
	"event_kind":                     {"event_kind"},
	"event_type":                     {"event_type_name"},
	"location":                       {"location_name"},
	"project_transport_organization": {"project_transport_organization_name"},
	"material_type":                  {"material_type_name"},
	"driver":                         {"driver_name"},
	"project":                        {"project_name"},
	"broker":                         {"broker_name"},
	"plan":                           {"plan_id"},
	"plan_status":                    {"plan_status"},
	"event_date_local":               {"event_date_local"},
	"event_hour_local":               {"event_hour_local"},
	"event_dow_local":                {"event_dow_local"},
	"is_app_owned":                   {"is_app_owned"},
	"is_manual_override":             {"is_manual_override"},
	"transport_order_is_managed":     {"transport_order_is_managed"},
}

type doPTPEventTimeSummaryCreateOptions struct {
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

func newDoPTPEventTimeSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan event time summary",
		Long: `Create a project transport plan event time summary.

Aggregates time spent on transport events (loading, unloading, wait times) for
project transport plans. Use this to analyze dwell times at pickup and delivery
locations, identify bottlenecks, and optimize site operations.

Group-by attributes:
  broker                              Group by broker
  driver                              Group by driver
  event_date_local                    Group by local event date
  event_dow_local                     Group by day of week (0-6)
  event_hour_local                    Group by hour (0-23)
  event_kind                          Group by event kind (actual/other)
  event_type                          Group by event type
  is_app_owned                        Group by app ownership
  is_manual_override                  Group by manual override status
  location                            Group by location
  material_type                       Group by material type
  plan                                Group by transport plan
  plan_status                         Group by plan status
  project                             Group by project
  project_transport_organization      Group by transport organization
  transport_order_is_managed          Group by managed status

Metrics:
  By default: event_count, duration_minutes_avg, duration_minutes_p90

  Available metrics:
    event_count                 Total events
    duration_minutes_sum        Total duration
    duration_minutes_avg        Average duration
    duration_minutes_min        Minimum duration
    duration_minutes_max        Maximum duration
    duration_minutes_stddev     Standard deviation
    duration_minutes_p50        50th percentile
    duration_minutes_p90        90th percentile
    duration_minutes_p95        95th percentile

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    broker                              Broker ID
    driver                              Driver ID
    event_date_local                    Specific date (YYYY-MM-DD)
    event_date_local_min                Minimum date (YYYY-MM-DD)
    event_date_local_max                Maximum date (YYYY-MM-DD)
    event_kind                          Event kind (actual/other)
    event_type                          Event type ID
    has_full_duration                   Has full duration (true/false, default true)
    is_app_owned                        App owned (true/false)
    is_manual_override                  Manual override (true/false)
    location                            Location ID
    material_type                       Material type ID
    plan                                Transport plan ID
    plan_status                         Plan status
    project                             Project ID
    project_transport_organization      Transport organization ID
    start_at_confidence_min             Min start confidence (0.0-1.0)
    start_at_confidence_max             Max start confidence (0.0-1.0)
    end_at_confidence_min               Min end confidence (0.0-1.0)
    end_at_confidence_max               Max end confidence (0.0-1.0)
    transport_order_is_managed          Managed status (true/false)`,
		Example: `  # Summary grouped by event type
  xbe summarize ptp-event-time-summary create --group-by event_type --filter broker=123 --filter event_date_local_min=2025-01-01 --filter event_date_local_max=2025-01-31

  # Summary by location
  xbe summarize ptp-event-time-summary create --group-by location --filter broker=123

  # Summary by hour of day
  xbe summarize ptp-event-time-summary create --group-by event_hour_local --filter broker=123 --filter event_date_local_min=2025-01-01 --filter event_date_local_max=2025-01-31

  # Summary with all metrics
  xbe summarize ptp-event-time-summary create --group-by event_type --filter broker=123 --all-metrics

  # JSON output
  xbe summarize ptp-event-time-summary create --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPTPEventTimeSummaryCreate,
	}
	initDoPTPEventTimeSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doPTPEventTimeSummaryCmd.AddCommand(newDoPTPEventTimeSummaryCreateCmd())
}

func initDoPTPEventTimeSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to event_type unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to event_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPTPEventTimeSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPTPEventTimeSummaryCreateOptions(cmd)
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

	filters, err := parsePTPEventTimeSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolvePTPEventTimeSummaryGroupBy(opts)
	sort := resolvePTPEventTimeSummarySort(opts)
	metrics := resolvePTPEventTimeSummaryMetrics(opts)

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
			"type":       "project-transport-plan-event-time-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-event-time-summaries", jsonBody)
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
	headers, values = selectPTPEventTimeSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := ptpEventTimeSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildPTPEventTimeSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderPTPEventTimeSummaryTable(cmd, headers, values)
}

func parseDoPTPEventTimeSummaryCreateOptions(cmd *cobra.Command) (doPTPEventTimeSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doPTPEventTimeSummaryCreateOptions{}, err
	}

	return doPTPEventTimeSummaryCreateOptions{
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

type ptpEventTimeSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildPTPEventTimeSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderPTPEventTimeSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No PTP event time summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatPTPEventTimeSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parsePTPEventTimeSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolvePTPEventTimeSummaryGroupBy(opts doPTPEventTimeSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"event_type"}
}

func resolvePTPEventTimeSummarySort(opts doPTPEventTimeSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"event_count:desc"}
}

func resolvePTPEventTimeSummaryMetrics(opts doPTPEventTimeSummaryCreateOptions) []string {
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

	return ptpEventTimeSummaryDefaultMetrics
}

func selectPTPEventTimeSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := ptpEventTimeSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := ptpEventTimeSummaryGroupByDisplayColumns[group]; ok {
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

func formatPTPEventTimeSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatPTPEventTimeSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatPTPEventTimeSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatPTPEventTimeSummaryNumber(header, typed)
	case float32:
		return formatPTPEventTimeSummaryNumber(header, float64(typed))
	case int:
		return formatPTPEventTimeSummaryNumber(header, float64(typed))
	case int64:
		return formatPTPEventTimeSummaryNumber(header, float64(typed))
	case int32:
		return formatPTPEventTimeSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatPTPEventTimeSummaryString(header, value string) string {
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

func formatPTPEventTimeSummaryNumber(header string, value float64) string {
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
