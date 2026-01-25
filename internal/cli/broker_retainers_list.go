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

type brokerRetainersListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Status  string
	Buyer   string
	Seller  string
}

type brokerRetainerRow struct {
	ID               string `json:"id"`
	Status           string `json:"status,omitempty"`
	TermStartOn      string `json:"term_start_on,omitempty"`
	TermEndOn        string `json:"term_end_on,omitempty"`
	ExpectedEarnings any    `json:"expected_earnings,omitempty"`
	ActualEarnings   any    `json:"actual_earnings,omitempty"`
	BuyerType        string `json:"buyer_type,omitempty"`
	BuyerID          string `json:"buyer_id,omitempty"`
	BuyerName        string `json:"buyer,omitempty"`
	SellerType       string `json:"seller_type,omitempty"`
	SellerID         string `json:"seller_id,omitempty"`
	SellerName       string `json:"seller,omitempty"`
}

func newBrokerRetainersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List broker retainers",
		Long: `List broker retainers with filtering and pagination.

Output Columns:
  ID          Broker retainer identifier
  STATUS      Retainer status
  BUYER       Broker (buyer)
  SELLER      Trucker (seller)
  TERM START  Term start date
  TERM END    Term end date
  EXPECTED    Expected earnings
  ACTUAL      Actual earnings

Filters:
  --status  Filter by status
  --buyer   Filter by buyer (Type|ID, e.g., Broker|123)
  --seller  Filter by seller (Type|ID, e.g., Trucker|456)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.
  Example: --sort seller.company_name

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List broker retainers
  xbe view broker-retainers list

  # Filter by status
  xbe view broker-retainers list --status active

  # Filter by buyer
  xbe view broker-retainers list --buyer Broker|123

  # Sort by seller company name
  xbe view broker-retainers list --sort seller.company_name

  # Output as JSON
  xbe view broker-retainers list --json`,
		RunE: runBrokerRetainersList,
	}
	initBrokerRetainersListFlags(cmd)
	return cmd
}

func init() {
	brokerRetainersCmd.AddCommand(newBrokerRetainersListCmd())
}

func initBrokerRetainersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("buyer", "", "Filter by buyer (Type|ID, e.g., Broker|123)")
	cmd.Flags().String("seller", "", "Filter by seller (Type|ID, e.g., Trucker|456)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runBrokerRetainersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseBrokerRetainersListOptions(cmd)
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
	query.Set("fields[broker-retainers]", strings.Join([]string{
		"status",
		"term-start-on",
		"term-end-on",
		"expected-earnings",
		"actual-earnings",
		"buyer",
		"seller",
	}, ","))
	query.Set("include", "buyer,seller")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "id")
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[buyer]", opts.Buyer)
	setFilterIfPresent(query, "filter[seller]", opts.Seller)

	body, _, err := client.Get(cmd.Context(), "/v1/broker-retainers", query)
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

	rows := buildBrokerRetainerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderBrokerRetainersTable(cmd, rows)
}

func parseBrokerRetainersListOptions(cmd *cobra.Command) (brokerRetainersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return brokerRetainersListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Status:  status,
		Buyer:   buyer,
		Seller:  seller,
	}, nil
}

func buildBrokerRetainerRows(resp jsonAPIResponse) []brokerRetainerRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]brokerRetainerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildBrokerRetainerRow(resource, included))
	}
	return rows
}

func brokerRetainerRowFromSingle(resp jsonAPISingleResponse) brokerRetainerRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildBrokerRetainerRow(resp.Data, included)
}

func buildBrokerRetainerRow(resource jsonAPIResource, included map[string]jsonAPIResource) brokerRetainerRow {
	attrs := resource.Attributes
	row := brokerRetainerRow{
		ID:               resource.ID,
		Status:           stringAttr(attrs, "status"),
		TermStartOn:      formatDate(stringAttr(attrs, "term-start-on")),
		TermEndOn:        formatDate(stringAttr(attrs, "term-end-on")),
		ExpectedEarnings: anyAttr(attrs, "expected-earnings"),
		ActualEarnings:   anyAttr(attrs, "actual-earnings"),
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

func renderBrokerRetainersTable(cmd *cobra.Command, rows []brokerRetainerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No broker retainers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tBUYER\tSELLER\tTERM START\tTERM END\tEXPECTED\tACTUAL")
	for _, row := range rows {
		buyerLabel := formatRelated(row.BuyerName, row.BuyerID)
		sellerLabel := formatRelated(row.SellerName, row.SellerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(buyerLabel, 28),
			truncateString(sellerLabel, 28),
			row.TermStartOn,
			row.TermEndOn,
			formatAnyValue(row.ExpectedEarnings),
			formatAnyValue(row.ActualEarnings),
		)
	}
	return writer.Flush()
}
