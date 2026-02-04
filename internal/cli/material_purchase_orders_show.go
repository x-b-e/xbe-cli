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

type materialPurchaseOrdersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type materialPurchaseOrderReleaseSummary struct {
	ID       string `json:"id"`
	Status   string `json:"status,omitempty"`
	Quantity string `json:"quantity,omitempty"`
}

type materialPurchaseOrderDetails struct {
	ID                        string                                `json:"id"`
	Status                    string                                `json:"status"`
	IsManagingRedemption      bool                                  `json:"is_managing_redemption"`
	Quantity                  string                                `json:"quantity"`
	TransactionAtMin          string                                `json:"transaction_at_min,omitempty"`
	TransactionAtMax          string                                `json:"transaction_at_max,omitempty"`
	ExternalPurchaseOrderID   string                                `json:"external_purchase_order_id,omitempty"`
	ExternalSalesOrderID      string                                `json:"external_sales_order_id,omitempty"`
	Broker                    string                                `json:"broker,omitempty"`
	BrokerID                  string                                `json:"broker_id,omitempty"`
	MaterialSupplier          string                                `json:"material_supplier,omitempty"`
	MaterialSupplierID        string                                `json:"material_supplier_id,omitempty"`
	Customer                  string                                `json:"customer,omitempty"`
	CustomerID                string                                `json:"customer_id,omitempty"`
	MaterialType              string                                `json:"material_type,omitempty"`
	MaterialTypeID            string                                `json:"material_type_id,omitempty"`
	MaterialSite              string                                `json:"material_site,omitempty"`
	MaterialSiteID            string                                `json:"material_site_id,omitempty"`
	JobSite                   string                                `json:"job_site,omitempty"`
	JobSiteID                 string                                `json:"job_site_id,omitempty"`
	UnitOfMeasure             string                                `json:"unit_of_measure,omitempty"`
	UnitOfMeasureID           string                                `json:"unit_of_measure_id,omitempty"`
	Releases                  []materialPurchaseOrderReleaseSummary `json:"releases,omitempty"`
	ExternalIdentificationIDs []string                              `json:"external_identification_ids,omitempty"`
	FileAttachmentIDs         []string                              `json:"file_attachment_ids,omitempty"`
}

func newMaterialPurchaseOrdersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show material purchase order details",
		Long: `Show the full details of a material purchase order.

Includes order attributes, relationships, and release references.

Arguments:
  <id>    Material purchase order ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # View a material purchase order
  xbe view material-purchase-orders show 123

  # Output as JSON
  xbe view material-purchase-orders show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runMaterialPurchaseOrdersShow,
	}
	initMaterialPurchaseOrdersShowFlags(cmd)
	return cmd
}

func init() {
	materialPurchaseOrdersCmd.AddCommand(newMaterialPurchaseOrdersShowCmd())
}

func initMaterialPurchaseOrdersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runMaterialPurchaseOrdersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseMaterialPurchaseOrdersShowOptions(cmd)
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
		return fmt.Errorf("material purchase order id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[material-purchase-orders]", "status,is-managing-redemption,transaction-at-min,transaction-at-max,quantity,external-purchase-order-id,external-sales-order-id,broker,material-supplier,customer,material-site,material-type,job-site,unit-of-measure,releases,external-identifications,file-attachments")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[material-sites]", "name")
	query.Set("fields[material-types]", "name,display-name")
	query.Set("fields[job-sites]", "name")
	query.Set("fields[unit-of-measures]", "name,abbreviation")
	query.Set("fields[material-purchase-order-releases]", "status,quantity")
	query.Set("include", "broker,material-supplier,customer,material-site,material-type,job-site,unit-of-measure,releases")

	body, _, err := client.Get(cmd.Context(), "/v1/material-purchase-orders/"+id, query)
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

	details := buildMaterialPurchaseOrderDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderMaterialPurchaseOrderDetails(cmd, details)
}

func parseMaterialPurchaseOrdersShowOptions(cmd *cobra.Command) (materialPurchaseOrdersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	noAuth, _ := cmd.Flags().GetBool("no-auth")

	return materialPurchaseOrdersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildMaterialPurchaseOrderDetails(resp jsonAPISingleResponse) materialPurchaseOrderDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := materialPurchaseOrderDetails{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		IsManagingRedemption:    boolAttr(attrs, "is-managing-redemption"),
		Quantity:                strings.TrimSpace(stringAttr(attrs, "quantity")),
		TransactionAtMin:        stringAttr(attrs, "transaction-at-min"),
		TransactionAtMax:        stringAttr(attrs, "transaction-at-max"),
		ExternalPurchaseOrderID: stringAttr(attrs, "external-purchase-order-id"),
		ExternalSalesOrderID:    stringAttr(attrs, "external-sales-order-id"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Broker = stringAttr(broker.Attributes, "company-name")
		}
	}
	if rel, ok := resource.Relationships["material-supplier"]; ok && rel.Data != nil {
		details.MaterialSupplierID = rel.Data.ID
		if supplier, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSupplier = firstNonEmpty(
				stringAttr(supplier.Attributes, "company-name"),
				stringAttr(supplier.Attributes, "name"),
			)
		}
	}
	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		details.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Customer = firstNonEmpty(
				stringAttr(customer.Attributes, "company-name"),
				stringAttr(customer.Attributes, "name"),
			)
		}
	}
	if rel, ok := resource.Relationships["material-type"]; ok && rel.Data != nil {
		details.MaterialTypeID = rel.Data.ID
		if mt, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialType = materialTypeLabel(mt.Attributes)
		}
	}
	if rel, ok := resource.Relationships["material-site"]; ok && rel.Data != nil {
		details.MaterialSiteID = rel.Data.ID
		if ms, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MaterialSite = stringAttr(ms.Attributes, "name")
		}
	}
	if rel, ok := resource.Relationships["job-site"]; ok && rel.Data != nil {
		details.JobSiteID = rel.Data.ID
		if js, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.JobSite = stringAttr(js.Attributes, "name")
		}
	}
	if rel, ok := resource.Relationships["unit-of-measure"]; ok && rel.Data != nil {
		details.UnitOfMeasureID = rel.Data.ID
		if uom, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UnitOfMeasure = unitOfMeasureLabel(uom.Attributes)
		}
	}

	if rel, ok := resource.Relationships["releases"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				summary := materialPurchaseOrderReleaseSummary{ID: ref.ID}
				if release, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					summary.Status = stringAttr(release.Attributes, "status")
					summary.Quantity = strings.TrimSpace(stringAttr(release.Attributes, "quantity"))
				}
				details.Releases = append(details.Releases, summary)
			}
		}
	}

	if rel, ok := resource.Relationships["external-identifications"]; ok && rel.raw != nil {
		details.ExternalIdentificationIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["file-attachments"]; ok && rel.raw != nil {
		details.FileAttachmentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderMaterialPurchaseOrderDetails(cmd *cobra.Command, details materialPurchaseOrderDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Status: %s\n", details.Status)
	fmt.Fprintf(out, "Managing Redemption: %t\n", details.IsManagingRedemption)
	if details.Quantity != "" {
		if details.UnitOfMeasure != "" {
			fmt.Fprintf(out, "Quantity: %s %s\n", details.Quantity, details.UnitOfMeasure)
		} else {
			fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
		}
	}
	if details.TransactionAtMin != "" {
		fmt.Fprintf(out, "Transaction At Min: %s\n", details.TransactionAtMin)
	}
	if details.TransactionAtMax != "" {
		fmt.Fprintf(out, "Transaction At Max: %s\n", details.TransactionAtMax)
	}
	if details.ExternalPurchaseOrderID != "" {
		fmt.Fprintf(out, "External Purchase Order ID: %s\n", details.ExternalPurchaseOrderID)
	}
	if details.ExternalSalesOrderID != "" {
		fmt.Fprintf(out, "External Sales Order ID: %s\n", details.ExternalSalesOrderID)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Relationships:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	if details.Broker != "" {
		fmt.Fprintf(out, "  Broker: %s (%s)\n", details.Broker, details.BrokerID)
	} else if details.BrokerID != "" {
		fmt.Fprintf(out, "  Broker: %s\n", details.BrokerID)
	}
	if details.MaterialSupplier != "" {
		fmt.Fprintf(out, "  Material Supplier: %s (%s)\n", details.MaterialSupplier, details.MaterialSupplierID)
	} else if details.MaterialSupplierID != "" {
		fmt.Fprintf(out, "  Material Supplier: %s\n", details.MaterialSupplierID)
	}
	if details.Customer != "" {
		fmt.Fprintf(out, "  Customer: %s (%s)\n", details.Customer, details.CustomerID)
	} else if details.CustomerID != "" {
		fmt.Fprintf(out, "  Customer: %s\n", details.CustomerID)
	}
	if details.MaterialType != "" {
		fmt.Fprintf(out, "  Material Type: %s (%s)\n", details.MaterialType, details.MaterialTypeID)
	} else if details.MaterialTypeID != "" {
		fmt.Fprintf(out, "  Material Type: %s\n", details.MaterialTypeID)
	}
	if details.MaterialSite != "" {
		fmt.Fprintf(out, "  Material Site: %s (%s)\n", details.MaterialSite, details.MaterialSiteID)
	} else if details.MaterialSiteID != "" {
		fmt.Fprintf(out, "  Material Site: %s\n", details.MaterialSiteID)
	}
	if details.JobSite != "" {
		fmt.Fprintf(out, "  Job Site: %s (%s)\n", details.JobSite, details.JobSiteID)
	} else if details.JobSiteID != "" {
		fmt.Fprintf(out, "  Job Site: %s\n", details.JobSiteID)
	}
	if details.UnitOfMeasureID != "" {
		if details.UnitOfMeasure != "" {
			fmt.Fprintf(out, "  Unit Of Measure: %s (%s)\n", details.UnitOfMeasure, details.UnitOfMeasureID)
		} else {
			fmt.Fprintf(out, "  Unit Of Measure: %s\n", details.UnitOfMeasureID)
		}
	}

	if len(details.Releases) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Releases:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, release := range details.Releases {
			line := release.ID
			if release.Status != "" {
				line = fmt.Sprintf("%s (%s)", line, release.Status)
			}
			if release.Quantity != "" {
				line = fmt.Sprintf("%s qty=%s", line, release.Quantity)
			}
			fmt.Fprintf(out, "  %s\n", line)
		}
	}

	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "External Identifications: %s\n", strings.Join(details.ExternalIdentificationIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "File Attachments: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}

	return nil
}
