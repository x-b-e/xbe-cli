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

type doProjectPhaseRevenueItemActualsCreateOptions struct {
	BaseURL                                  string
	Token                                    string
	JSON                                     bool
	ProjectPhaseRevenueItem                  string
	JobProductionPlan                        string
	JobProductionPlanProjectPhaseRevenueItem string
	Quantity                                 string
	RevenueDate                              string
	QuantityStrategyExplicit                 string
	CreatedBy                                string
}

func newDoProjectPhaseRevenueItemActualsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase revenue item actual",
		Long: `Create a project phase revenue item actual.

Required flags:
  --project-phase-revenue-item  Project phase revenue item ID

Optional flags:
  --job-production-plan                   Job production plan ID
  --job-production-plan-project-phase-revenue-item  Job production plan project phase revenue item ID
  --quantity                               Actual quantity (direct)
  --revenue-date                           Revenue date (YYYY-MM-DD)
  --quantity-strategy-explicit             Quantity strategy (direct or indirect)
  --created-by                             Created-by user ID (admin only)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project phase revenue item actual
  xbe do project-phase-revenue-item-actuals create \
    --project-phase-revenue-item 123 \
    --job-production-plan-project-phase-revenue-item 456 \
    --quantity 10 \
    --revenue-date 2026-01-23

  # JSON output
  xbe do project-phase-revenue-item-actuals create \
    --project-phase-revenue-item 123 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseRevenueItemActualsCreate,
	}
	initDoProjectPhaseRevenueItemActualsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseRevenueItemActualsCmd.AddCommand(newDoProjectPhaseRevenueItemActualsCreateCmd())
}

func initDoProjectPhaseRevenueItemActualsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-phase-revenue-item", "", "Project phase revenue item ID (required)")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("job-production-plan-project-phase-revenue-item", "", "Job production plan project phase revenue item ID")
	cmd.Flags().String("quantity", "", "Actual quantity")
	cmd.Flags().String("revenue-date", "", "Revenue date (YYYY-MM-DD)")
	cmd.Flags().String("quantity-strategy-explicit", "", "Quantity strategy (direct or indirect)")
	cmd.Flags().String("created-by", "", "Created-by user ID (admin only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-phase-revenue-item")
}

func runDoProjectPhaseRevenueItemActualsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseRevenueItemActualsCreateOptions(cmd)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("quantity") {
		attributes["quantity"] = opts.Quantity
	}
	if cmd.Flags().Changed("revenue-date") {
		attributes["revenue-date"] = opts.RevenueDate
	}
	if cmd.Flags().Changed("quantity-strategy-explicit") {
		attributes["quantity-strategy-explicit"] = opts.QuantityStrategyExplicit
	}

	relationships := map[string]any{
		"project-phase-revenue-item": map[string]any{
			"data": map[string]any{
				"type": "project-phase-revenue-items",
				"id":   opts.ProjectPhaseRevenueItem,
			},
		},
	}

	if opts.JobProductionPlan != "" {
		relationships["job-production-plan"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		}
	}
	if opts.JobProductionPlanProjectPhaseRevenueItem != "" {
		relationships["job-production-plan-project-phase-revenue-item"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-project-phase-revenue-items",
				"id":   opts.JobProductionPlanProjectPhaseRevenueItem,
			},
		}
	}
	if opts.CreatedBy != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "project-phase-revenue-item-actuals",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-revenue-item-actuals", jsonBody)
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

	row := projectPhaseRevenueItemActualRow{
		ID:                       resp.Data.ID,
		RevenueDate:              formatDate(stringAttr(resp.Data.Attributes, "revenue-date")),
		Quantity:                 stringAttr(resp.Data.Attributes, "quantity"),
		QuantityStrategyExplicit: stringAttr(resp.Data.Attributes, "quantity-strategy-explicit"),
		PricePerUnit:             stringAttr(resp.Data.Attributes, "price-per-unit"),
		RevenueAmount:            stringAttr(resp.Data.Attributes, "revenue-amount"),
		QuantityIndirect:         stringAttr(resp.Data.Meta, "quantity_indirect"),
		CostAmount:               stringAttr(resp.Data.Meta, "cost_amount"),
	}
	if rel, ok := resp.Data.Relationships["project-phase-revenue-item"]; ok && rel.Data != nil {
		row.ProjectPhaseRevenueItemID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-production-plan-project-phase-revenue-item"]; ok && rel.Data != nil {
		row.JobProductionPlanProjectPhaseRevenueItemID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase revenue item actual %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectPhaseRevenueItemActualsCreateOptions(cmd *cobra.Command) (doProjectPhaseRevenueItemActualsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectPhaseRevenueItem, _ := cmd.Flags().GetString("project-phase-revenue-item")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	jobProductionPlanProjectPhaseRevenueItem, _ := cmd.Flags().GetString("job-production-plan-project-phase-revenue-item")
	quantity, _ := cmd.Flags().GetString("quantity")
	revenueDate, _ := cmd.Flags().GetString("revenue-date")
	quantityStrategyExplicit, _ := cmd.Flags().GetString("quantity-strategy-explicit")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseRevenueItemActualsCreateOptions{
		BaseURL:                                  baseURL,
		Token:                                    token,
		JSON:                                     jsonOut,
		ProjectPhaseRevenueItem:                  projectPhaseRevenueItem,
		JobProductionPlan:                        jobProductionPlan,
		JobProductionPlanProjectPhaseRevenueItem: jobProductionPlanProjectPhaseRevenueItem,
		Quantity:                                 quantity,
		RevenueDate:                              revenueDate,
		QuantityStrategyExplicit:                 quantityStrategyExplicit,
		CreatedBy:                                createdBy,
	}, nil
}
