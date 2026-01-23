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

type doJobProductionPlanUncompletionsCreateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	JobProductionPlan                 string
	Comment                           string
	SuppressStatusChangeNotifications bool
}

func newDoJobProductionPlanUncompletionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Uncomplete a job production plan",
		Long: `Uncomplete a completed job production plan.

Required flags:
  --job-production-plan  Job production plan ID to uncomplete

Optional flags:
  --comment                               Uncompletion comment
  --suppress-status-change-notifications  Suppress status change notifications`,
		Example: `  # Uncomplete a job production plan
  xbe do job-production-plan-uncompletions create --job-production-plan 12345

  # Uncomplete with comment
  xbe do job-production-plan-uncompletions create --job-production-plan 12345 --comment "Reopened"

  # Uncomplete without notifications
  xbe do job-production-plan-uncompletions create \
    --job-production-plan 12345 \
    --suppress-status-change-notifications

  # JSON output
  xbe do job-production-plan-uncompletions create --job-production-plan 12345 --json`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanUncompletionsCreate,
	}
	initDoJobProductionPlanUncompletionsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanUncompletionsCmd.AddCommand(newDoJobProductionPlanUncompletionsCreateCmd())
}

func initDoJobProductionPlanUncompletionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("comment", "", "Uncompletion comment")
	cmd.Flags().Bool("suppress-status-change-notifications", false, "Suppress status change notifications")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanUncompletionsCreate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanUncompletionsCreateOptions(cmd)
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
			"type":          "job-production-plan-uncompletions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-uncompletions", jsonBody)
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

	row := buildJobProductionPlanUncompletionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if row.JobProductionPlanID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan uncompletion %s for plan %s\n", row.ID, row.JobProductionPlanID)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan uncompletion %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanUncompletionsCreateOptions(cmd *cobra.Command) (doJobProductionPlanUncompletionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	comment, _ := cmd.Flags().GetString("comment")
	suppressStatusChangeNotifications, _ := cmd.Flags().GetBool("suppress-status-change-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanUncompletionsCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		JobProductionPlan:                 jobProductionPlan,
		Comment:                           comment,
		SuppressStatusChangeNotifications: suppressStatusChangeNotifications,
	}, nil
}

type jobProductionPlanUncompletionRow struct {
	ID                                string `json:"id"`
	JobProductionPlanID               string `json:"job_production_plan_id"`
	Comment                           string `json:"comment,omitempty"`
	SuppressStatusChangeNotifications bool   `json:"suppress_status_change_notifications"`
}

func buildJobProductionPlanUncompletionRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanUncompletionRow {
	attrs := resp.Data.Attributes
	row := jobProductionPlanUncompletionRow{
		ID:                                resp.Data.ID,
		Comment:                           stringAttr(attrs, "comment"),
		SuppressStatusChangeNotifications: boolAttr(attrs, "suppress-status-change-notifications"),
	}
	if rel, ok := resp.Data.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	return row
}
