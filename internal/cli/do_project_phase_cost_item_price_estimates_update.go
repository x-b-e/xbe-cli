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

type doProjectPhaseCostItemPriceEstimatesUpdateOptions struct {
	BaseURL            string
	Token              string
	JSON               bool
	ID                 string
	Estimate           string
	ProjectEstimateSet string
	CreatedBy          string
}

func newDoProjectPhaseCostItemPriceEstimatesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project phase cost item price estimate",
		Long: `Update a project phase cost item price estimate.

Optional fields:
  --estimate              Estimate JSON object describing a probability distribution
  --project-estimate-set  Project estimate set ID
  --created-by            Created by user ID

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update estimate
  xbe do project-phase-cost-item-price-estimates update 123 \
    --estimate '{"class_name":"NormalDistribution","mean":12,"standard_deviation":2.5}'

  # Move to a different estimate set
  xbe do project-phase-cost-item-price-estimates update 123 --project-estimate-set 456`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectPhaseCostItemPriceEstimatesUpdate,
	}
	initDoProjectPhaseCostItemPriceEstimatesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectPhaseCostItemPriceEstimatesCmd.AddCommand(newDoProjectPhaseCostItemPriceEstimatesUpdateCmd())
}

func initDoProjectPhaseCostItemPriceEstimatesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("estimate", "", "Estimate JSON object")
	cmd.Flags().String("project-estimate-set", "", "Project estimate set ID")
	cmd.Flags().String("created-by", "", "Created by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectPhaseCostItemPriceEstimatesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectPhaseCostItemPriceEstimatesUpdateOptions(cmd, args)
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

	id := strings.TrimSpace(opts.ID)
	if id == "" {
		return fmt.Errorf("price estimate id is required")
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

	relationships := map[string]any{}
	if cmd.Flags().Changed("project-estimate-set") {
		value := strings.TrimSpace(opts.ProjectEstimateSet)
		if value == "" {
			err := fmt.Errorf("--project-estimate-set cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["project-estimate-set"] = map[string]any{
			"data": map[string]any{
				"type": "project-estimate-sets",
				"id":   value,
			},
		}
	}
	if cmd.Flags().Changed("created-by") {
		value := strings.TrimSpace(opts.CreatedBy)
		if value == "" {
			err := fmt.Errorf("--created-by cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["created-by"] = map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   value,
			},
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	payload := map[string]any{
		"type": "project-phase-cost-item-price-estimates",
		"id":   id,
	}
	if len(attributes) > 0 {
		payload["attributes"] = attributes
	}
	if len(relationships) > 0 {
		payload["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": payload,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/project-phase-cost-item-price-estimates/"+id, jsonBody)
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

	if opts.JSON {
		row := buildProjectPhaseCostItemPriceEstimateRowFromSingle(resp)
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project phase cost item price estimate %s\n", resp.Data.ID)
	return nil
}

func parseDoProjectPhaseCostItemPriceEstimatesUpdateOptions(cmd *cobra.Command, args []string) (doProjectPhaseCostItemPriceEstimatesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	estimate, _ := cmd.Flags().GetString("estimate")
	projectEstimateSet, _ := cmd.Flags().GetString("project-estimate-set")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectPhaseCostItemPriceEstimatesUpdateOptions{
		BaseURL:            baseURL,
		Token:              token,
		JSON:               jsonOut,
		ID:                 args[0],
		Estimate:           estimate,
		ProjectEstimateSet: projectEstimateSet,
		CreatedBy:          createdBy,
	}, nil
}
