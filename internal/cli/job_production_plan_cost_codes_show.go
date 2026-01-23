package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanCostCodesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanCostCodeDetails struct {
	ID                            string `json:"id"`
	JobProductionPlanID           string `json:"job_production_plan_id,omitempty"`
	CostCodeID                    string `json:"cost_code_id,omitempty"`
	ProjectResourceClassification string `json:"project_resource_classification_id,omitempty"`
}

func newJobProductionPlanCostCodesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan cost code details",
		Long: `Show the full details of a job production plan cost code.

Output Fields:
  ID                 Job production plan cost code identifier
  Job Production Plan  Associated job production plan ID
  Cost Code          Associated cost code ID
  Resource Class     Project resource classification ID

Arguments:
  <id>    The job production plan cost code ID (required). You can find IDs using the list command.`,
		Example: `  # Show a job production plan cost code
  xbe view job-production-plan-cost-codes show 123

  # Get JSON output
  xbe view job-production-plan-cost-codes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanCostCodesShow,
	}
	initJobProductionPlanCostCodesShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanCostCodesCmd.AddCommand(newJobProductionPlanCostCodesShowCmd())
}

func initJobProductionPlanCostCodesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanCostCodesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanCostCodesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("job production plan cost code id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "job-production-plan,cost-code,project-resource-classification")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-cost-codes/"+id, query)
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

	details := buildJobProductionPlanCostCodeDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanCostCodeDetails(cmd, details)
}

func parseJobProductionPlanCostCodesShowOptions(cmd *cobra.Command) (jobProductionPlanCostCodesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanCostCodesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanCostCodeDetails(resp jsonAPISingleResponse) jobProductionPlanCostCodeDetails {
	resource := resp.Data
	details := jobProductionPlanCostCodeDetails{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["cost-code"]; ok && rel.Data != nil {
		details.CostCodeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["project-resource-classification"]; ok && rel.Data != nil {
		details.ProjectResourceClassification = rel.Data.ID
	}

	return details
}

func renderJobProductionPlanCostCodeDetails(cmd *cobra.Command, details jobProductionPlanCostCodeDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.CostCodeID != "" {
		fmt.Fprintf(out, "Cost Code: %s\n", details.CostCodeID)
	}
	if details.ProjectResourceClassification != "" {
		fmt.Fprintf(out, "Resource Class: %s\n", details.ProjectResourceClassification)
	}

	return nil
}
