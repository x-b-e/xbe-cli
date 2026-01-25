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

type doProjectBidLocationsUpdateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ID                   string
	Name                 string
	Notes                string
	Geometry             string
	Address              string
	AddressLatitude      string
	AddressLongitude     string
	AddressPlaceID       string
	AddressPlusCode      string
	SkipAddressGeocoding bool
}

func newDoProjectBidLocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project bid location",
		Long: `Update a project bid location.

Optional flags:
  --name                   Location name
  --notes                  Location notes
  --geometry               Geometry in WKT or EWKB hex
  --address                Full address
  --address-latitude       Address latitude
  --address-longitude      Address longitude
  --address-place-id       Address place ID
  --address-plus-code      Address plus code
  --skip-address-geocoding Skip geocoding the address

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a project bid location name
  xbe do project-bid-locations update 123 --name "Updated Name"

  # Update geometry
  xbe do project-bid-locations update 123 --geometry "LINESTRING(-77.0365 38.8977,-77.0400 38.9000)"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectBidLocationsUpdate,
	}
	initDoProjectBidLocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectBidLocationsCmd.AddCommand(newDoProjectBidLocationsUpdateCmd())
}

func initDoProjectBidLocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Location name")
	cmd.Flags().String("notes", "", "Location notes")
	cmd.Flags().String("geometry", "", "Geometry in WKT or EWKB hex")
	cmd.Flags().String("address", "", "Full address")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().String("address-place-id", "", "Address place ID")
	cmd.Flags().String("address-plus-code", "", "Address plus code")
	cmd.Flags().Bool("skip-address-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectBidLocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectBidLocationsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("geometry") {
		attributes["geometry"] = opts.Geometry
	}
	if cmd.Flags().Changed("address") {
		attributes["address"] = opts.Address
	}
	if cmd.Flags().Changed("address-latitude") {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if cmd.Flags().Changed("address-longitude") {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if cmd.Flags().Changed("address-place-id") {
		attributes["address-place-id"] = opts.AddressPlaceID
	}
	if cmd.Flags().Changed("address-plus-code") {
		attributes["address-plus-code"] = opts.AddressPlusCode
	}
	if cmd.Flags().Changed("skip-address-geocoding") {
		attributes["skip-address-geocoding"] = opts.SkipAddressGeocoding
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "project-bid-locations",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-bid-locations/"+opts.ID, jsonBody)
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

	row := projectBidLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	label := firstNonEmpty(row.Name, row.ID)
	fmt.Fprintf(cmd.OutOrStdout(), "Updated project bid location %s\n", label)
	return nil
}

func parseDoProjectBidLocationsUpdateOptions(cmd *cobra.Command, args []string) (doProjectBidLocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	notes, _ := cmd.Flags().GetString("notes")
	geometry, _ := cmd.Flags().GetString("geometry")
	address, _ := cmd.Flags().GetString("address")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	addressPlaceID, _ := cmd.Flags().GetString("address-place-id")
	addressPlusCode, _ := cmd.Flags().GetString("address-plus-code")
	skipAddressGeocoding, _ := cmd.Flags().GetBool("skip-address-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectBidLocationsUpdateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ID:                   args[0],
		Name:                 name,
		Notes:                notes,
		Geometry:             geometry,
		Address:              address,
		AddressLatitude:      addressLatitude,
		AddressLongitude:     addressLongitude,
		AddressPlaceID:       addressPlaceID,
		AddressPlusCode:      addressPlusCode,
		SkipAddressGeocoding: skipAddressGeocoding,
	}, nil
}
