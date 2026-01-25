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

type doMaterialSiteInventoryLocationsUpdateOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	ID              string
	DisplayName     string
	Latitude        string
	Longitude       string
	UnitOfMeasureID string
}

func newDoMaterialSiteInventoryLocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material site inventory location",
		Long: `Update a material site inventory location.

Optional flags:
  --display-name-explicit  Display name override
  --latitude               Latitude coordinate (use with --longitude)
  --longitude              Longitude coordinate (use with --latitude)
  --unit-of-measure        Unit of measure ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update display name
  xbe do material-site-inventory-locations update 123 --display-name-explicit "New Name"

  # Update coordinates
  xbe do material-site-inventory-locations update 123 --latitude 41.882 --longitude -87.624`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialSiteInventoryLocationsUpdate,
	}
	initDoMaterialSiteInventoryLocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialSiteInventoryLocationsCmd.AddCommand(newDoMaterialSiteInventoryLocationsUpdateCmd())
}

func initDoMaterialSiteInventoryLocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("display-name-explicit", "", "Display name override")
	cmd.Flags().String("latitude", "", "Latitude coordinate")
	cmd.Flags().String("longitude", "", "Longitude coordinate")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialSiteInventoryLocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialSiteInventoryLocationsUpdateOptions(cmd, args)
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
	var relationships map[string]any

	if cmd.Flags().Changed("display-name-explicit") {
		attributes["display-name-explicit"] = opts.DisplayName
	}

	latChanged := cmd.Flags().Changed("latitude")
	lonChanged := cmd.Flags().Changed("longitude")
	if latChanged || lonChanged {
		if !(latChanged && lonChanged) {
			err := fmt.Errorf("latitude and longitude must be provided together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["latitude"] = opts.Latitude
		attributes["longitude"] = opts.Longitude
	}

	if cmd.Flags().Changed("unit-of-measure") {
		if strings.TrimSpace(opts.UnitOfMeasureID) == "" {
			err := fmt.Errorf("unit-of-measure id is required when updating unit-of-measure")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships = map[string]any{
			"unit-of-measure": map[string]any{
				"data": map[string]any{
					"type": "unit-of-measures",
					"id":   opts.UnitOfMeasureID,
				},
			},
		}
	}

	if len(attributes) == 0 && relationships == nil {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "material-site-inventory-locations",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if relationships != nil {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/material-site-inventory-locations/"+opts.ID, jsonBody)
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

	row := materialSiteInventoryLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material site inventory location %s\n", row.ID)
	return nil
}

func parseDoMaterialSiteInventoryLocationsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialSiteInventoryLocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	displayName, _ := cmd.Flags().GetString("display-name-explicit")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	unitOfMeasureID, _ := cmd.Flags().GetString("unit-of-measure")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialSiteInventoryLocationsUpdateOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		ID:              args[0],
		DisplayName:     displayName,
		Latitude:        latitude,
		Longitude:       longitude,
		UnitOfMeasureID: unitOfMeasureID,
	}, nil
}
