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

type doJobProductionPlanCompletionsCreateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	JobProductionPlanID               string
	Comment                           string
	SuppressStatusChangeNotifications bool
}

type jobProductionPlanCompletionRow struct {
	ID                                string `json:"id"`
	JobProductionPlanID               string `json:"job_production_plan_id,omitempty"`
	Comment                           string `json:"comment,omitempty"`
	SuppressStatusChangeNotifications bool   `json:"suppress_status_change_notifications"`
}

func newDoJobProductionPlanCompletionsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Complete a job production plan",
		Long: `Complete a job production plan.

This action transitions the plan status to complete. Only plans in approved
status can be completed.

Required flags:
  --job-production-plan   Job production plan ID

Optional flags:
  --comment                                 Comment for the completion
  --suppress-status-change-notifications    Suppress status change notifications`,
		Example: `  # Complete a job production plan
  xbe do job-production-plan-completions create --job-production-plan 123 --comment "Finished"

  # Complete and suppress notifications
  xbe do job-production-plan-completions create --job-production-plan 123 --suppress-status-change-notifications

  # JSON output
  xbe do job-production-plan-completions create --job-production-plan 123 --json`,
		RunE: runDoJobProductionPlanCompletionsCreate,
	}
	initDoJobProductionPlanCompletionsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanCompletionsCmd.AddCommand(newDoJobProductionPlanCompletionsCreateCmd())
}

func initDoJobProductionPlanCompletionsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("comment", "", "Comment for the completion")
	cmd.Flags().Bool("suppress-status-change-notifications", false, "Suppress status change notifications")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("job-production-plan")
}

func runDoJobProductionPlanCompletionsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanCompletionsCreateOptions(cmd)
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
				"id":   opts.JobProductionPlanID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-completions",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-completions", jsonBody)
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

	row := buildJobProductionPlanCompletionRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan completion %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanCompletionsCreateOptions(cmd *cobra.Command) (doJobProductionPlanCompletionsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanID, _ := cmd.Flags().GetString("job-production-plan")
	comment, _ := cmd.Flags().GetString("comment")
	suppressNotifications, _ := cmd.Flags().GetBool("suppress-status-change-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanCompletionsCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		JobProductionPlanID:               jobProductionPlanID,
		Comment:                           comment,
		SuppressStatusChangeNotifications: suppressNotifications,
	}, nil
}

func buildJobProductionPlanCompletionRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanCompletionRow {
	resource := resp.Data
	row := jobProductionPlanCompletionRow{
		ID:                                resource.ID,
		Comment:                           stringAttr(resource.Attributes, "comment"),
		SuppressStatusChangeNotifications: boolAttr(resource.Attributes, "suppress-status-change-notifications"),
	}
	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		row.JobProductionPlanID = rel.Data.ID
	}
	return row
}
