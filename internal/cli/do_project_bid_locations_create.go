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

type doProjectBidLocationsCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	Project              string
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

func newDoProjectBidLocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project bid location",
		Long: `Create a project bid location.

Required flags:
  --project     Project ID
  --geometry    Geometry in WKT or EWKB hex

Optional flags:
  --name                   Location name
  --notes                  Location notes
  --address                Full address
  --address-latitude       Address latitude
  --address-longitude      Address longitude
  --address-place-id       Address place ID
  --address-plus-code      Address plus code
  --skip-address-geocoding Skip geocoding the address

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project bid location
  xbe do project-bid-locations create --project 123 --geometry "POINT(-77.0365 38.8977)"

  # Create with address details
  xbe do project-bid-locations create --project 123 --geometry "POINT(-77.0365 38.8977)" \
    --name "Main Yard" --address "1600 Pennsylvania Ave NW, Washington, DC"`,
		RunE: runDoProjectBidLocationsCreate,
	}
	initDoProjectBidLocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectBidLocationsCmd.AddCommand(newDoProjectBidLocationsCreateCmd())
}

func initDoProjectBidLocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID (required)")
	cmd.Flags().String("geometry", "", "Geometry in WKT or EWKB hex (required)")
	cmd.Flags().String("name", "", "Location name")
	cmd.Flags().String("notes", "", "Location notes")
	cmd.Flags().String("address", "", "Full address")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().String("address-place-id", "", "Address place ID")
	cmd.Flags().String("address-plus-code", "", "Address plus code")
	cmd.Flags().Bool("skip-address-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("geometry")
}

func runDoProjectBidLocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectBidLocationsCreateOptions(cmd)
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
		"geometry": opts.Geometry,
	}

	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.Notes != "" {
		attributes["notes"] = opts.Notes
	}
	if opts.Address != "" {
		attributes["address"] = opts.Address
	}
	if opts.AddressLatitude != "" {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if opts.AddressLongitude != "" {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if opts.AddressPlaceID != "" {
		attributes["address-place-id"] = opts.AddressPlaceID
	}
	if opts.AddressPlusCode != "" {
		attributes["address-plus-code"] = opts.AddressPlusCode
	}
	if opts.SkipAddressGeocoding {
		attributes["skip-address-geocoding"] = true
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-bid-locations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-bid-locations", jsonBody)
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
	fmt.Fprintf(cmd.OutOrStdout(), "Created project bid location %s\n", label)
	return nil
}

func parseDoProjectBidLocationsCreateOptions(cmd *cobra.Command) (doProjectBidLocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
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

	return doProjectBidLocationsCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		Project:              project,
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
