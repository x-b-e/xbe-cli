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

type doJobProductionPlanCancellationsCreateOptions struct {
	BaseURL                                 string
	Token                                   string
	JSON                                    bool
	JobProductionPlan                       string
	JobProductionPlanCancellationReasonType string
	Comment                                 string
	SuppressStatusChangeNotifications       bool
}

func newDoJobProductionPlanCancellationsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Cancel a job production plan",
		Long: `Cancel a job production plan.

Job production plans must be in approved status and cannot have time cards or
material transactions when being cancelled.

Required flags:
  --job-production-plan   Job production plan ID

Optional flags:
  --job-production-plan-cancellation-reason-type   Cancellation reason type ID
  --comment                                        Cancellation comment
  --suppress-status-change-notifications           Suppress status change notifications

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Cancel a job production plan with a reason and comment
  xbe do job-production-plan-cancellations create \
    --job-production-plan 123 \
    --job-production-plan-cancellation-reason-type 45 \
    --comment "Weather delay"

  # Cancel and suppress notifications
  xbe do job-production-plan-cancellations create \
    --job-production-plan 123 \
    --suppress-status-change-notifications`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanCancellationsCreate,
	}
	initDoJobProductionPlanCancellationsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanCancellationsCmd.AddCommand(newDoJobProductionPlanCancellationsCreateCmd())
}

func initDoJobProductionPlanCancellationsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("job-production-plan-cancellation-reason-type", "", "Cancellation reason type ID")
	cmd.Flags().String("comment", "", "Cancellation comment")
	cmd.Flags().Bool("suppress-status-change-notifications", false, "Suppress status change notifications")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanCancellationsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanCancellationsCreateOptions(cmd)
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
	if opts.JobProductionPlanCancellationReasonType != "" {
		relationships["job-production-plan-cancellation-reason-type"] = map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-cancellation-reason-types",
				"id":   opts.JobProductionPlanCancellationReasonType,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-cancellations",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-cancellations", jsonBody)
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

	row := buildJobProductionPlanCancellationRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan cancellation %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanCancellationsCreateOptions(cmd *cobra.Command) (doJobProductionPlanCancellationsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	cancellationType, _ := cmd.Flags().GetString("job-production-plan-cancellation-reason-type")
	comment, _ := cmd.Flags().GetString("comment")
	suppressStatusChangeNotifications, _ := cmd.Flags().GetBool("suppress-status-change-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanCancellationsCreateOptions{
		BaseURL:                                 baseURL,
		Token:                                   token,
		JSON:                                    jsonOut,
		JobProductionPlan:                       jobProductionPlan,
		JobProductionPlanCancellationReasonType: cancellationType,
		Comment:                                 comment,
		SuppressStatusChangeNotifications:       suppressStatusChangeNotifications,
	}, nil
}
