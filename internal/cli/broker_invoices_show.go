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

type brokerInvoicesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type brokerInvoiceDetails struct {
	ID                                string   `json:"id"`
	Status                            string   `json:"status,omitempty"`
	InvoiceDate                       string   `json:"invoice_date,omitempty"`
	DueOn                             string   `json:"due_on,omitempty"`
	AdjustmentAmount                  string   `json:"adjustment_amount,omitempty"`
	CurrencyCode                      string   `json:"currency_code,omitempty"`
	TotalAmount                       string   `json:"total_amount,omitempty"`
	TimeCardAmount                    string   `json:"time_card_amount,omitempty"`
	Notes                             string   `json:"notes,omitempty"`
	ExplicitBuyerName                 string   `json:"explicit_buyer_name,omitempty"`
	ExplicitBuyerAddress              string   `json:"explicit_buyer_address,omitempty"`
	ShiftDateMin                      string   `json:"shift_date_min,omitempty"`
	ShiftDateMax                      string   `json:"shift_date_max,omitempty"`
	BusinessUnitIDs                   []string `json:"business_unit_ids,omitempty"`
	CustomerIDs                       []string `json:"customer_ids,omitempty"`
	IsManagementServiceType           bool     `json:"is_management_service_type,omitempty"`
	OrganizationInvoicesBatchStatuses any      `json:"organization_invoices_batch_statuses,omitempty"`
	CustomerTimeCardAmount            string   `json:"customer_time_card_amount,omitempty"`
	CustomerTotalAmount               string   `json:"customer_total_amount,omitempty"`
	BrokerTimeCardAmount              string   `json:"broker_time_card_amount,omitempty"`
	BrokerTotalAmount                 string   `json:"broker_total_amount,omitempty"`
	CurrentRevisionNumber             string   `json:"current_revision_number,omitempty"`
	BuyerType                         string   `json:"buyer_type,omitempty"`
	BuyerID                           string   `json:"buyer_id,omitempty"`
	BuyerName                         string   `json:"buyer,omitempty"`
	SellerType                        string   `json:"seller_type,omitempty"`
	SellerID                          string   `json:"seller_id,omitempty"`
	SellerName                        string   `json:"seller,omitempty"`
	TimeCardIDs                       []string `json:"time_card_ids,omitempty"`
	InvoiceStatusChangeIDs            []string `json:"invoice_status_change_ids,omitempty"`
	ExternalIdentificationIDs         []string `json:"external_identification_ids,omitempty"`
	InvoiceOrganizationBatchStatusIDs []string `json:"invoice_organization_batch_status_ids,omitempty"`
}

func newBrokerInvoicesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show broker invoice details",
		Long: `Show the full details of a broker invoice.

Arguments:
  <id>  The broker invoice ID (required).`,
		Example: `  # Show a broker invoice
  xbe view broker-invoices show 123

  # Output as JSON
  xbe view broker-invoices show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runBrokerInvoicesShow,
	}
	initBrokerInvoicesShowFlags(cmd)
	return cmd
}

func init() {
	brokerInvoicesCmd.AddCommand(newBrokerInvoicesShowCmd())
}

func initBrokerInvoicesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerInvoicesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseBrokerInvoicesShowOptions(cmd)
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
		return fmt.Errorf("broker invoice id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[broker-invoices]", strings.Join([]string{
		"invoice-date",
		"due-on",
		"status",
		"adjustment-amount",
		"currency-code",
		"total-amount",
		"time-card-amount",
		"notes",
		"explicit-buyer-name",
		"explicit-buyer-address",
		"shift-date-min",
		"shift-date-max",
		"business-unit-ids",
		"customer-ids",
		"is-management-service-type",
		"organization-invoices-batch-statuses",
		"customer-time-card-amount",
		"customer-total-amount",
		"broker-time-card-amount",
		"broker-total-amount",
		"current-revision-number",
		"buyer",
		"seller",
		"time-cards",
		"invoice-status-changes",
		"external-identifications",
		"invoice-organization-batch-statuses",
	}, ","))
	query.Set("include", "buyer,seller")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/broker-invoices/"+id, query)
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

	details := buildBrokerInvoiceDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderBrokerInvoiceDetails(cmd, details)
}

func parseBrokerInvoicesShowOptions(cmd *cobra.Command) (brokerInvoicesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerInvoicesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildBrokerInvoiceDetails(resp jsonAPISingleResponse) brokerInvoiceDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := brokerInvoiceDetails{
		ID:                                resp.Data.ID,
		Status:                            stringAttr(attrs, "status"),
		InvoiceDate:                       formatDate(stringAttr(attrs, "invoice-date")),
		DueOn:                             formatDate(stringAttr(attrs, "due-on")),
		AdjustmentAmount:                  stringAttr(attrs, "adjustment-amount"),
		CurrencyCode:                      stringAttr(attrs, "currency-code"),
		TotalAmount:                       stringAttr(attrs, "total-amount"),
		TimeCardAmount:                    stringAttr(attrs, "time-card-amount"),
		Notes:                             stringAttr(attrs, "notes"),
		ExplicitBuyerName:                 stringAttr(attrs, "explicit-buyer-name"),
		ExplicitBuyerAddress:              stringAttr(attrs, "explicit-buyer-address"),
		ShiftDateMin:                      formatDate(stringAttr(attrs, "shift-date-min")),
		ShiftDateMax:                      formatDate(stringAttr(attrs, "shift-date-max")),
		BusinessUnitIDs:                   stringSliceAttr(attrs, "business-unit-ids"),
		CustomerIDs:                       stringSliceAttr(attrs, "customer-ids"),
		IsManagementServiceType:           boolAttr(attrs, "is-management-service-type"),
		OrganizationInvoicesBatchStatuses: anyAttr(attrs, "organization-invoices-batch-statuses"),
		CustomerTimeCardAmount:            stringAttr(attrs, "customer-time-card-amount"),
		CustomerTotalAmount:               stringAttr(attrs, "customer-total-amount"),
		BrokerTimeCardAmount:              stringAttr(attrs, "broker-time-card-amount"),
		BrokerTotalAmount:                 stringAttr(attrs, "broker-total-amount"),
		CurrentRevisionNumber:             stringAttr(attrs, "current-revision-number"),
	}

	if rel, ok := resp.Data.Relationships["buyer"]; ok && rel.Data != nil {
		details.BuyerType = rel.Data.Type
		details.BuyerID = rel.Data.ID
		if buyer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.BuyerName = firstNonEmpty(
				stringAttr(buyer.Attributes, "company-name"),
				stringAttr(buyer.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["seller"]; ok && rel.Data != nil {
		details.SellerType = rel.Data.Type
		details.SellerID = rel.Data.ID
		if seller, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.SellerName = firstNonEmpty(
				stringAttr(seller.Attributes, "company-name"),
				stringAttr(seller.Attributes, "name"),
			)
		}
	}

	if rel, ok := resp.Data.Relationships["time-cards"]; ok {
		details.TimeCardIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["invoice-status-changes"]; ok {
		details.InvoiceStatusChangeIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["external-identifications"]; ok {
		details.ExternalIdentificationIDs = relationshipIDList(rel)
	}
	if rel, ok := resp.Data.Relationships["invoice-organization-batch-statuses"]; ok {
		details.InvoiceOrganizationBatchStatusIDs = relationshipIDList(rel)
	}

	return details
}

func renderBrokerInvoiceDetails(cmd *cobra.Command, details brokerInvoiceDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.InvoiceDate != "" {
		fmt.Fprintf(out, "Invoice Date: %s\n", details.InvoiceDate)
	}
	if details.DueOn != "" {
		fmt.Fprintf(out, "Due On: %s\n", details.DueOn)
	}
	if details.AdjustmentAmount != "" {
		fmt.Fprintf(out, "Adjustment Amount: %s\n", details.AdjustmentAmount)
	}
	if details.CurrencyCode != "" {
		fmt.Fprintf(out, "Currency Code: %s\n", details.CurrencyCode)
	}
	if details.TotalAmount != "" {
		fmt.Fprintf(out, "Total Amount: %s\n", details.TotalAmount)
	}
	if details.TimeCardAmount != "" {
		fmt.Fprintf(out, "Time Card Amount: %s\n", details.TimeCardAmount)
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	if details.ExplicitBuyerName != "" {
		fmt.Fprintf(out, "Explicit Buyer Name: %s\n", details.ExplicitBuyerName)
	}
	if details.ExplicitBuyerAddress != "" {
		fmt.Fprintf(out, "Explicit Buyer Address: %s\n", details.ExplicitBuyerAddress)
	}
	if details.ShiftDateMin != "" || details.ShiftDateMax != "" {
		fmt.Fprintf(out, "Shift Date Window: %s - %s\n", details.ShiftDateMin, details.ShiftDateMax)
	}
	if len(details.BusinessUnitIDs) > 0 {
		fmt.Fprintf(out, "Business Units: %s\n", strings.Join(details.BusinessUnitIDs, ", "))
	}
	if len(details.CustomerIDs) > 0 {
		fmt.Fprintf(out, "Customer IDs: %s\n", strings.Join(details.CustomerIDs, ", "))
	}
	fmt.Fprintf(out, "Management Service Type: %t\n", details.IsManagementServiceType)
	if formatted := formatAnyJSON(details.OrganizationInvoicesBatchStatuses); formatted != "" {
		fmt.Fprintln(out, "Organization Invoice Batch Statuses:")
		fmt.Fprintln(out, formatted)
	}
	if details.CustomerTimeCardAmount != "" {
		fmt.Fprintf(out, "Customer Time Card Amount: %s\n", details.CustomerTimeCardAmount)
	}
	if details.CustomerTotalAmount != "" {
		fmt.Fprintf(out, "Customer Total Amount: %s\n", details.CustomerTotalAmount)
	}
	if details.BrokerTimeCardAmount != "" {
		fmt.Fprintf(out, "Broker Time Card Amount: %s\n", details.BrokerTimeCardAmount)
	}
	if details.BrokerTotalAmount != "" {
		fmt.Fprintf(out, "Broker Total Amount: %s\n", details.BrokerTotalAmount)
	}
	if details.CurrentRevisionNumber != "" {
		fmt.Fprintf(out, "Current Revision: %s\n", details.CurrentRevisionNumber)
	}
	if details.BuyerType != "" || details.BuyerID != "" || details.BuyerName != "" {
		buyerLabel := formatRelated(details.BuyerName, formatPolymorphic(details.BuyerType, details.BuyerID))
		fmt.Fprintf(out, "Buyer: %s\n", buyerLabel)
	}
	if details.SellerType != "" || details.SellerID != "" || details.SellerName != "" {
		sellerLabel := formatRelated(details.SellerName, formatPolymorphic(details.SellerType, details.SellerID))
		fmt.Fprintf(out, "Seller: %s\n", sellerLabel)
	}
	if len(details.TimeCardIDs) > 0 {
		fmt.Fprintf(out, "Time Cards: %s\n", strings.Join(details.TimeCardIDs, ", "))
	}
	if len(details.InvoiceStatusChangeIDs) > 0 {
		fmt.Fprintf(out, "Invoice Status Changes: %s\n", strings.Join(details.InvoiceStatusChangeIDs, ", "))
	}
	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintf(out, "External Identifications: %s\n", strings.Join(details.ExternalIdentificationIDs, ", "))
	}
	if len(details.InvoiceOrganizationBatchStatusIDs) > 0 {
		fmt.Fprintf(out, "Invoice Organization Batch Statuses: %s\n", strings.Join(details.InvoiceOrganizationBatchStatusIDs, ", "))
	}

	return nil
}
