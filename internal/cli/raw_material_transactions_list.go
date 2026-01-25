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

type rawMaterialTransactionsListOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	NoAuth               bool
	Limit                int
	Offset               int
	Date                 string
	DateMin              string
	DateMax              string
	MaterialSite         string
	MaterialSiteID       string
	Broker               string
	BrokerID             string
	MaterialSupplierName string
	TicketNumber         string
	JobNumber            string
	HaulerType           string
	SalesCustomerID      string
	TruckName            string
	SiteID               string
	MaterialName         string
}

type rawMaterialTransactionRow struct {
	ID                    string `json:"id"`
	TicketNumber          string `json:"ticket_number,omitempty"`
	TicketJobNumber       string `json:"ticket_job_number,omitempty"`
	TransactionAt         string `json:"transaction_at,omitempty"`
	TransactionDate       string `json:"transaction_date,omitempty"`
	TruckName             string `json:"truck_name,omitempty"`
	MaterialName          string `json:"material_name,omitempty"`
	SiteID                string `json:"site_id,omitempty"`
	HaulerType            string `json:"hauler_type,omitempty"`
	SalesCustomerID       string `json:"sales_customer_id,omitempty"`
	SourceID              string `json:"material_site_id,omitempty"`
	MaterialTransactionID string `json:"material_transaction_id,omitempty"`
}

func newRawMaterialTransactionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw material transactions",
		Long: `List raw material transactions with filtering and pagination.

Raw material transactions represent the raw ticket data ingested from material
sites before normalization. Use filters to narrow by site, broker, ticket number,
or transaction date.

Output Columns:
  ID        Raw material transaction ID
  DATE      Transaction date
  TICKET    Ticket number
  JOB       Ticket job number
  TRUCK     Truck name or identifier
  MATERIAL  Material identifier/name from the raw ticket
  SITE      Raw site ID
  MTXN      Related material transaction ID (if available)

Filters:
  --date                   Filter by transaction date (YYYY-MM-DD)
  --date-min               Filter by minimum transaction timestamp (YYYY-MM-DD)
  --date-max               Filter by maximum transaction timestamp (YYYY-MM-DD)
  --material-site          Filter by material site ID (comma-separated)
  --material-site-id       Filter by material site ID (comma-separated)
  --broker                 Filter by broker ID (comma-separated)
  --broker-id              Filter by broker ID (comma-separated)
  --material-supplier-name Filter by material supplier name (starts with)
  --ticket-number          Filter by ticket number
  --job-number             Filter by ticket job number
  --hauler-type            Filter by hauler type
  --sales-customer-id      Filter by sales customer ID (comma-separated)
  --truck-name             Filter by truck name
  --site-id                Filter by raw site ID
  --material-name          Filter by material ID/name from raw ticket

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List recent raw material transactions
  xbe view raw-material-transactions list

  # Filter by transaction date
  xbe view raw-material-transactions list --date 2025-01-15

  # Filter by ticket number
  xbe view raw-material-transactions list --ticket-number T12345

  # Filter by material site
  xbe view raw-material-transactions list --material-site 456

  # JSON output
  xbe view raw-material-transactions list --json`,
		RunE: runRawMaterialTransactionsList,
	}
	initRawMaterialTransactionsListFlags(cmd)
	return cmd
}

func init() {
	rawMaterialTransactionsCmd.AddCommand(newRawMaterialTransactionsListCmd())
}

func initRawMaterialTransactionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("date", "", "Filter by transaction date (YYYY-MM-DD)")
	cmd.Flags().String("date-min", "", "Filter by minimum transaction date (YYYY-MM-DD)")
	cmd.Flags().String("date-max", "", "Filter by maximum transaction date (YYYY-MM-DD)")
	cmd.Flags().String("material-site", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("material-site-id", "", "Filter by material site ID (comma-separated for multiple)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("broker-id", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("material-supplier-name", "", "Filter by material supplier name (starts with)")
	cmd.Flags().String("ticket-number", "", "Filter by ticket number")
	cmd.Flags().String("job-number", "", "Filter by ticket job number")
	cmd.Flags().String("hauler-type", "", "Filter by hauler type")
	cmd.Flags().String("sales-customer-id", "", "Filter by sales customer ID (comma-separated for multiple)")
	cmd.Flags().String("truck-name", "", "Filter by truck name")
	cmd.Flags().String("site-id", "", "Filter by raw site ID")
	cmd.Flags().String("material-name", "", "Filter by material ID/name from raw ticket")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawMaterialTransactionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawMaterialTransactionsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "-transaction-at")
	query.Set("fields[raw-material-transactions]", "ticket-job-number,uniqueid,ticket-number,transaction-at,truck-name,material-name,site-id,hauler-type,sales-customer-id,material-transaction,source")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	materialSiteFilter := strings.TrimSpace(opts.MaterialSite)
	if strings.TrimSpace(opts.MaterialSiteID) != "" {
		if materialSiteFilter == "" {
			materialSiteFilter = strings.TrimSpace(opts.MaterialSiteID)
		} else {
			materialSiteFilter = materialSiteFilter + "," + strings.TrimSpace(opts.MaterialSiteID)
		}
	}
	setFilterIfPresent(query, "filter[material_site]", materialSiteFilter)

	brokerFilter := strings.TrimSpace(opts.Broker)
	if strings.TrimSpace(opts.BrokerID) != "" {
		if brokerFilter == "" {
			brokerFilter = strings.TrimSpace(opts.BrokerID)
		} else {
			brokerFilter = brokerFilter + "," + strings.TrimSpace(opts.BrokerID)
		}
	}
	setFilterIfPresent(query, "filter[broker]", brokerFilter)
	setFilterIfPresent(query, "filter[material_supplier_name]", opts.MaterialSupplierName)
	setFilterIfPresent(query, "filter[ticket_number]", opts.TicketNumber)
	setFilterIfPresent(query, "filter[job_number]", opts.JobNumber)
	setFilterIfPresent(query, "filter[hauler_type]", opts.HaulerType)
	setFilterIfPresent(query, "filter[sales_customer_id]", opts.SalesCustomerID)
	setFilterIfPresent(query, "filter[truck_name]", opts.TruckName)
	setFilterIfPresent(query, "filter[site_id]", opts.SiteID)
	setFilterIfPresent(query, "filter[material_name]", opts.MaterialName)

	if opts.Date != "" {
		query.Set("filter[transaction_date]", opts.Date)
	} else {
		setFilterIfPresent(query, "filter[transaction_at_min]", opts.DateMin)
		setFilterIfPresent(query, "filter[transaction_at_max]", opts.DateMax)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/raw-material-transactions", query)
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

	rows := buildRawMaterialTransactionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawMaterialTransactionsTable(cmd, rows)
}

func parseRawMaterialTransactionsListOptions(cmd *cobra.Command) (rawMaterialTransactionsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	date, err := cmd.Flags().GetString("date")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	dateMin, err := cmd.Flags().GetString("date-min")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	dateMax, err := cmd.Flags().GetString("date-max")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	materialSite, err := cmd.Flags().GetString("material-site")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	materialSiteID, err := cmd.Flags().GetString("material-site-id")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	brokerID, err := cmd.Flags().GetString("broker-id")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	materialSupplierName, err := cmd.Flags().GetString("material-supplier-name")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	ticketNumber, err := cmd.Flags().GetString("ticket-number")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	jobNumber, err := cmd.Flags().GetString("job-number")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	haulerType, err := cmd.Flags().GetString("hauler-type")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	salesCustomerID, err := cmd.Flags().GetString("sales-customer-id")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	truckName, err := cmd.Flags().GetString("truck-name")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	siteID, err := cmd.Flags().GetString("site-id")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	materialName, err := cmd.Flags().GetString("material-name")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return rawMaterialTransactionsListOptions{}, err
	}

	return rawMaterialTransactionsListOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		NoAuth:               noAuth,
		Limit:                limit,
		Offset:               offset,
		Date:                 date,
		DateMin:              dateMin,
		DateMax:              dateMax,
		MaterialSite:         materialSite,
		MaterialSiteID:       materialSiteID,
		Broker:               broker,
		BrokerID:             brokerID,
		MaterialSupplierName: materialSupplierName,
		TicketNumber:         ticketNumber,
		JobNumber:            jobNumber,
		HaulerType:           haulerType,
		SalesCustomerID:      salesCustomerID,
		TruckName:            truckName,
		SiteID:               siteID,
		MaterialName:         materialName,
	}, nil
}

func buildRawMaterialTransactionRows(resp jsonAPIResponse) []rawMaterialTransactionRow {
	rows := make([]rawMaterialTransactionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := rawMaterialTransactionRow{
			ID:              resource.ID,
			TicketNumber:    stringAttr(attrs, "ticket-number"),
			TicketJobNumber: stringAttr(attrs, "ticket-job-number"),
			TransactionAt:   stringAttr(attrs, "transaction-at"),
			TransactionDate: formatDate(stringAttr(attrs, "transaction-at")),
			TruckName:       stringAttr(attrs, "truck-name"),
			MaterialName:    stringAttr(attrs, "material-name"),
			SiteID:          stringAttr(attrs, "site-id"),
			HaulerType:      stringAttr(attrs, "hauler-type"),
			SalesCustomerID: stringAttr(attrs, "sales-customer-id"),
		}

		if rel, ok := resource.Relationships["source"]; ok && rel.Data != nil {
			row.SourceID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["material-transaction"]; ok && rel.Data != nil {
			row.MaterialTransactionID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderRawMaterialTransactionsTable(cmd *cobra.Command, rows []rawMaterialTransactionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No raw material transactions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDATE\tTICKET\tJOB\tTRUCK\tMATERIAL\tSITE\tMTXN")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TransactionDate,
			truncateString(row.TicketNumber, 16),
			truncateString(row.TicketJobNumber, 12),
			truncateString(row.TruckName, 16),
			truncateString(row.MaterialName, 16),
			truncateString(row.SiteID, 10),
			truncateString(row.MaterialTransactionID, 10),
		)
	}
	return writer.Flush()
}
