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

type organizationInvoicesBatchesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Processed    string
	Broker       string
	CreatedBy    string
	ChangedBy    string
	Organization string
	Invoices     string
	InvoicesID   string
	CreatedAtMin string
	CreatedAtMax string
}

type organizationInvoicesBatchRow struct {
	ID               string   `json:"id"`
	Status           string   `json:"status,omitempty"`
	InvoiceTypes     []string `json:"invoice_types,omitempty"`
	OrganizationType string   `json:"organization_type,omitempty"`
	OrganizationID   string   `json:"organization_id,omitempty"`
	BrokerID         string   `json:"broker_id,omitempty"`
	CreatedByID      string   `json:"created_by_id,omitempty"`
	UpdatedByID      string   `json:"updated_by_id,omitempty"`
}

func newOrganizationInvoicesBatchesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batches",
		Long: `List organization invoices batches with filtering and pagination.

Output Columns:
  ID            Organization invoices batch identifier
  STATUS        Processing status
  ORG TYPE      Organization type
  ORG ID        Organization ID
  BROKER        Broker ID
  INVOICE TYPES Invoice type list
  CREATED BY    Creator user ID

Filters:
  --processed        Filter by processed status (true/false)
  --broker           Filter by broker ID (comma-separated for multiple)
  --created-by       Filter by created by user ID (comma-separated for multiple)
  --changed-by       Filter by changed by user ID (comma-separated for multiple)
  --organization     Filter by organization (Type|ID, e.g. Broker|123)
  --invoices         Filter by invoice IDs (comma-separated for multiple)
  --invoices-id      Filter by invoice IDs using invoices_id filter (comma-separated)
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List organization invoices batches
  xbe view organization-invoices-batches list

  # Filter by processed status
  xbe view organization-invoices-batches list --processed true

  # Filter by organization
  xbe view organization-invoices-batches list --organization "Broker|123"

  # Filter by broker
  xbe view organization-invoices-batches list --broker 456

  # Output as JSON
  xbe view organization-invoices-batches list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchesList,
	}
	initOrganizationInvoicesBatchesListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchesCmd.AddCommand(newOrganizationInvoicesBatchesListCmd())
}

func initOrganizationInvoicesBatchesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("processed", "", "Filter by processed status (true/false)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("created-by", "", "Filter by created by user ID (comma-separated for multiple)")
	cmd.Flags().String("changed-by", "", "Filter by changed by user ID (comma-separated for multiple)")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Broker|123)")
	cmd.Flags().String("invoices", "", "Filter by invoice IDs (comma-separated for multiple)")
	cmd.Flags().String("invoices-id", "", "Filter by invoice IDs using invoices_id filter (comma-separated)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchesListOptions(cmd)
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
	query.Set("fields[organization-invoices-batches]", "status,invoice-types,organization,broker,created-by,updated-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[processed]", opts.Processed)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[changed-by]", opts.ChangedBy)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[invoices]", opts.Invoices)
	setFilterIfPresent(query, "filter[invoices_id]", opts.InvoicesID)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batches", query)
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

	rows := buildOrganizationInvoicesBatchRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchesTable(cmd, rows)
}

func parseOrganizationInvoicesBatchesListOptions(cmd *cobra.Command) (organizationInvoicesBatchesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	processed, _ := cmd.Flags().GetString("processed")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	changedBy, _ := cmd.Flags().GetString("changed-by")
	organization, _ := cmd.Flags().GetString("organization")
	invoices, _ := cmd.Flags().GetString("invoices")
	invoicesID, _ := cmd.Flags().GetString("invoices-id")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Processed:    processed,
		Broker:       broker,
		CreatedBy:    createdBy,
		ChangedBy:    changedBy,
		Organization: organization,
		Invoices:     invoices,
		InvoicesID:   invoicesID,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
	}, nil
}

func buildOrganizationInvoicesBatchRows(resp jsonAPIResponse) []organizationInvoicesBatchRow {
	rows := make([]organizationInvoicesBatchRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildOrganizationInvoicesBatchRow(resource))
	}
	return rows
}

func buildOrganizationInvoicesBatchRow(resource jsonAPIResource) organizationInvoicesBatchRow {
	attrs := resource.Attributes
	row := organizationInvoicesBatchRow{
		ID:           resource.ID,
		Status:       stringAttr(attrs, "status"),
		InvoiceTypes: stringSliceAttr(attrs, "invoice-types"),
		BrokerID:     relationshipIDFromMap(resource.Relationships, "broker"),
		CreatedByID:  relationshipIDFromMap(resource.Relationships, "created-by"),
		UpdatedByID:  relationshipIDFromMap(resource.Relationships, "updated-by"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}

func renderOrganizationInvoicesBatchesTable(cmd *cobra.Command, rows []organizationInvoicesBatchRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batches found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tORG TYPE\tORG ID\tBROKER\tINVOICE TYPES\tCREATED BY")
	for _, row := range rows {
		invoiceTypes := strings.Join(row.InvoiceTypes, ",")
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 12),
			truncateString(row.OrganizationType, 12),
			truncateString(row.OrganizationID, 14),
			truncateString(row.BrokerID, 14),
			truncateString(invoiceTypes, 24),
			truncateString(row.CreatedByID, 14),
		)
	}

	return writer.Flush()
}
