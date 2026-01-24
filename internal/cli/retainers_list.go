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

type retainersListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Status       string
	Buyer        string
	Seller       string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
}

type retainerRow struct {
	ID                               string `json:"id"`
	PolymorphicType                  string `json:"polymorphic_type,omitempty"`
	Status                           string `json:"status,omitempty"`
	TerminatedOn                     string `json:"terminated_on,omitempty"`
	ExpectedEarnings                 string `json:"expected_earnings,omitempty"`
	ActualEarnings                   string `json:"actual_earnings,omitempty"`
	ConsumptionPct                   string `json:"consumption_pct,omitempty"`
	TermStartOn                      string `json:"term_start_on,omitempty"`
	TermEndOn                        string `json:"term_end_on,omitempty"`
	MaximumExpectedDailyHours        string `json:"maximum_expected_daily_hours,omitempty"`
	MaximumTravelMinutes             string `json:"maximum_travel_minutes,omitempty"`
	BillableTravelMinutesPerTravelMi string `json:"billable_travel_minutes_per_travel_mile,omitempty"`
	BuyerType                        string `json:"buyer_type,omitempty"`
	BuyerID                          string `json:"buyer_id,omitempty"`
	SellerType                       string `json:"seller_type,omitempty"`
	SellerID                         string `json:"seller_id,omitempty"`
}

func newRetainersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List retainers",
		Long: `List retainers with filtering and pagination.

Output Columns:
  ID            Retainer identifier
  TYPE          Retainer type
  STATUS        Retainer status
  BUYER         Buyer (type/id)
  SELLER        Seller (type/id)
  TERM START    Term start date
  TERM END      Term end date
  EXPECTED      Expected earnings
  ACTUAL        Actual earnings
  CONSUMPTION   Consumption percentage

Filters:
  --status         Filter by status (editing, active, terminated, expired, closed)
  --buyer          Filter by buyer (Type|ID, e.g. Broker|123)
  --seller         Filter by seller (Type|ID, e.g. Trucker|456)
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List retainers
  xbe view retainers list

  # Filter by status
  xbe view retainers list --status active

  # Filter by buyer
  xbe view retainers list --buyer Broker|123

  # Filter by seller
  xbe view retainers list --seller Trucker|456

  # Output as JSON
  xbe view retainers list --json`,
		Args: cobra.NoArgs,
		RunE: runRetainersList,
	}
	initRetainersListFlags(cmd)
	return cmd
}

func init() {
	retainersCmd.AddCommand(newRetainersListCmd())
}

func initRetainersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("buyer", "", "Filter by buyer (Type|ID, e.g. Broker|123)")
	cmd.Flags().String("seller", "", "Filter by seller (Type|ID, e.g. Trucker|456)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRetainersListOptions(cmd)
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
	query.Set("fields[retainers]", "polymorphic-type,status,term-start-on,term-end-on,expected-earnings,actual-earnings,consumption-pct,buyer,seller")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[buyer]", opts.Buyer)
	setFilterIfPresent(query, "filter[seller]", opts.Seller)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/retainers", query)
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

	rows := buildRetainerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRetainersTable(cmd, rows)
}

func parseRetainersListOptions(cmd *cobra.Command) (retainersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainersListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Status:       status,
		Buyer:        buyer,
		Seller:       seller,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
	}, nil
}

func buildRetainerRows(resp jsonAPIResponse) []retainerRow {
	rows := make([]retainerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildRetainerRow(resource))
	}
	return rows
}

func buildRetainerRow(resource jsonAPIResource) retainerRow {
	row := retainerRow{
		ID:                               resource.ID,
		PolymorphicType:                  stringAttr(resource.Attributes, "polymorphic-type"),
		Status:                           stringAttr(resource.Attributes, "status"),
		TerminatedOn:                     stringAttr(resource.Attributes, "terminated-on"),
		ExpectedEarnings:                 stringAttr(resource.Attributes, "expected-earnings"),
		ActualEarnings:                   stringAttr(resource.Attributes, "actual-earnings"),
		ConsumptionPct:                   stringAttr(resource.Attributes, "consumption-pct"),
		TermStartOn:                      stringAttr(resource.Attributes, "term-start-on"),
		TermEndOn:                        stringAttr(resource.Attributes, "term-end-on"),
		MaximumExpectedDailyHours:        stringAttr(resource.Attributes, "maximum-expected-daily-hours"),
		MaximumTravelMinutes:             stringAttr(resource.Attributes, "maximum-travel-minutes"),
		BillableTravelMinutesPerTravelMi: stringAttr(resource.Attributes, "billable-travel-minutes-per-travel-mile"),
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		row.BuyerType = rel.Data.Type
		row.BuyerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		row.SellerType = rel.Data.Type
		row.SellerID = rel.Data.ID
	}

	return row
}

func buildRetainerRowFromSingle(resp jsonAPISingleResponse) retainerRow {
	return buildRetainerRow(resp.Data)
}

func renderRetainersTable(cmd *cobra.Command, rows []retainerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No retainers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTYPE\tSTATUS\tBUYER\tSELLER\tTERM START\tTERM END\tEXPECTED\tACTUAL\tCONSUMPTION")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.PolymorphicType, 18),
			row.Status,
			truncateString(formatTypeID(row.BuyerType, row.BuyerID), 32),
			truncateString(formatTypeID(row.SellerType, row.SellerID), 32),
			row.TermStartOn,
			row.TermEndOn,
			row.ExpectedEarnings,
			row.ActualEarnings,
			row.ConsumptionPct,
		)
	}
	return writer.Flush()
}
