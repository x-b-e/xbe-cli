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

type doProductionMeasurementsUpdateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	ID                         string
	WidthInches                string
	DepthInches                string
	LengthFeet                 string
	SpeedFeetPerMinute         string
	SpeedFeetPerMinutePossible string
	DensityLbsPerCubicFoot     string
	Note                       string
	WidthDisplayUnitOfMeasure  string
	PassCount                  string
}

func newDoProductionMeasurementsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a production measurement",
		Long: `Update a production measurement.

Optional flags:
  --width-inches                   Width in inches
  --depth-inches                   Depth in inches
  --length-feet                    Length in feet
  --speed-feet-per-minute          Speed in feet per minute
  --speed-feet-per-minute-possible Possible speed in feet per minute
  --density-lbs-per-cubic-foot     Density in lbs per cubic foot
  --note                           Note
  --width-display-unit-of-measure  Width display unit of measure (inches, feet)
  --pass-count                     Pass count

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update dimensions
  xbe do production-measurements update 123 \
    --width-inches 120 \
    --depth-inches 6 \
    --length-feet 600

  # Update display unit and pass count
  xbe do production-measurements update 123 --width-display-unit-of-measure feet --pass-count 2`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProductionMeasurementsUpdate,
	}
	initDoProductionMeasurementsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProductionMeasurementsCmd.AddCommand(newDoProductionMeasurementsUpdateCmd())
}

func initDoProductionMeasurementsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("width-inches", "", "Width in inches")
	cmd.Flags().String("depth-inches", "", "Depth in inches")
	cmd.Flags().String("length-feet", "", "Length in feet")
	cmd.Flags().String("speed-feet-per-minute", "", "Speed in feet per minute")
	cmd.Flags().String("speed-feet-per-minute-possible", "", "Possible speed in feet per minute")
	cmd.Flags().String("density-lbs-per-cubic-foot", "", "Density in lbs per cubic foot")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().String("width-display-unit-of-measure", "", "Width display unit of measure (inches, feet)")
	cmd.Flags().String("pass-count", "", "Pass count")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProductionMeasurementsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProductionMeasurementsUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("production measurement id is required")
	}

	attributes := map[string]any{}

	if cmd.Flags().Changed("width-inches") {
		attributes["width-inches"] = opts.WidthInches
	}
	if cmd.Flags().Changed("depth-inches") {
		attributes["depth-inches"] = opts.DepthInches
	}
	if cmd.Flags().Changed("length-feet") {
		attributes["length-feet"] = opts.LengthFeet
	}
	if cmd.Flags().Changed("speed-feet-per-minute") {
		attributes["speed-feet-per-minute"] = opts.SpeedFeetPerMinute
	}
	if cmd.Flags().Changed("speed-feet-per-minute-possible") {
		attributes["speed-feet-per-minute-possible"] = opts.SpeedFeetPerMinutePossible
	}
	if cmd.Flags().Changed("density-lbs-per-cubic-foot") {
		attributes["density-lbs-per-cubic-foot"] = opts.DensityLbsPerCubicFoot
	}
	if cmd.Flags().Changed("note") {
		attributes["note"] = opts.Note
	}
	if cmd.Flags().Changed("width-display-unit-of-measure") {
		attributes["width-display-unit-of-measure"] = opts.WidthDisplayUnitOfMeasure
	}
	if cmd.Flags().Changed("pass-count") {
		attributes["pass-count"] = opts.PassCount
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "production-measurements",
			"id":         id,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/production-measurements/"+id, jsonBody)
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
		row := productionMeasurementRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated production measurement %s\n", resp.Data.ID)
	return nil
}

func parseDoProductionMeasurementsUpdateOptions(cmd *cobra.Command, args []string) (doProductionMeasurementsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	widthInches, _ := cmd.Flags().GetString("width-inches")
	depthInches, _ := cmd.Flags().GetString("depth-inches")
	lengthFeet, _ := cmd.Flags().GetString("length-feet")
	speedFeetPerMinute, _ := cmd.Flags().GetString("speed-feet-per-minute")
	speedFeetPerMinutePossible, _ := cmd.Flags().GetString("speed-feet-per-minute-possible")
	densityLbsPerCubicFoot, _ := cmd.Flags().GetString("density-lbs-per-cubic-foot")
	note, _ := cmd.Flags().GetString("note")
	widthDisplayUnitOfMeasure, _ := cmd.Flags().GetString("width-display-unit-of-measure")
	passCount, _ := cmd.Flags().GetString("pass-count")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProductionMeasurementsUpdateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		ID:                         args[0],
		WidthInches:                widthInches,
		DepthInches:                depthInches,
		LengthFeet:                 lengthFeet,
		SpeedFeetPerMinute:         speedFeetPerMinute,
		SpeedFeetPerMinutePossible: speedFeetPerMinutePossible,
		DensityLbsPerCubicFoot:     densityLbsPerCubicFoot,
		Note:                       note,
		WidthDisplayUnitOfMeasure:  widthDisplayUnitOfMeasure,
		PassCount:                  passCount,
	}, nil
}
