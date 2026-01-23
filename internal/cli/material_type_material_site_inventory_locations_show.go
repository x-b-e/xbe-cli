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

type materialTypeMaterialSiteInventoryLocationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialTypeMaterialSiteInventoryLocationDetails struct {
	ID                              string `json:"id"`
	MaterialTypeID                  string `json:"material_type_id,omitempty"`
	MaterialTypeName                string `json:"material_type_name,omitempty"`
	MaterialSiteInventoryLocationID string `json:"material_site_inventory_location_id,omitempty"`
	MaterialSiteInventoryLocation   string `json:"material_site_inventory_location,omitempty"`
}

func newMaterialTypeMaterialSiteInventoryLocationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material type material site inventory location details",
		Long: `Show the full details of a material type material site inventory location.

Includes the associated material type and material site inventory location.

Arguments:
  <id>  The material type material site inventory location ID (required).`,
		Example: `  # Show a material type material site inventory location
  xbe view material-type-material-site-inventory-locations show 123

  # Output as JSON
  xbe view material-type-material-site-inventory-locations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialTypeMaterialSiteInventoryLocationsShow,
	}
	initMaterialTypeMaterialSiteInventoryLocationsShowFlags(cmd)
	return cmd
}

func init() {
	materialTypeMaterialSiteInventoryLocationsCmd.AddCommand(newMaterialTypeMaterialSiteInventoryLocationsShowCmd())
}

func initMaterialTypeMaterialSiteInventoryLocationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTypeMaterialSiteInventoryLocationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialTypeMaterialSiteInventoryLocationsShowOptions(cmd)
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
		return fmt.Errorf("material type material site inventory location id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-type-material-site-inventory-locations]", "material-type,material-site-inventory-location")
	query.Set("include", "material-type,material-site-inventory-location")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[material-site-inventory-locations]", "qualified-name,display-name-explicit")

	body, _, err := client.Get(cmd.Context(), "/v1/material-type-material-site-inventory-locations/"+id, query)
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

	details := buildMaterialTypeMaterialSiteInventoryLocationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialTypeMaterialSiteInventoryLocationDetails(cmd, details)
}

func parseMaterialTypeMaterialSiteInventoryLocationsShowOptions(cmd *cobra.Command) (materialTypeMaterialSiteInventoryLocationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTypeMaterialSiteInventoryLocationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialTypeMaterialSiteInventoryLocationDetails(resp jsonAPISingleResponse) materialTypeMaterialSiteInventoryLocationDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialTypeMaterialSiteInventoryLocationDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTypeName = materialTypeLabel(materialType.Attributes)
		}
	}

	if rel, ok := resp.Data.Relationships["material-site-inventory-location"]; ok && rel.Data != nil {
		details.MaterialSiteInventoryLocationID = rel.Data.ID
		if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSiteInventoryLocation = materialSiteInventoryLocationLabel(location.Attributes)
		}
	}

	return details
}

func renderMaterialTypeMaterialSiteInventoryLocationDetails(cmd *cobra.Command, details materialTypeMaterialSiteInventoryLocationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.MaterialTypeID != "" {
		label := details.MaterialTypeID
		if details.MaterialTypeName != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialTypeName, details.MaterialTypeID)
		}
		fmt.Fprintf(out, "Material Type: %s\n", label)
	}
	if details.MaterialSiteInventoryLocationID != "" {
		label := details.MaterialSiteInventoryLocationID
		if details.MaterialSiteInventoryLocation != "" {
			label = fmt.Sprintf("%s (%s)", details.MaterialSiteInventoryLocation, details.MaterialSiteInventoryLocationID)
		}
		fmt.Fprintf(out, "Inventory Location: %s\n", label)
	}

	return nil
}
