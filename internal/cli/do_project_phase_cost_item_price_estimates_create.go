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

type doProjectPhaseCostItemPriceEstimatesCreateOptions struct {
	BaseURL              string
	Token                string
	JSON                 bool
	ProjectPhaseCostItem string
	ProjectEstimateSet   string
	CreatedBy            string
	Estimate             string
}

type projectPhaseCostItemPriceEstimateRowCreate struct {
	ID                     string `json:"id"`
	ProjectPhaseCostItemID string `json:"project_phase_cost_item_id,omitempty"`
	ProjectEstimateSetID   string `json:"project_estimate_set_id,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
	Estimate               any    `json:"estimate,omitempty"`
}

func newDoProjectPhaseCostItemPriceEstimatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase cost item price estimate",
		Long: `Create a project phase cost item price estimate.

Required flags:
  --project-phase-cost-item  Project phase cost item ID
  --project-estimate-set     Project estimate set ID

Optional flags:
  --estimate    Estimate JSON object describing a probability distribution
  --created-by  Created by user ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a normal distribution estimate
  xbe do project-phase-cost-item-price-estimates create \
    --project-phase-cost-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'

  # Create a triangular distribution estimate
  xbe do project-phase-cost-item-price-estimates create \
    --project-phase-cost-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"TriangularDistribution","minimum":5,"mode":10,"maximum":15}'`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseCostItemPriceEstimatesCreate,
	}
	initDoProjectPhaseCostItemPriceEstimatesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseCostItemPriceEstimatesCmd.AddCommand(newDoProjectPhaseCostItemPriceEstimatesCreateCmd())
}

func initDoProjectPhaseCostItemPriceEstimatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-phase-cost-item", "", "Project phase cost item ID (required)")
	cmd.Flags().String("project-estimate-set", "", "Project estimate set ID (required)")
	cmd.Flags().String("estimate", "", "Estimate JSON object (optional)")
	cmd.Flags().String("created-by", "", "Created by user ID (optional)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("project-phase-cost-item")
	cmd.MarkFlagRequired("project-estimate-set")
}

func runDoProjectPhaseCostItemPriceEstimatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseCostItemPriceEstimatesCreateOptions(cmd)
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

	projectPhaseCostItem := strings.TrimSpace(opts.ProjectPhaseCostItem)
	if projectPhaseCostItem == "" {
		err := fmt.Errorf("--project-phase-cost-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	projectEstimateSet := strings.TrimSpace(opts.ProjectEstimateSet)
	if projectEstimateSet == "" {
		err := fmt.Errorf("--project-estimate-set is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("estimate") {
		estimate, err := parseEstimateInput(opts.Estimate)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if estimate != nil || strings.TrimSpace(opts.Estimate) == "null" {
			attributes["estimate"] = estimate
		}
	}

	relationships := map[string]any{
		"project-phase-cost-item": map[string]any{
			"data": map[string]any{
				"type": "project-phase-cost-items",
				"id":   projectPhaseCostItem,
			},
		},
		"project-estimate-set": map[string]any{
			"data": map[string]any{
				"type": "project-estimate-sets",
				"id":   projectEstimateSet,
			},
		},
	}
	if strings.TrimSpace(opts.CreatedBy) != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   strings.TrimSpace(opts.CreatedBy),
			},
		}
	}

	data := map[string]any{
		"type":          "project-phase-cost-item-price-estimates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-cost-item-price-estimates", jsonBody)
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

	row := buildProjectPhaseCostItemPriceEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase cost item price estimate %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseCostItemPriceEstimatesCreateOptions(cmd *cobra.Command) (doProjectPhaseCostItemPriceEstimatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectPhaseCostItem, _ := cmd.Flags().GetString("project-phase-cost-item")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	estimate, _ := cmd.Flags().GetString("estimate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseCostItemPriceEstimatesCreateOptions{
		BaseURL:              baseURL,
		Token:                token,
		JSON:                 jsonOut,
		ProjectPhaseCostItem: projectPhaseCostItem,
		ProjectEstimateSet:   projectEstimateSet,
		CreatedBy:            createdBy,
		Estimate:             estimate,
	}, nil
}

func buildProjectPhaseCostItemPriceEstimateRowFromSingle(resp jsonAPISingleResponse) projectPhaseCostItemPriceEstimateRowCreate {
	resource := resp.Data
	row := projectPhaseCostItemPriceEstimateRowCreate{
		ID:                     resource.ID,
		Estimate:               resource.Attributes["estimate"],
		ProjectPhaseCostItemID: relationshipIDFromMap(resource.Relationships, "project-phase-cost-item"),
		ProjectEstimateSetID:   relationshipIDFromMap(resource.Relationships, "project-estimate-set"),
		CreatedByID:            relationshipIDFromMap(resource.Relationships, "created-by"),
	}
	return row
}
