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

type doTrailersUpdateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	// Trailer ID
	ID string

	// Attributes
	Number                    string
	Composition               string
	Kind                      string
	FrameKind                 string
	TarpKind                  string
	CapacityLbs               int
	RearWheelOffsetInches     int
	Hitch                     string
	HasOverweightPermitStatus string

	// Boolean attributes
	CoalChute     string
	InsulatedBed  string
	BedLiner      string
	SludgeLocks   string
	Vibrator      string
	IntransitHeat string
	CanPave       string
	InService     string

	// Physical specs
	CurbWeightLbs int
	HeightInches  int

	// Address
	ParkingAddress              string
	ParkingAddressPlaceID       string
	ParkingAddressPlusCode      string
	SkipParkingAddressGeocoding string

	// Relationships
	Trucker               string
	Tractor               string
	TrailerClassification string
	BrokeredTrailer       string
}

func newDoTrailersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a trailer",
		Long: `Update a trailer.

Provide the trailer ID as an argument and specify the fields to update.

Updatable fields:
  Basic:
    --number                      Trailer number/identifier

  Trailer details:
    --composition                 Trailer composition
    --kind                        Trailer kind
    --frame-kind                  Frame kind
    --tarp-kind                   Tarp kind
    --capacity-lbs                Capacity in pounds
    --rear-wheel-offset-inches    Rear wheel offset in inches
    --hitch                       Hitch type
    --has-overweight-permit-status Overweight permit status

  Features:
    --coal-chute                  Has coal chute (true/false)
    --insulated-bed               Has insulated bed (true/false)
    --bed-liner                   Has bed liner (true/false)
    --sludge-locks                Has sludge locks (true/false)
    --vibrator                    Has vibrator (true/false)
    --intransit-heat              Intransit heat (true/false)
    --can-pave                    Can pave (true/false)

  Status:
    --in-service                  In service status (true/false)

  Physical specs:
    --curb-weight-lbs             Curb weight in pounds
    --height-inches               Height in inches

  Parking address:
    --parking-address             Parking address
    --parking-address-place-id    Google Place ID
    --parking-address-plus-code   Plus code
    --skip-parking-address-geocoding Skip geocoding (true/false)

  Relationships:
    --trucker                     Trucker ID
    --tractor                     Tractor ID
    --trailer-classification      Trailer classification ID
    --brokered-trailer            Brokered trailer ID`,
		Example: `  # Update trailer number
  xbe do trailers update 123 --number "TR101"

  # Update trailer details
  xbe do trailers update 123 --kind "flatbed" --capacity-lbs 60000

  # Update in-service status
  xbe do trailers update 123 --in-service false

  # Update parking address
  xbe do trailers update 123 --parking-address "456 Oak St, Chicago, IL"

  # Get JSON output
  xbe do trailers update 123 --number "TR102" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTrailersUpdate,
	}
	initDoTrailersUpdateFlags(cmd)
	return cmd
}

func init() {
	doTrailersCmd.AddCommand(newDoTrailersUpdateCmd())
}

func initDoTrailersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")

	// Attributes
	cmd.Flags().String("number", "", "Trailer number")
	cmd.Flags().String("composition", "", "Trailer composition")
	cmd.Flags().String("kind", "", "Trailer kind")
	cmd.Flags().String("frame-kind", "", "Frame kind")
	cmd.Flags().String("tarp-kind", "", "Tarp kind")
	cmd.Flags().Int("capacity-lbs", 0, "Capacity in pounds")
	cmd.Flags().Int("rear-wheel-offset-inches", 0, "Rear wheel offset in inches")
	cmd.Flags().String("hitch", "", "Hitch type")
	cmd.Flags().String("has-overweight-permit-status", "", "Overweight permit status")

	// Boolean attributes
	cmd.Flags().String("coal-chute", "", "Has coal chute (true/false)")
	cmd.Flags().String("insulated-bed", "", "Has insulated bed (true/false)")
	cmd.Flags().String("bed-liner", "", "Has bed liner (true/false)")
	cmd.Flags().String("sludge-locks", "", "Has sludge locks (true/false)")
	cmd.Flags().String("vibrator", "", "Has vibrator (true/false)")
	cmd.Flags().String("intransit-heat", "", "Intransit heat (true/false)")
	cmd.Flags().String("can-pave", "", "Can pave (true/false)")
	cmd.Flags().String("in-service", "", "In service status (true/false)")

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
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID")
	cmd.Flags().String("brokered-trailer", "", "Brokered trailer ID")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTrailersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTrailersUpdateOptions(cmd, args[0])
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
	if cmd.Flags().Changed("composition") {
		attributes["composition"] = opts.Composition
	}
	if cmd.Flags().Changed("kind") {
		attributes["kind"] = opts.Kind
	}
	if cmd.Flags().Changed("frame-kind") {
		attributes["frame-kind"] = opts.FrameKind
	}
	if cmd.Flags().Changed("tarp-kind") {
		attributes["tarp-kind"] = opts.TarpKind
	}
	if cmd.Flags().Changed("capacity-lbs") {
		attributes["capacity-lbs"] = opts.CapacityLbs
	}
	if cmd.Flags().Changed("rear-wheel-offset-inches") {
		attributes["rear-wheel-offset-inches"] = opts.RearWheelOffsetInches
	}
	if cmd.Flags().Changed("hitch") {
		attributes["hitch"] = opts.Hitch
	}
	if cmd.Flags().Changed("has-overweight-permit-status") {
		attributes["has-overweight-permit-status"] = opts.HasOverweightPermitStatus
	}
	if cmd.Flags().Changed("coal-chute") {
		attributes["coal-chute"] = opts.CoalChute == "true"
	}
	if cmd.Flags().Changed("insulated-bed") {
		attributes["insulated-bed"] = opts.InsulatedBed == "true"
	}
	if cmd.Flags().Changed("bed-liner") {
		attributes["bed-liner"] = opts.BedLiner == "true"
	}
	if cmd.Flags().Changed("sludge-locks") {
		attributes["sludge-locks"] = opts.SludgeLocks == "true"
	}
	if cmd.Flags().Changed("vibrator") {
		attributes["vibrator"] = opts.Vibrator == "true"
	}
	if cmd.Flags().Changed("intransit-heat") {
		attributes["intransit-heat"] = opts.IntransitHeat == "true"
	}
	if cmd.Flags().Changed("can-pave") {
		attributes["can-pave"] = opts.CanPave == "true"
	}
	if cmd.Flags().Changed("in-service") {
		attributes["in-service"] = opts.InService == "true"
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
	if cmd.Flags().Changed("tractor") {
		if opts.Tractor == "" {
			relationships["tractor"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["tractor"] = map[string]any{
				"data": map[string]string{
					"type": "tractors",
					"id":   opts.Tractor,
				},
			}
		}
	}
	if cmd.Flags().Changed("trailer-classification") {
		if opts.TrailerClassification == "" {
			relationships["trailer-classification"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["trailer-classification"] = map[string]any{
				"data": map[string]string{
					"type": "trailer-classifications",
					"id":   opts.TrailerClassification,
				},
			}
		}
	}
	if cmd.Flags().Changed("brokered-trailer") {
		if opts.BrokeredTrailer == "" {
			relationships["brokered-trailer"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["brokered-trailer"] = map[string]any{
				"data": map[string]string{
					"type": "trailers",
					"id":   opts.BrokeredTrailer,
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
		"type": "trailers",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/trailers/"+opts.ID, jsonBody)
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

	row := trailerRow{
		ID:     resp.Data.ID,
		Number: stringAttr(resp.Data.Attributes, "number"),
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated trailer %s (%s)\n", row.ID, row.Number)
	return nil
}

func parseDoTrailersUpdateOptions(cmd *cobra.Command, id string) (doTrailersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Attributes
	number, _ := cmd.Flags().GetString("number")
	composition, _ := cmd.Flags().GetString("composition")
	kind, _ := cmd.Flags().GetString("kind")
	frameKind, _ := cmd.Flags().GetString("frame-kind")
	tarpKind, _ := cmd.Flags().GetString("tarp-kind")
	capacityLbs, _ := cmd.Flags().GetInt("capacity-lbs")
	rearWheelOffsetInches, _ := cmd.Flags().GetInt("rear-wheel-offset-inches")
	hitch, _ := cmd.Flags().GetString("hitch")
	hasOverweightPermitStatus, _ := cmd.Flags().GetString("has-overweight-permit-status")

	// Boolean attributes
	coalChute, _ := cmd.Flags().GetString("coal-chute")
	insulatedBed, _ := cmd.Flags().GetString("insulated-bed")
	bedLiner, _ := cmd.Flags().GetString("bed-liner")
	sludgeLocks, _ := cmd.Flags().GetString("sludge-locks")
	vibrator, _ := cmd.Flags().GetString("vibrator")
	intransitHeat, _ := cmd.Flags().GetString("intransit-heat")
	canPave, _ := cmd.Flags().GetString("can-pave")
	inService, _ := cmd.Flags().GetString("in-service")

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
	tractor, _ := cmd.Flags().GetString("tractor")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	brokeredTrailer, _ := cmd.Flags().GetString("brokered-trailer")

	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTrailersUpdateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,

		ID: id,

		Number:                    number,
		Composition:               composition,
		Kind:                      kind,
		FrameKind:                 frameKind,
		TarpKind:                  tarpKind,
		CapacityLbs:               capacityLbs,
		RearWheelOffsetInches:     rearWheelOffsetInches,
		Hitch:                     hitch,
		HasOverweightPermitStatus: hasOverweightPermitStatus,

		CoalChute:     coalChute,
		InsulatedBed:  insulatedBed,
		BedLiner:      bedLiner,
		SludgeLocks:   sludgeLocks,
		Vibrator:      vibrator,
		IntransitHeat: intransitHeat,
		CanPave:       canPave,
		InService:     inService,

		CurbWeightLbs: curbWeightLbs,
		HeightInches:  heightInches,

		ParkingAddress:              parkingAddress,
		ParkingAddressPlaceID:       parkingAddressPlaceID,
		ParkingAddressPlusCode:      parkingAddressPlusCode,
		SkipParkingAddressGeocoding: skipParkingAddressGeocoding,

		Trucker:               trucker,
		Tractor:               tractor,
		TrailerClassification: trailerClassification,
		BrokeredTrailer:       brokeredTrailer,
	}, nil
}
