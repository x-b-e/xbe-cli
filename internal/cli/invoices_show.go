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

type invoicesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type invoiceDetails struct {
	ID                                string            `json:"id"`
	Status                            string            `json:"status,omitempty"`
	InvoiceDate                       string            `json:"invoice_date,omitempty"`
	DueOn                             string            `json:"due_on,omitempty"`
	AdjustmentAmount                  string            `json:"adjustment_amount,omitempty"`
	CurrencyCode                      string            `json:"currency_code,omitempty"`
	TotalAmount                       string            `json:"total_amount,omitempty"`
	Notes                             string            `json:"notes,omitempty"`
	ExplicitBuyerName                 string            `json:"explicit_buyer_name,omitempty"`
	ExplicitBuyerAddress              string            `json:"explicit_buyer_address,omitempty"`
	ShiftDateMin                      string            `json:"shift_date_min,omitempty"`
	ShiftDateMax                      string            `json:"shift_date_max,omitempty"`
	BusinessUnitIDs                   []string          `json:"business_unit_ids,omitempty"`
	CustomerIDs                       []string          `json:"customer_ids,omitempty"`
	IsManagementServiceType           bool              `json:"is_management_service_type"`
	OrganizationInvoicesBatchStatuses map[string]string `json:"organization_invoices_batch_statuses,omitempty"`
	CustomerTimeCardAmount            string            `json:"customer_time_card_amount,omitempty"`
	CustomerTotalAmount               string            `json:"customer_total_amount,omitempty"`
	BrokerTimeCardAmount              string            `json:"broker_time_card_amount,omitempty"`
	BrokerTotalAmount                 string            `json:"broker_total_amount,omitempty"`
	CurrentRevisionNumber             string            `json:"current_revision_number,omitempty"`
	BuyerID                           string            `json:"buyer_id,omitempty"`
	BuyerType                         string            `json:"buyer_type,omitempty"`
	SellerID                          string            `json:"seller_id,omitempty"`
	SellerType                        string            `json:"seller_type,omitempty"`
	TimeCardIDs                       []string          `json:"time_card_ids,omitempty"`
	InvoiceStatusChangeIDs            []string          `json:"invoice_status_change_ids,omitempty"`
	ExternalIdentificationIDs         []string          `json:"external_identification_ids,omitempty"`
	InvoiceOrganizationBatchStatusIDs []string          `json:"invoice_organization_batch_status_ids,omitempty"`
}

func newInvoicesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show invoice details",
		Long: `Show the full details of an invoice.

Output Fields:
  ID
  Status
  Invoice Date
  Due On
  Adjustment Amount
  Currency Code
  Total Amount
  Notes
  Explicit Buyer Name
  Explicit Buyer Address
  Shift Date Min
  Shift Date Max
  Business Unit IDs
  Customer IDs
  Is Management Service Type
  Organization Invoice Batch Statuses
  Customer Time Card Amount
  Customer Total Amount
  Broker Time Card Amount
  Broker Total Amount
  Current Revision Number
  Buyer / Seller (type + ID)
  Time Card IDs
  Invoice Status Change IDs
  External Identification IDs
  Invoice Organization Batch Status IDs

Arguments:
  <id>    The invoice ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show invoice details
  xbe view invoices show 123

  # Get JSON output
  xbe view invoices show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runInvoicesShow,
	}
	initInvoicesShowFlags(cmd)
	return cmd
}

func init() {
	invoicesCmd.AddCommand(newInvoicesShowCmd())
}

func initInvoicesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoicesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseInvoicesShowOptions(cmd)
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
		return fmt.Errorf("invoice id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[invoices]", "invoice-date,due-on,status,adjustment-amount,currency-code,total-amount,notes,explicit-buyer-name,explicit-buyer-address,shift-date-min,shift-date-max,business-unit-ids,customer-ids,is-management-service-type,organization-invoices-batch-statuses,customer-time-card-amount,customer-total-amount,broker-time-card-amount,broker-total-amount,current-revision-number,buyer,seller,time-cards,invoice-status-changes,external-identifications,invoice-organization-batch-statuses")

	body, _, err := client.Get(cmd.Context(), "/v1/invoices/"+id, query)
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

	details := buildInvoiceDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderInvoiceDetails(cmd, details)
}

func parseInvoicesShowOptions(cmd *cobra.Command) (invoicesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoicesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildInvoiceDetails(resp jsonAPISingleResponse) invoiceDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := invoiceDetails{
		ID:                                resource.ID,
		Status:                            stringAttr(attrs, "status"),
		InvoiceDate:                       formatDate(stringAttr(attrs, "invoice-date")),
		DueOn:                             formatDate(stringAttr(attrs, "due-on")),
		AdjustmentAmount:                  stringAttr(attrs, "adjustment-amount"),
		CurrencyCode:                      stringAttr(attrs, "currency-code"),
		TotalAmount:                       stringAttr(attrs, "total-amount"),
		Notes:                             stringAttr(attrs, "notes"),
		ExplicitBuyerName:                 stringAttr(attrs, "explicit-buyer-name"),
		ExplicitBuyerAddress:              stringAttr(attrs, "explicit-buyer-address"),
		ShiftDateMin:                      formatDate(stringAttr(attrs, "shift-date-min")),
		ShiftDateMax:                      formatDate(stringAttr(attrs, "shift-date-max")),
		BusinessUnitIDs:                   stringSliceAttr(attrs, "business-unit-ids"),
		CustomerIDs:                       stringSliceAttr(attrs, "customer-ids"),
		IsManagementServiceType:           boolAttr(attrs, "is-management-service-type"),
		OrganizationInvoicesBatchStatuses: stringMapAttr(attrs, "organization-invoices-batch-statuses"),
		CustomerTimeCardAmount:            stringAttr(attrs, "customer-time-card-amount"),
		CustomerTotalAmount:               stringAttr(attrs, "customer-total-amount"),
		BrokerTimeCardAmount:              stringAttr(attrs, "broker-time-card-amount"),
		BrokerTotalAmount:                 stringAttr(attrs, "broker-total-amount"),
		CurrentRevisionNumber:             stringAttr(attrs, "current-revision-number"),
		TimeCardIDs:                       relationshipIDsFromMap(resource.Relationships, "time-cards"),
		InvoiceStatusChangeIDs:            relationshipIDsFromMap(resource.Relationships, "invoice-status-changes"),
		ExternalIdentificationIDs:         relationshipIDsFromMap(resource.Relationships, "external-identifications"),
		InvoiceOrganizationBatchStatusIDs: relationshipIDsFromMap(resource.Relationships, "invoice-organization-batch-statuses"),
	}

	details.BuyerID, details.BuyerType = relationshipRefFromMap(resource.Relationships, "buyer")
	details.SellerID, details.SellerType = relationshipRefFromMap(resource.Relationships, "seller")

	return details
}

func renderInvoiceDetails(cmd *cobra.Command, details invoiceDetails) error {
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
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	if details.ExplicitBuyerName != "" {
		fmt.Fprintf(out, "Explicit Buyer Name: %s\n", details.ExplicitBuyerName)
	}
	if details.ExplicitBuyerAddress != "" {
		fmt.Fprintf(out, "Explicit Buyer Address: %s\n", details.ExplicitBuyerAddress)
	}
	if details.ShiftDateMin != "" {
		fmt.Fprintf(out, "Shift Date Min: %s\n", details.ShiftDateMin)
	}
	if details.ShiftDateMax != "" {
		fmt.Fprintf(out, "Shift Date Max: %s\n", details.ShiftDateMax)
	}
	if len(details.BusinessUnitIDs) > 0 {
		fmt.Fprintf(out, "Business Unit IDs: %s\n", strings.Join(details.BusinessUnitIDs, ", "))
	}
	if len(details.CustomerIDs) > 0 {
		fmt.Fprintf(out, "Customer IDs: %s\n", strings.Join(details.CustomerIDs, ", "))
	}
	fmt.Fprintf(out, "Is Management Service Type: %t\n", details.IsManagementServiceType)
	if batchStatuses := formatStringMap(details.OrganizationInvoicesBatchStatuses); batchStatuses != "" {
		fmt.Fprintf(out, "Organization Invoice Batch Statuses: %s\n", batchStatuses)
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
		fmt.Fprintf(out, "Current Revision Number: %s\n", details.CurrentRevisionNumber)
	}
	if details.BuyerID != "" || details.BuyerType != "" {
		fmt.Fprintf(out, "Buyer: %s\n", formatRelationshipLabel(details.BuyerType, details.BuyerID))
	}
	if details.SellerID != "" || details.SellerType != "" {
		fmt.Fprintf(out, "Seller: %s\n", formatRelationshipLabel(details.SellerType, details.SellerID))
	}
	if len(details.TimeCardIDs) > 0 {
		fmt.Fprintf(out, "Time Card IDs: %s\n", strings.Join(details.TimeCardIDs, ", "))
	}
	if len(details.InvoiceStatusChangeIDs) > 0 {
		fmt.Fprintf(out, "Invoice Status Change IDs: %s\n", strings.Join(details.InvoiceStatusChangeIDs, ", "))
	}
	if len(details.ExternalIdentificationIDs) > 0 {
		fmt.Fprintf(out, "External Identification IDs: %s\n", strings.Join(details.ExternalIdentificationIDs, ", "))
	}
	if len(details.InvoiceOrganizationBatchStatusIDs) > 0 {
		fmt.Fprintf(out, "Invoice Organization Batch Status IDs: %s\n", strings.Join(details.InvoiceOrganizationBatchStatusIDs, ", "))
	}

	return nil
}
