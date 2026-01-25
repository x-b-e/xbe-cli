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

type materialSiteInventoryLocationsListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	QualifiedName      string
	UnitOfMeasure      string
	MaterialSite       string
	BrokerID           string
	Broker             string
	MaterialSupplierID string
	MaterialSupplier   string
}

type materialSiteInventoryLocationRow struct {
	ID                 string `json:"id"`
	QualifiedName      string `json:"qualified_name,omitempty"`
	DisplayName        string `json:"display_name_explicit,omitempty"`
	Latitude           string `json:"latitude,omitempty"`
	Longitude          string `json:"longitude,omitempty"`
	MaterialSiteID     string `json:"material_site_id,omitempty"`
	MaterialSite       string `json:"material_site,omitempty"`
	UnitOfMeasureID    string `json:"unit_of_measure_id,omitempty"`
	UnitOfMeasure      string `json:"unit_of_measure,omitempty"`
	BrokerID           string `json:"broker_id,omitempty"`
	Broker             string `json:"broker,omitempty"`
	MaterialSupplierID string `json:"material_supplier_id,omitempty"`
	MaterialSupplier   string `json:"material_supplier,omitempty"`
}

func newMaterialSiteInventoryLocationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material site inventory locations",
		Long: `List material site inventory locations with filtering and pagination.

Output Columns:
  ID            Inventory location identifier
  NAME          Display name (falls back to qualified name)
  MATERIAL SITE Material site name
  UNIT          Unit of measure
  SUPPLIER      Material supplier name

Filters:
  --qualified-name       Filter by qualified name
  --material-site        Filter by material site ID (comma-separated for multiple)
  --unit-of-measure      Filter by unit of measure ID (comma-separated for multiple)
  --broker-id            Filter by broker ID (comma-separated for multiple)
  --broker               Filter by broker ID (comma-separated for multiple)
  --material-supplier-id Filter by material supplier ID (comma-separated for multiple)
  --material-supplier    Filter by material supplier ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List material site inventory locations
  xbe view material-site-inventory-locations list

  # Filter by material site
  xbe view material-site-inventory-locations list --material-site 123

  # Filter by qualified name
  xbe view material-site-inventory-locations list --qualified-name "Plant A"

  # Filter by broker
  xbe view material-site-inventory-locations list --broker 456

  # Output as JSON
  xbe view material-site-inventory-locations list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialSiteInventoryLocationsList,
	}
	initMaterialSiteInventoryLocationsListFlags(cmd)
	return cmd
}

func init() {
	materialSiteInventoryLocationsCmd.AddCommand(newMaterialSiteInventoryLocationsListCmd())
}

func initMaterialSiteInventoryLocationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("qualified-name", "", "Filter by qualified name")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID (comma-separated for multiple)")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("material-supplier-id", "", "Filter by material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialSiteInventoryLocationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialSiteInventoryLocationsListOptions(cmd)
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
	query.Set("fields[material-site-inventory-locations]", "qualified-name,display-name-explicit,latitude,longitude,material-site,unit-of-measure,broker,material-supplier")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("include", "material-site,unit-of-measure,broker,material-supplier")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[qualified_name]", opts.QualifiedName)
	setFilterIfPresent(query, "filter[unit_of_measure]", opts.UnitOfMeasure)
	setFilterIfPresent(query, "filter[material_site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[broker_id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[material_supplier_id]", opts.MaterialSupplierID)
	setFilterIfPresent(query, "filter[material_supplier]", opts.MaterialSupplier)

	body, _, err := client.Get(cmd.Context(), "/v1/material-site-inventory-locations", query)
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

	rows := buildMaterialSiteInventoryLocationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialSiteInventoryLocationsTable(cmd, rows)
}

func parseMaterialSiteInventoryLocationsListOptions(cmd *cobra.Command) (materialSiteInventoryLocationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	qualifiedName, _ := cmd.Flags().GetString("qualified-name")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	materialSite, _ := cmd.Flags().GetString("material-site")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	broker, _ := cmd.Flags().GetString("broker")
	materialSupplierID, _ := cmd.Flags().GetString("material-supplier-id")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialSiteInventoryLocationsListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Sort:               sort,
		QualifiedName:      qualifiedName,
		UnitOfMeasure:      unitOfMeasure,
		MaterialSite:       materialSite,
		BrokerID:           brokerID,
		Broker:             broker,
		MaterialSupplierID: materialSupplierID,
		MaterialSupplier:   materialSupplier,
	}, nil
}

func buildMaterialSiteInventoryLocationRows(resp jsonAPIResponse) []materialSiteInventoryLocationRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialSiteInventoryLocationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialSiteInventoryLocationRowFromResource(resource, included))
	}

	return rows
}

func buildMaterialSiteInventoryLocationRowFromResource(resource jsonAPIResource, included map[string]jsonAPIResource) materialSiteInventoryLocationRow {
	attrs := resource.Attributes
	row := materialSiteInventoryLocationRow{
		ID:            resource.ID,
		QualifiedName: stringAttr(attrs, "qualified-name"),
		DisplayName:   stringAttr(attrs, "display-name-explicit"),
		Latitude:      stringAttr(attrs, "latitude"),
		Longitude:     stringAttr(attrs, "longitude"),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
		if ms, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialSite = stringAttr(ms.Attributes, "name")
		}
	}

	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		row.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.UnitOfMeasure = firstNonEmpty(
				stringAttr(uom.Attributes, "abbreviation"),
				stringAttr(uom.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.Broker = stringAttr(broker.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		row.MaterialSupplierID = rel.Data.ID
		if supplier, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialSupplier = stringAttr(supplier.Attributes, "name")
		}
	}

	return row
}

func renderMaterialSiteInventoryLocationsTable(cmd *cobra.Command, rows []materialSiteInventoryLocationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material site inventory locations found.")
		return nil
	}

	const (
		maxName     = 28
		maxSite     = 24
		maxUnit     = 16
		maxSupplier = 24
	)

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tMATERIAL SITE\tUNIT\tSUPPLIER")
	for _, row := range rows {
		nameLabel := firstNonEmpty(row.DisplayName, row.QualifiedName)
		siteLabel := firstNonEmpty(row.MaterialSite, row.MaterialSiteID)
		unitLabel := firstNonEmpty(row.UnitOfMeasure, row.UnitOfMeasureID)
		supplierLabel := firstNonEmpty(row.MaterialSupplier, row.MaterialSupplierID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(nameLabel, maxName),
			truncateString(siteLabel, maxSite),
			truncateString(unitLabel, maxUnit),
			truncateString(supplierLabel, maxSupplier),
		)
	}
	return writer.Flush()
}
