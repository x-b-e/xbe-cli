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

type projectPhaseCostItemsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseCostItemDetails struct {
	ID                              string   `json:"id"`
	ProjectPhaseRevenueItemID       string   `json:"project_phase_revenue_item_id,omitempty"`
	ProjectCostClassificationID     string   `json:"project_cost_classification_id,omitempty"`
	ProjectCostClassificationName   string   `json:"project_cost_classification_name,omitempty"`
	ProjectResourceClassificationID string   `json:"project_resource_classification_id,omitempty"`
	UnitOfMeasureID                 string   `json:"unit_of_measure_id,omitempty"`
	CostCodeID                      string   `json:"cost_code_id,omitempty"`
	IsRevenueQuantityDriver         bool     `json:"is_revenue_quantity_driver"`
	PriceEstimateID                 string   `json:"price_estimate_id,omitempty"`
	QuantityEstimateID              string   `json:"quantity_estimate_id,omitempty"`
	PriceEstimateIDs                []string `json:"price_estimate_ids,omitempty"`
	QuantityEstimateIDs             []string `json:"quantity_estimate_ids,omitempty"`
	ProjectPhaseCostItemActualIDs   []string `json:"project_phase_cost_item_actual_ids,omitempty"`
}

func newProjectPhaseCostItemsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase cost item details",
		Long: `Show the full details of a project phase cost item.

Output Fields:
  ID                       Cost item identifier
  Project Phase Revenue Item ID
  Project Cost Classification ID
  Project Cost Classification Name
  Project Resource Classification ID
  Unit of Measure ID
  Cost Code ID
  Is Revenue Quantity Driver
  Price Estimate ID
  Quantity Estimate ID
  Price Estimate IDs
  Quantity Estimate IDs
  Project Phase Cost Item Actual IDs

Arguments:
  <id>    The project phase cost item ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project phase cost item
  xbe view project-phase-cost-items show 123

  # JSON output
  xbe view project-phase-cost-items show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseCostItemsShow,
	}
	initProjectPhaseCostItemsShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseCostItemsCmd.AddCommand(newProjectPhaseCostItemsShowCmd())
}

func initProjectPhaseCostItemsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseCostItemsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectPhaseCostItemsShowOptions(cmd)
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
		return fmt.Errorf("project phase cost item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-cost-items]", "project-cost-classification-name,is-revenue-quantity-driver,project-phase-revenue-item,project-cost-classification,project-resource-classification,unit-of-measure,cost-code,price-estimate,price-estimates,quantity-estimate,quantity-estimates,project-phase-cost-item-actuals")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-cost-items/"+id, query)
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

	details := buildProjectPhaseCostItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseCostItemDetails(cmd, details)
}

func parseProjectPhaseCostItemsShowOptions(cmd *cobra.Command) (projectPhaseCostItemsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseCostItemsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseCostItemDetails(resp jsonAPISingleResponse) projectPhaseCostItemDetails {
	resource := resp.Data
	attrs := resource.Attributes

	return projectPhaseCostItemDetails{
		ID:                              resource.ID,
		ProjectPhaseRevenueItemID:       relationshipIDFromMap(resource.Relationships, "project-phase-revenue-item"),
		ProjectCostClassificationID:     relationshipIDFromMap(resource.Relationships, "project-cost-classification"),
		ProjectCostClassificationName:   stringAttr(attrs, "project-cost-classification-name"),
		ProjectResourceClassificationID: relationshipIDFromMap(resource.Relationships, "project-resource-classification"),
		UnitOfMeasureID:                 relationshipIDFromMap(resource.Relationships, "unit-of-measure"),
		CostCodeID:                      relationshipIDFromMap(resource.Relationships, "cost-code"),
		IsRevenueQuantityDriver:         boolAttr(attrs, "is-revenue-quantity-driver"),
		PriceEstimateID:                 relationshipIDFromMap(resource.Relationships, "price-estimate"),
		QuantityEstimateID:              relationshipIDFromMap(resource.Relationships, "quantity-estimate"),
		PriceEstimateIDs:                relationshipIDsFromMap(resource.Relationships, "price-estimates"),
		QuantityEstimateIDs:             relationshipIDsFromMap(resource.Relationships, "quantity-estimates"),
		ProjectPhaseCostItemActualIDs:   relationshipIDsFromMap(resource.Relationships, "project-phase-cost-item-actuals"),
	}
}

func renderProjectPhaseCostItemDetails(cmd *cobra.Command, details projectPhaseCostItemDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectPhaseRevenueItemID != "" {
		fmt.Fprintf(out, "Project Phase Revenue Item ID: %s\n", details.ProjectPhaseRevenueItemID)
	}
	if details.ProjectCostClassificationID != "" {
		fmt.Fprintf(out, "Project Cost Classification ID: %s\n", details.ProjectCostClassificationID)
	}
	if details.ProjectCostClassificationName != "" {
		fmt.Fprintf(out, "Project Cost Classification Name: %s\n", details.ProjectCostClassificationName)
	}
	if details.ProjectResourceClassificationID != "" {
		fmt.Fprintf(out, "Project Resource Classification ID: %s\n", details.ProjectResourceClassificationID)
	}
	if details.UnitOfMeasureID != "" {
		fmt.Fprintf(out, "Unit of Measure ID: %s\n", details.UnitOfMeasureID)
	}
	if details.CostCodeID != "" {
		fmt.Fprintf(out, "Cost Code ID: %s\n", details.CostCodeID)
	}
	fmt.Fprintf(out, "Is Revenue Quantity Driver: %t\n", details.IsRevenueQuantityDriver)
	if details.PriceEstimateID != "" {
		fmt.Fprintf(out, "Price Estimate ID: %s\n", details.PriceEstimateID)
	}
	if details.QuantityEstimateID != "" {
		fmt.Fprintf(out, "Quantity Estimate ID: %s\n", details.QuantityEstimateID)
	}
	if len(details.PriceEstimateIDs) > 0 {
		fmt.Fprintf(out, "Price Estimate IDs: %s\n", strings.Join(details.PriceEstimateIDs, ", "))
	}
	if len(details.QuantityEstimateIDs) > 0 {
		fmt.Fprintf(out, "Quantity Estimate IDs: %s\n", strings.Join(details.QuantityEstimateIDs, ", "))
	}
	if len(details.ProjectPhaseCostItemActualIDs) > 0 {
		fmt.Fprintf(out, "Project Phase Cost Item Actual IDs: %s\n", strings.Join(details.ProjectPhaseCostItemActualIDs, ", "))
	}

	return nil
}
