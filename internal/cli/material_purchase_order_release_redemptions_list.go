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

type materialPurchaseOrderReleaseRedemptionsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	Release                  string
	TicketNumber             string
	MaterialTransaction      string
	Broker                   string
	BrokerID                 string
	Trucker                  string
	TruckerID                string
	Driver                   string
	DriverID                 string
	TenderJobScheduleShift   string
	TenderJobScheduleShiftID string
	MaterialSupplier         string
	MaterialSupplierID       string
	MaterialType             string
	MaterialTypeID           string
	PurchaseOrder            string
	PurchaseOrderID          string
}

type materialPurchaseOrderReleaseRedemptionRow struct {
	ID                              string `json:"id"`
	TicketNumber                    string `json:"ticket_number,omitempty"`
	ReleaseID                       string `json:"release_id,omitempty"`
	ReleaseNumber                   string `json:"release_number,omitempty"`
	ReleaseStatus                   string `json:"release_status,omitempty"`
	PurchaseOrderID                 string `json:"purchase_order_id,omitempty"`
	PurchaseOrderStatus             string `json:"purchase_order_status,omitempty"`
	MaterialTransactionID           string `json:"material_transaction_id,omitempty"`
	MaterialTransactionTicketNumber string `json:"material_transaction_ticket_number,omitempty"`
	TruckerID                       string `json:"trucker_id,omitempty"`
	TruckerName                     string `json:"trucker_name,omitempty"`
	DriverID                        string `json:"driver_id,omitempty"`
	DriverName                      string `json:"driver_name,omitempty"`
	BrokerID                        string `json:"broker_id,omitempty"`
	BrokerName                      string `json:"broker_name,omitempty"`
	MaterialSupplierID              string `json:"material_supplier_id,omitempty"`
	MaterialSupplierName            string `json:"material_supplier_name,omitempty"`
	MaterialTypeID                  string `json:"material_type_id,omitempty"`
	MaterialTypeName                string `json:"material_type_name,omitempty"`
	TenderJobScheduleShiftID        string `json:"tender_job_schedule_shift_id,omitempty"`
}

func newMaterialPurchaseOrderReleaseRedemptionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List material purchase order release redemptions",
		Long: `List material purchase order release redemptions with filtering and pagination.

Output Columns:
  ID           Redemption identifier
  TICKET       Ticket number
  RELEASE      Release number or ID
  ORDER        Purchase order reference
  MTXN         Material transaction ticket or ID
  SUPPLIER     Material supplier name
  MATERIAL     Material type
  TRUCKER      Trucker company
  DRIVER       Driver name

Filters:
  --release                      Filter by release ID
  --ticket-number                Filter by ticket number
  --material-transaction         Filter by material transaction ID
  --broker                        Filter by broker ID
  --broker-id                     Filter by broker ID (joined)
  --trucker                       Filter by trucker ID
  --trucker-id                    Filter by trucker ID (joined)
  --driver                        Filter by driver ID
  --driver-id                     Filter by driver ID (joined)
  --tender-job-schedule-shift     Filter by tender job schedule shift ID
  --tender-job-schedule-shift-id  Filter by tender job schedule shift ID (joined)
  --material-supplier             Filter by material supplier ID
  --material-supplier-id          Filter by material supplier ID (joined)
  --material-type                 Filter by material type ID
  --material-type-id              Filter by material type ID (joined)
  --purchase-order                Filter by purchase order ID
  --purchase-order-id             Filter by purchase order ID (joined)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List release redemptions
  xbe view material-purchase-order-release-redemptions list

  # Filter by release
  xbe view material-purchase-order-release-redemptions list --release 123

  # Filter by ticket number
  xbe view material-purchase-order-release-redemptions list --ticket-number T-100

  # Output as JSON
  xbe view material-purchase-order-release-redemptions list --json`,
		Args: cobra.NoArgs,
		RunE: runMaterialPurchaseOrderReleaseRedemptionsList,
	}
	initMaterialPurchaseOrderReleaseRedemptionsListFlags(cmd)
	return cmd
}

func init() {
	materialPurchaseOrderReleaseRedemptionsCmd.AddCommand(newMaterialPurchaseOrderReleaseRedemptionsListCmd())
}

func initMaterialPurchaseOrderReleaseRedemptionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("release", "", "Filter by release ID")
	cmd.Flags().String("ticket-number", "", "Filter by ticket number")
	cmd.Flags().String("material-transaction", "", "Filter by material transaction ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (joined)")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trucker-id", "", "Filter by trucker ID (joined)")
	cmd.Flags().String("driver", "", "Filter by driver ID")
	cmd.Flags().String("driver-id", "", "Filter by driver ID (joined)")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("tender-job-schedule-shift-id", "", "Filter by tender job schedule shift ID (joined)")
	cmd.Flags().String("material-supplier", "", "Filter by material supplier ID")
	cmd.Flags().String("material-supplier-id", "", "Filter by material supplier ID (joined)")
	cmd.Flags().String("material-type", "", "Filter by material type ID")
	cmd.Flags().String("material-type-id", "", "Filter by material type ID (joined)")
	cmd.Flags().String("purchase-order", "", "Filter by purchase order ID")
	cmd.Flags().String("purchase-order-id", "", "Filter by purchase order ID (joined)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialPurchaseOrderReleaseRedemptionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseMaterialPurchaseOrderReleaseRedemptionsListOptions(cmd)
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
	query.Set("fields[material-purchase-order-release-redemptions]", "ticket-number,release,purchase-order,material-transaction,trucker,driver,tender-job-schedule-shift,broker,material-supplier,material-type")
	query.Set("include", "release,purchase-order,material-transaction,trucker,driver,tender-job-schedule-shift,broker,material-supplier,material-type")
	query.Set("fields[material-purchase-order-releases]", "status,quantity")
	query.Set("fields[material-purchase-orders]", "status")
	query.Set("fields[material-transactions]", "ticket-number")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("fields[tender-job-schedule-shifts]", "job-schedule-shift")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[material-types]", "name,display-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[release]", opts.Release)
	setFilterIfPresent(query, "filter[ticket-number]", opts.TicketNumber)
	setFilterIfPresent(query, "filter[material-transaction]", opts.MaterialTransaction)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[broker-id]", opts.BrokerID)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trucker-id]", opts.TruckerID)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)
	setFilterIfPresent(query, "filter[driver-id]", opts.DriverID)
	setFilterIfPresent(query, "filter[tender-job-schedule-shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[tender-job-schedule-shift-id]", opts.TenderJobScheduleShiftID)
	setFilterIfPresent(query, "filter[material-supplier]", opts.MaterialSupplier)
	setFilterIfPresent(query, "filter[material-supplier-id]", opts.MaterialSupplierID)
	setFilterIfPresent(query, "filter[material-type]", opts.MaterialType)
	setFilterIfPresent(query, "filter[material-type-id]", opts.MaterialTypeID)
	setFilterIfPresent(query, "filter[purchase-order]", opts.PurchaseOrder)
	setFilterIfPresent(query, "filter[purchase-order-id]", opts.PurchaseOrderID)

	body, _, err := client.Get(cmd.Context(), "/v1/material-purchase-order-release-redemptions", query)
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

	rows := buildMaterialPurchaseOrderReleaseRedemptionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderMaterialPurchaseOrderReleaseRedemptionsTable(cmd, rows)
}

func parseMaterialPurchaseOrderReleaseRedemptionsListOptions(cmd *cobra.Command) (materialPurchaseOrderReleaseRedemptionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	release, _ := cmd.Flags().GetString("release")
	ticketNumber, _ := cmd.Flags().GetString("ticket-number")
	materialTransaction, _ := cmd.Flags().GetString("material-transaction")
	broker, _ := cmd.Flags().GetString("broker")
	brokerID, _ := cmd.Flags().GetString("broker-id")
	trucker, _ := cmd.Flags().GetString("trucker")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	driver, _ := cmd.Flags().GetString("driver")
	driverID, _ := cmd.Flags().GetString("driver-id")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	tenderJobScheduleShiftID, _ := cmd.Flags().GetString("tender-job-schedule-shift-id")
	materialSupplier, _ := cmd.Flags().GetString("material-supplier")
	materialSupplierID, _ := cmd.Flags().GetString("material-supplier-id")
	materialType, _ := cmd.Flags().GetString("material-type")
	materialTypeID, _ := cmd.Flags().GetString("material-type-id")
	purchaseOrder, _ := cmd.Flags().GetString("purchase-order")
	purchaseOrderID, _ := cmd.Flags().GetString("purchase-order-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialPurchaseOrderReleaseRedemptionsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		Release:                  release,
		TicketNumber:             ticketNumber,
		MaterialTransaction:      materialTransaction,
		Broker:                   broker,
		BrokerID:                 brokerID,
		Trucker:                  trucker,
		TruckerID:                truckerID,
		Driver:                   driver,
		DriverID:                 driverID,
		TenderJobScheduleShift:   tenderJobScheduleShift,
		TenderJobScheduleShiftID: tenderJobScheduleShiftID,
		MaterialSupplier:         materialSupplier,
		MaterialSupplierID:       materialSupplierID,
		MaterialType:             materialType,
		MaterialTypeID:           materialTypeID,
		PurchaseOrder:            purchaseOrder,
		PurchaseOrderID:          purchaseOrderID,
	}, nil
}

func buildMaterialPurchaseOrderReleaseRedemptionRows(resp jsonAPIResponse) []materialPurchaseOrderReleaseRedemptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]materialPurchaseOrderReleaseRedemptionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildMaterialPurchaseOrderReleaseRedemptionRow(resource, included))
	}
	return rows
}

func materialPurchaseOrderReleaseRedemptionRowFromSingle(resp jsonAPISingleResponse) materialPurchaseOrderReleaseRedemptionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	return buildMaterialPurchaseOrderReleaseRedemptionRow(resp.Data, included)
}

func buildMaterialPurchaseOrderReleaseRedemptionRow(resource jsonAPIResource, included map[string]jsonAPIResource) materialPurchaseOrderReleaseRedemptionRow {
	row := materialPurchaseOrderReleaseRedemptionRow{
		ID:           resource.ID,
		TicketNumber: stringAttr(resource.Attributes, "ticket-number"),
	}

	if rel, ok := resource.Relationships["release"]; ok && rel.Data != nil {
		row.ReleaseID = rel.Data.ID
		if release, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := release.Attributes
			row.ReleaseNumber = stringAttr(attrs, "release-number")
			row.ReleaseStatus = stringAttr(attrs, "status")
		}
	}

	if rel, ok := resource.Relationships["purchase-order"]; ok && rel.Data != nil {
		row.PurchaseOrderID = rel.Data.ID
		if po, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := po.Attributes
			row.PurchaseOrderStatus = stringAttr(attrs, "status")
		}
	}

	if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
		row.MaterialTransactionID = rel.Data.ID
		if mtxn, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialTransactionTicketNumber = stringAttr(mtxn.Attributes, "ticket-number")
		}
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = strings.TrimSpace(stringAttr(trucker.Attributes, "company-name"))
		}
	}

	if rel, ok := resource.Relationships["driver"]; ok && rel.Data != nil {
		row.DriverID = rel.Data.ID
		if driver, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.DriverName = strings.TrimSpace(stringAttr(driver.Attributes, "name"))
		}
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		row.MaterialSupplierID = rel.Data.ID
		if supplier, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialSupplierName = strings.TrimSpace(firstNonEmpty(
				stringAttr(supplier.Attributes, "company-name"),
				stringAttr(supplier.Attributes, "name"),
			))
		}
	}

	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		row.MaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.MaterialTypeName = strings.TrimSpace(firstNonEmpty(
				stringAttr(materialType.Attributes, "display-name"),
				stringAttr(materialType.Attributes, "name"),
			))
		}
	}

	return row
}

func renderMaterialPurchaseOrderReleaseRedemptionsTable(cmd *cobra.Command, rows []materialPurchaseOrderReleaseRedemptionRow) error {
	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tTICKET\tRELEASE\tORDER\tMTXN\tSUPPLIER\tMATERIAL\tTRUCKER\tDRIVER")
	for _, row := range rows {
		releaseLabel := firstNonEmpty(row.ReleaseNumber, row.ReleaseID)
		purchaseOrderLabel := row.PurchaseOrderID
		mtxnLabel := firstNonEmpty(row.MaterialTransactionTicketNumber, row.MaterialTransactionID)
		supplierLabel := firstNonEmpty(row.MaterialSupplierName, row.MaterialSupplierID)
		materialLabel := firstNonEmpty(row.MaterialTypeName, row.MaterialTypeID)
		truckerLabel := firstNonEmpty(row.TruckerName, row.TruckerID)
		driverLabel := firstNonEmpty(row.DriverName, row.DriverID)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TicketNumber,
			releaseLabel,
			purchaseOrderLabel,
			mtxnLabel,
			supplierLabel,
			materialLabel,
			truckerLabel,
			driverLabel,
		)
	}

	return w.Flush()
}
