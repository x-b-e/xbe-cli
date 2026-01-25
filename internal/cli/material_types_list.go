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

type materialTypesListOptions struct {
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Broker                      string
	Name                        string
	Q                           string
	IsArchived                  string
	ParentMaterialType          string
	MaterialSites               string
	ForMaterialSuppliers        string
	HasMaterialSupplier         string
	RequiringMixDesign          string
	MissingRequiredMixDesign    string
	MaterialSupplier            string
	ParentMaterialTypeID        string
	MaterialSubTypes            string
	ExplicitDisplayName         string
	UltimateParentMaterialType  string
	UltimateParentQualifiedName string
	HierarchyLike               string
	MixDesignDescriptionLike    string
	IsSoftDeleted               string
	IsAvailableAt               string
	HasInventoryEstimates       string
}

type materialTypeRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	IsArchived  bool   `json:"is_archived"`
	ParentID    string `json:"parent_id,omitempty"`
	ParentName  string `json:"parent_name,omitempty"`
}

func newMaterialTypesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material types",
		Long: `List material types with filtering and pagination.

Material types define the materials that can be hauled or used in jobs. They
can be organized hierarchically with parent types and sub-types.

Output Columns:
  ID           Material type identifier
  NAME         Material type name
  DISPLAY      Display name (may differ from name)
  ARCHIVED     Whether the material type is archived
  PARENT       Parent material type name (if hierarchical)

Filters:
  --broker               Filter by broker ID
  --name                 Filter by name (partial match, case-insensitive)
  --q                    General search
  --is-archived          Filter by archived status (true/false)
  --parent-material-type Filter by parent material type ID`,
		Example: `  # List all material types
  xbe view material-types list

  # Filter by broker
  xbe view material-types list --broker 123

  # Search by name
  xbe view material-types list --name "gravel"

  # Filter active (non-archived) only
  xbe view material-types list --is-archived false

  # Output as JSON
  xbe view material-types list --json`,
		RunE: runMaterialTypesList,
	}
	initMaterialTypesListFlags(cmd)
	return cmd
}

func init() {
	materialTypesCmd.AddCommand(newMaterialTypesListCmd())
}

func initMaterialTypesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("name", "", "Filter by name (partial match)")
	cmd.Flags().String("q", "", "General search")
	cmd.Flags().String("is-archived", "", "Filter by archived status (true/false)")
	cmd.Flags().String("parent-material-type", "", "Filter by parent material type ID")
	cmd.Flags().String("material-sites", "", "Filter by material site IDs (comma-separated)")
	cmd.Flags().String("for-material-suppliers", "", "Filter for material suppliers (comma-separated IDs)")
	cmd.Flags().String("has-material-supplier", "", "Filter by having material supplier (true/false)")
	cmd.Flags().String("requiring-mix-design", "", "Filter by requiring mix design (true/false)")
	cmd.Flags().String("missing-required-mix-design", "", "Filter by missing required mix design (true/false)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("parent-material-type-id", "", "Filter by parent material type ID (empty for top-level)")
	cmd.Flags().String("material-sub-types", "", "Filter by material sub type IDs (comma-separated)")
	cmd.Flags().String("explicit-display-name", "", "Filter by explicit display name")
	cmd.Flags().String("ultimate-parent-material-type", "", "Filter by ultimate parent material type ID")
	cmd.Flags().String("ultimate-parent-qualified-name", "", "Filter by ultimate parent qualified name")
	cmd.Flags().String("hierarchy-like", "", "Filter by hierarchy (partial match)")
	cmd.Flags().String("mix-design-description-like", "", "Filter by mix design description (partial match)")
	cmd.Flags().String("is-soft-deleted", "", "Filter by soft deleted status (true/false)")
	cmd.Flags().String("is-available-at", "", "Filter by availability at datetime (ISO 8601)")
	cmd.Flags().String("has-inventory-estimates", "", "Filter by having inventory estimates (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialTypesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialTypesListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[material-types]", "name,display-name,is-archived,parent-material-type")
	query.Set("include", "parent-material-type")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[is-archived]", opts.IsArchived)
	setFilterIfPresent(query, "filter[parent-material-type]", opts.ParentMaterialType)
	setFilterIfPresent(query, "filter[material-sites]", opts.MaterialSites)
	setFilterIfPresent(query, "filter[for-material-suppliers]", opts.ForMaterialSuppliers)
	setFilterIfPresent(query, "filter[has-material-supplier]", opts.HasMaterialSupplier)
	setFilterIfPresent(query, "filter[requiring-mix-design]", opts.RequiringMixDesign)
	setFilterIfPresent(query, "filter[missing-required-mix-design]", opts.MissingRequiredMixDesign)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[parent-material-type-id]", opts.ParentMaterialTypeID)
	setFilterIfPresent(query, "filter[material-sub-types]", opts.MaterialSubTypes)
	setFilterIfPresent(query, "filter[explicit-display-name]", opts.ExplicitDisplayName)
	setFilterIfPresent(query, "filter[ultimate-parent-material-type]", opts.UltimateParentMaterialType)
	setFilterIfPresent(query, "filter[ultimate-parent-qualified-name]", opts.UltimateParentQualifiedName)
	setFilterIfPresent(query, "filter[hierarchy-like]", opts.HierarchyLike)
	setFilterIfPresent(query, "filter[mix-design-description-like]", opts.MixDesignDescriptionLike)
	setFilterIfPresent(query, "filter[is-soft-deleted]", opts.IsSoftDeleted)
	setFilterIfPresent(query, "filter[is-available-at]", opts.IsAvailableAt)
	setFilterIfPresent(query, "filter[has-inventory-estimates]", opts.HasInventoryEstimates)

	body, _, err := client.Get(cmd.Context(), "/v1/material-types", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildMaterialTypeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialTypesTable(cmd, rows)
}

func parseMaterialTypesListOptions(cmd *cobra.Command) (materialTypesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	broker, _ := cmd.Flags().GetString("broker")
	name, _ := cmd.Flags().GetString("name")
	q, _ := cmd.Flags().GetString("q")
	isArchived, _ := cmd.Flags().GetString("is-archived")
	parentMaterialType, _ := cmd.Flags().GetString("parent-material-type")
	materialSites, _ := cmd.Flags().GetString("material-sites")
	forMaterialSuppliers, _ := cmd.Flags().GetString("for-material-suppliers")
	hasMaterialSupplier, _ := cmd.Flags().GetString("has-material-supplier")
	requiringMixDesign, _ := cmd.Flags().GetString("requiring-mix-design")
	missingRequiredMixDesign, _ := cmd.Flags().GetString("missing-required-mix-design")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	parentMaterialTypeID, _ := cmd.Flags().GetString("parent-material-type-id")
	materialSubTypes, _ := cmd.Flags().GetString("material-sub-types")
	explicitDisplayName, _ := cmd.Flags().GetString("explicit-display-name")
	ultimateParentMaterialType, _ := cmd.Flags().GetString("ultimate-parent-material-type")
	ultimateParentQualifiedName, _ := cmd.Flags().GetString("ultimate-parent-qualified-name")
	hierarchyLike, _ := cmd.Flags().GetString("hierarchy-like")
	mixDesignDescriptionLike, _ := cmd.Flags().GetString("mix-design-description-like")
	isSoftDeleted, _ := cmd.Flags().GetString("is-soft-deleted")
	isAvailableAt, _ := cmd.Flags().GetString("is-available-at")
	hasInventoryEstimates, _ := cmd.Flags().GetString("has-inventory-estimates")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialTypesListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Broker:                      broker,
		Name:                        name,
		Q:                           q,
		IsArchived:                  isArchived,
		ParentMaterialType:          parentMaterialType,
		MaterialSites:               materialSites,
		ForMaterialSuppliers:        forMaterialSuppliers,
		HasMaterialSupplier:         hasMaterialSupplier,
		RequiringMixDesign:          requiringMixDesign,
		MissingRequiredMixDesign:    missingRequiredMixDesign,
		MaterialSupplier:            materialSupplier,
		ParentMaterialTypeID:        parentMaterialTypeID,
		MaterialSubTypes:            materialSubTypes,
		ExplicitDisplayName:         explicitDisplayName,
		UltimateParentMaterialType:  ultimateParentMaterialType,
		UltimateParentQualifiedName: ultimateParentQualifiedName,
		HierarchyLike:               hierarchyLike,
		MixDesignDescriptionLike:    mixDesignDescriptionLike,
		IsSoftDeleted:               isSoftDeleted,
		IsAvailableAt:               isAvailableAt,
		HasInventoryEstimates:       hasInventoryEstimates,
	}, nil
}

func buildMaterialTypeRows(resp jsonAPIResponse) []materialTypeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]materialTypeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := materialTypeRow{
			ID:          resource.ID,
			Name:        stringAttr(resource.Attributes, "name"),
			DisplayName: stringAttr(resource.Attributes, "display-name"),
			IsArchived:  boolAttr(resource.Attributes, "is-archived"),
		}

		if rel, ok := resource.Relationships["parent-material-type"]; ok && rel.Data != nil {
			row.ParentID = rel.Data.ID
			if parent, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ParentName = stringAttr(parent.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderMaterialTypesTable(cmd *cobra.Command, rows []materialTypeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material types found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tDISPLAY\tARCHIVED\tPARENT")
	for _, row := range rows {
		archived := "no"
		if row.IsArchived {
			archived = "yes"
		}
		displayName := row.DisplayName
		if displayName == row.Name {
			displayName = ""
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 30),
			truncateString(displayName, 25),
			archived,
			truncateString(row.ParentName, 25),
		)
	}
	return writer.Flush()
}
