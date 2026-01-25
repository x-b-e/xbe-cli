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

type doJobProductionPlanAlarmSubscribersCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	JobProductionPlanAlarm string
	Subscriber             string
}

func newDoJobProductionPlanAlarmSubscribersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a job production plan alarm subscriber",
		Long: `Create a job production plan alarm subscriber.

Required flags:
  --job-production-plan-alarm  Job production plan alarm ID (required)
  --subscriber                 Subscriber user ID (required)`,
		Example: `  # Subscribe a user to an alarm
  xbe do job-production-plan-alarm-subscribers create --job-production-plan-alarm 123 --subscriber 456`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanAlarmSubscribersCreate,
	}
	initDoJobProductionPlanAlarmSubscribersCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanAlarmSubscribersCmd.AddCommand(newDoJobProductionPlanAlarmSubscribersCreateCmd())
}

func initDoJobProductionPlanAlarmSubscribersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan-alarm", "", "Job production plan alarm ID (required)")
	cmd.Flags().String("subscriber", "", "Subscriber user ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanAlarmSubscribersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanAlarmSubscribersCreateOptions(cmd)
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

	if opts.JobProductionPlanAlarm == "" {
		err := fmt.Errorf("--job-production-plan-alarm is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if opts.Subscriber == "" {
		err := fmt.Errorf("--subscriber is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"job-production-plan-alarm": map[string]any{
			"data": map[string]any{
				"type": "job-production-plan-alarms",
				"id":   opts.JobProductionPlanAlarm,
			},
		},
		"subscriber": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   opts.Subscriber,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-alarm-subscribers",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-alarm-subscribers", jsonBody)
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

	row := jobProductionPlanAlarmSubscriberRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan alarm subscriber %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanAlarmSubscribersCreateOptions(cmd *cobra.Command) (doJobProductionPlanAlarmSubscribersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlanAlarm, _ := cmd.Flags().GetString("job-production-plan-alarm")
	subscriber, _ := cmd.Flags().GetString("subscriber")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanAlarmSubscribersCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		JobProductionPlanAlarm: jobProductionPlanAlarm,
		Subscriber:             subscriber,
	}, nil
}
