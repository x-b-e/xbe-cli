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

type truckerInvoicePaymentsListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	TruckerID string
	Trucker   string
}

type truckerInvoicePaymentRow struct {
	ID            string `json:"id"`
	PaymentAmount string `json:"payment_amount,omitempty"`
	PaymentDate   string `json:"payment_date,omitempty"`
	TruckerID     string `json:"trucker_id,omitempty"`
}

func newTruckerInvoicePaymentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List trucker invoice payments",
		Long: `List trucker invoice payments from QuickBooks.

Output Columns:
  ID        Payment identifier (QuickBooks bill payment ID)
  DATE      Payment date
  AMOUNT    Payment amount
  TRUCKER   Trucker ID

Filters:
  --trucker-id  Filter by trucker ID (required; scopes the QuickBooks company)
  --trucker     Filter by trucker relationship ID

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List payments for a trucker
  xbe view trucker-invoice-payments list --trucker-id 123

  # Filter by trucker relationship
  xbe view trucker-invoice-payments list --trucker 123

  # JSON output
  xbe view trucker-invoice-payments list --trucker-id 123 --json`,
		Args: cobra.NoArgs,
		RunE: runTruckerInvoicePaymentsList,
	}
	initTruckerInvoicePaymentsListFlags(cmd)
	return cmd
}

func init() {
	truckerInvoicePaymentsCmd.AddCommand(newTruckerInvoicePaymentsListCmd())
}

func initTruckerInvoicePaymentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("trucker-id", "", "Filter by trucker ID (required; scopes QuickBooks company)")
	cmd.Flags().String("trucker", "", "Filter by trucker relationship ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTruckerInvoicePaymentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTruckerInvoicePaymentsListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.TruckerID) == "" && strings.TrimSpace(opts.Trucker) == "" {
		err := fmt.Errorf("either --trucker-id or --trucker is required to scope QuickBooks payments")
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
	query.Set("fields[trucker-invoice-payments]", "payment-amount,payment-date,trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[trucker_id]", opts.TruckerID)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)

	body, _, err := client.Get(cmd.Context(), "/v1/trucker-invoice-payments", query)
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

	rows := buildTruckerInvoicePaymentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTruckerInvoicePaymentsTable(cmd, rows)
}

func parseTruckerInvoicePaymentsListOptions(cmd *cobra.Command) (truckerInvoicePaymentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	truckerID, _ := cmd.Flags().GetString("trucker-id")
	trucker, _ := cmd.Flags().GetString("trucker")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return truckerInvoicePaymentsListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		TruckerID: truckerID,
		Trucker:   trucker,
	}, nil
}

func buildTruckerInvoicePaymentRows(resp jsonAPIResponse) []truckerInvoicePaymentRow {
	rows := make([]truckerInvoicePaymentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildTruckerInvoicePaymentRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildTruckerInvoicePaymentRow(resource jsonAPIResource) truckerInvoicePaymentRow {
	attrs := resource.Attributes
	row := truckerInvoicePaymentRow{
		ID:            resource.ID,
		PaymentAmount: stringAttr(attrs, "payment-amount"),
		PaymentDate:   formatDate(stringAttr(attrs, "payment-date")),
	}

	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}

	return row
}

func renderTruckerInvoicePaymentsTable(cmd *cobra.Command, rows []truckerInvoicePaymentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No trucker invoice payments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDATE\tAMOUNT\tTRUCKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.PaymentDate,
			row.PaymentAmount,
			row.TruckerID,
		)
	}
	return writer.Flush()
}
