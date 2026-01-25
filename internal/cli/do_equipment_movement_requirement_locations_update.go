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

type doEquipmentMovementRequirementLocationsUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	Name      string
	Latitude  string
	Longitude string
}

func newDoEquipmentMovementRequirementLocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an equipment movement requirement location",
		Long: `Update an equipment movement requirement location.

Optional flags:
  --name        Location name
  --latitude    Latitude coordinate
  --longitude   Longitude coordinate`,
		Example: `  # Update location name
  xbe do equipment-movement-requirement-locations update 123 --name "Updated Yard"

  # Update coordinates
  xbe do equipment-movement-requirement-locations update 123 \
    --latitude 37.8 \
    --longitude -122.4`,
		Args: cobra.ExactArgs(1),
		RunE: runDoEquipmentMovementRequirementLocationsUpdate,
	}
	initDoEquipmentMovementRequirementLocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doEquipmentMovementRequirementLocationsCmd.AddCommand(newDoEquipmentMovementRequirementLocationsUpdateCmd())
}

func initDoEquipmentMovementRequirementLocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Location name")
	cmd.Flags().String("latitude", "", "Latitude coordinate")
	cmd.Flags().String("longitude", "", "Longitude coordinate")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoEquipmentMovementRequirementLocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoEquipmentMovementRequirementLocationsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("latitude") {
		attributes["latitude"] = opts.Latitude
	}
	if cmd.Flags().Changed("longitude") {
		attributes["longitude"] = opts.Longitude
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "equipment-movement-requirement-locations",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/equipment-movement-requirement-locations/"+opts.ID, jsonBody)
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

	row := buildEquipmentMovementRequirementLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated equipment movement requirement location %s\n", row.ID)
	return nil
}

func parseDoEquipmentMovementRequirementLocationsUpdateOptions(cmd *cobra.Command, args []string) (doEquipmentMovementRequirementLocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	latitude, _ := cmd.Flags().GetString("latitude")
	longitude, _ := cmd.Flags().GetString("longitude")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doEquipmentMovementRequirementLocationsUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		Name:      name,
		Latitude:  latitude,
		Longitude: longitude,
	}, nil
}
