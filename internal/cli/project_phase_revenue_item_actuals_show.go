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

type projectPhaseRevenueItemActualsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectPhaseRevenueItemActualDetails struct {
	ID                                         string   `json:"id"`
	Quantity                                   string   `json:"quantity,omitempty"`
	RevenueDate                                string   `json:"revenue_date,omitempty"`
	QuantityStrategyExplicit                   string   `json:"quantity_strategy_explicit,omitempty"`
	PricePerUnit                               string   `json:"price_per_unit,omitempty"`
	RevenueAmount                              string   `json:"revenue_amount,omitempty"`
	QuantityIndirect                           string   `json:"quantity_indirect,omitempty"`
	CostAmount                                 string   `json:"cost_amount,omitempty"`
	ProjectPhaseRevenueItemID                  string   `json:"project_phase_revenue_item_id,omitempty"`
	JobProductionPlanProjectPhaseRevenueItemID string   `json:"job_production_plan_project_phase_revenue_item_id,omitempty"`
	JobProductionPlanID                        string   `json:"job_production_plan_id,omitempty"`
	CreatedByID                                string   `json:"created_by_id,omitempty"`
	ProjectPhaseCostItemActualIDs              []string `json:"project_phase_cost_item_actual_ids,omitempty"`
	CommentIDs                                 []string `json:"comment_ids,omitempty"`
	FileAttachmentIDs                          []string `json:"file_attachment_ids,omitempty"`
}

func newProjectPhaseRevenueItemActualsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project phase revenue item actual details",
		Long: `Show the full details of a project phase revenue item actual.

Output Fields:
  ID                              Revenue item actual identifier
  Quantity                        Actual quantity (resolved)
  Revenue Date                    Revenue date
  Quantity Strategy               Quantity strategy (direct/indirect)
  Price Per Unit                  Resolved price per unit
  Revenue Amount                  Total revenue amount
  Quantity Indirect               Indirect quantity (if available)
  Cost Amount                     Total related cost amount (if available)
  Project Phase Revenue Item ID   Related project phase revenue item
  Job Production Plan Project Phase Revenue Item ID  Related job production plan project phase revenue item
  Job Production Plan ID           Related job production plan
  Created By ID                    Created-by user ID
  Project Phase Cost Item Actual IDs  Related cost item actual IDs
  Comment IDs                      Related comment IDs
  File Attachment IDs              Related file attachment IDs

Arguments:
  <id>    Project phase revenue item actual ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a project phase revenue item actual
  xbe view project-phase-revenue-item-actuals show 123

  # JSON output
  xbe view project-phase-revenue-item-actuals show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectPhaseRevenueItemActualsShow,
	}
	initProjectPhaseRevenueItemActualsShowFlags(cmd)
	return cmd
}

func init() {
	projectPhaseRevenueItemActualsCmd.AddCommand(newProjectPhaseRevenueItemActualsShowCmd())
}

func initProjectPhaseRevenueItemActualsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectPhaseRevenueItemActualsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseProjectPhaseRevenueItemActualsShowOptions(cmd)
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
		return fmt.Errorf("project phase revenue item actual id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[project-phase-revenue-item-actuals]", "quantity,revenue-date,quantity-strategy-explicit,price-per-unit,revenue-amount,project-phase-revenue-item,job-production-plan,job-production-plan-project-phase-revenue-item,created-by,comments,file-attachments,project-phase-cost-item-actuals")
	query.Set("meta[project-phase-revenue-item-actuals]", "quantity-indirect,cost-amount")

	body, _, err := client.Get(cmd.Context(), "/v1/project-phase-revenue-item-actuals/"+id, query)
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

	details := buildProjectPhaseRevenueItemActualDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectPhaseRevenueItemActualDetails(cmd, details)
}

func parseProjectPhaseRevenueItemActualsShowOptions(cmd *cobra.Command) (projectPhaseRevenueItemActualsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectPhaseRevenueItemActualsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectPhaseRevenueItemActualDetails(resp jsonAPISingleResponse) projectPhaseRevenueItemActualDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := projectPhaseRevenueItemActualDetails{
		ID:                       resource.ID,
		Quantity:                 stringAttr(attrs, "quantity"),
		RevenueDate:              formatDate(stringAttr(attrs, "revenue-date")),
		QuantityStrategyExplicit: stringAttr(attrs, "quantity-strategy-explicit"),
		PricePerUnit:             stringAttr(attrs, "price-per-unit"),
		RevenueAmount:            stringAttr(attrs, "revenue-amount"),
		QuantityIndirect:         stringAttr(resource.Meta, "quantity_indirect"),
		CostAmount:               stringAttr(resource.Meta, "cost_amount"),
	}

	if rel, ok := resource.Relationships["project-phase-revenue-item"]; ok && rel.Data != nil {
		details.ProjectPhaseRevenueItemID = rel.Data.ID
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
	if rel, ok := resource.Relationships["project-phase-cost-item-actuals"]; ok {
		details.ProjectPhaseCostItemActualIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["comments"]; ok {
		details.CommentIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["file-attachments"]; ok {
		details.FileAttachmentIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderProjectPhaseRevenueItemActualDetails(cmd *cobra.Command, details projectPhaseRevenueItemActualDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectPhaseRevenueItemID != "" {
		fmt.Fprintf(out, "Project Phase Revenue Item ID: %s\n", details.ProjectPhaseRevenueItemID)
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
	if details.RevenueDate != "" {
		fmt.Fprintf(out, "Revenue Date: %s\n", details.RevenueDate)
	}
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.QuantityStrategyExplicit != "" {
		fmt.Fprintf(out, "Quantity Strategy: %s\n", details.QuantityStrategyExplicit)
	}
	if details.PricePerUnit != "" {
		fmt.Fprintf(out, "Price Per Unit: %s\n", details.PricePerUnit)
	}
	if details.RevenueAmount != "" {
		fmt.Fprintf(out, "Revenue Amount: %s\n", details.RevenueAmount)
	}
	if details.QuantityIndirect != "" {
		fmt.Fprintf(out, "Quantity Indirect: %s\n", details.QuantityIndirect)
	}
	if details.CostAmount != "" {
		fmt.Fprintf(out, "Cost Amount: %s\n", details.CostAmount)
	}
	if len(details.ProjectPhaseCostItemActualIDs) > 0 {
		fmt.Fprintf(out, "Project Phase Cost Item Actual IDs: %s\n", strings.Join(details.ProjectPhaseCostItemActualIDs, ", "))
	}
	if len(details.CommentIDs) > 0 {
		fmt.Fprintf(out, "Comment IDs: %s\n", strings.Join(details.CommentIDs, ", "))
	}
	if len(details.FileAttachmentIDs) > 0 {
		fmt.Fprintf(out, "File Attachment IDs: %s\n", strings.Join(details.FileAttachmentIDs, ", "))
	}

	return nil
}
