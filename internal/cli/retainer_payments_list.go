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

type retainerPaymentsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	RetainerPeriod string
	Status         string
	RetainerType   string
	Buyer          string
	Seller         string
	PayOnMin       string
}

type retainerPaymentRow struct {
	ID               string `json:"id"`
	RetainerID       string `json:"retainer_id,omitempty"`
	RetainerPeriodID string `json:"retainer_period_id,omitempty"`
	Status           string `json:"status,omitempty"`
	Amount           string `json:"amount,omitempty"`
	CreatedOn        string `json:"created_on,omitempty"`
	PayOn            string `json:"pay_on,omitempty"`
	Kind             string `json:"kind,omitempty"`
}

func newRetainerPaymentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List retainer payments",
		Long: `List retainer payments with filtering and pagination.

Retainer payments represent scheduled or recorded payments tied to retainer periods.

Output Columns:
  ID        Retainer payment ID
  RETAINER  Retainer ID
  PERIOD    Retainer period ID
  STATUS    Payment status
  AMOUNT    Payment amount
  CREATED   Payment creation date
  PAY_ON    Scheduled pay-on date
  KIND      Payment kind (pre or closing)

Filters:
  --retainer-period  Filter by retainer period ID
  --status           Filter by status (editing, approved, batched, exported)
  --retainer-type    Filter by retainer type (BrokerRetainer, CustomerRetainer)
  --buyer            Filter by buyer organization ID
  --seller           Filter by seller organization ID
  --pay-on-min       Filter by minimum pay-on date (YYYY-MM-DD)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List retainer payments
  xbe view retainer-payments list

  # Filter by retainer period
  xbe view retainer-payments list --retainer-period 123

  # Filter by status
  xbe view retainer-payments list --status approved

  # Filter by retainer type
  xbe view retainer-payments list --retainer-type CustomerRetainer

  # Filter by buyer
  xbe view retainer-payments list --buyer 456

  # Filter by pay-on date
  xbe view retainer-payments list --pay-on-min 2025-01-15

  # Output as JSON
  xbe view retainer-payments list --json`,
		Args: cobra.NoArgs,
		RunE: runRetainerPaymentsList,
	}
	initRetainerPaymentsListFlags(cmd)
	return cmd
}

func init() {
	retainerPaymentsCmd.AddCommand(newRetainerPaymentsListCmd())
}

func initRetainerPaymentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("retainer-period", "", "Filter by retainer period ID")
	cmd.Flags().String("status", "", "Filter by status (editing, approved, batched, exported)")
	cmd.Flags().String("retainer-type", "", "Filter by retainer type (BrokerRetainer, CustomerRetainer)")
	cmd.Flags().String("buyer", "", "Filter by buyer organization ID")
	cmd.Flags().String("seller", "", "Filter by seller organization ID")
	cmd.Flags().String("pay-on-min", "", "Filter by minimum pay-on date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runRetainerPaymentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseRetainerPaymentsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[retainer-payments]", "status,amount,created-on,pay-on,kind,retainer-period,retainer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[retainer-period]", opts.RetainerPeriod)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[retainer-type]", opts.RetainerType)
	setFilterIfPresent(query, "filter[buyer]", opts.Buyer)
	setFilterIfPresent(query, "filter[seller]", opts.Seller)
	setFilterIfPresent(query, "filter[pay-on-min]", opts.PayOnMin)

	body, _, err := client.Get(cmd.Context(), "/v1/retainer-payments", query)
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

	rows := buildRetainerPaymentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderRetainerPaymentsTable(cmd, rows)
}

func parseRetainerPaymentsListOptions(cmd *cobra.Command) (retainerPaymentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	retainerPeriod, _ := cmd.Flags().GetString("retainer-period")
	status, _ := cmd.Flags().GetString("status")
	retainerType, _ := cmd.Flags().GetString("retainer-type")
	buyer, _ := cmd.Flags().GetString("buyer")
	seller, _ := cmd.Flags().GetString("seller")
	payOnMin, _ := cmd.Flags().GetString("pay-on-min")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return retainerPaymentsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		RetainerPeriod: retainerPeriod,
		Status:         status,
		RetainerType:   retainerType,
		Buyer:          buyer,
		Seller:         seller,
		PayOnMin:       payOnMin,
	}, nil
}

func buildRetainerPaymentRows(resp jsonAPIResponse) []retainerPaymentRow {
	rows := make([]retainerPaymentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := retainerPaymentRow{
			ID:        resource.ID,
			Status:    stringAttr(attrs, "status"),
			Amount:    stringAttr(attrs, "amount"),
			CreatedOn: formatDate(stringAttr(attrs, "created-on")),
			PayOn:     formatDate(stringAttr(attrs, "pay-on")),
			Kind:      stringAttr(attrs, "kind"),
		}

		if rel, ok := resource.Relationships["retainer"]; ok && rel.Data != nil {
			row.RetainerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["retainer-period"]; ok && rel.Data != nil {
			row.RetainerPeriodID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderRetainerPaymentsTable(cmd *cobra.Command, rows []retainerPaymentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No retainer payments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tRETAINER\tPERIOD\tSTATUS\tAMOUNT\tCREATED\tPAY_ON\tKIND")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.RetainerID,
			row.RetainerPeriodID,
			row.Status,
			row.Amount,
			row.CreatedOn,
			row.PayOn,
			row.Kind,
		)
	}
	return writer.Flush()
}
