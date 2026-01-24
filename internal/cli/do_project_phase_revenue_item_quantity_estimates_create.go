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

type doProjectPhaseRevenueItemQuantityEstimatesCreateOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	ProjectPhaseRevenueItem string
	ProjectEstimateSet      string
	CreatedBy               string
	Description             string
	Estimate                string
}

func newDoProjectPhaseRevenueItemQuantityEstimatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project phase revenue item quantity estimate",
		Long: `Create a project phase revenue item quantity estimate.

Required:
  --project-phase-revenue-item  Project phase revenue item ID
  --project-estimate-set        Project estimate set ID

Optional:
  --estimate     Estimate JSON payload
  --description  Estimate description
  --created-by   Creator user ID

Estimate JSON should include a class_name and distribution parameters.
Examples:
  NormalDistribution: {"class_name":"NormalDistribution","mean":10,"standard_deviation":2}
  TriangularDistribution: {"class_name":"TriangularDistribution","minimum":5,"mode":7,"maximum":12}`,
		Example: `  # Create with a normal distribution
  xbe do project-phase-revenue-item-quantity-estimates create \
    --project-phase-revenue-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'

  # Create with a description
  xbe do project-phase-revenue-item-quantity-estimates create \
    --project-phase-revenue-item 123 \
    --project-estimate-set 456 \
    --description "Initial estimate"

  # Output as JSON
  xbe do project-phase-revenue-item-quantity-estimates create \
    --project-phase-revenue-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"TriangularDistribution","minimum":5,"mode":7,"maximum":12}' \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectPhaseRevenueItemQuantityEstimatesCreate,
	}
	initDoProjectPhaseRevenueItemQuantityEstimatesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseRevenueItemQuantityEstimatesCmd.AddCommand(newDoProjectPhaseRevenueItemQuantityEstimatesCreateCmd())
}

func initDoProjectPhaseRevenueItemQuantityEstimatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-phase-revenue-item", "", "Project phase revenue item ID (required)")
	cmd.Flags().String("project-estimate-set", "", "Project estimate set ID (required)")
	cmd.Flags().String("estimate", "", "Estimate JSON payload")
	cmd.Flags().String("description", "", "Estimate description")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseRevenueItemQuantityEstimatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectPhaseRevenueItemQuantityEstimatesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ProjectPhaseRevenueItem) == "" {
		err := fmt.Errorf("--project-phase-revenue-item is required")
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
	if strings.TrimSpace(opts.Description) != "" {
		attributes["description"] = opts.Description
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
		"project-phase-revenue-item": map[string]any{
			"data": map[string]any{
				"type": "project-phase-revenue-items",
				"id":   opts.ProjectPhaseRevenueItem,
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
		"type":          "project-phase-revenue-item-quantity-estimates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-phase-revenue-item-quantity-estimates", jsonBody)
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

	row := projectPhaseRevenueItemQuantityEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project phase revenue item quantity estimate %s\n", row.ID)
	return nil
}

func parseDoProjectPhaseRevenueItemQuantityEstimatesCreateOptions(cmd *cobra.Command) (doProjectPhaseRevenueItemQuantityEstimatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectPhaseRevenueItem, _ := cmd.Flags().GetString("project-phase-revenue-item")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	estimate, _ := cmd.Flags().GetString("estimate")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseRevenueItemQuantityEstimatesCreateOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		ProjectPhaseRevenueItem: projectPhaseRevenueItem,
		ProjectEstimateSet:      projectEstimateSet,
		CreatedBy:               createdBy,
		Description:             description,
		Estimate:                estimate,
	}, nil
}
