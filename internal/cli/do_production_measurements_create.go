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

type doProductionMeasurementsCreateOptions struct {
	BaseURL                    string
	Token                      string
	JSON                       bool
	JobProductionPlanSegment   string
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

func newDoProductionMeasurementsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a production measurement",
		Long: `Create a production measurement.

Required flags:
  --job-production-plan-segment   Job production plan segment ID
  --width-display-unit-of-measure Width display unit of measure (inches, feet)
  --pass-count                    Pass count

Optional flags:
  --width-inches                  Width in inches
  --depth-inches                  Depth in inches
  --length-feet                   Length in feet
  --speed-feet-per-minute         Speed in feet per minute
  --speed-feet-per-minute-possible Possible speed in feet per minute
  --density-lbs-per-cubic-foot    Density in lbs per cubic foot
  --note                          Note

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a production measurement
  xbe do production-measurements create \
    --job-production-plan-segment 123 \
    --width-display-unit-of-measure inches \
    --pass-count 1 \
    --width-inches 144 \
    --depth-inches 6 \
    --length-feet 500 \
    --speed-feet-per-minute 35 \
    --density-lbs-per-cubic-foot 145`,
		Args: cobra.NoArgs,
		RunE: runDoProductionMeasurementsCreate,
	}
	initDoProductionMeasurementsCreateFlags(cmd)
	return cmd
}

func init() {
	doProductionMeasurementsCmd.AddCommand(newDoProductionMeasurementsCreateCmd())
}

func initDoProductionMeasurementsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan-segment", "", "Job production plan segment ID (required)")
	cmd.Flags().String("width-inches", "", "Width in inches")
	cmd.Flags().String("depth-inches", "", "Depth in inches")
	cmd.Flags().String("length-feet", "", "Length in feet")
	cmd.Flags().String("speed-feet-per-minute", "", "Speed in feet per minute")
	cmd.Flags().String("speed-feet-per-minute-possible", "", "Possible speed in feet per minute")
	cmd.Flags().String("density-lbs-per-cubic-foot", "", "Density in lbs per cubic foot")
	cmd.Flags().String("note", "", "Note")
	cmd.Flags().String("width-display-unit-of-measure", "", "Width display unit of measure (inches, feet, required)")
	cmd.Flags().String("pass-count", "", "Pass count (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("job-production-plan-segment")
	cmd.MarkFlagRequired("width-display-unit-of-measure")
	cmd.MarkFlagRequired("pass-count")
}

func runDoProductionMeasurementsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProductionMeasurementsCreateOptions(cmd)
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
		"width-display-unit-of-measure": opts.WidthDisplayUnitOfMeasure,
		"pass-count":                    opts.PassCount,
	}
	if opts.WidthInches != "" {
		attributes["width-inches"] = opts.WidthInches
	}
	if opts.DepthInches != "" {
		attributes["depth-inches"] = opts.DepthInches
	}
	if opts.LengthFeet != "" {
		attributes["length-feet"] = opts.LengthFeet
	}
	if opts.SpeedFeetPerMinute != "" {
		attributes["speed-feet-per-minute"] = opts.SpeedFeetPerMinute
	}
	if opts.SpeedFeetPerMinutePossible != "" {
		attributes["speed-feet-per-minute-possible"] = opts.SpeedFeetPerMinutePossible
	}
	if opts.DensityLbsPerCubicFoot != "" {
		attributes["density-lbs-per-cubic-foot"] = opts.DensityLbsPerCubicFoot
	}
	if opts.Note != "" {
		attributes["note"] = opts.Note
	}

	relationships := map[string]any{
		"job-production-plan-segment": map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-segments",
				"id":   opts.JobProductionPlanSegment,
			},
		},
	}

	data := map[string]any{
		"type":          "production-measurements",
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

	body, _, err := client.Post(cmd.Context(), "/v1/production-measurements", jsonBody)
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

	row := productionMeasurementRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created production measurement %s\n", row.ID)
	return nil
}

func parseDoProductionMeasurementsCreateOptions(cmd *cobra.Command) (doProductionMeasurementsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanSegment, _ := cmd.Flags().GetString("job-production-plan-segment")
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

	return doProductionMeasurementsCreateOptions{
		BaseURL:                    baseURL,
		Token:                      token,
		JSON:                       jsonOut,
		JobProductionPlanSegment:   jobProductionPlanSegment,
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
