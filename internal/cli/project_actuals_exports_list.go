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

type projectActualsExportsListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	OrganizationFormatter string
	Status                string
	Broker                string
	Project               string
	CreatedBy             string
	Organization          string
	OrganizationID        string
	OrganizationType      string
	NotOrganizationType   string
	JobProductionPlans    string
}

type projectActualsExportRow struct {
	ID                      string `json:"id"`
	Status                  string `json:"status,omitempty"`
	FileName                string `json:"file_name,omitempty"`
	OrganizationType        string `json:"organization_type,omitempty"`
	OrganizationID          string `json:"organization_id,omitempty"`
	BrokerID                string `json:"broker_id,omitempty"`
	ProjectID               string `json:"project_id,omitempty"`
	OrganizationFormatterID string `json:"organization_formatter_id,omitempty"`
	CreatedByID             string `json:"created_by_id,omitempty"`
}

func newProjectActualsExportsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project actuals exports",
		Long: `List project actuals exports with filtering and pagination.

Output Columns:
  ID         Export identifier
  STATUS     Processing status
  FILE NAME  Generated file name
  ORG TYPE   Organization type
  ORG ID     Organization ID
  BROKER     Broker ID
  PROJECT    Project ID
  FORMATTER  Organization formatter ID
  CREATED BY Creator user ID

Filters:
  --organization-formatter  Filter by organization formatter ID
  --status                  Filter by status (processing, processed, failed)
  --broker                  Filter by broker ID
  --project                 Filter by project ID
  --created-by              Filter by created-by user ID
  --organization            Filter by organization (Type|ID, e.g. Broker|123)
  --organization-id         Filter by organization ID (requires --organization-type)
  --organization-type       Filter by organization type (e.g. Broker, Customer)
  --not-organization-type   Exclude organization type (e.g. Broker)
  --job-production-plans    Filter by job production plan IDs (comma-separated)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List exports
  xbe view project-actuals-exports list

  # Filter by status
  xbe view project-actuals-exports list --status processed

  # Filter by organization
  xbe view project-actuals-exports list --organization "Broker|123"

  # Output as JSON
  xbe view project-actuals-exports list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectActualsExportsList,
	}
	initProjectActualsExportsListFlags(cmd)
	return cmd
}

func init() {
	projectActualsExportsCmd.AddCommand(newProjectActualsExportsListCmd())
}

func initProjectActualsExportsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
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
	cmd.Flags().String("job-production-plans", "", "Filter by job production plan IDs (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectActualsExportsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectActualsExportsListOptions(cmd)
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
	query.Set("fields[project-actuals-exports]", "status,file-name,organization,broker,project,organization-formatter,created-by")

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
	setFilterIfPresent(query, "filter[job_production_plans]", opts.JobProductionPlans)

	body, _, err := client.Get(cmd.Context(), "/v1/project-actuals-exports", query)
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

	rows := buildProjectActualsExportRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectActualsExportsTable(cmd, rows)
}

func parseProjectActualsExportsListOptions(cmd *cobra.Command) (projectActualsExportsListOptions, error) {
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
	jobProductionPlans, _ := cmd.Flags().GetString("job-production-plans")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectActualsExportsListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		OrganizationFormatter: organizationFormatter,
		Status:                status,
		Broker:                broker,
		Project:               project,
		CreatedBy:             createdBy,
		Organization:          organization,
		OrganizationID:        organizationID,
		OrganizationType:      organizationType,
		NotOrganizationType:   notOrganizationType,
		JobProductionPlans:    jobProductionPlans,
	}, nil
}

func buildProjectActualsExportRows(resp jsonAPIResponse) []projectActualsExportRow {
	rows := make([]projectActualsExportRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildProjectActualsExportRow(resource))
	}
	return rows
}

func buildProjectActualsExportRow(resource jsonAPIResource) projectActualsExportRow {
	attrs := resource.Attributes
	row := projectActualsExportRow{
		ID:                      resource.ID,
		Status:                  stringAttr(attrs, "status"),
		FileName:                stringAttr(attrs, "file-name"),
		BrokerID:                relationshipIDFromMap(resource.Relationships, "broker"),
		ProjectID:               relationshipIDFromMap(resource.Relationships, "project"),
		OrganizationFormatterID: relationshipIDFromMap(resource.Relationships, "organization-formatter"),
		CreatedByID:             relationshipIDFromMap(resource.Relationships, "created-by"),
	}

	if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
		row.OrganizationType = rel.Data.Type
		row.OrganizationID = rel.Data.ID
	}

	return row
}

func renderProjectActualsExportsTable(cmd *cobra.Command, rows []projectActualsExportRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project actuals exports found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tFILE NAME\tORG TYPE\tORG ID\tBROKER\tPROJECT\tFORMATTER\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Status, 12),
			truncateString(row.FileName, 30),
			truncateString(row.OrganizationType, 12),
			truncateString(row.OrganizationID, 14),
			truncateString(row.BrokerID, 14),
			truncateString(row.ProjectID, 14),
			truncateString(row.OrganizationFormatterID, 14),
			truncateString(row.CreatedByID, 14),
		)
	}

	return writer.Flush()
}
