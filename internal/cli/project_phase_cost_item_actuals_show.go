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

type projectPhaseCostItemActualsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseCostItemActualDetails struct {
	ID                                         string   `json:"id"`
	Quantity                                   string   `json:"quantity,omitempty"`
	PricePerUnitExplicit                       string   `json:"price_per_unit_explicit,omitempty"`
	PricePerUnit                               string   `json:"price_per_unit,omitempty"`
	CostAmount                                 string   `json:"cost_amount,omitempty"`
	ProjectPhaseCostItemID                     string   `json:"project_phase_cost_item_id,omitempty"`
	ProjectPhaseRevenueItemActualID            string   `json:"project_phase_revenue_item_actual_id,omitempty"`
	JobProductionPlanProjectPhaseRevenueItemID string   `json:"job_production_plan_project_phase_revenue_item_id,omitempty"`
	JobProductionPlanID                        string   `json:"job_production_plan_id,omitempty"`
	CreatedByID                                string   `json:"created_by_id,omitempty"`
	CostCodeAllocationID                       string   `json:"cost_code_allocation_id,omitempty"`
	CostCodeAllocationType                     string   `json:"cost_code_allocation_type,omitempty"`
	CommentIDs                                 []string `json:"comment_ids,omitempty"`
	FileAttachmentIDs                          []string `json:"file_attachment_ids,omitempty"`
}

func newProjectPhaseCostItemActualsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase cost item actual details",
		Long: `Show the full details of a project phase cost item actual.

Output Fields:
  ID                              Cost item actual identifier
  Quantity                        Actual quantity
  Price Per Unit (Explicit)       Explicit price per unit (if set)
  Price Per Unit                  Resolved price per unit
  Cost Amount                     Total cost amount
  Project Phase Cost Item ID      Related project phase cost item
  Project Phase Revenue Item Actual ID  Related project phase revenue item actual
  Job Production Plan Project Phase Revenue Item ID  Related job production plan project phase revenue item
  Job Production Plan ID           Related job production plan
  Created By ID                    Created-by user ID
  Cost Code Allocation             Allocation type and ID (if present)
  Comment IDs                      Related comment IDs
  File Attachment IDs              Related file attachment IDs

Arguments:
  <id>    Project phase cost item actual ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project phase cost item actual
  xbe view project-phase-cost-item-actuals show 123

  # JSON output
  xbe view project-phase-cost-item-actuals show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseCostItemActualsShow,
	}
	initProjectPhaseCostItemActualsShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseCostItemActualsCmd.AddCommand(newProjectPhaseCostItemActualsShowCmd())
}

func initProjectPhaseCostItemActualsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseCostItemActualsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectPhaseCostItemActualsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project phase cost item actual id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-cost-item-actuals]", "quantity,price-per-unit-explicit,price-per-unit,cost-amount,project-phase-cost-item,project-phase-revenue-item-actual,job-production-plan,job-production-plan-project-phase-revenue-item,created-by,cost-code-allocation,comments,file-attachments")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-cost-item-actuals/"+id, query)
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

	details := buildProjectPhaseCostItemActualDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseCostItemActualDetails(cmd, details)
}

func parseProjectPhaseCostItemActualsShowOptions(cmd *cobra.Command) (projectPhaseCostItemActualsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseCostItemActualsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseCostItemActualDetails(resp jsonAPISingleResponse) projectPhaseCostItemActualDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := projectPhaseCostItemActualDetails{
		ID:                   resource.ID,
		Quantity:             stringAttr(attrs, "quantity"),
		PricePerUnitExplicit: stringAttr(attrs, "price-per-unit-explicit"),
		PricePerUnit:         stringAttr(attrs, "price-per-unit"),
		CostAmount:           stringAttr(attrs, "cost-amount"),
	}

	if rel, ok := resource.Relationships["project-phase-cost-item"]; ok && rel.Data != nil {
		details.ProjectPhaseCostItemID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-phase-revenue-item-actual"]; ok && rel.Data != nil {
		details.ProjectPhaseRevenueItemActualID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan-project-phase-revenue-item"]; ok && rel.Data != nil {
		details.JobProductionPlanProjectPhaseRevenueItemID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["cost-code-allocation"]; ok && rel.Data != nil {
		details.CostCodeAllocationID = rel.Data.ID
		details.CostCodeAllocationType = rel.Data.Type
	}
	if rel, ok := resource.Relationships["comments"]; ok {
		details.CommentIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["file-attachments"]; ok {
		details.FileAttachmentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectPhaseCostItemActualDetails(cmd *cobra.Command, details projectPhaseCostItemActualDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectPhaseCostItemID != "" {
		fmt.Fprintf(out, "Project Phase Cost Item ID: %s\n", details.ProjectPhaseCostItemID)
	}
	if details.ProjectPhaseRevenueItemActualID != "" {
		fmt.Fprintf(out, "Project Phase Revenue Item Actual ID: %s\n", details.ProjectPhaseRevenueItemActualID)
	}
	if details.JobProductionPlanProjectPhaseRevenueItemID != "" {
		fmt.Fprintf(out, "Job Production Plan Project Phase Revenue Item ID: %s\n", details.JobProductionPlanProjectPhaseRevenueItemID)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "Created By ID: %s\n", details.CreatedByID)
	}
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.PricePerUnitExplicit != "" {
		fmt.Fprintf(out, "Price Per Unit (Explicit): %s\n", details.PricePerUnitExplicit)
	}
	if details.PricePerUnit != "" {
		fmt.Fprintf(out, "Price Per Unit: %s\n", details.PricePerUnit)
	}
	if details.CostAmount != "" {
		fmt.Fprintf(out, "Cost Amount: %s\n", details.CostAmount)
	}
	if details.CostCodeAllocationID != "" {
		if details.CostCodeAllocationType != "" {
			fmt.Fprintf(out, "Cost Code Allocation: %s %s\n", details.CostCodeAllocationType, details.CostCodeAllocationID)
		} else {
			fmt.Fprintf(out, "Cost Code Allocation ID: %s\n", details.CostCodeAllocationID)
		}
	}
	if len(details.CommentIDs) > 0 {
		fmt.Fprintf(out, "Comment IDs: %s\n", strings.Join(details.CommentIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachment IDs: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}

	return nil
}
