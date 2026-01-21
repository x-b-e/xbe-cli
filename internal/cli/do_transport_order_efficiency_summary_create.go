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

var transportOrderEfficiencySummaryDefaultMetrics = []string{
	"transport_order_count",
	"ordered_miles_sum",
	"routed_miles_sum",
	"deviated_miles_sum",
}

var transportOrderEfficiencySummaryGroupByAllColumns = map[string][]string{
	"ordered_date":           {"ordered_date"},
	"pickup_date":            {"pickup_date"},
	"customer":               {"customer_id", "customer_name"},
	"broker":                 {"broker_id", "broker_name"},
	"project_division":       {"project_division_id", "project_division_name"},
	"project_office":         {"project_office_id", "project_office_name"},
	"project_category":       {"project_category_id", "project_category_name"},
	"driver":                 {"driver_id", "driver_name"},
	"is_managed":             {"is_managed"},
	"was_unmanaged":          {"was_unmanaged"},
	"transport_order_status": {"transport_order_status"},
	"transport_order":        {"transport_order_id"},
}

var transportOrderEfficiencySummaryGroupByDisplayColumns = map[string][]string{
	"ordered_date":           {"ordered_date"},
	"pickup_date":            {"pickup_date"},
	"customer":               {"customer_name"},
	"broker":                 {"broker_name"},
	"project_division":       {"project_division_name"},
	"project_office":         {"project_office_name"},
	"project_category":       {"project_category_name"},
	"driver":                 {"driver_name"},
	"is_managed":             {"is_managed"},
	"was_unmanaged":          {"was_unmanaged"},
	"transport_order_status": {"transport_order_status"},
	"transport_order":        {"transport_order_id"},
}

type doTransportOrderEfficiencySummaryCreateOptions struct {
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

func newDoTransportOrderEfficiencySummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a transport order efficiency summary",
		Long: `Create a transport order efficiency summary.

Aggregates routing efficiency metrics for transport orders comparing ordered
(planned) miles vs actual routed miles and deviations. Use this to identify
routing inefficiencies, measure driver compliance with planned routes, and
optimize dispatch decisions.

Group-by attributes:
  broker                    Group by broker
  customer                  Group by customer
  driver                    Group by driver
  is_managed                Group by managed status
  ordered_date              Group by ordered date
  pickup_date               Group by pickup date
  project_category          Group by project category
  project_division          Group by project division
  project_office            Group by project office
  transport_order           Group by transport order
  transport_order_status    Group by transport order status
  was_unmanaged             Group by was unmanaged status

Metrics:
  By default: transport_order_count, ordered_miles_sum, routed_miles_sum, deviated_miles_sum

  Available metrics:
    transport_order_count               Count of transport orders
    ordered_minutes_sum                 Sum of ordered minutes
    ordered_minutes_avg                 Average ordered minutes
    ordered_miles_sum                   Sum of ordered miles
    ordered_miles_avg                   Average ordered miles
    routed_minutes_sum                  Sum of routed minutes
    routed_minutes_avg                  Average routed minutes
    routed_miles_sum                    Sum of routed miles
    routed_miles_avg                    Average routed miles
    deviated_minutes_sum                Sum of deviated minutes
    deviated_minutes_avg                Average deviated minutes
    deviated_miles_sum                  Sum of deviated miles
    deviated_miles_avg                  Average deviated miles
    routed_minutes_over_ordered_pct     Routed vs ordered minutes percentage
    routed_miles_over_ordered_pct       Routed vs ordered miles percentage
    deviated_minutes_over_routed_pct    Deviated vs routed minutes percentage
    deviated_miles_over_routed_pct      Deviated vs routed miles percentage
    routed_minutes_minus_ordered        Routed minus ordered minutes
    routed_miles_minus_ordered          Routed minus ordered miles
    deviated_minutes_minus_routed       Deviated minus routed minutes
    deviated_miles_minus_routed         Deviated minus routed miles
    project_transport_plan_driver_count PTP driver count
    project_transport_plan_driver_confirmation_count PTP driver confirmation count

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    broker                   Broker ID
    customer                 Customer ID
    driver                   Driver ID
    is_managed               Managed status (true/false)
    was_unmanaged            Was unmanaged status (true/false)
    ordered_date             Specific ordered date (YYYY-MM-DD)
    ordered_date_min         Minimum ordered date (YYYY-MM-DD)
    ordered_date_max         Maximum ordered date (YYYY-MM-DD)
    pickup_date              Specific pickup date (YYYY-MM-DD)
    pickup_date_min          Minimum pickup date (YYYY-MM-DD)
    pickup_date_max          Maximum pickup date (YYYY-MM-DD)
    project_division         Project division ID
    project_office           Project office ID
    project_category         Project category ID
    transport_order          Transport order ID
    transport_order_status   Transport order status`,
		Example: `  # Summary grouped by customer
  xbe summarize transport-order-efficiency-summary create --group-by customer --filter broker=123 --filter ordered_date_min=2025-01-01 --filter ordered_date_max=2025-01-31

  # Summary by driver
  xbe summarize transport-order-efficiency-summary create --group-by driver --filter broker=123 --filter pickup_date_min=2025-01-01 --filter pickup_date_max=2025-01-31

  # Summary with all metrics
  xbe summarize transport-order-efficiency-summary create --group-by customer --filter broker=123 --all-metrics

  # Total summary (no group-by)
  xbe summarize transport-order-efficiency-summary create --group-by "" --filter broker=123 --filter ordered_date_min=2025-01-01 --filter ordered_date_max=2025-01-31

  # JSON output
  xbe summarize transport-order-efficiency-summary create --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTransportOrderEfficiencySummaryCreate,
	}
	initDoTransportOrderEfficiencySummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doTransportOrderEfficiencySummaryCmd.AddCommand(newDoTransportOrderEfficiencySummaryCreateCmd())
}

func initDoTransportOrderEfficiencySummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to customer unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to transport_order_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTransportOrderEfficiencySummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTransportOrderEfficiencySummaryCreateOptions(cmd)
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

	filters, err := parseTransportOrderEfficiencySummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolveTransportOrderEfficiencySummaryGroupBy(opts)
	sort := resolveTransportOrderEfficiencySummarySort(opts)
	metrics := resolveTransportOrderEfficiencySummaryMetrics(opts)

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
			"type":       "transport-order-efficiency-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/transport-order-efficiency-summaries", jsonBody)
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
	headers, values = selectTransportOrderEfficiencySummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := transportOrderEfficiencySummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildTransportOrderEfficiencySummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderTransportOrderEfficiencySummaryTable(cmd, headers, values)
}

func parseDoTransportOrderEfficiencySummaryCreateOptions(cmd *cobra.Command) (doTransportOrderEfficiencySummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doTransportOrderEfficiencySummaryCreateOptions{}, err
	}

	return doTransportOrderEfficiencySummaryCreateOptions{
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

type transportOrderEfficiencySummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildTransportOrderEfficiencySummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderTransportOrderEfficiencySummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No transport order efficiency summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatTransportOrderEfficiencySummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parseTransportOrderEfficiencySummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolveTransportOrderEfficiencySummaryGroupBy(opts doTransportOrderEfficiencySummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"customer"}
}

func resolveTransportOrderEfficiencySummarySort(opts doTransportOrderEfficiencySummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"transport_order_count:desc"}
}

func resolveTransportOrderEfficiencySummaryMetrics(opts doTransportOrderEfficiencySummaryCreateOptions) []string {
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

	return transportOrderEfficiencySummaryDefaultMetrics
}

func selectTransportOrderEfficiencySummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := transportOrderEfficiencySummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := transportOrderEfficiencySummaryGroupByDisplayColumns[group]; ok {
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

func formatTransportOrderEfficiencySummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatTransportOrderEfficiencySummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatTransportOrderEfficiencySummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatTransportOrderEfficiencySummaryNumber(header, typed)
	case float32:
		return formatTransportOrderEfficiencySummaryNumber(header, float64(typed))
	case int:
		return formatTransportOrderEfficiencySummaryNumber(header, float64(typed))
	case int64:
		return formatTransportOrderEfficiencySummaryNumber(header, float64(typed))
	case int32:
		return formatTransportOrderEfficiencySummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatTransportOrderEfficiencySummaryString(header, value string) string {
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

func formatTransportOrderEfficiencySummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_pct"):
		percent := value
		if value <= 1.5 {
			percent = value * 100
		}
		return fmt.Sprintf("%.1f%%", percent)
	case strings.HasSuffix(lowerHeader, "_count"):
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
