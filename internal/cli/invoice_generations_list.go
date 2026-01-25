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

type invoiceGenerationsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Organization        string
	OrganizationID      string
	OrganizationType    string
	NotOrganizationType string
	TimeCards           string
	CompletedAtMin      string
	CompletedAtMax      string
	IsCompletedAt       string
	IsCompleted         string
	CreatedAtMin        string
	CreatedAtMax        string
	IsCreatedAt         string
	UpdatedAtMin        string
	UpdatedAtMax        string
	IsUpdatedAt         string
}

type invoiceGenerationRow struct {
	ID               string   `json:"id"`
	Status           string   `json:"status,omitempty"`
	Organization     string   `json:"organization,omitempty"`
	OrganizationType string   `json:"organization_type,omitempty"`
	OrganizationID   string   `json:"organization_id,omitempty"`
	InvoicingDate    string   `json:"invoicing_date,omitempty"`
	CompletedAt      string   `json:"completed_at,omitempty"`
	IsRunning        bool     `json:"is_running,omitempty"`
	IsParent         bool     `json:"is_parent,omitempty"`
	IsCompleted      bool     `json:"is_completed,omitempty"`
	TimeCardCount    int      `json:"time_card_count,omitempty"`
	TimeCardIDs      []string `json:"time_card_ids,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
}

func newInvoiceGenerationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoice generations",
		Long: `List invoice generations with filtering and pagination.

Output Columns:
  ID         Invoice generation identifier
  STATUS     Generation status
  ORG        Organization name or type/id
  INVOICING  Invoicing date
  COMPLETED  Completion timestamp
  RUNNING    Running status
  PARENT     Parent generation flag
  CARDS      Time card count

Filters:
  --organization           Filter by organization (format: Type|ID, e.g., Broker|123)
  --organization-id        Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type      Filter by organization type (e.g., Broker, Customer, Trucker)
  --not-organization-type  Exclude organization type (e.g., Customer)
  --time-cards             Filter by time card IDs (comma-separated)
  --completed-at-min       Filter by completed-at on/after (ISO 8601)
  --completed-at-max       Filter by completed-at on/before (ISO 8601)
  --is-completed-at        Filter by has completed-at (true/false)
  --is-completed           Filter by completion status (true/false)
  --created-at-min         Filter by created-at on/after (ISO 8601)
  --created-at-max         Filter by created-at on/before (ISO 8601)
  --is-created-at          Filter by has created-at (true/false)
  --updated-at-min         Filter by updated-at on/after (ISO 8601)
  --updated-at-max         Filter by updated-at on/before (ISO 8601)
  --is-updated-at          Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List invoice generations
  xbe view invoice-generations list

  # Filter by organization
  xbe view invoice-generations list --organization "Broker|123"

  # Filter by time cards
  xbe view invoice-generations list --time-cards 101,102

  # Output as JSON
  xbe view invoice-generations list --json`,
		Args: cobra.NoArgs,
		RunE: runInvoiceGenerationsList,
	}
	initInvoiceGenerationsListFlags(cmd)
	return cmd
}

func init() {
	invoiceGenerationsCmd.AddCommand(newInvoiceGenerationsListCmd())
}

func initInvoiceGenerationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization", "", "Filter by organization (format: Type|ID, e.g., Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer, Trucker)")
	cmd.Flags().String("not-organization-type", "", "Exclude organization type (e.g., Customer)")
	cmd.Flags().String("time-cards", "", "Filter by time card IDs (comma-separated)")
	cmd.Flags().String("completed-at-min", "", "Filter by completed-at on/after (ISO 8601)")
	cmd.Flags().String("completed-at-max", "", "Filter by completed-at on/before (ISO 8601)")
	cmd.Flags().String("is-completed-at", "", "Filter by has completed-at (true/false)")
	cmd.Flags().String("is-completed", "", "Filter by completion status (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runInvoiceGenerationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseInvoiceGenerationsListOptions(cmd)
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
	query.Set("fields[invoice-generations]", strings.Join([]string{
		"status",
		"invoicing-date",
		"completed-at",
		"is-running",
		"is-parent",
		"is-completed",
		"time-card-ids",
		"organization",
		"created-at",
		"updated-at",
	}, ","))
	query.Set("include", "organization")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")

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
	setFilterIfPresent(query, "filter[time-cards]", opts.TimeCards)
	setFilterIfPresent(query, "filter[completed-at-min]", opts.CompletedAtMin)
	setFilterIfPresent(query, "filter[completed-at-max]", opts.CompletedAtMax)
	setFilterIfPresent(query, "filter[is-completed-at]", opts.IsCompletedAt)
	setFilterIfPresent(query, "filter[is-completed]", opts.IsCompleted)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/invoice-generations", query)
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

	rows := buildInvoiceGenerationRows(resp)
	if strings.TrimSpace(opts.NotOrganizationType) != "" {
		rows = filterInvoiceGenerationsByOrganizationType(rows, opts.NotOrganizationType)
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderInvoiceGenerationsTable(cmd, rows)
}

func parseInvoiceGenerationsListOptions(cmd *cobra.Command) (invoiceGenerationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organization, _ := cmd.Flags().GetString("organization")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	notOrganizationType, _ := cmd.Flags().GetString("not-organization-type")
	timeCards, _ := cmd.Flags().GetString("time-cards")
	completedAtMin, _ := cmd.Flags().GetString("completed-at-min")
	completedAtMax, _ := cmd.Flags().GetString("completed-at-max")
	isCompletedAt, _ := cmd.Flags().GetString("is-completed-at")
	isCompleted, _ := cmd.Flags().GetString("is-completed")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return invoiceGenerationsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Organization:        organization,
		OrganizationID:      organizationID,
		OrganizationType:    organizationType,
		NotOrganizationType: notOrganizationType,
		TimeCards:           timeCards,
		CompletedAtMin:      completedAtMin,
		CompletedAtMax:      completedAtMax,
		IsCompletedAt:       isCompletedAt,
		IsCompleted:         isCompleted,
		CreatedAtMin:        createdAtMin,
		CreatedAtMax:        createdAtMax,
		IsCreatedAt:         isCreatedAt,
		UpdatedAtMin:        updatedAtMin,
		UpdatedAtMax:        updatedAtMax,
		IsUpdatedAt:         isUpdatedAt,
	}, nil
}

func buildInvoiceGenerationRows(resp jsonAPIResponse) []invoiceGenerationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]invoiceGenerationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildInvoiceGenerationRow(resource, included))
	}

	return rows
}

func invoiceGenerationRowFromSingle(resp jsonAPISingleResponse) invoiceGenerationRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildInvoiceGenerationRow(resp.Data, included)
}

func buildInvoiceGenerationRow(resource jsonAPIResource, included map[string]jsonAPIResource) invoiceGenerationRow {
	attrs := resource.Attributes
	cardIDs := stringSliceAttr(attrs, "time-card-ids")

	row := invoiceGenerationRow{
		ID:            resource.ID,
		Status:        stringAttr(attrs, "status"),
		InvoicingDate: formatDate(stringAttr(attrs, "invoicing-date")),
		CompletedAt:   formatDateTime(stringAttr(attrs, "completed-at")),
		IsRunning:     boolAttr(attrs, "is-running"),
		IsParent:      boolAttr(attrs, "is-parent"),
		IsCompleted:   boolAttr(attrs, "is-completed"),
		TimeCardIDs:   cardIDs,
		TimeCardCount: len(cardIDs),
		CreatedAt:     formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:     formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.Organization = firstNonEmpty(
				stringAttr(org.Attributes, "company-name"),
				stringAttr(org.Attributes, "name"),
			)
		}
	}

	return row
}

func renderInvoiceGenerationsTable(cmd *cobra.Command, rows []invoiceGenerationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No invoice generations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tORG\tINVOICING\tCOMPLETED\tRUNNING\tPARENT\tCARDS")
	for _, row := range rows {
		orgLabel := formatRelated(row.Organization, formatPolymorphic(row.OrganizationType, row.OrganizationID))
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\t%t\t%d\n",
			row.ID,
			row.Status,
			truncateString(orgLabel, 28),
			row.InvoicingDate,
			row.CompletedAt,
			row.IsRunning,
			row.IsParent,
			row.TimeCardCount,
		)
	}
	return writer.Flush()
}

func filterInvoiceGenerationsByOrganizationType(rows []invoiceGenerationRow, organizationType string) []invoiceGenerationRow {
	filterType := normalizeOrganizationType(organizationType)
	if filterType == "" {
		return rows
	}
	filtered := make([]invoiceGenerationRow, 0, len(rows))
	for _, row := range rows {
		if normalizeOrganizationType(row.OrganizationType) != filterType {
			filtered = append(filtered, row)
		}
	}
	return filtered
}
