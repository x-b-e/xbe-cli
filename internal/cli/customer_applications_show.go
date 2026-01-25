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

type customerApplicationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type customerApplicationDetails struct {
	ID                                    string   `json:"id"`
	Status                                string   `json:"status,omitempty"`
	CompanyName                           string   `json:"company_name,omitempty"`
	CompanyURL                            string   `json:"company_url,omitempty"`
	Notes                                 string   `json:"notes,omitempty"`
	RequiresUnionDrivers                  bool     `json:"requires_union_drivers"`
	IsTruckingCompany                     bool     `json:"is_trucking_company"`
	EstimatedAnnualMaterialTransportSpend string   `json:"estimated_annual_material_transport_spend,omitempty"`
	CompanyAddress                        string   `json:"company_address,omitempty"`
	CompanyAddressLatitude                string   `json:"company_address_latitude,omitempty"`
	CompanyAddressLongitude               string   `json:"company_address_longitude,omitempty"`
	CompanyAddressPlaceID                 string   `json:"company_address_place_id,omitempty"`
	CompanyAddressPlusCode                string   `json:"company_address_plus_code,omitempty"`
	CompanyAddressFormatted               string   `json:"company_address_formatted,omitempty"`
	CompanyAddressTimeZoneID              string   `json:"company_address_time_zone_id,omitempty"`
	CompanyAddressCity                    string   `json:"company_address_city,omitempty"`
	CompanyAddressStateCode               string   `json:"company_address_state_code,omitempty"`
	SkipCompanyAddressGeocoding           bool     `json:"skip_company_address_geocoding"`
	IsCompanyAddressFormattedAddress      bool     `json:"is_company_address_formatted_address"`
	BrokerID                              string   `json:"broker_id,omitempty"`
	BrokerName                            string   `json:"broker_name,omitempty"`
	UserID                                string   `json:"user_id,omitempty"`
	UserName                              string   `json:"user_name,omitempty"`
	UserEmail                             string   `json:"user_email,omitempty"`
	CustomerID                            string   `json:"customer_id,omitempty"`
	CustomerName                          string   `json:"customer_name,omitempty"`
	JobTypeIDs                            []string `json:"job_type_ids,omitempty"`
	JobTypeNames                          []string `json:"job_type_names,omitempty"`
}

func newCustomerApplicationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show customer application details",
		Long: `Show the full details of a customer application.

Output Fields:
  ID
  Status
  Company Name
  Company URL
  Notes
  Requires Union Drivers
  Is Trucking Company
  Estimated Annual Material Transport Spend
  Company Address
  Company Address Latitude
  Company Address Longitude
  Company Address Place ID
  Company Address Plus Code
  Company Address Formatted
  Company Address Time Zone ID
  Company Address City
  Company Address State Code
  Skip Company Address Geocoding
  Is Company Address Formatted Address
  Broker (name and ID)
  User (name, email, and ID)
  Customer (name and ID)
  Job Type Names
  Job Type IDs

Arguments:
  <id>    Customer application ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a customer application
  xbe view customer-applications show 123

  # JSON output
  xbe view customer-applications show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCustomerApplicationsShow,
	}
	initCustomerApplicationsShowFlags(cmd)
	return cmd
}

func init() {
	customerApplicationsCmd.AddCommand(newCustomerApplicationsShowCmd())
}

func initCustomerApplicationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerApplicationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCustomerApplicationsShowOptions(cmd)
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
		return fmt.Errorf("customer application id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[customer-applications]", "status,company-name,company-url,notes,requires-union-drivers,is-trucking-company,estimated-annual-material-transport-spend,company-address,company-address-latitude,company-address-longitude,company-address-place-id,company-address-plus-code,company-address-formatted,company-address-time-zone-id,company-address-city,company-address-state-code,skip-company-address-geocoding,is-company-address-formatted-address,broker,user,customer,job-types")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[job-types]", "name")
	query.Set("include", "broker,user,customer,job-types")

	body, _, err := client.Get(cmd.Context(), "/v1/customer-applications/"+id, query)
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

	details := buildCustomerApplicationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCustomerApplicationDetails(cmd, details)
}

func parseCustomerApplicationsShowOptions(cmd *cobra.Command) (customerApplicationsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return customerApplicationsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return customerApplicationsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return customerApplicationsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return customerApplicationsShowOptions{}, err
	}

	return customerApplicationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCustomerApplicationDetails(resp jsonAPISingleResponse) customerApplicationDetails {
	attrs := resp.Data.Attributes
	details := customerApplicationDetails{
		ID:                                    resp.Data.ID,
		Status:                                stringAttr(attrs, "status"),
		CompanyName:                           strings.TrimSpace(stringAttr(attrs, "company-name")),
		CompanyURL:                            strings.TrimSpace(stringAttr(attrs, "company-url")),
		Notes:                                 strings.TrimSpace(stringAttr(attrs, "notes")),
		RequiresUnionDrivers:                  boolAttr(attrs, "requires-union-drivers"),
		IsTruckingCompany:                     boolAttr(attrs, "is-trucking-company"),
		EstimatedAnnualMaterialTransportSpend: stringAttr(attrs, "estimated-annual-material-transport-spend"),
		CompanyAddress:                        strings.TrimSpace(stringAttr(attrs, "company-address")),
		CompanyAddressLatitude:                stringAttr(attrs, "company-address-latitude"),
		CompanyAddressLongitude:               stringAttr(attrs, "company-address-longitude"),
		CompanyAddressPlaceID:                 stringAttr(attrs, "company-address-place-id"),
		CompanyAddressPlusCode:                stringAttr(attrs, "company-address-plus-code"),
		CompanyAddressFormatted:               stringAttr(attrs, "company-address-formatted"),
		CompanyAddressTimeZoneID:              stringAttr(attrs, "company-address-time-zone-id"),
		CompanyAddressCity:                    stringAttr(attrs, "company-address-city"),
		CompanyAddressStateCode:               stringAttr(attrs, "company-address-state-code"),
		SkipCompanyAddressGeocoding:           boolAttr(attrs, "skip-company-address-geocoding"),
		IsCompanyAddressFormattedAddress:      boolAttr(attrs, "is-company-address-formatted-address"),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			details.UserEmail = strings.TrimSpace(stringAttr(user.Attributes, "email-address"))
		}
	}

	if rel, ok := resp.Data.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CustomerName = strings.TrimSpace(stringAttr(customer.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["job-types"]; ok {
		details.JobTypeIDs = relationshipIDStrings(rel)
		for _, ref := range relationshipIDs(rel) {
			if ref.ID == "" {
				continue
			}
			if jobType, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
				name := strings.TrimSpace(stringAttr(jobType.Attributes, "name"))
				if name != "" {
					details.JobTypeNames = append(details.JobTypeNames, name)
				}
			}
		}
	}

	return details
}

func renderCustomerApplicationDetails(cmd *cobra.Command, details customerApplicationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CompanyName != "" {
		fmt.Fprintf(out, "Company Name: %s\n", details.CompanyName)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.UserName != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserName)
	}
	if details.UserEmail != "" {
		fmt.Fprintf(out, "User Email: %s\n", details.UserEmail)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.CustomerName != "" {
		fmt.Fprintf(out, "Customer: %s\n", details.CustomerName)
	}
	if details.CustomerID != "" {
		fmt.Fprintf(out, "Customer ID: %s\n", details.CustomerID)
	}
	if len(details.JobTypeNames) > 0 {
		fmt.Fprintf(out, "Job Types: %s\n", strings.Join(details.JobTypeNames, ", "))
	}
	if len(details.JobTypeIDs) > 0 {
		fmt.Fprintf(out, "Job Type IDs: %s\n", strings.Join(details.JobTypeIDs, ", "))
	}
	if details.CompanyURL != "" {
		fmt.Fprintf(out, "Company URL: %s\n", details.CompanyURL)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	fmt.Fprintf(out, "Requires Union Drivers: %s\n", formatBool(details.RequiresUnionDrivers))
	fmt.Fprintf(out, "Is Trucking Company: %s\n", formatBool(details.IsTruckingCompany))
	if details.EstimatedAnnualMaterialTransportSpend != "" {
		fmt.Fprintf(out, "Estimated Annual Material Transport Spend: %s\n", details.EstimatedAnnualMaterialTransportSpend)
	}
	if details.CompanyAddress != "" {
		fmt.Fprintf(out, "Company Address: %s\n", details.CompanyAddress)
	}
	if details.CompanyAddressLatitude != "" {
		fmt.Fprintf(out, "Company Address Latitude: %s\n", details.CompanyAddressLatitude)
	}
	if details.CompanyAddressLongitude != "" {
		fmt.Fprintf(out, "Company Address Longitude: %s\n", details.CompanyAddressLongitude)
	}
	if details.CompanyAddressPlaceID != "" {
		fmt.Fprintf(out, "Company Address Place ID: %s\n", details.CompanyAddressPlaceID)
	}
	if details.CompanyAddressPlusCode != "" {
		fmt.Fprintf(out, "Company Address Plus Code: %s\n", details.CompanyAddressPlusCode)
	}
	if details.CompanyAddressFormatted != "" {
		fmt.Fprintf(out, "Company Address Formatted: %s\n", details.CompanyAddressFormatted)
	}
	if details.CompanyAddressTimeZoneID != "" {
		fmt.Fprintf(out, "Company Address Time Zone ID: %s\n", details.CompanyAddressTimeZoneID)
	}
	if details.CompanyAddressCity != "" {
		fmt.Fprintf(out, "Company Address City: %s\n", details.CompanyAddressCity)
	}
	if details.CompanyAddressStateCode != "" {
		fmt.Fprintf(out, "Company Address State Code: %s\n", details.CompanyAddressStateCode)
	}
	fmt.Fprintf(out, "Skip Company Address Geocoding: %s\n", formatBool(details.SkipCompanyAddressGeocoding))
	fmt.Fprintf(out, "Is Company Address Formatted Address: %s\n", formatBool(details.IsCompanyAddressFormattedAddress))

	return nil
}
