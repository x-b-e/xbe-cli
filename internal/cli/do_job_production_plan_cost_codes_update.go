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

type doJobProductionPlanCostCodesUpdateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ID                            string
	ProjectResourceClassification string
}

func newDoJobProductionPlanCostCodesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan cost code",
		Long: `Update a job production plan cost code.

Optional:
  --project-resource-classification  Project resource classification ID`,
		Example: `  # Update project resource classification
  xbe do job-production-plan-cost-codes update 123 --project-resource-classification 789`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanCostCodesUpdate,
	}
	initDoJobProductionPlanCostCodesUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanCostCodesCmd.AddCommand(newDoJobProductionPlanCostCodesUpdateCmd())
}

func initDoJobProductionPlanCostCodesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project-resource-classification", "", "Project resource classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanCostCodesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanCostCodesUpdateOptions(cmd, args)
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

	data := map[string]any{
		"type": "job-production-plan-cost-codes",
		"id":   opts.ID,
	}

	relationships := map[string]any{}
	if cmd.Flags().Changed("project-resource-classification") {
		relationships["project-resource-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-resource-classifications",
				"id":   opts.ProjectResourceClassification,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data["relationships"] = relationships

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-cost-codes/"+opts.ID, jsonBody)
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
		row := jobProductionPlanCostCodeRow{ID: resp.Data.ID}
		if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
			row.JobProductionPlanID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["cost-code"]; ok && rel.Data != nil {
			row.CostCodeID = rel.Data.ID
		}
		if rel, ok := resp.Data.Relationships["project-resource-classification"]; ok && rel.Data != nil {
			row.ProjectResourceClassification = rel.Data.ID
		}
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan cost code %s\n", resp.Data.ID)
	return nil
}

func parseDoJobProductionPlanCostCodesUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanCostCodesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	projectResourceClassification, _ := cmd.Flags().GetString("project-resource-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanCostCodesUpdateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ID:                            args[0],
		ProjectResourceClassification: projectResourceClassification,
	}, nil
}
