package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doCycleTimeComparisonsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	CoordinatesOne   string
	CoordinatesTwo   string
	ProximityMeters  string
	TransactionAtMin string
	TransactionAtMax string
	CycleCount       string
	CycleCountSet    bool
}

type cycleTimeComparisonRow struct {
	ID                            string             `json:"id"`
	CoordinatesOne                []float64          `json:"coordinates_one,omitempty"`
	CoordinatesTwo                []float64          `json:"coordinates_two,omitempty"`
	ProximityMeters               *float64           `json:"proximity_meters,omitempty"`
	TransactionAtMin              string             `json:"transaction_at_min,omitempty"`
	TransactionAtMax              string             `json:"transaction_at_max,omitempty"`
	CycleCount                    *int               `json:"cycle_count,omitempty"`
	RoundTripDrivingMinutesMedian *float64           `json:"round_trip_driving_minutes_median,omitempty"`
	Percentiles                   map[string]float64 `json:"percentiles,omitempty"`
}

func newDoCycleTimeComparisonsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a cycle time comparison",
		Long: `Create a cycle time comparison.

Cycle time comparisons estimate cycle durations between two coordinate points
using accepted material transactions within a proximity radius.

Required flags:
  --coordinates-one   JSON array [lat,lon] for the first point
  --coordinates-two   JSON array [lat,lon] for the second point
  --proximity-meters  Radius in meters for nearby job/material sites

Optional flags:
  --transaction-at-min  Minimum transaction timestamp (ISO8601)
  --transaction-at-max  Maximum transaction timestamp (ISO8601)
  --cycle-count         Limit number of cycles sampled (integer)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Compare cycle times between two points
  xbe do cycle-time-comparisons create \
    --coordinates-one '[37.7749,-122.4194]' \
    --coordinates-two '[37.8044,-122.2712]' \
    --proximity-meters 5000

  # Limit to a date range and sample size
  xbe do cycle-time-comparisons create \
    --coordinates-one '[37.7749,-122.4194]' \
    --coordinates-two '[37.8044,-122.2712]' \
    --proximity-meters 5000 \
    --transaction-at-min 2024-01-01T00:00:00Z \
    --transaction-at-max 2024-12-31T23:59:59Z \
    --cycle-count 250

  # JSON output
  xbe do cycle-time-comparisons create \
    --coordinates-one '[37.7749,-122.4194]' \
    --coordinates-two '[37.8044,-122.2712]' \
    --proximity-meters 5000 --json`,
		Args: cobra.NoArgs,
		RunE: runDoCycleTimeComparisonsCreate,
	}
	initDoCycleTimeComparisonsCreateFlags(cmd)
	return cmd
}

func init() {
	doCycleTimeComparisonsCmd.AddCommand(newDoCycleTimeComparisonsCreateCmd())
}

func initDoCycleTimeComparisonsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("coordinates-one", "", "JSON array [lat,lon] for the first point (required)")
	cmd.Flags().String("coordinates-two", "", "JSON array [lat,lon] for the second point (required)")
	cmd.Flags().String("proximity-meters", "", "Radius in meters for nearby sites (required)")
	cmd.Flags().String("transaction-at-min", "", "Minimum transaction timestamp (ISO8601)")
	cmd.Flags().String("transaction-at-max", "", "Maximum transaction timestamp (ISO8601)")
	cmd.Flags().String("cycle-count", "", "Limit number of cycles sampled (integer)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("coordinates-one")
	_ = cmd.MarkFlagRequired("coordinates-two")
	_ = cmd.MarkFlagRequired("proximity-meters")
}

func runDoCycleTimeComparisonsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoCycleTimeComparisonsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	coordinatesOne, err := parseCoordinatePair(opts.CoordinatesOne)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	coordinatesTwo, err := parseCoordinatePair(opts.CoordinatesTwo)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	proximityMeters, err := parseProximityMeters(opts.ProximityMeters)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"coordinates-one":  coordinatesOne,
		"coordinates-two":  coordinatesTwo,
		"proximity-meters": proximityMeters,
	}

	if opts.TransactionAtMin != "" {
		attributes["transaction-at-min"] = opts.TransactionAtMin
	}
	if opts.TransactionAtMax != "" {
		attributes["transaction-at-max"] = opts.TransactionAtMax
	}
	if opts.CycleCountSet {
		if strings.TrimSpace(opts.CycleCount) == "" {
			err := fmt.Errorf("--cycle-count must be a positive integer")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		cycleCount, err := strconv.Atoi(opts.CycleCount)
		if err != nil || cycleCount <= 0 {
			err := fmt.Errorf("--cycle-count must be a positive integer")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["cycle-count"] = cycleCount
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "cycle-time-comparisons",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/cycle-time-comparisons", jsonBody)
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

	row := buildCycleTimeComparisonRowFromSingle(resp)
	if len(row.CoordinatesOne) == 0 {
		row.CoordinatesOne = coordinatesOne
	}
	if len(row.CoordinatesTwo) == 0 {
		row.CoordinatesTwo = coordinatesTwo
	}
	if row.ProximityMeters == nil {
		row.ProximityMeters = &proximityMeters
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderCycleTimeComparisonDetails(cmd, row)
}

func parseDoCycleTimeComparisonsCreateOptions(cmd *cobra.Command) (doCycleTimeComparisonsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	coordinatesOne, _ := cmd.Flags().GetString("coordinates-one")
	coordinatesTwo, _ := cmd.Flags().GetString("coordinates-two")
	proximityMeters, _ := cmd.Flags().GetString("proximity-meters")
	transactionAtMin, _ := cmd.Flags().GetString("transaction-at-min")
	transactionAtMax, _ := cmd.Flags().GetString("transaction-at-max")
	cycleCount, _ := cmd.Flags().GetString("cycle-count")
	cycleCountSet := cmd.Flags().Changed("cycle-count")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doCycleTimeComparisonsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		CoordinatesOne:   coordinatesOne,
		CoordinatesTwo:   coordinatesTwo,
		ProximityMeters:  proximityMeters,
		TransactionAtMin: transactionAtMin,
		TransactionAtMax: transactionAtMax,
		CycleCount:       cycleCount,
		CycleCountSet:    cycleCountSet,
	}, nil
}

func parseCoordinatePair(raw string) ([]float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("coordinates must be provided as [lat,lon]")
	}
	var coords []float64
	if err := json.Unmarshal([]byte(raw), &coords); err != nil {
		return nil, fmt.Errorf("invalid coordinates JSON (expected [lat,lon]): %w", err)
	}
	if len(coords) != 2 {
		return nil, fmt.Errorf("coordinates must be [lat,lon]")
	}
	if coords[0] < -90 || coords[0] > 90 {
		return nil, fmt.Errorf("latitude must be between -90 and 90")
	}
	if coords[1] < -180 || coords[1] > 180 {
		return nil, fmt.Errorf("longitude must be between -180 and 180")
	}
	return coords, nil
}

func parseProximityMeters(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("--proximity-meters is required")
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid --proximity-meters: %w", err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("--proximity-meters must be greater than 0")
	}
	return value, nil
}

func buildCycleTimeComparisonRowFromSingle(resp jsonAPISingleResponse) cycleTimeComparisonRow {
	resource := resp.Data
	return buildCycleTimeComparisonRow(resource)
}

func buildCycleTimeComparisonRow(resource jsonAPIResource) cycleTimeComparisonRow {
	attrs := resource.Attributes
	return cycleTimeComparisonRow{
		ID:                            resource.ID,
		CoordinatesOne:                floatSliceAttr(attrs, "coordinates-one"),
		CoordinatesTwo:                floatSliceAttr(attrs, "coordinates-two"),
		ProximityMeters:               floatAttrPointer(attrs, "proximity-meters"),
		TransactionAtMin:              stringAttr(attrs, "transaction-at-min"),
		TransactionAtMax:              stringAttr(attrs, "transaction-at-max"),
		CycleCount:                    intAttrPointer(attrs, "cycle-count"),
		RoundTripDrivingMinutesMedian: floatAttrPointer(attrs, "round-trip-driving-minutes-median"),
		Percentiles:                   parseCycleTimeComparisonPercentiles(attrs),
	}
}

func parseCycleTimeComparisonPercentiles(attrs map[string]any) map[string]float64 {
	if attrs == nil {
		return nil
	}
	value, ok := attrs["percentiles"]
	if !ok || value == nil {
		return nil
	}

	out := map[string]float64{}
	switch typed := value.(type) {
	case map[string]any:
		for key, raw := range typed {
			if val, ok := parseFloatValue(raw); ok {
				out[key] = val
			}
		}
	case map[string]float64:
		for key, val := range typed {
			out[key] = val
		}
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func floatSliceAttr(attrs map[string]any, key string) []float64 {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case []float64:
		return typed
	case []float32:
		out := make([]float64, 0, len(typed))
		for _, v := range typed {
			out = append(out, float64(v))
		}
		return out
	case []any:
		out := make([]float64, 0, len(typed))
		for _, item := range typed {
			if val, ok := parseFloatValue(item); ok {
				out = append(out, val)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func parseFloatValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		if f, err := typed.Float64(); err == nil {
			return f, true
		}
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func renderCycleTimeComparisonDetails(cmd *cobra.Command, row cycleTimeComparisonRow) error {
	out := cmd.OutOrStdout()

	if row.ID != "" {
		fmt.Fprintf(out, "Cycle time comparison %s\n", row.ID)
	}
	if coords := formatCoordinatePair(row.CoordinatesOne); coords != "" {
		fmt.Fprintf(out, "Coordinates One: %s\n", coords)
	}
	if coords := formatCoordinatePair(row.CoordinatesTwo); coords != "" {
		fmt.Fprintf(out, "Coordinates Two: %s\n", coords)
	}
	if row.ProximityMeters != nil {
		fmt.Fprintf(out, "Proximity: %s\n", formatMeters(row.ProximityMeters))
	}
	if row.TransactionAtMin != "" {
		fmt.Fprintf(out, "Transaction At Min: %s\n", row.TransactionAtMin)
	}
	if row.TransactionAtMax != "" {
		fmt.Fprintf(out, "Transaction At Max: %s\n", row.TransactionAtMax)
	}
	if row.CycleCount != nil {
		fmt.Fprintf(out, "Cycle Count: %d\n", *row.CycleCount)
	}
	if row.RoundTripDrivingMinutesMedian != nil {
		fmt.Fprintf(out, "Round Trip Driving Minutes (Median): %s\n", formatCycleTimeFloatPointer(row.RoundTripDrivingMinutesMedian, 2))
	}

	if len(row.Percentiles) == 0 {
		return nil
	}

	fmt.Fprintln(out, "Cycle Minutes Percentiles:")
	writer := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "PERCENTILE\tMINUTES")
	for _, percentile := range sortedPercentileKeys(row.Percentiles) {
		value := row.Percentiles[percentile]
		fmt.Fprintf(writer, "P%s\t%s\n", percentile, formatCycleTimeFloatValue(value, 2))
	}
	return writer.Flush()
}

func sortedPercentileKeys(percentiles map[string]float64) []string {
	intKeys := make([]int, 0, len(percentiles))
	otherKeys := make([]string, 0)
	for key := range percentiles {
		if parsed, err := strconv.Atoi(key); err == nil {
			intKeys = append(intKeys, parsed)
		} else {
			otherKeys = append(otherKeys, key)
		}
	}
	sort.Ints(intKeys)
	sort.Strings(otherKeys)

	keys := make([]string, 0, len(intKeys)+len(otherKeys))
	for _, key := range intKeys {
		keys = append(keys, strconv.Itoa(key))
	}
	keys = append(keys, otherKeys...)
	return keys
}

func formatCoordinatePair(coords []float64) string {
	if len(coords) != 2 {
		return ""
	}
	return fmt.Sprintf("%.6f, %.6f", coords[0], coords[1])
}

func formatMeters(value *float64) string {
	if value == nil {
		return ""
	}
	if math.Mod(*value, 1) == 0 {
		return fmt.Sprintf("%.0f m", *value)
	}
	return fmt.Sprintf("%.2f m", *value)
}

func formatCycleTimeFloatPointer(value *float64, precision int) string {
	if value == nil {
		return ""
	}
	return formatCycleTimeFloatValue(*value, precision)
}

func formatCycleTimeFloatValue(value float64, precision int) string {
	format := "%." + strconv.Itoa(precision) + "f"
	return fmt.Sprintf(format, value)
}
