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

type brokerInvoicesListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	Buyer                         string
	Seller                        string
	Status                        string
	InvoiceDate                   string
	InvoiceDateMin                string
	InvoiceDateMax                string
	HasInvoiceDate                string
	DueOn                         string
	DueOnMin                      string
	DueOnMax                      string
	HasDueOn                      string
	TicketNumber                  string
	MaterialTransactionTicketNums string
	Tender                        string
	IsManagementServiceType       string
	BusinessUnit                  string
	NotBusinessUnit               string
	Customer                      string
	MaterialTransactionCostCodes  string
	AllocatedCostCodes            string
	Broker                        string
	HasTicketReport               string
	BatchStatus                   string
	HavingPlansWithJobNumberLike  string
}

type brokerInvoiceRow struct {
	ID             string `json:"id"`
	Status         string `json:"status,omitempty"`
	InvoiceDate    string `json:"invoice_date,omitempty"`
	DueOn          string `json:"due_on,omitempty"`
	TotalAmount    string `json:"total_amount,omitempty"`
	TimeCardAmount string `json:"time_card_amount,omitempty"`
	CurrencyCode   string `json:"currency_code,omitempty"`
	BuyerType      string `json:"buyer_type,omitempty"`
	BuyerID        string `json:"buyer_id,omitempty"`
	BuyerName      string `json:"buyer,omitempty"`
	SellerType     string `json:"seller_type,omitempty"`
	SellerID       string `json:"seller_id,omitempty"`
	SellerName     string `json:"seller,omitempty"`
}

func newBrokerInvoicesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List broker invoices",
		Long: `List broker invoices with filtering and pagination.

Output Columns:
  ID        Invoice identifier
  STATUS    Invoice status
  CUSTOMER  Buyer (customer)
  BROKER    Seller (broker)
  INVOICE   Invoice date
  DUE       Due date
  TOTAL     Total amount
  CARD AMT  Time card amount

Filters:
  --buyer                                Filter by buyer (Type|ID, e.g., Customer|123)
  --seller                               Filter by seller (Type|ID, e.g., Broker|456)
  --status                               Filter by status
  --invoice-date                         Filter by invoice date (YYYY-MM-DD)
  --invoice-date-min                     Filter by minimum invoice date (YYYY-MM-DD)
  --invoice-date-max                     Filter by maximum invoice date (YYYY-MM-DD)
  --has-invoice-date                     Filter by presence of invoice date (true/false)
  --due-on                               Filter by due date (YYYY-MM-DD)
  --due-on-min                           Filter by minimum due date (YYYY-MM-DD)
  --due-on-max                           Filter by maximum due date (YYYY-MM-DD)
  --has-due-on                           Filter by presence of due date (true/false)
  --ticket-number                        Filter by time card ticket number
  --material-transaction-ticket-numbers  Filter by material transaction ticket numbers (comma-separated)
  --tender                               Filter by tender ID(s)
  --is-management-service-type           Filter by management service type (true/false)
  --business-unit                        Filter by business unit ID(s)
  --not-business-unit                    Exclude business unit ID(s)
  --customer                             Filter by customer ID(s)
  --material-transaction-cost-codes      Filter by material transaction cost codes
  --allocated-cost-codes                 Filter by allocated cost codes
  --broker                               Filter by broker ID(s)
  --has-ticket-report                    Filter by has ticket report (true/false)
  --batch-status                         Filter by batch status (format: Type|ID|status, e.g., customers|89|never_processed)
  --having-plans-with-job-number-like    Filter by job number match

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List broker invoices
  xbe view broker-invoices list

  # Filter by customer and status
  xbe view broker-invoices list --customer 123 --status approved

  # Filter by invoice date range
  xbe view broker-invoices list --invoice-date-min 2025-01-01 --invoice-date-max 2025-01-31

  # Output as JSON
  xbe view broker-invoices list --json`,
		Args: cobra.NoArgs,
		RunE: runBrokerInvoicesList,
	}
	initBrokerInvoicesListFlags(cmd)
	return cmd
}

func init() {
	brokerInvoicesCmd.AddCommand(newBrokerInvoicesListCmd())
}

func initBrokerInvoicesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("buyer", "", "Filter by buyer (Type|ID, e.g., Customer|123)")
	cmd.Flags().String("seller", "", "Filter by seller (Type|ID, e.g., Broker|456)")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("invoice-date", "", "Filter by invoice date (YYYY-MM-DD)")
	cmd.Flags().String("invoice-date-min", "", "Filter by minimum invoice date (YYYY-MM-DD)")
	cmd.Flags().String("invoice-date-max", "", "Filter by maximum invoice date (YYYY-MM-DD)")
	cmd.Flags().String("has-invoice-date", "", "Filter by presence of invoice date (true/false)")
	cmd.Flags().String("due-on", "", "Filter by due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-min", "", "Filter by minimum due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-max", "", "Filter by maximum due date (YYYY-MM-DD)")
	cmd.Flags().String("has-due-on", "", "Filter by presence of due date (true/false)")
	cmd.Flags().String("ticket-number", "", "Filter by time card ticket number")
	cmd.Flags().String("material-transaction-ticket-numbers", "", "Filter by material transaction ticket numbers (comma-separated)")
	cmd.Flags().String("tender", "", "Filter by tender ID(s)")
	cmd.Flags().String("is-management-service-type", "", "Filter by management service type (true/false)")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID(s)")
	cmd.Flags().String("not-business-unit", "", "Exclude business unit ID(s)")
	cmd.Flags().String("customer", "", "Filter by customer ID(s)")
	cmd.Flags().String("material-transaction-cost-codes", "", "Filter by material transaction cost codes")
	cmd.Flags().String("allocated-cost-codes", "", "Filter by allocated cost codes")
	cmd.Flags().String("broker", "", "Filter by broker ID(s)")
	cmd.Flags().String("has-ticket-report", "", "Filter by has ticket report (true/false)")
	cmd.Flags().String("batch-status", "", "Filter by batch status (format: Type|ID|status, e.g., customers|89|never_processed)")
	cmd.Flags().String("having-plans-with-job-number-like", "", "Filter by job number match")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerInvoicesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokerInvoicesListOptions(cmd)
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
	query.Set("fields[broker-invoices]", strings.Join([]string{
		"status",
		"invoice-date",
		"due-on",
		"total-amount",
		"time-card-amount",
		"currency-code",
		"buyer",
		"seller",
	}, ","))
	query.Set("include", "buyer,seller")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[buyer]", opts.Buyer)
	setFilterIfPresent(query, "filter[seller]", opts.Seller)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[invoice-date]", opts.InvoiceDate)
	setFilterIfPresent(query, "filter[invoice-date-min]", opts.InvoiceDateMin)
	setFilterIfPresent(query, "filter[invoice-date-max]", opts.InvoiceDateMax)
	setFilterIfPresent(query, "filter[has-invoice-date]", opts.HasInvoiceDate)
	setFilterIfPresent(query, "filter[due-on]", opts.DueOn)
	setFilterIfPresent(query, "filter[due-on-min]", opts.DueOnMin)
	setFilterIfPresent(query, "filter[due-on-max]", opts.DueOnMax)
	setFilterIfPresent(query, "filter[has-due-on]", opts.HasDueOn)
	setFilterIfPresent(query, "filter[ticket-number]", opts.TicketNumber)
	setFilterIfPresent(query, "filter[material-transaction-ticket-numbers]", opts.MaterialTransactionTicketNums)
	setFilterIfPresent(query, "filter[tender]", opts.Tender)
	setFilterIfPresent(query, "filter[is-management-service-type]", opts.IsManagementServiceType)
	setFilterIfPresent(query, "filter[business-unit]", opts.BusinessUnit)
	setFilterIfPresent(query, "filter[not-business-unit]", opts.NotBusinessUnit)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[material-transaction-cost-codes]", opts.MaterialTransactionCostCodes)
	setFilterIfPresent(query, "filter[allocated-cost-codes]", opts.AllocatedCostCodes)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[has-ticket-report]", opts.HasTicketReport)
	setFilterIfPresent(query, "filter[batch-status]", opts.BatchStatus)
	setFilterIfPresent(query, "filter[having-plans-with-job-number-like]", opts.HavingPlansWithJobNumberLike)

	body, _, err := client.Get(cmd.Context(), "/v1/broker-invoices", query)
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

	rows := buildBrokerInvoiceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokerInvoicesTable(cmd, rows)
}

func parseBrokerInvoicesListOptions(cmd *cobra.Command) (brokerInvoicesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	status, _ := cmd.Flags().GetString("status")
	invoiceDate, _ := cmd.Flags().GetString("invoice-date")
	invoiceDateMin, _ := cmd.Flags().GetString("invoice-date-min")
	invoiceDateMax, _ := cmd.Flags().GetString("invoice-date-max")
	hasInvoiceDate, _ := cmd.Flags().GetString("has-invoice-date")
	dueOn, _ := cmd.Flags().GetString("due-on")
	dueOnMin, _ := cmd.Flags().GetString("due-on-min")
	dueOnMax, _ := cmd.Flags().GetString("due-on-max")
	hasDueOn, _ := cmd.Flags().GetString("has-due-on")
	ticketNumber, _ := cmd.Flags().GetString("ticket-number")
	materialTransactionTicketNums, _ := cmd.Flags().GetString("material-transaction-ticket-numbers")
	tender, _ := cmd.Flags().GetString("tender")
	isManagementServiceType, _ := cmd.Flags().GetString("is-management-service-type")
	businessUnit, _ := cmd.Flags().GetString("business-unit")
	notBusinessUnit, _ := cmd.Flags().GetString("not-business-unit")
	customer, _ := cmd.Flags().GetString("customer")
	materialTransactionCostCodes, _ := cmd.Flags().GetString("material-transaction-cost-codes")
	allocatedCostCodes, _ := cmd.Flags().GetString("allocated-cost-codes")
	broker, _ := cmd.Flags().GetString("broker")
	hasTicketReport, _ := cmd.Flags().GetString("has-ticket-report")
	batchStatus, _ := cmd.Flags().GetString("batch-status")
	havingPlansWithJobNumberLike, _ := cmd.Flags().GetString("having-plans-with-job-number-like")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerInvoicesListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		Buyer:                         buyer,
		Seller:                        seller,
		Status:                        status,
		InvoiceDate:                   invoiceDate,
		InvoiceDateMin:                invoiceDateMin,
		InvoiceDateMax:                invoiceDateMax,
		HasInvoiceDate:                hasInvoiceDate,
		DueOn:                         dueOn,
		DueOnMin:                      dueOnMin,
		DueOnMax:                      dueOnMax,
		HasDueOn:                      hasDueOn,
		TicketNumber:                  ticketNumber,
		MaterialTransactionTicketNums: materialTransactionTicketNums,
		Tender:                        tender,
		IsManagementServiceType:       isManagementServiceType,
		BusinessUnit:                  businessUnit,
		NotBusinessUnit:               notBusinessUnit,
		Customer:                      customer,
		MaterialTransactionCostCodes:  materialTransactionCostCodes,
		AllocatedCostCodes:            allocatedCostCodes,
		Broker:                        broker,
		HasTicketReport:               hasTicketReport,
		BatchStatus:                   batchStatus,
		HavingPlansWithJobNumberLike:  havingPlansWithJobNumberLike,
	}, nil
}

func buildBrokerInvoiceRows(resp jsonAPIResponse) []brokerInvoiceRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]brokerInvoiceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildBrokerInvoiceRow(resource, included))
	}

	return rows
}

func brokerInvoiceRowFromSingle(resp jsonAPISingleResponse) brokerInvoiceRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildBrokerInvoiceRow(resp.Data, included)
}

func buildBrokerInvoiceRow(resource jsonAPIResource, included map[string]jsonAPIResource) brokerInvoiceRow {
	attrs := resource.Attributes

	row := brokerInvoiceRow{
		ID:             resource.ID,
		Status:         stringAttr(attrs, "status"),
		InvoiceDate:    formatDate(stringAttr(attrs, "invoice-date")),
		DueOn:          formatDate(stringAttr(attrs, "due-on")),
		TotalAmount:    stringAttr(attrs, "total-amount"),
		TimeCardAmount: stringAttr(attrs, "time-card-amount"),
		CurrencyCode:   stringAttr(attrs, "currency-code"),
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		row.BuyerType = rel.Data.Type
		row.BuyerID = rel.Data.ID
		if buyer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BuyerName = firstNonEmpty(
				stringAttr(buyer.Attributes, "company-name"),
				stringAttr(buyer.Attributes, "name"),
			)
		}
	}

	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		row.SellerType = rel.Data.Type
		row.SellerID = rel.Data.ID
		if seller, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.SellerName = firstNonEmpty(
				stringAttr(seller.Attributes, "company-name"),
				stringAttr(seller.Attributes, "name"),
			)
		}
	}

	return row
}

func renderBrokerInvoicesTable(cmd *cobra.Command, rows []brokerInvoiceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No broker invoices found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tCUSTOMER\tBROKER\tINVOICE\tDUE\tTOTAL\tCARD AMT")
	for _, row := range rows {
		buyerLabel := formatRelated(row.BuyerName, row.BuyerID)
		sellerLabel := formatRelated(row.SellerName, row.SellerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(buyerLabel, 28),
			truncateString(sellerLabel, 28),
			row.InvoiceDate,
			row.DueOn,
			row.TotalAmount,
			row.TimeCardAmount,
		)
	}
	return writer.Flush()
}
