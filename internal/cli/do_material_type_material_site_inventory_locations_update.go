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

type doMaterialTypeMaterialSiteInventoryLocationsUpdateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ID                            string
	MaterialType                  string
	MaterialSiteInventoryLocation string
}

func newDoMaterialTypeMaterialSiteInventoryLocationsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a material type material site inventory location",
		Long: `Update a material type material site inventory location.

All flags are optional. Only provided flags will update the mapping.

Optional flags:
  --material-type                    Material type ID
  --material-site-inventory-location Material site inventory location ID`,
		Example: `  # Update material type
  xbe do material-type-material-site-inventory-locations update 123 --material-type 456

  # Update inventory location
  xbe do material-type-material-site-inventory-locations update 123 --material-site-inventory-location 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoMaterialTypeMaterialSiteInventoryLocationsUpdate,
	}
	initDoMaterialTypeMaterialSiteInventoryLocationsUpdateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypeMaterialSiteInventoryLocationsCmd.AddCommand(newDoMaterialTypeMaterialSiteInventoryLocationsUpdateCmd())
}

func initDoMaterialTypeMaterialSiteInventoryLocationsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-type", "", "Material type ID")
	cmd.Flags().String("material-site-inventory-location", "", "Material site inventory location ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoMaterialTypeMaterialSiteInventoryLocationsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoMaterialTypeMaterialSiteInventoryLocationsUpdateOptions(cmd, args)
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

	relationships := map[string]any{}

	if cmd.Flags().Changed("material-type") {
		if strings.TrimSpace(opts.MaterialType) == "" {
			err := fmt.Errorf("--material-type cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["material-type"] = map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		}
	}

	if cmd.Flags().Changed("material-site-inventory-location") {
		if strings.TrimSpace(opts.MaterialSiteInventoryLocation) == "" {
			err := fmt.Errorf("--material-site-inventory-location cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["material-site-inventory-location"] = map[string]any{
			"data": map[string]any{
				"type": "material-site-inventory-locations",
				"id":   opts.MaterialSiteInventoryLocation,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-type-material-site-inventory-locations",
			"id":            opts.ID,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/material-type-material-site-inventory-locations/"+opts.ID, jsonBody)
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

	row := materialTypeMaterialSiteInventoryLocationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated material type material site inventory location %s\n", row.ID)
	return nil
}

func parseDoMaterialTypeMaterialSiteInventoryLocationsUpdateOptions(cmd *cobra.Command, args []string) (doMaterialTypeMaterialSiteInventoryLocationsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialSiteInventoryLocation, _ := cmd.Flags().GetString("material-site-inventory-location")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypeMaterialSiteInventoryLocationsUpdateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ID:                            args[0],
		MaterialType:                  materialType,
		MaterialSiteInventoryLocation: materialSiteInventoryLocation,
	}, nil
}
