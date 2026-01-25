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

type invoiceRevisionizingInvoiceRevisionsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	InvoiceRevisionizingWork string
	InvoiceRevision          string
	Invoice                  string
	CreatedAtMin             string
	CreatedAtMax             string
	UpdatedAtMin             string
	UpdatedAtMax             string
}

type invoiceRevisionizingInvoiceRevisionRow struct {
	ID                         string `json:"id"`
	InvoiceRevisionizingWorkID string `json:"invoice_revisionizing_work_id,omitempty"`
	InvoiceRevisionID          string `json:"invoice_revision_id,omitempty"`
	InvoiceID                  string `json:"invoice_id,omitempty"`
	RevisionNumber             string `json:"revision_number,omitempty"`
	CreatedAt                  string `json:"created_at,omitempty"`
	UpdatedAt                  string `json:"updated_at,omitempty"`
}

func newInvoiceRevisionizingInvoiceRevisionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoice revisionizing invoice revisions",
		Long: `List invoice revisionizing invoice revisions.

Output Columns:
  ID              Link identifier
  WORK            Invoice revisionizing work ID
  REVISION        Invoice revision ID
  INVOICE         Invoice ID
  REVISION NUMBER Revision number for the invoice
  CREATED AT      Creation timestamp
  UPDATED AT      Last update timestamp

Filters:
  --invoice-revisionizing-work  Filter by invoice revisionizing work ID
  --invoice-revision            Filter by invoice revision ID
  --invoice                     Filter by invoice ID
  --created-at-min              Filter by created-at on/after (ISO 8601)
  --created-at-max              Filter by created-at on/before (ISO 8601)
  --updated-at-min              Filter by updated-at on/after (ISO 8601)
  --updated-at-max              Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List invoice revisionizing invoice revisions
  xbe view invoice-revisionizing-invoice-revisions list

  # Filter by revisionizing work
  xbe view invoice-revisionizing-invoice-revisions list --invoice-revisionizing-work 123

  # Filter by invoice revision
  xbe view invoice-revisionizing-invoice-revisions list --invoice-revision 456

  # Filter by invoice
  xbe view invoice-revisionizing-invoice-revisions list --invoice 789

  # Output as JSON
  xbe view invoice-revisionizing-invoice-revisions list --json`,
		Args: cobra.NoArgs,
		RunE: runInvoiceRevisionizingInvoiceRevisionsList,
	}
	initInvoiceRevisionizingInvoiceRevisionsListFlags(cmd)
	return cmd
}

func init() {
	invoiceRevisionizingInvoiceRevisionsCmd.AddCommand(newInvoiceRevisionizingInvoiceRevisionsListCmd())
}

func initInvoiceRevisionizingInvoiceRevisionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("invoice-revisionizing-work", "", "Filter by invoice revisionizing work ID")
	cmd.Flags().String("invoice-revision", "", "Filter by invoice revision ID")
	cmd.Flags().String("invoice", "", "Filter by invoice ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceRevisionizingInvoiceRevisionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInvoiceRevisionizingInvoiceRevisionsListOptions(cmd)
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
	query.Set("fields[invoice-revisionizing-invoice-revisions]", "revision-number,created-at,updated-at,invoice-revisionizing-work,invoice-revision,invoice")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[invoice_revisionizing_work]", opts.InvoiceRevisionizingWork)
	setFilterIfPresent(query, "filter[invoice_revision]", opts.InvoiceRevision)
	setFilterIfPresent(query, "filter[invoice]", opts.Invoice)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-revisionizing-invoice-revisions", query)
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

	rows := buildInvoiceRevisionizingInvoiceRevisionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInvoiceRevisionizingInvoiceRevisionsTable(cmd, rows)
}

func parseInvoiceRevisionizingInvoiceRevisionsListOptions(cmd *cobra.Command) (invoiceRevisionizingInvoiceRevisionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	invoiceRevisionizingWork, _ := cmd.Flags().GetString("invoice-revisionizing-work")
	invoiceRevision, _ := cmd.Flags().GetString("invoice-revision")
	invoice, _ := cmd.Flags().GetString("invoice")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceRevisionizingInvoiceRevisionsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		InvoiceRevisionizingWork: invoiceRevisionizingWork,
		InvoiceRevision:          invoiceRevision,
		Invoice:                  invoice,
		CreatedAtMin:             createdAtMin,
		CreatedAtMax:             createdAtMax,
		UpdatedAtMin:             updatedAtMin,
		UpdatedAtMax:             updatedAtMax,
	}, nil
}

func buildInvoiceRevisionizingInvoiceRevisionRows(resp jsonAPIResponse) []invoiceRevisionizingInvoiceRevisionRow {
	rows := make([]invoiceRevisionizingInvoiceRevisionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := invoiceRevisionizingInvoiceRevisionRow{
			ID:             resource.ID,
			RevisionNumber: stringAttr(attrs, "revision-number"),
			CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
			UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
		}

		if rel, ok := resource.Relationships["invoice-revisionizing-work"]; ok && rel.Data != nil {
			row.InvoiceRevisionizingWorkID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["invoice-revision"]; ok && rel.Data != nil {
			row.InvoiceRevisionID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
			row.InvoiceID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildInvoiceRevisionizingInvoiceRevisionRowFromSingle(resp jsonAPISingleResponse) invoiceRevisionizingInvoiceRevisionRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := invoiceRevisionizingInvoiceRevisionRow{
		ID:             resource.ID,
		RevisionNumber: stringAttr(attrs, "revision-number"),
		CreatedAt:      formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:      formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["invoice-revisionizing-work"]; ok && rel.Data != nil {
		row.InvoiceRevisionizingWorkID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["invoice-revision"]; ok && rel.Data != nil {
		row.InvoiceRevisionID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
		row.InvoiceID = rel.Data.ID
	}

	return row
}

func renderInvoiceRevisionizingInvoiceRevisionsTable(cmd *cobra.Command, rows []invoiceRevisionizingInvoiceRevisionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No invoice revisionizing invoice revisions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tWORK\tREVISION\tINVOICE\tREVISION NUMBER\tCREATED AT\tUPDATED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.InvoiceRevisionizingWorkID,
			row.InvoiceRevisionID,
			row.InvoiceID,
			row.RevisionNumber,
			row.CreatedAt,
			row.UpdatedAt,
		)
	}
	writer.Flush()
	return nil
}
