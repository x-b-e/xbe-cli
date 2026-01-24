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

type doProjectRevenueItemQuantityEstimatesUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	ProjectEstimateSet string
	CreatedBy          string
	Description        string
	Estimate           string
}

func newDoProjectRevenueItemQuantityEstimatesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project revenue item quantity estimate",
		Long: `Update a project revenue item quantity estimate.

All flags are optional. Only provided flags will be updated.

Attributes:
  --estimate     Estimate JSON payload (use null to clear)
  --description  Estimate description

Relationships:
  --project-estimate-set  Project estimate set ID
  --created-by            Creator user ID`,
		Example: `  # Update the estimate distribution
  xbe do project-revenue-item-quantity-estimates update 123 \
    --estimate '{"class_name":"TriangularDistribution","minimum":5,"mode":7,"maximum":12}'

  # Update the description
  xbe do project-revenue-item-quantity-estimates update 123 --description "Updated estimate"

  # Update the estimate set relationship
  xbe do project-revenue-item-quantity-estimates update 123 --project-estimate-set 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectRevenueItemQuantityEstimatesUpdate,
	}
	initDoProjectRevenueItemQuantityEstimatesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectRevenueItemQuantityEstimatesCmd.AddCommand(newDoProjectRevenueItemQuantityEstimatesUpdateCmd())
}

func initDoProjectRevenueItemQuantityEstimatesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-estimate-set", "", "Project estimate set ID")
	cmd.Flags().String("created-by", "", "Creator user ID")
	cmd.Flags().String("estimate", "", "Estimate JSON payload (use null to clear)")
	cmd.Flags().String("description", "", "Estimate description")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectRevenueItemQuantityEstimatesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectRevenueItemQuantityEstimatesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("description") {
		attributes["description"] = opts.Description
	}

	if cmd.Flags().Changed("estimate") {
		if strings.TrimSpace(opts.Estimate) == "" {
			err := fmt.Errorf("--estimate cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		estimatePayload, err := parseEstimatePayload(opts.Estimate)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["estimate"] = estimatePayload
	}

	if cmd.Flags().Changed("project-estimate-set") {
		if strings.TrimSpace(opts.ProjectEstimateSet) == "" {
			err := fmt.Errorf("--project-estimate-set cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["project-estimate-set"] = map[string]any{
			"data": map[string]any{
				"type": "project-estimate-sets",
				"id":   opts.ProjectEstimateSet,
			},
		}
	}

	if cmd.Flags().Changed("created-by") {
		if strings.TrimSpace(opts.CreatedBy) == "" {
			err := fmt.Errorf("--created-by cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.CreatedBy,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("at least one field must be specified for update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "project-revenue-item-quantity-estimates",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-revenue-item-quantity-estimates/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project revenue item quantity estimate %s\n", row.ID)
	return nil
}

func parseDoProjectRevenueItemQuantityEstimatesUpdateOptions(cmd *cobra.Command, args []string) (doProjectRevenueItemQuantityEstimatesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	estimate, _ := cmd.Flags().GetString("estimate")
	description, _ := cmd.Flags().GetString("description")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectRevenueItemQuantityEstimatesUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		ProjectEstimateSet: projectEstimateSet,
		CreatedBy:          createdBy,
		Description:        description,
		Estimate:           estimate,
	}, nil
}
