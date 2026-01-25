package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPaveFrameActualStatisticsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	Latitude                     string
	Longitude                    string
	HourMinimumTempF             string
	HourMaximumPrecipIn          string
	WindowMinimumPavingHourPct   string
	Window                       string
	AggLevel                     string
	CalculateResultsBeforeCreate bool
	DateMin                      string
	DateMax                      string
	WorkDays                     []string
}

func newDoPaveFrameActualStatisticsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pave frame actual statistic",
		Long: `Create a pave frame actual statistic.

Required flags:
  --latitude                       Latitude (-90 to 90)
  --longitude                      Longitude (-180 to 180)
  --hour-minimum-temp-f            Minimum hourly temperature (F)
  --hour-maximum-precip-in         Maximum hourly precipitation (in)
  --window-minimum-paving-hour-pct Minimum paving hour percent (0-1)
  --agg-level                      Aggregation level (day_of_year, week, month, year)
  --work-days                      Work days (comma-separated, 0=Sunday ... 6=Saturday)

Optional flags:
  --window                         Window (day/night)
  --calculate-results-before-create Calculate results before create (true/false)
  --date-min                       Minimum date (YYYY-MM-DD)
  --date-max                       Maximum date (YYYY-MM-DD)`,
		Example: `  # Create a statistic
  xbe do pave-frame-actual-statistics create --latitude 41.88 --longitude -87.62 \
    --hour-minimum-temp-f 45 --hour-maximum-precip-in 0.1 --window-minimum-paving-hour-pct 0.6 \
    --agg-level month --work-days 1,2,3,4,5

  # Create with date range
  xbe do pave-frame-actual-statistics create --latitude 41.88 --longitude -87.62 \
    --hour-minimum-temp-f 45 --hour-maximum-precip-in 0.1 --window-minimum-paving-hour-pct 0.6 \
    --agg-level week --work-days 1,2,3,4,5 --date-min 2024-01-01 --date-max 2024-12-31`,
		Args: cobra.NoArgs,
		RunE: runDoPaveFrameActualStatisticsCreate,
	}
	initDoPaveFrameActualStatisticsCreateFlags(cmd)
	return cmd
}

func init() {
	doPaveFrameActualStatisticsCmd.AddCommand(newDoPaveFrameActualStatisticsCreateCmd())
}

func initDoPaveFrameActualStatisticsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("latitude", "", "Latitude (-90 to 90) (required)")
	cmd.Flags().String("longitude", "", "Longitude (-180 to 180) (required)")
	cmd.Flags().String("hour-minimum-temp-f", "", "Minimum hourly temperature (F) (required)")
	cmd.Flags().String("hour-maximum-precip-in", "", "Maximum hourly precipitation (in) (required)")
	cmd.Flags().String("window-minimum-paving-hour-pct", "", "Minimum paving hour percent (0-1) (required)")
	cmd.Flags().String("window", "", "Window (day/night)")
	cmd.Flags().String("agg-level", "", "Aggregation level (day_of_year, week, month, year) (required)")
	cmd.Flags().Bool("calculate-results-before-create", false, "Calculate results before create")
	cmd.Flags().String("date-min", "", "Minimum date (YYYY-MM-DD)")
	cmd.Flags().String("date-max", "", "Maximum date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("work-days", nil, "Work days (comma-separated, 0=Sunday ... 6=Saturday) (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("latitude")
	_ = cmd.MarkFlagRequired("longitude")
	_ = cmd.MarkFlagRequired("hour-minimum-temp-f")
	_ = cmd.MarkFlagRequired("hour-maximum-precip-in")
	_ = cmd.MarkFlagRequired("window-minimum-paving-hour-pct")
	_ = cmd.MarkFlagRequired("agg-level")
	_ = cmd.MarkFlagRequired("work-days")
}

func runDoPaveFrameActualStatisticsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPaveFrameActualStatisticsCreateOptions(cmd)
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

	workDays, err := parseWorkDays(opts.WorkDays)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if len(workDays) == 0 {
		err := errors.New("work-days is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"latitude":                       opts.Latitude,
		"longitude":                      opts.Longitude,
		"hour-minimum-temp-f":            opts.HourMinimumTempF,
		"hour-maximum-precip-in":         opts.HourMaximumPrecipIn,
		"window-minimum-paving-hour-pct": opts.WindowMinimumPavingHourPct,
		"agg-level":                      opts.AggLevel,
		"work-days":                      workDays,
	}

	if opts.Window != "" {
		attributes["window"] = opts.Window
	}
	if cmd.Flags().Changed("calculate-results-before-create") {
		attributes["calculate-results-before-create"] = opts.CalculateResultsBeforeCreate
	}
	if opts.DateMin != "" {
		attributes["date-min"] = opts.DateMin
	}
	if opts.DateMax != "" {
		attributes["date-max"] = opts.DateMax
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "pave-frame-actual-statistics",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/pave-frame-actual-statistics", jsonBody)
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

	if opts.JSON {
		row := buildPaveFrameActualStatisticRow(resp.Data)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created pave frame actual statistic %s\n", resp.Data.ID)
	return nil
}

func parseDoPaveFrameActualStatisticsCreateOptions(cmd *cobra.Command) (doPaveFrameActualStatisticsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	hourMinimumTempF, _ := cmd.Flags().GetString("hour-minimum-temp-f")
	hourMaximumPrecipIn, _ := cmd.Flags().GetString("hour-maximum-precip-in")
	windowMinimumPavingHourPct, _ := cmd.Flags().GetString("window-minimum-paving-hour-pct")
	window, _ := cmd.Flags().GetString("window")
	aggLevel, _ := cmd.Flags().GetString("agg-level")
	calculateResultsBeforeCreate, _ := cmd.Flags().GetBool("calculate-results-before-create")
	dateMin, _ := cmd.Flags().GetString("date-min")
	dateMax, _ := cmd.Flags().GetString("date-max")
	workDays, _ := cmd.Flags().GetStringSlice("work-days")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPaveFrameActualStatisticsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		Latitude:                     latitude,
		Longitude:                    longitude,
		HourMinimumTempF:             hourMinimumTempF,
		HourMaximumPrecipIn:          hourMaximumPrecipIn,
		WindowMinimumPavingHourPct:   windowMinimumPavingHourPct,
		Window:                       window,
		AggLevel:                     aggLevel,
		CalculateResultsBeforeCreate: calculateResultsBeforeCreate,
		DateMin:                      dateMin,
		DateMax:                      dateMax,
		WorkDays:                     workDays,
	}, nil
}

func parseWorkDays(values []string) ([]int, error) {
	cleaned := cleanStringSlice(values)
	if len(cleaned) == 0 {
		return nil, nil
	}
	workDays := make([]int, 0, len(cleaned))
	for _, value := range cleaned {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			parsed, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid work-days value %q", part)
			}
			workDays = append(workDays, parsed)
		}
	}
	return workDays, nil
}
