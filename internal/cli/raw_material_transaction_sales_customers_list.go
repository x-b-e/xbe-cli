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

type rawMaterialTransactionSalesCustomersListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Sort     string
	Broker   string
	Customer string
}

type rawMaterialTransactionSalesCustomerRow struct {
	ID                 string `json:"id"`
	RawSalesCustomerID string `json:"raw_sales_customer_id,omitempty"`
	CustomerID         string `json:"customer_id,omitempty"`
	CustomerName       string `json:"customer_name,omitempty"`
	BrokerID           string `json:"broker_id,omitempty"`
	BrokerName         string `json:"broker_name,omitempty"`
}

func newRawMaterialTransactionSalesCustomersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List raw material transaction sales customers",
		Long: `List raw material transaction sales customers with filtering and pagination.

Output Columns:
  ID             Raw material transaction sales customer identifier
  RAW SALES ID   Raw sales customer identifier
  CUSTOMER       Customer name or ID
  BROKER         Broker name or ID

Filters:
  --customer  Filter by customer ID
  --broker    Filter by broker ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List raw material transaction sales customers
  xbe view raw-material-transaction-sales-customers list

  # Filter by customer
  xbe view raw-material-transaction-sales-customers list --customer 123

  # Filter by broker
  xbe view raw-material-transaction-sales-customers list --broker 456

  # Output as JSON
  xbe view raw-material-transaction-sales-customers list --json`,
		Args: cobra.NoArgs,
		RunE: runRawMaterialTransactionSalesCustomersList,
	}
	initRawMaterialTransactionSalesCustomersListFlags(cmd)
	return cmd
}

func init() {
	rawMaterialTransactionSalesCustomersCmd.AddCommand(newRawMaterialTransactionSalesCustomersListCmd())
}

func initRawMaterialTransactionSalesCustomersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRawMaterialTransactionSalesCustomersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRawMaterialTransactionSalesCustomersListOptions(cmd)
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
	query.Set("fields[raw-material-transaction-sales-customers]", "raw-sales-customer-id,customer,broker")
	query.Set("include", "customer,broker")
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

	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/raw-material-transaction-sales-customers", query)
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

	rows := buildRawMaterialTransactionSalesCustomerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRawMaterialTransactionSalesCustomersTable(cmd, rows)
}

func parseRawMaterialTransactionSalesCustomersListOptions(cmd *cobra.Command) (rawMaterialTransactionSalesCustomersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rawMaterialTransactionSalesCustomersListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Sort:     sort,
		Customer: customer,
		Broker:   broker,
	}, nil
}

func buildRawMaterialTransactionSalesCustomerRows(resp jsonAPIResponse) []rawMaterialTransactionSalesCustomerRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]rawMaterialTransactionSalesCustomerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRawMaterialTransactionSalesCustomerRow(resource, included))
	}
	return rows
}

func rawMaterialTransactionSalesCustomerRowFromSingle(resp jsonAPISingleResponse) rawMaterialTransactionSalesCustomerRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildRawMaterialTransactionSalesCustomerRow(resp.Data, included)
}

func buildRawMaterialTransactionSalesCustomerRow(resource jsonAPIResource, included map[string]jsonAPIResource) rawMaterialTransactionSalesCustomerRow {
	row := rawMaterialTransactionSalesCustomerRow{
		ID:                 resource.ID,
		RawSalesCustomerID: stringAttr(resource.Attributes, "raw-sales-customer-id"),
	}

	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.CustomerName = stringAttr(customer.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.BrokerName = stringAttr(broker.Attributes, "company-name")
		}
	}

	return row
}

func renderRawMaterialTransactionSalesCustomersTable(cmd *cobra.Command, rows []rawMaterialTransactionSalesCustomerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No raw material transaction sales customers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRAW SALES ID\tCUSTOMER\tBROKER")
	for _, row := range rows {
		customer := firstNonEmpty(row.CustomerName, row.CustomerID)
		broker := firstNonEmpty(row.BrokerName, row.BrokerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.RawSalesCustomerID, 24),
			truncateString(customer, 28),
			truncateString(broker, 28),
		)
	}
	return writer.Flush()
}
