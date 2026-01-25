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

type doJobProductionPlanTimeCardApproversCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	JobProductionPlan string
	User              string
}

func newDoJobProductionPlanTimeCardApproversCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan time card approver",
		Long: `Create a job production plan time card approver.

Required flags:
  --job-production-plan  Job production plan ID (required)
  --user                 User ID (required)`,
		Example: `  # Create a time card approver for a job production plan
  xbe do job-production-plan-time-card-approvers create \
    --job-production-plan 123 \
    --user 456

  # JSON output
  xbe do job-production-plan-time-card-approvers create \
    --job-production-plan 123 \
    --user 456 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanTimeCardApproversCreate,
	}
	initDoJobProductionPlanTimeCardApproversCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanTimeCardApproversCmd.AddCommand(newDoJobProductionPlanTimeCardApproversCreateCmd())
}

func initDoJobProductionPlanTimeCardApproversCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanTimeCardApproversCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanTimeCardApproversCreateOptions(cmd)
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

	if strings.TrimSpace(opts.JobProductionPlan) == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.User) == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.User,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-time-card-approvers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-time-card-approvers", jsonBody)
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

	row := buildJobProductionPlanTimeCardApproverRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan time card approver %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanTimeCardApproversCreateOptions(cmd *cobra.Command) (doJobProductionPlanTimeCardApproversCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	user, _ := cmd.Flags().GetString("user")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanTimeCardApproversCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		JobProductionPlan: jobProductionPlan,
		User:              user,
	}, nil
}

func buildJobProductionPlanTimeCardApproverRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanTimeCardApproverRow {
	row := jobProductionPlanTimeCardApproverRow{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		row.UserID = rel.Data.ID
	}

	return row
}
