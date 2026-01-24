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

type doProjectTransportLocationsUpdateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	ID                           string
	Name                         string
	ExternalTmsCompanyID         string
	GeocodingMethod              string
	AddressLatitude              string
	AddressLongitude             string
	AddressFull                  string
	AddressStreetOne             string
	AddressStreetTwo             string
	AddressCity                  string
	AddressStateCode             string
	AddressCountryCode           string
	AddressPostalCode            string
	AddressTimeZoneID            string
	AddressSPLC                  string
	IsActive                     bool
	IsValidForStop               bool
	SkipDetection                bool
	ProjectTransportOrganization string
}

func newDoProjectTransportLocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project transport location",
		Long: `Update an existing project transport location.

Provide the location ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --name                          Location name
  --external-tms-company-id        External TMS company ID
  --geocoding-method               Geocoding method
  --address-latitude               Address latitude
  --address-longitude              Address longitude
  --address-full                   Full address string
  --address-street-one             Address line 1
  --address-street-two             Address line 2
  --address-city                   City
  --address-state-code             State/region code
  --address-country-code           Country code (US/CA/MX)
  --address-postal-code            Postal code
  --address-time-zone-id           Time zone identifier
  --address-splc                   SPLC (9-digit)
  --is-active                      Whether the location is active
  --is-valid-for-stop              Whether the location is valid for stops
  --skip-detection                 Skip managed location detection
  --project-transport-organization Project transport organization ID (empty to clear)`,
		Example: `  # Update name
  xbe do project-transport-locations update 123 --name "Updated Name"

  # Update coordinates and geocoding method
  xbe do project-transport-locations update 123 --geocoding-method explicit --address-latitude 41.9 --address-longitude -87.6

  # Clear project transport organization
  xbe do project-transport-locations update 123 --project-transport-organization ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportLocationsUpdate,
	}
	initDoProjectTransportLocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportLocationsCmd.AddCommand(newDoProjectTransportLocationsUpdateCmd())
}

func initDoProjectTransportLocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Location name")
	cmd.Flags().String("external-tms-company-id", "", "External TMS company ID")
	cmd.Flags().String("geocoding-method", "", "Geocoding method")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().String("address-full", "", "Full address string")
	cmd.Flags().String("address-street-one", "", "Address line 1")
	cmd.Flags().String("address-street-two", "", "Address line 2")
	cmd.Flags().String("address-city", "", "City")
	cmd.Flags().String("address-state-code", "", "State/region code")
	cmd.Flags().String("address-country-code", "", "Country code (US/CA/MX)")
	cmd.Flags().String("address-postal-code", "", "Postal code")
	cmd.Flags().String("address-time-zone-id", "", "Time zone identifier")
	cmd.Flags().String("address-splc", "", "SPLC (9-digit)")
	cmd.Flags().Bool("is-active", true, "Whether the location is active")
	cmd.Flags().Bool("is-valid-for-stop", true, "Whether the location is valid for stops")
	cmd.Flags().Bool("skip-detection", false, "Skip managed location detection")
	cmd.Flags().String("project-transport-organization", "", "Project transport organization ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportLocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportLocationsUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("external-tms-company-id") {
		attributes["external-tms-company-id"] = opts.ExternalTmsCompanyID
	}
	if cmd.Flags().Changed("geocoding-method") {
		attributes["geocoding-method"] = opts.GeocodingMethod
	}
	if cmd.Flags().Changed("address-latitude") {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if cmd.Flags().Changed("address-longitude") {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if cmd.Flags().Changed("address-full") {
		attributes["address-full"] = opts.AddressFull
	}
	if cmd.Flags().Changed("address-street-one") {
		attributes["address-street-one"] = opts.AddressStreetOne
	}
	if cmd.Flags().Changed("address-street-two") {
		attributes["address-street-two"] = opts.AddressStreetTwo
	}
	if cmd.Flags().Changed("address-city") {
		attributes["address-city"] = opts.AddressCity
	}
	if cmd.Flags().Changed("address-state-code") {
		attributes["address-state-code"] = opts.AddressStateCode
	}
	if cmd.Flags().Changed("address-country-code") {
		attributes["address-country-code"] = opts.AddressCountryCode
	}
	if cmd.Flags().Changed("address-postal-code") {
		attributes["address-postal-code"] = opts.AddressPostalCode
	}
	if cmd.Flags().Changed("address-time-zone-id") {
		attributes["address-time-zone-id"] = opts.AddressTimeZoneID
	}
	if cmd.Flags().Changed("address-splc") {
		attributes["address-splc"] = opts.AddressSPLC
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}
	if cmd.Flags().Changed("is-valid-for-stop") {
		attributes["is-valid-for-stop"] = opts.IsValidForStop
	}
	if cmd.Flags().Changed("skip-detection") {
		attributes["skip-detection"] = opts.SkipDetection
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("project-transport-organization") {
		if opts.ProjectTransportOrganization == "" {
			relationships["project-transport-organization"] = map[string]any{
				"data": nil,
			}
		} else {
			relationships["project-transport-organization"] = map[string]any{
				"data": map[string]any{
					"type": "project-transport-organizations",
					"id":   opts.ProjectTransportOrganization,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field or --project-transport-organization")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type": "project-transport-locations",
			"id":   opts.ID,
		},
	}
	if len(attributes) > 0 {
		requestBody["data"].(map[string]any)["attributes"] = attributes
	}
	if len(relationships) > 0 {
		requestBody["data"].(map[string]any)["relationships"] = relationships
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-locations/"+opts.ID, jsonBody)
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

	row := buildProjectTransportLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport location %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProjectTransportLocationsUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportLocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	externalTmsCompanyID, _ := cmd.Flags().GetString("external-tms-company-id")
	geocodingMethod, _ := cmd.Flags().GetString("geocoding-method")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	addressFull, _ := cmd.Flags().GetString("address-full")
	addressStreetOne, _ := cmd.Flags().GetString("address-street-one")
	addressStreetTwo, _ := cmd.Flags().GetString("address-street-two")
	addressCity, _ := cmd.Flags().GetString("address-city")
	addressStateCode, _ := cmd.Flags().GetString("address-state-code")
	addressCountryCode, _ := cmd.Flags().GetString("address-country-code")
	addressPostalCode, _ := cmd.Flags().GetString("address-postal-code")
	addressTimeZoneID, _ := cmd.Flags().GetString("address-time-zone-id")
	addressSPLC, _ := cmd.Flags().GetString("address-splc")
	isActive, _ := cmd.Flags().GetBool("is-active")
	isValidForStop, _ := cmd.Flags().GetBool("is-valid-for-stop")
	skipDetection, _ := cmd.Flags().GetBool("skip-detection")
	projectTransportOrganization, _ := cmd.Flags().GetString("project-transport-organization")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportLocationsUpdateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		ID:                           args[0],
		Name:                         name,
		ExternalTmsCompanyID:         externalTmsCompanyID,
		GeocodingMethod:              geocodingMethod,
		AddressLatitude:              addressLatitude,
		AddressLongitude:             addressLongitude,
		AddressFull:                  addressFull,
		AddressStreetOne:             addressStreetOne,
		AddressStreetTwo:             addressStreetTwo,
		AddressCity:                  addressCity,
		AddressStateCode:             addressStateCode,
		AddressCountryCode:           addressCountryCode,
		AddressPostalCode:            addressPostalCode,
		AddressTimeZoneID:            addressTimeZoneID,
		AddressSPLC:                  addressSPLC,
		IsActive:                     isActive,
		IsValidForStop:               isValidForStop,
		SkipDetection:                skipDetection,
		ProjectTransportOrganization: projectTransportOrganization,
	}, nil
}
