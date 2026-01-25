package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type paveFrameActualStatisticsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type paveFrameActualStatisticDetails struct {
	ID                           string   `json:"id"`
	Name                         string   `json:"name,omitempty"`
	Latitude                     string   `json:"latitude,omitempty"`
	Longitude                    string   `json:"longitude,omitempty"`
	HourMinimumTempF             string   `json:"hour_minimum_temp_f,omitempty"`
	HourMaximumPrecipIn          string   `json:"hour_maximum_precip_in,omitempty"`
	WindowMinimumPavingHourPct   string   `json:"window_minimum_paving_hour_pct,omitempty"`
	Window                       string   `json:"window,omitempty"`
	AggLevel                     string   `json:"agg_level,omitempty"`
	CalculateResultsBeforeCreate bool     `json:"calculate_results_before_create,omitempty"`
	DateMin                      string   `json:"date_min,omitempty"`
	DateMax                      string   `json:"date_max,omitempty"`
	WorkDays                     []string `json:"work_days,omitempty"`
	Results                      any      `json:"results,omitempty"`
	ResultsAt                    string   `json:"results_at,omitempty"`
}

func newPaveFrameActualStatisticsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show pave frame actual statistic details",
		Long: `Show the full details of a pave frame actual statistic.

Includes location, aggregation settings, thresholds, and computed results.

Arguments:
  <id>  Statistic ID (required). Find IDs using the list command.`,
		Example: `  # Show a statistic
  xbe view pave-frame-actual-statistics show 123

  # Output JSON
  xbe view pave-frame-actual-statistics show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPaveFrameActualStatisticsShow,
	}
	initPaveFrameActualStatisticsShowFlags(cmd)
	return cmd
}

func init() {
	paveFrameActualStatisticsCmd.AddCommand(newPaveFrameActualStatisticsShowCmd())
}

func initPaveFrameActualStatisticsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPaveFrameActualStatisticsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePaveFrameActualStatisticsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("pave frame actual statistic id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[pave-frame-actual-statistics]", "name,latitude,longitude,hour-minimum-temp-f,hour-maximum-precip-in,window-minimum-paving-hour-pct,window,agg-level,calculate-results-before-create,date-min,date-max,work-days,results,results-at")

	body, _, err := client.Get(cmd.Context(), "/v1/pave-frame-actual-statistics/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildPaveFrameActualStatisticDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPaveFrameActualStatisticDetails(cmd, details)
}

func parsePaveFrameActualStatisticsShowOptions(cmd *cobra.Command) (paveFrameActualStatisticsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return paveFrameActualStatisticsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPaveFrameActualStatisticDetails(resp jsonAPISingleResponse) paveFrameActualStatisticDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return paveFrameActualStatisticDetails{
		ID:                           resource.ID,
		Name:                         stringAttr(attrs, "name"),
		Latitude:                     stringAttr(attrs, "latitude"),
		Longitude:                    stringAttr(attrs, "longitude"),
		HourMinimumTempF:             stringAttr(attrs, "hour-minimum-temp-f"),
		HourMaximumPrecipIn:          stringAttr(attrs, "hour-maximum-precip-in"),
		WindowMinimumPavingHourPct:   stringAttr(attrs, "window-minimum-paving-hour-pct"),
		Window:                       stringAttr(attrs, "window"),
		AggLevel:                     stringAttr(attrs, "agg-level"),
		CalculateResultsBeforeCreate: boolAttr(attrs, "calculate-results-before-create"),
		DateMin:                      stringAttr(attrs, "date-min"),
		DateMax:                      stringAttr(attrs, "date-max"),
		WorkDays:                     stringSliceAttr(attrs, "work-days"),
		Results:                      anyAttr(attrs, "results"),
		ResultsAt:                    stringAttr(attrs, "results-at"),
	}
}

func renderPaveFrameActualStatisticDetails(cmd *cobra.Command, details paveFrameActualStatisticDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Latitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", details.Latitude)
	}
	if details.Longitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", details.Longitude)
	}
	if details.Window != "" {
		fmt.Fprintf(out, "Window: %s\n", details.Window)
	}
	if details.AggLevel != "" {
		fmt.Fprintf(out, "Aggregation Level: %s\n", details.AggLevel)
	}
	if details.DateMin != "" {
		fmt.Fprintf(out, "Date Min: %s\n", details.DateMin)
	}
	if details.DateMax != "" {
		fmt.Fprintf(out, "Date Max: %s\n", details.DateMax)
	}
	if details.HourMinimumTempF != "" {
		fmt.Fprintf(out, "Hour Minimum Temp (F): %s\n", details.HourMinimumTempF)
	}
	if details.HourMaximumPrecipIn != "" {
		fmt.Fprintf(out, "Hour Maximum Precip (in): %s\n", details.HourMaximumPrecipIn)
	}
	if details.WindowMinimumPavingHourPct != "" {
		fmt.Fprintf(out, "Window Minimum Paving Hour Pct: %s\n", details.WindowMinimumPavingHourPct)
	}
	if len(details.WorkDays) > 0 {
		fmt.Fprintf(out, "Work Days: %s\n", strings.Join(details.WorkDays, ", "))
	}
	fmt.Fprintf(out, "Calculate Results Before Create: %t\n", details.CalculateResultsBeforeCreate)
	if details.ResultsAt != "" {
		fmt.Fprintf(out, "Results At: %s\n", details.ResultsAt)
	}

	if details.Results != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Results:")
		if formatted := formatAnyJSON(details.Results); formatted != "" {
			fmt.Fprintln(out, formatted)
		}
	}

	return nil
}
