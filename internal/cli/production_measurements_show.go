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

type productionMeasurementsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type productionMeasurementDetails struct {
	ID                         string `json:"id"`
	JobProductionPlanSegment   string `json:"job_production_plan_segment_id,omitempty"`
	WidthInches                string `json:"width_inches,omitempty"`
	DepthInches                string `json:"depth_inches,omitempty"`
	LengthFeet                 string `json:"length_feet,omitempty"`
	SpeedFeetPerMinute         string `json:"speed_feet_per_minute,omitempty"`
	SpeedFeetPerMinutePossible string `json:"speed_feet_per_minute_possible,omitempty"`
	DensityLbsPerCubicFoot     string `json:"density_lbs_per_cubic_foot,omitempty"`
	WidthDisplayUnitOfMeasure  string `json:"width_display_unit_of_measure,omitempty"`
	PassCount                  string `json:"pass_count,omitempty"`
	Note                       string `json:"note,omitempty"`
	VolumeCubicYards           string `json:"volume_cubic_yards,omitempty"`
	MassTons                   string `json:"mass_tons,omitempty"`
	DurationMinutes            string `json:"duration_minutes,omitempty"`
	RateTonsPerHour            string `json:"rate_tons_per_hour,omitempty"`
	RateCubicYardsPerHour      string `json:"rate_cubic_yards_per_hour,omitempty"`
	CreatedAt                  string `json:"created_at,omitempty"`
	UpdatedAt                  string `json:"updated_at,omitempty"`
}

func newProductionMeasurementsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show production measurement details",
		Long: `Show the full details of a production measurement.

Output Fields:
  ID
  Job Production Plan Segment ID
  Width Inches
  Depth Inches
  Length Feet
  Speed Feet Per Minute
  Speed Feet Per Minute Possible
  Density Lbs Per Cubic Foot
  Width Display Unit Of Measure
  Pass Count
  Note
  Volume Cubic Yards
  Mass Tons
  Duration Minutes
  Rate Tons Per Hour
  Rate Cubic Yards Per Hour
  Created At
  Updated At

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The production measurement ID (required). You can find IDs using the list command.`,
		Example: `  # Show a production measurement
  xbe view production-measurements show 123

  # Get JSON output
  xbe view production-measurements show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProductionMeasurementsShow,
	}
	initProductionMeasurementsShowFlags(cmd)
	return cmd
}

func init() {
	productionMeasurementsCmd.AddCommand(newProductionMeasurementsShowCmd())
}

func initProductionMeasurementsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProductionMeasurementsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProductionMeasurementsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("production measurement id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[production-measurements]", "width-inches,depth-inches,length-feet,speed-feet-per-minute,speed-feet-per-minute-possible,density-lbs-per-cubic-foot,note,width-display-unit-of-measure,pass-count,volume-cubic-yards,mass-tons,duration-minutes,rate-tons-per-hour,rate-cubic-yards-per-hour,job-production-plan-segment,created-at,updated-at")

	body, _, err := client.Get(cmd.Context(), "/v1/production-measurements/"+id, query)
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

	details := buildProductionMeasurementDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProductionMeasurementDetails(cmd, details)
}

func parseProductionMeasurementsShowOptions(cmd *cobra.Command) (productionMeasurementsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return productionMeasurementsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProductionMeasurementDetails(resp jsonAPISingleResponse) productionMeasurementDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := productionMeasurementDetails{
		ID:                         resource.ID,
		WidthInches:                stringAttr(attrs, "width-inches"),
		DepthInches:                stringAttr(attrs, "depth-inches"),
		LengthFeet:                 stringAttr(attrs, "length-feet"),
		SpeedFeetPerMinute:         stringAttr(attrs, "speed-feet-per-minute"),
		SpeedFeetPerMinutePossible: stringAttr(attrs, "speed-feet-per-minute-possible"),
		DensityLbsPerCubicFoot:     stringAttr(attrs, "density-lbs-per-cubic-foot"),
		WidthDisplayUnitOfMeasure:  stringAttr(attrs, "width-display-unit-of-measure"),
		PassCount:                  stringAttr(attrs, "pass-count"),
		Note:                       stringAttr(attrs, "note"),
		VolumeCubicYards:           stringAttr(attrs, "volume-cubic-yards"),
		MassTons:                   stringAttr(attrs, "mass-tons"),
		DurationMinutes:            stringAttr(attrs, "duration-minutes"),
		RateTonsPerHour:            stringAttr(attrs, "rate-tons-per-hour"),
		RateCubicYardsPerHour:      stringAttr(attrs, "rate-cubic-yards-per-hour"),
		CreatedAt:                  formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:                  formatDateTime(stringAttr(attrs, "updated-at")),
	}

	details.JobProductionPlanSegment = relationshipIDFromMap(resource.Relationships, "job-production-plan-segment")

	return details
}

func renderProductionMeasurementDetails(cmd *cobra.Command, details productionMeasurementDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanSegment != "" {
		fmt.Fprintf(out, "Job Production Plan Segment ID: %s\n", details.JobProductionPlanSegment)
	}
	fmt.Fprintf(out, "Width Inches: %s\n", details.WidthInches)
	fmt.Fprintf(out, "Depth Inches: %s\n", details.DepthInches)
	fmt.Fprintf(out, "Length Feet: %s\n", details.LengthFeet)
	fmt.Fprintf(out, "Speed Feet Per Minute: %s\n", details.SpeedFeetPerMinute)
	fmt.Fprintf(out, "Speed Feet Per Minute Possible: %s\n", details.SpeedFeetPerMinutePossible)
	fmt.Fprintf(out, "Density Lbs Per Cubic Foot: %s\n", details.DensityLbsPerCubicFoot)
	fmt.Fprintf(out, "Width Display Unit Of Measure: %s\n", details.WidthDisplayUnitOfMeasure)
	fmt.Fprintf(out, "Pass Count: %s\n", details.PassCount)
	fmt.Fprintf(out, "Note: %s\n", details.Note)
	fmt.Fprintf(out, "Volume Cubic Yards: %s\n", details.VolumeCubicYards)
	fmt.Fprintf(out, "Mass Tons: %s\n", details.MassTons)
	fmt.Fprintf(out, "Duration Minutes: %s\n", details.DurationMinutes)
	fmt.Fprintf(out, "Rate Tons Per Hour: %s\n", details.RateTonsPerHour)
	fmt.Fprintf(out, "Rate Cubic Yards Per Hour: %s\n", details.RateCubicYardsPerHour)
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
