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

type doProjectPhaseCostItemsCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ProjectPhaseRevenueItem       string
	ProjectCostClassification     string
	ProjectResourceClassification string
	UnitOfMeasure                 string
	CostCode                      string
	IsRevenueQuantityDriver       string
}

func newDoProjectPhaseCostItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase cost item",
		Long: `Create a project phase cost item.

Required flags:
  --project-phase-revenue-item   Project phase revenue item ID
  --project-cost-classification  Project cost classification ID

Optional flags:
  --project-resource-classification  Project resource classification ID
  --unit-of-measure                  Unit of measure ID
  --cost-code                        Cost code ID
  --is-revenue-quantity-driver       Whether this item drives revenue quantity (true/false)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a project phase cost item
  xbe do project-phase-cost-items create \
    --project-phase-revenue-item 123 \
    --project-cost-classification 456

  # Create with optional relationships
  xbe do project-phase-cost-items create \
    --project-phase-revenue-item 123 \
    --project-cost-classification 456 \
    --project-resource-classification 789 \
    --unit-of-measure 321 \
    --cost-code 654 \
    --is-revenue-quantity-driver true`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseCostItemsCreate,
	}
	initDoProjectPhaseCostItemsCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseCostItemsCmd.AddCommand(newDoProjectPhaseCostItemsCreateCmd())
}

func initDoProjectPhaseCostItemsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-phase-revenue-item", "", "Project phase revenue item ID (required)")
	cmd.Flags().String("project-cost-classification", "", "Project cost classification ID (required)")
	cmd.Flags().String("project-resource-classification", "", "Project resource classification ID")
	cmd.Flags().String("unit-of-measure", "", "Unit of measure ID")
	cmd.Flags().String("cost-code", "", "Cost code ID")
	cmd.Flags().String("is-revenue-quantity-driver", "", "Revenue quantity driver flag (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project-phase-revenue-item")
	_ = cmd.MarkFlagRequired("project-cost-classification")
}

func runDoProjectPhaseCostItemsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseCostItemsCreateOptions(cmd)
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

	projectPhaseRevenueItem := strings.TrimSpace(opts.ProjectPhaseRevenueItem)
	if projectPhaseRevenueItem == "" {
		err := fmt.Errorf("--project-phase-revenue-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	projectCostClassification := strings.TrimSpace(opts.ProjectCostClassification)
	if projectCostClassification == "" {
		err := fmt.Errorf("--project-cost-classification is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if opts.IsRevenueQuantityDriver != "" {
		attributes["is-revenue-quantity-driver"] = opts.IsRevenueQuantityDriver == "true"
	}

	relationships := map[string]any{
		"project-phase-revenue-item": map[string]any{
			"data": map[string]any{
				"type": "project-phase-revenue-items",
				"id":   projectPhaseRevenueItem,
			},
		},
		"project-cost-classification": map[string]any{
			"data": map[string]any{
				"type": "project-cost-classifications",
				"id":   projectCostClassification,
			},
		},
	}

	if strings.TrimSpace(opts.ProjectResourceClassification) != "" {
		relationships["project-resource-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-resource-classifications",
				"id":   strings.TrimSpace(opts.ProjectResourceClassification),
			},
		}
	}
	if strings.TrimSpace(opts.UnitOfMeasure) != "" {
		relationships["unit-of-measure"] = map[string]any{
			"data": map[string]any{
				"type": "unit-of-measures",
				"id":   strings.TrimSpace(opts.UnitOfMeasure),
			},
		}
	}
	if strings.TrimSpace(opts.CostCode) != "" {
		relationships["cost-code"] = map[string]any{
			"data": map[string]any{
				"type": "cost-codes",
				"id":   strings.TrimSpace(opts.CostCode),
			},
		}
	}

	data := map[string]any{
		"type":          "project-phase-cost-items",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-cost-items", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase cost item %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseCostItemsCreateOptions(cmd *cobra.Command) (doProjectPhaseCostItemsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectPhaseRevenueItem, _ := cmd.Flags().GetString("project-phase-revenue-item")
	projectCostClassification, _ := cmd.Flags().GetString("project-cost-classification")
	projectResourceClassification, _ := cmd.Flags().GetString("project-resource-classification")
	unitOfMeasure, _ := cmd.Flags().GetString("unit-of-measure")
	costCode, _ := cmd.Flags().GetString("cost-code")
	isRevenueQuantityDriver, _ := cmd.Flags().GetString("is-revenue-quantity-driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseCostItemsCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ProjectPhaseRevenueItem:       projectPhaseRevenueItem,
		ProjectCostClassification:     projectCostClassification,
		ProjectResourceClassification: projectResourceClassification,
		UnitOfMeasure:                 unitOfMeasure,
		CostCode:                      costCode,
		IsRevenueQuantityDriver:       isRevenueQuantityDriver,
	}, nil
}
