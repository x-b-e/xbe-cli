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

var materialSiteReadingSummaryDefaultMetrics = []string{
	"value_avg",
}

var materialSiteReadingSummaryGroupByAllColumns = map[string][]string{
	"minute":                              {"minute", "duration_seconds"},
	"five_minute":                         {"five_minute", "duration_seconds"},
	"ten_minute":                          {"ten_minute", "duration_seconds"},
	"quarter_hour":                        {"quarter_hour", "duration_seconds"},
	"half_hour":                           {"half_hour", "duration_seconds"},
	"hour":                                {"hour", "duration_seconds"},
	"material_site":                       {"material_site_id", "material_site_name"},
	"material_site_measure":               {"material_site_measure_id", "material_site_measure_name"},
	"material_site_reading_material_type": {"material_site_reading_material_type_id", "material_site_reading_material_type_external_id"},
	"material_site_reading_material_type_presence": {"material_site_reading_material_type_presence"},
}

var materialSiteReadingSummaryGroupByDisplayColumns = map[string][]string{
	"minute":                              {"minute"},
	"five_minute":                         {"five_minute"},
	"ten_minute":                          {"ten_minute"},
	"quarter_hour":                        {"quarter_hour"},
	"half_hour":                           {"half_hour"},
	"hour":                                {"hour"},
	"material_site":                       {"material_site_name"},
	"material_site_measure":               {"material_site_measure_name"},
	"material_site_reading_material_type": {"material_site_reading_material_type_external_id", "material_site_reading_material_type_id"},
	"material_site_reading_material_type_presence": {"material_site_reading_material_type_presence"},
}

type doMaterialSiteReadingSummaryCreateOptions struct {
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

func newDoMaterialSiteReadingSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material site reading summary",
		Long: `Create a material site reading summary.

This command aggregates material site readings into time buckets for a specific
site and measure. You must provide material_site, material_site_measure,
reading_at_min, and reading_at_max filters. The time window must be positive and
no longer than 1 day.

Group-by attributes:
  minute                                     Group by minute
  five_minute                                Group by 5-minute buckets
  ten_minute                                 Group by 10-minute buckets
  quarter_hour                               Group by 15-minute buckets
  half_hour                                  Group by 30-minute buckets
  hour                                       Group by hour
  material_site                              Group by material site (material_site_id, material_site_name)
  material_site_measure                      Group by material site measure (material_site_measure_id, material_site_measure_name)
  material_site_reading_material_type        Group by reading material type (material_site_reading_material_type_id, material_site_reading_material_type_external_id)
  material_site_reading_material_type_presence  Group by whether a reading material type is present

Metrics:
  By default: value_avg

  Available metrics:
    value_avg                  Average reading value

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Required filters:
    material_site                              Material site ID
    material_site_measure                      Material site measure ID
    reading_at_min                             Minimum reading timestamp (ISO 8601)
    reading_at_max                             Maximum reading timestamp (ISO 8601)

  Optional filters:
    material_site_reading_material_type        Reading material type ID
    material_site_reading_material_type_presence  Require material type presence (true/false)
    minute                                     Filter by minute bucket
    five_minute                                Filter by 5-minute bucket
    ten_minute                                 Filter by 10-minute bucket
    quarter_hour                               Filter by 15-minute bucket
    half_hour                                  Filter by 30-minute bucket
    hour                                       Filter by hour bucket`,
		Example: `  # Minute-level summary for a material site measure
  xbe summarize material-site-reading-summary create --group-by minute \
    --filter material_site=123 --filter material_site_measure=456 \
    --filter reading_at_min=2025-01-01T00:00:00Z --filter reading_at_max=2025-01-01T00:30:00Z

  # Hourly summary for a 12-hour window
  xbe summarize material-site-reading-summary create --group-by hour \
    --filter material_site=123 --filter material_site_measure=456 \
    --filter reading_at_min=2025-01-01T00:00:00Z --filter reading_at_max=2025-01-01T12:00:00Z

  # Filter to readings with a material type
  xbe summarize material-site-reading-summary create --group-by minute \
    --filter material_site=123 --filter material_site_measure=456 \
    --filter reading_at_min=2025-01-01T00:00:00Z --filter reading_at_max=2025-01-01T00:30:00Z \
    --filter material_site_reading_material_type_presence=true

  # JSON output
  xbe summarize material-site-reading-summary create --filter material_site=123 \
    --filter material_site_measure=456 --filter reading_at_min=2025-01-01T00:00:00Z \
    --filter reading_at_max=2025-01-01T00:30:00Z --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialSiteReadingSummaryCreate,
	}
	initDoMaterialSiteReadingSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteReadingSummaryCmd.AddCommand(newDoMaterialSiteReadingSummaryCreateCmd())
}

func initDoMaterialSiteReadingSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to minute unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated, e.g. value_avg:desc). Defaults to value_avg:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all available metrics")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"material_site\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSiteReadingSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSiteReadingSummaryCreateOptions(cmd)
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

	filters, err := parseMaterialSiteReadingSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolveMaterialSiteReadingSummaryGroupBy(opts)
	sort := resolveMaterialSiteReadingSummarySort(opts)
	metrics := resolveMaterialSiteReadingSummaryMetrics(opts)

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
			"type":       "material-site-reading-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-site-reading-summaries", jsonBody)
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
	headers, values = selectMaterialSiteReadingSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := materialSiteReadingSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildMaterialSiteReadingSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderMaterialSiteReadingSummaryTable(cmd, headers, values)
}

func parseDoMaterialSiteReadingSummaryCreateOptions(cmd *cobra.Command) (doMaterialSiteReadingSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doMaterialSiteReadingSummaryCreateOptions{}, err
	}

	return doMaterialSiteReadingSummaryCreateOptions{
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

type materialSiteReadingSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildMaterialSiteReadingSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderMaterialSiteReadingSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material site reading summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatMaterialSiteReadingSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parseMaterialSiteReadingSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolveMaterialSiteReadingSummaryGroupBy(opts doMaterialSiteReadingSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"minute"}
}

func resolveMaterialSiteReadingSummarySort(opts doMaterialSiteReadingSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"value_avg:desc"}
}

func resolveMaterialSiteReadingSummaryMetrics(opts doMaterialSiteReadingSummaryCreateOptions) []string {
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

	return materialSiteReadingSummaryDefaultMetrics
}

func selectMaterialSiteReadingSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := materialSiteReadingSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := materialSiteReadingSummaryGroupByDisplayColumns[group]; ok {
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

func formatMaterialSiteReadingSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatMaterialSiteReadingSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatMaterialSiteReadingSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatMaterialSiteReadingSummaryNumber(header, typed)
	case float32:
		return formatMaterialSiteReadingSummaryNumber(header, float64(typed))
	case int:
		return formatMaterialSiteReadingSummaryNumber(header, float64(typed))
	case int64:
		return formatMaterialSiteReadingSummaryNumber(header, float64(typed))
	case int32:
		return formatMaterialSiteReadingSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatMaterialSiteReadingSummaryString(header, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lowerHeader := strings.ToLower(header)
	switch {
	case strings.HasSuffix(lowerHeader, "_name"), strings.HasSuffix(lowerHeader, "_external_id"):
		return truncateString(value, 35)
	default:
		return value
	}
}

func formatMaterialSiteReadingSummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_count"):
		return strconv.FormatInt(int64(value+0.5), 10)
	case strings.HasSuffix(lowerHeader, "_seconds"):
		return strconv.FormatInt(int64(value+0.5), 10)
	default:
		if value == float64(int64(value)) {
			return strconv.FormatInt(int64(value), 10)
		}
		return fmt.Sprintf("%.2f", value)
	}
}
