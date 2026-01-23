package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type materialTypeMaterialSiteInventoryLocationsListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	MaterialType                  string
	MaterialSiteInventoryLocation string
}

type materialTypeMaterialSiteInventoryLocationRow struct {
	ID                              string `json:"id"`
	MaterialTypeID                  string `json:"material_type_id,omitempty"`
	MaterialTypeName                string `json:"material_type_name,omitempty"`
	MaterialSiteInventoryLocationID string `json:"material_site_inventory_location_id,omitempty"`
	MaterialSiteInventoryLocation   string `json:"material_site_inventory_location,omitempty"`
}

func newMaterialTypeMaterialSiteInventoryLocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material type material site inventory locations",
		Long: `List material type material site inventory locations with filtering and pagination.

Material type material site inventory locations associate supplier-specific
material types with inventory locations at a material site.

Output Columns:
  ID                   Mapping identifier
  MATERIAL TYPE        Material type name or ID
  INVENTORY LOCATION   Inventory location label or ID

Filters:
  --material-type                    Filter by material type ID
  --material-site-inventory-location Filter by material site inventory location ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List material type material site inventory locations
  xbe view material-type-material-site-inventory-locations list

  # Filter by material type
  xbe view material-type-material-site-inventory-locations list --material-type 123

  # Filter by inventory location
  xbe view material-type-material-site-inventory-locations list --material-site-inventory-location 456

  # Output as JSON
  xbe view material-type-material-site-inventory-locations list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTypeMaterialSiteInventoryLocationsList,
	}
	initMaterialTypeMaterialSiteInventoryLocationsListFlags(cmd)
	return cmd
}

func init() {
	materialTypeMaterialSiteInventoryLocationsCmd.AddCommand(newMaterialTypeMaterialSiteInventoryLocationsListCmd())
}

func initMaterialTypeMaterialSiteInventoryLocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("material-site-inventory-location", "", "Filter by material site inventory location ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTypeMaterialSiteInventoryLocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTypeMaterialSiteInventoryLocationsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-type-material-site-inventory-locations]", "material-type,material-site-inventory-location")
	query.Set("include", "material-type,material-site-inventory-location")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[material-site-inventory-locations]", "qualified-name,display-name-explicit")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "id")
	}
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material-site-inventory-location]", opts.MaterialSiteInventoryLocation)

	body, _, err := client.Get(cmd.Context(), "/v1/material-type-material-site-inventory-locations", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildMaterialTypeMaterialSiteInventoryLocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTypeMaterialSiteInventoryLocationsTable(cmd, rows)
}

func parseMaterialTypeMaterialSiteInventoryLocationsListOptions(cmd *cobra.Command) (materialTypeMaterialSiteInventoryLocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialSiteInventoryLocation, _ := cmd.Flags().GetString("material-site-inventory-location")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTypeMaterialSiteInventoryLocationsListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		MaterialType:                  materialType,
		MaterialSiteInventoryLocation: materialSiteInventoryLocation,
	}, nil
}

func buildMaterialTypeMaterialSiteInventoryLocationRows(resp jsonAPIResponse) []materialTypeMaterialSiteInventoryLocationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialTypeMaterialSiteInventoryLocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialTypeMaterialSiteInventoryLocationRow{
			ID: resource.ID,
		}

		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
			if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialTypeName = materialTypeLabel(materialType.Attributes)
			}
		}

		if rel, ok := resource.Relationships["material-site-inventory-location"]; ok && rel.Data != nil {
			row.MaterialSiteInventoryLocationID = rel.Data.ID
			if location, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialSiteInventoryLocation = materialSiteInventoryLocationLabel(location.Attributes)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderMaterialTypeMaterialSiteInventoryLocationsTable(cmd *cobra.Command, rows []materialTypeMaterialSiteInventoryLocationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material type material site inventory locations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMATERIAL TYPE\tINVENTORY LOCATION")
	for _, row := range rows {
		typeLabel := row.MaterialTypeName
		if typeLabel == "" {
			typeLabel = row.MaterialTypeID
		}
		locationLabel := row.MaterialSiteInventoryLocation
		if locationLabel == "" {
			locationLabel = row.MaterialSiteInventoryLocationID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(typeLabel, 30),
			truncateString(locationLabel, 30),
		)
	}
	return writer.Flush()
}

func materialSiteInventoryLocationLabel(attrs map[string]any) string {
	displayName := strings.TrimSpace(stringAttr(attrs, "display-name-explicit"))
	if displayName != "" {
		return displayName
	}
	qualifiedName := strings.TrimSpace(stringAttr(attrs, "qualified-name"))
	if qualifiedName != "" {
		return qualifiedName
	}
	return ""
}
