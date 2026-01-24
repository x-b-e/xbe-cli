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

type doProjectPhaseCostItemActualsCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ProjectPhaseCostItem          string
	ProjectPhaseRevenueItemActual string
	Quantity                      string
	PricePerUnitExplicit          string
	CreatedBy                     string
}

func newDoProjectPhaseCostItemActualsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase cost item actual",
		Long: `Create a project phase cost item actual.

Required flags:
  --project-phase-cost-item            Project phase cost item ID
  --project-phase-revenue-item-actual  Project phase revenue item actual ID

Optional flags:
  --quantity                Actual quantity
  --price-per-unit-explicit Explicit price per unit
  --created-by              Created-by user ID (admin only)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project phase cost item actual
  xbe do project-phase-cost-item-actuals create \
    --project-phase-cost-item 123 \
    --project-phase-revenue-item-actual 456 \
    --quantity 10 \
    --price-per-unit-explicit 25.5

  # JSON output
  xbe do project-phase-cost-item-actuals create \
    --project-phase-cost-item 123 \
    --project-phase-revenue-item-actual 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseCostItemActualsCreate,
	}
	initDoProjectPhaseCostItemActualsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseCostItemActualsCmd.AddCommand(newDoProjectPhaseCostItemActualsCreateCmd())
}

func initDoProjectPhaseCostItemActualsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-phase-cost-item", "", "Project phase cost item ID (required)")
	cmd.Flags().String("project-phase-revenue-item-actual", "", "Project phase revenue item actual ID (required)")
	cmd.Flags().String("quantity", "", "Actual quantity")
	cmd.Flags().String("price-per-unit-explicit", "", "Explicit price per unit")
	cmd.Flags().String("created-by", "", "Created-by user ID (admin only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-phase-cost-item")
	_ = cmd.MarkFlagRequired("project-phase-revenue-item-actual")
}

func runDoProjectPhaseCostItemActualsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseCostItemActualsCreateOptions(cmd)
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
	if cmd.Flags().Changed("price-per-unit-explicit") {
		attributes["price-per-unit-explicit"] = opts.PricePerUnitExplicit
	}

	relationships := map[string]any{
		"project-phase-cost-item": map[string]any{
			"data": map[string]any{
				"type": "project-phase-cost-items",
				"id":   opts.ProjectPhaseCostItem,
			},
		},
		"project-phase-revenue-item-actual": map[string]any{
			"data": map[string]any{
				"type": "project-phase-revenue-item-actuals",
				"id":   opts.ProjectPhaseRevenueItemActual,
			},
		},
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
			"type":          "project-phase-cost-item-actuals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-cost-item-actuals", jsonBody)
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

	row := projectPhaseCostItemActualRow{
		ID:                   resp.Data.ID,
		Quantity:             stringAttr(resp.Data.Attributes, "quantity"),
		PricePerUnitExplicit: stringAttr(resp.Data.Attributes, "price-per-unit-explicit"),
		PricePerUnit:         stringAttr(resp.Data.Attributes, "price-per-unit"),
		CostAmount:           stringAttr(resp.Data.Attributes, "cost-amount"),
	}
	if rel, ok := resp.Data.Relationships["project-phase-cost-item"]; ok && rel.Data != nil {
		row.ProjectPhaseCostItemID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["project-phase-revenue-item-actual"]; ok && rel.Data != nil {
		row.ProjectPhaseRevenueItemActualID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["job-production-plan-project-phase-revenue-item"]; ok && rel.Data != nil {
		row.JobProductionPlanProjectPhaseRevenueItemID = rel.Data.ID
	}

	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase cost item actual %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectPhaseCostItemActualsCreateOptions(cmd *cobra.Command) (doProjectPhaseCostItemActualsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectPhaseCostItem, _ := cmd.Flags().GetString("project-phase-cost-item")
	projectPhaseRevenueItemActual, _ := cmd.Flags().GetString("project-phase-revenue-item-actual")
	quantity, _ := cmd.Flags().GetString("quantity")
	pricePerUnitExplicit, _ := cmd.Flags().GetString("price-per-unit-explicit")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseCostItemActualsCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ProjectPhaseCostItem:          projectPhaseCostItem,
		ProjectPhaseRevenueItemActual: projectPhaseRevenueItemActual,
		Quantity:                      quantity,
		PricePerUnitExplicit:          pricePerUnitExplicit,
		CreatedBy:                     createdBy,
	}, nil
}
