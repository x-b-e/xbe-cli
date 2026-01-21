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

var deviceLocationEventSummaryDefaultMetrics = []string{
	"event_count",
}

var deviceLocationEventSummaryGroupByAllColumns = map[string][]string{
	"device":              {"device_id"},
	"device_identifier":   {"device_identifier"},
	"user":                {"user_id", "user_name"},
	"event_on":            {"event_on"},
	"device_name":         {"device_name"},
	"device_model":        {"device_model"},
	"device_version":      {"device_version"},
	"device_platform":     {"device_platform"},
	"device_manufacturer": {"device_manufacturer"},
	"native_app_version":  {"native_app_version"},
	"native_ota_version":  {"native_ota_version"},
}

var deviceLocationEventSummaryGroupByDisplayColumns = map[string][]string{
	"device":              {"device_id"},
	"device_identifier":   {"device_identifier"},
	"user":                {"user_name"},
	"event_on":            {"event_on"},
	"device_name":         {"device_name"},
	"device_model":        {"device_model"},
	"device_version":      {"device_version"},
	"device_platform":     {"device_platform"},
	"device_manufacturer": {"device_manufacturer"},
	"native_app_version":  {"native_app_version"},
	"native_ota_version":  {"native_ota_version"},
}

type doDeviceLocationEventSummaryCreateOptions struct {
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

func newDoDeviceLocationEventSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a device location event summary",
		Long: `Create a device location event summary.

Aggregates GPS location events from mobile devices and tablets. Use this to
analyze device activity, track app versions in the field, and monitor GPS
reporting frequency by device type or manufacturer.

Group-by attributes:
  device                   Group by device ID
  device_identifier        Group by device identifier
  device_manufacturer      Group by device manufacturer
  device_model             Group by device model
  device_name              Group by device name
  device_platform          Group by device platform
  device_version           Group by device version
  event_on                 Group by event date
  native_app_version       Group by native app version
  native_ota_version       Group by native OTA version
  user                     Group by user

Metrics:
  By default: event_count

  Available metrics:
    event_count              Count of location events

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    device                   Device ID
    device_identifier        Device identifier
    user                     User ID
    event_on                 Specific event date (YYYY-MM-DD)
    event_on_min             Minimum event date (YYYY-MM-DD)
    event_on_max             Maximum event date (YYYY-MM-DD)
    device_name              Device name
    device_model             Device model
    device_version           Device version
    device_platform          Device platform
    device_manufacturer      Device manufacturer
    native_app_version       Native app version
    native_ota_version       Native OTA version`,
		Example: `  # Summary grouped by device
  xbe summarize device-location-event-summary create --group-by device --filter event_on_min=2025-01-01 --filter event_on_max=2025-01-31

  # Summary by user
  xbe summarize device-location-event-summary create --group-by user --filter event_on_min=2025-01-01 --filter event_on_max=2025-01-31

  # Summary by device platform
  xbe summarize device-location-event-summary create --group-by device_platform --filter event_on_min=2025-01-01 --filter event_on_max=2025-01-31

  # Summary by event date
  xbe summarize device-location-event-summary create --group-by event_on --filter user=123 --filter event_on_min=2025-01-01 --filter event_on_max=2025-01-31

  # JSON output
  xbe summarize device-location-event-summary create --filter event_on_min=2025-01-01 --filter event_on_max=2025-01-31 --json`,
		Args: cobra.NoArgs,
		RunE: runDoDeviceLocationEventSummaryCreate,
	}
	initDoDeviceLocationEventSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doDeviceLocationEventSummaryCmd.AddCommand(newDoDeviceLocationEventSummaryCreateCmd())
}

func initDoDeviceLocationEventSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to device unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to event_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"user\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDeviceLocationEventSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDeviceLocationEventSummaryCreateOptions(cmd)
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

	filters, err := parseDeviceLocationEventSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolveDeviceLocationEventSummaryGroupBy(opts)
	sort := resolveDeviceLocationEventSummarySort(opts)
	metrics := resolveDeviceLocationEventSummaryMetrics(opts)

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
			"type":       "device-location-event-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/device-location-event-summaries", jsonBody)
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
	headers, values = selectDeviceLocationEventSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := deviceLocationEventSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildDeviceLocationEventSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderDeviceLocationEventSummaryTable(cmd, headers, values)
}

func parseDoDeviceLocationEventSummaryCreateOptions(cmd *cobra.Command) (doDeviceLocationEventSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doDeviceLocationEventSummaryCreateOptions{}, err
	}

	return doDeviceLocationEventSummaryCreateOptions{
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

type deviceLocationEventSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildDeviceLocationEventSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderDeviceLocationEventSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No device location event summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatDeviceLocationEventSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parseDeviceLocationEventSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolveDeviceLocationEventSummaryGroupBy(opts doDeviceLocationEventSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"device"}
}

func resolveDeviceLocationEventSummarySort(opts doDeviceLocationEventSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"event_count:desc"}
}

func resolveDeviceLocationEventSummaryMetrics(opts doDeviceLocationEventSummaryCreateOptions) []string {
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

	return deviceLocationEventSummaryDefaultMetrics
}

func selectDeviceLocationEventSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := deviceLocationEventSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := deviceLocationEventSummaryGroupByDisplayColumns[group]; ok {
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

func formatDeviceLocationEventSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatDeviceLocationEventSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatDeviceLocationEventSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatDeviceLocationEventSummaryNumber(header, typed)
	case float32:
		return formatDeviceLocationEventSummaryNumber(header, float64(typed))
	case int:
		return formatDeviceLocationEventSummaryNumber(header, float64(typed))
	case int64:
		return formatDeviceLocationEventSummaryNumber(header, float64(typed))
	case int32:
		return formatDeviceLocationEventSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatDeviceLocationEventSummaryString(header, value string) string {
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

func formatDeviceLocationEventSummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_count"):
		return strconv.FormatInt(int64(value+0.5), 10)
	default:
		if value == float64(int64(value)) {
			return strconv.FormatInt(int64(value), 10)
		}
		return fmt.Sprintf("%.2f", value)
	}
}
