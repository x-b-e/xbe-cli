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

type doProjectTransportLocationsCreateOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
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
	Broker                       string
	ProjectTransportOrganization string
}

func newDoProjectTransportLocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project transport location",
		Long: `Create a new project transport location.

Required flags:
  --geocoding-method  Geocoding method (required)
  --broker            Broker ID (required)

Optional flags:
  --name                          Location name
  --external-tms-company-id        External TMS company ID
  --project-transport-organization Project transport organization ID
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
  --is-active                      Whether the location is active (default true)
  --is-valid-for-stop              Whether the location is valid for stops (default true)
  --skip-detection                 Skip managed location detection`,
		Example: `  # Create a project transport location with explicit coordinates
  xbe do project-transport-locations create \\
    --name "North Yard" \\
    --geocoding-method explicit \\
    --address-latitude 41.8781 \\
    --address-longitude -87.6298 \\
    --broker 123

  # Create with address and organization
  xbe do project-transport-locations create \\
    --name "Plant A" \\
    --geocoding-method forward \\
    --address-full "123 Main St, Chicago, IL 60601" \\
    --project-transport-organization 456 \\
    --broker 123`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportLocationsCreate,
	}
	initDoProjectTransportLocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportLocationsCmd.AddCommand(newDoProjectTransportLocationsCreateCmd())
}

func initDoProjectTransportLocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Location name")
	cmd.Flags().String("external-tms-company-id", "", "External TMS company ID")
	cmd.Flags().String("geocoding-method", "", "Geocoding method (required)")
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
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("project-transport-organization", "", "Project transport organization ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportLocationsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportLocationsCreateOptions(cmd)
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

	if opts.GeocodingMethod == "" {
		err := fmt.Errorf("--geocoding-method is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Broker == "" {
		err := fmt.Errorf("--broker is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"geocoding-method":  opts.GeocodingMethod,
		"is-active":         opts.IsActive,
		"is-valid-for-stop": opts.IsValidForStop,
	}

	if opts.Name != "" {
		attributes["name"] = opts.Name
	}
	if opts.ExternalTmsCompanyID != "" {
		attributes["external-tms-company-id"] = opts.ExternalTmsCompanyID
	}
	if opts.AddressLatitude != "" {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if opts.AddressLongitude != "" {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if opts.AddressFull != "" {
		attributes["address-full"] = opts.AddressFull
	}
	if opts.AddressStreetOne != "" {
		attributes["address-street-one"] = opts.AddressStreetOne
	}
	if opts.AddressStreetTwo != "" {
		attributes["address-street-two"] = opts.AddressStreetTwo
	}
	if opts.AddressCity != "" {
		attributes["address-city"] = opts.AddressCity
	}
	if opts.AddressStateCode != "" {
		attributes["address-state-code"] = opts.AddressStateCode
	}
	if opts.AddressCountryCode != "" {
		attributes["address-country-code"] = opts.AddressCountryCode
	}
	if opts.AddressPostalCode != "" {
		attributes["address-postal-code"] = opts.AddressPostalCode
	}
	if opts.AddressTimeZoneID != "" {
		attributes["address-time-zone-id"] = opts.AddressTimeZoneID
	}
	if opts.AddressSPLC != "" {
		attributes["address-splc"] = opts.AddressSPLC
	}
	if cmd.Flags().Changed("skip-detection") {
		attributes["skip-detection"] = opts.SkipDetection
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.Broker,
			},
		},
	}

	if opts.ProjectTransportOrganization != "" {
		relationships["project-transport-organization"] = map[string]any{
			"data": map[string]any{
				"type": "project-transport-organizations",
				"id":   opts.ProjectTransportOrganization,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-transport-locations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-locations", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport location %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoProjectTransportLocationsCreateOptions(cmd *cobra.Command) (doProjectTransportLocationsCreateOptions, error) {
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
	broker, _ := cmd.Flags().GetString("broker")
	projectTransportOrganization, _ := cmd.Flags().GetString("project-transport-organization")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportLocationsCreateOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
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
		Broker:                       broker,
		ProjectTransportOrganization: projectTransportOrganization,
	}, nil
}

func buildProjectTransportLocationRowFromSingle(resp jsonAPISingleResponse) projectTransportLocationRow {
	attrs := resp.Data.Attributes
	row := projectTransportLocationRow{
		ID:                   resp.Data.ID,
		Name:                 stringAttr(attrs, "name"),
		ExternalTmsCompanyID: stringAttr(attrs, "external-tms-company-id"),
		GeocodingMethod:      stringAttr(attrs, "geocoding-method"),
		AddressCity:          stringAttr(attrs, "address-city"),
		AddressStateCode:     stringAttr(attrs, "address-state-code"),
		IsActive:             boolAttr(attrs, "is-active"),
		IsValidForStop:       boolAttr(attrs, "is-valid-for-stop"),
		DistanceInMiles:      attrs["distance-in-miles"],
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-transport-organization"]; ok && rel.Data != nil {
		row.ProjectTransportOrganizationID = rel.Data.ID
	}

	return row
}
