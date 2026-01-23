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

type doTractorsUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	// Tractor ID
	ID string

	// Attributes
	Number                string
	TruckManufacturerName string
	TruckModelName        string
	TruckModelYear        int
	ColorName             string

	// Registration
	PlateNumber           string
	PlateJurisdictionCode string
	VIN                   string
	Apportioned           string

	// Status
	InService     string
	IntransitHeat string

	// Physical specs
	CurbWeightLbs int
	HeightInches  int

	// Address
	ParkingAddress              string
	ParkingAddressPlaceID       string
	ParkingAddressPlusCode      string
	SkipParkingAddressGeocoding string

	// Relationship
	Trucker         string
	BrokeredTractor string
}

func newDoTractorsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tractor",
		Long: `Update a tractor.

Provide the tractor ID as an argument and specify the fields to update.

Updatable fields:
  Basic:
    --number                    Tractor number/identifier

  Truck details:
    --truck-manufacturer-name   Manufacturer name
    --truck-model-name          Model name
    --truck-model-year          Model year (integer)
    --color-name                Color name

  Registration:
    --plate-number              License plate number
    --plate-jurisdiction-code   Plate jurisdiction code
    --vin                       Vehicle identification number
    --apportioned               Apportioned status (true/false)

  Status:
    --in-service                In service status (true/false)
    --intransit-heat            Intransit heat (true/false)

  Physical specs:
    --curb-weight-lbs           Curb weight in pounds
    --height-inches             Height in inches

  Parking address:
    --parking-address           Parking address
    --parking-address-place-id  Google Place ID
    --parking-address-plus-code Plus code
    --skip-parking-address-geocoding Skip geocoding (true/false)

  Relationships:
    --trucker                   Trucker ID
    --brokered-tractor          Brokered tractor ID`,
		Example: `  # Update tractor number
  xbe do tractors update 123 --number "T101"

  # Update truck details
  xbe do tractors update 123 --truck-manufacturer-name "Freightliner" --truck-model-year 2024

  # Update in-service status
  xbe do tractors update 123 --in-service false

  # Update parking address
  xbe do tractors update 123 --parking-address "456 Oak St, Chicago, IL"

  # Get JSON output
  xbe do tractors update 123 --number "T102" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTractorsUpdate,
	}
	initDoTractorsUpdateFlags(cmd)
	return cmd
}

func init() {
	doTractorsCmd.AddCommand(newDoTractorsUpdateCmd())
}

func initDoTractorsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")

	// Attributes
	cmd.Flags().String("number", "", "Tractor number")
	cmd.Flags().String("truck-manufacturer-name", "", "Manufacturer name")
	cmd.Flags().String("truck-model-name", "", "Model name")
	cmd.Flags().Int("truck-model-year", 0, "Model year")
	cmd.Flags().String("color-name", "", "Color name")

	// Registration
	cmd.Flags().String("plate-number", "", "License plate number")
	cmd.Flags().String("plate-jurisdiction-code", "", "Plate jurisdiction code")
	cmd.Flags().String("vin", "", "Vehicle identification number")
	cmd.Flags().String("apportioned", "", "Apportioned status (true/false)")

	// Status
	cmd.Flags().String("in-service", "", "In service status (true/false)")
	cmd.Flags().String("intransit-heat", "", "Intransit heat (true/false)")

	// Physical specs
	cmd.Flags().Int("curb-weight-lbs", 0, "Curb weight in pounds")
	cmd.Flags().Int("height-inches", 0, "Height in inches")

	// Parking address
	cmd.Flags().String("parking-address", "", "Parking address")
	cmd.Flags().String("parking-address-place-id", "", "Google Place ID")
	cmd.Flags().String("parking-address-plus-code", "", "Plus code")
	cmd.Flags().String("skip-parking-address-geocoding", "", "Skip geocoding (true/false)")

	// Relationships
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("brokered-tractor", "", "Brokered tractor ID")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorsUpdateOptions(cmd, args[0])
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require authentication for write operations
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

	// Build attributes from changed flags
	attributes := map[string]any{}

	if cmd.Flags().Changed("number") {
		attributes["number"] = opts.Number
	}
	if cmd.Flags().Changed("truck-manufacturer-name") {
		attributes["truck-manufacturer-name"] = opts.TruckManufacturerName
	}
	if cmd.Flags().Changed("truck-model-name") {
		attributes["truck-model-name"] = opts.TruckModelName
	}
	if cmd.Flags().Changed("truck-model-year") {
		attributes["truck-model-year"] = opts.TruckModelYear
	}
	if cmd.Flags().Changed("color-name") {
		attributes["color-name"] = opts.ColorName
	}
	if cmd.Flags().Changed("plate-number") {
		attributes["plate-number"] = opts.PlateNumber
	}
	if cmd.Flags().Changed("plate-jurisdiction-code") {
		attributes["plate-jurisdiction-code"] = opts.PlateJurisdictionCode
	}
	if cmd.Flags().Changed("vin") {
		attributes["vin"] = opts.VIN
	}
	if cmd.Flags().Changed("apportioned") {
		attributes["apportioned"] = opts.Apportioned == "true"
	}
	if cmd.Flags().Changed("in-service") {
		attributes["in-service"] = opts.InService == "true"
	}
	if cmd.Flags().Changed("intransit-heat") {
		attributes["intransit-heat"] = opts.IntransitHeat == "true"
	}
	if cmd.Flags().Changed("curb-weight-lbs") {
		attributes["curb-weight-lbs"] = opts.CurbWeightLbs
	}
	if cmd.Flags().Changed("height-inches") {
		attributes["height-inches"] = opts.HeightInches
	}
	if cmd.Flags().Changed("parking-address") {
		attributes["parking-address"] = opts.ParkingAddress
	}
	if cmd.Flags().Changed("parking-address-place-id") {
		attributes["parking-address-place-id"] = opts.ParkingAddressPlaceID
	}
	if cmd.Flags().Changed("parking-address-plus-code") {
		attributes["parking-address-plus-code"] = opts.ParkingAddressPlusCode
	}
	if cmd.Flags().Changed("skip-parking-address-geocoding") {
		attributes["skip-parking-address-geocoding"] = opts.SkipParkingAddressGeocoding == "true"
	}

	// Build relationships from changed flags
	relationships := map[string]any{}

	if cmd.Flags().Changed("trucker") {
		relationships["trucker"] = map[string]any{
			"data": map[string]string{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		}
	}
	if cmd.Flags().Changed("brokered-tractor") {
		if opts.BrokeredTractor == "" {
			relationships["brokered-tractor"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["brokered-tractor"] = map[string]any{
				"data": map[string]string{
					"type": "tractors",
					"id":   opts.BrokeredTractor,
				},
			}
		}
	}

	// Require at least one field to update
	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestData := map[string]any{
		"type": "tractors",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		requestData["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestData["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/tractors/"+opts.ID, jsonBody)
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

	row := tractorRow{
		ID:     resp.Data.ID,
		Number: stringAttr(resp.Data.Attributes, "number"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated tractor %s (%s)\n", row.ID, row.Number)
	return nil
}

func parseDoTractorsUpdateOptions(cmd *cobra.Command, id string) (doTractorsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Attributes
	number, _ := cmd.Flags().GetString("number")
	truckManufacturerName, _ := cmd.Flags().GetString("truck-manufacturer-name")
	truckModelName, _ := cmd.Flags().GetString("truck-model-name")
	truckModelYear, _ := cmd.Flags().GetInt("truck-model-year")
	colorName, _ := cmd.Flags().GetString("color-name")

	// Registration
	plateNumber, _ := cmd.Flags().GetString("plate-number")
	plateJurisdictionCode, _ := cmd.Flags().GetString("plate-jurisdiction-code")
	vin, _ := cmd.Flags().GetString("vin")
	apportioned, _ := cmd.Flags().GetString("apportioned")

	// Status
	inService, _ := cmd.Flags().GetString("in-service")
	intransitHeat, _ := cmd.Flags().GetString("intransit-heat")

	// Physical specs
	curbWeightLbs, _ := cmd.Flags().GetInt("curb-weight-lbs")
	heightInches, _ := cmd.Flags().GetInt("height-inches")

	// Parking address
	parkingAddress, _ := cmd.Flags().GetString("parking-address")
	parkingAddressPlaceID, _ := cmd.Flags().GetString("parking-address-place-id")
	parkingAddressPlusCode, _ := cmd.Flags().GetString("parking-address-plus-code")
	skipParkingAddressGeocoding, _ := cmd.Flags().GetString("skip-parking-address-geocoding")

	// Relationships
	trucker, _ := cmd.Flags().GetString("trucker")
	brokeredTractor, _ := cmd.Flags().GetString("brokered-tractor")

	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorsUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,

		ID: id,

		Number:                number,
		TruckManufacturerName: truckManufacturerName,
		TruckModelName:        truckModelName,
		TruckModelYear:        truckModelYear,
		ColorName:             colorName,

		PlateNumber:           plateNumber,
		PlateJurisdictionCode: plateJurisdictionCode,
		VIN:                   vin,
		Apportioned:           apportioned,

		InService:     inService,
		IntransitHeat: intransitHeat,

		CurbWeightLbs: curbWeightLbs,
		HeightInches:  heightInches,

		ParkingAddress:              parkingAddress,
		ParkingAddressPlaceID:       parkingAddressPlaceID,
		ParkingAddressPlusCode:      parkingAddressPlusCode,
		SkipParkingAddressGeocoding: skipParkingAddressGeocoding,

		Trucker:         trucker,
		BrokeredTractor: brokeredTractor,
	}, nil
}
