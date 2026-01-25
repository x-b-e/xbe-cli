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

type doTractorsCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	// Required
	Number  string
	Trucker string

	// Truck details
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
	BrokeredTractor string
}

func newDoTractorsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tractor",
		Long: `Create a new tractor.

Required flags:
  --number     The tractor number/identifier (required)
  --trucker    The trucker ID (required)

Optional flags:
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
    --brokered-tractor          Brokered tractor ID`,
		Example: `  # Create a tractor with required fields
  xbe do tractors create --number "T100" --trucker 123

  # Create with full details
  xbe do tractors create --number "T200" --trucker 123 --truck-manufacturer-name "Peterbilt" --truck-model-name "579" --truck-model-year 2023

  # Create with parking address
  xbe do tractors create --number "T300" --trucker 123 --parking-address "123 Main St, Chicago, IL"

  # Get JSON output
  xbe do tractors create --number "T400" --trucker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTractorsCreate,
	}
	initDoTractorsCreateFlags(cmd)
	return cmd
}

func init() {
	doTractorsCmd.AddCommand(newDoTractorsCreateCmd())
}

func initDoTractorsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")

	// Required
	cmd.Flags().String("number", "", "Tractor number (required)")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")

	// Truck details
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

	// Relationship
	cmd.Flags().String("brokered-tractor", "", "Brokered tractor ID")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTractorsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTractorsCreateOptions(cmd)
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

	// Require number
	if opts.Number == "" {
		err := fmt.Errorf("--number is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Require trucker
	if opts.Trucker == "" {
		err := fmt.Errorf("--trucker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// Build attributes
	attributes := map[string]any{
		"number": opts.Number,
	}

	// Truck details
	if opts.TruckManufacturerName != "" {
		attributes["truck-manufacturer-name"] = opts.TruckManufacturerName
	}
	if opts.TruckModelName != "" {
		attributes["truck-model-name"] = opts.TruckModelName
	}
	if cmd.Flags().Changed("truck-model-year") {
		attributes["truck-model-year"] = opts.TruckModelYear
	}
	if opts.ColorName != "" {
		attributes["color-name"] = opts.ColorName
	}

	// Registration
	if opts.PlateNumber != "" {
		attributes["plate-number"] = opts.PlateNumber
	}
	if opts.PlateJurisdictionCode != "" {
		attributes["plate-jurisdiction-code"] = opts.PlateJurisdictionCode
	}
	if opts.VIN != "" {
		attributes["vin"] = opts.VIN
	}
	if opts.Apportioned != "" {
		attributes["apportioned"] = opts.Apportioned == "true"
	}

	// Status
	if opts.InService != "" {
		attributes["in-service"] = opts.InService == "true"
	}
	if opts.IntransitHeat != "" {
		attributes["intransit-heat"] = opts.IntransitHeat == "true"
	}

	// Physical specs
	if cmd.Flags().Changed("curb-weight-lbs") {
		attributes["curb-weight-lbs"] = opts.CurbWeightLbs
	}
	if cmd.Flags().Changed("height-inches") {
		attributes["height-inches"] = opts.HeightInches
	}

	// Parking address
	if opts.ParkingAddress != "" {
		attributes["parking-address"] = opts.ParkingAddress
	}
	if opts.ParkingAddressPlaceID != "" {
		attributes["parking-address-place-id"] = opts.ParkingAddressPlaceID
	}
	if opts.ParkingAddressPlusCode != "" {
		attributes["parking-address-plus-code"] = opts.ParkingAddressPlusCode
	}
	if opts.SkipParkingAddressGeocoding != "" {
		attributes["skip-parking-address-geocoding"] = opts.SkipParkingAddressGeocoding == "true"
	}

	// Build relationships
	relationships := map[string]any{
		"trucker": map[string]any{
			"data": map[string]string{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		},
	}

	// Optional relationship
	if opts.BrokeredTractor != "" {
		relationships["brokered-tractor"] = map[string]any{
			"data": map[string]string{
				"type": "tractors",
				"id":   opts.BrokeredTractor,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "tractors",
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

	body, _, err := client.Post(cmd.Context(), "/v1/tractors", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created tractor %s (%s)\n", row.ID, row.Number)
	return nil
}

func parseDoTractorsCreateOptions(cmd *cobra.Command) (doTractorsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Required
	number, _ := cmd.Flags().GetString("number")
	trucker, _ := cmd.Flags().GetString("trucker")

	// Truck details
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

	// Relationship
	brokeredTractor, _ := cmd.Flags().GetString("brokered-tractor")

	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTractorsCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,

		Number:  number,
		Trucker: trucker,

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

		BrokeredTractor: brokeredTractor,
	}, nil
}
