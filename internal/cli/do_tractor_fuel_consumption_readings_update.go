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

type doTractorFuelConsumptionReadingsUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	ReadingOn     string
	ReadingTime   string
	DateSequence  string
	StateCode     string
	Value         string
	DriverDay     string
	UnitOfMeasure string
}

func newDoTractorFuelConsumptionReadingsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tractor fuel consumption reading",
		Long: `Update a tractor fuel consumption reading.

Writable fields:
  --reading-on      Reading date (YYYY-MM-DD)
  --reading-time    Reading time (HH:MM or HH:MM:SS)
  --date-sequence   Sequence number for the day
  --state-code      State code (US abbreviation)
  --value           Fuel consumption value

Writable relationships:
  --driver-day      Driver day ID
  --unit-of-measure Unit of measure ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update reading value
  xbe do tractor-fuel-consumption-readings update 123 --value 10.5

  # Update reading date and unit of measure
  xbe do tractor-fuel-consumption-readings update 123 --reading-on 2025-01-16 --unit-of-measure 45`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTractorFuelConsumptionReadingsUpdate,
	}
	initDoTractorFuelConsumptionReadingsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTractorFuelConsumptionReadingsCmd.AddCommand(newDoTractorFuelConsumptionReadingsUpdateCmd())
}

func initDoTractorFuelConsumptionReadingsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("reading-on", "", "Reading date (YYYY-MM-DD)")
	cmd.Flags().String("reading-time", "", "Reading time (HH:MM or HH:MM:SS)")
	cmd.Flags().String("date-sequence", "", "Sequence number for the day")
	cmd.Flags().String("state-code", "", "State code (US abbreviation)")
	cmd.Flags().String("value", "", "Fuel consumption value")
	cmd.Flags().String("driver-day", "", "Driver day ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorFuelConsumptionReadingsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorFuelConsumptionReadingsUpdateOptions(cmd, args[0])
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("tractor fuel consumption reading id is required")
	}

	attributes := map[string]any{}
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

	relationships := map[string]any{}
	if cmd.Flags().Changed("driver-day") {
		if strings.TrimSpace(opts.DriverDay) == "" {
			err := fmt.Errorf("--driver-day cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["driver-day"] = map[string]any{
			"data": map[string]any{
				"type": "trucker-shift-sets",
				"id":   opts.DriverDay,
			},
		}
	}
	if cmd.Flags().Changed("unit-of-measure") {
		if strings.TrimSpace(opts.UnitOfMeasure) == "" {
			err := fmt.Errorf("--unit-of-measure cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	payload := map[string]any{
		"type": "tractor-fuel-consumption-readings",
		"id":   id,
	}
	if len(attributes) > 0 {
		payload["attributes"] = attributes
	}
	if len(relationships) > 0 {
		payload["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": payload,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/tractor-fuel-consumption-readings/"+id, jsonBody)
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
		row := tractorFuelConsumptionReadingRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tractor fuel consumption reading %s\n", resp.Data.ID)
	return nil
}

func parseDoTractorFuelConsumptionReadingsUpdateOptions(cmd *cobra.Command, id string) (doTractorFuelConsumptionReadingsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	readingOn, _ := cmd.Flags().GetString("reading-on")
	readingTime, _ := cmd.Flags().GetString("reading-time")
	dateSequence, _ := cmd.Flags().GetString("date-sequence")
	stateCode, _ := cmd.Flags().GetString("state-code")
	value, _ := cmd.Flags().GetString("value")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorFuelConsumptionReadingsUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            id,
		ReadingOn:     readingOn,
		ReadingTime:   readingTime,
		DateSequence:  dateSequence,
		StateCode:     stateCode,
		Value:         value,
		DriverDay:     driverDay,
		UnitOfMeasure: unitOfMeasure,
	}, nil
}
