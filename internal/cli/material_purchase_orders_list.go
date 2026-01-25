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

type materialPurchaseOrdersListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	Status                    string
	Broker                    string
	MaterialSupplier          string
	Customer                  string
	MaterialSite              string
	MaterialType              string
	JobSite                   string
	UnitOfMeasure             string
	TransactionAtMin          string
	TransactionAtMax          string
	Quantity                  string
	BaseMaterialType          string
	ExternalPurchaseOrderID   string
	ExternalSalesOrderID      string
	ExternalIdentificationVal string
}

type materialPurchaseOrderRow struct {
	ID                      string `json:"id"`
	Status                  string `json:"status"`
	Quantity                string `json:"quantity"`
	UnitOfMeasure           string `json:"unit_of_measure,omitempty"`
	MaterialType            string `json:"material_type,omitempty"`
	MaterialSupplier        string `json:"material_supplier,omitempty"`
	Customer                string `json:"customer,omitempty"`
	JobSite                 string `json:"job_site,omitempty"`
	ExternalPurchaseOrderID string `json:"external_purchase_order_id,omitempty"`
	ExternalSalesOrderID    string `json:"external_sales_order_id,omitempty"`
}

func newMaterialPurchaseOrdersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material purchase orders",
		Long: `List material purchase orders with filtering and pagination.

Output Columns:
  ID        Material purchase order identifier
  STATUS    Workflow status
  QTY       Ordered quantity
  UOM       Unit of measure
  MATERIAL  Material type
  SUPPLIER  Material supplier
  CUSTOMER  Customer (if set)
  JOB SITE  Job site (if set)
  PO ID     External purchase order ID (if set)
  SO ID     External sales order ID (if set)

Filters:
  --status                     Filter by status (editing,approved,closed)
  --broker                     Filter by broker ID (comma-separated)
  --material-supplier           Filter by material supplier ID (comma-separated)
  --customer                   Filter by customer ID (comma-separated)
  --material-site              Filter by material site ID (comma-separated)
  --material-type              Filter by material type ID (comma-separated)
  --job-site                   Filter by job site ID (comma-separated)
  --unit-of-measure            Filter by unit of measure ID (comma-separated)
  --transaction-at-min         Filter by minimum transaction datetime (ISO 8601)
  --transaction-at-max         Filter by maximum transaction datetime (ISO 8601)
  --quantity                   Filter by quantity
  --base-material-type         Filter by base material type ID (comma-separated)
  --external-purchase-order-id Filter by external purchase order ID
  --external-sales-order-id    Filter by external sales order ID
  --external-identification-value Filter by external identification value

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List material purchase orders
  xbe view material-purchase-orders list

  # Filter by status
  xbe view material-purchase-orders list --status approved

  # Filter by supplier and job site
  xbe view material-purchase-orders list --material-supplier 123 --job-site 456

  # Filter by external purchase order ID
  xbe view material-purchase-orders list --external-purchase-order-id PO-1001

  # Output as JSON
  xbe view material-purchase-orders list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialPurchaseOrdersList,
	}
	initMaterialPurchaseOrdersListFlags(cmd)
	return cmd
}

func init() {
	materialPurchaseOrdersCmd.AddCommand(newMaterialPurchaseOrdersListCmd())
}

func initMaterialPurchaseOrdersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status (editing,approved,closed)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID (comma-separated)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated)")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated)")
	cmd.Flags().String("material-type", "", "Filter by material type ID (comma-separated)")
	cmd.Flags().String("job-site", "", "Filter by job site ID (comma-separated)")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID (comma-separated)")
	cmd.Flags().String("transaction-at-min", "", "Filter by minimum transaction datetime (ISO 8601)")
	cmd.Flags().String("transaction-at-max", "", "Filter by maximum transaction datetime (ISO 8601)")
	cmd.Flags().String("quantity", "", "Filter by quantity")
	cmd.Flags().String("base-material-type", "", "Filter by base material type ID (comma-separated)")
	cmd.Flags().String("external-purchase-order-id", "", "Filter by external purchase order ID")
	cmd.Flags().String("external-sales-order-id", "", "Filter by external sales order ID")
	cmd.Flags().String("external-identification-value", "", "Filter by external identification value")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialPurchaseOrdersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialPurchaseOrdersListOptions(cmd)
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
	query.Set("fields[material-purchase-orders]", "status,quantity,is-managing-redemption,transaction-at-min,transaction-at-max,external-purchase-order-id,external-sales-order-id,broker,material-supplier,customer,material-site,material-type,job-site,unit-of-measure")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("include", "broker,material-supplier,customer,material-site,material-type,job-site,unit-of-measure")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[material-site]", opts.MaterialSite)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[job-site]", opts.JobSite)
	setFilterIfPresent(query, "filter[unit-of-measure]", opts.UnitOfMeasure)
	setFilterIfPresent(query, "filter[transaction-at-min]", opts.TransactionAtMin)
	setFilterIfPresent(query, "filter[transaction-at-max]", opts.TransactionAtMax)
	setFilterIfPresent(query, "filter[quantity]", opts.Quantity)
	setFilterIfPresent(query, "filter[base-material-type]", opts.BaseMaterialType)
	setFilterIfPresent(query, "filter[external-purchase-order-id]", opts.ExternalPurchaseOrderID)
	setFilterIfPresent(query, "filter[external-sales-order-id]", opts.ExternalSalesOrderID)
	setFilterIfPresent(query, "filter[external-identification-value]", opts.ExternalIdentificationVal)

	body, _, err := client.Get(cmd.Context(), "/v1/material-purchase-orders", query)
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

	rows := buildMaterialPurchaseOrderRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialPurchaseOrdersTable(cmd, rows)
}

func parseMaterialPurchaseOrdersListOptions(cmd *cobra.Command) (materialPurchaseOrdersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	broker, _ := cmd.Flags().GetString("broker")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	customer, _ := cmd.Flags().GetString("customer")
	materialSite, _ := cmd.Flags().GetString("material-site")
	materialType, _ := cmd.Flags().GetString("material-type")
	jobSite, _ := cmd.Flags().GetString("job-site")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	transactionAtMin, _ := cmd.Flags().GetString("transaction-at-min")
	transactionAtMax, _ := cmd.Flags().GetString("transaction-at-max")
	quantity, _ := cmd.Flags().GetString("quantity")
	baseMaterialType, _ := cmd.Flags().GetString("base-material-type")
	externalPurchaseOrderID, _ := cmd.Flags().GetString("external-purchase-order-id")
	externalSalesOrderID, _ := cmd.Flags().GetString("external-sales-order-id")
	externalIdentificationValue, _ := cmd.Flags().GetString("external-identification-value")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialPurchaseOrdersListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		Status:                    status,
		Broker:                    broker,
		MaterialSupplier:          materialSupplier,
		Customer:                  customer,
		MaterialSite:              materialSite,
		MaterialType:              materialType,
		JobSite:                   jobSite,
		UnitOfMeasure:             unitOfMeasure,
		TransactionAtMin:          transactionAtMin,
		TransactionAtMax:          transactionAtMax,
		Quantity:                  quantity,
		BaseMaterialType:          baseMaterialType,
		ExternalPurchaseOrderID:   externalPurchaseOrderID,
		ExternalSalesOrderID:      externalSalesOrderID,
		ExternalIdentificationVal: externalIdentificationValue,
	}, nil
}

func buildMaterialPurchaseOrderRows(resp jsonAPIResponse) []materialPurchaseOrderRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialPurchaseOrderRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := materialPurchaseOrderRow{
			ID:                      resource.ID,
			Status:                  stringAttr(attrs, "status"),
			Quantity:                strings.TrimSpace(stringAttr(attrs, "quantity")),
			ExternalPurchaseOrderID: stringAttr(attrs, "external-purchase-order-id"),
			ExternalSalesOrderID:    stringAttr(attrs, "external-sales-order-id"),
		}

		if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
			if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.UnitOfMeasure = unitOfMeasureLabel(uom.Attributes)
			}
		}
		if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
			if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialType = materialTypeLabel(mt.Attributes)
			}
		}
		if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
			if supplier, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.MaterialSupplier = firstNonEmpty(
					stringAttr(supplier.Attributes, "company-name"),
					stringAttr(supplier.Attributes, "name"),
				)
			}
		}
		if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
			if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Customer = firstNonEmpty(
					stringAttr(customer.Attributes, "company-name"),
					stringAttr(customer.Attributes, "name"),
				)
			}
		}
		if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
			if jobSite, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.JobSite = stringAttr(jobSite.Attributes, "name")
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderMaterialPurchaseOrdersTable(cmd *cobra.Command, rows []materialPurchaseOrderRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No material purchase orders found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tQTY\tUOM\tMATERIAL\tSUPPLIER\tCUSTOMER\tJOB SITE\tPO ID\tSO ID")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Quantity,
			truncateString(row.UnitOfMeasure, 8),
			truncateString(row.MaterialType, 20),
			truncateString(row.MaterialSupplier, 20),
			truncateString(row.Customer, 20),
			truncateString(row.JobSite, 20),
			truncateString(row.ExternalPurchaseOrderID, 14),
			truncateString(row.ExternalSalesOrderID, 14),
		)
	}
	return writer.Flush()
}

func unitOfMeasureLabel(attrs map[string]any) string {
	abbr := strings.TrimSpace(stringAttr(attrs, "abbreviation"))
	if abbr != "" {
		return abbr
	}
	return strings.TrimSpace(stringAttr(attrs, "name"))
}

func materialTypeLabel(attrs map[string]any) string {
	display := strings.TrimSpace(stringAttr(attrs, "display-name"))
	if display != "" {
		return display
	}
	return strings.TrimSpace(stringAttr(attrs, "name"))
}
