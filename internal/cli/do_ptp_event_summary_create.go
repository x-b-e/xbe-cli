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

var ptpEventSummaryDefaultMetrics = []string{
	"count",
	"location_assignment_prediction_correct_pct",
}

var ptpEventSummaryGroupByAllColumns = map[string][]string{
	"created_date":                      {"created_date"},
	"event_type":                        {"event_type_id", "event_type_name"},
	"broker":                            {"broker_id", "broker_name"},
	"transport_order_project_office":    {"transport_order_project_office_id", "transport_order_project_office_name"},
	"transport_order_project_category":  {"transport_order_project_category_id", "transport_order_project_category_name"},
	"project_transport_plan_created_by": {"project_transport_plan_created_by_id", "project_transport_plan_created_by_name"},
	"transport_order_is_managed":        {"transport_order_is_managed"},
	"has_transport_order_stop_role":     {"has_transport_order_stop_role"},
	"has_location_prediction_position":  {"has_location_prediction_position"},
}

var ptpEventSummaryGroupByDisplayColumns = map[string][]string{
	"created_date":                      {"created_date"},
	"event_type":                        {"event_type_name"},
	"broker":                            {"broker_name"},
	"transport_order_project_office":    {"transport_order_project_office_name"},
	"transport_order_project_category":  {"transport_order_project_category_name"},
	"project_transport_plan_created_by": {"project_transport_plan_created_by_name"},
	"transport_order_is_managed":        {"transport_order_is_managed"},
	"has_transport_order_stop_role":     {"has_transport_order_stop_role"},
	"has_location_prediction_position":  {"has_location_prediction_position"},
}

type doPTPEventSummaryCreateOptions struct {
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

func newDoPTPEventSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan event summary",
		Long: `Create a project transport plan event summary.

Aggregates transport events (pickups, deliveries, arrivals, departures) for
project transport plans. Use this to analyze event patterns by type and measure
the accuracy of location prediction algorithms.

Group-by attributes:
  broker                              Group by broker
  created_date                        Group by creation date
  event_type                          Group by event type
  has_location_prediction_position    Group by location prediction availability
  has_transport_order_stop_role       Group by stop role availability
  project_transport_plan_created_by   Group by creator
  transport_order_is_managed          Group by managed status
  transport_order_project_category    Group by project category
  transport_order_project_office      Group by project office

Metrics:
  By default: count, location_assignment_prediction_correct_pct

  Available metrics:
    count                                        Event count
    location_assignment_prediction_missing_pct   Prediction missing percentage
    location_assignment_prediction_correct_pct   Prediction correct percentage
    location_assignment_prediction_incorrect_pct Prediction incorrect percentage
    location_assignment_prediction_top_2_pct     Top 2 percentage
    location_assignment_prediction_top_3_pct     Top 3 percentage
    location_assignment_prediction_top_5_pct     Top 5 percentage
    location_assignment_prediction_top_10_pct    Top 10 percentage

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    broker                              Broker ID
    created_date                        Specific date (YYYY-MM-DD)
    created_date_min                    Minimum date (YYYY-MM-DD)
    created_date_max                    Maximum date (YYYY-MM-DD)
    event_type                          Event type ID
    has_transport_order_stop_role       Has stop role (true/false)
    has_location_prediction_position    Has location prediction (true/false)`,
		Example: `  # Summary grouped by event type
  xbe summarize ptp-event-summary create --group-by event_type --filter broker=123 --filter created_date_min=2025-01-01 --filter created_date_max=2025-01-31

  # Summary by broker
  xbe summarize ptp-event-summary create --group-by broker --filter broker=123

  # Summary with all metrics
  xbe summarize ptp-event-summary create --group-by event_type --filter broker=123 --all-metrics

  # JSON output
  xbe summarize ptp-event-summary create --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPTPEventSummaryCreate,
	}
	initDoPTPEventSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doPTPEventSummaryCmd.AddCommand(newDoPTPEventSummaryCreateCmd())
}

func initDoPTPEventSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to event_type unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPTPEventSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPTPEventSummaryCreateOptions(cmd)
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

	filters, err := parsePTPEventSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolvePTPEventSummaryGroupBy(opts)
	sort := resolvePTPEventSummarySort(opts)
	metrics := resolvePTPEventSummaryMetrics(opts)

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
			"type":       "project-transport-plan-event-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-event-summaries", jsonBody)
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
	headers, values = selectPTPEventSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := ptpEventSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildPTPEventSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderPTPEventSummaryTable(cmd, headers, values)
}

func parseDoPTPEventSummaryCreateOptions(cmd *cobra.Command) (doPTPEventSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doPTPEventSummaryCreateOptions{}, err
	}

	return doPTPEventSummaryCreateOptions{
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

type ptpEventSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildPTPEventSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderPTPEventSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No PTP event summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatPTPEventSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parsePTPEventSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolvePTPEventSummaryGroupBy(opts doPTPEventSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"event_type"}
}

func resolvePTPEventSummarySort(opts doPTPEventSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"count:desc"}
}

func resolvePTPEventSummaryMetrics(opts doPTPEventSummaryCreateOptions) []string {
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

	return ptpEventSummaryDefaultMetrics
}

func selectPTPEventSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := ptpEventSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := ptpEventSummaryGroupByDisplayColumns[group]; ok {
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

func formatPTPEventSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatPTPEventSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatPTPEventSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatPTPEventSummaryNumber(header, typed)
	case float32:
		return formatPTPEventSummaryNumber(header, float64(typed))
	case int:
		return formatPTPEventSummaryNumber(header, float64(typed))
	case int64:
		return formatPTPEventSummaryNumber(header, float64(typed))
	case int32:
		return formatPTPEventSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatPTPEventSummaryString(header, value string) string {
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

func formatPTPEventSummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_pct"):
		percent := value
		if value <= 1.5 {
			percent = value * 100
		}
		return fmt.Sprintf("%.1f%%", percent)
	case lowerHeader == "count" || strings.HasSuffix(lowerHeader, "_count"):
		return strconv.FormatInt(int64(value+0.5), 10)
	default:
		if value == float64(int64(value)) {
			return strconv.FormatInt(int64(value), 10)
		}
		return fmt.Sprintf("%.2f", value)
	}
}
