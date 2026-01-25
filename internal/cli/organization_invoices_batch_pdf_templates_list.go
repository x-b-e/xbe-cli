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

type organizationInvoicesBatchPdfTemplatesListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	Organization     string
	OrganizationType string
	OrganizationID   string
	Broker           string
	IsActive         string
	IsGlobal         string
	CreatedBy        string
	CreatedAtMin     string
	CreatedAtMax     string
	IsCreatedAt      string
}

type organizationInvoicesBatchPdfTemplateRow struct {
	ID               string `json:"id"`
	Description      string `json:"description,omitempty"`
	IsActive         bool   `json:"is_active"`
	IsGlobal         bool   `json:"is_global"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
}

func newOrganizationInvoicesBatchPdfTemplatesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organization invoices batch PDF templates",
		Long: `List organization invoices batch PDF templates with filtering and pagination.

Output Columns:
  ID            Template identifier
  DESCRIPTION   Template description
  ACTIVE        Whether the template is active
  GLOBAL        Whether the template is global
  ORGANIZATION  Organization (Type/ID)
  BROKER        Broker ID
  CREATED BY    Created-by user ID
  CREATED AT    Created timestamp

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  --organization      Filter by organization (format: Type|ID, e.g., Broker|123)
  --organization-id   Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type Filter by organization type (e.g., Broker, Customer, Trucker)
  --broker            Filter by broker ID
  --is-active         Filter by active status (true/false)
  --is-global         Filter by global status (true/false)
  --created-by        Filter by created-by user ID
  --created-at-min    Filter by created-at on/after (ISO 8601)
  --created-at-max    Filter by created-at on/before (ISO 8601)
  --is-created-at     Filter by has created-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List templates
  xbe view organization-invoices-batch-pdf-templates list

  # Filter by organization
  xbe view organization-invoices-batch-pdf-templates list --organization "Broker|123"

  # Filter by broker
  xbe view organization-invoices-batch-pdf-templates list --broker 456

  # Filter by global templates
  xbe view organization-invoices-batch-pdf-templates list --is-global true

  # Output as JSON
  xbe view organization-invoices-batch-pdf-templates list --json`,
		Args: cobra.NoArgs,
		RunE: runOrganizationInvoicesBatchPdfTemplatesList,
	}
	initOrganizationInvoicesBatchPdfTemplatesListFlags(cmd)
	return cmd
}

func init() {
	organizationInvoicesBatchPdfTemplatesCmd.AddCommand(newOrganizationInvoicesBatchPdfTemplatesListCmd())
}

func initOrganizationInvoicesBatchPdfTemplatesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization", "", "Filter by organization (format: Type|ID, e.g., Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer, Trucker)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("is-global", "", "Filter by global status (true/false)")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOrganizationInvoicesBatchPdfTemplatesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOrganizationInvoicesBatchPdfTemplatesListOptions(cmd)
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
	query.Set("fields[organization-invoices-batch-pdf-templates]", strings.Join([]string{
		"description",
		"is-active",
		"is-global",
		"organization",
		"broker",
		"created-by",
		"created-at",
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
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[is-active]", opts.IsActive)
	setFilterIfPresent(query, "filter[is-global]", opts.IsGlobal)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/organization-invoices-batch-pdf-templates", query)
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

	rows := buildOrganizationInvoicesBatchPdfTemplateRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOrganizationInvoicesBatchPdfTemplatesTable(cmd, rows)
}

func parseOrganizationInvoicesBatchPdfTemplatesListOptions(cmd *cobra.Command) (organizationInvoicesBatchPdfTemplatesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	broker, _ := cmd.Flags().GetString("broker")
	isActive, _ := cmd.Flags().GetString("is-active")
	isGlobal, _ := cmd.Flags().GetString("is-global")
	createdBy, _ := cmd.Flags().GetString("created-by")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return organizationInvoicesBatchPdfTemplatesListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		Organization:     organization,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
		Broker:           broker,
		IsActive:         isActive,
		IsGlobal:         isGlobal,
		CreatedBy:        createdBy,
		CreatedAtMin:     createdAtMin,
		CreatedAtMax:     createdAtMax,
		IsCreatedAt:      isCreatedAt,
	}, nil
}

func buildOrganizationInvoicesBatchPdfTemplateRows(resp jsonAPIResponse) []organizationInvoicesBatchPdfTemplateRow {
	rows := make([]organizationInvoicesBatchPdfTemplateRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildOrganizationInvoicesBatchPdfTemplateRow(resource))
	}
	return rows
}

func buildOrganizationInvoicesBatchPdfTemplateRow(resource jsonAPIResource) organizationInvoicesBatchPdfTemplateRow {
	row := organizationInvoicesBatchPdfTemplateRow{
		ID:          resource.ID,
		Description: stringAttr(resource.Attributes, "description"),
		IsActive:    boolAttr(resource.Attributes, "is-active"),
		IsGlobal:    boolAttr(resource.Attributes, "is-global"),
		CreatedAt:   formatDateTime(stringAttr(resource.Attributes, "created-at")),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func buildOrganizationInvoicesBatchPdfTemplateRowFromSingle(resp jsonAPISingleResponse) organizationInvoicesBatchPdfTemplateRow {
	return buildOrganizationInvoicesBatchPdfTemplateRow(resp.Data)
}

func renderOrganizationInvoicesBatchPdfTemplatesTable(cmd *cobra.Command, rows []organizationInvoicesBatchPdfTemplateRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No organization invoices batch PDF templates found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDESCRIPTION\tACTIVE\tGLOBAL\tORGANIZATION\tBROKER\tCREATED BY\tCREATED AT")
	for _, row := range rows {
		organization := formatPolymorphic(row.OrganizationType, row.OrganizationID)
		if organization == "" {
			organization = "-"
		}
		broker := row.BrokerID
		if broker == "" {
			broker = "-"
		}
		createdBy := row.CreatedByID
		if createdBy == "" {
			createdBy = "-"
		}
		createdAt := row.CreatedAt
		if createdAt == "" {
			createdAt = "-"
		}

		fmt.Fprintf(writer, "%s\t%s\t%t\t%t\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Description, 40),
			row.IsActive,
			row.IsGlobal,
			organization,
			broker,
			createdBy,
			createdAt,
		)
	}
	return writer.Flush()
}
