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

type tractorFuelConsumptionReadingsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tractorFuelConsumptionReadingDetails struct {
	ID                        string  `json:"id"`
	ReadingOn                 string  `json:"reading_on,omitempty"`
	ReadingTime               string  `json:"reading_time,omitempty"`
	DateSequence              int     `json:"date_sequence,omitempty"`
	StateCode                 string  `json:"state_code,omitempty"`
	Value                     float64 `json:"value,omitempty"`
	TractorID                 string  `json:"tractor_id,omitempty"`
	TractorNumber             string  `json:"tractor,omitempty"`
	DriverDayID               string  `json:"driver_day_id,omitempty"`
	DriverDayStartOn          string  `json:"driver_day_start_on,omitempty"`
	UnitOfMeasureID           string  `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasureName         string  `json:"unit_of_measure,omitempty"`
	UnitOfMeasureAbbreviation string  `json:"unit_of_measure_abbreviation,omitempty"`
}

func newTractorFuelConsumptionReadingsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tractor fuel consumption reading details",
		Long: `Show the full details of a tractor fuel consumption reading.

Output Fields:
  ID
  Reading On
  Reading Time
  Date Sequence
  Value
  State Code
  Tractor
  Driver Day
  Unit Of Measure

Arguments:
  <id>    The tractor fuel consumption reading ID (required). You can find IDs using the list command.`,
		Example: `  # Show a tractor fuel consumption reading
  xbe view tractor-fuel-consumption-readings show 123

  # Get JSON output
  xbe view tractor-fuel-consumption-readings show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTractorFuelConsumptionReadingsShow,
	}
	initTractorFuelConsumptionReadingsShowFlags(cmd)
	return cmd
}

func init() {
	tractorFuelConsumptionReadingsCmd.AddCommand(newTractorFuelConsumptionReadingsShowCmd())
}

func initTractorFuelConsumptionReadingsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTractorFuelConsumptionReadingsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTractorFuelConsumptionReadingsShowOptions(cmd)
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
		return fmt.Errorf("tractor fuel consumption reading id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tractor-fuel-consumption-readings]", "reading-on,reading-time,date-sequence,state-code,value,tractor,driver-day,unit-of-measure")
	query.Set("fields[tractors]", "number")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[trucker-shift-sets]", "start-on")
	query.Set("include", "tractor,driver-day,unit-of-measure")

	body, _, err := client.Get(cmd.Context(), "/v1/tractor-fuel-consumption-readings/"+id, query)
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

	details := buildTractorFuelConsumptionReadingDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTractorFuelConsumptionReadingDetails(cmd, details)
}

func parseTractorFuelConsumptionReadingsShowOptions(cmd *cobra.Command) (tractorFuelConsumptionReadingsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tractorFuelConsumptionReadingsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTractorFuelConsumptionReadingDetails(resp jsonAPISingleResponse) tractorFuelConsumptionReadingDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := tractorFuelConsumptionReadingDetails{
		ID:           resource.ID,
		ReadingOn:    formatDate(stringAttr(attrs, "reading-on")),
		ReadingTime:  formatTime(stringAttr(attrs, "reading-time")),
		DateSequence: intAttr(attrs, "date-sequence"),
		StateCode:    stringAttr(attrs, "state-code"),
		Value:        floatAttr(attrs, "value"),
	}

	details.TractorID = relationshipIDFromMap(resource.Relationships, "tractor")
	details.DriverDayID = relationshipIDFromMap(resource.Relationships, "driver-day")
	details.UnitOfMeasureID = relationshipIDFromMap(resource.Relationships, "unit-of-measure")

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.TractorID != "" {
		if tractor, ok := included[resourceKey("tractors", details.TractorID)]; ok {
			details.TractorNumber = stringAttr(tractor.Attributes, "number")
		}
	}

	if details.DriverDayID != "" {
		if driverDay, ok := included[resourceKey("trucker-shift-sets", details.DriverDayID)]; ok {
			details.DriverDayStartOn = formatDate(stringAttr(driverDay.Attributes, "start-on"))
		}
	}

	if details.UnitOfMeasureID != "" {
		if uom, ok := included[resourceKey("unit-of-measures", details.UnitOfMeasureID)]; ok {
			details.UnitOfMeasureName = stringAttr(uom.Attributes, "name")
			details.UnitOfMeasureAbbreviation = stringAttr(uom.Attributes, "abbreviation")
		}
	}

	return details
}

func renderTractorFuelConsumptionReadingDetails(cmd *cobra.Command, details tractorFuelConsumptionReadingDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ReadingOn != "" {
		fmt.Fprintf(out, "Reading On: %s\n", details.ReadingOn)
	}
	if details.ReadingTime != "" {
		fmt.Fprintf(out, "Reading Time: %s\n", details.ReadingTime)
	}
	if details.DateSequence != 0 {
		fmt.Fprintf(out, "Date Sequence: %d\n", details.DateSequence)
	}
	fmt.Fprintf(out, "Value: %.2f\n", details.Value)
	if details.StateCode != "" {
		fmt.Fprintf(out, "State Code: %s\n", details.StateCode)
	}

	if details.TractorID != "" {
		label := details.TractorID
		if details.TractorNumber != "" {
			label = fmt.Sprintf("%s (%s)", details.TractorNumber, details.TractorID)
		}
		fmt.Fprintf(out, "Tractor: %s\n", label)
	}

	if details.DriverDayID != "" {
		label := details.DriverDayID
		if details.DriverDayStartOn != "" {
			label = fmt.Sprintf("%s (%s)", details.DriverDayStartOn, details.DriverDayID)
		}
		fmt.Fprintf(out, "Driver Day: %s\n", label)
	}

	if details.UnitOfMeasureID != "" {
		label := details.UnitOfMeasureID
		name := firstNonEmpty(details.UnitOfMeasureAbbreviation, details.UnitOfMeasureName)
		if name != "" {
			label = fmt.Sprintf("%s (%s)", name, details.UnitOfMeasureID)
		}
		fmt.Fprintf(out, "Unit Of Measure: %s\n", label)
	}

	return nil
}
