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

type doJobProductionPlanUnabandonmentsCreateOptions struct {
	BaseURL                           string
	Token                             string
	JSON                              bool
	JobProductionPlan                 string
	Comment                           string
	SuppressStatusChangeNotifications bool
}

func newDoJobProductionPlanUnabandonmentsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Unabandon a job production plan",
		Long: `Unabandon a job production plan.

Job production plans must be in abandoned status to be unabandoned.

Required flags:
  --job-production-plan   Job production plan ID

Optional flags:
  --comment                               Unabandonment comment
  --suppress-status-change-notifications  Suppress status change notifications

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Unabandon a job production plan with a comment
  xbe do job-production-plan-unabandonments create \
    --job-production-plan 123 \
    --comment "Restoring plan"

  # Unabandon and suppress notifications
  xbe do job-production-plan-unabandonments create \
    --job-production-plan 123 \
    --suppress-status-change-notifications`,
		Args: cobra.NoArgs,
		RunE: runDoJobProductionPlanUnabandonmentsCreate,
	}
	initDoJobProductionPlanUnabandonmentsCreateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanUnabandonmentsCmd.AddCommand(newDoJobProductionPlanUnabandonmentsCreateCmd())
}

func initDoJobProductionPlanUnabandonmentsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan", "", "Job production plan ID")
	cmd.Flags().String("comment", "", "Unabandonment comment")
	cmd.Flags().Bool("suppress-status-change-notifications", false, "Suppress status change notifications")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanUnabandonmentsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoJobProductionPlanUnabandonmentsCreateOptions(cmd)
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

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "job-production-plan-unabandonments",
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

	body, _, err := client.Post(cmd.Context(), "/v1/job-production-plan-unabandonments", jsonBody)
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

	row := buildJobProductionPlanUnabandonmentRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created job production plan unabandonment %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanUnabandonmentsCreateOptions(cmd *cobra.Command) (doJobProductionPlanUnabandonmentsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	comment, _ := cmd.Flags().GetString("comment")
	suppressStatusChangeNotifications, _ := cmd.Flags().GetBool("suppress-status-change-notifications")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanUnabandonmentsCreateOptions{
		BaseURL:                           baseURL,
		Token:                             token,
		JSON:                              jsonOut,
		JobProductionPlan:                 jobProductionPlan,
		Comment:                           comment,
		SuppressStatusChangeNotifications: suppressStatusChangeNotifications,
	}, nil
}
