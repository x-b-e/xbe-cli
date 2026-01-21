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

type doTruckScopesCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	TrailerClassificationIDs []string
	AuthorizedStateCodes     []string
	Address                  string
	AddressLatitude          string
	AddressLongitude         string
	AddressPlaceID           string
	AddressPlusCode          string
	AddressProximityMeters   int
	SkipGeocoding            bool
	OrganizationType         string
	OrganizationID           string
}

func newDoTruckScopesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new truck scope",
		Long: `Create a new truck scope.

Required flags:
  --organization-type  Organization type (e.g., brokers, truckers) (required)
  --organization-id    Organization ID (required)

Optional flags:
  --trailer-classification-ids  Trailer classification IDs (comma-separated or repeated)
  --authorized-state-codes      Authorized state codes (comma-separated or repeated)
  --address                     Street address
  --address-latitude            Address latitude
  --address-longitude           Address longitude
  --address-place-id            Google Place ID
  --address-plus-code           Plus code
  --address-proximity-meters    Address proximity in meters
  --skip-geocoding              Skip geocoding the address`,
		Example: `  # Create a truck scope with authorized states
  xbe do truck-scopes create --authorized-state-codes "IL,IN,WI" --organization-type brokers --organization-id 123

  # Create with address and proximity
  xbe do truck-scopes create --address "123 Main St" --address-proximity-meters 50000 --organization-type truckers --organization-id 456`,
		Args: cobra.NoArgs,
		RunE: runDoTruckScopesCreate,
	}
	initDoTruckScopesCreateFlags(cmd)
	return cmd
}

func init() {
	doTruckScopesCmd.AddCommand(newDoTruckScopesCreateCmd())
}

func initDoTruckScopesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().StringSlice("trailer-classification-ids", nil, "Trailer classification IDs (comma-separated or repeated)")
	cmd.Flags().StringSlice("authorized-state-codes", nil, "Authorized state codes (comma-separated or repeated)")
	cmd.Flags().String("address", "", "Street address")
	cmd.Flags().String("address-latitude", "", "Address latitude")
	cmd.Flags().String("address-longitude", "", "Address longitude")
	cmd.Flags().String("address-place-id", "", "Google Place ID")
	cmd.Flags().String("address-plus-code", "", "Plus code")
	cmd.Flags().Int("address-proximity-meters", 0, "Address proximity in meters")
	cmd.Flags().Bool("skip-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("organization-type", "", "Organization type (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckScopesCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckScopesCreateOptions(cmd)
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

	if opts.OrganizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.OrganizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if len(opts.TrailerClassificationIDs) > 0 {
		attributes["trailer-classification-ids"] = opts.TrailerClassificationIDs
	}
	if len(opts.AuthorizedStateCodes) > 0 {
		attributes["authorized-state-codes"] = opts.AuthorizedStateCodes
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
	if cmd.Flags().Changed("address-proximity-meters") {
		attributes["address-proximity-meters"] = opts.AddressProximityMeters
	}
	if opts.SkipGeocoding {
		attributes["skip-geocoding"] = true
	}

	relationships := map[string]any{
		"organization": map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "truck-scopes",
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

	body, _, err := client.Post(cmd.Context(), "/v1/truck-scopes", jsonBody)
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

	row := buildTruckScopeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created truck scope %s\n", row.ID)
	return nil
}

func parseDoTruckScopesCreateOptions(cmd *cobra.Command) (doTruckScopesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	trailerClassificationIDs, _ := cmd.Flags().GetStringSlice("trailer-classification-ids")
	authorizedStateCodes, _ := cmd.Flags().GetStringSlice("authorized-state-codes")
	address, _ := cmd.Flags().GetString("address")
	addressLatitude, _ := cmd.Flags().GetString("address-latitude")
	addressLongitude, _ := cmd.Flags().GetString("address-longitude")
	addressPlaceID, _ := cmd.Flags().GetString("address-place-id")
	addressPlusCode, _ := cmd.Flags().GetString("address-plus-code")
	addressProximityMeters, _ := cmd.Flags().GetInt("address-proximity-meters")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-geocoding")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckScopesCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		TrailerClassificationIDs: trailerClassificationIDs,
		AuthorizedStateCodes:     authorizedStateCodes,
		Address:                  address,
		AddressLatitude:          addressLatitude,
		AddressLongitude:         addressLongitude,
		AddressPlaceID:           addressPlaceID,
		AddressPlusCode:          addressPlusCode,
		AddressProximityMeters:   addressProximityMeters,
		SkipGeocoding:            skipGeocoding,
		OrganizationType:         organizationType,
		OrganizationID:           organizationID,
	}, nil
}

func buildTruckScopeRowFromSingle(resp jsonAPISingleResponse) truckScopeRow {
	attrs := resp.Data.Attributes

	row := truckScopeRow{
		ID:                     resp.Data.ID,
		Address:                stringAttr(attrs, "address"),
		AddressCity:            stringAttr(attrs, "address-city"),
		AddressStateCode:       stringAttr(attrs, "address-state-code"),
		AddressProximityMeters: attrs["address-proximity-meters"],
	}

	if ids, ok := attrs["trailer-classification-ids"].([]any); ok {
		for _, id := range ids {
			if s, ok := id.(string); ok {
				row.TrailerClassificationIDs = append(row.TrailerClassificationIDs, s)
			} else if n, ok := id.(float64); ok {
				row.TrailerClassificationIDs = append(row.TrailerClassificationIDs, fmt.Sprintf("%.0f", n))
			}
		}
	}

	if codes, ok := attrs["authorized-state-codes"].([]any); ok {
		for _, code := range codes {
			if s, ok := code.(string); ok {
				row.AuthorizedStateCodes = append(row.AuthorizedStateCodes, s)
			}
		}
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}
