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

type projectPhaseCostItemsListOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	NoAuth                        bool
	Limit                         int
	Offset                        int
	Sort                          string
	ProjectPhaseRevenueItem       string
	ProjectCostClassification     string
	ProjectResourceClassification string
	UnitOfMeasure                 string
	ProjectPhase                  string
	Project                       string
	IsRevenueQuantityDriver       string
}

type projectPhaseCostItemRow struct {
	ID                              string `json:"id"`
	ProjectPhaseRevenueItemID       string `json:"project_phase_revenue_item_id,omitempty"`
	ProjectCostClassificationID     string `json:"project_cost_classification_id,omitempty"`
	ProjectCostClassificationName   string `json:"project_cost_classification_name,omitempty"`
	ProjectResourceClassificationID string `json:"project_resource_classification_id,omitempty"`
	UnitOfMeasureID                 string `json:"unit_of_measure_id,omitempty"`
	CostCodeID                      string `json:"cost_code_id,omitempty"`
	IsRevenueQuantityDriver         bool   `json:"is_revenue_quantity_driver"`
}

func newProjectPhaseCostItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project phase cost items",
		Long: `List project phase cost items.

Project phase cost items define the cost components for project phase revenue items.

Output Columns:
  ID             Cost item identifier
  REV_ITEM       Project phase revenue item ID
  COST_CLASS     Project cost classification name or ID
  RESOURCE_CLASS Project resource classification ID
  UOM            Unit of measure ID
  COST_CODE      Cost code ID
  REV_QTY_DRIVER Whether this item drives revenue quantity

Filters:
  --project-phase-revenue-item      Filter by project phase revenue item ID
  --project-cost-classification     Filter by project cost classification ID
  --project-resource-classification Filter by project resource classification ID
  --unit-of-measure                 Filter by unit of measure ID
  --project-phase                   Filter by project phase ID
  --project                         Filter by project ID
  --is-revenue-quantity-driver      Filter by revenue quantity driver (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List project phase cost items
  xbe view project-phase-cost-items list

  # Filter by project phase revenue item
  xbe view project-phase-cost-items list --project-phase-revenue-item 123

  # Filter by project
  xbe view project-phase-cost-items list --project 456

  # Filter by revenue quantity driver
  xbe view project-phase-cost-items list --is-revenue-quantity-driver true

  # Output as JSON
  xbe view project-phase-cost-items list --json`,
		Args: cobra.NoArgs,
		RunE: runProjectPhaseCostItemsList,
	}
	initProjectPhaseCostItemsListFlags(cmd)
	return cmd
}

func init() {
	projectPhaseCostItemsCmd.AddCommand(newProjectPhaseCostItemsListCmd())
}

func initProjectPhaseCostItemsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("project-phase-revenue-item", "", "Filter by project phase revenue item ID")
	cmd.Flags().String("project-cost-classification", "", "Filter by project cost classification ID")
	cmd.Flags().String("project-resource-classification", "", "Filter by project resource classification ID")
	cmd.Flags().String("unit-of-measure", "", "Filter by unit of measure ID")
	cmd.Flags().String("project-phase", "", "Filter by project phase ID")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("is-revenue-quantity-driver", "", "Filter by revenue quantity driver (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseCostItemsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProjectPhaseCostItemsListOptions(cmd)
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
	query.Set("fields[project-phase-cost-items]", "project-cost-classification-name,is-revenue-quantity-driver,project-phase-revenue-item,project-cost-classification,project-resource-classification,unit-of-measure,cost-code")

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
	setFilterIfPresent(query, "filter[project-cost-classification]", opts.ProjectCostClassification)
	setFilterIfPresent(query, "filter[project-resource-classification]", opts.ProjectResourceClassification)
	setFilterIfPresent(query, "filter[unit-of-measure]", opts.UnitOfMeasure)
	setFilterIfPresent(query, "filter[project-phase]", opts.ProjectPhase)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[is-revenue-quantity-driver]", opts.IsRevenueQuantityDriver)

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-cost-items", query)
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

	rows := buildProjectPhaseCostItemRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProjectPhaseCostItemsTable(cmd, rows)
}

func parseProjectPhaseCostItemsListOptions(cmd *cobra.Command) (projectPhaseCostItemsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	projectPhaseRevenueItem, _ := cmd.Flags().GetString("project-phase-revenue-item")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	projectResourceClassification, _ := cmd.Flags().GetString("project-resource-classification")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	projectPhase, _ := cmd.Flags().GetString("project-phase")
	project, _ := cmd.Flags().GetString("project")
	isRevenueQuantityDriver, _ := cmd.Flags().GetString("is-revenue-quantity-driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseCostItemsListOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		NoAuth:                        noAuth,
		Limit:                         limit,
		Offset:                        offset,
		Sort:                          sort,
		ProjectPhaseRevenueItem:       projectPhaseRevenueItem,
		ProjectCostClassification:     projectCostClassification,
		ProjectResourceClassification: projectResourceClassification,
		UnitOfMeasure:                 unitOfMeasure,
		ProjectPhase:                  projectPhase,
		Project:                       project,
		IsRevenueQuantityDriver:       isRevenueQuantityDriver,
	}, nil
}

func buildProjectPhaseCostItemRows(resp jsonAPIResponse) []projectPhaseCostItemRow {
	rows := make([]projectPhaseCostItemRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := projectPhaseCostItemRow{
			ID:                              resource.ID,
			ProjectCostClassificationName:   stringAttr(resource.Attributes, "project-cost-classification-name"),
			IsRevenueQuantityDriver:         boolAttr(resource.Attributes, "is-revenue-quantity-driver"),
			ProjectPhaseRevenueItemID:       relationshipIDFromMap(resource.Relationships, "project-phase-revenue-item"),
			ProjectCostClassificationID:     relationshipIDFromMap(resource.Relationships, "project-cost-classification"),
			ProjectResourceClassificationID: relationshipIDFromMap(resource.Relationships, "project-resource-classification"),
			UnitOfMeasureID:                 relationshipIDFromMap(resource.Relationships, "unit-of-measure"),
			CostCodeID:                      relationshipIDFromMap(resource.Relationships, "cost-code"),
		}

		rows = append(rows, row)
	}
	return rows
}

func buildProjectPhaseCostItemRowFromSingle(resp jsonAPISingleResponse) projectPhaseCostItemRow {
	resource := resp.Data
	return projectPhaseCostItemRow{
		ID:                              resource.ID,
		ProjectCostClassificationName:   stringAttr(resource.Attributes, "project-cost-classification-name"),
		IsRevenueQuantityDriver:         boolAttr(resource.Attributes, "is-revenue-quantity-driver"),
		ProjectPhaseRevenueItemID:       relationshipIDFromMap(resource.Relationships, "project-phase-revenue-item"),
		ProjectCostClassificationID:     relationshipIDFromMap(resource.Relationships, "project-cost-classification"),
		ProjectResourceClassificationID: relationshipIDFromMap(resource.Relationships, "project-resource-classification"),
		UnitOfMeasureID:                 relationshipIDFromMap(resource.Relationships, "unit-of-measure"),
		CostCodeID:                      relationshipIDFromMap(resource.Relationships, "cost-code"),
	}
}

func renderProjectPhaseCostItemsTable(cmd *cobra.Command, rows []projectPhaseCostItemRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No project phase cost items found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tREV_ITEM\tCOST_CLASS\tRESOURCE_CLASS\tUOM\tCOST_CODE\tREV_QTY_DRIVER")
	for _, row := range rows {
		costClass := row.ProjectCostClassificationName
		if costClass == "" {
			costClass = row.ProjectCostClassificationID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%t\n",
			row.ID,
			row.ProjectPhaseRevenueItemID,
			truncateString(costClass, 30),
			truncateString(row.ProjectResourceClassificationID, 30),
			truncateString(row.UnitOfMeasureID, 15),
			truncateString(row.CostCodeID, 20),
			row.IsRevenueQuantityDriver,
		)
	}
	return writer.Flush()
}
