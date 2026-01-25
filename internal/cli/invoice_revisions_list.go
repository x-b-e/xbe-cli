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

type invoiceRevisionsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Revision     string
	Invoice      string
	InvoiceID    string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type invoiceRevisionRow struct {
	ID          string `json:"id"`
	Revision    string `json:"revision,omitempty"`
	InvoiceType string `json:"invoice_type,omitempty"`
	InvoiceID   string `json:"invoice_id,omitempty"`
}

func newInvoiceRevisionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoice revisions",
		Long: `List invoice revisions with filtering and pagination.

Output Columns:
  ID       Invoice revision identifier
  REV      Revision number
  INVOICE  Invoice reference

Filters:
  --revision        Filter by revision number
  --invoice         Filter by invoice (format: Type|ID, e.g., broker-invoices|123)
  --invoice-id      Filter by invoice ID
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --is-created-at   Filter by has created-at (true/false)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)
  --is-updated-at   Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List invoice revisions
  xbe view invoice-revisions list

  # Filter by invoice revision number
  xbe view invoice-revisions list --revision 2

  # Filter by invoice
  xbe view invoice-revisions list --invoice broker-invoices|123

  # Output as JSON
  xbe view invoice-revisions list --json`,
		Args: cobra.NoArgs,
		RunE: runInvoiceRevisionsList,
	}
	initInvoiceRevisionsListFlags(cmd)
	return cmd
}

func init() {
	invoiceRevisionsCmd.AddCommand(newInvoiceRevisionsListCmd())
}

func initInvoiceRevisionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("revision", "", "Filter by revision number")
	cmd.Flags().String("invoice", "", "Filter by invoice (format: Type|ID, e.g., broker-invoices|123)")
	cmd.Flags().String("invoice-id", "", "Filter by invoice ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceRevisionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInvoiceRevisionsListOptions(cmd)
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
	query.Set("fields[invoice-revisions]", strings.Join([]string{
		"revision",
		"invoice",
	}, ","))

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[revision]", opts.Revision)
	setFilterIfPresent(query, "filter[invoice]", opts.Invoice)
	setFilterIfPresent(query, "filter[invoice-id]", opts.InvoiceID)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-revisions", query)
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

	rows := buildInvoiceRevisionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInvoiceRevisionsTable(cmd, rows)
}

func parseInvoiceRevisionsListOptions(cmd *cobra.Command) (invoiceRevisionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	revision, _ := cmd.Flags().GetString("revision")
	invoice, _ := cmd.Flags().GetString("invoice")
	invoiceID, _ := cmd.Flags().GetString("invoice-id")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceRevisionsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Revision:     revision,
		Invoice:      invoice,
		InvoiceID:    invoiceID,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildInvoiceRevisionRows(resp jsonAPIResponse) []invoiceRevisionRow {
	rows := make([]invoiceRevisionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildInvoiceRevisionRow(resource))
	}
	return rows
}

func buildInvoiceRevisionRow(resource jsonAPIResource) invoiceRevisionRow {
	attrs := resource.Attributes
	row := invoiceRevisionRow{
		ID:       resource.ID,
		Revision: stringAttr(attrs, "revision"),
	}

	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceType = rel.Data.Type
		row.InvoiceID = rel.Data.ID
	}

	return row
}

func renderInvoiceRevisionsTable(cmd *cobra.Command, rows []invoiceRevisionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No invoice revisions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREV\tINVOICE")
	for _, row := range rows {
		invoice := formatPolymorphic(row.InvoiceType, row.InvoiceID)
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.Revision,
			truncateString(invoice, 28),
		)
	}
	return writer.Flush()
}
