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

type doProjectRevenueItemQuantityEstimatesCreateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ProjectRevenueItem string
	ProjectEstimateSet string
	CreatedBy          string
	Description        string
	Estimate           string
}

func newDoProjectRevenueItemQuantityEstimatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project revenue item quantity estimate",
		Long: `Create a project revenue item quantity estimate.

Required:
  --project-revenue-item  Project revenue item ID
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
  xbe do project-revenue-item-quantity-estimates create \
    --project-revenue-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"NormalDistribution","mean":10,"standard_deviation":2}'

  # Create with a description
  xbe do project-revenue-item-quantity-estimates create \
    --project-revenue-item 123 \
    --project-estimate-set 456 \
    --description "Initial estimate"

  # Output as JSON
  xbe do project-revenue-item-quantity-estimates create \
    --project-revenue-item 123 \
    --project-estimate-set 456 \
    --estimate '{"class_name":"TriangularDistribution","minimum":5,"mode":7,"maximum":12}' \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoProjectRevenueItemQuantityEstimatesCreate,
	}
	initDoProjectRevenueItemQuantityEstimatesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectRevenueItemQuantityEstimatesCmd.AddCommand(newDoProjectRevenueItemQuantityEstimatesCreateCmd())
}

func initDoProjectRevenueItemQuantityEstimatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-revenue-item", "", "Project revenue item ID (required)")
	cmd.Flags().String("project-estimate-set", "", "Project estimate set ID (required)")
	cmd.Flags().String("estimate", "", "Estimate JSON payload")
	cmd.Flags().String("description", "", "Estimate description")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectRevenueItemQuantityEstimatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectRevenueItemQuantityEstimatesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.ProjectRevenueItem) == "" {
		err := fmt.Errorf("--project-revenue-item is required")
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
		"project-revenue-item": map[string]any{
			"data": map[string]any{
				"type": "project-revenue-items",
				"id":   opts.ProjectRevenueItem,
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
		"type":          "project-revenue-item-quantity-estimates",
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-revenue-item-quantity-estimates", jsonBody)
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

	row := projectRevenueItemQuantityEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created project revenue item quantity estimate %s\n", row.ID)
	return nil
}

func parseDoProjectRevenueItemQuantityEstimatesCreateOptions(cmd *cobra.Command) (doProjectRevenueItemQuantityEstimatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectRevenueItem, _ := cmd.Flags().GetString("project-revenue-item")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	estimate, _ := cmd.Flags().GetString("estimate")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectRevenueItemQuantityEstimatesCreateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ProjectRevenueItem: projectRevenueItem,
		ProjectEstimateSet: projectEstimateSet,
		CreatedBy:          createdBy,
		Description:        description,
		Estimate:           estimate,
	}, nil
}
