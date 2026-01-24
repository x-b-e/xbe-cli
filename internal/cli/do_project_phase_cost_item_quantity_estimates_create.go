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

type doProjectPhaseCostItemQuantityEstimatesCreateOptions struct {
	BaseURL                  string
	Token                    string
	JSON                     bool
	ProjectPhaseCostItem     string
	ProjectEstimateSet       string
	CreatedBy                string
	RevenueItemQuantityBasis string
	Estimate                 string
}

func newDoProjectPhaseCostItemQuantityEstimatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase cost item quantity estimate",
		Long: `Create a project phase cost item quantity estimate.

Required:
  --project-phase-cost-item  Project phase cost item ID
  --project-estimate-set     Project estimate set ID

Optional:
  --estimate                     Estimate JSON payload
  --revenue-item-quantity-basis  Revenue item quantity basis (must be > 0)
  --created-by                   Creator user ID

Estimate JSON should include a class_name and distribution parameters.
Examples:
  NormalDistribution: {"class_name":"NormalDistribution","mean":10,"standard_deviation":2}
  TriangularDistribution: {"class_name":"TriangularDistribution","minimum":5,"mode":7,"maximum":12}`,
		Example: `  # Create with a normal distribution
  xbe do project-phase-cost-item-quantity-estimates create \
    --project-phase-cost-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'

  # Create with a quantity basis
  xbe do project-phase-cost-item-quantity-estimates create \
    --project-phase-cost-item 123 \
    --project-estimate-set 456 \
    --revenue-item-quantity-basis 15.5

  # Output as JSON
  xbe do project-phase-cost-item-quantity-estimates create \
    --project-phase-cost-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"TriangularDistribution","minimum":5,"mode":7,"maximum":12}' \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseCostItemQuantityEstimatesCreate,
	}
	initDoProjectPhaseCostItemQuantityEstimatesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseCostItemQuantityEstimatesCmd.AddCommand(newDoProjectPhaseCostItemQuantityEstimatesCreateCmd())
}

func initDoProjectPhaseCostItemQuantityEstimatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-phase-cost-item", "", "Project phase cost item ID (required)")
	cmd.Flags().String("project-estimate-set", "", "Project estimate set ID (required)")
	cmd.Flags().String("estimate", "", "Estimate JSON payload")
	cmd.Flags().String("revenue-item-quantity-basis", "", "Revenue item quantity basis")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseCostItemQuantityEstimatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseCostItemQuantityEstimatesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ProjectPhaseCostItem) == "" {
		err := fmt.Errorf("--project-phase-cost-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if strings.TrimSpace(opts.ProjectEstimateSet) == "" {
		err := fmt.Errorf("--project-estimate-set is required")
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
	if strings.TrimSpace(opts.RevenueItemQuantityBasis) != "" {
		attributes["revenue-item-quantity-basis"] = opts.RevenueItemQuantityBasis
	}
	if strings.TrimSpace(opts.Estimate) != "" {
		estimatePayload, err := parseEstimatePayload(opts.Estimate)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["estimate"] = estimatePayload
	}

	relationships := map[string]any{
		"project-phase-cost-item": map[string]any{
			"data": map[string]any{
				"type": "project-phase-cost-items",
				"id":   opts.ProjectPhaseCostItem,
			},
		},
		"project-estimate-set": map[string]any{
			"data": map[string]any{
				"type": "project-estimate-sets",
				"id":   opts.ProjectEstimateSet,
			},
		},
	}
	if strings.TrimSpace(opts.CreatedBy) != "" {
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	data := map[string]any{
		"type":          "project-phase-cost-item-quantity-estimates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-cost-item-quantity-estimates", jsonBody)
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

	row := projectPhaseCostItemQuantityEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase cost item quantity estimate %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseCostItemQuantityEstimatesCreateOptions(cmd *cobra.Command) (doProjectPhaseCostItemQuantityEstimatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectPhaseCostItem, _ := cmd.Flags().GetString("project-phase-cost-item")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	revenueItemQuantityBasis, _ := cmd.Flags().GetString("revenue-item-quantity-basis")
	estimate, _ := cmd.Flags().GetString("estimate")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseCostItemQuantityEstimatesCreateOptions{
		BaseURL:                  baseURL,
		Token:                    token,
		JSON:                     jsonOut,
		ProjectPhaseCostItem:     projectPhaseCostItem,
		ProjectEstimateSet:       projectEstimateSet,
		CreatedBy:                createdBy,
		RevenueItemQuantityBasis: revenueItemQuantityBasis,
		Estimate:                 estimate,
	}, nil
}

func parseEstimatePayload(raw string) (any, error) {
	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("invalid estimate JSON: %w", err)
	}
	if payload == nil {
		return nil, nil
	}
	if _, ok := payload.(map[string]any); !ok {
		return nil, fmt.Errorf("estimate must be a JSON object")
	}
	return payload, nil
}
