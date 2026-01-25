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

type organizationInvoicesBatchInvoicesListOptions struct {
	BaseURL                   string
	Token                     string
	JSON                      bool
	NoAuth                    bool
	Limit                     int
	Offset                    int
	Sort                      string
	OrganizationInvoicesBatch string
	Invoice                   string
	InvoiceID                 string
	Organization              string
	OrganizationType          string
	OrganizationID            string
	CreatedBy                 string
	ChangedBy                 string
	Successful                string
}

type organizationInvoicesBatchInvoiceRow struct {
	ID                          string `json:"id"`
	Status                      string `json:"status,omitempty"`
	BatchStatus                 string `json:"batch_status,omitempty"`
	InvoiceAmount               string `json:"invoice_amount,omitempty"`
	IsRevised                   bool   `json:"is_revised"`
	InvoiceID                   string `json:"invoice_id,omitempty"`
	InvoiceType                 string `json:"invoice_type,omitempty"`
	OrganizationInvoicesBatchID string `json:"organization_invoices_batch_id,omitempty"`
	OrganizationID              string `json:"organization_id,omitempty"`
	OrganizationType            string `json:"organization_type,omitempty"`
	CreatedByID                 string `json:"created_by_id,omitempty"`
	ChangedByID                 string `json:"changed_by_id,omitempty"`
	UpdatedByID                 string `json:"updated_by_id,omitempty"`
}

func newOrganizationInvoicesBatchInvoicesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch invoices",
		Long: `List organization invoices batch invoices with filtering and pagination.

Output Columns:
  ID              Batch invoice identifier
  INVOICE         Invoice (Type/ID)
  BATCH           Organization invoices batch ID
  ORGANIZATION    Organization (Type/ID)
  STATUS          Batch invoice status
  BATCH STATUS    Batch status for the invoice
  INVOICE AMOUNT  Invoice amount from the revision
  REVISED         Whether the invoice revision differs

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  --organization-invoices-batch  Filter by organization invoices batch ID
  --invoice                      Filter by invoice (format: Type|ID, e.g., BrokerInvoice|123)
  --invoice-id                   Filter by invoice ID
  --organization                 Filter by organization (format: Type|ID, e.g., Broker|123)
  --organization-id              Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type            Filter by organization type (e.g., Broker, Customer, Trucker)
  --created-by                   Filter by created-by user ID
  --changed-by                   Filter by changed-by user ID
  --successful                   Filter by successful status (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List batch invoices
  xbe view organization-invoices-batch-invoices list

  # Filter by batch
  xbe view organization-invoices-batch-invoices list --organization-invoices-batch 123

  # Filter by invoice ID
  xbe view organization-invoices-batch-invoices list --invoice-id 456

  # Filter by organization
  xbe view organization-invoices-batch-invoices list --organization "Broker|123"

  # Filter by successful status
  xbe view organization-invoices-batch-invoices list --successful true

  # Output as JSON
  xbe view organization-invoices-batch-invoices list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchInvoicesList,
	}
	initOrganizationInvoicesBatchInvoicesListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchInvoicesCmd.AddCommand(newOrganizationInvoicesBatchInvoicesListCmd())
}

func initOrganizationInvoicesBatchInvoicesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-invoices-batch", "", "Filter by organization invoices batch ID")
	cmd.Flags().String("invoice", "", "Filter by invoice (format: Type|ID)")
	cmd.Flags().String("invoice-id", "", "Filter by invoice ID")
	cmd.Flags().String("organization", "", "Filter by organization (format: Type|ID, e.g., Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer, Trucker)")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("changed-by", "", "Filter by changed-by user ID")
	cmd.Flags().String("successful", "", "Filter by successful status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchInvoicesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchInvoicesListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-invoices]", "status,batch-status,invoice-amount,is-revised,invoice,organization-invoices-batch,organization,created-by,updated-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization-invoices-batch]", opts.OrganizationInvoicesBatch)
	setFilterIfPresent(query, "filter[invoice]", opts.Invoice)
	setFilterIfPresent(query, "filter[invoice-id]", opts.InvoiceID)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	organizationIDFilter, err := buildOrganizationIDFilter(opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if organizationIDFilter != "" {
		query.Set("filter[organization-id]", organizationIDFilter)
	} else {
		setFilterIfPresent(query, "filter[organization-type]", opts.OrganizationType)
	}
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[changed-by]", opts.ChangedBy)
	setFilterIfPresent(query, "filter[successful]", opts.Successful)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-invoices", query)
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

	rows := buildOrganizationInvoicesBatchInvoiceRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchInvoicesTable(cmd, rows)
}

func parseOrganizationInvoicesBatchInvoicesListOptions(cmd *cobra.Command) (organizationInvoicesBatchInvoicesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organizationInvoicesBatch, _ := cmd.Flags().GetString("organization-invoices-batch")
	invoice, _ := cmd.Flags().GetString("invoice")
	invoiceID, _ := cmd.Flags().GetString("invoice-id")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	createdBy, _ := cmd.Flags().GetString("created-by")
	changedBy, _ := cmd.Flags().GetString("changed-by")
	successful, _ := cmd.Flags().GetString("successful")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchInvoicesListOptions{
		BaseURL:                   baseURL,
		Token:                     token,
		JSON:                      jsonOut,
		NoAuth:                    noAuth,
		Limit:                     limit,
		Offset:                    offset,
		Sort:                      sort,
		OrganizationInvoicesBatch: organizationInvoicesBatch,
		Invoice:                   invoice,
		InvoiceID:                 invoiceID,
		Organization:              organization,
		OrganizationType:          organizationType,
		OrganizationID:            organizationID,
		CreatedBy:                 createdBy,
		ChangedBy:                 changedBy,
		Successful:                successful,
	}, nil
}

func buildOrganizationInvoicesBatchInvoiceRows(resp jsonAPIResponse) []organizationInvoicesBatchInvoiceRow {
	rows := make([]organizationInvoicesBatchInvoiceRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := organizationInvoicesBatchInvoiceRow{
			ID:            resource.ID,
			Status:        stringAttr(resource.Attributes, "status"),
			BatchStatus:   stringAttr(resource.Attributes, "batch-status"),
			InvoiceAmount: stringAttr(resource.Attributes, "invoice-amount"),
			IsRevised:     boolAttr(resource.Attributes, "is-revised"),
		}
		if rel, ok := resource.Relationships["invoice"]; ok && rel.Data != nil {
			row.InvoiceID = rel.Data.ID
			row.InvoiceType = rel.Data.Type
		}
		if rel, ok := resource.Relationships["organization-invoices-batch"]; ok && rel.Data != nil {
			row.OrganizationInvoicesBatchID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationID = rel.Data.ID
			row.OrganizationType = rel.Data.Type
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
			row.ChangedByID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["updated-by"]; ok && rel.Data != nil {
			row.UpdatedByID = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderOrganizationInvoicesBatchInvoicesTable(cmd *cobra.Command, rows []organizationInvoicesBatchInvoiceRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch invoices found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINVOICE\tBATCH\tORGANIZATION\tSTATUS\tBATCH STATUS\tINVOICE AMOUNT\tREVISED")
	for _, row := range rows {
		invoice := formatPolymorphic(row.InvoiceType, row.InvoiceID)
		organization := formatPolymorphic(row.OrganizationType, row.OrganizationID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\n",
			row.ID,
			invoice,
			row.OrganizationInvoicesBatchID,
			organization,
			row.Status,
			row.BatchStatus,
			formatAnyValue(row.InvoiceAmount),
			row.IsRevised,
		)
	}
	return writer.Flush()
}
