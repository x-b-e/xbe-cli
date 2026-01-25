package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectBidLocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectBidLocationDetails struct {
	ID                                string   `json:"id"`
	Name                              string   `json:"name,omitempty"`
	Notes                             string   `json:"notes,omitempty"`
	Geometry                          string   `json:"geometry,omitempty"`
	GeometryGeoJSON                   string   `json:"geometry_geojson,omitempty"`
	CentroidLatitude                  string   `json:"centroid_latitude,omitempty"`
	CentroidLongitude                 string   `json:"centroid_longitude,omitempty"`
	Address                           string   `json:"address,omitempty"`
	AddressFormatted                  string   `json:"address_formatted,omitempty"`
	AddressTimeZoneID                 string   `json:"address_time_zone_id,omitempty"`
	AddressCity                       string   `json:"address_city,omitempty"`
	AddressStateCode                  string   `json:"address_state_code,omitempty"`
	AddressCounty                     string   `json:"address_county,omitempty"`
	AddressCountryCode                string   `json:"address_country_code,omitempty"`
	AddressLatitude                   string   `json:"address_latitude,omitempty"`
	AddressLongitude                  string   `json:"address_longitude,omitempty"`
	AddressPlaceID                    string   `json:"address_place_id,omitempty"`
	AddressPlusCode                   string   `json:"address_plus_code,omitempty"`
	SkipAddressGeocoding              bool     `json:"skip_address_geocoding"`
	ProjectID                         string   `json:"project_id,omitempty"`
	ProjectName                       string   `json:"project_name,omitempty"`
	BrokerID                          string   `json:"broker_id,omitempty"`
	BrokerName                        string   `json:"broker_name,omitempty"`
	MaterialTypeIDs                   []string `json:"material_type_ids,omitempty"`
	ProjectBidLocationMaterialTypeIDs []string `json:"project_bid_location_material_type_ids,omitempty"`
}

func newProjectBidLocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project bid location details",
		Long: `Show the full details of a project bid location.

Output Fields:
  ID                              Project bid location identifier
  Name                            Location name
  Notes                           Location notes
  Geometry                        Geometry WKT or EWKB
  Geometry GeoJSON                Geometry rendered as GeoJSON
  Centroid Latitude               Centroid latitude
  Centroid Longitude              Centroid longitude
  Address                         Full address
  Address Formatted               Formatted address
  Address Time Zone ID            Time zone identifier
  Address City                    City
  Address State Code              State code
  Address County                  County
  Address Country Code            Country code
  Address Latitude                Address latitude
  Address Longitude               Address longitude
  Address Place ID                Place ID
  Address Plus Code               Plus code
  Skip Address Geocoding          Skip geocoding the address
  Project                         Project name or ID
  Project ID                      Project identifier
  Broker                          Broker name
  Broker ID                       Broker identifier
  Material Type IDs               Material type identifiers
  Project Bid Location Material Type IDs   Project bid location material type IDs

Arguments:
  <id>    The project bid location ID (required). Use the list command to find IDs.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project bid location
  xbe view project-bid-locations show 123

  # JSON output
  xbe view project-bid-locations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectBidLocationsShow,
	}
	initProjectBidLocationsShowFlags(cmd)
	return cmd
}

func init() {
	projectBidLocationsCmd.AddCommand(newProjectBidLocationsShowCmd())
}

func initProjectBidLocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectBidLocationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectBidLocationsShowOptions(cmd)
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
		return fmt.Errorf("project bid location id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-bid-locations]", "name,notes,geometry,geometry-geojson,centroid-latitude,centroid-longitude,address,address-formatted,address-time-zone-id,address-city,address-state-code,address-county,address-country-code,address-latitude,address-longitude,address-place-id,address-plus-code,skip-address-geocoding,project,broker,material-types,project-bid-location-material-types")
	query.Set("fields[projects]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "project,broker")

	body, _, err := client.Get(cmd.Context(), "/v1/project-bid-locations/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectBidLocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectBidLocationDetails(cmd, details)
}

func parseProjectBidLocationsShowOptions(cmd *cobra.Command) (projectBidLocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectBidLocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectBidLocationDetails(resp jsonAPISingleResponse) projectBidLocationDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := projectBidLocationDetails{
		ID:                                resp.Data.ID,
		Name:                              stringAttr(attrs, "name"),
		Notes:                             stringAttr(attrs, "notes"),
		Geometry:                          stringAttr(attrs, "geometry"),
		GeometryGeoJSON:                   stringAttr(attrs, "geometry-geojson"),
		CentroidLatitude:                  stringAttr(attrs, "centroid-latitude"),
		CentroidLongitude:                 stringAttr(attrs, "centroid-longitude"),
		Address:                           stringAttr(attrs, "address"),
		AddressFormatted:                  stringAttr(attrs, "address-formatted"),
		AddressTimeZoneID:                 stringAttr(attrs, "address-time-zone-id"),
		AddressCity:                       stringAttr(attrs, "address-city"),
		AddressStateCode:                  stringAttr(attrs, "address-state-code"),
		AddressCounty:                     stringAttr(attrs, "address-county"),
		AddressCountryCode:                stringAttr(attrs, "address-country-code"),
		AddressLatitude:                   stringAttr(attrs, "address-latitude"),
		AddressLongitude:                  stringAttr(attrs, "address-longitude"),
		AddressPlaceID:                    stringAttr(attrs, "address-place-id"),
		AddressPlusCode:                   stringAttr(attrs, "address-plus-code"),
		SkipAddressGeocoding:              boolAttr(attrs, "skip-address-geocoding"),
		MaterialTypeIDs:                   nil,
		ProjectBidLocationMaterialTypeIDs: nil,
	}

	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if project, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = stringAttr(project.Attributes, "name")
		}
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}
	if rel, ok := resp.Data.Relationships["material-types"]; ok {
		details.MaterialTypeIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resp.Data.Relationships["project-bid-location-material-types"]; ok {
		details.ProjectBidLocationMaterialTypeIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectBidLocationDetails(cmd *cobra.Command, details projectBidLocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	if details.Geometry != "" {
		fmt.Fprintf(out, "Geometry: %s\n", details.Geometry)
	}
	if details.GeometryGeoJSON != "" {
		fmt.Fprintf(out, "Geometry GeoJSON: %s\n", details.GeometryGeoJSON)
	}
	if details.CentroidLatitude != "" {
		fmt.Fprintf(out, "Centroid Latitude: %s\n", details.CentroidLatitude)
	}
	if details.CentroidLongitude != "" {
		fmt.Fprintf(out, "Centroid Longitude: %s\n", details.CentroidLongitude)
	}
	if details.Address != "" {
		fmt.Fprintf(out, "Address: %s\n", details.Address)
	}
	if details.AddressFormatted != "" {
		fmt.Fprintf(out, "Address Formatted: %s\n", details.AddressFormatted)
	}
	if details.AddressTimeZoneID != "" {
		fmt.Fprintf(out, "Address Time Zone ID: %s\n", details.AddressTimeZoneID)
	}
	if details.AddressCity != "" {
		fmt.Fprintf(out, "Address City: %s\n", details.AddressCity)
	}
	if details.AddressStateCode != "" {
		fmt.Fprintf(out, "Address State Code: %s\n", details.AddressStateCode)
	}
	if details.AddressCounty != "" {
		fmt.Fprintf(out, "Address County: %s\n", details.AddressCounty)
	}
	if details.AddressCountryCode != "" {
		fmt.Fprintf(out, "Address Country Code: %s\n", details.AddressCountryCode)
	}
	if details.AddressLatitude != "" {
		fmt.Fprintf(out, "Address Latitude: %s\n", details.AddressLatitude)
	}
	if details.AddressLongitude != "" {
		fmt.Fprintf(out, "Address Longitude: %s\n", details.AddressLongitude)
	}
	if details.AddressPlaceID != "" {
		fmt.Fprintf(out, "Address Place ID: %s\n", details.AddressPlaceID)
	}
	if details.AddressPlusCode != "" {
		fmt.Fprintf(out, "Address Plus Code: %s\n", details.AddressPlusCode)
	}
	fmt.Fprintf(out, "Skip Address Geocoding: %s\n", formatBool(details.SkipAddressGeocoding))
	if details.ProjectName != "" {
		fmt.Fprintf(out, "Project: %s\n", details.ProjectName)
	}
	if details.ProjectID != "" {
		fmt.Fprintf(out, "Project ID: %s\n", details.ProjectID)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if len(details.MaterialTypeIDs) > 0 {
		fmt.Fprintf(out, "Material Type IDs: %s\n", strings.Join(details.MaterialTypeIDs, ", "))
	}
	if len(details.ProjectBidLocationMaterialTypeIDs) > 0 {
		fmt.Fprintf(out, "Project Bid Location Material Type IDs: %s\n", strings.Join(details.ProjectBidLocationMaterialTypeIDs, ", "))
	}

	return nil
}
