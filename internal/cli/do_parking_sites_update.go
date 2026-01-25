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

type doParkingSitesUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	IsActive      bool
	ActiveStart   string
	ActiveEnd     string
	Address       string
	Latitude      string
	Longitude     string
	SkipGeocoding bool
}

func newDoParkingSitesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a parking site",
		Long: `Update a parking site.

Optional:
  --is-active      Whether the site is active
  --active-start   Active start time (ISO 8601)
  --active-end     Active end time (ISO 8601)
  --address        Address of the parking site
  --latitude       Latitude coordinate
  --longitude      Longitude coordinate
  --skip-geocoding Skip geocoding the address`,
		Example: `  # Mark a parking site as active
  xbe do parking-sites update 123 --is-active

  # Update the address
  xbe do parking-sites update 123 --address "456 Oak Ave"

  # Set active end time
  xbe do parking-sites update 123 --active-end "2024-12-31T23:59:59Z"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoParkingSitesUpdate,
	}
	initDoParkingSitesUpdateFlags(cmd)
	return cmd
}

func init() {
	doParkingSitesCmd.AddCommand(newDoParkingSitesUpdateCmd())
}

func initDoParkingSitesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-active", false, "Whether the site is active")
	cmd.Flags().String("active-start", "", "Active start time (ISO 8601)")
	cmd.Flags().String("active-end", "", "Active end time (ISO 8601)")
	cmd.Flags().String("address", "", "Address of the parking site")
	cmd.Flags().String("latitude", "", "Latitude coordinate")
	cmd.Flags().String("longitude", "", "Longitude coordinate")
	cmd.Flags().Bool("skip-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoParkingSitesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoParkingSitesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}
	if cmd.Flags().Changed("active-start") {
		attributes["active-start-at"] = opts.ActiveStart
	}
	if cmd.Flags().Changed("active-end") {
		attributes["active-end-at"] = opts.ActiveEnd
	}
	if cmd.Flags().Changed("address") {
		attributes["address"] = opts.Address
	}
	if cmd.Flags().Changed("latitude") {
		attributes["address-latitude"] = opts.Latitude
	}
	if cmd.Flags().Changed("longitude") {
		attributes["address-longitude"] = opts.Longitude
	}
	if cmd.Flags().Changed("skip-geocoding") {
		attributes["skip-address-geocoding"] = opts.SkipGeocoding
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "parking-sites",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/parking-sites/"+opts.ID, jsonBody)
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

	if opts.JSON {
		row := parkingSiteRow{
			ID:            resp.Data.ID,
			IsActive:      boolAttr(resp.Data.Attributes, "is-active"),
			ActiveStartAt: stringAttr(resp.Data.Attributes, "active-start-at"),
			ActiveEndAt:   stringAttr(resp.Data.Attributes, "active-end-at"),
			Address:       stringAttr(resp.Data.Attributes, "address"),
		}
		if rel, ok := resp.Data.Relationships["parked"]; ok && rel.Data != nil {
			row.ParkedType = rel.Data.Type
			row.ParkedID = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated parking site %s\n", resp.Data.ID)
	return nil
}

func parseDoParkingSitesUpdateOptions(cmd *cobra.Command, args []string) (doParkingSitesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isActive, _ := cmd.Flags().GetBool("is-active")
	activeStart, _ := cmd.Flags().GetString("active-start")
	activeEnd, _ := cmd.Flags().GetString("active-end")
	address, _ := cmd.Flags().GetString("address")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doParkingSitesUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		IsActive:      isActive,
		ActiveStart:   activeStart,
		ActiveEnd:     activeEnd,
		Address:       address,
		Latitude:      latitude,
		Longitude:     longitude,
		SkipGeocoding: skipGeocoding,
	}, nil
}
