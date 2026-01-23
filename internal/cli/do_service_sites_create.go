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

type doServiceSitesCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	Name             string
	BrokerID         string
	Address          string
	AddressLatitude  string
	AddressLongitude string
	SkipGeocoding    bool
}

func newDoServiceSitesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new service site",
		Long: `Create a new service site.

Required flags:
  --name      Service site name
  --broker    Broker ID (required)
  --address   Full address

Optional flags:
  --address-latitude    Address latitude
  --address-longitude   Address longitude
  --skip-geocoding      Skip geocoding the address`,
		Example: `  # Create a service site
  xbe do service-sites create --name "North Yard" --broker 123 --address "100 Main St, Chicago, IL"

  # Create with coordinates
  xbe do service-sites create --name "South Yard" --broker 123 --address "200 Main St, Chicago, IL" \
    --address-latitude "41.8781" --address-longitude "-87.6298" --skip-geocoding`,
		RunE: runDoServiceSitesCreate,
	}
	initDoServiceSitesCreateFlags(cmd)
	return cmd
}

func init() {
	doServiceSitesCmd.AddCommand(newDoServiceSitesCreateCmd())
}

func initDoServiceSitesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Service site name (required)")
	cmd.Flags().String("broker", "", "Broker ID (required)")
	cmd.Flags().String("address", "", "Full address (required)")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().Bool("skip-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("broker")
	_ = cmd.MarkFlagRequired("address")
}

func runDoServiceSitesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoServiceSitesCreateOptions(cmd)
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
		"name":    opts.Name,
		"address": opts.Address,
	}

	if opts.AddressLatitude != "" {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if opts.AddressLongitude != "" {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if opts.SkipGeocoding {
		attributes["skip-address-geocoding"] = true
	}

	relationships := map[string]any{
		"broker": map[string]any{
			"data": map[string]any{
				"type": "brokers",
				"id":   opts.BrokerID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "service-sites",
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

	body, _, err := client.Post(cmd.Context(), "/v1/service-sites", jsonBody)
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

	row := serviceSiteRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created service site %s (%s)\n", row.ID, row.Name)
	return nil
}

func parseDoServiceSitesCreateOptions(cmd *cobra.Command) (doServiceSitesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	brokerID, _ := cmd.Flags().GetString("broker")
	address, _ := cmd.Flags().GetString("address")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doServiceSitesCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		Name:             name,
		BrokerID:         brokerID,
		Address:          address,
		AddressLatitude:  addressLatitude,
		AddressLongitude: addressLongitude,
		SkipGeocoding:    skipGeocoding,
	}, nil
}

func serviceSiteRowFromSingle(resp jsonAPISingleResponse) serviceSiteRow {
	row := serviceSiteRow{
		ID:      resp.Data.ID,
		Name:    stringAttr(resp.Data.Attributes, "name"),
		Address: stringAttr(resp.Data.Attributes, "address"),
	}
	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	return row
}
