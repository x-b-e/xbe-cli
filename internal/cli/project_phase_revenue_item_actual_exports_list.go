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

type projectPhaseRevenueItemActualExportsListOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	NoAuth                   bool
	Limit                    int
	Offset                   int
	Sort                     string
	OrganizationFormatter    string
	Status                   string
	Broker                   string
	Project                  string
	CreatedBy                string
	Organization             string
	OrganizationID           string
	OrganizationType         string
	NotOrganizationType      string
	ProjectPhaseRevenueItems string
	RevenueDate              string
	RevenueDateMin           string
	RevenueDateMax           string
	HasRevenueDate           string
}

type projectPhaseRevenueItemActualExportRow struct {
	ID                      string `json:"id"`
	Status                  string `json:"status,omitempty"`
	FileName                string `json:"file_name,omitempty"`
	RevenueDate             string `json:"revenue_date,omitempty"`
	OrganizationType        string `json:"organization_type,omitempty"`
	OrganizationID          string `json:"organization_id,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	OrganizationFormatterID string `json:"organization_formatter_id,omitempty"`
	CreatedByID             string `json:"created_by_id,omitempty"`
}

func newProjectPhaseRevenueItemActualExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project phase revenue item actual exports",
		Long: `List project phase revenue item actual exports with filtering and pagination.

Output Columns:
  ID            Export identifier
  STATUS        Processing status
  REVENUE DATE  Revenue date for exported actuals
  FILE NAME     Generated file name
  ORG TYPE      Organization type
  ORG ID        Organization ID
  BROKER        Broker ID
  FORMATTER     Organization formatter ID
  CREATED BY    Creator user ID

Filters:
  --organization-formatter       Filter by organization formatter ID
  --status                       Filter by status (processing, processed, failed)
  --broker                       Filter by broker ID
  --project                      Filter by project ID
  --created-by                   Filter by created-by user ID
  --organization                 Filter by organization (Type|ID, e.g. Broker|123)
  --organization-id              Filter by organization ID (requires --organization-type)
  --organization-type            Filter by organization type (e.g. Broker, Customer)
  --not-organization-type        Exclude organization type (e.g. Broker)
  --project-phase-revenue-items  Filter by project phase revenue item IDs (comma-separated)
  --revenue-date                 Filter by revenue date (YYYY-MM-DD)
  --revenue-date-min             Filter by revenue date on/after (YYYY-MM-DD)
  --revenue-date-max             Filter by revenue date on/before (YYYY-MM-DD)
  --has-revenue-date             Filter by presence of revenue date (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List exports
  xbe view project-phase-revenue-item-actual-exports list

  # Filter by status
  xbe view project-phase-revenue-item-actual-exports list --status processed

  # Filter by revenue date
  xbe view project-phase-revenue-item-actual-exports list --revenue-date 2025-01-15

  # Output as JSON
  xbe view project-phase-revenue-item-actual-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectPhaseRevenueItemActualExportsList,
	}
	initProjectPhaseRevenueItemActualExportsListFlags(cmd)
	return cmd
}

func init() {
	projectPhaseRevenueItemActualExportsCmd.AddCommand(newProjectPhaseRevenueItemActualExportsListCmd())
}

func initProjectPhaseRevenueItemActualExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization-formatter", "", "Filter by organization formatter ID")
	cmd.Flags().String("status", "", "Filter by status (processing, processed, failed)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g. Broker)")
	cmd.Flags().String("not-organization-type", "", "Exclude organization type (e.g. Broker)")
	cmd.Flags().String("project-phase-revenue-items", "", "Filter by project phase revenue item IDs (comma-separated)")
	cmd.Flags().String("revenue-date", "", "Filter by revenue date (YYYY-MM-DD)")
	cmd.Flags().String("revenue-date-min", "", "Filter by revenue date on/after (YYYY-MM-DD)")
	cmd.Flags().String("revenue-date-max", "", "Filter by revenue date on/before (YYYY-MM-DD)")
	cmd.Flags().String("has-revenue-date", "", "Filter by presence of revenue date (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseRevenueItemActualExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectPhaseRevenueItemActualExportsListOptions(cmd)
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
	query.Set("fields[project-phase-revenue-item-actual-exports]", "status,file-name,revenue-date,organization,broker,organization-formatter,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization_formatter]", opts.OrganizationFormatter)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	if opts.OrganizationID != "" {
		if strings.Contains(opts.OrganizationID, "|") {
			query.Set("filter[organization_id]", opts.OrganizationID)
		} else if opts.OrganizationType != "" {
			query.Set("filter[organization_id]", opts.OrganizationType+"|"+opts.OrganizationID)
		} else {
			return fmt.Errorf("--organization-id requires --organization-type")
		}
	}
	setFilterIfPresent(query, "filter[organization_type]", opts.OrganizationType)
	setFilterIfPresent(query, "filter[not_organization_type]", opts.NotOrganizationType)
	setFilterIfPresent(query, "filter[project_phase_revenue_items]", opts.ProjectPhaseRevenueItems)
	setFilterIfPresent(query, "filter[revenue-date]", opts.RevenueDate)
	setFilterIfPresent(query, "filter[revenue-date-min]", opts.RevenueDateMin)
	setFilterIfPresent(query, "filter[revenue-date-max]", opts.RevenueDateMax)
	setFilterIfPresent(query, "filter[has-revenue-date]", opts.HasRevenueDate)

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-revenue-item-actual-exports", query)
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

	rows := buildProjectPhaseRevenueItemActualExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectPhaseRevenueItemActualExportsTable(cmd, rows)
}

func parseProjectPhaseRevenueItemActualExportsListOptions(cmd *cobra.Command) (projectPhaseRevenueItemActualExportsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organizationFormatter, _ := cmd.Flags().GetString("organization-formatter")
	status, _ := cmd.Flags().GetString("status")
	broker, _ := cmd.Flags().GetString("broker")
	project, _ := cmd.Flags().GetString("project")
	createdBy, _ := cmd.Flags().GetString("created-by")
	organization, _ := cmd.Flags().GetString("organization")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	notOrganizationType, _ := cmd.Flags().GetString("not-organization-type")
	projectPhaseRevenueItems, _ := cmd.Flags().GetString("project-phase-revenue-items")
	revenueDate, _ := cmd.Flags().GetString("revenue-date")
	revenueDateMin, _ := cmd.Flags().GetString("revenue-date-min")
	revenueDateMax, _ := cmd.Flags().GetString("revenue-date-max")
	hasRevenueDate, _ := cmd.Flags().GetString("has-revenue-date")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseRevenueItemActualExportsListOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		NoAuth:                   noAuth,
		Limit:                    limit,
		Offset:                   offset,
		Sort:                     sort,
		OrganizationFormatter:    organizationFormatter,
		Status:                   status,
		Broker:                   broker,
		Project:                  project,
		CreatedBy:                createdBy,
		Organization:             organization,
		OrganizationID:           organizationID,
		OrganizationType:         organizationType,
		NotOrganizationType:      notOrganizationType,
		ProjectPhaseRevenueItems: projectPhaseRevenueItems,
		RevenueDate:              revenueDate,
		RevenueDateMin:           revenueDateMin,
		RevenueDateMax:           revenueDateMax,
		HasRevenueDate:           hasRevenueDate,
	}, nil
}

func buildProjectPhaseRevenueItemActualExportRows(resp jsonAPIResponse) []projectPhaseRevenueItemActualExportRow {
	rows := make([]projectPhaseRevenueItemActualExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectPhaseRevenueItemActualExportRow(resource))
	}
	return rows
}

func buildProjectPhaseRevenueItemActualExportRow(resource jsonAPIResource) projectPhaseRevenueItemActualExportRow {
	attrs := resource.Attributes
	row := projectPhaseRevenueItemActualExportRow{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		FileName:                stringAttr(attrs, "file-name"),
		RevenueDate:             stringAttr(attrs, "revenue-date"),
		BrokerID:                relationshipIDFromMap(resource.Relationships, "broker"),
		OrganizationFormatterID: relationshipIDFromMap(resource.Relationships, "organization-formatter"),
		CreatedByID:             relationshipIDFromMap(resource.Relationships, "created-by"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}

func renderProjectPhaseRevenueItemActualExportsTable(cmd *cobra.Command, rows []projectPhaseRevenueItemActualExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project phase revenue item actual exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tREVENUE DATE\tFILE NAME\tORG TYPE\tORG ID\tBROKER\tFORMATTER\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 12),
			truncateString(row.RevenueDate, 12),
			truncateString(row.FileName, 30),
			truncateString(row.OrganizationType, 12),
			truncateString(row.OrganizationID, 14),
			truncateString(row.BrokerID, 14),
			truncateString(row.OrganizationFormatterID, 14),
			truncateString(row.CreatedByID, 14),
		)
	}

	return writer.Flush()
}
