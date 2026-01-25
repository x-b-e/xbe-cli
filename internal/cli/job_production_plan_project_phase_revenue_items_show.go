package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanProjectPhaseRevenueItemsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanProjectPhaseRevenueItemDetails struct {
	ID                               string   `json:"id"`
	JobProductionPlanID              string   `json:"job_production_plan_id,omitempty"`
	ProjectPhaseRevenueItemID        string   `json:"project_phase_revenue_item_id,omitempty"`
	ProjectPhaseRevenueItemActualIDs []string `json:"project_phase_revenue_item_actual_ids,omitempty"`
	ProjectPhaseCostItemActualIDs    []string `json:"project_phase_cost_item_actual_ids,omitempty"`
	Quantity                         string   `json:"quantity,omitempty"`
	CreatedAt                        string   `json:"created_at,omitempty"`
	UpdatedAt                        string   `json:"updated_at,omitempty"`
}

func newJobProductionPlanProjectPhaseRevenueItemsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan project phase revenue item details",
		Long: `Show the full details of a job production plan project phase revenue item.

Output Fields:
  ID
  Job Production Plan ID
  Project Phase Revenue Item ID
  Project Phase Revenue Item Actual IDs
  Project Phase Cost Item Actual IDs
  Quantity
  Created At
  Updated At

Arguments:
  <id>    The item ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a job production plan project phase revenue item
  xbe view job-production-plan-project-phase-revenue-items show 123

  # Get JSON output
  xbe view job-production-plan-project-phase-revenue-items show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanProjectPhaseRevenueItemsShow,
	}
	initJobProductionPlanProjectPhaseRevenueItemsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanProjectPhaseRevenueItemsCmd.AddCommand(newJobProductionPlanProjectPhaseRevenueItemsShowCmd())
}

func initJobProductionPlanProjectPhaseRevenueItemsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanProjectPhaseRevenueItemsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanProjectPhaseRevenueItemsShowOptions(cmd)
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
		return fmt.Errorf("job production plan project phase revenue item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-project-phase-revenue-items/"+id, nil)
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

	details := buildJobProductionPlanProjectPhaseRevenueItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanProjectPhaseRevenueItemDetails(cmd, details)
}

func parseJobProductionPlanProjectPhaseRevenueItemsShowOptions(cmd *cobra.Command) (jobProductionPlanProjectPhaseRevenueItemsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanProjectPhaseRevenueItemsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanProjectPhaseRevenueItemDetails(resp jsonAPISingleResponse) jobProductionPlanProjectPhaseRevenueItemDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanProjectPhaseRevenueItemDetails{
		ID:        resource.ID,
		Quantity:  stringAttr(attrs, "quantity"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-phase-revenue-item"]; ok && rel.Data != nil {
		details.ProjectPhaseRevenueItemID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-phase-revenue-item-actuals"]; ok {
		details.ProjectPhaseRevenueItemActualIDs = relationshipIDStrings(rel)
	}
	if rel, ok := resource.Relationships["project-phase-cost-item-actuals"]; ok {
		details.ProjectPhaseCostItemActualIDs = relationshipIDStrings(rel)
	}

	return details
}

func renderJobProductionPlanProjectPhaseRevenueItemDetails(cmd *cobra.Command, details jobProductionPlanProjectPhaseRevenueItemDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}
	if details.ProjectPhaseRevenueItemID != "" {
		fmt.Fprintf(out, "Project Phase Revenue Item ID: %s\n", details.ProjectPhaseRevenueItemID)
	}
	if len(details.ProjectPhaseRevenueItemActualIDs) > 0 {
		fmt.Fprintf(out, "Project Phase Revenue Item Actual IDs: %s\n", strings.Join(details.ProjectPhaseRevenueItemActualIDs, ", "))
	}
	if len(details.ProjectPhaseCostItemActualIDs) > 0 {
		fmt.Fprintf(out, "Project Phase Cost Item Actual IDs: %s\n", strings.Join(details.ProjectPhaseCostItemActualIDs, ", "))
	}
	if details.Quantity != "" {
		fmt.Fprintf(out, "Quantity: %s\n", details.Quantity)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}

func relationshipIDStrings(rel jsonAPIRelationship) []string {
	if rel.raw == nil {
		return nil
	}
	var refs []jsonAPIResourceIdentifier
	if err := json.Unmarshal(rel.raw, &refs); err != nil {
		return nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.ID)
	}
	return ids
}
