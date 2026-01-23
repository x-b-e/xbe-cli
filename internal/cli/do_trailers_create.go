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

type doTrailersCreateOptions struct {
	BaseURL string
	Token   string
	JSON    bool

	// Required
	Number  string
	Trucker string

	// Trailer details
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
	Tractor               string
	TrailerClassification string
	BrokeredTrailer       string
}

func newDoTrailersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new trailer",
		Long: `Create a new trailer.

Required flags:
  --number     The trailer number/identifier (required)
  --trucker    The trucker ID (required)

Optional flags:
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
    --tractor                     Tractor ID
    --trailer-classification      Trailer classification ID
    --brokered-trailer            Brokered trailer ID`,
		Example: `  # Create a trailer with required fields
  xbe do trailers create --number "TR100" --trucker 123

  # Create with full details
  xbe do trailers create --number "TR200" --trucker 123 --kind "dump" --composition "steel" --capacity-lbs 50000

  # Create with parking address
  xbe do trailers create --number "TR300" --trucker 123 --parking-address "123 Main St, Chicago, IL"

  # Get JSON output
  xbe do trailers create --number "TR400" --trucker 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoTrailersCreate,
	}
	initDoTrailersCreateFlags(cmd)
	return cmd
}

func init() {
	doTrailersCmd.AddCommand(newDoTrailersCreateCmd())
}

func initDoTrailersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")

	// Required
	cmd.Flags().String("number", "", "Trailer number (required)")
	cmd.Flags().String("trucker", "", "Trucker ID (required)")

	// Trailer details
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
	cmd.Flags().String("tractor", "", "Tractor ID")
	cmd.Flags().String("trailer-classification", "", "Trailer classification ID")
	cmd.Flags().String("brokered-trailer", "", "Brokered trailer ID")

	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTrailersCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTrailersCreateOptions(cmd)
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

	// Trailer details
	if opts.Composition != "" {
		attributes["composition"] = opts.Composition
	}
	if opts.Kind != "" {
		attributes["kind"] = opts.Kind
	}
	if opts.FrameKind != "" {
		attributes["frame-kind"] = opts.FrameKind
	}
	if opts.TarpKind != "" {
		attributes["tarp-kind"] = opts.TarpKind
	}
	if cmd.Flags().Changed("capacity-lbs") {
		attributes["capacity-lbs"] = opts.CapacityLbs
	}
	if cmd.Flags().Changed("rear-wheel-offset-inches") {
		attributes["rear-wheel-offset-inches"] = opts.RearWheelOffsetInches
	}
	if opts.Hitch != "" {
		attributes["hitch"] = opts.Hitch
	}
	if opts.HasOverweightPermitStatus != "" {
		attributes["has-overweight-permit-status"] = opts.HasOverweightPermitStatus
	}

	// Boolean attributes
	if opts.CoalChute != "" {
		attributes["coal-chute"] = opts.CoalChute == "true"
	}
	if opts.InsulatedBed != "" {
		attributes["insulated-bed"] = opts.InsulatedBed == "true"
	}
	if opts.BedLiner != "" {
		attributes["bed-liner"] = opts.BedLiner == "true"
	}
	if opts.SludgeLocks != "" {
		attributes["sludge-locks"] = opts.SludgeLocks == "true"
	}
	if opts.Vibrator != "" {
		attributes["vibrator"] = opts.Vibrator == "true"
	}
	if opts.IntransitHeat != "" {
		attributes["intransit-heat"] = opts.IntransitHeat == "true"
	}
	if opts.CanPave != "" {
		attributes["can-pave"] = opts.CanPave == "true"
	}
	if opts.InService != "" {
		attributes["in-service"] = opts.InService == "true"
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

	// Optional relationships
	if opts.Tractor != "" {
		relationships["tractor"] = map[string]any{
			"data": map[string]string{
				"type": "tractors",
				"id":   opts.Tractor,
			},
		}
	}
	if opts.TrailerClassification != "" {
		relationships["trailer-classification"] = map[string]any{
			"data": map[string]string{
				"type": "trailer-classifications",
				"id":   opts.TrailerClassification,
			},
		}
	}
	if opts.BrokeredTrailer != "" {
		relationships["brokered-trailer"] = map[string]any{
			"data": map[string]string{
				"type": "trailers",
				"id":   opts.BrokeredTrailer,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "trailers",
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

	body, _, err := client.Post(cmd.Context(), "/v1/trailers", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created trailer %s (%s)\n", row.ID, row.Number)
	return nil
}

func parseDoTrailersCreateOptions(cmd *cobra.Command) (doTrailersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")

	// Required
	number, _ := cmd.Flags().GetString("number")
	trucker, _ := cmd.Flags().GetString("trucker")

	// Trailer details
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
	tractor, _ := cmd.Flags().GetString("tractor")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	brokeredTrailer, _ := cmd.Flags().GetString("brokered-trailer")

	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTrailersCreateOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,

		Number:  number,
		Trucker: trucker,

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

		Tractor:               tractor,
		TrailerClassification: trailerClassification,
		BrokeredTrailer:       brokeredTrailer,
	}, nil
}
