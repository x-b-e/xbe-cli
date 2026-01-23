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

type doTractorOdometerReadingsCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	TractorID     string
	DriverDayID   string
	UnitOfMeasure string
	DateSequence  string
	ReadingOn     string
	ReadingTime   string
	StateCode     string
	Value         string
}

func newDoTractorOdometerReadingsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tractor odometer reading",
		Long: `Create a tractor odometer reading.

Required flags:
  --tractor          Tractor ID (required)
  --unit-of-measure  Unit of measure ID (required)
  --value            Reading value (required)
  --state-code       State code (2-letter, required unless derived from tractor)

Optional flags:
  --driver-day    Driver day ID
  --reading-on    Reading date (YYYY-MM-DD)
  --reading-time  Reading time (HH:MM or HH:MM:SS)
  --date-sequence Date sequence (integer)`,
		Example: `  # Create an odometer reading
  xbe do tractor-odometer-readings create --tractor 123 --unit-of-measure 456 --value 120345.6 --state-code IL

  # Create with date and time
  xbe do tractor-odometer-readings create --tractor 123 --unit-of-measure 456 --value 120350 --state-code IL --reading-on 2025-01-15 --reading-time 08:30`,
		Args: cobra.NoArgs,
		RunE: runDoTractorOdometerReadingsCreate,
	}
	initDoTractorOdometerReadingsCreateFlags(cmd)
	return cmd
}

func init() {
	doTractorOdometerReadingsCmd.AddCommand(newDoTractorOdometerReadingsCreateCmd())
}

func initDoTractorOdometerReadingsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tractor", "", "Tractor ID (required)")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID (required)")
	cmd.Flags().String("value", "", "Reading value (required)")
	cmd.Flags().String("state-code", "", "State code (2-letter, required unless derived from tractor)")
	cmd.Flags().String("driver-day", "", "Driver day ID")
	cmd.Flags().String("reading-on", "", "Reading date (YYYY-MM-DD)")
	cmd.Flags().String("reading-time", "", "Reading time (HH:MM or HH:MM:SS)")
	cmd.Flags().String("date-sequence", "", "Date sequence (integer)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorOdometerReadingsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoTractorOdometerReadingsCreateOptions(cmd)
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

	if opts.TractorID == "" {
		err := fmt.Errorf("--tractor is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.UnitOfMeasure == "" {
		err := fmt.Errorf("--unit-of-measure is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.Value == "" {
		err := fmt.Errorf("--value is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.StateCode == "" {
		err := fmt.Errorf("--state-code is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.DateSequence != "" {
		attributes["date-sequence"] = opts.DateSequence
	}
	if opts.ReadingOn != "" {
		attributes["reading-on"] = opts.ReadingOn
	}
	if opts.ReadingTime != "" {
		attributes["reading-time"] = opts.ReadingTime
	}
	if opts.StateCode != "" {
		attributes["state-code"] = opts.StateCode
	}
	if opts.Value != "" {
		attributes["value"] = opts.Value
	}

	relationships := map[string]any{
		"tractor": map[string]any{
			"data": map[string]any{
				"type": "tractors",
				"id":   opts.TractorID,
			},
		},
		"unit-of-measure": map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   opts.UnitOfMeasure,
			},
		},
	}

	if opts.DriverDayID != "" {
		relationships["driver-day"] = map[string]any{
			"data": map[string]any{
				"type": "driver-days",
				"id":   opts.DriverDayID,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tractor-odometer-readings",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/tractor-odometer-readings", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created tractor odometer reading %s\n", row.ID)
	return nil
}

func parseDoTractorOdometerReadingsCreateOptions(cmd *cobra.Command) (doTractorOdometerReadingsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tractorID, _ := cmd.Flags().GetString("tractor")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	value, _ := cmd.Flags().GetString("value")
	stateCode, _ := cmd.Flags().GetString("state-code")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	readingOn, _ := cmd.Flags().GetString("reading-on")
	readingTime, _ := cmd.Flags().GetString("reading-time")
	dateSequence, _ := cmd.Flags().GetString("date-sequence")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorOdometerReadingsCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		TractorID:     tractorID,
		UnitOfMeasure: unitOfMeasure,
		Value:         value,
		StateCode:     stateCode,
		DriverDayID:   driverDay,
		ReadingOn:     readingOn,
		ReadingTime:   readingTime,
		DateSequence:  dateSequence,
	}, nil
}

func buildTractorOdometerReadingRowFromSingle(resp jsonAPISingleResponse) tractorOdometerReadingRow {
	attrs := resp.Data.Attributes
	row := tractorOdometerReadingRow{
		ID:           resp.Data.ID,
		DateSequence: stringAttr(attrs, "date-sequence"),
		ReadingOn:    formatDate(stringAttr(attrs, "reading-on")),
		ReadingTime:  formatTime(stringAttr(attrs, "reading-time")),
		StateCode:    stringAttr(attrs, "state-code"),
		Value:        stringAttr(attrs, "value"),
	}

	if rel, ok := resp.Data.Relationships["tractor"]; ok && rel.Data != nil {
		row.TractorID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["driver-day"]; ok && rel.Data != nil {
		row.DriverDayID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}
