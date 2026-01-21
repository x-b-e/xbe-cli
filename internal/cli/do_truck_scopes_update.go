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

type doTruckScopesUpdateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ID                       string
	TrailerClassificationIDs []string
	AuthorizedStateCodes     []string
	Address                  string
	AddressLatitude          string
	AddressLongitude         string
	AddressPlaceID           string
	AddressPlusCode          string
	AddressProximityMeters   int
	SkipGeocoding            bool
}

func newDoTruckScopesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing truck scope",
		Long: `Update an existing truck scope.

Provide the scope ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --trailer-classification-ids  Trailer classification IDs (comma-separated or repeated)
  --authorized-state-codes      Authorized state codes (comma-separated or repeated)
  --address                     Street address
  --address-latitude            Address latitude
  --address-longitude           Address longitude
  --address-place-id            Google Place ID
  --address-plus-code           Plus code
  --address-proximity-meters    Address proximity in meters
  --skip-geocoding              Skip geocoding the address`,
		Example: `  # Update authorized state codes
  xbe do truck-scopes update 123 --authorized-state-codes "IL,IN,WI,MI"

  # Update address proximity
  xbe do truck-scopes update 123 --address-proximity-meters 75000`,
		Args: cobra.ExactArgs(1),
		RunE: runDoTruckScopesUpdate,
	}
	initDoTruckScopesUpdateFlags(cmd)
	return cmd
}

func init() {
	doTruckScopesCmd.AddCommand(newDoTruckScopesUpdateCmd())
}

func initDoTruckScopesUpdateFlags(cmd *cobra.Command) {
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
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoTruckScopesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoTruckScopesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("trailer-classification-ids") {
		attributes["trailer-classification-ids"] = opts.TrailerClassificationIDs
	}
	if cmd.Flags().Changed("authorized-state-codes") {
		attributes["authorized-state-codes"] = opts.AuthorizedStateCodes
	}
	if cmd.Flags().Changed("address") {
		attributes["address"] = opts.Address
	}
	if cmd.Flags().Changed("address-latitude") {
		attributes["address-latitude"] = opts.AddressLatitude
	}
	if cmd.Flags().Changed("address-longitude") {
		attributes["address-longitude"] = opts.AddressLongitude
	}
	if cmd.Flags().Changed("address-place-id") {
		attributes["address-place-id"] = opts.AddressPlaceID
	}
	if cmd.Flags().Changed("address-plus-code") {
		attributes["address-plus-code"] = opts.AddressPlusCode
	}
	if cmd.Flags().Changed("address-proximity-meters") {
		attributes["address-proximity-meters"] = opts.AddressProximityMeters
	}
	if cmd.Flags().Changed("skip-geocoding") {
		attributes["skip-geocoding"] = opts.SkipGeocoding
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "truck-scopes",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/truck-scopes/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated truck scope %s\n", row.ID)
	return nil
}

func parseDoTruckScopesUpdateOptions(cmd *cobra.Command, args []string) (doTruckScopesUpdateOptions, error) {
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
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doTruckScopesUpdateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ID:                       args[0],
		TrailerClassificationIDs: trailerClassificationIDs,
		AuthorizedStateCodes:     authorizedStateCodes,
		Address:                  address,
		AddressLatitude:          addressLatitude,
		AddressLongitude:         addressLongitude,
		AddressPlaceID:           addressPlaceID,
		AddressPlusCode:          addressPlusCode,
		AddressProximityMeters:   addressProximityMeters,
		SkipGeocoding:            skipGeocoding,
	}, nil
}
