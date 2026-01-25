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

type doParkingSitesCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ParkedType    string
	ParkedID      string
	IsActive      bool
	ActiveStart   string
	ActiveEnd     string
	Address       string
	Latitude      string
	Longitude     string
	SkipGeocoding bool
}

func newDoParkingSitesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a parking site",
		Long: `Create a parking site.

Required:
  --parked-type    Type of parked item (e.g., trailers, tractors)
  --parked-id      ID of parked item

Optional:
  --is-active      Whether the site is active
  --active-start   Active start time (ISO 8601)
  --active-end     Active end time (ISO 8601)
  --address        Address of the parking site
  --latitude       Latitude coordinate
  --longitude      Longitude coordinate
  --skip-geocoding Skip geocoding the address`,
		Example: `  # Create a parking site for a trailer
  xbe do parking-sites create --parked-type trailers --parked-id 123

  # Create an active parking site with address
  xbe do parking-sites create --parked-type tractors --parked-id 456 --is-active --address "123 Main St"

  # Create with coordinates
  xbe do parking-sites create --parked-type trailers --parked-id 789 --latitude 40.7128 --longitude -74.0060`,
		RunE: runDoParkingSitesCreate,
	}
	initDoParkingSitesCreateFlags(cmd)
	return cmd
}

func init() {
	doParkingSitesCmd.AddCommand(newDoParkingSitesCreateCmd())
}

func initDoParkingSitesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("parked-type", "", "Type of parked item (e.g., trailers, tractors)")
	cmd.Flags().String("parked-id", "", "ID of parked item")
	cmd.Flags().Bool("is-active", false, "Whether the site is active")
	cmd.Flags().String("active-start", "", "Active start time (ISO 8601)")
	cmd.Flags().String("active-end", "", "Active end time (ISO 8601)")
	cmd.Flags().String("address", "", "Address of the parking site")
	cmd.Flags().String("latitude", "", "Latitude coordinate")
	cmd.Flags().String("longitude", "", "Longitude coordinate")
	cmd.Flags().Bool("skip-geocoding", false, "Skip geocoding the address")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("parked-type")
	_ = cmd.MarkFlagRequired("parked-id")
}

func runDoParkingSitesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoParkingSitesCreateOptions(cmd)
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
	if opts.ActiveStart != "" {
		attributes["active-start-at"] = opts.ActiveStart
	}
	if opts.ActiveEnd != "" {
		attributes["active-end-at"] = opts.ActiveEnd
	}
	if opts.Address != "" {
		attributes["address"] = opts.Address
	}
	if opts.Latitude != "" {
		attributes["address-latitude"] = opts.Latitude
	}
	if opts.Longitude != "" {
		attributes["address-longitude"] = opts.Longitude
	}
	if opts.SkipGeocoding {
		attributes["skip-address-geocoding"] = true
	}

	relationships := map[string]any{
		"parked": map[string]any{
			"data": map[string]any{
				"type": opts.ParkedType,
				"id":   opts.ParkedID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "parking-sites",
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

	body, _, err := client.Post(cmd.Context(), "/v1/parking-sites", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created parking site %s\n", resp.Data.ID)
	return nil
}

func parseDoParkingSitesCreateOptions(cmd *cobra.Command) (doParkingSitesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	parkedType, _ := cmd.Flags().GetString("parked-type")
	parkedID, _ := cmd.Flags().GetString("parked-id")
	isActive, _ := cmd.Flags().GetBool("is-active")
	activeStart, _ := cmd.Flags().GetString("active-start")
	activeEnd, _ := cmd.Flags().GetString("active-end")
	address, _ := cmd.Flags().GetString("address")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	skipGeocoding, _ := cmd.Flags().GetBool("skip-geocoding")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doParkingSitesCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ParkedType:    parkedType,
		ParkedID:      parkedID,
		IsActive:      isActive,
		ActiveStart:   activeStart,
		ActiveEnd:     activeEnd,
		Address:       address,
		Latitude:      latitude,
		Longitude:     longitude,
		SkipGeocoding: skipGeocoding,
	}, nil
}
