package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doMaterialTransactionRateSummariesCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	MaterialSite            string
	StartAt                 string
	EndAt                   string
	GroupBy                 string
	MaterialTypeHierarchies string
	Sparse                  bool
}

type materialTransactionRateSummaryRow struct {
	MaterialSiteID string  `json:"material_site_id,omitempty"`
	Date           string  `json:"date,omitempty"`
	Hour           int     `json:"hour,omitempty"`
	TonsSum        float64 `json:"tons_sum,omitempty"`
}

type materialTransactionRateSummary struct {
	ID                          string
	MaterialSiteID              string
	TimeZoneID                  string
	StartAt                     string
	EndAt                       string
	GroupBy                     string
	MaterialTypeHierarchies     string
	Results                     []materialTransactionRateSummaryRow
	SparseResults               []materialTransactionRateSummaryRow
	DescriptiveStatistics       map[string]any
	SparseDescriptiveStatistics map[string]any
}

type materialTransactionRateSummaryOutput struct {
	MaterialSiteID          string                              `json:"material_site_id,omitempty"`
	TimeZoneID              string                              `json:"time_zone_id,omitempty"`
	StartAt                 string                              `json:"start_at,omitempty"`
	EndAt                   string                              `json:"end_at,omitempty"`
	GroupBy                 string                              `json:"group_by,omitempty"`
	MaterialTypeHierarchies string                              `json:"material_type_hierarchies,omitempty"`
	Sparse                  bool                                `json:"sparse"`
	Results                 []materialTransactionRateSummaryRow `json:"results,omitempty"`
	DescriptiveStatistics   map[string]any                      `json:"descriptive_statistics,omitempty"`
}

func newDoMaterialTransactionRateSummariesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material transaction rate summary",
		Long: `Create a material transaction rate summary.

Material transaction rate summaries return hourly tons for a material site.
Provide a date range to limit the summary window. Results are grouped by hour
and include descriptive statistics for tons.

Required flags:
  --material-site  Material site ID (required)

Optional flags:
  --start-at                   Start timestamp (ISO-8601)
  --end-at                     End timestamp (ISO-8601)
  --group-by                   Group by interval (default: hour)
  --material-type-hierarchies  Material type hierarchies (comma-separated)
  --sparse                      Use sparse results (exclude zero-ton rows)`,
		Example: `  # Generate hourly rate summary for a material site
  xbe do material-transaction-rate-summaries create --material-site 123 --start-at 2025-01-01T00:00:00Z --end-at 2025-01-02T00:00:00Z

  # Filter by material type hierarchy
  xbe do material-transaction-rate-summaries create --material-site 123 --material-type-hierarchies "aggregate,asphalt"

  # Use sparse results (non-zero rows only)
  xbe do material-transaction-rate-summaries create --material-site 123 --start-at 2025-01-01T00:00:00Z --end-at 2025-01-02T00:00:00Z --sparse

  # Output JSON
  xbe do material-transaction-rate-summaries create --material-site 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTransactionRateSummariesCreate,
	}
	initDoMaterialTransactionRateSummariesCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTransactionRateSummariesCmd.AddCommand(newDoMaterialTransactionRateSummariesCreateCmd())
}

func initDoMaterialTransactionRateSummariesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-site", "", "Material site ID (required)")
	cmd.Flags().String("start-at", "", "Start timestamp (ISO-8601)")
	cmd.Flags().String("end-at", "", "End timestamp (ISO-8601)")
	cmd.Flags().String("group-by", "hour", "Group by interval (default: hour)")
	cmd.Flags().String("material-type-hierarchies", "", "Material type hierarchies (comma-separated)")
	cmd.Flags().Bool("sparse", false, "Use sparse results (exclude zero-ton rows)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTransactionRateSummariesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTransactionRateSummariesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.MaterialSite) == "" {
		err := fmt.Errorf("--material-site is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.GroupBy) == "" {
		err := fmt.Errorf("--group-by cannot be empty")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"group-by": opts.GroupBy,
	}
	if strings.TrimSpace(opts.StartAt) != "" {
		attributes["start-at"] = opts.StartAt
	}
	if strings.TrimSpace(opts.EndAt) != "" {
		attributes["end-at"] = opts.EndAt
	}
	if strings.TrimSpace(opts.MaterialTypeHierarchies) != "" {
		attributes["material-type-hierarchies"] = opts.MaterialTypeHierarchies
	}

	relationships := map[string]any{
		"material-site": map[string]any{
			"data": map[string]any{
				"type": "material-sites",
				"id":   opts.MaterialSite,
			},
		},
	}

	data := map[string]any{
		"type":          "material-transaction-rate-summaries",
		"relationships": relationships,
		"attributes":    attributes,
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-transaction-rate-summaries", jsonBody)
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

	summary := materialTransactionRateSummaryFromSingle(resp)
	results := summary.Results
	stats := summary.DescriptiveStatistics
	if opts.Sparse {
		results = summary.SparseResults
		stats = summary.SparseDescriptiveStatistics
	}

	if opts.JSON {
		output := materialTransactionRateSummaryOutput{
			MaterialSiteID:          summary.MaterialSiteID,
			TimeZoneID:              summary.TimeZoneID,
			StartAt:                 summary.StartAt,
			EndAt:                   summary.EndAt,
			GroupBy:                 summary.GroupBy,
			MaterialTypeHierarchies: summary.MaterialTypeHierarchies,
			Sparse:                  opts.Sparse,
			Results:                 results,
			DescriptiveStatistics:   stats,
		}
		return writeJSON(cmd.OutOrStdout(), output)
	}

	summary.Results = results
	summary.DescriptiveStatistics = stats
	return renderMaterialTransactionRateSummary(cmd, summary, opts.Sparse)
}

func parseDoMaterialTransactionRateSummariesCreateOptions(cmd *cobra.Command) (doMaterialTransactionRateSummariesCreateOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}
	materialSite, err := cmd.Flags().GetString("material-site")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}
	startAt, err := cmd.Flags().GetString("start-at")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}
	endAt, err := cmd.Flags().GetString("end-at")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}
	groupBy, err := cmd.Flags().GetString("group-by")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}
	materialTypeHierarchies, err := cmd.Flags().GetString("material-type-hierarchies")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}
	sparse, err := cmd.Flags().GetBool("sparse")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return doMaterialTransactionRateSummariesCreateOptions{}, err
	}

	return doMaterialTransactionRateSummariesCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		MaterialSite:            materialSite,
		StartAt:                 startAt,
		EndAt:                   endAt,
		GroupBy:                 groupBy,
		MaterialTypeHierarchies: materialTypeHierarchies,
		Sparse:                  sparse,
	}, nil
}

func materialTransactionRateSummaryFromSingle(resp jsonAPISingleResponse) materialTransactionRateSummary {
	attrs := resp.Data.Attributes

	summary := materialTransactionRateSummary{
		ID:                          resp.Data.ID,
		StartAt:                     formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                       formatDateTime(stringAttr(attrs, "end-at")),
		GroupBy:                     stringAttr(attrs, "group-by"),
		MaterialTypeHierarchies:     stringAttr(attrs, "material-type-hierarchies"),
		TimeZoneID:                  stringAttr(attrs, "time-zone-id"),
		Results:                     parseMaterialTransactionRateSummaryRows(attrs["results"]),
		SparseResults:               parseMaterialTransactionRateSummaryRows(attrs["sparse-results"]),
		DescriptiveStatistics:       mapStringAnyAttr(attrs, "descriptive-statistics"),
		SparseDescriptiveStatistics: mapStringAnyAttr(attrs, "sparse-descriptive-statistics"),
	}

	if rel, ok := resp.Data.Relationships["material-site"]; ok && rel.Data != nil {
		summary.MaterialSiteID = rel.Data.ID
	}

	return summary
}

func parseMaterialTransactionRateSummaryRows(value any) []materialTransactionRateSummaryRow {
	switch typed := value.(type) {
	case []materialTransactionRateSummaryRow:
		return typed
	case []map[string]any:
		rows := make([]materialTransactionRateSummaryRow, 0, len(typed))
		for _, row := range typed {
			rows = append(rows, materialTransactionRateSummaryRowFromMap(row))
		}
		return rows
	case []any:
		rows := make([]materialTransactionRateSummaryRow, 0, len(typed))
		for _, item := range typed {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			rows = append(rows, materialTransactionRateSummaryRowFromMap(row))
		}
		return rows
	default:
		return nil
	}
}

func materialTransactionRateSummaryRowFromMap(row map[string]any) materialTransactionRateSummaryRow {
	materialSiteID := stringFromAny(rowValue(row, "material_site_id", "material-site-id"))
	date := stringFromAny(rowValue(row, "date"))
	hour := intFromAny(rowValue(row, "hour"))
	tonsSum := floatFromAny(rowValue(row, "tons_sum", "tons-sum"))

	return materialTransactionRateSummaryRow{
		MaterialSiteID: materialSiteID,
		Date:           date,
		Hour:           hour,
		TonsSum:        tonsSum,
	}
}

func renderMaterialTransactionRateSummary(cmd *cobra.Command, summary materialTransactionRateSummary, sparse bool) error {
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "Material Transaction Rate Summary")
	if summary.MaterialSiteID != "" {
		fmt.Fprintf(out, "Material Site: %s\n", summary.MaterialSiteID)
	}
	if summary.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone: %s\n", summary.TimeZoneID)
	}
	if summary.GroupBy != "" {
		fmt.Fprintf(out, "Group By: %s\n", summary.GroupBy)
	}
	if summary.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", summary.StartAt)
	}
	if summary.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", summary.EndAt)
	}
	if strings.TrimSpace(summary.MaterialTypeHierarchies) != "" {
		fmt.Fprintf(out, "Material Type Hierarchies: %s\n", summary.MaterialTypeHierarchies)
	}
	if sparse {
		fmt.Fprintln(out, "Results: sparse (non-zero tons only)")
	}

	if len(summary.Results) == 0 {
		fmt.Fprintln(out, "No material transaction rate summary data found.")
		return nil
	}

	fmt.Fprintln(out)
	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "DATE\tHOUR\tTONS")
	for _, row := range summary.Results {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", row.Date, formatHour(row.Hour), formatTonsSum(row.TonsSum))
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	if len(summary.DescriptiveStatistics) == 0 {
		return nil
	}

	fmt.Fprintln(out)
	if sparse {
		fmt.Fprintln(out, "Sparse descriptive statistics")
	} else {
		fmt.Fprintln(out, "Descriptive statistics")
	}
	statsWriter := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(statsWriter, "METRIC\tVALUE")
	keys := make([]string, 0, len(summary.DescriptiveStatistics))
	for key := range summary.DescriptiveStatistics {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(statsWriter, "%s\t%s\n", key, formatStatsValue(summary.DescriptiveStatistics[key]))
	}
	return statsWriter.Flush()
}

func mapStringAnyAttr(attrs map[string]any, key string) map[string]any {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return nil
}

func rowValue(row map[string]any, keys ...string) any {
	if row == nil {
		return nil
	}
	for _, key := range keys {
		if value, ok := row[key]; ok {
			return value
		}
	}
	return nil
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case int32:
		return int(typed)
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	case json.Number:
		if i, err := typed.Int64(); err == nil {
			return int(i)
		}
		if f, err := typed.Float64(); err == nil {
			return int(f)
		}
	case string:
		if i, err := strconv.Atoi(typed); err == nil {
			return i
		}
	}
	return 0
}

func floatFromAny(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case int32:
		return float64(typed)
	case json.Number:
		if f, err := typed.Float64(); err == nil {
			return f
		}
	case string:
		if f, err := strconv.ParseFloat(typed, 64); err == nil {
			return f
		}
	}
	return 0
}

func formatHour(hour int) string {
	if hour < 0 {
		return ""
	}
	return strconv.Itoa(hour)
}

func formatTonsSum(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func formatStatsValue(value any) string {
	switch typed := value.(type) {
	case float64:
		return formatStatsFloat(typed)
	case float32:
		return formatStatsFloat(float64(typed))
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case json.Number:
		if i, err := typed.Int64(); err == nil {
			return strconv.FormatInt(i, 10)
		}
		if f, err := typed.Float64(); err == nil {
			return formatStatsFloat(f)
		}
		return typed.String()
	case string:
		return strings.TrimSpace(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func formatStatsFloat(value float64) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return fmt.Sprintf("%.4f", value)
}
