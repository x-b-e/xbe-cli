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

type doTractorOdometerReadingsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	DriverDayID   string
	UnitOfMeasure string
	DateSequence  string
	ReadingOn     string
	ReadingTime   string
	StateCode     string
	Value         string
}

func newDoTractorOdometerReadingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tractor odometer reading",
		Long: `Update a tractor odometer reading.

Optional flags:
  --unit-of-measure  Unit of measure ID
  --driver-day       Driver day ID
  --reading-on       Reading date (YYYY-MM-DD)
  --reading-time     Reading time (HH:MM or HH:MM:SS)
  --date-sequence    Date sequence (integer)
  --state-code       State code (2-letter)
  --value            Reading value`,
		Example: `  # Update reading value
  xbe do tractor-odometer-readings update 123 --value 120400

  # Update reading date and time
  xbe do tractor-odometer-readings update 123 --reading-on 2025-02-01 --reading-time 09:15`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTractorOdometerReadingsUpdate,
	}
	initDoTractorOdometerReadingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTractorOdometerReadingsCmd.AddCommand(newDoTractorOdometerReadingsUpdateCmd())
}

func initDoTractorOdometerReadingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("driver-day", "", "Driver day ID")
	cmd.Flags().String("reading-on", "", "Reading date (YYYY-MM-DD)")
	cmd.Flags().String("reading-time", "", "Reading time (HH:MM or HH:MM:SS)")
	cmd.Flags().String("date-sequence", "", "Date sequence (integer)")
	cmd.Flags().String("state-code", "", "State code (2-letter)")
	cmd.Flags().String("value", "", "Reading value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorOdometerReadingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorOdometerReadingsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("reading-on") {
		attributes["reading-on"] = opts.ReadingOn
	}
	if cmd.Flags().Changed("reading-time") {
		attributes["reading-time"] = opts.ReadingTime
	}
	if cmd.Flags().Changed("date-sequence") {
		attributes["date-sequence"] = opts.DateSequence
	}
	if cmd.Flags().Changed("state-code") {
		attributes["state-code"] = opts.StateCode
	}
	if cmd.Flags().Changed("value") {
		attributes["value"] = opts.Value
	}

	if cmd.Flags().Changed("unit-of-measure") {
		relationships["unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		}
	}
	if cmd.Flags().Changed("driver-day") {
		relationships["driver-day"] = map[string]any{
			"data": map[string]any{
				"type": "driver-days",
				"id":   opts.DriverDayID,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "tractor-odometer-readings",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
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

	body, _, err := client.Patch(cmd.Context(), "/v1/tractor-odometer-readings/"+opts.ID, jsonBody)
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

	row := buildTractorOdometerReadingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tractor odometer reading %s\n", row.ID)
	return nil
}

func parseDoTractorOdometerReadingsUpdateOptions(cmd *cobra.Command, args []string) (doTractorOdometerReadingsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	readingOn, _ := cmd.Flags().GetString("reading-on")
	readingTime, _ := cmd.Flags().GetString("reading-time")
	dateSequence, _ := cmd.Flags().GetString("date-sequence")
	stateCode, _ := cmd.Flags().GetString("state-code")
	value, _ := cmd.Flags().GetString("value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	id := strings.TrimSpace(args[0])
	if id == "" {
		return doTractorOdometerReadingsUpdateOptions{}, fmt.Errorf("tractor odometer reading id is required")
	}

	return doTractorOdometerReadingsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            id,
		UnitOfMeasure: unitOfMeasure,
		DriverDayID:   driverDay,
		ReadingOn:     readingOn,
		ReadingTime:   readingTime,
		DateSequence:  dateSequence,
		StateCode:     stateCode,
		Value:         value,
	}, nil
}
