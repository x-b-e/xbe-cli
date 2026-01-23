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

type doJobProductionPlanApprovalsCreateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	JobProductionPlan                 string
	Comment                           string
	SuppressStatusChangeNotifications bool
	SkipValidateRequiredMixDesigns    bool
}

func newDoJobProductionPlanApprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Approve a job production plan",
		Long: `Approve a submitted job production plan.

Required flags:
  --job-production-plan  Job production plan ID to approve

Optional flags:
  --comment                               Approval comment
  --suppress-status-change-notifications  Suppress approval notifications
  --skip-validate-required-mix-designs    Skip required mix design validation`,
		Example: `  # Approve a job production plan
  xbe do job-production-plan-approvals create --job-production-plan 12345

  # Approve with comment
  xbe do job-production-plan-approvals create --job-production-plan 12345 --comment "Approved"

  # Approve while skipping mix design validation
  xbe do job-production-plan-approvals create \
    --job-production-plan 12345 \
    --skip-validate-required-mix-designs

  # JSON output
  xbe do job-production-plan-approvals create --job-production-plan 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanApprovalsCreate,
	}
	initDoJobProductionPlanApprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanApprovalsCmd.AddCommand(newDoJobProductionPlanApprovalsCreateCmd())
}

func initDoJobProductionPlanApprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("comment", "", "Approval comment")
	cmd.Flags().Bool("suppress-status-change-notifications", false, "Suppress status change notifications")
	cmd.Flags().Bool("skip-validate-required-mix-designs", false, "Skip required mix design validation")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanApprovalsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanApprovalsCreateOptions(cmd)
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

	if opts.JobProductionPlan == "" {
		err := fmt.Errorf("--job-production-plan is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}

	if opts.Comment != "" {
		attributes["comment"] = opts.Comment
	}
	if cmd.Flags().Changed("suppress-status-change-notifications") {
		attributes["suppress-status-change-notifications"] = opts.SuppressStatusChangeNotifications
	}
	if cmd.Flags().Changed("skip-validate-required-mix-designs") {
		attributes["skip-validate-required-mix-designs"] = opts.SkipValidateRequiredMixDesigns
	}

	relationships := map[string]any{
		"job-production-plan": map[string]any{
			"data": map[string]any{
				"type": "job-production-plans",
				"id":   opts.JobProductionPlan,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-approvals",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-approvals", jsonBody)
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

	row := buildJobProductionPlanApprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.JobProductionPlanID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan approval %s for plan %s\n", row.ID, row.JobProductionPlanID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan approval %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanApprovalsCreateOptions(cmd *cobra.Command) (doJobProductionPlanApprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	comment, _ := cmd.Flags().GetString("comment")
	suppressStatusChangeNotifications, _ := cmd.Flags().GetBool("suppress-status-change-notifications")
	skipValidateRequiredMixDesigns, _ := cmd.Flags().GetBool("skip-validate-required-mix-designs")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanApprovalsCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		JobProductionPlan:                 jobProductionPlan,
		Comment:                           comment,
		SuppressStatusChangeNotifications: suppressStatusChangeNotifications,
		SkipValidateRequiredMixDesigns:    skipValidateRequiredMixDesigns,
	}, nil
}

type jobProductionPlanApprovalRow struct {
	ID                                string `json:"id"`
	JobProductionPlanID               string `json:"job_production_plan_id"`
	Comment                           string `json:"comment,omitempty"`
	SuppressStatusChangeNotifications bool   `json:"suppress_status_change_notifications"`
	SkipValidateRequiredMixDesigns    bool   `json:"skip_validate_required_mix_designs"`
}

func buildJobProductionPlanApprovalRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanApprovalRow {
	attrs := resp.Data.Attributes
	row := jobProductionPlanApprovalRow{
		ID:                                resp.Data.ID,
		Comment:                           stringAttr(attrs, "comment"),
		SuppressStatusChangeNotifications: boolAttr(attrs, "suppress-status-change-notifications"),
		SkipValidateRequiredMixDesigns:    boolAttr(attrs, "skip-validate-required-mix-designs"),
	}
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	return row
}
