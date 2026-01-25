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

type doPaveFrameActualHoursCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	Date        string
	Hour        string
	Window      string
	Latitude    string
	Longitude   string
	TempMinF    string
	Precip1hrIn string
}

func newDoPaveFrameActualHoursCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pave frame actual hour",
		Long: `Create a pave frame actual hour.

Required:
  --date           Date (YYYY-MM-DD)
  --hour           Hour of day (0-23)
  --window         Window (day/night)
  --latitude       Latitude
  --longitude      Longitude
  --temp-min-f     Minimum temperature (F)
  --precip-1hr-in  Precipitation in last hour (inches)

Note: Only admin users can create pave frame actual hours.`,
		Example: `  # Create a pave frame actual hour
  xbe do pave-frame-actual-hours create --date 2024-01-15 --hour 9 --window day \
    --latitude 38.9 --longitude -77.0 --temp-min-f 42.3 --precip-1hr-in 0.05`,
		RunE: runDoPaveFrameActualHoursCreate,
	}
	initDoPaveFrameActualHoursCreateFlags(cmd)
	return cmd
}

func init() {
	doPaveFrameActualHoursCmd.AddCommand(newDoPaveFrameActualHoursCreateCmd())
}

func initDoPaveFrameActualHoursCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("date", "", "Date (YYYY-MM-DD)")
	cmd.Flags().String("hour", "", "Hour of day (0-23)")
	cmd.Flags().String("window", "", "Window (day/night)")
	cmd.Flags().String("latitude", "", "Latitude")
	cmd.Flags().String("longitude", "", "Longitude")
	cmd.Flags().String("temp-min-f", "", "Minimum temperature (F)")
	cmd.Flags().String("precip-1hr-in", "", "Precipitation in last hour (inches)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("date")
	_ = cmd.MarkFlagRequired("hour")
	_ = cmd.MarkFlagRequired("window")
	_ = cmd.MarkFlagRequired("latitude")
	_ = cmd.MarkFlagRequired("longitude")
	_ = cmd.MarkFlagRequired("temp-min-f")
	_ = cmd.MarkFlagRequired("precip-1hr-in")
}

func runDoPaveFrameActualHoursCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPaveFrameActualHoursCreateOptions(cmd)
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

	attributes := map[string]any{
		"date":          opts.Date,
		"hour":          opts.Hour,
		"window":        opts.Window,
		"latitude":      opts.Latitude,
		"longitude":     opts.Longitude,
		"temp-min-f":    opts.TempMinF,
		"precip-1hr-in": opts.Precip1hrIn,
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "pave-frame-actual-hours",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/pave-frame-actual-hours", jsonBody)
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

	row := buildPaveFrameActualHourRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created pave frame actual hour %s\n", row.ID)
	return nil
}

func parseDoPaveFrameActualHoursCreateOptions(cmd *cobra.Command) (doPaveFrameActualHoursCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	date, _ := cmd.Flags().GetString("date")
	hour, _ := cmd.Flags().GetString("hour")
	window, _ := cmd.Flags().GetString("window")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	tempMinF, _ := cmd.Flags().GetString("temp-min-f")
	precip1hrIn, _ := cmd.Flags().GetString("precip-1hr-in")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPaveFrameActualHoursCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		Date:        date,
		Hour:        hour,
		Window:      window,
		Latitude:    latitude,
		Longitude:   longitude,
		TempMinF:    tempMinF,
		Precip1hrIn: precip1hrIn,
	}, nil
}
