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

type doEquipmentMovementTripDispatchesCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	EquipmentMovementTrip        string
	EquipmentMovementRequirement string
	InboundEquipmentRequirement  string
	OutboundEquipmentRequirement string
	OriginLocation               string
	DestinationLocation          string
	Trucker                      string
	Driver                       string
	Trailer                      string
	TellClerkSynchronously       bool
}

func newDoEquipmentMovementTripDispatchesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an equipment movement trip dispatch",
		Long: `Create an equipment movement trip dispatch.

Input modes (choose one):
  --equipment-movement-trip            Start from an existing equipment movement trip
  --equipment-movement-requirement     Start from an existing equipment movement requirement
  --inbound-equipment-requirement and/or --outbound-equipment-requirement
                                       Start from equipment requirements (may include origin/destination locations)

Optional assignment flags:
  --trucker        Trucker ID
  --driver         Driver (user) ID
  --trailer        Trailer ID

Optional attributes:
  --tell-clerk-synchronously   Process fulfillment synchronously

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create from an existing trip
  xbe do equipment-movement-trip-dispatches create \
    --equipment-movement-trip 123

  # Create from an equipment movement requirement
  xbe do equipment-movement-trip-dispatches create \
    --equipment-movement-requirement 456

  # Create from equipment requirements
  xbe do equipment-movement-trip-dispatches create \
    --inbound-equipment-requirement 789 \
    --origin-location 321

  # Create with assignments
  xbe do equipment-movement-trip-dispatches create \
    --equipment-movement-trip 123 \
    --trucker 55 \
    --driver 66 \
    --trailer 77`,
		Args: cobra.NoArgs,
		RunE: runDoEquipmentMovementTripDispatchesCreate,
	}
	initDoEquipmentMovementTripDispatchesCreateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementTripDispatchesCmd.AddCommand(newDoEquipmentMovementTripDispatchesCreateCmd())
}

func initDoEquipmentMovementTripDispatchesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("equipment-movement-trip", "", "Equipment movement trip ID")
	cmd.Flags().String("equipment-movement-requirement", "", "Equipment movement requirement ID")
	cmd.Flags().String("inbound-equipment-requirement", "", "Inbound equipment requirement ID")
	cmd.Flags().String("outbound-equipment-requirement", "", "Outbound equipment requirement ID")
	cmd.Flags().String("origin-location", "", "Origin location ID (equipment movement requirement location)")
	cmd.Flags().String("destination-location", "", "Destination location ID (equipment movement requirement location)")
	cmd.Flags().String("trucker", "", "Trucker ID")
	cmd.Flags().String("driver", "", "Driver (user) ID")
	cmd.Flags().String("trailer", "", "Trailer ID")
	cmd.Flags().Bool("tell-clerk-synchronously", false, "Process fulfillment synchronously")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementTripDispatchesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoEquipmentMovementTripDispatchesCreateOptions(cmd)
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

	hasTrip := strings.TrimSpace(opts.EquipmentMovementTrip) != ""
	hasRequirement := strings.TrimSpace(opts.EquipmentMovementRequirement) != ""
	hasEquipmentRequirements := strings.TrimSpace(opts.InboundEquipmentRequirement) != "" || strings.TrimSpace(opts.OutboundEquipmentRequirement) != ""
	hasLocations := strings.TrimSpace(opts.OriginLocation) != "" || strings.TrimSpace(opts.DestinationLocation) != ""

	modeCount := 0
	if hasTrip {
		modeCount++
	}
	if hasRequirement {
		modeCount++
	}
	if hasEquipmentRequirements || hasLocations {
		modeCount++
	}

	if modeCount == 0 {
		err := fmt.Errorf("must specify one input mode: --equipment-movement-trip, --equipment-movement-requirement, or equipment requirement flags")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if modeCount > 1 {
		err := fmt.Errorf("only one input mode can be specified")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if hasLocations && !hasEquipmentRequirements {
		err := fmt.Errorf("origin/destination locations require inbound or outbound equipment requirements")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("tell-clerk-synchronously") {
		attributes["tell-clerk-synchronously"] = opts.TellClerkSynchronously
	}

	relationships := map[string]any{}

	if hasTrip {
		relationships["equipment-movement-trip"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-trips",
				"id":   opts.EquipmentMovementTrip,
			},
		}
	}
	if hasRequirement {
		relationships["equipment-movement-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-requirements",
				"id":   opts.EquipmentMovementRequirement,
			},
		}
	}
	if strings.TrimSpace(opts.InboundEquipmentRequirement) != "" {
		relationships["inbound-equipment-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "crew-requirements",
				"id":   opts.InboundEquipmentRequirement,
			},
		}
	}
	if strings.TrimSpace(opts.OutboundEquipmentRequirement) != "" {
		relationships["outbound-equipment-requirement"] = map[string]any{
			"data": map[string]any{
				"type": "crew-requirements",
				"id":   opts.OutboundEquipmentRequirement,
			},
		}
	}
	if strings.TrimSpace(opts.OriginLocation) != "" {
		relationships["origin-location"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-requirement-locations",
				"id":   opts.OriginLocation,
			},
		}
	}
	if strings.TrimSpace(opts.DestinationLocation) != "" {
		relationships["destination-location"] = map[string]any{
			"data": map[string]any{
				"type": "equipment-movement-requirement-locations",
				"id":   opts.DestinationLocation,
			},
		}
	}
	if strings.TrimSpace(opts.Trucker) != "" {
		relationships["trucker"] = map[string]any{
			"data": map[string]any{
				"type": "truckers",
				"id":   opts.Trucker,
			},
		}
	}
	if strings.TrimSpace(opts.Driver) != "" {
		relationships["driver"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Driver,
			},
		}
	}
	if strings.TrimSpace(opts.Trailer) != "" {
		relationships["trailer"] = map[string]any{
			"data": map[string]any{
				"type": "trailers",
				"id":   opts.Trailer,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "equipment-movement-trip-dispatches",
			"relationships": relationships,
		},
	}

	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/equipment-movement-trip-dispatches", jsonBody)
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

	row := buildEquipmentMovementTripDispatchRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created equipment movement trip dispatch %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementTripDispatchesCreateOptions(cmd *cobra.Command) (doEquipmentMovementTripDispatchesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	equipmentMovementTrip, _ := cmd.Flags().GetString("equipment-movement-trip")
	equipmentMovementRequirement, _ := cmd.Flags().GetString("equipment-movement-requirement")
	inboundEquipmentRequirement, _ := cmd.Flags().GetString("inbound-equipment-requirement")
	outboundEquipmentRequirement, _ := cmd.Flags().GetString("outbound-equipment-requirement")
	originLocation, _ := cmd.Flags().GetString("origin-location")
	destinationLocation, _ := cmd.Flags().GetString("destination-location")
	trucker, _ := cmd.Flags().GetString("trucker")
	driver, _ := cmd.Flags().GetString("driver")
	trailer, _ := cmd.Flags().GetString("trailer")
	tellClerkSynchronously, _ := cmd.Flags().GetBool("tell-clerk-synchronously")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementTripDispatchesCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		EquipmentMovementTrip:        equipmentMovementTrip,
		EquipmentMovementRequirement: equipmentMovementRequirement,
		InboundEquipmentRequirement:  inboundEquipmentRequirement,
		OutboundEquipmentRequirement: outboundEquipmentRequirement,
		OriginLocation:               originLocation,
		DestinationLocation:          destinationLocation,
		Trucker:                      trucker,
		Driver:                       driver,
		Trailer:                      trailer,
		TellClerkSynchronously:       tellClerkSynchronously,
	}, nil
}

func buildEquipmentMovementTripDispatchRowFromSingle(resp jsonAPISingleResponse) equipmentMovementTripDispatchRow {
	attrs := resp.Data.Attributes
	row := equipmentMovementTripDispatchRow{
		ID:        resp.Data.ID,
		Status:    stringAttr(attrs, "status"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
	}

	if rel, ok := resp.Data.Relationships["equipment-movement-trip"]; ok && rel.Data != nil {
		row.EquipmentMovementTripID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["equipment-movement-requirement"]; ok && rel.Data != nil {
		row.EquipmentMovementRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["inbound-equipment-requirement"]; ok && rel.Data != nil {
		row.InboundEquipmentRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["outbound-equipment-requirement"]; ok && rel.Data != nil {
		row.OutboundEquipmentRequirementID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["trailer"]; ok && rel.Data != nil {
		row.TrailerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["lineup-dispatch"]; ok && rel.Data != nil {
		row.LineupDispatchID = rel.Data.ID
	}

	return row
}
