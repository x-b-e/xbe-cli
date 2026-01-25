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

type doProjectPhaseCostItemsUpdateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ID                            string
	ProjectCostClassification     string
	ProjectResourceClassification string
	UnitOfMeasure                 string
	CostCode                      string
	IsRevenueQuantityDriver       string
	PriceEstimate                 string
	QuantityEstimate              string
}

func newDoProjectPhaseCostItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project phase cost item",
		Long: `Update an existing project phase cost item.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The project phase cost item ID (required)

Flags:
  --project-cost-classification     Update project cost classification ID (empty to clear)
  --project-resource-classification Update project resource classification ID (empty to clear)
  --unit-of-measure                 Update unit of measure ID (empty to clear)
  --cost-code                       Update cost code ID (empty to clear)
  --is-revenue-quantity-driver      Update revenue quantity driver flag (true/false)
  --price-estimate                  Update price estimate ID (empty to clear)
  --quantity-estimate               Update quantity estimate ID (empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update revenue quantity driver
  xbe do project-phase-cost-items update 123 --is-revenue-quantity-driver true

  # Update classification
  xbe do project-phase-cost-items update 123 --project-cost-classification 456

  # Clear cost code
  xbe do project-phase-cost-items update 123 --cost-code ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectPhaseCostItemsUpdate,
	}
	initDoProjectPhaseCostItemsUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseCostItemsCmd.AddCommand(newDoProjectPhaseCostItemsUpdateCmd())
}

func initDoProjectPhaseCostItemsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-cost-classification", "", "Project cost classification ID")
	cmd.Flags().String("project-resource-classification", "", "Project resource classification ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("cost-code", "", "Cost code ID")
	cmd.Flags().String("is-revenue-quantity-driver", "", "Revenue quantity driver flag (true/false)")
	cmd.Flags().String("price-estimate", "", "Price estimate ID")
	cmd.Flags().String("quantity-estimate", "", "Quantity estimate ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseCostItemsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectPhaseCostItemsUpdateOptions(cmd, args)
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
	relationships := map[string]any{}
	hasChanges := false

	if opts.IsRevenueQuantityDriver != "" {
		attributes["is-revenue-quantity-driver"] = opts.IsRevenueQuantityDriver == "true"
		hasChanges = true
	}

	if cmd.Flags().Changed("project-cost-classification") {
		if opts.ProjectCostClassification == "" {
			relationships["project-cost-classification"] = map[string]any{"data": nil}
		} else {
			relationships["project-cost-classification"] = map[string]any{
				"data": map[string]any{
					"type": "project-cost-classifications",
					"id":   opts.ProjectCostClassification,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("project-resource-classification") {
		if opts.ProjectResourceClassification == "" {
			relationships["project-resource-classification"] = map[string]any{"data": nil}
		} else {
			relationships["project-resource-classification"] = map[string]any{
				"data": map[string]any{
					"type": "project-resource-classifications",
					"id":   opts.ProjectResourceClassification,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("unit-of-measure") {
		if opts.UnitOfMeasure == "" {
			relationships["unit-of-measure"] = map[string]any{"data": nil}
		} else {
			relationships["unit-of-measure"] = map[string]any{
				"data": map[string]any{
					"type": "unit-of-measures",
					"id":   opts.UnitOfMeasure,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("cost-code") {
		if opts.CostCode == "" {
			relationships["cost-code"] = map[string]any{"data": nil}
		} else {
			relationships["cost-code"] = map[string]any{
				"data": map[string]any{
					"type": "cost-codes",
					"id":   opts.CostCode,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("price-estimate") {
		if opts.PriceEstimate == "" {
			relationships["price-estimate"] = map[string]any{"data": nil}
		} else {
			relationships["price-estimate"] = map[string]any{
				"data": map[string]any{
					"type": "project-phase-cost-item-price-estimates",
					"id":   opts.PriceEstimate,
				},
			}
		}
		hasChanges = true
	}

	if cmd.Flags().Changed("quantity-estimate") {
		if opts.QuantityEstimate == "" {
			relationships["quantity-estimate"] = map[string]any{"data": nil}
		} else {
			relationships["quantity-estimate"] = map[string]any{
				"data": map[string]any{
					"type": "project-phase-cost-item-quantity-estimates",
					"id":   opts.QuantityEstimate,
				},
			}
		}
		hasChanges = true
	}

	if !hasChanges {
		err := fmt.Errorf("no fields to update; specify at least one flag")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-phase-cost-items",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-phase-cost-items/"+opts.ID, jsonBody)
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

	row := buildProjectPhaseCostItemRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project phase cost item %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseCostItemsUpdateOptions(cmd *cobra.Command, args []string) (doProjectPhaseCostItemsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	projectResourceClassification, _ := cmd.Flags().GetString("project-resource-classification")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	costCode, _ := cmd.Flags().GetString("cost-code")
	isRevenueQuantityDriver, _ := cmd.Flags().GetString("is-revenue-quantity-driver")
	priceEstimate, _ := cmd.Flags().GetString("price-estimate")
	quantityEstimate, _ := cmd.Flags().GetString("quantity-estimate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseCostItemsUpdateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ID:                            args[0],
		ProjectCostClassification:     projectCostClassification,
		ProjectResourceClassification: projectResourceClassification,
		UnitOfMeasure:                 unitOfMeasure,
		CostCode:                      costCode,
		IsRevenueQuantityDriver:       isRevenueQuantityDriver,
		PriceEstimate:                 priceEstimate,
		QuantityEstimate:              quantityEstimate,
	}, nil
}
