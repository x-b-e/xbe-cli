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

type inventoryEstimatesListOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	NoAuth             bool
	Limit              int
	Offset             int
	Sort               string
	EstimatedAt        string
	AmountTons         string
	MaterialSite       string
	MaterialType       string
	MaterialSupplierID string
	MaterialSupplier   string
	BrokerID           string
	Broker             string
	CreatedBy          string
}

type inventoryEstimateRow struct {
	ID                          string `json:"id"`
	EstimatedAt                 string `json:"estimated_at,omitempty"`
	AmountTons                  string `json:"amount_tons,omitempty"`
	Description                 string `json:"description,omitempty"`
	MaterialSiteID              string `json:"material_site_id,omitempty"`
	MaterialTypeID              string `json:"material_type_id,omitempty"`
	MaterialSupplierID          string `json:"material_supplier_id,omitempty"`
	BrokerID                    string `json:"broker_id,omitempty"`
	CreatedByID                 string `json:"created_by_id,omitempty"`
	MostRecentInventoryChangeID string `json:"most_recent_inventory_change_id,omitempty"`
}

func newInventoryEstimatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inventory estimates",
		Long: `List inventory estimates.

Output Columns:
  ID           Inventory estimate identifier
  ESTIMATED AT Estimated timestamp
  AMOUNT TONS  Estimated amount (tons)
  MATERIAL SITE Material site ID
  MATERIAL TYPE Material type ID
  DESCRIPTION  Description (truncated)

Filters:
  --estimated-at         Filter by estimated-at timestamp (ISO 8601)
  --amount-tons          Filter by amount in tons
  --material-site        Filter by material site ID
  --material-type        Filter by material type ID
  --material-supplier-id Filter by material supplier ID (join)
  --material-supplier    Filter by material supplier ID
  --broker-id            Filter by broker ID (join)
  --broker               Filter by broker ID
  --created-by           Filter by created-by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List inventory estimates
  xbe view inventory-estimates list

  # Filter by material site and type
  xbe view inventory-estimates list --material-site 123 --material-type 456

  # Filter by estimated-at
  xbe view inventory-estimates list --estimated-at 2025-01-05T08:00:00Z

  # Output as JSON
  xbe view inventory-estimates list --json`,
		Args: cobra.NoArgs,
		RunE: runInventoryEstimatesList,
	}
	initInventoryEstimatesListFlags(cmd)
	return cmd
}

func init() {
	inventoryEstimatesCmd.AddCommand(newInventoryEstimatesListCmd())
}

func initInventoryEstimatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("estimated-at", "", "Filter by estimated-at timestamp (ISO 8601)")
	cmd.Flags().String("amount-tons", "", "Filter by amount in tons")
	cmd.Flags().String("material-site", "", "Filter by material site ID")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("material-supplier-id", "", "Filter by material supplier ID (join)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (join)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInventoryEstimatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInventoryEstimatesListOptions(cmd)
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
	query.Set("fields[inventory-estimates]", "estimated-at,amount-tons,description,material-site,material-type,material-supplier,broker,created-by,most-recent-inventory-change")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[estimated-at]", opts.EstimatedAt)
	setFilterIfPresent(query, "filter[amount-tons]", opts.AmountTons)
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material-supplier-id]", opts.MaterialSupplierID)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/inventory-estimates", query)
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

	rows := buildInventoryEstimateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInventoryEstimatesTable(cmd, rows)
}

func parseInventoryEstimatesListOptions(cmd *cobra.Command) (inventoryEstimatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	estimatedAt, _ := cmd.Flags().GetString("estimated-at")
	amountTons, _ := cmd.Flags().GetString("amount-tons")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialSupplierID, _ := cmd.Flags().GetString("material-supplier-id")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return inventoryEstimatesListOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		NoAuth:             noAuth,
		Limit:              limit,
		Offset:             offset,
		Sort:               sort,
		EstimatedAt:        estimatedAt,
		AmountTons:         amountTons,
		MaterialSite:       materialSite,
		MaterialType:       materialType,
		MaterialSupplierID: materialSupplierID,
		MaterialSupplier:   materialSupplier,
		BrokerID:           brokerID,
		Broker:             broker,
		CreatedBy:          createdBy,
	}, nil
}

func buildInventoryEstimateRows(resp jsonAPIResponse) []inventoryEstimateRow {
	rows := make([]inventoryEstimateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildInventoryEstimateRow(resource))
	}
	return rows
}

func buildInventoryEstimateRowFromSingle(resp jsonAPISingleResponse) inventoryEstimateRow {
	return buildInventoryEstimateRow(resp.Data)
}

func buildInventoryEstimateRow(resource jsonAPIResource) inventoryEstimateRow {
	attrs := resource.Attributes
	row := inventoryEstimateRow{
		ID:          resource.ID,
		EstimatedAt: formatDateTime(stringAttr(attrs, "estimated-at")),
		AmountTons:  stringAttr(attrs, "amount-tons"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		row.MaterialSiteID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		row.MaterialSupplierID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["most-recent-inventory-change"]; ok && rel.Data != nil {
		row.MostRecentInventoryChangeID = rel.Data.ID
	}

	return row
}

func renderInventoryEstimatesTable(cmd *cobra.Command, rows []inventoryEstimateRow) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tESTIMATED AT\tAMOUNT TONS\tMATERIAL SITE\tMATERIAL TYPE\tDESCRIPTION")
	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EstimatedAt,
			row.AmountTons,
			row.MaterialSiteID,
			row.MaterialTypeID,
			truncateString(row.Description, 40),
		)
	}
	return w.Flush()
}
