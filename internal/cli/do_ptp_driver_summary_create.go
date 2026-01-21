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

var ptpDriverSummaryDefaultMetrics = []string{
	"count",
	"confirmation_pct",
	"driver_assignment_prediction_correct_pct",
}

var ptpDriverSummaryGroupByAllColumns = map[string][]string{
	"ordered_date":                   {"ordered_date"},
	"pickup_date":                    {"pickup_date"},
	"customer":                       {"customer_id", "customer_name"},
	"broker":                         {"broker_id", "broker_name"},
	"project_division":               {"project_division_id", "project_division_name"},
	"project_office":                 {"project_office_id", "project_office_name"},
	"project_category":               {"project_category_id", "project_category_name"},
	"driver":                         {"driver_id", "driver_name"},
	"transport_order_status":         {"transport_order_status"},
	"is_managed":                     {"is_managed"},
	"was_unmanaged":                  {"was_unmanaged"},
	"has_driver_prediction_position": {"has_driver_prediction_position"},
}

var ptpDriverSummaryGroupByDisplayColumns = map[string][]string{
	"ordered_date":                   {"ordered_date"},
	"pickup_date":                    {"pickup_date"},
	"customer":                       {"customer_name"},
	"broker":                         {"broker_name"},
	"project_division":               {"project_division_name"},
	"project_office":                 {"project_office_name"},
	"project_category":               {"project_category_name"},
	"driver":                         {"driver_name"},
	"transport_order_status":         {"transport_order_status"},
	"is_managed":                     {"is_managed"},
	"was_unmanaged":                  {"was_unmanaged"},
	"has_driver_prediction_position": {"has_driver_prediction_position"},
}

type doPTPDriverSummaryCreateOptions struct {
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

func newDoPTPDriverSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan driver summary",
		Long: `Create a project transport plan driver summary.

Aggregates driver assignment data for project transport plans. Use this to analyze
how accurately drivers are being assigned to loads, track confirmation rates, and
measure the success of driver prediction algorithms.

Group-by attributes:
  broker                          Group by broker
  customer                        Group by customer
  driver                          Group by driver
  has_driver_prediction_position  Group by prediction position availability
  is_managed                      Group by managed status
  ordered_date                    Group by ordered date
  pickup_date                     Group by pickup date
  project_category                Group by project category
  project_division                Group by project division
  project_office                  Group by project office
  transport_order_status          Group by transport order status
  was_unmanaged                   Group by was unmanaged status

Metrics:
  By default: count, confirmation_pct, driver_assignment_prediction_correct_pct

  Available metrics:
    count                                        Driver count
    confirmation_count                           Confirmation count
    confirmation_pct                             Confirmation percentage
    driver_assignment_prediction_present_pct     Prediction present percentage
    driver_assignment_prediction_missing_pct     Prediction missing percentage
    driver_assignment_prediction_not_in_candidates_pct  Not in candidates percentage
    driver_assignment_prediction_in_candidates_pct      In candidates percentage
    driver_assignment_prediction_correct_pct     Correct percentage
    driver_assignment_prediction_incorrect_pct   Incorrect percentage
    driver_assignment_prediction_top_2_pct       Top 2 percentage
    driver_assignment_prediction_top_3_pct       Top 3 percentage
    driver_assignment_prediction_top_5_pct       Top 5 percentage
    driver_assignment_prediction_top_10_pct      Top 10 percentage
    transport_order_count                        Transport order count
    ordered_minutes_sum/avg                      Ordered minutes
    ordered_miles_sum/avg                        Ordered miles
    routed_minutes_sum/avg                       Routed minutes
    routed_miles_sum/avg                         Routed miles
    deviated_minutes_sum/avg                     Deviated minutes
    deviated_miles_sum/avg                       Deviated miles
    routed_minutes_over_ordered_pct              Routed vs ordered minutes percentage
    routed_miles_over_ordered_pct                Routed vs ordered miles percentage
    deviated_minutes_over_routed_pct             Deviated vs routed minutes percentage
    deviated_miles_over_routed_pct               Deviated vs routed miles percentage

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    broker, customer, driver, is_managed, was_unmanaged,
    has_driver_prediction_position, ordered_date, ordered_date_min,
    ordered_date_max, pickup_date, pickup_date_min, pickup_date_max,
    project_division, project_office, project_category,
    transport_order_status, transport_order`,
		Example: `  # Summary grouped by driver
  xbe summarize ptp-driver-summary create --group-by driver --filter broker=123 --filter ordered_date_min=2025-01-01 --filter ordered_date_max=2025-01-31

  # Summary by customer
  xbe summarize ptp-driver-summary create --group-by customer --filter broker=123

  # Summary with all metrics
  xbe summarize ptp-driver-summary create --group-by driver --filter broker=123 --all-metrics

  # Total summary (no group-by)
  xbe summarize ptp-driver-summary create --group-by "" --filter broker=123 --filter ordered_date_min=2025-01-01 --filter ordered_date_max=2025-01-31

  # JSON output
  xbe summarize ptp-driver-summary create --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoPTPDriverSummaryCreate,
	}
	initDoPTPDriverSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doPTPDriverSummaryCmd.AddCommand(newDoPTPDriverSummaryCreateCmd())
}

func initDoPTPDriverSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to driver unless set.")
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

func runDoPTPDriverSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPTPDriverSummaryCreateOptions(cmd)
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

	filters, err := parsePTPDriverSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolvePTPDriverSummaryGroupBy(opts)
	sort := resolvePTPDriverSummarySort(opts)
	metrics := resolvePTPDriverSummaryMetrics(opts)

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
			"type":       "project-transport-plan-driver-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-driver-summaries", jsonBody)
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
	headers, values = selectPTPDriverSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := ptpDriverSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildPTPDriverSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderPTPDriverSummaryTable(cmd, headers, values)
}

func parseDoPTPDriverSummaryCreateOptions(cmd *cobra.Command) (doPTPDriverSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doPTPDriverSummaryCreateOptions{}, err
	}

	return doPTPDriverSummaryCreateOptions{
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

type ptpDriverSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildPTPDriverSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderPTPDriverSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No PTP driver summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatPTPDriverSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parsePTPDriverSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolvePTPDriverSummaryGroupBy(opts doPTPDriverSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"driver"}
}

func resolvePTPDriverSummarySort(opts doPTPDriverSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"count:desc"}
}

func resolvePTPDriverSummaryMetrics(opts doPTPDriverSummaryCreateOptions) []string {
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

	return ptpDriverSummaryDefaultMetrics
}

func selectPTPDriverSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := ptpDriverSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := ptpDriverSummaryGroupByDisplayColumns[group]; ok {
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

func formatPTPDriverSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatPTPDriverSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatPTPDriverSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatPTPDriverSummaryNumber(header, typed)
	case float32:
		return formatPTPDriverSummaryNumber(header, float64(typed))
	case int:
		return formatPTPDriverSummaryNumber(header, float64(typed))
	case int64:
		return formatPTPDriverSummaryNumber(header, float64(typed))
	case int32:
		return formatPTPDriverSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatPTPDriverSummaryString(header, value string) string {
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

func formatPTPDriverSummaryNumber(header string, value float64) string {
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
