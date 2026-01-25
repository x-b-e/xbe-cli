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

type projectPhaseRevenueItemActualsListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	ProjectPhaseRevenueItem string
	JobProductionPlan       string
	CreatedBy               string
	Project                 string
	RevenueDate             string
	RevenueDateMin          string
	RevenueDateMax          string
}

type projectPhaseRevenueItemActualRow struct {
	ID                                         string `json:"id"`
	ProjectPhaseRevenueItemID                  string `json:"project_phase_revenue_item_id,omitempty"`
	JobProductionPlanProjectPhaseRevenueItemID string `json:"job_production_plan_project_phase_revenue_item_id,omitempty"`
	JobProductionPlanID                        string `json:"job_production_plan_id,omitempty"`
	RevenueDate                                string `json:"revenue_date,omitempty"`
	Quantity                                   string `json:"quantity,omitempty"`
	QuantityStrategyExplicit                   string `json:"quantity_strategy_explicit,omitempty"`
	PricePerUnit                               string `json:"price_per_unit,omitempty"`
	RevenueAmount                              string `json:"revenue_amount,omitempty"`
	QuantityIndirect                           string `json:"quantity_indirect,omitempty"`
	CostAmount                                 string `json:"cost_amount,omitempty"`
}

func newProjectPhaseRevenueItemActualsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project phase revenue item actuals",
		Long: `List project phase revenue item actuals with filtering and pagination.

Output Columns:
  ID          Revenue item actual identifier
  REV ITEM    Project phase revenue item ID
  JOB PLAN    Job production plan ID
  REV DATE    Revenue date
  QTY         Quantity (resolved)
  STRATEGY    Quantity strategy (direct/indirect)
  PRICE/UNIT  Price per unit (resolved)
  REV AMOUNT  Total revenue amount

Filters:
  --project-phase-revenue-item  Filter by project phase revenue item ID
  --job-production-plan         Filter by job production plan ID
  --created-by                  Filter by created-by user ID
  --project                     Filter by project ID
  --revenue-date                Filter by revenue date (YYYY-MM-DD)
  --revenue-date-min            Filter by revenue date on/after (YYYY-MM-DD)
  --revenue-date-max            Filter by revenue date on/before (YYYY-MM-DD)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project phase revenue item actuals
  xbe view project-phase-revenue-item-actuals list

  # Filter by project phase revenue item
  xbe view project-phase-revenue-item-actuals list --project-phase-revenue-item 123

  # Filter by job production plan
  xbe view project-phase-revenue-item-actuals list --job-production-plan 456

  # Output as JSON
  xbe view project-phase-revenue-item-actuals list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectPhaseRevenueItemActualsList,
	}
	initProjectPhaseRevenueItemActualsListFlags(cmd)
	return cmd
}

func init() {
	projectPhaseRevenueItemActualsCmd.AddCommand(newProjectPhaseRevenueItemActualsListCmd())
}

func initProjectPhaseRevenueItemActualsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-phase-revenue-item", "", "Filter by project phase revenue item ID")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("revenue-date", "", "Filter by revenue date (YYYY-MM-DD)")
	cmd.Flags().String("revenue-date-min", "", "Filter by revenue date on/after (YYYY-MM-DD)")
	cmd.Flags().String("revenue-date-max", "", "Filter by revenue date on/before (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseRevenueItemActualsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectPhaseRevenueItemActualsListOptions(cmd)
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
	query.Set("fields[project-phase-revenue-item-actuals]", "quantity,revenue-date,quantity-strategy-explicit,price-per-unit,revenue-amount,project-phase-revenue-item,job-production-plan,job-production-plan-project-phase-revenue-item")
	query.Set("meta[project-phase-revenue-item-actuals]", "quantity-indirect,cost-amount")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[project-phase-revenue-item]", opts.ProjectPhaseRevenueItem)
	setFilterIfPresent(query, "filter[job-production-plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[revenue-date]", opts.RevenueDate)
	setFilterIfPresent(query, "filter[revenue-date-min]", opts.RevenueDateMin)
	setFilterIfPresent(query, "filter[revenue-date-max]", opts.RevenueDateMax)

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-revenue-item-actuals", query)
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

	rows := buildProjectPhaseRevenueItemActualRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectPhaseRevenueItemActualsTable(cmd, rows)
}

func parseProjectPhaseRevenueItemActualsListOptions(cmd *cobra.Command) (projectPhaseRevenueItemActualsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectPhaseRevenueItem, _ := cmd.Flags().GetString("project-phase-revenue-item")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	createdBy, _ := cmd.Flags().GetString("created-by")
	project, _ := cmd.Flags().GetString("project")
	revenueDate, _ := cmd.Flags().GetString("revenue-date")
	revenueDateMin, _ := cmd.Flags().GetString("revenue-date-min")
	revenueDateMax, _ := cmd.Flags().GetString("revenue-date-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseRevenueItemActualsListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		ProjectPhaseRevenueItem: projectPhaseRevenueItem,
		JobProductionPlan:       jobProductionPlan,
		CreatedBy:               createdBy,
		Project:                 project,
		RevenueDate:             revenueDate,
		RevenueDateMin:          revenueDateMin,
		RevenueDateMax:          revenueDateMax,
	}, nil
}

func buildProjectPhaseRevenueItemActualRows(resp jsonAPIResponse) []projectPhaseRevenueItemActualRow {
	rows := make([]projectPhaseRevenueItemActualRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectPhaseRevenueItemActualRow{
			ID:                       resource.ID,
			RevenueDate:              formatDate(stringAttr(resource.Attributes, "revenue-date")),
			Quantity:                 stringAttr(resource.Attributes, "quantity"),
			QuantityStrategyExplicit: stringAttr(resource.Attributes, "quantity-strategy-explicit"),
			PricePerUnit:             stringAttr(resource.Attributes, "price-per-unit"),
			RevenueAmount:            stringAttr(resource.Attributes, "revenue-amount"),
			QuantityIndirect:         stringAttr(resource.Meta, "quantity_indirect"),
			CostAmount:               stringAttr(resource.Meta, "cost_amount"),
		}

		if rel, ok := resource.Relationships["project-phase-revenue-item"]; ok && rel.Data != nil {
			row.ProjectPhaseRevenueItemID = rel.Data.ID
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

func renderProjectPhaseRevenueItemActualsTable(cmd *cobra.Command, rows []projectPhaseRevenueItemActualRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project phase revenue item actuals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREV ITEM\tJOB PLAN\tREV DATE\tQTY\tSTRATEGY\tPRICE/UNIT\tREV AMOUNT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ProjectPhaseRevenueItemID,
			row.JobProductionPlanID,
			row.RevenueDate,
			row.Quantity,
			row.QuantityStrategyExplicit,
			row.PricePerUnit,
			row.RevenueAmount,
		)
	}
	return writer.Flush()
}
