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

type timeCardInvoicesListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	ShiftStartsAfter string
	InvoiceType      string
	Seller           string
	InvoiceStatus    string
}

type timeCardInvoiceRow struct {
	ID            string `json:"id"`
	InvoiceID     string `json:"invoice_id,omitempty"`
	InvoiceType   string `json:"invoice_type,omitempty"`
	InvoiceStatus string `json:"invoice_status,omitempty"`
	SellerID      string `json:"seller_id,omitempty"`
	TimeCardID    string `json:"time_card_id,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

func newTimeCardInvoicesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time card invoices",
		Long: `List time card invoices with filtering and pagination.

Time card invoices associate time cards with invoices for billing and approval.

Output Columns:
  ID          Time card invoice identifier
  INVOICE     Invoice ID
  TYPE        Invoice type
  STATUS      Invoice status
  TIME_CARD   Time card ID
  CREATED_AT  Creation timestamp

Filters:
  --shift-starts-after  Filter by shift start on/after (ISO 8601)
  --invoice-type        Filter by invoice type (comma-separated for multiple)
  --invoice-status      Filter by invoice status (comma-separated for multiple)
  --seller              Filter by invoice seller ID (comma-separated for multiple)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time card invoices
  xbe view time-card-invoices list

  # Filter by shift start
  xbe view time-card-invoices list --shift-starts-after 2025-01-01T00:00:00Z

  # Filter by invoice status
  xbe view time-card-invoices list --invoice-status approved

  # Filter by seller
  xbe view time-card-invoices list --seller 123

  # Output as JSON
  xbe view time-card-invoices list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeCardInvoicesList,
	}
	initTimeCardInvoicesListFlags(cmd)
	return cmd
}

func init() {
	timeCardInvoicesCmd.AddCommand(newTimeCardInvoicesListCmd())
}

func initTimeCardInvoicesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("shift-starts-after", "", "Filter by shift start on/after (ISO 8601)")
	cmd.Flags().String("invoice-type", "", "Filter by invoice type (comma-separated for multiple)")
	cmd.Flags().String("invoice-status", "", "Filter by invoice status (comma-separated for multiple)")
	cmd.Flags().String("seller", "", "Filter by invoice seller ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardInvoicesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeCardInvoicesListOptions(cmd)
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
	query.Set("fields[time-card-invoices]", "created-at,updated-at,invoice,time-card")
	query.Set("fields[invoices]", "status,seller")
	query.Set("include", "invoice")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[shift-starts-after]", opts.ShiftStartsAfter)
	setFilterIfPresent(query, "filter[invoice-type]", opts.InvoiceType)
	setFilterIfPresent(query, "filter[invoice-status]", opts.InvoiceStatus)
	setFilterIfPresent(query, "filter[seller]", opts.Seller)

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-invoices", query)
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

	rows := buildTimeCardInvoiceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeCardInvoicesTable(cmd, rows)
}

func parseTimeCardInvoicesListOptions(cmd *cobra.Command) (timeCardInvoicesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	shiftStartsAfter, _ := cmd.Flags().GetString("shift-starts-after")
	invoiceType, _ := cmd.Flags().GetString("invoice-type")
	invoiceStatus, _ := cmd.Flags().GetString("invoice-status")
	seller, _ := cmd.Flags().GetString("seller")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardInvoicesListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		ShiftStartsAfter: shiftStartsAfter,
		InvoiceType:      invoiceType,
		InvoiceStatus:    invoiceStatus,
		Seller:           seller,
	}, nil
}

func buildTimeCardInvoiceRows(resp jsonAPIResponse) []timeCardInvoiceRow {
	included := map[string]jsonAPIResource{}
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]timeCardInvoiceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		invoiceID := relationshipIDFromMap(resource.Relationships, "invoice")
		timeCardID := relationshipIDFromMap(resource.Relationships, "time-card")
		row := timeCardInvoiceRow{
			ID:         resource.ID,
			InvoiceID:  invoiceID,
			TimeCardID: timeCardID,
			CreatedAt:  formatDateTime(stringAttr(resource.Attributes, "created-at")),
			UpdatedAt:  formatDateTime(stringAttr(resource.Attributes, "updated-at")),
		}

		if invoiceID != "" {
			if invoice, ok := included[resourceKey("invoices", invoiceID)]; ok {
				row.InvoiceStatus = stringAttr(invoice.Attributes, "status")
				row.InvoiceType = resolveInvoiceType(invoice.Attributes, invoice.Type)
				row.SellerID = relationshipIDFromMap(invoice.Relationships, "seller")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func resolveInvoiceType(attrs map[string]any, resourceType string) string {
	invoiceType := stringAttr(attrs, "type")
	if invoiceType == "" {
		invoiceType = stringAttr(attrs, "invoice-type")
	}
	if invoiceType == "" {
		invoiceType = strings.TrimSpace(resourceType)
	}
	return invoiceType
}

func renderTimeCardInvoicesTable(cmd *cobra.Command, rows []timeCardInvoiceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time card invoices found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINVOICE\tTYPE\tSTATUS\tTIME_CARD\tCREATED_AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.InvoiceID,
			truncateString(row.InvoiceType, 20),
			truncateString(row.InvoiceStatus, 20),
			row.TimeCardID,
			row.CreatedAt,
		)
	}
	return writer.Flush()
}
