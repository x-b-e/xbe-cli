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

type projectTransportLocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportLocationDetails struct {
	ID                                   string   `json:"id"`
	Name                                 string   `json:"name,omitempty"`
	ExternalTmsCompanyID                 string   `json:"external_tms_company_id,omitempty"`
	GeocodingMethod                      string   `json:"geocoding_method,omitempty"`
	AddressLatitude                      string   `json:"address_latitude,omitempty"`
	AddressLongitude                     string   `json:"address_longitude,omitempty"`
	AddressFull                          string   `json:"address_full,omitempty"`
	AddressStreetOne                     string   `json:"address_street_one,omitempty"`
	AddressStreetTwo                     string   `json:"address_street_two,omitempty"`
	AddressCity                          string   `json:"address_city,omitempty"`
	AddressStateCode                     string   `json:"address_state_code,omitempty"`
	AddressCountryCode                   string   `json:"address_country_code,omitempty"`
	AddressPostalCode                    string   `json:"address_postal_code,omitempty"`
	AddressTimeZoneID                    string   `json:"address_time_zone_id,omitempty"`
	AddressSPLC                          string   `json:"address_splc,omitempty"`
	IsActive                             bool     `json:"is_active"`
	IsValidForStop                       bool     `json:"is_valid_for_stop"`
	SkipDetection                        bool     `json:"skip_detection"`
	DistanceInMiles                      any      `json:"distance_in_miles,omitempty"`
	ActivityRadiusM                      any      `json:"activity_radius_m,omitempty"`
	IsManaged                            bool     `json:"is_managed"`
	BrokerID                             string   `json:"broker_id,omitempty"`
	BrokerName                           string   `json:"broker_name,omitempty"`
	ProjectTransportOrganizationID       string   `json:"project_transport_organization_id,omitempty"`
	ProjectTransportOrganizationName     string   `json:"project_transport_organization_name,omitempty"`
	NearestProjectOfficeCachedID         string   `json:"nearest_project_office_cached_id,omitempty"`
	NearestProjectOfficeCachedName       string   `json:"nearest_project_office_cached_name,omitempty"`
	UpdatedByID                          string   `json:"updated_by_id,omitempty"`
	UpdatedByName                        string   `json:"updated_by_name,omitempty"`
	ProjectTransportLocationEventTypeIDs []string `json:"project_transport_location_event_type_ids,omitempty"`
	ProjectTransportPlanStopIDs          []string `json:"project_transport_plan_stop_ids,omitempty"`
	PickupProjectMaterialTypeIDs         []string `json:"pickup_project_material_type_ids,omitempty"`
	DeliveryProjectMaterialTypeIDs       []string `json:"delivery_project_material_type_ids,omitempty"`
	ExternalIdentificationIDs            []string `json:"external_identification_ids,omitempty"`
}

func newProjectTransportLocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport location details",
		Long: `Show the full details of a project transport location.

Includes address details, geocoding metadata, and related entities.

Arguments:
  <id>  The project transport location ID (required).`,
		Example: `  # Show a project transport location
  xbe view project-transport-locations show 123

  # Output as JSON
  xbe view project-transport-locations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportLocationsShow,
	}
	initProjectTransportLocationsShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportLocationsCmd.AddCommand(newProjectTransportLocationsShowCmd())
}

func initProjectTransportLocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportLocationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectTransportLocationsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport location id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-transport-locations]", "name,external-tms-company-id,geocoding-method,address-latitude,address-longitude,address-full,address-street-one,address-street-two,address-city,address-state-code,address-country-code,address-postal-code,address-time-zone-id,address-splc,is-active,is-valid-for-stop,skip-detection,distance-in-miles,activity-radius-m,is-managed,broker,project-transport-organization,nearest-project-office-cached,updated-by,project-transport-location-event-types,project-transport-plan-stops,pickup-project-material-types,delivery-project-material-types,external-identifications")
	query.Set("include", "broker,project-transport-organization,nearest-project-office-cached,updated-by")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[project-transport-organizations]", "name")
	query.Set("fields[project-offices]", "name,abbreviation")
	query.Set("fields[users]", "name")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-locations/"+id, query)
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

	details := buildProjectTransportLocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportLocationDetails(cmd, details)
}

func parseProjectTransportLocationsShowOptions(cmd *cobra.Command) (projectTransportLocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportLocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportLocationDetails(resp jsonAPISingleResponse) projectTransportLocationDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := projectTransportLocationDetails{
		ID:                   resp.Data.ID,
		Name:                 stringAttr(attrs, "name"),
		ExternalTmsCompanyID: stringAttr(attrs, "external-tms-company-id"),
		GeocodingMethod:      stringAttr(attrs, "geocoding-method"),
		AddressLatitude:      stringAttr(attrs, "address-latitude"),
		AddressLongitude:     stringAttr(attrs, "address-longitude"),
		AddressFull:          stringAttr(attrs, "address-full"),
		AddressStreetOne:     stringAttr(attrs, "address-street-one"),
		AddressStreetTwo:     stringAttr(attrs, "address-street-two"),
		AddressCity:          stringAttr(attrs, "address-city"),
		AddressStateCode:     stringAttr(attrs, "address-state-code"),
		AddressCountryCode:   stringAttr(attrs, "address-country-code"),
		AddressPostalCode:    stringAttr(attrs, "address-postal-code"),
		AddressTimeZoneID:    stringAttr(attrs, "address-time-zone-id"),
		AddressSPLC:          stringAttr(attrs, "address-splc"),
		IsActive:             boolAttr(attrs, "is-active"),
		IsValidForStop:       boolAttr(attrs, "is-valid-for-stop"),
		SkipDetection:        boolAttr(attrs, "skip-detection"),
		DistanceInMiles:      attrs["distance-in-miles"],
		ActivityRadiusM:      attrs["activity-radius-m"],
		IsManaged:            boolAttr(attrs, "is-managed"),
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resp.Data.Relationships["project-transport-organization"]; ok && rel.Data != nil {
		details.ProjectTransportOrganizationID = rel.Data.ID
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectTransportOrganizationName = stringAttr(org.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["nearest-project-office-cached"]; ok && rel.Data != nil {
		details.NearestProjectOfficeCachedID = rel.Data.ID
		if office, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.NearestProjectOfficeCachedName = stringAttr(office.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["updated-by"]; ok && rel.Data != nil {
		details.UpdatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UpdatedByName = stringAttr(user.Attributes, "name")
		}
	}

	if rel, ok := resp.Data.Relationships["project-transport-location-event-types"]; ok && rel.raw != nil {
		details.ProjectTransportLocationEventTypeIDs = extractRelationshipIDs(rel)
	}
	if rel, ok := resp.Data.Relationships["project-transport-plan-stops"]; ok && rel.raw != nil {
		details.ProjectTransportPlanStopIDs = extractRelationshipIDs(rel)
	}
	if rel, ok := resp.Data.Relationships["pickup-project-material-types"]; ok && rel.raw != nil {
		details.PickupProjectMaterialTypeIDs = extractRelationshipIDs(rel)
	}
	if rel, ok := resp.Data.Relationships["delivery-project-material-types"]; ok && rel.raw != nil {
		details.DeliveryProjectMaterialTypeIDs = extractRelationshipIDs(rel)
	}
	if rel, ok := resp.Data.Relationships["external-identifications"]; ok && rel.raw != nil {
		details.ExternalIdentificationIDs = extractRelationshipIDs(rel)
	}

	return details
}

func renderProjectTransportLocationDetails(cmd *cobra.Command, d projectTransportLocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	if d.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", d.Name)
	}
	if d.ExternalTmsCompanyID != "" {
		fmt.Fprintf(out, "External TMS Company ID: %s\n", d.ExternalTmsCompanyID)
	}
	if d.GeocodingMethod != "" {
		fmt.Fprintf(out, "Geocoding Method: %s\n", d.GeocodingMethod)
	}

	if d.AddressFull != "" {
		fmt.Fprintf(out, "Address: %s\n", d.AddressFull)
	}
	if d.AddressStreetOne != "" {
		fmt.Fprintf(out, "Street 1: %s\n", d.AddressStreetOne)
	}
	if d.AddressStreetTwo != "" {
		fmt.Fprintf(out, "Street 2: %s\n", d.AddressStreetTwo)
	}
	if d.AddressCity != "" {
		fmt.Fprintf(out, "City: %s\n", d.AddressCity)
	}
	if d.AddressStateCode != "" {
		fmt.Fprintf(out, "State: %s\n", d.AddressStateCode)
	}
	if d.AddressPostalCode != "" {
		fmt.Fprintf(out, "Postal Code: %s\n", d.AddressPostalCode)
	}
	if d.AddressCountryCode != "" {
		fmt.Fprintf(out, "Country: %s\n", d.AddressCountryCode)
	}
	if d.AddressLatitude != "" || d.AddressLongitude != "" {
		fmt.Fprintf(out, "Coordinates: %s, %s\n", d.AddressLatitude, d.AddressLongitude)
	}
	if d.AddressTimeZoneID != "" {
		fmt.Fprintf(out, "Time Zone: %s\n", d.AddressTimeZoneID)
	}
	if d.AddressSPLC != "" {
		fmt.Fprintf(out, "SPLC: %s\n", d.AddressSPLC)
	}

	fmt.Fprintf(out, "Active: %t\n", d.IsActive)
	fmt.Fprintf(out, "Valid For Stop: %t\n", d.IsValidForStop)
	fmt.Fprintf(out, "Skip Detection: %t\n", d.SkipDetection)
	fmt.Fprintf(out, "Managed: %t\n", d.IsManaged)

	if d.DistanceInMiles != nil {
		fmt.Fprintf(out, "Distance (mi): %s\n", formatDistanceMiles(d.DistanceInMiles))
	}
	if d.ActivityRadiusM != nil {
		fmt.Fprintf(out, "Activity Radius (m): %v\n", d.ActivityRadiusM)
	}

	if d.BrokerID != "" || d.ProjectTransportOrganizationID != "" || d.NearestProjectOfficeCachedID != "" || d.UpdatedByID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Relationships:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.BrokerName != "" {
			fmt.Fprintf(out, "  Broker: %s (ID: %s)\n", d.BrokerName, d.BrokerID)
		} else if d.BrokerID != "" {
			fmt.Fprintf(out, "  Broker ID: %s\n", d.BrokerID)
		}
		if d.ProjectTransportOrganizationName != "" {
			fmt.Fprintf(out, "  Project Transport Organization: %s (ID: %s)\n", d.ProjectTransportOrganizationName, d.ProjectTransportOrganizationID)
		} else if d.ProjectTransportOrganizationID != "" {
			fmt.Fprintf(out, "  Project Transport Organization ID: %s\n", d.ProjectTransportOrganizationID)
		}
		if d.NearestProjectOfficeCachedName != "" {
			fmt.Fprintf(out, "  Nearest Project Office: %s (ID: %s)\n", d.NearestProjectOfficeCachedName, d.NearestProjectOfficeCachedID)
		} else if d.NearestProjectOfficeCachedID != "" {
			fmt.Fprintf(out, "  Nearest Project Office ID: %s\n", d.NearestProjectOfficeCachedID)
		}
		if d.UpdatedByName != "" {
			fmt.Fprintf(out, "  Updated By: %s (ID: %s)\n", d.UpdatedByName, d.UpdatedByID)
		} else if d.UpdatedByID != "" {
			fmt.Fprintf(out, "  Updated By ID: %s\n", d.UpdatedByID)
		}
	}

	if len(d.ProjectTransportLocationEventTypeIDs) > 0 || len(d.ProjectTransportPlanStopIDs) > 0 || len(d.PickupProjectMaterialTypeIDs) > 0 || len(d.DeliveryProjectMaterialTypeIDs) > 0 || len(d.ExternalIdentificationIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Related IDs:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if len(d.ProjectTransportLocationEventTypeIDs) > 0 {
			fmt.Fprintf(out, "  Event Type IDs: %s\n", strings.Join(d.ProjectTransportLocationEventTypeIDs, ", "))
		}
		if len(d.ProjectTransportPlanStopIDs) > 0 {
			fmt.Fprintf(out, "  Plan Stop IDs: %s\n", strings.Join(d.ProjectTransportPlanStopIDs, ", "))
		}
		if len(d.PickupProjectMaterialTypeIDs) > 0 {
			fmt.Fprintf(out, "  Pickup Material Type IDs: %s\n", strings.Join(d.PickupProjectMaterialTypeIDs, ", "))
		}
		if len(d.DeliveryProjectMaterialTypeIDs) > 0 {
			fmt.Fprintf(out, "  Delivery Material Type IDs: %s\n", strings.Join(d.DeliveryProjectMaterialTypeIDs, ", "))
		}
		if len(d.ExternalIdentificationIDs) > 0 {
			fmt.Fprintf(out, "  External Identification IDs: %s\n", strings.Join(d.ExternalIdentificationIDs, ", "))
		}
	}

	return nil
}
