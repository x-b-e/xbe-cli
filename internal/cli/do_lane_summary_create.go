package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

var laneSummaryDefaultMetrics = []string{
	"cycle_count",
	"material_transaction_count",
	"cycle_minutes_median",
	"calculated_travel_minutes_median",
	"tons_sum",
}

var laneSummaryGroupByAllColumns = map[string][]string{
	"material_transaction": {"material_transaction_id"},
	"material_type":        {"material_type_id", "material_type_name"},
	"broker":               {"broker_id", "broker_name"},
	"customer":             {"customer_id", "customer_name"},
	"trucker":              {"trucker_id", "trucker_name"},
	"driver":               {"driver_id", "driver_name"},
	"trailer":              {"trailer_id", "trailer_number"},
	"job_production_plan":  {"job_production_plan_id", "job_name", "job_number"},
	"job_site":             {"job_site_id", "job_site_name"},
	"material_supplier":    {"material_supplier_id", "material_supplier_name"},
	"material_site":        {"material_site_id", "material_site_name"},
	"business_unit":        {"business_unit_id", "business_unit_name", "business_unit_external_id"},
	"origin":               {"origin_id", "origin_name", "origin_organization_name", "origin_type", "origin_latitude", "origin_longitude"},
	"destination":          {"destination_id", "destination_name", "destination_organization_name", "destination_type", "destination_latitude", "destination_longitude"},
	"date":                 {"date"},
	"month":                {"month"},
	"year":                 {"year"},
	"hour":                 {"hour"},
}

var laneSummaryGroupByDisplayColumns = map[string][]string{
	"material_transaction": {"material_transaction_id"},
	"material_type":        {"material_type_name"},
	"broker":               {"broker_name"},
	"customer":             {"customer_name"},
	"trucker":              {"trucker_name"},
	"driver":               {"driver_name"},
	"trailer":              {"trailer_number"},
	"job_production_plan":  {"job_name", "job_number"},
	"job_site":             {"job_site_name"},
	"material_supplier":    {"material_supplier_name"},
	"material_site":        {"material_site_name"},
	"business_unit":        {"business_unit_name"},
	"origin":               {"origin_name"},
	"destination":          {"destination_name"},
	"date":                 {"date"},
	"month":                {"month"},
	"year":                 {"year"},
	"hour":                 {"hour"},
}

type doLaneSummaryCreateOptions struct {
	BaseURL                            string
	Token                              string
	JSON                               bool
	GroupByRaw                         string
	GroupBySet                         bool
	SortRaw                            string
	SortSet                            bool
	Limit                              int
	MetricsRaw                         string
	MetricsSet                         bool
	Metrics                            []string
	AllMetrics                         bool
	MinTransactions                    int
	FiltersJSON                        string
	FilterPairs                        []string
	OptionalFeatures                   []string
	IncludeDriverMovementDurations     bool
	UseDriverDayTripLeadMinutes        bool
	BetaDriverMovementSegmentDurations bool
}

func newDoLaneSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lane summary (cycle summary)",
		Long: `Create a lane summary (cycle summary).

This command posts to the cycle summaries endpoint and returns headers/values for
cycle metrics grouped by the selected attributes.

Group-by attributes (examples):
  material_transaction, material_type, broker, customer, trucker, driver, trailer,
  job_production_plan, job_site, material_supplier, material_site, business_unit,
  origin, destination, date, month, year, hour

Metrics:
  By default the CLI shows the same compact set as the Cycle Summary report:
  cycle_count, material_transaction_count, cycle_minutes_median,
  calculated_travel_minutes_median, tons_sum

  Use --metrics or --metric to customize the metric columns, or --all-metrics
  to include every metric returned by the API.

  Available metrics:
    cycle_count
    material_transaction_count
    cycle_minutes_median
    cycle_minutes_decile
    cycle_minutes_min
    job_site_wait_time_median
    material_site_wait_time_median
    other_site_wait_time_median
    calculated_travel_miles_median
    calculated_travel_miles_mean
    calculated_travel_miles_min
    calculated_travel_miles_max
    calculated_travel_minutes_median
    calculated_travel_minutes_mean
    calculated_travel_minutes_min
    calculated_travel_minutes_max
    pickup_dwell_minutes_mean
    pickup_dwell_minutes_min
    pickup_dwell_minutes_max
    pickup_dwell_minutes_median
    pickup_dwell_minutes_p90
    loaded_driving_minutes_mean
    loaded_driving_minutes_min
    loaded_driving_minutes_max
    loaded_driving_minutes_median
    loaded_driving_minutes_p90
    loaded_driving_minutes_total
    delivery_dwell_minutes_mean
    delivery_dwell_minutes_min
    delivery_dwell_minutes_max
    delivery_dwell_minutes_median
    delivery_dwell_minutes_p90
    unloaded_driving_minutes_mean
    unloaded_driving_minutes_min
    unloaded_driving_minutes_max
    unloaded_driving_minutes_median
    unloaded_driving_minutes_p90
    unloaded_driving_minutes_total
    cycle_pct
    tons_sum
    tons_mean
    cost_per_ton
    effective_cost_per_hour_median
    calculated_non_travel_minutes_median
    calculated_non_travel_minutes_decile
    dwell_and_drive_count

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    material_transaction
    material_type
    broker
    customer
    trucker
    driver
    trailer
    material_supplier
    material_site
    business_unit
    job_production_plan
    job_number
    job_name
    job_site
    origin (Type|ID, e.g. MaterialSite|123 or JobSite|456)
    destination (Type|ID, e.g. MaterialSite|123 or JobSite|456)
    has_trip (true/false)
    is-managed (true/false)
    material_type_ultimate_parent_qualified_names
    date
    date_min
    date_max
    transaction_at_min
    transaction_at_max
    month
    year
    hour

  Metric filters:
    material_transaction_count__min

Optional features:
  --include-driver-movement-durations (cycle_summary_include_dmd)
  --use-driver-day-trip-lead-minutes (cycle_summary_use_driver_day_trip_lead_minutes)
  --beta-driver-movement-segment-durations (cycle_summary_beta_driver_movement_segment_durations)
  --optional-feature <name> (repeatable)`,
		Example: `  # Lane summary grouped by origin and destination
  xbe summarize lane-summary create --group-by origin,destination --filter broker=123 --filter transaction_at_min=2025-01-17T00:00:00Z --filter transaction_at_max=2025-01-17T23:59:59Z

  # Total summary (no group-by)
  xbe summarize lane-summary create --group-by "" --filter broker=123 --filter date_min=2025-01-01 --filter date_max=2025-01-31

  # Include optional features
  xbe summarize lane-summary create --filter broker=123 --use-driver-day-trip-lead-minutes --beta-driver-movement-segment-durations

  # Limit to high-volume lanes only
  xbe summarize lane-summary create --filter broker=123 --min-transactions 25

  # Customize metrics
  xbe summarize lane-summary create --filter broker=123 --metrics cycle_minutes_median,tons_sum

  # JSON output with explicit filters
  xbe summarize lane-summary create --filters '{"broker":"123","has_trip":true}' --json`,
		Args: cobra.NoArgs,
		RunE: runDoLaneSummaryCreate,
	}
	initDoLaneSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doLaneSummaryCmd.AddCommand(newDoLaneSummaryCreateCmd())
}

func initDoLaneSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to origin,destination unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to material_transaction_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().Int("min-transactions", 0, "Drop rows with fewer than this many transactions")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().Bool("include-driver-movement-durations", false, "Enable driver movement durations (cycle_summary_include_dmd)")
	cmd.Flags().Bool("use-driver-day-trip-lead-minutes", false, "Enable driver day trip lead minutes metrics")
	cmd.Flags().Bool("beta-driver-movement-segment-durations", false, "Enable beta driver movement segment durations")
	cmd.Flags().StringArray("optional-feature", nil, "Additional optional features (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoLaneSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoLaneSummaryCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	filters, err := parseLaneSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.MinTransactions > 0 {
		filters["material_transaction_count__min"] = opts.MinTransactions
	}

	groupBy := resolveLaneSummaryGroupBy(opts)
	sort := resolveLaneSummarySort(opts)
	metrics := resolveLaneSummaryMetrics(opts)

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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "cycle-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	optionalFeatures := buildLaneSummaryOptionalFeatures(opts)
	query := url.Values{}
	if len(optionalFeatures) > 0 {
		query.Set("meta[optional-features]", strings.Join(optionalFeatures, ","))
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.PostWithQuery(cmd.Context(), "/v1/cycle-summaries", query, jsonBody)
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
	headers, values = selectLaneSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics, cmd.ErrOrStderr())

	if opts.JSON {
		output := laneSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildLaneSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderLaneSummaryTable(cmd, headers, values)
}

func parseDoLaneSummaryCreateOptions(cmd *cobra.Command) (doLaneSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	minTransactions, err := cmd.Flags().GetInt("min-transactions")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	optionalFeatures, err := cmd.Flags().GetStringArray("optional-feature")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	includeDriverMovementDurations, err := cmd.Flags().GetBool("include-driver-movement-durations")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	useDriverDayTripLeadMinutes, err := cmd.Flags().GetBool("use-driver-day-trip-lead-minutes")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	betaDriverMovementSegmentDurations, err := cmd.Flags().GetBool("beta-driver-movement-segment-durations")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doLaneSummaryCreateOptions{}, err
	}

	return doLaneSummaryCreateOptions{
		BaseURL:                            baseURL,
		Token:                              token,
		JSON:                               jsonOut,
		GroupByRaw:                         groupByRaw,
		GroupBySet:                         cmd.Flags().Changed("group-by"),
		SortRaw:                            sortRaw,
		SortSet:                            cmd.Flags().Changed("sort"),
		Limit:                              limit,
		MetricsRaw:                         metricsRaw,
		MetricsSet:                         cmd.Flags().Changed("metrics"),
		Metrics:                            metrics,
		AllMetrics:                         allMetrics,
		MinTransactions:                    minTransactions,
		FiltersJSON:                        filtersJSON,
		FilterPairs:                        filterPairs,
		OptionalFeatures:                   optionalFeatures,
		IncludeDriverMovementDurations:     includeDriverMovementDurations,
		UseDriverDayTripLeadMinutes:        useDriverDayTripLeadMinutes,
		BetaDriverMovementSegmentDurations: betaDriverMovementSegmentDurations,
	}, nil
}

type laneSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildLaneSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderLaneSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lane summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatLaneSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parseLaneSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolveLaneSummaryGroupBy(opts doLaneSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"origin", "destination"}
}

func resolveLaneSummarySort(opts doLaneSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"material_transaction_count:desc"}
}

func resolveLaneSummaryMetrics(opts doLaneSummaryCreateOptions) []string {
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

	return laneSummaryDefaultMetrics
}

func selectLaneSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool, warn io.Writer) ([]string, [][]any) {
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
		if cols, ok := laneSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := laneSummaryGroupByDisplayColumns[group]; ok {
			displayColumns = append(displayColumns, cols...)
		} else if group != "" {
			displayColumns = append(displayColumns, group)
		}
	}

	seen := make(map[string]struct{}, len(headers))
	selectedHeaders := []string{}
	selectedIndexes := []int{}
	missingColumns := []string{}

	addHeader := func(header string) {
		if header == "" {
			return
		}
		if _, ok := seen[header]; ok {
			return
		}
		idx, ok := headerIndex[header]
		if !ok {
			missingColumns = append(missingColumns, header)
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

	if len(missingColumns) > 0 && warn != nil {
		fmt.Fprintf(warn, "Warning: requested columns not found: %s\n", strings.Join(missingColumns, ", "))
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

func buildLaneSummaryOptionalFeatures(opts doLaneSummaryCreateOptions) []string {
	features := append([]string{}, opts.OptionalFeatures...)
	if opts.IncludeDriverMovementDurations {
		features = append(features, "cycle_summary_include_dmd")
	}
	if opts.UseDriverDayTripLeadMinutes {
		features = append(features, "cycle_summary_use_driver_day_trip_lead_minutes")
	}
	if opts.BetaDriverMovementSegmentDurations {
		features = append(features, "cycle_summary_beta_driver_movement_segment_durations")
	}
	return uniqueStrings(features)
}

func splitCommaList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func valuesAttr(attrs map[string]any, key string) [][]any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}

	rows := [][]any{}
	typed, ok := value.([]any)
	if !ok {
		return nil
	}

	for _, row := range typed {
		if rowSlice, ok := row.([]any); ok {
			rows = append(rows, rowSlice)
		} else {
			rows = append(rows, []any{row})
		}
	}

	if len(rows) == 0 {
		return nil
	}
	return rows
}

func formatLaneSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatLaneSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatLaneSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatLaneSummaryNumber(header, typed)
	case float32:
		return formatLaneSummaryNumber(header, float64(typed))
	case int:
		return formatLaneSummaryNumber(header, float64(typed))
	case int64:
		return formatLaneSummaryNumber(header, float64(typed))
	case int32:
		return formatLaneSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatLaneSummaryString(header, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lowerHeader := strings.ToLower(header)
	switch {
	case strings.Contains(lowerHeader, "organization_name"):
		return truncateString(value, 32)
	case strings.HasSuffix(lowerHeader, "_name"):
		return truncateString(value, 28)
	case strings.HasSuffix(lowerHeader, "_type"):
		return truncateString(value, 14)
	default:
		return value
	}
}

func formatLaneSummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.Contains(lowerHeader, "latitude") || strings.Contains(lowerHeader, "longitude"):
		return strconv.FormatFloat(value, 'f', 5, 64)
	case strings.HasSuffix(lowerHeader, "_pct"):
		percent := value
		if value <= 1.5 {
			percent = value * 100
		}
		return fmt.Sprintf("%.1f%%", percent)
	case strings.HasSuffix(lowerHeader, "_count"):
		return strconv.FormatInt(int64(value+0.5), 10)
	case strings.Contains(lowerHeader, "cost_per") || strings.Contains(lowerHeader, "effective_cost_per_hour"):
		return fmt.Sprintf("$%.2f", value)
	case strings.Contains(lowerHeader, "tons"):
		return fmt.Sprintf("%.2f", value)
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
