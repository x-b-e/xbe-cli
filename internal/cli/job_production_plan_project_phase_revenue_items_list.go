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

type jobProductionPlanProjectPhaseRevenueItemsListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	JobProductionPlan       string
	ProjectPhaseRevenueItem string
	Project                 string
	ProjectRevenueItem      string
	ProjectPhase            string
	JobProductionPlanStatus string
}

type jobProductionPlanProjectPhaseRevenueItemRow struct {
	ID                        string `json:"id"`
	JobProductionPlanID       string `json:"job_production_plan_id,omitempty"`
	ProjectPhaseRevenueItemID string `json:"project_phase_revenue_item_id,omitempty"`
	Quantity                  string `json:"quantity,omitempty"`
}

func newJobProductionPlanProjectPhaseRevenueItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan project phase revenue items",
		Long: `List job production plan project phase revenue items.

Output Columns:
  ID                       Item identifier
  JOB_PRODUCTION_PLAN      Job production plan ID
  PROJECT_PHASE_REVENUE_ITEM Project phase revenue item ID
  QUANTITY                 Planned quantity

Filters:
  --job-production-plan       Filter by job production plan ID
  --project-phase-revenue-item Filter by project phase revenue item ID
  --project                   Filter by project ID
  --project-revenue-item       Filter by project revenue item ID
  --project-phase              Filter by project phase ID
  --job-production-plan-status Filter by job production plan status

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List all items
  xbe view job-production-plan-project-phase-revenue-items list

  # Filter by job production plan
  xbe view job-production-plan-project-phase-revenue-items list --job-production-plan 123

  # Filter by project phase revenue item
  xbe view job-production-plan-project-phase-revenue-items list --project-phase-revenue-item 456

  # Output as JSON
  xbe view job-production-plan-project-phase-revenue-items list --json`,
		RunE: runJobProductionPlanProjectPhaseRevenueItemsList,
	}
	initJobProductionPlanProjectPhaseRevenueItemsListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanProjectPhaseRevenueItemsCmd.AddCommand(newJobProductionPlanProjectPhaseRevenueItemsListCmd())
}

func initJobProductionPlanProjectPhaseRevenueItemsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("project-phase-revenue-item", "", "Filter by project phase revenue item ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("project-revenue-item", "", "Filter by project revenue item ID")
	cmd.Flags().String("project-phase", "", "Filter by project phase ID")
	cmd.Flags().String("job-production-plan-status", "", "Filter by job production plan status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanProjectPhaseRevenueItemsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanProjectPhaseRevenueItemsListOptions(cmd)
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
	query.Set("fields[job-production-plan-project-phase-revenue-items]", "quantity,job-production-plan,project-phase-revenue-item")
	query.Set("include", "job-production-plan,project-phase-revenue-item")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[project-phase-revenue-item]", opts.ProjectPhaseRevenueItem)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[project-revenue-item]", opts.ProjectRevenueItem)
	setFilterIfPresent(query, "filter[project-phase]", opts.ProjectPhase)
	setFilterIfPresent(query, "filter[job-production-plan-status]", opts.JobProductionPlanStatus)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-project-phase-revenue-items", query)
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

	rows := buildJobProductionPlanProjectPhaseRevenueItemRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanProjectPhaseRevenueItemsTable(cmd, rows)
}

func parseJobProductionPlanProjectPhaseRevenueItemsListOptions(cmd *cobra.Command) (jobProductionPlanProjectPhaseRevenueItemsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	projectPhaseRevenueItem, _ := cmd.Flags().GetString("project-phase-revenue-item")
	project, _ := cmd.Flags().GetString("project")
	projectRevenueItem, _ := cmd.Flags().GetString("project-revenue-item")
	projectPhase, _ := cmd.Flags().GetString("project-phase")
	jobProductionPlanStatus, _ := cmd.Flags().GetString("job-production-plan-status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanProjectPhaseRevenueItemsListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		JobProductionPlan:       jobProductionPlan,
		ProjectPhaseRevenueItem: projectPhaseRevenueItem,
		Project:                 project,
		ProjectRevenueItem:      projectRevenueItem,
		ProjectPhase:            projectPhase,
		JobProductionPlanStatus: jobProductionPlanStatus,
	}, nil
}

func buildJobProductionPlanProjectPhaseRevenueItemRows(resp jsonAPIResponse) []jobProductionPlanProjectPhaseRevenueItemRow {
	rows := make([]jobProductionPlanProjectPhaseRevenueItemRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := jobProductionPlanProjectPhaseRevenueItemRow{
			ID:       resource.ID,
			Quantity: stringAttr(resource.Attributes, "quantity"),
		}

		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-phase-revenue-item"]; ok && rel.Data != nil {
			row.ProjectPhaseRevenueItemID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderJobProductionPlanProjectPhaseRevenueItemsTable(cmd *cobra.Command, rows []jobProductionPlanProjectPhaseRevenueItemRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan project phase revenue items found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB_PRODUCTION_PLAN\tPROJECT_PHASE_REVENUE_ITEM\tQUANTITY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobProductionPlanID,
			row.ProjectPhaseRevenueItemID,
			row.Quantity,
		)
	}
	return writer.Flush()
}
