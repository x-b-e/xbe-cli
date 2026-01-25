package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doPaveFrameActualStatisticsUpdateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ID                           string
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

func newDoPaveFrameActualStatisticsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a pave frame actual statistic",
		Long: `Update an existing pave frame actual statistic.

Optional flags:
  --latitude                       Latitude (-90 to 90)
  --longitude                      Longitude (-180 to 180)
  --hour-minimum-temp-f            Minimum hourly temperature (F)
  --hour-maximum-precip-in         Maximum hourly precipitation (in)
  --window-minimum-paving-hour-pct Minimum paving hour percent (0-1)
  --window                         Window (day/night)
  --agg-level                      Aggregation level (day_of_year, week, month, year)
  --calculate-results-before-create Calculate results before create (true/false)
  --date-min                       Minimum date (YYYY-MM-DD)
  --date-max                       Maximum date (YYYY-MM-DD)
  --work-days                      Work days (comma-separated, 0=Sunday ... 6=Saturday)`,
		Example: `  # Update window and aggregation level
  xbe do pave-frame-actual-statistics update 123 --window night --agg-level week

  # Update thresholds
  xbe do pave-frame-actual-statistics update 123 --hour-minimum-temp-f 50 --hour-maximum-precip-in 0.05`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPaveFrameActualStatisticsUpdate,
	}
	initDoPaveFrameActualStatisticsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPaveFrameActualStatisticsCmd.AddCommand(newDoPaveFrameActualStatisticsUpdateCmd())
}

func initDoPaveFrameActualStatisticsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("latitude", "", "Latitude (-90 to 90)")
	cmd.Flags().String("longitude", "", "Longitude (-180 to 180)")
	cmd.Flags().String("hour-minimum-temp-f", "", "Minimum hourly temperature (F)")
	cmd.Flags().String("hour-maximum-precip-in", "", "Maximum hourly precipitation (in)")
	cmd.Flags().String("window-minimum-paving-hour-pct", "", "Minimum paving hour percent (0-1)")
	cmd.Flags().String("window", "", "Window (day/night)")
	cmd.Flags().String("agg-level", "", "Aggregation level (day_of_year, week, month, year)")
	cmd.Flags().Bool("calculate-results-before-create", false, "Calculate results before create")
	cmd.Flags().String("date-min", "", "Minimum date (YYYY-MM-DD)")
	cmd.Flags().String("date-max", "", "Maximum date (YYYY-MM-DD)")
	cmd.Flags().StringSlice("work-days", nil, "Work days (comma-separated, 0=Sunday ... 6=Saturday)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPaveFrameActualStatisticsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPaveFrameActualStatisticsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}

	if cmd.Flags().Changed("latitude") {
		attributes["latitude"] = opts.Latitude
	}
	if cmd.Flags().Changed("longitude") {
		attributes["longitude"] = opts.Longitude
	}
	if cmd.Flags().Changed("hour-minimum-temp-f") {
		attributes["hour-minimum-temp-f"] = opts.HourMinimumTempF
	}
	if cmd.Flags().Changed("hour-maximum-precip-in") {
		attributes["hour-maximum-precip-in"] = opts.HourMaximumPrecipIn
	}
	if cmd.Flags().Changed("window-minimum-paving-hour-pct") {
		attributes["window-minimum-paving-hour-pct"] = opts.WindowMinimumPavingHourPct
	}
	if cmd.Flags().Changed("window") {
		attributes["window"] = opts.Window
	}
	if cmd.Flags().Changed("agg-level") {
		attributes["agg-level"] = opts.AggLevel
	}
	if cmd.Flags().Changed("calculate-results-before-create") {
		attributes["calculate-results-before-create"] = opts.CalculateResultsBeforeCreate
	}
	if cmd.Flags().Changed("date-min") {
		attributes["date-min"] = opts.DateMin
	}
	if cmd.Flags().Changed("date-max") {
		attributes["date-max"] = opts.DateMax
	}
	if cmd.Flags().Changed("work-days") {
		workDays, err := parseWorkDays(opts.WorkDays)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if len(workDays) == 0 {
			attributes["work-days"] = []int{}
		} else {
			attributes["work-days"] = workDays
		}
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "pave-frame-actual-statistics",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	path := fmt.Sprintf("/v1/pave-frame-actual-statistics/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated pave frame actual statistic %s\n", resp.Data.ID)
	return nil
}

func parseDoPaveFrameActualStatisticsUpdateOptions(cmd *cobra.Command, args []string) (doPaveFrameActualStatisticsUpdateOptions, error) {
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

	return doPaveFrameActualStatisticsUpdateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ID:                           args[0],
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
