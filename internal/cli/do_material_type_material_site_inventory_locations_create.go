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

type doMaterialTypeMaterialSiteInventoryLocationsCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	MaterialType                  string
	MaterialSiteInventoryLocation string
}

func newDoMaterialTypeMaterialSiteInventoryLocationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a material type material site inventory location",
		Long: `Create a material type material site inventory location.

Required flags:
  --material-type                    Material type ID (required)
  --material-site-inventory-location Material site inventory location ID (required)

Notes:
  Material types must be supplier-specific and share the same supplier
  as the inventory location.`,
		Example: `  # Create a mapping between a material type and inventory location
  xbe do material-type-material-site-inventory-locations create \\
    --material-type 123 \\
    --material-site-inventory-location 456`,
		Args: cobra.NoArgs,
		RunE: runDoMaterialTypeMaterialSiteInventoryLocationsCreate,
	}
	initDoMaterialTypeMaterialSiteInventoryLocationsCreateFlags(cmd)
	return cmd
}

func init() {
	doMaterialTypeMaterialSiteInventoryLocationsCmd.AddCommand(newDoMaterialTypeMaterialSiteInventoryLocationsCreateCmd())
}

func initDoMaterialTypeMaterialSiteInventoryLocationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("material-type", "", "Material type ID (required)")
	cmd.Flags().String("material-site-inventory-location", "", "Material site inventory location ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("material-type")
	_ = cmd.MarkFlagRequired("material-site-inventory-location")
}

func runDoMaterialTypeMaterialSiteInventoryLocationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoMaterialTypeMaterialSiteInventoryLocationsCreateOptions(cmd)
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

	relationships := map[string]any{
		"material-type": map[string]any{
			"data": map[string]any{
				"type": "material-types",
				"id":   opts.MaterialType,
			},
		},
		"material-site-inventory-location": map[string]any{
			"data": map[string]any{
				"type": "material-site-inventory-locations",
				"id":   opts.MaterialSiteInventoryLocation,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "material-type-material-site-inventory-locations",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/material-type-material-site-inventory-locations", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created material type material site inventory location %s\n", row.ID)
	return nil
}

func parseDoMaterialTypeMaterialSiteInventoryLocationsCreateOptions(cmd *cobra.Command) (doMaterialTypeMaterialSiteInventoryLocationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialSiteInventoryLocation, _ := cmd.Flags().GetString("material-site-inventory-location")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doMaterialTypeMaterialSiteInventoryLocationsCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		MaterialType:                  materialType,
		MaterialSiteInventoryLocation: materialSiteInventoryLocation,
	}, nil
}

func materialTypeMaterialSiteInventoryLocationRowFromSingle(resp jsonAPISingleResponse) materialTypeMaterialSiteInventoryLocationRow {
	row := materialTypeMaterialSiteInventoryLocationRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["material-site-inventory-location"]; ok && rel.Data != nil {
		row.MaterialSiteInventoryLocationID = rel.Data.ID
	}

	return row
}
