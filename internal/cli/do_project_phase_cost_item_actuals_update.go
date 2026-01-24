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

type doProjectPhaseCostItemActualsUpdateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ID                            string
	ProjectPhaseRevenueItemActual string
	Quantity                      string
	PricePerUnitExplicit          string
	CreatedBy                     string
}

func newDoProjectPhaseCostItemActualsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project phase cost item actual",
		Long: `Update a project phase cost item actual.

Optional flags:
  --quantity                        Actual quantity
  --price-per-unit-explicit         Explicit price per unit
  --project-phase-revenue-item-actual  Project phase revenue item actual ID
  --created-by                      Created-by user ID (admin only)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update quantity
  xbe do project-phase-cost-item-actuals update 123 --quantity 12

  # Update price per unit
  xbe do project-phase-cost-item-actuals update 123 --price-per-unit-explicit 30.5

  # JSON output
  xbe do project-phase-cost-item-actuals update 123 --quantity 12 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectPhaseCostItemActualsUpdate,
	}
	initDoProjectPhaseCostItemActualsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseCostItemActualsCmd.AddCommand(newDoProjectPhaseCostItemActualsUpdateCmd())
}

func initDoProjectPhaseCostItemActualsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("quantity", "", "Actual quantity")
	cmd.Flags().String("price-per-unit-explicit", "", "Explicit price per unit")
	cmd.Flags().String("project-phase-revenue-item-actual", "", "Project phase revenue item actual ID")
	cmd.Flags().String("created-by", "", "Created-by user ID (admin only)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseCostItemActualsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectPhaseCostItemActualsUpdateOptions(cmd, args)
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

	relationships := map[string]any{}
	if cmd.Flags().Changed("project-phase-revenue-item-actual") {
		relationships["project-phase-revenue-item-actual"] = map[string]any{
			"data": map[string]any{
				"type": "project-phase-revenue-item-actuals",
				"id":   opts.ProjectPhaseRevenueItemActual,
			},
		}
	}
	if cmd.Flags().Changed("created-by") {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-phase-cost-item-actuals",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-phase-cost-item-actuals/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project phase cost item actual %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectPhaseCostItemActualsUpdateOptions(cmd *cobra.Command, args []string) (doProjectPhaseCostItemActualsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	quantity, _ := cmd.Flags().GetString("quantity")
	pricePerUnitExplicit, _ := cmd.Flags().GetString("price-per-unit-explicit")
	projectPhaseRevenueItemActual, _ := cmd.Flags().GetString("project-phase-revenue-item-actual")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseCostItemActualsUpdateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ID:                            args[0],
		ProjectPhaseRevenueItemActual: projectPhaseRevenueItemActual,
		Quantity:                      quantity,
		PricePerUnitExplicit:          pricePerUnitExplicit,
		CreatedBy:                     createdBy,
	}, nil
}
