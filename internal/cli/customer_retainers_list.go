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

type customerRetainersListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Status       string
	Customer     string
	Broker       string
	CreatedAtMin string
	CreatedAtMax string
	UpdatedAtMin string
	UpdatedAtMax string
	NotID        string
}

type customerRetainerRow struct {
	ID               string `json:"id"`
	Status           string `json:"status,omitempty"`
	CustomerID       string `json:"customer_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	TermStartOn      string `json:"term_start_on,omitempty"`
	TermEndOn        string `json:"term_end_on,omitempty"`
	ExpectedEarnings string `json:"expected_earnings,omitempty"`
	ActualEarnings   string `json:"actual_earnings,omitempty"`
	ConsumptionPct   string `json:"consumption_pct,omitempty"`
}

func newCustomerRetainersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customer retainers",
		Long: `List customer retainers with filtering and pagination.

Output Columns:
  ID            Retainer identifier
  STATUS        Retainer status
  CUSTOMER      Customer ID
  BROKER        Broker ID
  TERM START    Term start date
  TERM END      Term end date
  EXPECTED      Expected earnings
  ACTUAL        Actual earnings
  CONSUMPTION   Consumption percentage

Filters:
  --status         Filter by status (editing, active, terminated, expired, closed)
  --customer       Filter by customer ID
  --broker         Filter by broker ID
  --created-at-min Filter by created-at on/after (ISO 8601)
  --created-at-max Filter by created-at on/before (ISO 8601)
  --updated-at-min Filter by updated-at on/after (ISO 8601)
  --updated-at-max Filter by updated-at on/before (ISO 8601)
  --not-id         Exclude by retainer ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List customer retainers
  xbe view customer-retainers list

  # Filter by customer and status
  xbe view customer-retainers list --customer 123 --status active

  # Filter by broker
  xbe view customer-retainers list --broker 456

  # Output as JSON
  xbe view customer-retainers list --json`,
		Args: cobra.NoArgs,
		RunE: runCustomerRetainersList,
	}
	initCustomerRetainersListFlags(cmd)
	return cmd
}

func init() {
	customerRetainersCmd.AddCommand(newCustomerRetainersListCmd())
}

func initCustomerRetainersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status (editing, active, terminated, expired, closed)")
	cmd.Flags().String("customer", "", "Filter by customer ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("not-id", "", "Exclude by retainer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCustomerRetainersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCustomerRetainersListOptions(cmd)
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
	query.Set("fields[customer-retainers]", "status,expected-earnings,actual-earnings,consumption-pct,term-start-on,term-end-on,buyer,seller")

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
	setFilterIfPresent(query, "filter[buyer]", opts.Customer)
	setFilterIfPresent(query, "filter[seller]", opts.Broker)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[not-id]", opts.NotID)

	body, _, err := client.Get(cmd.Context(), "/v1/customer-retainers", query)
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

	rows := buildCustomerRetainerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCustomerRetainersTable(cmd, rows)
}

func parseCustomerRetainersListOptions(cmd *cobra.Command) (customerRetainersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	customer, _ := cmd.Flags().GetString("customer")
	broker, _ := cmd.Flags().GetString("broker")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	notID, _ := cmd.Flags().GetString("not-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return customerRetainersListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Status:       status,
		Customer:     customer,
		Broker:       broker,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		NotID:        notID,
	}, nil
}

func buildCustomerRetainerRows(resp jsonAPIResponse) []customerRetainerRow {
	rows := make([]customerRetainerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildCustomerRetainerRow(resource))
	}
	return rows
}

func buildCustomerRetainerRow(resource jsonAPIResource) customerRetainerRow {
	attrs := resource.Attributes
	row := customerRetainerRow{
		ID:               resource.ID,
		Status:           stringAttr(attrs, "status"),
		TermStartOn:      formatDate(stringAttr(attrs, "term-start-on")),
		TermEndOn:        formatDate(stringAttr(attrs, "term-end-on")),
		ExpectedEarnings: stringAttr(attrs, "expected-earnings"),
		ActualEarnings:   stringAttr(attrs, "actual-earnings"),
		ConsumptionPct:   stringAttr(attrs, "consumption-pct"),
	}

	if rel, ok := resource.Relationships["buyer"]; ok && rel.Data != nil {
		row.CustomerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["seller"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}

	return row
}

func buildCustomerRetainerRowFromSingle(resp jsonAPISingleResponse) customerRetainerRow {
	return buildCustomerRetainerRow(resp.Data)
}

func renderCustomerRetainersTable(cmd *cobra.Command, rows []customerRetainerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No customer retainers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tCUSTOMER\tBROKER\tTERM START\tTERM END\tEXPECTED\tACTUAL\tCONSUMPTION")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.CustomerID,
			row.BrokerID,
			row.TermStartOn,
			row.TermEndOn,
			row.ExpectedEarnings,
			row.ActualEarnings,
			row.ConsumptionPct,
		)
	}
	return writer.Flush()
}
