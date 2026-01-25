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

type projectsListOptions struct {
	BaseURL                      string
	Token                        string
	JSON                         bool
	NoAuth                       bool
	Limit                        int
	Offset                       int
	Name                         string
	Status                       string
	CreatedAtMin                 string
	CreatedAtMax                 string
	Broker                       string
	Customer                     string
	ProjectManager               string
	Estimator                    string
	Developer                    string
	ProjectOffice                string
	Q                            string
	Number                       string
	IsActive                     string
	IsManaged                    string
	JobStartOn                   string
	JobStartOnMin                string
	JobStartOnMax                string
	DueOn                        string
	DueOnMin                     string
	DueOnMax                     string
	NameLike                     string
	HasMaterialTransactionOrders string
	IsProjectManager             string
	ProjectTransportPlan         string
	IsTransportOnly              string
	JobProductionPlanPlanner     string
}

func newProjectsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		Long: `List projects with filtering and pagination.

Returns a list of projects matching the specified criteria.

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Date Filters:
  Use ISO 8601 format (YYYY-MM-DD) for date filters.`,
		Example: `  # List projects
  xbe view projects list

  # Search by project name
  xbe view projects list --name "Highway"

  # Filter by status
  xbe view projects list --status active

  # Filter by date range
  xbe view projects list --created-at-min 2024-01-01 --created-at-max 2024-12-31

  # Paginate results
  xbe view projects list --limit 20 --offset 40

  # Output as JSON
  xbe view projects list --json`,
		RunE: runProjectsList,
	}
	initProjectsListFlags(cmd)
	return cmd
}

func init() {
	projectsCmd.AddCommand(newProjectsListCmd())
}

func initProjectsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("name", "", "Filter by project name (partial match)")
	cmd.Flags().String("status", "", "Filter by project status")
	cmd.Flags().String("created-at-min", "", "Filter by minimum created date (YYYY-MM-DD)")
	cmd.Flags().String("created-at-max", "", "Filter by maximum created date (YYYY-MM-DD)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("customer", "", "Filter by customer ID (comma-separated for multiple)")
	cmd.Flags().String("project-manager", "", "Filter by project manager user ID (comma-separated for multiple)")
	cmd.Flags().String("estimator", "", "Filter by estimator user ID (comma-separated for multiple)")
	cmd.Flags().String("developer", "", "Filter by developer ID (comma-separated for multiple)")
	cmd.Flags().String("project-office", "", "Filter by project office ID (comma-separated for multiple)")
	cmd.Flags().String("q", "", "Full-text search")
	cmd.Flags().String("number", "", "Filter by project number")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("is-managed", "", "Filter by managed status (true/false)")
	cmd.Flags().String("job-start-on", "", "Filter by job start date (YYYY-MM-DD)")
	cmd.Flags().String("job-start-on-min", "", "Filter by minimum job start date (YYYY-MM-DD)")
	cmd.Flags().String("job-start-on-max", "", "Filter by maximum job start date (YYYY-MM-DD)")
	cmd.Flags().String("due-on", "", "Filter by due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-min", "", "Filter by minimum due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-max", "", "Filter by maximum due date (YYYY-MM-DD)")
	cmd.Flags().String("name-like", "", "Filter by name (partial match)")
	cmd.Flags().String("has-material-transaction-orders", "", "Filter by having material transaction orders (true/false)")
	cmd.Flags().String("is-project-manager", "", "Filter by having project manager (true/false)")
	cmd.Flags().String("project-transport-plan", "", "Filter by project transport plan ID (comma-separated for multiple)")
	cmd.Flags().String("is-transport-only", "", "Filter by transport only status (true/false)")
	cmd.Flags().String("job-production-plan-planner", "", "Filter by job production plan planner ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectsListOptions(cmd)
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
	query.Set("fields[projects]", "name,status,created-at")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[name]", opts.Name)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[customer]", opts.Customer)
	setFilterIfPresent(query, "filter[project-manager]", opts.ProjectManager)
	setFilterIfPresent(query, "filter[estimator]", opts.Estimator)
	setFilterIfPresent(query, "filter[developer]", opts.Developer)
	setFilterIfPresent(query, "filter[project-office]", opts.ProjectOffice)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[number]", opts.Number)
	setFilterIfPresent(query, "filter[is-active]", opts.IsActive)
	setFilterIfPresent(query, "filter[is-managed]", opts.IsManaged)
	setFilterIfPresent(query, "filter[job-start-on]", opts.JobStartOn)
	setFilterIfPresent(query, "filter[job-start-on-min]", opts.JobStartOnMin)
	setFilterIfPresent(query, "filter[job-start-on-max]", opts.JobStartOnMax)
	setFilterIfPresent(query, "filter[due-on]", opts.DueOn)
	setFilterIfPresent(query, "filter[due-on-min]", opts.DueOnMin)
	setFilterIfPresent(query, "filter[due-on-max]", opts.DueOnMax)
	setFilterIfPresent(query, "filter[name-like]", opts.NameLike)
	setFilterIfPresent(query, "filter[has-material-transaction-orders]", opts.HasMaterialTransactionOrders)
	setFilterIfPresent(query, "filter[is-project-manager]", opts.IsProjectManager)
	setFilterIfPresent(query, "filter[project-transport-plan]", opts.ProjectTransportPlan)
	setFilterIfPresent(query, "filter[is-transport-only]", opts.IsTransportOnly)
	setFilterIfPresent(query, "filter[job-production-plan-planner]", opts.JobProductionPlanPlanner)

	body, _, err := client.Get(cmd.Context(), "/v1/projects", query)
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

	if opts.JSON {
		rows := buildProjectRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectsList(cmd, resp)
}

func parseProjectsListOptions(cmd *cobra.Command) (projectsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return projectsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return projectsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return projectsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return projectsListOptions{}, err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return projectsListOptions{}, err
	}
	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return projectsListOptions{}, err
	}
	createdAtMin, err := cmd.Flags().GetString("created-at-min")
	if err != nil {
		return projectsListOptions{}, err
	}
	createdAtMax, err := cmd.Flags().GetString("created-at-max")
	if err != nil {
		return projectsListOptions{}, err
	}
	broker, err := cmd.Flags().GetString("broker")
	if err != nil {
		return projectsListOptions{}, err
	}
	customer, err := cmd.Flags().GetString("customer")
	if err != nil {
		return projectsListOptions{}, err
	}
	projectManager, err := cmd.Flags().GetString("project-manager")
	if err != nil {
		return projectsListOptions{}, err
	}
	estimator, err := cmd.Flags().GetString("estimator")
	if err != nil {
		return projectsListOptions{}, err
	}
	developer, err := cmd.Flags().GetString("developer")
	if err != nil {
		return projectsListOptions{}, err
	}
	projectOffice, err := cmd.Flags().GetString("project-office")
	if err != nil {
		return projectsListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return projectsListOptions{}, err
	}
	number, err := cmd.Flags().GetString("number")
	if err != nil {
		return projectsListOptions{}, err
	}
	isActive, err := cmd.Flags().GetString("is-active")
	if err != nil {
		return projectsListOptions{}, err
	}
	isManaged, err := cmd.Flags().GetString("is-managed")
	if err != nil {
		return projectsListOptions{}, err
	}
	jobStartOn, err := cmd.Flags().GetString("job-start-on")
	if err != nil {
		return projectsListOptions{}, err
	}
	jobStartOnMin, err := cmd.Flags().GetString("job-start-on-min")
	if err != nil {
		return projectsListOptions{}, err
	}
	jobStartOnMax, err := cmd.Flags().GetString("job-start-on-max")
	if err != nil {
		return projectsListOptions{}, err
	}
	dueOn, err := cmd.Flags().GetString("due-on")
	if err != nil {
		return projectsListOptions{}, err
	}
	dueOnMin, err := cmd.Flags().GetString("due-on-min")
	if err != nil {
		return projectsListOptions{}, err
	}
	dueOnMax, err := cmd.Flags().GetString("due-on-max")
	if err != nil {
		return projectsListOptions{}, err
	}
	nameLike, err := cmd.Flags().GetString("name-like")
	if err != nil {
		return projectsListOptions{}, err
	}
	hasMaterialTransactionOrders, err := cmd.Flags().GetString("has-material-transaction-orders")
	if err != nil {
		return projectsListOptions{}, err
	}
	isProjectManager, err := cmd.Flags().GetString("is-project-manager")
	if err != nil {
		return projectsListOptions{}, err
	}
	projectTransportPlan, err := cmd.Flags().GetString("project-transport-plan")
	if err != nil {
		return projectsListOptions{}, err
	}
	isTransportOnly, err := cmd.Flags().GetString("is-transport-only")
	if err != nil {
		return projectsListOptions{}, err
	}
	jobProductionPlanPlanner, err := cmd.Flags().GetString("job-production-plan-planner")
	if err != nil {
		return projectsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return projectsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return projectsListOptions{}, err
	}

	return projectsListOptions{
		BaseURL:                      baseURL,
		Token:                        token,
		JSON:                         jsonOut,
		NoAuth:                       noAuth,
		Limit:                        limit,
		Offset:                       offset,
		Name:                         name,
		Status:                       status,
		CreatedAtMin:                 createdAtMin,
		CreatedAtMax:                 createdAtMax,
		Broker:                       broker,
		Customer:                     customer,
		ProjectManager:               projectManager,
		Estimator:                    estimator,
		Developer:                    developer,
		ProjectOffice:                projectOffice,
		Q:                            q,
		Number:                       number,
		IsActive:                     isActive,
		IsManaged:                    isManaged,
		JobStartOn:                   jobStartOn,
		JobStartOnMin:                jobStartOnMin,
		JobStartOnMax:                jobStartOnMax,
		DueOn:                        dueOn,
		DueOnMin:                     dueOnMin,
		DueOnMax:                     dueOnMax,
		NameLike:                     nameLike,
		HasMaterialTransactionOrders: hasMaterialTransactionOrders,
		IsProjectManager:             isProjectManager,
		ProjectTransportPlan:         projectTransportPlan,
		IsTransportOnly:              isTransportOnly,
		JobProductionPlanPlanner:     jobProductionPlanPlanner,
	}, nil
}

type projectRow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func buildProjectRows(resp jsonAPIResponse) []projectRow {
	rows := make([]projectRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, projectRow{
			ID:        resource.ID,
			Name:      strings.TrimSpace(stringAttr(resource.Attributes, "name")),
			Status:    strings.TrimSpace(stringAttr(resource.Attributes, "status")),
			CreatedAt: strings.TrimSpace(stringAttr(resource.Attributes, "created-at")),
		})
	}
	return rows
}

func renderProjectsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildProjectRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No projects found.")
		return nil
	}

	const nameMax = 50
	const statusMax = 20

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tSTATUS\tCREATED")
	for _, row := range rows {
		createdAt := row.CreatedAt
		if len(createdAt) > 10 {
			createdAt = createdAt[:10]
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, nameMax),
			truncateString(row.Status, statusMax),
			createdAt,
		)
	}
	return writer.Flush()
}
