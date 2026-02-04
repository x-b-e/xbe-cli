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

type truckerApplicationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type truckerApplicationDetails struct {
	ID                          string `json:"id"`
	CompanyName                 string `json:"company_name,omitempty"`
	Status                      string `json:"status,omitempty"`
	CompanyAddress              string `json:"company_address,omitempty"`
	CompanyAddressFormatted     string `json:"company_address_formatted,omitempty"`
	CompanyAddressLatitude      string `json:"company_address_latitude,omitempty"`
	CompanyAddressLongitude     string `json:"company_address_longitude,omitempty"`
	CompanyAddressGeocoded      bool   `json:"company_address_geocoded"`
	IsCompanyAddressFormatted   bool   `json:"is_company_address_formatted_address"`
	CompanyAddressPlaceID       string `json:"company_address_place_id,omitempty"`
	CompanyAddressPlusCode      string `json:"company_address_plus_code,omitempty"`
	SkipCompanyAddressGeocoding bool   `json:"skip_company_address_geocoding"`
	HasUnionDrivers             bool   `json:"has_union_drivers"`
	EstimatedTrailerCapacity    string `json:"estimated_trailer_capacity,omitempty"`
	Notes                       string `json:"notes,omitempty"`
	ReferralCode                string `json:"referral_code,omitempty"`
	ViaDumptruckloadsdotcom     *bool  `json:"via_dumptruckloadsdotcom,omitempty"`
	DistanceFromSearchMiles     any    `json:"distance_from_search_miles,omitempty"`
	BrokerID                    string `json:"broker_id,omitempty"`
	BrokerName                  string `json:"broker_name,omitempty"`
	UserID                      string `json:"user_id,omitempty"`
	UserName                    string `json:"user_name,omitempty"`
	UserEmail                   string `json:"user_email,omitempty"`
	TruckerID                   string `json:"trucker_id,omitempty"`
	TruckerName                 string `json:"trucker_name,omitempty"`
	CreatedAt                   string `json:"created_at,omitempty"`
	UpdatedAt                   string `json:"updated_at,omitempty"`
}

func newTruckerApplicationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show trucker application details",
		Long: `Show the full details of a trucker application.

Output Fields:
  ID           Application identifier
  Company      Company name and address details
  Status       Application status
  Broker       Broker name or ID
  User         User name or email
  Trucker      Trucker name or ID (if approved)
  Metadata     Notes, referral code, union drivers, trailer capacity

Arguments:
  <id>  The trucker application ID (required).`,
		Example: `  # Show trucker application details
  xbe view trucker-applications show 123

  # Output as JSON
  xbe view trucker-applications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTruckerApplicationsShow,
	}
	initTruckerApplicationsShowFlags(cmd)
	return cmd
}

func init() {
	truckerApplicationsCmd.AddCommand(newTruckerApplicationsShowCmd())
}

func initTruckerApplicationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerApplicationsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTruckerApplicationsShowOptions(cmd)
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
		return fmt.Errorf("trucker application id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[trucker-applications]", "company-name,company-address,company-address-place-id,company-address-plus-code,skip-company-address-geocoding,company-address-geocoded,company-address-latitude,company-address-longitude,company-address-formatted,is-company-address-formatted-address,has-union-drivers,estimated-trailer-capacity,notes,referral-code,via-dumptruckloadsdotcom,status,created-at,updated-at,broker,user,trucker,distance-from-search-miles")
	query.Set("include", "broker,user,trucker")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-applications/"+id, query)
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

	details := buildTruckerApplicationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTruckerApplicationDetails(cmd, details)
}

func parseTruckerApplicationsShowOptions(cmd *cobra.Command) (truckerApplicationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerApplicationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTruckerApplicationDetails(resp jsonAPISingleResponse) truckerApplicationDetails {
	row := truckerApplicationRowFromSingle(resp)
	attrs := resp.Data.Attributes

	details := truckerApplicationDetails{
		ID:                          row.ID,
		CompanyName:                 row.CompanyName,
		Status:                      row.Status,
		CompanyAddress:              stringAttr(attrs, "company-address"),
		CompanyAddressFormatted:     stringAttr(attrs, "company-address-formatted"),
		CompanyAddressLatitude:      stringAttr(attrs, "company-address-latitude"),
		CompanyAddressLongitude:     stringAttr(attrs, "company-address-longitude"),
		CompanyAddressGeocoded:      boolAttr(attrs, "company-address-geocoded"),
		IsCompanyAddressFormatted:   boolAttr(attrs, "is-company-address-formatted-address"),
		CompanyAddressPlaceID:       stringAttr(attrs, "company-address-place-id"),
		CompanyAddressPlusCode:      stringAttr(attrs, "company-address-plus-code"),
		SkipCompanyAddressGeocoding: boolAttr(attrs, "skip-company-address-geocoding"),
		HasUnionDrivers:             boolAttr(attrs, "has-union-drivers"),
		EstimatedTrailerCapacity:    stringAttr(attrs, "estimated-trailer-capacity"),
		Notes:                       stringAttr(attrs, "notes"),
		ReferralCode:                stringAttr(attrs, "referral-code"),
		DistanceFromSearchMiles:     row.DistanceFromSearchMiles,
		BrokerID:                    row.BrokerID,
		BrokerName:                  row.BrokerName,
		UserID:                      row.UserID,
		UserName:                    row.UserName,
		UserEmail:                   row.UserEmail,
		TruckerID:                   row.TruckerID,
		TruckerName:                 row.TruckerName,
		CreatedAt:                   stringAttr(attrs, "created-at"),
		UpdatedAt:                   stringAttr(attrs, "updated-at"),
	}

	if val, ok := attrs["via-dumptruckloadsdotcom"]; ok && val != nil {
		via := boolAttr(attrs, "via-dumptruckloadsdotcom")
		details.ViaDumptruckloadsdotcom = &via
	}

	if val, ok := attrs["distance-from-search-miles"]; ok && val != nil {
		details.DistanceFromSearchMiles = val
	}

	return details
}

func renderTruckerApplicationDetails(cmd *cobra.Command, details truckerApplicationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CompanyName != "" {
		fmt.Fprintf(out, "Company Name: %s\n", details.CompanyName)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}

	if details.BrokerID != "" || details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", formatRelated(details.BrokerName, details.BrokerID))
	}

	userLabel := details.UserName
	if userLabel == "" {
		userLabel = details.UserEmail
	}
	if details.UserID != "" || userLabel != "" {
		fmt.Fprintf(out, "User: %s\n", formatRelated(userLabel, details.UserID))
	}

	if details.TruckerID != "" || details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker: %s\n", formatRelated(details.TruckerName, details.TruckerID))
	}

	if details.CompanyAddress != "" {
		fmt.Fprintf(out, "Company Address: %s\n", details.CompanyAddress)
	}
	if details.CompanyAddressFormatted != "" {
		fmt.Fprintf(out, "Company Address (formatted): %s\n", details.CompanyAddressFormatted)
	}
	if details.CompanyAddressLatitude != "" || details.CompanyAddressLongitude != "" {
		fmt.Fprintf(out, "Company Address Lat/Lng: %s, %s\n", details.CompanyAddressLatitude, details.CompanyAddressLongitude)
	}
	fmt.Fprintf(out, "Company Address Geocoded: %t\n", details.CompanyAddressGeocoded)
	fmt.Fprintf(out, "Company Address Formatted Flag: %t\n", details.IsCompanyAddressFormatted)

	if details.CompanyAddressPlaceID != "" {
		fmt.Fprintf(out, "Company Address Place ID: %s\n", details.CompanyAddressPlaceID)
	}
	if details.CompanyAddressPlusCode != "" {
		fmt.Fprintf(out, "Company Address Plus Code: %s\n", details.CompanyAddressPlusCode)
	}
	fmt.Fprintf(out, "Skip Company Address Geocoding: %t\n", details.SkipCompanyAddressGeocoding)
	fmt.Fprintf(out, "Has Union Drivers: %t\n", details.HasUnionDrivers)

	if details.EstimatedTrailerCapacity != "" {
		fmt.Fprintf(out, "Estimated Trailer Capacity: %s\n", details.EstimatedTrailerCapacity)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	if details.ReferralCode != "" {
		fmt.Fprintf(out, "Referral Code: %s\n", details.ReferralCode)
	}
	if details.ViaDumptruckloadsdotcom != nil {
		fmt.Fprintf(out, "Via Dumptruckloadsdotcom: %t\n", *details.ViaDumptruckloadsdotcom)
	}
	if details.DistanceFromSearchMiles != nil {
		fmt.Fprintf(out, "Distance From Search (mi): %s\n", formatDistanceMiles(details.DistanceFromSearchMiles))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
