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

type truckerInvoicesListOptions struct {
	BaseURL                          string
	Token                            string
	JSON                             bool
	NoAuth                           bool
	Limit                            int
	Offset                           int
	Sort                             string
	Buyer                            string
	Seller                           string
	Status                           string
	InvoiceDate                      string
	InvoiceDateMin                   string
	InvoiceDateMax                   string
	HasInvoiceDate                   string
	DueOn                            string
	DueOnMin                         string
	DueOnMax                         string
	HasDueOn                         string
	TicketNumber                     string
	MaterialTransactionTicketNumbers string
	Tender                           string
	IsManagementServiceType          string
	BusinessUnit                     string
	NotBusinessUnit                  string
	Customer                         string
	MaterialTransactionCostCodes     string
	AllocatedCostCodes               string
	Broker                           string
	HasTicketReport                  string
	BatchStatus                      string
	HavingPlansWithJobNumberLike     string
	QuickbooksID                     string
}

func newTruckerInvoicesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker invoices",
		Long: `List trucker invoices with filtering and pagination.

Trucker invoices represent bills issued to brokers for trucker time cards.

Output Columns:
  ID             Trucker invoice identifier
  STATUS         Invoice status
  INVOICE_DATE   Invoice date
  DUE_ON         Due date
  TOTAL          Total amount
  TIME_CARD_AMT  Time card amount
  CURRENCY       Currency code
  QB_ID          QuickBooks invoice ID
  BUYER          Buyer (type:id)
  SELLER         Seller (type:id)

Filters:
  --buyer                               Filter by buyer ID (comma-separated for multiple)
  --seller                              Filter by seller ID (comma-separated for multiple)
  --status                              Filter by status (comma-separated for multiple)
  --invoice-date                        Filter by invoice date (YYYY-MM-DD)
  --invoice-date-min                    Filter by minimum invoice date (YYYY-MM-DD)
  --invoice-date-max                    Filter by maximum invoice date (YYYY-MM-DD)
  --has-invoice-date                    Filter by invoice date presence (true/false)
  --due-on                              Filter by due date (YYYY-MM-DD)
  --due-on-min                          Filter by minimum due date (YYYY-MM-DD)
  --due-on-max                          Filter by maximum due date (YYYY-MM-DD)
  --has-due-on                          Filter by due date presence (true/false)
  --ticket-number                       Filter by time card ticket number (comma-separated for multiple)
  --material-transaction-ticket-numbers Filter by material transaction ticket numbers (comma-separated)
  --tender                              Filter by tender ID (comma-separated for multiple)
  --is-management-service-type          Filter by management service type (true/false)
  --business-unit                       Filter by business unit ID (comma-separated for multiple)
  --not-business-unit                   Exclude business unit ID (comma-separated for multiple)
  --customer                            Filter by customer ID (comma-separated for multiple)
  --material-transaction-cost-codes     Filter by material transaction cost codes (comma-separated)
  --allocated-cost-codes                Filter by allocated cost codes (comma-separated)
  --broker                              Filter by broker ID (comma-separated for multiple)
  --has-ticket-report                   Filter by ticket report presence (true/false)
  --batch-status                        Filter by batch status (org|id|status[,status])
  --having-plans-with-job-number-like   Filter by job number match (comma-separated)
  --quickbooks-id                       Filter by QuickBooks ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List trucker invoices
  xbe view trucker-invoices list

  # Filter by status
  xbe view trucker-invoices list --status approved

  # Filter by buyer and date range
  xbe view trucker-invoices list --buyer 123 --invoice-date-min 2025-01-01 --invoice-date-max 2025-01-31

  # Filter by QuickBooks ID
  xbe view trucker-invoices list --quickbooks-id 98765

  # Output as JSON
  xbe view trucker-invoices list --json`,
		Args: cobra.NoArgs,
		RunE: runTruckerInvoicesList,
	}
	initTruckerInvoicesListFlags(cmd)
	return cmd
}

func init() {
	truckerInvoicesCmd.AddCommand(newTruckerInvoicesListCmd())
}

func initTruckerInvoicesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("buyer", "", "Filter by buyer ID (comma-separated for multiple)")
	cmd.Flags().String("seller", "", "Filter by seller ID (comma-separated for multiple)")
	cmd.Flags().String("status", "", "Filter by status (comma-separated for multiple)")
	cmd.Flags().String("invoice-date", "", "Filter by invoice date (YYYY-MM-DD)")
	cmd.Flags().String("invoice-date-min", "", "Filter by minimum invoice date (YYYY-MM-DD)")
	cmd.Flags().String("invoice-date-max", "", "Filter by maximum invoice date (YYYY-MM-DD)")
	cmd.Flags().String("has-invoice-date", "", "Filter by invoice date presence (true/false)")
	cmd.Flags().String("due-on", "", "Filter by due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-min", "", "Filter by minimum due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-max", "", "Filter by maximum due date (YYYY-MM-DD)")
	cmd.Flags().String("has-due-on", "", "Filter by due date presence (true/false)")
	cmd.Flags().String("ticket-number", "", "Filter by time card ticket number (comma-separated for multiple)")
	cmd.Flags().String("material-transaction-ticket-numbers", "", "Filter by material transaction ticket numbers (comma-separated)")
	cmd.Flags().String("tender", "", "Filter by tender ID (comma-separated for multiple)")
	cmd.Flags().String("is-management-service-type", "", "Filter by management service type (true/false)")
	cmd.Flags().String("business-unit", "", "Filter by business unit ID (comma-separated for multiple)")
	cmd.Flags().String("not-business-unit", "", "Exclude business unit ID (comma-separated for multiple)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("material-transaction-cost-codes", "", "Filter by material transaction cost codes (comma-separated)")
	cmd.Flags().String("allocated-cost-codes", "", "Filter by allocated cost codes (comma-separated)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("has-ticket-report", "", "Filter by ticket report presence (true/false)")
	cmd.Flags().String("batch-status", "", "Filter by batch status (org|id|status[,status])")
	cmd.Flags().String("having-plans-with-job-number-like", "", "Filter by job number match (comma-separated)")
	cmd.Flags().String("quickbooks-id", "", "Filter by QuickBooks ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerInvoicesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerInvoicesListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[trucker-invoices]", "invoice-date,due-on,status,total-amount,time-card-amount,currency-code,quickbooks-id,buyer,seller")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
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
	setFilterIfPresent(query, "filter[material-transaction-ticket-numbers]", opts.MaterialTransactionTicketNumbers)
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
	setFilterIfPresent(query, "filter[quickbooks-id]", opts.QuickbooksID)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-invoices", query)
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

	rows := buildTruckerInvoiceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckerInvoicesTable(cmd, rows)
}

func parseTruckerInvoicesListOptions(cmd *cobra.Command) (truckerInvoicesListOptions, error) {
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
	materialTransactionTicketNumbers, _ := cmd.Flags().GetString("material-transaction-ticket-numbers")
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
	quickbooksID, _ := cmd.Flags().GetString("quickbooks-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerInvoicesListOptions{
		BaseURL:                          baseURL,
		Token:                            token,
		JSON:                             jsonOut,
		NoAuth:                           noAuth,
		Limit:                            limit,
		Offset:                           offset,
		Sort:                             sort,
		Buyer:                            buyer,
		Seller:                           seller,
		Status:                           status,
		InvoiceDate:                      invoiceDate,
		InvoiceDateMin:                   invoiceDateMin,
		InvoiceDateMax:                   invoiceDateMax,
		HasInvoiceDate:                   hasInvoiceDate,
		DueOn:                            dueOn,
		DueOnMin:                         dueOnMin,
		DueOnMax:                         dueOnMax,
		HasDueOn:                         hasDueOn,
		TicketNumber:                     ticketNumber,
		MaterialTransactionTicketNumbers: materialTransactionTicketNumbers,
		Tender:                           tender,
		IsManagementServiceType:          isManagementServiceType,
		BusinessUnit:                     businessUnit,
		NotBusinessUnit:                  notBusinessUnit,
		Customer:                         customer,
		MaterialTransactionCostCodes:     materialTransactionCostCodes,
		AllocatedCostCodes:               allocatedCostCodes,
		Broker:                           broker,
		HasTicketReport:                  hasTicketReport,
		BatchStatus:                      batchStatus,
		HavingPlansWithJobNumberLike:     havingPlansWithJobNumberLike,
		QuickbooksID:                     quickbooksID,
	}, nil
}

func renderTruckerInvoicesTable(cmd *cobra.Command, rows []truckerInvoiceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trucker invoices found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tINVOICE_DATE\tDUE_ON\tTOTAL\tTIME_CARD_AMT\tCURRENCY\tQB_ID\tBUYER\tSELLER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 18),
			row.InvoiceDate,
			row.DueOn,
			row.TotalAmount,
			row.TimeCardAmount,
			row.CurrencyCode,
			truncateString(row.QuickbooksID, 16),
			truncateString(formatRelationshipLabel(row.BuyerType, row.BuyerID), 24),
			truncateString(formatRelationshipLabel(row.SellerType, row.SellerID), 24),
		)
	}
	return writer.Flush()
}
