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

var jobProductionPlanSummaryDefaultMetrics = []string{
	"plan_count",
	"tons_sum",
	"truck_hours_sum",
	"cycle_count_sum",
}

var jobProductionPlanSummaryGroupByAllColumns = map[string][]string{
	"job_production_plan":              {"job_production_plan_id"},
	"date":                             {"date"},
	"month":                            {"month"},
	"dow":                              {"dow"},
	"week":                             {"week"},
	"year":                             {"year"},
	"project":                          {"project_id", "project_name"},
	"color_hex":                        {"color_hex"},
	"start_hour":                       {"start_hour"},
	"day_or_night":                     {"day_or_night"},
	"customer":                         {"customer_id", "customer_name"},
	"customer_is_controlled_by_broker": {"customer_is_controlled_by_broker"},
	"broker":                           {"broker_id", "broker_name"},
	"business_unit":                    {"business_unit_id", "business_unit_name"},
	"planner":                          {"planner_id", "planner_name"},
	"project_manager":                  {"project_manager_id", "project_manager_name"},
	"job_number":                       {"job_number"},
	"raw_job_number":                   {"raw_job_number"},
	"job_name":                         {"job_name"},
	"plan":                             {"plan"},
	"is_checksum_difference":           {"is_checksum_difference"},
	"is_tons":                          {"is_tons"},
	"status":                           {"status"},
	"is_contractor_not_default":        {"is_contractor_not_default"},
	"is_contractor_default":            {"is_contractor_default"},
	"contractor":                       {"contractor_id", "contractor_name"},
	"material_type_ultimate_parent_qualified_names": {"material_type_ultimate_parent_qualified_names"},
	"after": {"after"},
}

var jobProductionPlanSummaryGroupByDisplayColumns = map[string][]string{
	"job_production_plan":              {"job_production_plan_id"},
	"date":                             {"date"},
	"month":                            {"month"},
	"dow":                              {"dow"},
	"week":                             {"week"},
	"year":                             {"year"},
	"project":                          {"project_name"},
	"color_hex":                        {"color_hex"},
	"start_hour":                       {"start_hour"},
	"day_or_night":                     {"day_or_night"},
	"customer":                         {"customer_name"},
	"customer_is_controlled_by_broker": {"customer_is_controlled_by_broker"},
	"broker":                           {"broker_name"},
	"business_unit":                    {"business_unit_name"},
	"planner":                          {"planner_name"},
	"project_manager":                  {"project_manager_name"},
	"job_number":                       {"job_number"},
	"raw_job_number":                   {"raw_job_number"},
	"job_name":                         {"job_name"},
	"plan":                             {"plan"},
	"is_checksum_difference":           {"is_checksum_difference"},
	"is_tons":                          {"is_tons"},
	"status":                           {"status"},
	"is_contractor_not_default":        {"is_contractor_not_default"},
	"is_contractor_default":            {"is_contractor_default"},
	"contractor":                       {"contractor_name"},
	"material_type_ultimate_parent_qualified_names": {"material_type_ultimate_parent_qualified_names"},
	"after": {"after"},
}

type doJobProductionPlanSummaryCreateOptions struct {
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

func newDoJobProductionPlanSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan summary",
		Long: `Create a job production plan summary.

Aggregates production planning data for hauling jobs including tonnage targets,
actual production, truck hours, cycle counts, and schedule adherence. Use this
to compare planned vs actual production and analyze job performance.

REQUIRED: You must provide --start-on and --end-on date flags.

Group-by attributes:
  after                     Group by after flag
  broker                    Group by broker
  business_unit             Group by business unit
  color_hex                 Group by color
  contractor                Group by contractor
  customer                  Group by customer
  customer_is_controlled_by_broker  Group by customer control flag
  date                      Group by plan date
  day_or_night              Group by day or night
  dow                       Group by day of week
  is_checksum_difference    Group by checksum difference
  is_contractor_default     Group by contractor default flag
  is_contractor_not_default Group by contractor non-default flag
  is_tons                   Group by tons flag
  job_name                  Group by job name
  job_number                Group by job number
  job_production_plan       Group by job production plan ID
  material_type_ultimate_parent_qualified_names  Group by material type
  month                     Group by month
  plan                      Group by plan
  planner                   Group by planner
  project                   Group by project
  project_manager           Group by project manager
  raw_job_number            Group by raw job number
  start_hour                Group by start hour
  status                    Group by status
  week                      Group by week
  year                      Group by year

Metrics:
  By default: plan_count, tons_sum, truck_hours_sum, cycle_count_sum

  Available metrics include:
    plan_count, goal_tons_sum, tons_sum, ton_miles_sum,
    start_time_offset_minutes_sum, start_time_offset_minutes_mean,
    start_time_offset_minutes_median, tons_vs_goal_tons_avg,
    production_incident_net_impact_minutes_sum, raw_tons_sum, tons_avg,
    checksum_difference_sum, truck_count_sum, truck_count_avg,
    truck_hours_sum, truck_hours_avg, truck_hours_vs_truck_count_avg,
    truck_spend_sum, truck_spend_avg, truck_spend_per_ton,
    production_hours_sum, production_hours_avg, goal_production_hours_sum,
    goal_production_hours_avg, tons_per_productive_segment_avg,
    cycle_count_sum, cubic_yards_sum, gallons_sum, planned_tons_per_cycle,
    tons_per_cycle, calculated_miles_avg, cycle_minutes_avg,
    cycle_minutes_stddev_avg, cycle_minutes_decile_avg, cycle_minutes_median_avg,
    actual_practical_surplus_hours_sum, actual_practical_surplus_pct,
    actual_practical_trucking_hours_sum, planned_practical_surplus_hours_sum,
    planned_practical_surplus_pct, planned_practical_trucking_hours_sum

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    broker, business_unit, contractor, customer, date, job_name_or_number_like,
    job_number, material_type_ultimate_parent_qualified_names,
    material_type_ultimate_parent_count, planner, project, project_manager,
    status, timely_scheduled, and all group-by attributes

  Min/Max filters for metrics:
    goal_tons_min, goal_tons_max, production_hours_min, production_hours_max,
    tons_min, tons_max, truck_hours_min, truck_hours_max, etc.`,
		Example: `  # Summary grouped by customer
  xbe summarize job-production-plan-summary create --start-on 2025-01-01 --end-on 2025-01-31 --group-by customer --filter broker=123

  # Summary by project and date
  xbe summarize job-production-plan-summary create --start-on 2025-01-01 --end-on 2025-01-31 --group-by project,date --filter broker=123

  # Summary with all metrics
  xbe summarize job-production-plan-summary create --start-on 2025-01-01 --end-on 2025-01-31 --group-by customer --filter broker=123 --all-metrics

  # Total summary (no group-by)
  xbe summarize job-production-plan-summary create --start-on 2025-01-01 --end-on 2025-01-31 --group-by "" --filter broker=123

  # JSON output
  xbe summarize job-production-plan-summary create --start-on 2025-01-01 --end-on 2025-01-31 --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanSummaryCreate,
	}
	initDoJobProductionPlanSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanSummaryCmd.AddCommand(newDoJobProductionPlanSummaryCreateCmd())
}

func initDoJobProductionPlanSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("start-on", "", "Start date (YYYY-MM-DD) [required]")
	cmd.Flags().String("end-on", "", "End date (YYYY-MM-DD) [required]")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to customer unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated). Defaults to plan_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all metrics returned by the API")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanSummaryCreateOptions(cmd)
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

	filters, err := parseJobProductionPlanSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	groupBy := resolveJobProductionPlanSummaryGroupBy(opts)
	sort := resolveJobProductionPlanSummarySort(opts)
	metrics := resolveJobProductionPlanSummaryMetrics(opts)

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
			"type":       "job-production-plan-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-summaries", jsonBody)
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
	headers, values = selectJobProductionPlanSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := jobProductionPlanSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildJobProductionPlanSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderJobProductionPlanSummaryTable(cmd, headers, values)
}

func parseDoJobProductionPlanSummaryCreateOptions(cmd *cobra.Command) (doJobProductionPlanSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	startOn, err := cmd.Flags().GetString("start-on")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	endOn, err := cmd.Flags().GetString("end-on")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doJobProductionPlanSummaryCreateOptions{}, err
	}

	return doJobProductionPlanSummaryCreateOptions{
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

type jobProductionPlanSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildJobProductionPlanSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderJobProductionPlanSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatJobProductionPlanSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parseJobProductionPlanSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolveJobProductionPlanSummaryGroupBy(opts doJobProductionPlanSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"customer"}
}

func resolveJobProductionPlanSummarySort(opts doJobProductionPlanSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"plan_count:desc"}
}

func resolveJobProductionPlanSummaryMetrics(opts doJobProductionPlanSummaryCreateOptions) []string {
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

	return jobProductionPlanSummaryDefaultMetrics
}

func selectJobProductionPlanSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := jobProductionPlanSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := jobProductionPlanSummaryGroupByDisplayColumns[group]; ok {
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

func formatJobProductionPlanSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatJobProductionPlanSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatJobProductionPlanSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatJobProductionPlanSummaryNumber(header, typed)
	case float32:
		return formatJobProductionPlanSummaryNumber(header, float64(typed))
	case int:
		return formatJobProductionPlanSummaryNumber(header, float64(typed))
	case int64:
		return formatJobProductionPlanSummaryNumber(header, float64(typed))
	case int32:
		return formatJobProductionPlanSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatJobProductionPlanSummaryString(header, value string) string {
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

func formatJobProductionPlanSummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_pct"):
		percent := value
		if value <= 1.5 {
			percent = value * 100
		}
		return fmt.Sprintf("%.1f%%", percent)
	case strings.HasSuffix(lowerHeader, "_count") || lowerHeader == "plan_count":
		return strconv.FormatInt(int64(value+0.5), 10)
	case strings.Contains(lowerHeader, "spend") || strings.Contains(lowerHeader, "cost"):
		return fmt.Sprintf("$%.2f", value)
	case strings.Contains(lowerHeader, "tons") || strings.Contains(lowerHeader, "yards") || strings.Contains(lowerHeader, "gallons"):
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
