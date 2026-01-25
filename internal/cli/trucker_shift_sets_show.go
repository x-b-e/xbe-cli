package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type truckerShiftSetsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type truckerShiftSetDetails struct {
	ID                                  string   `json:"id"`
	StartOn                             string   `json:"start_on,omitempty"`
	EarliestStartAt                     string   `json:"earliest_start_at,omitempty"`
	LatestStartAt                       string   `json:"latest_start_at,omitempty"`
	TimeZoneID                          string   `json:"time_zone_id,omitempty"`
	OrderedShiftIDs                     []string `json:"ordered_shift_ids,omitempty"`
	NewShiftIDs                         []string `json:"new_shift_ids,omitempty"`
	ExplicitMobilizationBeforeMinutes   *int     `json:"explicit_mobilization_before_minutes,omitempty"`
	CalculatedMobilizationBeforeMinutes *int     `json:"calculated_mobilization_before_minutes,omitempty"`
	CalculatedMobilizationAfterMinutes  *int     `json:"calculated_mobilization_after_minutes,omitempty"`
	MobilizationBeforeMinutes           *int     `json:"mobilization_before_minutes,omitempty"`
	MobilizationAfterMinutes            *int     `json:"mobilization_after_minutes,omitempty"`
	ExplicitPreTripMinutes              *int     `json:"explicit_pre_trip_minutes,omitempty"`
	PreTripMinutes                      *int     `json:"pre_trip_minutes,omitempty"`
	ExplicitPostTripMinutes             *int     `json:"explicit_post_trip_minutes,omitempty"`
	PostTripMinutes                     *int     `json:"post_trip_minutes,omitempty"`
	ParkingSiteStartAt                  string   `json:"parking_site_start_at,omitempty"`
	IsCustomerAmountConstraintEnabled   bool     `json:"is_customer_amount_constraint_enabled"`
	IsBrokerAmountConstraintEnabled     bool     `json:"is_broker_amount_constraint_enabled"`
	IsTimeSheetEnabled                  bool     `json:"is_time_sheet_enabled"`
	CanCurrentUserEdit                  bool     `json:"can_current_user_edit"`
	IsManaged                           bool     `json:"is_managed"`
	OdometerStartValue                  *float64 `json:"odometer_start_value,omitempty"`
	OdometerEndValue                    *float64 `json:"odometer_end_value,omitempty"`
	OdometerUnitOfMeasureExplicit       string   `json:"odometer_unit_of_measure_explicit,omitempty"`
	OdometerUnitOfMeasure               string   `json:"odometer_unit_of_measure,omitempty"`
	OdometerDistance                    *float64 `json:"odometer_distance,omitempty"`
	TruckerID                           string   `json:"trucker_id,omitempty"`
	TruckerName                         string   `json:"trucker_name,omitempty"`
	BrokerID                            string   `json:"broker_id,omitempty"`
	BrokerName                          string   `json:"broker_name,omitempty"`
	DriverID                            string   `json:"driver_id,omitempty"`
	DriverName                          string   `json:"driver_name,omitempty"`
	TrailerID                           string   `json:"trailer_id,omitempty"`
	TrailerNumber                       string   `json:"trailer_number,omitempty"`
	TrailerClassificationID             string   `json:"trailer_classification_id,omitempty"`
	TrailerClassificationName           string   `json:"trailer_classification_name,omitempty"`
	TractorID                           string   `json:"tractor_id,omitempty"`
	TractorNumber                       string   `json:"tractor_number,omitempty"`
	ExplicitBrokerAmountConstraintID    string   `json:"explicit_broker_amount_constraint_id,omitempty"`
	DriverDayAdjustmentID               string   `json:"driver_day_adjustment_id,omitempty"`
	TenderJobScheduleShiftIDs           []string `json:"tender_job_schedule_shift_ids,omitempty"`
	TripIDs                             []string `json:"trip_ids,omitempty"`
	InvolvedDriverIDs                   []string `json:"involved_driver_ids,omitempty"`
	TimeSheetID                         string   `json:"time_sheet_id,omitempty"`
	TimeSheetIDs                        []string `json:"time_sheet_ids,omitempty"`
	FuelConsumptionReadingIDs           []string `json:"fuel_consumption_reading_ids,omitempty"`
	OdometerReadingIDs                  []string `json:"odometer_reading_ids,omitempty"`
	CommentIDs                          []string `json:"comment_ids,omitempty"`
}

func newTruckerShiftSetsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show trucker shift set details",
		Long: `Show the full details of a trucker shift set (driver day).

Output Fields:
  ID, dates/times, mobilization and trip minutes, time sheet settings,
  odometer values, and equipment assignments.

Relationships:
  Trucker, broker, driver, trailer, tractor, shifts, trips, time sheets,
  comments, and related readings.

Arguments:
  <id>    The trucker shift set ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a trucker shift set
  xbe view trucker-shift-sets show 123

  # JSON output
  xbe view trucker-shift-sets show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTruckerShiftSetsShow,
	}
	initTruckerShiftSetsShowFlags(cmd)
	return cmd
}

func init() {
	truckerShiftSetsCmd.AddCommand(newTruckerShiftSetsShowCmd())
}

func initTruckerShiftSetsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerShiftSetsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTruckerShiftSetsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("trucker shift set id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[trucker-shift-sets]", "earliest-start-at,latest-start-at,ordered-shift-ids,new-shift-ids,explicit-mobilization-before-minutes,time-zone-id,start-on,calculated-mobilization-before-minutes,calculated-mobilization-after-minutes,mobilization-before-minutes,mobilization-after-minutes,explicit-pre-trip-minutes,pre-trip-minutes,explicit-post-trip-minutes,post-trip-minutes,parking-site-start-at,is-customer-amount-constraint-enabled,is-broker-amount-constraint-enabled,is-time-sheet-enabled,can-current-user-edit,is-managed,odometer-start-value,odometer-end-value,odometer-unit-of-measure-explicit,odometer-unit-of-measure,odometer-distance,trucker,driver,broker,trailer,tractor,trailer-classification,explicit-broker-amount-constraint,driver-day-adjustment,tender-job-schedule-shifts,trips,time-sheets,time-sheet,fuel-consumption-readings,odometer-readings,comments")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("fields[trailers]", "number")
	query.Set("fields[tractors]", "number")
	query.Set("fields[trailer-classifications]", "name")
	query.Set("include", "trucker,broker,driver,tractor,trailer,trailer-classification")

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-shift-sets/"+id, query)
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

	details := buildTruckerShiftSetDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTruckerShiftSetDetails(cmd, details)
}

func parseTruckerShiftSetsShowOptions(cmd *cobra.Command) (truckerShiftSetsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerShiftSetsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTruckerShiftSetDetails(resp jsonAPISingleResponse) truckerShiftSetDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := truckerShiftSetDetails{
		ID:                                  resource.ID,
		StartOn:                             formatDate(stringAttr(attrs, "start-on")),
		EarliestStartAt:                     formatDateTime(stringAttr(attrs, "earliest-start-at")),
		LatestStartAt:                       formatDateTime(stringAttr(attrs, "latest-start-at")),
		TimeZoneID:                          stringAttr(attrs, "time-zone-id"),
		OrderedShiftIDs:                     stringSliceAttr(attrs, "ordered-shift-ids"),
		NewShiftIDs:                         stringSliceAttr(attrs, "new-shift-ids"),
		ExplicitMobilizationBeforeMinutes:   intAttrPointer(attrs, "explicit-mobilization-before-minutes"),
		CalculatedMobilizationBeforeMinutes: intAttrPointer(attrs, "calculated-mobilization-before-minutes"),
		CalculatedMobilizationAfterMinutes:  intAttrPointer(attrs, "calculated-mobilization-after-minutes"),
		MobilizationBeforeMinutes:           intAttrPointer(attrs, "mobilization-before-minutes"),
		MobilizationAfterMinutes:            intAttrPointer(attrs, "mobilization-after-minutes"),
		ExplicitPreTripMinutes:              intAttrPointer(attrs, "explicit-pre-trip-minutes"),
		PreTripMinutes:                      intAttrPointer(attrs, "pre-trip-minutes"),
		ExplicitPostTripMinutes:             intAttrPointer(attrs, "explicit-post-trip-minutes"),
		PostTripMinutes:                     intAttrPointer(attrs, "post-trip-minutes"),
		ParkingSiteStartAt:                  formatDateTime(stringAttr(attrs, "parking-site-start-at")),
		IsCustomerAmountConstraintEnabled:   boolAttr(attrs, "is-customer-amount-constraint-enabled"),
		IsBrokerAmountConstraintEnabled:     boolAttr(attrs, "is-broker-amount-constraint-enabled"),
		IsTimeSheetEnabled:                  boolAttr(attrs, "is-time-sheet-enabled"),
		CanCurrentUserEdit:                  boolAttr(attrs, "can-current-user-edit"),
		IsManaged:                           boolAttr(attrs, "is-managed"),
		OdometerStartValue:                  floatAttrPointer(attrs, "odometer-start-value"),
		OdometerEndValue:                    floatAttrPointer(attrs, "odometer-end-value"),
		OdometerUnitOfMeasureExplicit:       stringAttr(attrs, "odometer-unit-of-measure-explicit"),
		OdometerUnitOfMeasure:               stringAttr(attrs, "odometer-unit-of-measure"),
		OdometerDistance:                    floatAttrPointer(attrs, "odometer-distance"),
	}

	details.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
	details.BrokerID = relationshipIDFromMap(resource.Relationships, "broker")
	details.DriverID = relationshipIDFromMap(resource.Relationships, "driver")
	details.TrailerID = relationshipIDFromMap(resource.Relationships, "trailer")
	details.TrailerClassificationID = relationshipIDFromMap(resource.Relationships, "trailer-classification")
	details.TractorID = relationshipIDFromMap(resource.Relationships, "tractor")
	details.ExplicitBrokerAmountConstraintID = relationshipIDFromMap(resource.Relationships, "explicit-broker-amount-constraint")
	details.DriverDayAdjustmentID = relationshipIDFromMap(resource.Relationships, "driver-day-adjustment")
	details.TimeSheetID = relationshipIDFromMap(resource.Relationships, "time-sheet")
	details.TenderJobScheduleShiftIDs = relationshipIDsFromMap(resource.Relationships, "tender-job-schedule-shifts")
	details.TripIDs = relationshipIDsFromMap(resource.Relationships, "trips")
	details.InvolvedDriverIDs = relationshipIDsFromMap(resource.Relationships, "involved-drivers")
	details.TimeSheetIDs = relationshipIDsFromMap(resource.Relationships, "time-sheets")
	details.FuelConsumptionReadingIDs = relationshipIDsFromMap(resource.Relationships, "fuel-consumption-readings")
	details.OdometerReadingIDs = relationshipIDsFromMap(resource.Relationships, "odometer-readings")
	details.CommentIDs = relationshipIDsFromMap(resource.Relationships, "comments")

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if details.TruckerID != "" {
		if trucker, ok := included[resourceKey("truckers", details.TruckerID)]; ok {
			details.TruckerName = firstNonEmpty(
				stringAttr(trucker.Attributes, "company-name"),
				stringAttr(trucker.Attributes, "name"),
			)
		}
	}

	if details.BrokerID != "" {
		if broker, ok := included[resourceKey("brokers", details.BrokerID)]; ok {
			details.BrokerName = firstNonEmpty(
				stringAttr(broker.Attributes, "company-name"),
				stringAttr(broker.Attributes, "name"),
			)
		}
	}

	if details.DriverID != "" {
		if driver, ok := included[resourceKey("users", details.DriverID)]; ok {
			details.DriverName = stringAttr(driver.Attributes, "name")
		}
	}

	if details.TrailerID != "" {
		if trailer, ok := included[resourceKey("trailers", details.TrailerID)]; ok {
			details.TrailerNumber = stringAttr(trailer.Attributes, "number")
		}
	}

	if details.TrailerClassificationID != "" {
		if classification, ok := included[resourceKey("trailer-classifications", details.TrailerClassificationID)]; ok {
			details.TrailerClassificationName = stringAttr(classification.Attributes, "name")
		}
	}

	if details.TractorID != "" {
		if tractor, ok := included[resourceKey("tractors", details.TractorID)]; ok {
			details.TractorNumber = stringAttr(tractor.Attributes, "number")
		}
	}

	return details
}

func renderTruckerShiftSetDetails(cmd *cobra.Command, details truckerShiftSetDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.StartOn != "" {
		fmt.Fprintf(out, "Start On: %s\n", details.StartOn)
	}
	if details.EarliestStartAt != "" {
		fmt.Fprintf(out, "Earliest Start At: %s\n", details.EarliestStartAt)
	}
	if details.LatestStartAt != "" {
		fmt.Fprintf(out, "Latest Start At: %s\n", details.LatestStartAt)
	}
	if details.TimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone ID: %s\n", details.TimeZoneID)
	}
	if len(details.OrderedShiftIDs) > 0 {
		fmt.Fprintf(out, "Ordered Shift IDs: %s\n", strings.Join(details.OrderedShiftIDs, ", "))
	}
	if len(details.NewShiftIDs) > 0 {
		fmt.Fprintf(out, "New Shift IDs: %s\n", strings.Join(details.NewShiftIDs, ", "))
	}
	if details.ExplicitMobilizationBeforeMinutes != nil {
		fmt.Fprintf(out, "Explicit Mobilization Before Minutes: %d\n", *details.ExplicitMobilizationBeforeMinutes)
	}
	if details.CalculatedMobilizationBeforeMinutes != nil {
		fmt.Fprintf(out, "Calculated Mobilization Before Minutes: %d\n", *details.CalculatedMobilizationBeforeMinutes)
	}
	if details.CalculatedMobilizationAfterMinutes != nil {
		fmt.Fprintf(out, "Calculated Mobilization After Minutes: %d\n", *details.CalculatedMobilizationAfterMinutes)
	}
	if details.MobilizationBeforeMinutes != nil {
		fmt.Fprintf(out, "Mobilization Before Minutes: %d\n", *details.MobilizationBeforeMinutes)
	}
	if details.MobilizationAfterMinutes != nil {
		fmt.Fprintf(out, "Mobilization After Minutes: %d\n", *details.MobilizationAfterMinutes)
	}
	if details.ExplicitPreTripMinutes != nil {
		fmt.Fprintf(out, "Explicit Pre-Trip Minutes: %d\n", *details.ExplicitPreTripMinutes)
	}
	if details.PreTripMinutes != nil {
		fmt.Fprintf(out, "Pre-Trip Minutes: %d\n", *details.PreTripMinutes)
	}
	if details.ExplicitPostTripMinutes != nil {
		fmt.Fprintf(out, "Explicit Post-Trip Minutes: %d\n", *details.ExplicitPostTripMinutes)
	}
	if details.PostTripMinutes != nil {
		fmt.Fprintf(out, "Post-Trip Minutes: %d\n", *details.PostTripMinutes)
	}
	if details.ParkingSiteStartAt != "" {
		fmt.Fprintf(out, "Parking Site Start At: %s\n", details.ParkingSiteStartAt)
	}

	fmt.Fprintf(out, "Is Customer Amount Constraint Enabled: %t\n", details.IsCustomerAmountConstraintEnabled)
	fmt.Fprintf(out, "Is Broker Amount Constraint Enabled: %t\n", details.IsBrokerAmountConstraintEnabled)
	fmt.Fprintf(out, "Is Time Sheet Enabled: %t\n", details.IsTimeSheetEnabled)
	fmt.Fprintf(out, "Can Current User Edit: %t\n", details.CanCurrentUserEdit)
	fmt.Fprintf(out, "Is Managed: %t\n", details.IsManaged)

	if details.OdometerStartValue != nil {
		fmt.Fprintf(out, "Odometer Start Value: %v\n", *details.OdometerStartValue)
	}
	if details.OdometerEndValue != nil {
		fmt.Fprintf(out, "Odometer End Value: %v\n", *details.OdometerEndValue)
	}
	if details.OdometerUnitOfMeasureExplicit != "" {
		fmt.Fprintf(out, "Odometer Unit Of Measure Explicit: %s\n", details.OdometerUnitOfMeasureExplicit)
	}
	if details.OdometerUnitOfMeasure != "" {
		fmt.Fprintf(out, "Odometer Unit Of Measure: %s\n", details.OdometerUnitOfMeasure)
	}
	if details.OdometerDistance != nil {
		fmt.Fprintf(out, "Odometer Distance: %v\n", *details.OdometerDistance)
	}

	if details.TruckerID != "" || details.TruckerName != "" {
		name := details.TruckerName
		if name == "" {
			name = details.TruckerID
			fmt.Fprintf(out, "Trucker: %s\n", name)
		} else if details.TruckerID != "" {
			fmt.Fprintf(out, "Trucker: %s (%s)\n", name, details.TruckerID)
		} else {
			fmt.Fprintf(out, "Trucker: %s\n", name)
		}
	}

	if details.BrokerID != "" || details.BrokerName != "" {
		name := details.BrokerName
		if name == "" {
			name = details.BrokerID
			fmt.Fprintf(out, "Broker: %s\n", name)
		} else if details.BrokerID != "" {
			fmt.Fprintf(out, "Broker: %s (%s)\n", name, details.BrokerID)
		} else {
			fmt.Fprintf(out, "Broker: %s\n", name)
		}
	}

	if details.DriverID != "" || details.DriverName != "" {
		name := details.DriverName
		if name == "" {
			name = details.DriverID
			fmt.Fprintf(out, "Driver: %s\n", name)
		} else if details.DriverID != "" {
			fmt.Fprintf(out, "Driver: %s (%s)\n", name, details.DriverID)
		} else {
			fmt.Fprintf(out, "Driver: %s\n", name)
		}
	}

	if details.TrailerID != "" || details.TrailerNumber != "" {
		label := details.TrailerNumber
		if label == "" {
			label = details.TrailerID
			fmt.Fprintf(out, "Trailer: %s\n", label)
		} else if details.TrailerID != "" {
			fmt.Fprintf(out, "Trailer: %s (%s)\n", label, details.TrailerID)
		} else {
			fmt.Fprintf(out, "Trailer: %s\n", label)
		}
	}

	if details.TrailerClassificationID != "" || details.TrailerClassificationName != "" {
		label := details.TrailerClassificationName
		if label == "" {
			label = details.TrailerClassificationID
			fmt.Fprintf(out, "Trailer Classification: %s\n", label)
		} else if details.TrailerClassificationID != "" {
			fmt.Fprintf(out, "Trailer Classification: %s (%s)\n", label, details.TrailerClassificationID)
		} else {
			fmt.Fprintf(out, "Trailer Classification: %s\n", label)
		}
	}

	if details.TractorID != "" || details.TractorNumber != "" {
		label := details.TractorNumber
		if label == "" {
			label = details.TractorID
			fmt.Fprintf(out, "Tractor: %s\n", label)
		} else if details.TractorID != "" {
			fmt.Fprintf(out, "Tractor: %s (%s)\n", label, details.TractorID)
		} else {
			fmt.Fprintf(out, "Tractor: %s\n", label)
		}
	}

	if details.ExplicitBrokerAmountConstraintID != "" {
		fmt.Fprintf(out, "Explicit Broker Amount Constraint ID: %s\n", details.ExplicitBrokerAmountConstraintID)
	}
	if details.DriverDayAdjustmentID != "" {
		fmt.Fprintf(out, "Driver Day Adjustment ID: %s\n", details.DriverDayAdjustmentID)
	}
	if len(details.TenderJobScheduleShiftIDs) > 0 {
		fmt.Fprintf(out, "Tender Job Schedule Shift IDs: %s\n", strings.Join(details.TenderJobScheduleShiftIDs, ", "))
	}
	if len(details.TripIDs) > 0 {
		fmt.Fprintf(out, "Trip IDs: %s\n", strings.Join(details.TripIDs, ", "))
	}
	if len(details.InvolvedDriverIDs) > 0 {
		fmt.Fprintf(out, "Involved Driver IDs: %s\n", strings.Join(details.InvolvedDriverIDs, ", "))
	}
	if details.TimeSheetID != "" {
		fmt.Fprintf(out, "Time Sheet ID: %s\n", details.TimeSheetID)
	}
	if len(details.TimeSheetIDs) > 0 {
		fmt.Fprintf(out, "Time Sheet IDs: %s\n", strings.Join(details.TimeSheetIDs, ", "))
	}
	if len(details.FuelConsumptionReadingIDs) > 0 {
		fmt.Fprintf(out, "Fuel Consumption Reading IDs: %s\n", strings.Join(details.FuelConsumptionReadingIDs, ", "))
	}
	if len(details.OdometerReadingIDs) > 0 {
		fmt.Fprintf(out, "Odometer Reading IDs: %s\n", strings.Join(details.OdometerReadingIDs, ", "))
	}
	if len(details.CommentIDs) > 0 {
		fmt.Fprintf(out, "Comment IDs: %s\n", strings.Join(details.CommentIDs, ", "))
	}

	return nil
}

func intAttrPointer(attrs map[string]any, key string) *int {
	if attrs == nil {
		return nil
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case int:
		v := typed
		return &v
	case int64:
		v := int(typed)
		return &v
	case float64:
		v := int(typed)
		return &v
	case float32:
		v := int(typed)
		return &v
	case string:
		if parsed, err := strconv.Atoi(typed); err == nil {
			v := parsed
			return &v
		}
		if parsed, err := strconv.ParseFloat(typed, 64); err == nil {
			v := int(parsed)
			return &v
		}
	}
	return nil
}
