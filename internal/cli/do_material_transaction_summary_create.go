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

var materialTransactionSummaryDefaultMetrics = []string{
	"material_transaction_count",
	"tons_sum",
}

var materialTransactionSummaryGroupByAllColumns = map[string][]string{
	"material_type_fully_qualified_name_base": {"material_type_fully_qualified_name_base"},
	"broker":               {"broker_id", "broker_name"},
	"business_unit":        {"business_unit_id", "business_unit_name"},
	"customer":             {"customer_id", "customer_name"},
	"trucker":              {"trucker_id", "trucker_name"},
	"customer_segment":     {"customer_segment"},
	"date":                 {"date"},
	"month":                {"month"},
	"year":                 {"year"},
	"day_of_year":          {"day_of_year"},
	"day_of_week":          {"day_of_week"},
	"week_of_year":         {"week_of_year"},
	"status":               {"status"},
	"hour":                 {"hour"},
	"material_type":        {"material_type_id", "material_type_name", "material_type_display_name"},
	"material_site":        {"material_site_id", "material_site_name"},
	"material_supplier":    {"material_supplier_id", "material_supplier_name"},
	"project":              {"project_id", "project_name"},
	"job_production_plan":  {"job_production_plan_id", "job_production_plan_name"},
	"planner":              {"planner_id", "planner_name"},
	"project_manager":      {"project_manager_id", "project_manager_name"},
	"developer":            {"developer_id", "developer_name"},
	"direction":            {"direction"},
	"job_site":             {"job_site_id", "job_site_name"},
	"is_material_supplier_controlled_by_broker": {"is_material_supplier_controlled_by_broker"},
}

var materialTransactionSummaryGroupByDisplayColumns = map[string][]string{
	"material_type_fully_qualified_name_base": {"material_type_fully_qualified_name_base"},
	"broker":               {"broker_name"},
	"business_unit":        {"business_unit_name"},
	"customer":             {"customer_name"},
	"trucker":              {"trucker_name"},
	"customer_segment":     {"customer_segment"},
	"date":                 {"date"},
	"month":                {"month"},
	"year":                 {"year"},
	"day_of_year":          {"day_of_year"},
	"day_of_week":          {"day_of_week"},
	"week_of_year":         {"week_of_year"},
	"status":               {"status"},
	"hour":                 {"hour"},
	"material_type":        {"material_type_name"},
	"material_site":        {"material_site_name"},
	"material_supplier":    {"material_supplier_name"},
	"project":              {"project_name"},
	"job_production_plan":  {"job_production_plan_name"},
	"planner":              {"planner_name"},
	"project_manager":      {"project_manager_name"},
	"developer":            {"developer_name"},
	"direction":            {"direction"},
	"job_site":             {"job_site_name"},
	"is_material_supplier_controlled_by_broker": {"is_material_supplier_controlled_by_broker"},
}

type doMaterialTransactionSummaryCreateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	GroupByRaw      string
	GroupBySet      bool
	SortRaw         string
	SortSet         bool
	Limit           int
	MetricsRaw      string
	MetricsSet      bool
	Metrics         []string
	AllMetrics      bool
	MinTransactions int
	FiltersJSON     string
	FilterPairs     []string
}

func newDoMaterialTransactionSummaryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction summary",
		Long: `Create a material transaction summary.

This command posts to the material transaction summaries endpoint and returns
aggregated metrics grouped by the selected attributes.

Group-by attributes:
  broker                                 Group by broker (broker_id, broker_name)
  business_unit                          Group by business unit (business_unit_id, business_unit_name)
  customer                               Group by customer (customer_id, customer_name)
  customer_segment                       Group by customer segment (internal/external)
  date                                   Group by transaction date
  day_of_week                            Group by day of week (0=Sunday, 6=Saturday)
  day_of_year                            Group by day of year (1-366)
  developer                              Group by developer (developer_id, developer_name)
  direction                              Group by direction (inbound/outbound/other)
  hour                                   Group by hour of day (0-23)
  is_material_supplier_controlled_by_broker  Group by supplier control flag
  job_production_plan                    Group by job production plan (job_production_plan_id, job_production_plan_name)
  job_site                               Group by job site (job_site_id, job_site_name)
  material_site                          Group by material site (material_site_id, material_site_name)
  material_supplier                      Group by material supplier (material_supplier_id, material_supplier_name)
  material_type                          Group by material type (material_type_id, material_type_name, material_type_display_name)
  material_type_fully_qualified_name_base  Group by material type base name
  month                                  Group by month (1-12)
  planner                                Group by planner (planner_id, planner_name)
  project                                Group by project (project_id, project_name)
  project_manager                        Group by project manager (project_manager_id, project_manager_name)
  status                                 Group by transaction status
  trucker                                Group by trucker (trucker_id, trucker_name)
  week_of_year                           Group by week of year (1-53)
  year                                   Group by year

Metrics:
  By default the CLI shows: material_transaction_count, tons_sum

  Available metrics:
    material_transaction_count    Count of material transactions
    tons_sum                      Sum of tonnage
    tons_avg                      Average tonnage per transaction

  Use --metrics or --metric to customize the metric columns, or --all-metrics
  to include every available metric.

Filters:
  Use --filter key=value (repeatable) or --filters '{"key":"value"}'.

  Available filters:
    broker                                   Broker ID
    business_unit                            Business unit ID
    customer                                 Customer ID
    customer_segment                         Customer segment (internal/external)
    date                                     Specific date (YYYY-MM-DD)
    date_min                                 Minimum date (YYYY-MM-DD)
    date_max                                 Maximum date (YYYY-MM-DD)
    day_of_week                              Day of week (0-6)
    day_of_year                              Day of year (1-366)
    developer                                Developer ID
    direction                                Direction (inbound/outbound/other)
    hour                                     Hour of day (0-23)
    is_material_supplier_controlled_by_broker  Supplier control flag (true/false)
    job_production_plan                      Job production plan ID
    job_site                                 Job site ID
    material_site                            Material site ID
    material_supplier                        Material supplier ID
    material_type                            Material type ID
    material_type_fully_qualified_name_base  Material type base name
    month                                    Month (1-12)
    planner                                  Planner ID
    project                                  Project ID
    project_manager                          Project manager ID
    status                                   Transaction status
    trucker                                  Trucker ID
    week_of_year                             Week of year (1-53)
    year                                     Year

  Metric filters:
    material_transaction_count__min          Minimum transaction count`,
		Example: `  # Summary grouped by material site
  xbe do material-transaction-summary create --group-by material_site --filter broker=123 --filter date_min=2025-01-01 --filter date_max=2025-01-31

  # Summary by customer segment
  xbe do material-transaction-summary create --group-by customer_segment --filter broker=123

  # Summary by date
  xbe do material-transaction-summary create --group-by date --filter broker=123 --filter date_min=2025-01-01 --filter date_max=2025-01-31

  # Summary by direction (inbound/outbound)
  xbe do material-transaction-summary create --group-by direction --filter broker=123

  # Summary by material type with all metrics
  xbe do material-transaction-summary create --group-by material_type --filter broker=123 --all-metrics

  # Summary by trucker and customer
  xbe do material-transaction-summary create --group-by trucker,customer --filter broker=123

  # Summary by month and year
  xbe do material-transaction-summary create --group-by year,month --filter broker=123 --filter year=2025

  # Limit to high-volume results only
  xbe do material-transaction-summary create --filter broker=123 --min-transactions 10

  # Total summary (no group-by)
  xbe do material-transaction-summary create --group-by "" --filter broker=123 --filter date_min=2025-01-01

  # JSON output
  xbe do material-transaction-summary create --filter broker=123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionSummaryCreate,
	}
	initDoMaterialTransactionSummaryCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionSummaryCmd.AddCommand(newDoMaterialTransactionSummaryCreateCmd())
}

func initDoMaterialTransactionSummaryCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("group-by", "", "Group by attributes (comma-separated). Defaults to material_site unless set.")
	cmd.Flags().String("sort", "", "Sort fields (comma-separated, e.g. tons_sum:desc). Defaults to material_transaction_count:desc unless set.")
	cmd.Flags().Int("limit", 0, "Limit number of rows returned")
	cmd.Flags().String("metrics", "", "Metric columns to include (comma-separated)")
	cmd.Flags().StringArray("metric", nil, "Metric column to include (repeatable)")
	cmd.Flags().Bool("all-metrics", false, "Include all available metrics")
	cmd.Flags().Int("min-transactions", 0, "Drop rows with fewer than this many transactions")
	cmd.Flags().String("filters", "", "Filters JSON object (e.g. '{\"broker\":\"123\"}')")
	cmd.Flags().StringArray("filter", nil, "Filter in key=value format (repeatable)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionSummaryCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTransactionSummaryCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication
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

	filters, err := parseMaterialTransactionSummaryFilters(opts.FiltersJSON, opts.FilterPairs)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.MinTransactions > 0 {
		filters["material_transaction_count__min"] = opts.MinTransactions
	}

	groupBy := resolveMaterialTransactionSummaryGroupBy(opts)
	sort := resolveMaterialTransactionSummarySort(opts)
	metrics := resolveMaterialTransactionSummaryMetrics(opts)

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
			"type":       "material-transaction-summaries",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-summaries", jsonBody)
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
	headers, values = selectMaterialTransactionSummaryColumns(headers, values, groupBy, metrics, opts.AllMetrics)

	if opts.JSON {
		output := materialTransactionSummaryOutput{
			Headers: headers,
			Values:  values,
			Rows:    buildMaterialTransactionSummaryRows(headers, values),
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	return renderMaterialTransactionSummaryTable(cmd, headers, values)
}

func parseDoMaterialTransactionSummaryCreateOptions(cmd *cobra.Command) (doMaterialTransactionSummaryCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	groupByRaw, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	sortRaw, err := cmd.Flags().GetString("sort")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	metricsRaw, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	metrics, err := cmd.Flags().GetStringArray("metric")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	allMetrics, err := cmd.Flags().GetBool("all-metrics")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	minTransactions, err := cmd.Flags().GetInt("min-transactions")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	filtersJSON, err := cmd.Flags().GetString("filters")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	filterPairs, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doMaterialTransactionSummaryCreateOptions{}, err
	}

	return doMaterialTransactionSummaryCreateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		GroupByRaw:      groupByRaw,
		GroupBySet:      cmd.Flags().Changed("group-by"),
		SortRaw:         sortRaw,
		SortSet:         cmd.Flags().Changed("sort"),
		Limit:           limit,
		MetricsRaw:      metricsRaw,
		MetricsSet:      cmd.Flags().Changed("metrics"),
		Metrics:         metrics,
		AllMetrics:      allMetrics,
		MinTransactions: minTransactions,
		FiltersJSON:     filtersJSON,
		FilterPairs:     filterPairs,
	}, nil
}

type materialTransactionSummaryOutput struct {
	Headers []string         `json:"headers"`
	Values  [][]any          `json:"values"`
	Rows    []map[string]any `json:"rows,omitempty"`
}

func buildMaterialTransactionSummaryRows(headers []string, values [][]any) []map[string]any {
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

func renderMaterialTransactionSummaryTable(cmd *cobra.Command, headers []string, values [][]any) error {
	if len(headers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No headers returned.")
		return nil
	}
	if len(values) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material transaction summary data found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, strings.Join(headers, "\t"))

	for _, row := range values {
		cols := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				cols[i] = formatMaterialTransactionSummaryValue(headers[i], row[i])
			} else {
				cols[i] = ""
			}
		}
		fmt.Fprintln(writer, strings.Join(cols, "\t"))
	}

	return writer.Flush()
}

func parseMaterialTransactionSummaryFilters(rawJSON string, pairs []string) (map[string]any, error) {
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

func resolveMaterialTransactionSummaryGroupBy(opts doMaterialTransactionSummaryCreateOptions) []string {
	if opts.GroupBySet {
		return splitCommaList(opts.GroupByRaw)
	}
	return []string{"material_site"}
}

func resolveMaterialTransactionSummarySort(opts doMaterialTransactionSummaryCreateOptions) []string {
	if opts.SortSet {
		return splitCommaList(opts.SortRaw)
	}
	return []string{"material_transaction_count:desc"}
}

func resolveMaterialTransactionSummaryMetrics(opts doMaterialTransactionSummaryCreateOptions) []string {
	if opts.AllMetrics {
		return []string{"material_transaction_count", "tons_sum", "tons_avg"}
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

	return materialTransactionSummaryDefaultMetrics
}

func selectMaterialTransactionSummaryColumns(headers []string, values [][]any, groupBy []string, metrics []string, allMetrics bool) ([]string, [][]any) {
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
		if cols, ok := materialTransactionSummaryGroupByAllColumns[group]; ok {
			for _, col := range cols {
				fullGroupByColumns[col] = struct{}{}
			}
		} else if group != "" {
			fullGroupByColumns[group] = struct{}{}
		}

		if cols, ok := materialTransactionSummaryGroupByDisplayColumns[group]; ok {
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

func formatMaterialTransactionSummaryValue(header string, value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return formatMaterialTransactionSummaryString(header, typed)
	case json.Number:
		if number, err := typed.Float64(); err == nil {
			return formatMaterialTransactionSummaryNumber(header, number)
		}
		return typed.String()
	case float64:
		return formatMaterialTransactionSummaryNumber(header, typed)
	case float32:
		return formatMaterialTransactionSummaryNumber(header, float64(typed))
	case int:
		return formatMaterialTransactionSummaryNumber(header, float64(typed))
	case int64:
		return formatMaterialTransactionSummaryNumber(header, float64(typed))
	case int32:
		return formatMaterialTransactionSummaryNumber(header, float64(typed))
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatMaterialTransactionSummaryString(header, value string) string {
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

func formatMaterialTransactionSummaryNumber(header string, value float64) string {
	lowerHeader := strings.ToLower(header)

	switch {
	case strings.HasSuffix(lowerHeader, "_count"):
		return strconv.FormatInt(int64(value+0.5), 10)
	case strings.Contains(lowerHeader, "tons"):
		return fmt.Sprintf("%.2f", value)
	default:
		if value == float64(int64(value)) {
			return strconv.FormatInt(int64(value), 10)
		}
		return fmt.Sprintf("%.2f", value)
	}
}
