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

type doJobProductionPlanProjectPhaseRevenueItemsCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	JobProductionPlan       string
	ProjectPhaseRevenueItem string
	Quantity                string
}

func newDoJobProductionPlanProjectPhaseRevenueItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan project phase revenue item",
		Long: `Create a job production plan project phase revenue item.

Required flags:
  --job-production-plan        Job production plan ID (required)
  --project-phase-revenue-item Project phase revenue item ID (required)

Optional:
  --quantity                   Planned quantity

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an item with quantity
  xbe do job-production-plan-project-phase-revenue-items create \\
    --job-production-plan 123 \\
    --project-phase-revenue-item 456 \\
    --quantity 25`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanProjectPhaseRevenueItemsCreate,
	}
	initDoJobProductionPlanProjectPhaseRevenueItemsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanProjectPhaseRevenueItemsCmd.AddCommand(newDoJobProductionPlanProjectPhaseRevenueItemsCreateCmd())
}

func initDoJobProductionPlanProjectPhaseRevenueItemsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("project-phase-revenue-item", "", "Project phase revenue item ID (required)")
	cmd.Flags().String("quantity", "", "Planned quantity")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanProjectPhaseRevenueItemsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanProjectPhaseRevenueItemsCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	jobProductionPlanID := strings.TrimSpace(opts.JobProductionPlan)
	projectPhaseRevenueItemID := strings.TrimSpace(opts.ProjectPhaseRevenueItem)

	if jobProductionPlanID == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if projectPhaseRevenueItemID == "" {
		err := fmt.Errorf("--project-phase-revenue-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   jobProductionPlanID,
			},
		},
		"project-phase-revenue-item": map[string]any{
			"data": map[string]any{
				"type": "project-phase-revenue-items",
				"id":   projectPhaseRevenueItemID,
			},
		},
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}

	requestData := map[string]any{
		"type":          "job-production-plan-project-phase-revenue-items",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		requestData["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": requestData,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-project-phase-revenue-items", jsonBody)
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

	row := buildJobProductionPlanProjectPhaseRevenueItemRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan project phase revenue item %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanProjectPhaseRevenueItemsCreateOptions(cmd *cobra.Command) (doJobProductionPlanProjectPhaseRevenueItemsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	projectPhaseRevenueItem, _ := cmd.Flags().GetString("project-phase-revenue-item")
	quantity, _ := cmd.Flags().GetString("quantity")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanProjectPhaseRevenueItemsCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		JobProductionPlan:       jobProductionPlan,
		ProjectPhaseRevenueItem: projectPhaseRevenueItem,
		Quantity:                quantity,
	}, nil
}

func buildJobProductionPlanProjectPhaseRevenueItemRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanProjectPhaseRevenueItemRow {
	resource := resp.Data
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

	return row
}
