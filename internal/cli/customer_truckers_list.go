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

type customerTruckersListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	Sort     string
	Customer string
	Trucker  string
	Broker   string
}

type customerTruckerRow struct {
	ID           string `json:"id"`
	CustomerID   string `json:"customer_id,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
	TruckerID    string `json:"trucker_id,omitempty"`
	TruckerName  string `json:"trucker_name,omitempty"`
}

func newCustomerTruckersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer truckers",
		Long: `List customer truckers with filtering and pagination.

Output Columns:
  ID        Customer trucker link identifier
  CUSTOMER  Customer name or ID
  TRUCKER   Trucker name or ID

Filters:
  --customer  Filter by customer ID
  --trucker   Filter by trucker ID
  --broker    Filter by broker ID (via customer)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List customer truckers
  xbe view customer-truckers list

  # Filter by customer
  xbe view customer-truckers list --customer 123

  # Filter by trucker
  xbe view customer-truckers list --trucker 456

  # Filter by broker
  xbe view customer-truckers list --broker 789

  # Output as JSON
  xbe view customer-truckers list --json`,
		Args: cobra.NoArgs,
		RunE: runCustomerTruckersList,
	}
	initCustomerTruckersListFlags(cmd)
	return cmd
}

func init() {
	customerTruckersCmd.AddCommand(newCustomerTruckersListCmd())
}

func initCustomerTruckersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("broker", "", "Filter by broker ID (via customer)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerTruckersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerTruckersListOptions(cmd)
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
	query.Set("fields[customer-truckers]", "customer,trucker")
	query.Set("include", "customer,trucker")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")

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
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)

	body, _, err := client.Get(cmd.Context(), "/v1/customer-truckers", query)
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

	rows := buildCustomerTruckerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerTruckersTable(cmd, rows)
}

func parseCustomerTruckersListOptions(cmd *cobra.Command) (customerTruckersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	customer, _ := cmd.Flags().GetString("customer")
	trucker, _ := cmd.Flags().GetString("trucker")
	broker, _ := cmd.Flags().GetString("broker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerTruckersListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		Sort:     sort,
		Customer: customer,
		Trucker:  trucker,
		Broker:   broker,
	}, nil
}

func buildCustomerTruckerRows(resp jsonAPIResponse) []customerTruckerRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]customerTruckerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildCustomerTruckerRow(resource, included))
	}
	return rows
}

func customerTruckerRowFromSingle(resp jsonAPISingleResponse) customerTruckerRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildCustomerTruckerRow(resp.Data, included)
}

func buildCustomerTruckerRow(resource jsonAPIResource, included map[string]jsonAPIResource) customerTruckerRow {
	row := customerTruckerRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["customer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
		if customer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.CustomerName = stringAttr(customer.Attributes, "company-name")
		}
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
		if trucker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.TruckerName = stringAttr(trucker.Attributes, "company-name")
		}
	}

	return row
}

func renderCustomerTruckersTable(cmd *cobra.Command, rows []customerTruckerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer truckers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCUSTOMER\tTRUCKER")
	for _, row := range rows {
		customer := firstNonEmpty(row.CustomerName, row.CustomerID)
		trucker := firstNonEmpty(row.TruckerName, row.TruckerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(customer, 28),
			truncateString(trucker, 28),
		)
	}
	return writer.Flush()
}
