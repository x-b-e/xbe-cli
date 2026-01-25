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

type materialTypeConversionsListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	MaterialSupplier        string
	MaterialSite            string
	MaterialType            string
	ForeignMaterialSupplier string
	ForeignMaterialSite     string
	ForeignMaterialType     string
}

type materialTypeConversionRow struct {
	ID                        string `json:"id"`
	MaterialSupplierID        string `json:"material_supplier_id,omitempty"`
	MaterialSupplier          string `json:"material_supplier,omitempty"`
	MaterialSiteID            string `json:"material_site_id,omitempty"`
	MaterialSite              string `json:"material_site,omitempty"`
	MaterialTypeID            string `json:"material_type_id,omitempty"`
	MaterialType              string `json:"material_type,omitempty"`
	ForeignMaterialSupplierID string `json:"foreign_material_supplier_id,omitempty"`
	ForeignMaterialSupplier   string `json:"foreign_material_supplier,omitempty"`
	ForeignMaterialSiteID     string `json:"foreign_material_site_id,omitempty"`
	ForeignMaterialSite       string `json:"foreign_material_site,omitempty"`
	ForeignMaterialTypeID     string `json:"foreign_material_type_id,omitempty"`
	ForeignMaterialType       string `json:"foreign_material_type,omitempty"`
}

func newMaterialTypeConversionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material type conversions",
		Long: `List material type conversions with filtering and pagination.

Output Columns:
  ID               Conversion identifier
  SUPPLIER         Local material supplier
  SITE             Local material site (optional)
  MATERIAL_TYPE    Local material type
  FOREIGN_SUPPLIER Foreign material supplier
  FOREIGN_SITE     Foreign material site (optional)
  FOREIGN_TYPE     Foreign material type

Filters:
  --material-supplier        Filter by material supplier ID (comma-separated for multiple)
  --material-site            Filter by material site ID (comma-separated for multiple)
  --material-type            Filter by material type ID (comma-separated for multiple)
  --foreign-material-supplier Filter by foreign material supplier ID (comma-separated for multiple)
  --foreign-material-site    Filter by foreign material site ID (comma-separated for multiple)
  --foreign-material-type    Filter by foreign material type ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List conversions
  xbe view material-type-conversions list

  # Filter by material supplier
  xbe view material-type-conversions list --material-supplier 123

  # Filter by foreign material type
  xbe view material-type-conversions list --foreign-material-type 456

  # Output as JSON
  xbe view material-type-conversions list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialTypeConversionsList,
	}
	initMaterialTypeConversionsListFlags(cmd)
	return cmd
}

func init() {
	materialTypeConversionsCmd.AddCommand(newMaterialTypeConversionsListCmd())
}

func initMaterialTypeConversionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("material-type", "", "Filter by material type ID (comma-separated for multiple)")
	cmd.Flags().String("foreign-material-supplier", "", "Filter by foreign material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("foreign-material-site", "", "Filter by foreign material site ID (comma-separated for multiple)")
	cmd.Flags().String("foreign-material-type", "", "Filter by foreign material type ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTypeConversionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTypeConversionsListOptions(cmd)
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
	query.Set("fields[material-type-conversions]", "material-supplier,material-site,material-type,foreign-material-supplier,foreign-material-site,foreign-material-type")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "display-name,name")
	query.Set("include", "material-supplier,material-site,material-type,foreign-material-supplier,foreign-material-site,foreign-material-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[foreign-material-supplier]", opts.ForeignMaterialSupplier)
	setFilterIfPresent(query, "filter[foreign-material-site]", opts.ForeignMaterialSite)
	setFilterIfPresent(query, "filter[foreign-material-type]", opts.ForeignMaterialType)

	body, _, err := client.Get(cmd.Context(), "/v1/material-type-conversions", query)
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

	rows := buildMaterialTypeConversionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTypeConversionsTable(cmd, rows)
}

func parseMaterialTypeConversionsListOptions(cmd *cobra.Command) (materialTypeConversionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	foreignMaterialSupplier, _ := cmd.Flags().GetString("foreign-material-supplier")
	foreignMaterialSite, _ := cmd.Flags().GetString("foreign-material-site")
	foreignMaterialType, _ := cmd.Flags().GetString("foreign-material-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTypeConversionsListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		MaterialSupplier:        materialSupplier,
		MaterialSite:            materialSite,
		MaterialType:            materialType,
		ForeignMaterialSupplier: foreignMaterialSupplier,
		ForeignMaterialSite:     foreignMaterialSite,
		ForeignMaterialType:     foreignMaterialType,
	}, nil
}

func buildMaterialTypeConversionRows(resp jsonAPIResponse) []materialTypeConversionRow {
	included := map[string]map[string]any{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc.Attributes
	}

	rows := make([]materialTypeConversionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialTypeConversionRow{ID: resource.ID}

		row.MaterialSupplierID = relationshipIDFromMap(resource.Relationships, "material-supplier")
		row.MaterialSupplier = resolveMaterialSupplierName(row.MaterialSupplierID, included)
		row.MaterialSiteID = relationshipIDFromMap(resource.Relationships, "material-site")
		row.MaterialSite = resolveMaterialSiteName(row.MaterialSiteID, included)
		row.MaterialTypeID = relationshipIDFromMap(resource.Relationships, "material-type")
		row.MaterialType = resolveMaterialTypeName(row.MaterialTypeID, included)
		row.ForeignMaterialSupplierID = relationshipIDFromMap(resource.Relationships, "foreign-material-supplier")
		row.ForeignMaterialSupplier = resolveMaterialSupplierName(row.ForeignMaterialSupplierID, included)
		row.ForeignMaterialSiteID = relationshipIDFromMap(resource.Relationships, "foreign-material-site")
		row.ForeignMaterialSite = resolveMaterialSiteName(row.ForeignMaterialSiteID, included)
		row.ForeignMaterialTypeID = relationshipIDFromMap(resource.Relationships, "foreign-material-type")
		row.ForeignMaterialType = resolveMaterialTypeName(row.ForeignMaterialTypeID, included)

		rows = append(rows, row)
	}

	return rows
}

func renderMaterialTypeConversionsTable(cmd *cobra.Command, rows []materialTypeConversionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material type conversions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSUPPLIER\tSITE\tMATERIAL_TYPE\tFOREIGN_SUPPLIER\tFOREIGN_SITE\tFOREIGN_TYPE")
	for _, row := range rows {
		supplier := firstNonEmpty(row.MaterialSupplier, row.MaterialSupplierID)
		site := firstNonEmpty(row.MaterialSite, row.MaterialSiteID)
		materialType := firstNonEmpty(row.MaterialType, row.MaterialTypeID)
		foreignSupplier := firstNonEmpty(row.ForeignMaterialSupplier, row.ForeignMaterialSupplierID)
		foreignSite := firstNonEmpty(row.ForeignMaterialSite, row.ForeignMaterialSiteID)
		foreignType := firstNonEmpty(row.ForeignMaterialType, row.ForeignMaterialTypeID)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(supplier, 30),
			truncateString(site, 30),
			truncateString(materialType, 30),
			truncateString(foreignSupplier, 30),
			truncateString(foreignSite, 30),
			truncateString(foreignType, 30),
		)
	}

	return writer.Flush()
}

func resolveMaterialSupplierName(id string, included map[string]map[string]any) string {
	if id == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("material-suppliers", id)]; ok {
		return stringAttr(attrs, "name")
	}
	return ""
}

func resolveMaterialSiteName(id string, included map[string]map[string]any) string {
	if id == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("material-sites", id)]; ok {
		return stringAttr(attrs, "name")
	}
	return ""
}

func resolveMaterialTypeName(id string, included map[string]map[string]any) string {
	if id == "" {
		return ""
	}
	if attrs, ok := included[resourceKey("material-types", id)]; ok {
		display := stringAttr(attrs, "display-name")
		if display != "" {
			return display
		}
		return stringAttr(attrs, "name")
	}
	return ""
}
