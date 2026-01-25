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

type rateAgreementsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Name    string
	Status  string
	Buyer   string
	Seller  string
	Search  string
}

type rateAgreementRow struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	Status     string `json:"status,omitempty"`
	CanDelete  bool   `json:"can_delete"`
	SellerType string `json:"seller_type,omitempty"`
	SellerID   string `json:"seller_id,omitempty"`
	BuyerType  string `json:"buyer_type,omitempty"`
	BuyerID    string `json:"buyer_id,omitempty"`
}

func newRateAgreementsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rate agreements",
		Long: `List rate agreements.

Output Columns:
  ID         Rate agreement identifier
  NAME       Rate agreement name
  STATUS     Status (active/inactive)
  CAN DELETE Whether the rate agreement can be deleted
  SELLER     Seller type and ID
  BUYER      Buyer type and ID

Filters:
  --name    Filter by name
  --status  Filter by status (active/inactive)
  --buyer   Filter by buyer (format: Type|ID, e.g. Customer|123)
  --seller  Filter by seller (format: Type|ID, e.g. Broker|456)
  --search  Search rate agreements by name

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List rate agreements
  xbe view rate-agreements list

  # Filter by status
  xbe view rate-agreements list --status active

  # Filter by seller and buyer
  xbe view rate-agreements list --seller "Broker|123" --buyer "Customer|456"

  # Search by name
  xbe view rate-agreements list --search "Standard"

  # Output as JSON
  xbe view rate-agreements list --json`,
		Args: cobra.NoArgs,
		RunE: runRateAgreementsList,
	}
	initRateAgreementsListFlags(cmd)
	return cmd
}

func init() {
	rateAgreementsCmd.AddCommand(newRateAgreementsListCmd())
}

func initRateAgreementsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("name", "", "Filter by name")
	cmd.Flags().String("status", "", "Filter by status (active/inactive)")
	cmd.Flags().String("buyer", "", "Filter by buyer (Type|ID, e.g. Customer|123)")
	cmd.Flags().String("seller", "", "Filter by seller (Type|ID, e.g. Broker|456)")
	cmd.Flags().String("search", "", "Search rate agreements by name")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRateAgreementsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRateAgreementsListOptions(cmd)
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
	query.Set("fields[rate-agreements]", "name,status,can-delete,seller,buyer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[buyer]", opts.Buyer)
	setFilterIfPresent(query, "filter[seller]", opts.Seller)
	setFilterIfPresent(query, "filter[q]", opts.Search)

	body, _, err := client.Get(cmd.Context(), "/v1/rate-agreements", query)
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

	rows := buildRateAgreementRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRateAgreementsTable(cmd, rows)
}

func parseRateAgreementsListOptions(cmd *cobra.Command) (rateAgreementsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	name, _ := cmd.Flags().GetString("name")
	status, _ := cmd.Flags().GetString("status")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	search, _ := cmd.Flags().GetString("search")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return rateAgreementsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Name:    name,
		Status:  status,
		Buyer:   buyer,
		Seller:  seller,
		Search:  search,
	}, nil
}

func buildRateAgreementRows(resp jsonAPIResponse) []rateAgreementRow {
	rows := make([]rateAgreementRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRateAgreementRow(resource))
	}
	return rows
}

func buildRateAgreementRow(resource jsonAPIResource) rateAgreementRow {
	row := rateAgreementRow{
		ID:        resource.ID,
		Name:      strings.TrimSpace(stringAttr(resource.Attributes, "name")),
		Status:    stringAttr(resource.Attributes, "status"),
		CanDelete: boolAttr(resource.Attributes, "can-delete"),
	}

	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		row.SellerType = rel.Data.Type
		row.SellerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		row.BuyerType = rel.Data.Type
		row.BuyerID = rel.Data.ID
	}

	return row
}

func renderRateAgreementsTable(cmd *cobra.Command, rows []rateAgreementRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No rate agreements found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSTATUS\tCAN DELETE\tSELLER\tBUYER")
	for _, row := range rows {
		seller := formatRateAgreementParty(row.SellerType, row.SellerID)
		buyer := formatRateAgreementParty(row.BuyerType, row.BuyerID)

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 32),
			row.Status,
			formatBool(row.CanDelete),
			truncateString(seller, 30),
			truncateString(buyer, 30),
		)
	}
	return writer.Flush()
}

func formatRateAgreementParty(resourceType, resourceID string) string {
	if resourceType == "" || resourceID == "" {
		return ""
	}
	return resourceType + "/" + resourceID
}
