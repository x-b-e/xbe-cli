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

type doJobProductionPlanUnapprovalsCreateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	JobProductionPlan                 string
	Comment                           string
	SuppressStatusChangeNotifications bool
}

type jobProductionPlanUnapprovalRow struct {
	ID                                string `json:"id"`
	JobProductionPlanID               string `json:"job_production_plan_id,omitempty"`
	Comment                           string `json:"comment,omitempty"`
	SuppressStatusChangeNotifications bool   `json:"suppress_status_change_notifications"`
}

func newDoJobProductionPlanUnapprovalsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unapprove a job production plan",
		Long: `Unapprove a job production plan.

Required flags:
  --job-production-plan  Job production plan ID (required)

Optional flags:
  --comment                               Unapproval comment
  --suppress-status-change-notifications  Suppress status change notifications`,
		Example: `  # Unapprove a job production plan
  xbe do job-production-plan-unapprovals create --job-production-plan 123 --comment "Need revisions"

  # Unapprove and suppress notifications
  xbe do job-production-plan-unapprovals create --job-production-plan 123 --suppress-status-change-notifications`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanUnapprovalsCreate,
	}
	initDoJobProductionPlanUnapprovalsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanUnapprovalsCmd.AddCommand(newDoJobProductionPlanUnapprovalsCreateCmd())
}

func initDoJobProductionPlanUnapprovalsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("comment", "", "Unapproval comment")
	cmd.Flags().Bool("suppress-status-change-notifications", false, "Suppress status change notifications")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanUnapprovalsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanUnapprovalsCreateOptions(cmd)
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

	attributes := map[string]any{}
	if strings.TrimSpace(opts.Comment) != "" {
		attributes["comment"] = opts.Comment
	}
	if cmd.Flags().Changed("suppress-status-change-notifications") {
		attributes["suppress-status-change-notifications"] = opts.SuppressStatusChangeNotifications
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
			"type":          "job-production-plan-unapprovals",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-unapprovals", jsonBody)
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

	row := jobProductionPlanUnapprovalRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan unapproval %s\n", row.ID)
	return nil
}

func jobProductionPlanUnapprovalRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanUnapprovalRow {
	attrs := resp.Data.Attributes
	row := jobProductionPlanUnapprovalRow{
		ID:                                resp.Data.ID,
		Comment:                           stringAttr(attrs, "comment"),
		SuppressStatusChangeNotifications: boolAttr(attrs, "suppress-status-change-notifications"),
	}

	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}

	return row
}

func parseDoJobProductionPlanUnapprovalsCreateOptions(cmd *cobra.Command) (doJobProductionPlanUnapprovalsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	comment, _ := cmd.Flags().GetString("comment")
	suppressStatusChangeNotifications, _ := cmd.Flags().GetBool("suppress-status-change-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanUnapprovalsCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		JobProductionPlan:                 jobProductionPlan,
		Comment:                           comment,
		SuppressStatusChangeNotifications: suppressStatusChangeNotifications,
	}, nil
}
