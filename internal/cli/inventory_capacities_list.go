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

type inventoryCapacitiesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	MaterialSite string
	MaterialType string
}

type inventoryCapacityRow struct {
	ID               string `json:"id"`
	MaterialSiteID   string `json:"material_site_id,omitempty"`
	MaterialSiteName string `json:"material_site_name,omitempty"`
	MaterialTypeID   string `json:"material_type_id,omitempty"`
	MaterialTypeName string `json:"material_type_name,omitempty"`
	MaxCapacityTons  string `json:"max_capacity_tons,omitempty"`
	MinCapacityTons  string `json:"min_capacity_tons,omitempty"`
	ThresholdTons    string `json:"threshold_tons,omitempty"`
}

func newInventoryCapacitiesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inventory capacities",
		Long: `List inventory capacities with filtering and pagination.

Inventory capacities define min/max storage levels and alert thresholds for
material sites and material types.

Output Columns:
  ID             Inventory capacity identifier
  MATERIAL SITE  Material site name and ID
  MATERIAL TYPE  Material type name and ID
  MIN TONS       Minimum capacity (tons)
  MAX TONS       Maximum capacity (tons)
  THRESHOLD      Alert threshold (tons)

Filters:
  --material-site  Filter by material site ID
  --material-type  Filter by material type ID`,
		Example: `  # List inventory capacities
  xbe view inventory-capacities list

  # Filter by material site
  xbe view inventory-capacities list --material-site 123

  # Filter by material type
  xbe view inventory-capacities list --material-type 456

  # Output as JSON
  xbe view inventory-capacities list --json`,
		RunE: runInventoryCapacitiesList,
	}
	initInventoryCapacitiesListFlags(cmd)
	return cmd
}

func init() {
	inventoryCapacitiesCmd.AddCommand(newInventoryCapacitiesListCmd())
}

func initInventoryCapacitiesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInventoryCapacitiesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInventoryCapacitiesListOptions(cmd)
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
	query.Set("sort", "id")
	query.Set("fields[inventory-capacities]", "max-capacity-tons,min-capacity-tons,threshold-tons,material-site,material-type")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "name")
	query.Set("include", "material-site,material-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[material_site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material_type]", opts.MaterialType)

	body, _, err := client.Get(cmd.Context(), "/v1/inventory-capacities", query)
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

	rows := buildInventoryCapacityRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInventoryCapacitiesTable(cmd, rows)
}

func parseInventoryCapacitiesListOptions(cmd *cobra.Command) (inventoryCapacitiesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return inventoryCapacitiesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		MaterialSite: materialSite,
		MaterialType: materialType,
	}, nil
}

func buildInventoryCapacityRows(resp jsonAPIResponse) []inventoryCapacityRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]inventoryCapacityRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := inventoryCapacityRow{
			ID:              resource.ID,
			MaxCapacityTons: strings.TrimSpace(stringAttr(attrs, "max-capacity-tons")),
			MinCapacityTons: strings.TrimSpace(stringAttr(attrs, "min-capacity-tons")),
			ThresholdTons:   strings.TrimSpace(stringAttr(attrs, "threshold-tons")),
		}

		if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
			row.MaterialSiteID = rel.Data.ID
			if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialSiteName = strings.TrimSpace(stringAttr(site.Attributes, "name"))
			}
		}
		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
			if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialTypeName = strings.TrimSpace(stringAttr(mt.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderInventoryCapacitiesTable(cmd *cobra.Command, rows []inventoryCapacityRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No inventory capacities found.")
		return nil
	}

	const nameMax = 36

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tMATERIAL SITE\tMATERIAL TYPE\tMIN TONS\tMAX TONS\tTHRESHOLD")
	for _, row := range rows {
		siteLabel := formatInventoryCapacityLabel(row.MaterialSiteName, row.MaterialSiteID)
		typeLabel := formatInventoryCapacityLabel(row.MaterialTypeName, row.MaterialTypeID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(siteLabel, nameMax),
			truncateString(typeLabel, nameMax),
			row.MinCapacityTons,
			row.MaxCapacityTons,
			row.ThresholdTons,
		)
	}
	return writer.Flush()
}

func formatInventoryCapacityLabel(name, id string) string {
	if name != "" && id != "" {
		return fmt.Sprintf("%s (%s)", name, id)
	}
	if name != "" {
		return name
	}
	return id
}
