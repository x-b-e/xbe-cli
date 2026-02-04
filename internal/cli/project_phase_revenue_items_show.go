package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectPhaseRevenueItemsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseRevenueItemDetails struct {
	ID                                          string   `json:"id"`
	ProjectPhaseID                              string   `json:"project_phase_id,omitempty"`
	ProjectRevenueItemID                        string   `json:"project_revenue_item_id,omitempty"`
	ProjectRevenueClassificationID              string   `json:"project_revenue_classification_id,omitempty"`
	QuantityStrategy                            string   `json:"quantity_strategy,omitempty"`
	Note                                        string   `json:"note,omitempty"`
	PricePerUnit                                string   `json:"price_per_unit,omitempty"`
	ActualQuantity                              string   `json:"actual_quantity,omitempty"`
	ActualRevenueAmount                         string   `json:"actual_revenue_amount,omitempty"`
	ActualCostAmount                            string   `json:"actual_cost_amount,omitempty"`
	ActualProfitAmount                          string   `json:"actual_profit_amount,omitempty"`
	QuantityEstimateID                          string   `json:"quantity_estimate_id,omitempty"`
	QuantityEstimateIDs                         []string `json:"quantity_estimate_ids,omitempty"`
	ProjectPhaseCostItemIDs                     []string `json:"project_phase_cost_item_ids,omitempty"`
	ProjectPhaseRevenueItemActualIDs            []string `json:"project_phase_revenue_item_actual_ids,omitempty"`
	JobProductionPlanProjectPhaseRevenueItemIDs []string `json:"job_production_plan_project_phase_revenue_item_ids,omitempty"`
}

func newProjectPhaseRevenueItemsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase revenue item details",
		Long: `Show the full details of a project phase revenue item.

Output Fields:
  ID
  Project Phase ID
  Project Revenue Item ID
  Project Revenue Classification ID
  Quantity Strategy
  Note
  Price Per Unit
  Actual Quantity
  Actual Revenue Amount
  Actual Cost Amount
  Actual Profit Amount
  Quantity Estimate ID
  Quantity Estimate IDs
  Project Phase Cost Item IDs
  Project Phase Revenue Item Actual IDs
  Job Production Plan Project Phase Revenue Item IDs

Arguments:
  <id>    The project phase revenue item ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project phase revenue item
  xbe view project-phase-revenue-items show 123

  # JSON output
  xbe view project-phase-revenue-items show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseRevenueItemsShow,
	}
	initProjectPhaseRevenueItemsShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseRevenueItemsCmd.AddCommand(newProjectPhaseRevenueItemsShowCmd())
}

func initProjectPhaseRevenueItemsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseRevenueItemsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectPhaseRevenueItemsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project phase revenue item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-revenue-items]", "quantity-strategy,note,price-per-unit,actual-quantity,actual-revenue-amount,actual-cost-amount,actual-profit-amount,project-phase,project-revenue-item,project-revenue-classification,quantity-estimate,quantity-estimates,project-phase-cost-items,project-phase-revenue-item-actuals,job-production-plan-project-phase-revenue-items")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-revenue-items/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectPhaseRevenueItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseRevenueItemDetails(cmd, details)
}

func parseProjectPhaseRevenueItemsShowOptions(cmd *cobra.Command) (projectPhaseRevenueItemsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseRevenueItemsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseRevenueItemDetails(resp jsonAPISingleResponse) projectPhaseRevenueItemDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return projectPhaseRevenueItemDetails{
		ID:                               resource.ID,
		ProjectPhaseID:                   relationshipIDFromMap(resource.Relationships, "project-phase"),
		ProjectRevenueItemID:             relationshipIDFromMap(resource.Relationships, "project-revenue-item"),
		ProjectRevenueClassificationID:   relationshipIDFromMap(resource.Relationships, "project-revenue-classification"),
		QuantityStrategy:                 stringAttr(attrs, "quantity-strategy"),
		Note:                             stringAttr(attrs, "note"),
		PricePerUnit:                     stringAttr(attrs, "price-per-unit"),
		ActualQuantity:                   stringAttr(attrs, "actual-quantity"),
		ActualRevenueAmount:              stringAttr(attrs, "actual-revenue-amount"),
		ActualCostAmount:                 stringAttr(attrs, "actual-cost-amount"),
		ActualProfitAmount:               stringAttr(attrs, "actual-profit-amount"),
		QuantityEstimateID:               relationshipIDFromMap(resource.Relationships, "quantity-estimate"),
		QuantityEstimateIDs:              relationshipIDsFromMap(resource.Relationships, "quantity-estimates"),
		ProjectPhaseCostItemIDs:          relationshipIDsFromMap(resource.Relationships, "project-phase-cost-items"),
		ProjectPhaseRevenueItemActualIDs: relationshipIDsFromMap(resource.Relationships, "project-phase-revenue-item-actuals"),
		JobProductionPlanProjectPhaseRevenueItemIDs: relationshipIDsFromMap(resource.Relationships, "job-production-plan-project-phase-revenue-items"),
	}
}

func renderProjectPhaseRevenueItemDetails(cmd *cobra.Command, details projectPhaseRevenueItemDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectPhaseID != "" {
		fmt.Fprintf(out, "Project Phase ID: %s\n", details.ProjectPhaseID)
	}
	if details.ProjectRevenueItemID != "" {
		fmt.Fprintf(out, "Project Revenue Item ID: %s\n", details.ProjectRevenueItemID)
	}
	if details.ProjectRevenueClassificationID != "" {
		fmt.Fprintf(out, "Project Revenue Classification ID: %s\n", details.ProjectRevenueClassificationID)
	}
	if details.QuantityStrategy != "" {
		fmt.Fprintf(out, "Quantity Strategy: %s\n", details.QuantityStrategy)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.PricePerUnit != "" {
		fmt.Fprintf(out, "Price Per Unit: %s\n", details.PricePerUnit)
	}
	if details.ActualQuantity != "" {
		fmt.Fprintf(out, "Actual Quantity: %s\n", details.ActualQuantity)
	}
	if details.ActualRevenueAmount != "" {
		fmt.Fprintf(out, "Actual Revenue Amount: %s\n", details.ActualRevenueAmount)
	}
	if details.ActualCostAmount != "" {
		fmt.Fprintf(out, "Actual Cost Amount: %s\n", details.ActualCostAmount)
	}
	if details.ActualProfitAmount != "" {
		fmt.Fprintf(out, "Actual Profit Amount: %s\n", details.ActualProfitAmount)
	}
	if details.QuantityEstimateID != "" {
		fmt.Fprintf(out, "Quantity Estimate ID: %s\n", details.QuantityEstimateID)
	}
	if len(details.QuantityEstimateIDs) > 0 {
		fmt.Fprintf(out, "Quantity Estimate IDs: %s\n", strings.Join(details.QuantityEstimateIDs, ", "))
	}
	if len(details.ProjectPhaseCostItemIDs) > 0 {
		fmt.Fprintf(out, "Project Phase Cost Item IDs: %s\n", strings.Join(details.ProjectPhaseCostItemIDs, ", "))
	}
	if len(details.ProjectPhaseRevenueItemActualIDs) > 0 {
		fmt.Fprintf(out, "Project Phase Revenue Item Actual IDs: %s\n", strings.Join(details.ProjectPhaseRevenueItemActualIDs, ", "))
	}
	if len(details.JobProductionPlanProjectPhaseRevenueItemIDs) > 0 {
		fmt.Fprintf(out, "Job Production Plan Project Phase Revenue Item IDs: %s\n", strings.Join(details.JobProductionPlanProjectPhaseRevenueItemIDs, ", "))
	}

	return nil
}
