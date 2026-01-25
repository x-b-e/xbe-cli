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

type materialPurchaseOrderReleaseRedemptionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialPurchaseOrderReleaseRedemptionDetails struct {
	ID                              string `json:"id"`
	TicketNumber                    string `json:"ticket_number,omitempty"`
	ReleaseID                       string `json:"release_id,omitempty"`
	ReleaseNumber                   string `json:"release_number,omitempty"`
	ReleaseStatus                   string `json:"release_status,omitempty"`
	ReleaseQuantity                 any    `json:"release_quantity,omitempty"`
	PurchaseOrderID                 string `json:"purchase_order_id,omitempty"`
	PurchaseOrderStatus             string `json:"purchase_order_status,omitempty"`
	PurchaseOrderQuantity           any    `json:"purchase_order_quantity,omitempty"`
	PurchaseOrderTransactionAtMin   string `json:"purchase_order_transaction_at_min,omitempty"`
	PurchaseOrderTransactionAtMax   string `json:"purchase_order_transaction_at_max,omitempty"`
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

func newMaterialPurchaseOrderReleaseRedemptionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material purchase order release redemption details",
		Long: `Show the full details of a material purchase order release redemption.

Includes the linked release, purchase order, material transaction, and
associated parties.

Arguments:
  <id>  The release redemption ID (required).`,
		Example: `  # Show a release redemption
  xbe view material-purchase-order-release-redemptions show 123

  # Output as JSON
  xbe view material-purchase-order-release-redemptions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialPurchaseOrderReleaseRedemptionsShow,
	}
	initMaterialPurchaseOrderReleaseRedemptionsShowFlags(cmd)
	return cmd
}

func init() {
	materialPurchaseOrderReleaseRedemptionsCmd.AddCommand(newMaterialPurchaseOrderReleaseRedemptionsShowCmd())
}

func initMaterialPurchaseOrderReleaseRedemptionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialPurchaseOrderReleaseRedemptionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseMaterialPurchaseOrderReleaseRedemptionsShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("material purchase order release redemption id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-purchase-order-release-redemptions]", "ticket-number,release,purchase-order,material-transaction,trucker,driver,tender-job-schedule-shift,broker,material-supplier,material-type")
	query.Set("include", "release,purchase-order,material-transaction,trucker,driver,tender-job-schedule-shift,broker,material-supplier,material-type")
	query.Set("fields[material-purchase-order-releases]", "status,quantity")
	query.Set("fields[material-purchase-orders]", "status,quantity,transaction-at-min,transaction-at-max")
	query.Set("fields[material-transactions]", "ticket-number")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[users]", "name")
	query.Set("fields[tender-job-schedule-shifts]", "job-schedule-shift")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[material-types]", "name,display-name")

	body, _, err := client.Get(cmd.Context(), "/v1/material-purchase-order-release-redemptions/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildMaterialPurchaseOrderReleaseRedemptionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialPurchaseOrderReleaseRedemptionDetails(cmd, details)
}

func parseMaterialPurchaseOrderReleaseRedemptionsShowOptions(cmd *cobra.Command) (materialPurchaseOrderReleaseRedemptionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return materialPurchaseOrderReleaseRedemptionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialPurchaseOrderReleaseRedemptionDetails(resp jsonAPISingleResponse) materialPurchaseOrderReleaseRedemptionDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialPurchaseOrderReleaseRedemptionDetails{
		ID:           resp.Data.ID,
		TicketNumber: stringAttr(resp.Data.Attributes, "ticket-number"),
	}

	if rel, ok := resp.Data.Relationships["release"]; ok && rel.Data != nil {
		details.ReleaseID = rel.Data.ID
		if release, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := release.Attributes
			details.ReleaseNumber = stringAttr(attrs, "release-number")
			details.ReleaseStatus = stringAttr(attrs, "status")
			details.ReleaseQuantity = attrs["quantity"]
		}
	}

	if rel, ok := resp.Data.Relationships["purchase-order"]; ok && rel.Data != nil {
		details.PurchaseOrderID = rel.Data.ID
		if po, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := po.Attributes
			details.PurchaseOrderStatus = stringAttr(attrs, "status")
			details.PurchaseOrderQuantity = attrs["quantity"]
			details.PurchaseOrderTransactionAtMin = stringAttr(attrs, "transaction-at-min")
			details.PurchaseOrderTransactionAtMax = stringAttr(attrs, "transaction-at-max")
		}
	}

	if rel, ok := resp.Data.Relationships["material-transaction"]; ok && rel.Data != nil {
		details.MaterialTransactionID = rel.Data.ID
		if mtxn, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTransactionTicketNumber = stringAttr(mtxn.Attributes, "ticket-number")
		}
	}

	if rel, ok := resp.Data.Relationships["trucker"]; ok && rel.Data != nil {
		details.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.TruckerName = strings.TrimSpace(stringAttr(trucker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["driver"]; ok && rel.Data != nil {
		details.DriverID = rel.Data.ID
		if driver, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DriverName = strings.TrimSpace(stringAttr(driver.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BrokerName = strings.TrimSpace(stringAttr(broker.Attributes, "company-name"))
		}
	}

	if rel, ok := resp.Data.Relationships["material-supplier"]; ok && rel.Data != nil {
		details.MaterialSupplierID = rel.Data.ID
		if supplier, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSupplierName = strings.TrimSpace(firstNonEmpty(
				stringAttr(supplier.Attributes, "company-name"),
				stringAttr(supplier.Attributes, "name"),
			))
		}
	}

	if rel, ok := resp.Data.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if materialType, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialTypeName = strings.TrimSpace(firstNonEmpty(
				stringAttr(materialType.Attributes, "display-name"),
				stringAttr(materialType.Attributes, "name"),
			))
		}
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}

	return details
}

func renderMaterialPurchaseOrderReleaseRedemptionDetails(cmd *cobra.Command, details materialPurchaseOrderReleaseRedemptionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TicketNumber != "" {
		fmt.Fprintf(out, "Ticket Number: %s\n", details.TicketNumber)
	}
	if details.ReleaseID != "" {
		fmt.Fprintf(out, "Release ID: %s\n", details.ReleaseID)
	}
	if details.ReleaseNumber != "" {
		fmt.Fprintf(out, "Release Number: %s\n", details.ReleaseNumber)
	}
	if details.ReleaseStatus != "" {
		fmt.Fprintf(out, "Release Status: %s\n", details.ReleaseStatus)
	}
	if details.ReleaseQuantity != nil {
		fmt.Fprintf(out, "Release Quantity: %v\n", details.ReleaseQuantity)
	}
	if details.PurchaseOrderID != "" {
		fmt.Fprintf(out, "Purchase Order ID: %s\n", details.PurchaseOrderID)
	}
	if details.PurchaseOrderStatus != "" {
		fmt.Fprintf(out, "Purchase Order Status: %s\n", details.PurchaseOrderStatus)
	}
	if details.PurchaseOrderQuantity != nil {
		fmt.Fprintf(out, "Purchase Order Quantity: %v\n", details.PurchaseOrderQuantity)
	}
	if details.PurchaseOrderTransactionAtMin != "" {
		fmt.Fprintf(out, "Purchase Order Transaction At Min: %s\n", details.PurchaseOrderTransactionAtMin)
	}
	if details.PurchaseOrderTransactionAtMax != "" {
		fmt.Fprintf(out, "Purchase Order Transaction At Max: %s\n", details.PurchaseOrderTransactionAtMax)
	}
	if details.MaterialTransactionID != "" {
		fmt.Fprintf(out, "Material Transaction ID: %s\n", details.MaterialTransactionID)
	}
	if details.MaterialTransactionTicketNumber != "" {
		fmt.Fprintf(out, "Material Transaction Ticket: %s\n", details.MaterialTransactionTicketNumber)
	}
	if details.MaterialSupplierID != "" {
		fmt.Fprintf(out, "Material Supplier ID: %s\n", details.MaterialSupplierID)
	}
	if details.MaterialSupplierName != "" {
		fmt.Fprintf(out, "Material Supplier: %s\n", details.MaterialSupplierName)
	}
	if details.MaterialTypeID != "" {
		fmt.Fprintf(out, "Material Type ID: %s\n", details.MaterialTypeID)
	}
	if details.MaterialTypeName != "" {
		fmt.Fprintf(out, "Material Type: %s\n", details.MaterialTypeName)
	}
	if details.TruckerID != "" {
		fmt.Fprintf(out, "Trucker ID: %s\n", details.TruckerID)
	}
	if details.TruckerName != "" {
		fmt.Fprintf(out, "Trucker: %s\n", details.TruckerName)
	}
	if details.DriverID != "" {
		fmt.Fprintf(out, "Driver ID: %s\n", details.DriverID)
	}
	if details.DriverName != "" {
		fmt.Fprintf(out, "Driver: %s\n", details.DriverName)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.BrokerName != "" {
		fmt.Fprintf(out, "Broker: %s\n", details.BrokerName)
	}
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift ID: %s\n", details.TenderJobScheduleShiftID)
	}

	return nil
}
