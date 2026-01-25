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

type doJobProductionPlanCostCodesCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	JobProductionPlanID           string
	CostCodeID                    string
	ProjectResourceClassification string
}

func newDoJobProductionPlanCostCodesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan cost code",
		Long: `Create a job production plan cost code.

Required:
  --job-production-plan           Job production plan ID
  --cost-code                     Cost code ID

Optional:
  --project-resource-classification Project resource classification ID`,
		Example: `  # Create a job production plan cost code
  xbe do job-production-plan-cost-codes create --job-production-plan 123 --cost-code 456

  # Create with project resource classification
  xbe do job-production-plan-cost-codes create --job-production-plan 123 --cost-code 456 --project-resource-classification 789`,
		RunE: runDoJobProductionPlanCostCodesCreate,
	}
	initDoJobProductionPlanCostCodesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanCostCodesCmd.AddCommand(newDoJobProductionPlanCostCodesCreateCmd())
}

func initDoJobProductionPlanCostCodesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("cost-code", "", "Cost code ID")
	cmd.Flags().String("project-resource-classification", "", "Project resource classification ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("job-production-plan")
	_ = cmd.MarkFlagRequired("cost-code")
}

func runDoJobProductionPlanCostCodesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanCostCodesCreateOptions(cmd)
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

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlanID,
			},
		},
		"cost-code": map[string]any{
			"data": map[string]any{
				"type": "cost-codes",
				"id":   opts.CostCodeID,
			},
		},
	}

	if opts.ProjectResourceClassification != "" {
		relationships["project-resource-classification"] = map[string]any{
			"data": map[string]any{
				"type": "project-resource-classifications",
				"id":   opts.ProjectResourceClassification,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-cost-codes",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-cost-codes", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan cost code %s\n", resp.Data.ID)
	return nil
}

func parseDoJobProductionPlanCostCodesCreateOptions(cmd *cobra.Command) (doJobProductionPlanCostCodesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	costCodeID, _ := cmd.Flags().GetString("cost-code")
	projectResourceClassification, _ := cmd.Flags().GetString("project-resource-classification")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanCostCodesCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		JobProductionPlanID:           jobProductionPlanID,
		CostCodeID:                    costCodeID,
		ProjectResourceClassification: projectResourceClassification,
	}, nil
}
