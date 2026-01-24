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

type doPaveFrameActualHoursUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	Date        string
	Hour        string
	Window      string
	Latitude    string
	Longitude   string
	TempMinF    string
	Precip1hrIn string
}

func newDoPaveFrameActualHoursUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a pave frame actual hour",
		Long: `Update a pave frame actual hour.

Provide the record ID as an argument, then specify which fields to update.
Only specified fields will be modified.

Updatable fields:
  --date           Date (YYYY-MM-DD)
  --hour           Hour of day (0-23)
  --window         Window (day/night)
  --latitude       Latitude
  --longitude      Longitude
  --temp-min-f     Minimum temperature (F)
  --precip-1hr-in  Precipitation in last hour (inches)

Note: Only admin users can update pave frame actual hours.`,
		Example: `  # Update temperature and precipitation
  xbe do pave-frame-actual-hours update 123 --temp-min-f 45.0 --precip-1hr-in 0.1

  # Update date and window
  xbe do pave-frame-actual-hours update 123 --date 2024-01-16 --window night`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPaveFrameActualHoursUpdate,
	}
	initDoPaveFrameActualHoursUpdateFlags(cmd)
	return cmd
}

func init() {
	doPaveFrameActualHoursCmd.AddCommand(newDoPaveFrameActualHoursUpdateCmd())
}

func initDoPaveFrameActualHoursUpdateFlags(cmd *cobra.Command) {
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
}

func runDoPaveFrameActualHoursUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPaveFrameActualHoursUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("date") {
		attributes["date"] = opts.Date
	}
	if cmd.Flags().Changed("hour") {
		attributes["hour"] = opts.Hour
	}
	if cmd.Flags().Changed("window") {
		attributes["window"] = opts.Window
	}
	if cmd.Flags().Changed("latitude") {
		attributes["latitude"] = opts.Latitude
	}
	if cmd.Flags().Changed("longitude") {
		attributes["longitude"] = opts.Longitude
	}
	if cmd.Flags().Changed("temp-min-f") {
		attributes["temp-min-f"] = opts.TempMinF
	}
	if cmd.Flags().Changed("precip-1hr-in") {
		attributes["precip-1hr-in"] = opts.Precip1hrIn
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "pave-frame-actual-hours",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/pave-frame-actual-hours/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated pave frame actual hour %s\n", row.ID)
	return nil
}

func parseDoPaveFrameActualHoursUpdateOptions(cmd *cobra.Command, args []string) (doPaveFrameActualHoursUpdateOptions, error) {
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

	return doPaveFrameActualHoursUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		Date:        date,
		Hour:        hour,
		Window:      window,
		Latitude:    latitude,
		Longitude:   longitude,
		TempMinF:    tempMinF,
		Precip1hrIn: precip1hrIn,
	}, nil
}
