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

type materialSiteReadingMaterialTypesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	MaterialSite string
	MaterialType string
	ExternalID   string
}

type materialSiteReadingMaterialTypeRow struct {
	ID               string `json:"id"`
	ExternalID       string `json:"external_id,omitempty"`
	MaterialSiteID   string `json:"material_site_id,omitempty"`
	MaterialSiteName string `json:"material_site_name,omitempty"`
	MaterialTypeID   string `json:"material_type_id,omitempty"`
	MaterialTypeName string `json:"material_type_name,omitempty"`
}

func newMaterialSiteReadingMaterialTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material site reading material types",
		Long: `List material site reading material types with filtering and pagination.

Material site reading material types map external material identifiers from
plant or system readings to internal material types for a specific material site.

Output Columns:
  ID           Mapping identifier
  EXTERNAL ID  External identifier from the source system
  MATERIAL SITE  Material site name or ID
  MATERIAL TYPE  Material type name or ID

Filters:
  --material-site  Filter by material site ID
  --material-type  Filter by material type ID
  --external-id    Filter by external identifier

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List material site reading material types
  xbe view material-site-reading-material-types list

  # Filter by material site
  xbe view material-site-reading-material-types list --material-site 123

  # Filter by material type
  xbe view material-site-reading-material-types list --material-type 456

  # Filter by external ID
  xbe view material-site-reading-material-types list --external-id \"EXT-100\"

  # Output as JSON
  xbe view material-site-reading-material-types list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialSiteReadingMaterialTypesList,
	}
	initMaterialSiteReadingMaterialTypesListFlags(cmd)
	return cmd
}

func init() {
	materialSiteReadingMaterialTypesCmd.AddCommand(newMaterialSiteReadingMaterialTypesListCmd())
}

func initMaterialSiteReadingMaterialTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("external-id", "", "Filter by external identifier")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteReadingMaterialTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSiteReadingMaterialTypesListOptions(cmd)
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
	query.Set("fields[material-site-reading-material-types]", "external-id,material-site,material-type")
	query.Set("include", "material-site,material-type")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "name,display-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "external-id")
	}
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[external-id]", opts.ExternalID)

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-reading-material-types", query)
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

	rows := buildMaterialSiteReadingMaterialTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSiteReadingMaterialTypesTable(cmd, rows)
}

func parseMaterialSiteReadingMaterialTypesListOptions(cmd *cobra.Command) (materialSiteReadingMaterialTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	externalID, _ := cmd.Flags().GetString("external-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteReadingMaterialTypesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		MaterialSite: materialSite,
		MaterialType: materialType,
		ExternalID:   externalID,
	}, nil
}

func buildMaterialSiteReadingMaterialTypeRows(resp jsonAPIResponse) []materialSiteReadingMaterialTypeRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialSiteReadingMaterialTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialSiteReadingMaterialTypeRow{
			ID:         resource.ID,
			ExternalID: stringAttr(resource.Attributes, "external-id"),
		}

		if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
			row.MaterialSiteID = rel.Data.ID
			if site, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialSiteName = stringAttr(site.Attributes, "name")
			}
		}

		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			row.MaterialTypeID = rel.Data.ID
			if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialTypeName = materialTypeLabel(materialType.Attributes)
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderMaterialSiteReadingMaterialTypesTable(cmd *cobra.Command, rows []materialSiteReadingMaterialTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material site reading material types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEXTERNAL ID\tMATERIAL SITE\tMATERIAL TYPE")
	for _, row := range rows {
		siteLabel := row.MaterialSiteName
		if siteLabel == "" {
			siteLabel = row.MaterialSiteID
		}
		typeLabel := row.MaterialTypeName
		if typeLabel == "" {
			typeLabel = row.MaterialTypeID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.ExternalID, 30),
			truncateString(siteLabel, 30),
			truncateString(typeLabel, 30),
		)
	}
	return writer.Flush()
}

func materialTypeLabel(attrs map[string]any) string {
	displayName := strings.TrimSpace(stringAttr(attrs, "display-name"))
	if displayName != "" {
		return displayName
	}
	return strings.TrimSpace(stringAttr(attrs, "name"))
}
