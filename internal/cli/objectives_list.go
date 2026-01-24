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

type objectivesListOptions struct {
	BaseURL                                 string
	Token                                   string
	JSON                                    bool
	NoAuth                                  bool
	Limit                                   int
	Offset                                  int
	Sort                                    string
	Name                                    string
	Owner                                   string
	Organization                            string
	OrganizationType                        string
	OrganizationID                          string
	Status                                  string
	StartOn                                 string
	StartOnMin                              string
	StartOnMax                              string
	EndOn                                   string
	EndOnMin                                string
	EndOnMax                                string
	Commitment                              string
	Project                                 string
	IsTemplate                              string
	TemplateScope                           string
	Slug                                    string
	SalesResponsiblePerson                  string
	HasSalesResponsiblePerson               string
	WithoutCustomerSuccessResponsiblePerson string
}

type objectiveRow struct {
	ID                         string `json:"id"`
	Name                       string `json:"name,omitempty"`
	Status                     string `json:"status,omitempty"`
	Commitment                 string `json:"commitment,omitempty"`
	StartOn                    string `json:"start_on,omitempty"`
	EndOn                      string `json:"end_on,omitempty"`
	IsTemplate                 any    `json:"is_template,omitempty"`
	TemplateScope              string `json:"template_scope,omitempty"`
	Slug                       string `json:"slug,omitempty"`
	OwnerID                    string `json:"owner_id,omitempty"`
	OwnerName                  string `json:"owner_name,omitempty"`
	OrganizationID             string `json:"organization_id,omitempty"`
	OrganizationType           string `json:"organization_type,omitempty"`
	OrganizationName           string `json:"organization_name,omitempty"`
	ProjectID                  string `json:"project_id,omitempty"`
	ProjectName                string `json:"project_name,omitempty"`
	SalesResponsiblePersonID   string `json:"sales_responsible_person_id,omitempty"`
	SalesResponsiblePersonName string `json:"sales_responsible_person_name,omitempty"`
}

func newObjectivesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List objectives",
		Long: `List objectives with filtering and pagination.

Output Columns:
  ID           Objective identifier
  STATUS       Current status (abandoned excluded by default)
  NAME         Objective name
  COMMITMENT   Commitment level
  START        Start date
  END          End date
  OWNER        Owner name or ID
  ORG          Organization name or Type/ID
  TEMPLATE     Template flag

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  --name                                     Filter by objective name (partial match)
  --owner                                    Filter by owner user ID
  --organization                             Filter by organization (Type|ID)
  --organization-id                          Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type                        Filter by organization type
  --status                                   Filter by status (unknown, not_started, red, yellow, green, completed, abandoned)
  --start-on, --start-on-min, --start-on-max Filter by start date (YYYY-MM-DD)
  --end-on, --end-on-min, --end-on-max       Filter by end date (YYYY-MM-DD)
  --commitment                               Filter by commitment (committed, aspirational)
  --project                                  Filter by project ID
  --is-template                              Filter by template flag (true/false)
  --template-scope                           Filter by template scope (match_all, organization, project)
  --slug                                     Filter by slug
  --sales-responsible-person                 Filter by sales responsible person user ID
  --has-sales-responsible-person             Filter by whether a sales responsible person is set (true/false)
  --without-customer-success-responsible-person  Filter by objectives without customer success responsible person (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List objectives
  xbe view objectives list

  # Filter by owner
  xbe view objectives list --owner 123

  # Filter by organization
  xbe view objectives list --organization "Broker|456"

  # Filter by status
  xbe view objectives list --status green

  # Filter by template scope
  xbe view objectives list --is-template true --template-scope match_all

  # Output as JSON
  xbe view objectives list --json`,
		Args: cobra.NoArgs,
		RunE: runObjectivesList,
	}
	initObjectivesListFlags(cmd)
	return cmd
}

func init() {
	objectivesCmd.AddCommand(newObjectivesListCmd())
}

func initObjectivesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("name", "", "Filter by objective name (partial match)")
	cmd.Flags().String("owner", "", "Filter by owner user ID (comma-separated for multiple)")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("start-on", "", "Filter by start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-min", "", "Filter by minimum start date (YYYY-MM-DD)")
	cmd.Flags().String("start-on-max", "", "Filter by maximum start date (YYYY-MM-DD)")
	cmd.Flags().String("end-on", "", "Filter by end date (YYYY-MM-DD)")
	cmd.Flags().String("end-on-min", "", "Filter by minimum end date (YYYY-MM-DD)")
	cmd.Flags().String("end-on-max", "", "Filter by maximum end date (YYYY-MM-DD)")
	cmd.Flags().String("commitment", "", "Filter by commitment (committed, aspirational)")
	cmd.Flags().String("project", "", "Filter by project ID (comma-separated for multiple)")
	cmd.Flags().String("is-template", "", "Filter by template flag (true/false)")
	cmd.Flags().String("template-scope", "", "Filter by template scope")
	cmd.Flags().String("slug", "", "Filter by slug")
	cmd.Flags().String("sales-responsible-person", "", "Filter by sales responsible person user ID (comma-separated for multiple)")
	cmd.Flags().String("has-sales-responsible-person", "", "Filter by whether sales responsible person is set (true/false)")
	cmd.Flags().String("without-customer-success-responsible-person", "", "Filter by objectives without customer success responsible person (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runObjectivesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseObjectivesListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[objectives]", "name,status,start-on,end-on,commitment,is-template,template-scope,slug,owner,organization,project,sales-responsible-person")
	query.Set("include", "owner,organization,project,sales-responsible-person")
	query.Set("fields[users]", "name")
	query.Set("fields[projects]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[owner]", opts.Owner)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	organizationIDFilter, err := buildOrganizationIDFilter(opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if organizationIDFilter != "" {
		query.Set("filter[organization_id]", organizationIDFilter)
	} else {
		setFilterIfPresent(query, "filter[organization_type]", opts.OrganizationType)
	}
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[start-on]", opts.StartOn)
	setFilterIfPresent(query, "filter[start-on-min]", opts.StartOnMin)
	setFilterIfPresent(query, "filter[start-on-max]", opts.StartOnMax)
	setFilterIfPresent(query, "filter[end-on]", opts.EndOn)
	setFilterIfPresent(query, "filter[end-on-min]", opts.EndOnMin)
	setFilterIfPresent(query, "filter[end-on-max]", opts.EndOnMax)
	setFilterIfPresent(query, "filter[commitment]", opts.Commitment)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[is-template]", opts.IsTemplate)
	setFilterIfPresent(query, "filter[template-scope]", opts.TemplateScope)
	setFilterIfPresent(query, "filter[slug]", opts.Slug)
	setFilterIfPresent(query, "filter[sales-responsible-person]", opts.SalesResponsiblePerson)
	setFilterIfPresent(query, "filter[has-sales-responsible-person]", opts.HasSalesResponsiblePerson)
	setFilterIfPresent(query, "filter[without-customer-success-responsible-person]", opts.WithoutCustomerSuccessResponsiblePerson)

	body, _, err := client.Get(cmd.Context(), "/v1/objectives", query)
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

	rows := buildObjectiveRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderObjectivesTable(cmd, rows)
}

func parseObjectivesListOptions(cmd *cobra.Command) (objectivesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	name, _ := cmd.Flags().GetString("name")
	owner, _ := cmd.Flags().GetString("owner")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	status, _ := cmd.Flags().GetString("status")
	startOn, _ := cmd.Flags().GetString("start-on")
	startOnMin, _ := cmd.Flags().GetString("start-on-min")
	startOnMax, _ := cmd.Flags().GetString("start-on-max")
	endOn, _ := cmd.Flags().GetString("end-on")
	endOnMin, _ := cmd.Flags().GetString("end-on-min")
	endOnMax, _ := cmd.Flags().GetString("end-on-max")
	commitment, _ := cmd.Flags().GetString("commitment")
	project, _ := cmd.Flags().GetString("project")
	isTemplate, _ := cmd.Flags().GetString("is-template")
	templateScope, _ := cmd.Flags().GetString("template-scope")
	slug, _ := cmd.Flags().GetString("slug")
	salesResponsiblePerson, _ := cmd.Flags().GetString("sales-responsible-person")
	hasSalesResponsiblePerson, _ := cmd.Flags().GetString("has-sales-responsible-person")
	withoutCustomerSuccessResponsiblePerson, _ := cmd.Flags().GetString("without-customer-success-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return objectivesListOptions{
		BaseURL:                                 baseURL,
		Token:                                   token,
		JSON:                                    jsonOut,
		NoAuth:                                  noAuth,
		Limit:                                   limit,
		Offset:                                  offset,
		Sort:                                    sort,
		Name:                                    name,
		Owner:                                   owner,
		Organization:                            organization,
		OrganizationType:                        organizationType,
		OrganizationID:                          organizationID,
		Status:                                  status,
		StartOn:                                 startOn,
		StartOnMin:                              startOnMin,
		StartOnMax:                              startOnMax,
		EndOn:                                   endOn,
		EndOnMin:                                endOnMin,
		EndOnMax:                                endOnMax,
		Commitment:                              commitment,
		Project:                                 project,
		IsTemplate:                              isTemplate,
		TemplateScope:                           templateScope,
		Slug:                                    slug,
		SalesResponsiblePerson:                  salesResponsiblePerson,
		HasSalesResponsiblePerson:               hasSalesResponsiblePerson,
		WithoutCustomerSuccessResponsiblePerson: withoutCustomerSuccessResponsiblePerson,
	}, nil
}

func buildObjectiveRows(resp jsonAPIResponse) []objectiveRow {
	included := indexIncludedResources(resp.Included)
	rows := make([]objectiveRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := objectiveRow{
			ID:            resource.ID,
			Name:          strings.TrimSpace(stringAttr(attrs, "name")),
			Status:        stringAttr(attrs, "status"),
			Commitment:    stringAttr(attrs, "commitment"),
			StartOn:       formatDate(stringAttr(attrs, "start-on")),
			EndOn:         formatDate(stringAttr(attrs, "end-on")),
			IsTemplate:    attrs["is-template"],
			TemplateScope: stringAttr(attrs, "template-scope"),
			Slug:          stringAttr(attrs, "slug"),
		}

		if rel, ok := resource.Relationships["owner"]; ok && rel.Data != nil {
			row.OwnerID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.OwnerName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationID = rel.Data.ID
			row.OrganizationType = rel.Data.Type
			if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.OrganizationName = firstNonEmpty(
					stringAttr(org.Attributes, "company-name"),
					stringAttr(org.Attributes, "name"),
				)
			}
		}

		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
			if proj, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ProjectName = strings.TrimSpace(stringAttr(proj.Attributes, "name"))
			}
		}

		if rel, ok := resource.Relationships["sales-responsible-person"]; ok && rel.Data != nil {
			row.SalesResponsiblePersonID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.SalesResponsiblePersonName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderObjectivesTable(cmd *cobra.Command, rows []objectiveRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No objectives found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tNAME\tCOMMITMENT\tSTART\tEND\tOWNER\tORG\tTEMPLATE")
	for _, row := range rows {
		owner := firstNonEmpty(row.OwnerName, row.OwnerID)
		org := row.OrganizationName
		if org == "" && row.OrganizationType != "" && row.OrganizationID != "" {
			org = fmt.Sprintf("%s/%s", row.OrganizationType, row.OrganizationID)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Name, 40),
			row.Commitment,
			row.StartOn,
			row.EndOn,
			truncateString(owner, 25),
			truncateString(org, 30),
			formatAnyValue(row.IsTemplate),
		)
	}
	return writer.Flush()
}
