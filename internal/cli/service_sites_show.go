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

type serviceSitesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type serviceSiteDetails struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Address                  string `json:"address"`
	AddressLatitude          string `json:"address_latitude,omitempty"`
	AddressLongitude         string `json:"address_longitude,omitempty"`
	IsAddressFormatted       bool   `json:"is_address_formatted_address"`
	AddressGeocoded          bool   `json:"address_geocoded"`
	AddressTimeZoneIDCached  string `json:"address_time_zone_id_cached,omitempty"`
	AddressStateCodeCached   string `json:"address_state_code_cached,omitempty"`
	AddressCountryCodeCached string `json:"address_country_code_cached,omitempty"`
	BrokerID                 string `json:"broker_id,omitempty"`
	BrokerName               string `json:"broker_name,omitempty"`
}

func newServiceSitesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show service site details",
		Long: `Show the full details of a specific service site.

Service sites are locations used for service work orders.

Output Fields:
  ID                      Service site identifier
  Name                    Service site name
  Broker                  Broker name
  Broker ID               Broker identifier
  Address                 Full address
  Address Latitude        Latitude coordinate
  Address Longitude       Longitude coordinate
  Is Address Formatted    Whether the address was formatted
  Address Geocoded        Whether the address has geocoding
  Address Time Zone ID    Cached time zone identifier
  Address State Code      Cached state code
  Address Country Code    Cached country code

Arguments:
  <id>    The service site ID (required). You can find IDs using the list command.`,
		Example: `  # View a service site
  xbe view service-sites show 123

  # Get JSON output
  xbe view service-sites show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runServiceSitesShow,
	}
	initServiceSitesShowFlags(cmd)
	return cmd
}

func init() {
	serviceSitesCmd.AddCommand(newServiceSitesShowCmd())
}

func initServiceSitesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runServiceSitesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseServiceSitesShowOptions(cmd)
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
		return fmt.Errorf("service site id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[service-sites]", "name,address,address-latitude,address-longitude,is-address-formatted-address,address-geocoded,address-time-zone-id-cached,address-state-code-cached,address-country-code-cached,broker")
	query.Set("fields[brokers]", "company-name")
	query.Set("include", "broker")

	body, _, err := client.Get(cmd.Context(), "/v1/service-sites/"+id, query)
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

	details := buildServiceSiteDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderServiceSiteDetails(cmd, details)
}

func parseServiceSitesShowOptions(cmd *cobra.Command) (serviceSitesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return serviceSitesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return serviceSitesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return serviceSitesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return serviceSitesShowOptions{}, err
	}

	return serviceSitesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildServiceSiteDetails(resp jsonAPISingleResponse) serviceSiteDetails {
	attrs := resp.Data.Attributes
	details := serviceSiteDetails{
		ID:                       resp.Data.ID,
		Name:                     stringAttr(attrs, "name"),
		Address:                  stringAttr(attrs, "address"),
		AddressLatitude:          stringAttr(attrs, "address-latitude"),
		AddressLongitude:         stringAttr(attrs, "address-longitude"),
		IsAddressFormatted:       boolAttr(attrs, "is-address-formatted-address"),
		AddressGeocoded:          boolAttr(attrs, "address-geocoded"),
		AddressTimeZoneIDCached:  stringAttr(attrs, "address-time-zone-id-cached"),
		AddressStateCodeCached:   stringAttr(attrs, "address-state-code-cached"),
		AddressCountryCodeCached: stringAttr(attrs, "address-country-code-cached"),
	}

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	return details
}

func renderServiceSiteDetails(cmd *cobra.Command, details serviceSiteDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.Address != "" {
		fmt.Fprintf(out, "Address: %s\n", details.Address)
	}
	if details.AddressLatitude != "" {
		fmt.Fprintf(out, "Address Latitude: %s\n", details.AddressLatitude)
	}
	if details.AddressLongitude != "" {
		fmt.Fprintf(out, "Address Longitude: %s\n", details.AddressLongitude)
	}
	fmt.Fprintf(out, "Is Address Formatted: %s\n", formatBool(details.IsAddressFormatted))
	fmt.Fprintf(out, "Address Geocoded: %s\n", formatBool(details.AddressGeocoded))
	if details.AddressTimeZoneIDCached != "" {
		fmt.Fprintf(out, "Address Time Zone ID: %s\n", details.AddressTimeZoneIDCached)
	}
	if details.AddressStateCodeCached != "" {
		fmt.Fprintf(out, "Address State Code: %s\n", details.AddressStateCodeCached)
	}
	if details.AddressCountryCodeCached != "" {
		fmt.Fprintf(out, "Address Country Code: %s\n", details.AddressCountryCodeCached)
	}

	return nil
}
