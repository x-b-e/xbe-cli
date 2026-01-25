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

type projectPhaseCostItemActualsListOptions struct {
	BaseURL                                    string
	Token                                      string
	JSON                                       bool
	NoAuth                                     bool
	Limit                                      int
	Offset                                     int
	Sort                                       string
	ProjectPhaseCostItem                       string
	ProjectPhaseRevenueItemActual              string
	JobProductionPlanProjectPhaseRevenueItem   string
	JobProductionPlanProjectPhaseRevenueItemID string
	JobProductionPlan                          string
	JobProductionPlanID                        string
	CreatedBy                                  string
	Project                                    string
	Quantity                                   string
}

type projectPhaseCostItemActualRow struct {
	ID                                         string `json:"id"`
	ProjectPhaseCostItemID                     string `json:"project_phase_cost_item_id,omitempty"`
	ProjectPhaseRevenueItemActualID            string `json:"project_phase_revenue_item_actual_id,omitempty"`
	JobProductionPlanProjectPhaseRevenueItemID string `json:"job_production_plan_project_phase_revenue_item_id,omitempty"`
	JobProductionPlanID                        string `json:"job_production_plan_id,omitempty"`
	Quantity                                   string `json:"quantity,omitempty"`
	PricePerUnitExplicit                       string `json:"price_per_unit_explicit,omitempty"`
	PricePerUnit                               string `json:"price_per_unit,omitempty"`
	CostAmount                                 string `json:"cost_amount,omitempty"`
}

func newProjectPhaseCostItemActualsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project phase cost item actuals",
		Long: `List project phase cost item actuals with filtering and pagination.

Output Columns:
  ID            Cost item actual identifier
  COST ITEM     Project phase cost item ID
  REV ACTUAL    Project phase revenue item actual ID
  JOB PLAN      Job production plan ID
  QTY           Quantity
  PRICE/UNIT    Price per unit (resolved)
  COST AMOUNT   Total cost amount

Filters:
  --project-phase-cost-item                     Filter by project phase cost item ID
  --project-phase-revenue-item-actual           Filter by project phase revenue item actual ID
  --job-production-plan-project-phase-revenue-item      Filter by job production plan project phase revenue item ID
  --job-production-plan-project-phase-revenue-item-id   Filter by job production plan project phase revenue item ID (direct)
  --job-production-plan                          Filter by job production plan ID
  --job-production-plan-id                       Filter by job production plan ID (direct)
  --created-by                                   Filter by created-by user ID
  --project                                      Filter by project ID
  --quantity                                     Filter by quantity

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project phase cost item actuals
  xbe view project-phase-cost-item-actuals list

  # Filter by project phase cost item
  xbe view project-phase-cost-item-actuals list --project-phase-cost-item 123

  # Filter by job production plan
  xbe view project-phase-cost-item-actuals list --job-production-plan 456

  # Output as JSON
  xbe view project-phase-cost-item-actuals list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectPhaseCostItemActualsList,
	}
	initProjectPhaseCostItemActualsListFlags(cmd)
	return cmd
}

func init() {
	projectPhaseCostItemActualsCmd.AddCommand(newProjectPhaseCostItemActualsListCmd())
}

func initProjectPhaseCostItemActualsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-phase-cost-item", "", "Filter by project phase cost item ID")
	cmd.Flags().String("project-phase-revenue-item-actual", "", "Filter by project phase revenue item actual ID")
	cmd.Flags().String("job-production-plan-project-phase-revenue-item", "", "Filter by job production plan project phase revenue item ID")
	cmd.Flags().String("job-production-plan-project-phase-revenue-item-id", "", "Filter by job production plan project phase revenue item ID (direct)")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("job-production-plan-id", "", "Filter by job production plan ID (direct)")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("quantity", "", "Filter by quantity")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseCostItemActualsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectPhaseCostItemActualsListOptions(cmd)
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
	query.Set("fields[project-phase-cost-item-actuals]", "quantity,price-per-unit-explicit,price-per-unit,cost-amount,project-phase-cost-item,project-phase-revenue-item-actual,job-production-plan,job-production-plan-project-phase-revenue-item")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-phase-cost-item]", opts.ProjectPhaseCostItem)
	setFilterIfPresent(query, "filter[project-phase-revenue-item-actual]", opts.ProjectPhaseRevenueItemActual)
	setFilterIfPresent(query, "filter[job-production-plan-project-phase-revenue-item]", opts.JobProductionPlanProjectPhaseRevenueItem)
	setFilterIfPresent(query, "filter[job-production-plan-project-phase-revenue-item-id]", opts.JobProductionPlanProjectPhaseRevenueItemID)
	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[job-production-plan-id]", opts.JobProductionPlanID)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[quantity]", opts.Quantity)

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-cost-item-actuals", query)
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

	rows := buildProjectPhaseCostItemActualRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectPhaseCostItemActualsTable(cmd, rows)
}

func parseProjectPhaseCostItemActualsListOptions(cmd *cobra.Command) (projectPhaseCostItemActualsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectPhaseCostItem, _ := cmd.Flags().GetString("project-phase-cost-item")
	projectPhaseRevenueItemActual, _ := cmd.Flags().GetString("project-phase-revenue-item-actual")
	jobProductionPlanProjectPhaseRevenueItem, _ := cmd.Flags().GetString("job-production-plan-project-phase-revenue-item")
	jobProductionPlanProjectPhaseRevenueItemID, _ := cmd.Flags().GetString("job-production-plan-project-phase-revenue-item-id")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan-id")
	createdBy, _ := cmd.Flags().GetString("created-by")
	project, _ := cmd.Flags().GetString("project")
	quantity, _ := cmd.Flags().GetString("quantity")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseCostItemActualsListOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
		NoAuth:                                   noAuth,
		Limit:                                    limit,
		Offset:                                   offset,
		Sort:                                     sort,
		ProjectPhaseCostItem:                     projectPhaseCostItem,
		ProjectPhaseRevenueItemActual:            projectPhaseRevenueItemActual,
		JobProductionPlanProjectPhaseRevenueItem: jobProductionPlanProjectPhaseRevenueItem,
		JobProductionPlanProjectPhaseRevenueItemID: jobProductionPlanProjectPhaseRevenueItemID,
		JobProductionPlan:                          jobProductionPlan,
		JobProductionPlanID:                        jobProductionPlanID,
		CreatedBy:                                  createdBy,
		Project:                                    project,
		Quantity:                                   quantity,
	}, nil
}

func buildProjectPhaseCostItemActualRows(resp jsonAPIResponse) []projectPhaseCostItemActualRow {
	rows := make([]projectPhaseCostItemActualRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectPhaseCostItemActualRow{
			ID:                   resource.ID,
			Quantity:             stringAttr(resource.Attributes, "quantity"),
			PricePerUnitExplicit: stringAttr(resource.Attributes, "price-per-unit-explicit"),
			PricePerUnit:         stringAttr(resource.Attributes, "price-per-unit"),
			CostAmount:           stringAttr(resource.Attributes, "cost-amount"),
		}

		if rel, ok := resource.Relationships["project-phase-cost-item"]; ok && rel.Data != nil {
			row.ProjectPhaseCostItemID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["project-phase-revenue-item-actual"]; ok && rel.Data != nil {
			row.ProjectPhaseRevenueItemActualID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-production-plan-project-phase-revenue-item"]; ok && rel.Data != nil {
			row.JobProductionPlanProjectPhaseRevenueItemID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProjectPhaseCostItemActualsTable(cmd *cobra.Command, rows []projectPhaseCostItemActualRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project phase cost item actuals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCOST ITEM\tREV ACTUAL\tJOB PLAN\tQTY\tPRICE/UNIT\tCOST AMOUNT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectPhaseCostItemID,
			row.ProjectPhaseRevenueItemActualID,
			row.JobProductionPlanID,
			row.Quantity,
			row.PricePerUnit,
			row.CostAmount,
		)
	}
	return writer.Flush()
}
