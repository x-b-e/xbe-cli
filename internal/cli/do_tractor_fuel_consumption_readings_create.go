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

type doTractorFuelConsumptionReadingsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	Tractor       string
	UnitOfMeasure string
	DriverDay     string
	ReadingOn     string
	ReadingTime   string
	DateSequence  string
	StateCode     string
	Value         string
}

func newDoTractorFuelConsumptionReadingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tractor fuel consumption reading",
		Long: `Create a tractor fuel consumption reading.

Required flags:
  --tractor         Tractor ID
  --unit-of-measure Unit of measure ID
  --value           Fuel consumption value

Optional flags:
  --driver-day    Driver day ID
  --reading-on    Reading date (YYYY-MM-DD)
  --reading-time  Reading time (HH:MM or HH:MM:SS)
  --date-sequence Sequence number for the day
  --state-code    State code (US abbreviation)`,
		Example: `  # Create a tractor fuel consumption reading
  xbe do tractor-fuel-consumption-readings create \
    --tractor 123 \
    --unit-of-measure 45 \
    --value 12.5 \
    --reading-on 2025-01-15 \
    --reading-time 08:15 \
    --state-code CA`,
		Args: cobra.NoArgs,
		RunE: runDoTractorFuelConsumptionReadingsCreate,
	}
	initDoTractorFuelConsumptionReadingsCreateFlags(cmd)
	return cmd
}

func init() {
	doTractorFuelConsumptionReadingsCmd.AddCommand(newDoTractorFuelConsumptionReadingsCreateCmd())
}

func initDoTractorFuelConsumptionReadingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tractor", "", "Tractor ID (required)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (required)")
	cmd.Flags().String("driver-day", "", "Driver day ID")
	cmd.Flags().String("reading-on", "", "Reading date (YYYY-MM-DD)")
	cmd.Flags().String("reading-time", "", "Reading time (HH:MM or HH:MM:SS)")
	cmd.Flags().String("date-sequence", "", "Sequence number for the day")
	cmd.Flags().String("state-code", "", "State code (US abbreviation)")
	cmd.Flags().String("value", "", "Fuel consumption value (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("tractor")
	cmd.MarkFlagRequired("unit-of-measure")
	cmd.MarkFlagRequired("value")
}

func runDoTractorFuelConsumptionReadingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTractorFuelConsumptionReadingsCreateOptions(cmd)
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

	attributes := map[string]any{
		"value": opts.Value,
	}
	if opts.ReadingOn != "" {
		attributes["reading-on"] = opts.ReadingOn
	}
	if opts.ReadingTime != "" {
		attributes["reading-time"] = opts.ReadingTime
	}
	if opts.DateSequence != "" {
		attributes["date-sequence"] = opts.DateSequence
	}
	if opts.StateCode != "" {
		attributes["state-code"] = opts.StateCode
	}

	relationships := map[string]any{
		"tractor": map[string]any{
			"data": map[string]any{
				"type": "tractors",
				"id":   opts.Tractor,
			},
		},
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		},
	}
	if opts.DriverDay != "" {
		relationships["driver-day"] = map[string]any{
			"data": map[string]any{
				"type": "trucker-shift-sets",
				"id":   opts.DriverDay,
			},
		}
	}

	data := map[string]any{
		"type":          "tractor-fuel-consumption-readings",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/tractor-fuel-consumption-readings", jsonBody)
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

	row := tractorFuelConsumptionReadingRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created tractor fuel consumption reading %s\n", row.ID)
	return nil
}

func parseDoTractorFuelConsumptionReadingsCreateOptions(cmd *cobra.Command) (doTractorFuelConsumptionReadingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tractor, _ := cmd.Flags().GetString("tractor")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	readingOn, _ := cmd.Flags().GetString("reading-on")
	readingTime, _ := cmd.Flags().GetString("reading-time")
	dateSequence, _ := cmd.Flags().GetString("date-sequence")
	stateCode, _ := cmd.Flags().GetString("state-code")
	value, _ := cmd.Flags().GetString("value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorFuelConsumptionReadingsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		Tractor:       tractor,
		UnitOfMeasure: unitOfMeasure,
		DriverDay:     driverDay,
		ReadingOn:     readingOn,
		ReadingTime:   readingTime,
		DateSequence:  dateSequence,
		StateCode:     stateCode,
		Value:         value,
	}, nil
}

func tractorFuelConsumptionReadingRowFromSingle(resp jsonAPISingleResponse) tractorFuelConsumptionReadingRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := tractorFuelConsumptionReadingRow{
		ID:              resource.ID,
		ReadingOn:       formatDate(stringAttr(attrs, "reading-on")),
		ReadingTime:     formatTime(stringAttr(attrs, "reading-time")),
		DateSequence:    intAttr(attrs, "date-sequence"),
		StateCode:       stringAttr(attrs, "state-code"),
		Value:           floatAttr(attrs, "value"),
		TractorID:       relationshipIDFromMap(resource.Relationships, "tractor"),
		DriverDayID:     relationshipIDFromMap(resource.Relationships, "driver-day"),
		UnitOfMeasureID: relationshipIDFromMap(resource.Relationships, "unit-of-measure"),
	}

	return row
}
