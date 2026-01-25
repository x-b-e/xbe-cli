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

type doJobProductionPlanBroadcastMessagesCreateOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	JobProductionPlan string
	Message           string
	Summary           string
	UserIDs           []string
}

func newDoJobProductionPlanBroadcastMessagesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan broadcast message",
		Long: `Create a job production plan broadcast message.

Required flags:
  --job-production-plan  Job production plan ID (required)
  --message              Message body (required)

Optional flags:
  --summary              Short summary (max 100 characters)
  --user-ids             Recipient user IDs (comma-separated or repeated)

Notes:
  If --user-ids is provided, values must be part of the default recipients
  for the job production plan.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a broadcast message
  xbe do job-production-plan-broadcast-messages create \
    --job-production-plan 123 \
    --message "Crew arrival moved to 7:30 AM" \
    --summary "Start time update"

  # Create with specific recipients
  xbe do job-production-plan-broadcast-messages create \
    --job-production-plan 123 \
    --message "Only drivers needed at 8:00 AM" \
    --user-ids 12,34`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanBroadcastMessagesCreate,
	}
	initDoJobProductionPlanBroadcastMessagesCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanBroadcastMessagesCmd.AddCommand(newDoJobProductionPlanBroadcastMessagesCreateCmd())
}

func initDoJobProductionPlanBroadcastMessagesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID (required)")
	cmd.Flags().String("message", "", "Message body (required)")
	cmd.Flags().String("summary", "", "Short summary (max 100 characters)")
	cmd.Flags().StringSlice("user-ids", nil, "Recipient user IDs (comma-separated or repeated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanBroadcastMessagesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanBroadcastMessagesCreateOptions(cmd)
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
	if opts.Message == "" {
		err := fmt.Errorf("--message is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"message": opts.Message,
	}
	if opts.Summary != "" {
		attributes["summary"] = opts.Summary
	}
	if len(opts.UserIDs) > 0 {
		attributes["user-ids"] = opts.UserIDs
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
			"type":          "job-production-plan-broadcast-messages",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-broadcast-messages", jsonBody)
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

	row := buildJobProductionPlanBroadcastMessageRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan broadcast message %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanBroadcastMessagesCreateOptions(cmd *cobra.Command) (doJobProductionPlanBroadcastMessagesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	message, _ := cmd.Flags().GetString("message")
	summary, _ := cmd.Flags().GetString("summary")
	userIDs, _ := cmd.Flags().GetStringSlice("user-ids")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanBroadcastMessagesCreateOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		JobProductionPlan: jobProductionPlan,
		Message:           message,
		Summary:           summary,
		UserIDs:           userIDs,
	}, nil
}

func buildJobProductionPlanBroadcastMessageRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanBroadcastMessageRow {
	included := make(map[string]jsonAPIResource)
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource
	}

	return buildJobProductionPlanBroadcastMessageRow(resp.Data, included)
}
