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

type doJobProductionPlanStatusChangesUpdateOptions struct {
	BaseURL                                 string
	Token                                   string
	JSON                                    bool
	ID                                      string
	JobProductionPlanCancellationReasonType string
}

func newDoJobProductionPlanStatusChangesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a job production plan status change",
		Long: `Update a job production plan status change.

Arguments:
  <id>  The status change ID (required)

Optional flags:
  --job-production-plan-cancellation-reason-type  Set cancellation reason type ID (use empty string to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Set a cancellation reason type
  xbe do job-production-plan-status-changes update 123 --job-production-plan-cancellation-reason-type 456

  # Clear a cancellation reason type
  xbe do job-production-plan-status-changes update 123 --job-production-plan-cancellation-reason-type ""`,
		Args: cobra.ExactArgs(1),
		RunE: runDoJobProductionPlanStatusChangesUpdate,
	}
	initDoJobProductionPlanStatusChangesUpdateFlags(cmd)
	return cmd
}

func init() {
	doJobProductionPlanStatusChangesCmd.AddCommand(newDoJobProductionPlanStatusChangesUpdateCmd())
}

func initDoJobProductionPlanStatusChangesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("job-production-plan-cancellation-reason-type", "", "Set cancellation reason type ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoJobProductionPlanStatusChangesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoJobProductionPlanStatusChangesUpdateOptions(cmd, args)
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
	relationships := map[string]any{}

	if cmd.Flags().Changed("job-production-plan-cancellation-reason-type") {
		reasonID := strings.TrimSpace(opts.JobProductionPlanCancellationReasonType)
		if reasonID == "" {
			relationships["job-production-plan-cancellation-reason-type"] = map[string]any{"data": nil}
		} else {
			relationships["job-production-plan-cancellation-reason-type"] = map[string]any{
				"data": map[string]any{
					"type": "job-production-plan-cancellation-reason-types",
					"id":   reasonID,
				},
			}
		}
	}

	if len(attributes) == 0 && len(relationships) == 0 {
		err := fmt.Errorf("no attributes or relationships to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type": "job-production-plan-status-changes",
		"id":   opts.ID,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}
	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/job-production-plan-status-changes/"+opts.ID, jsonBody)
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

	row := buildJobProductionPlanStatusChangeRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated job production plan status change %s\n", row.ID)
	return nil
}

func parseDoJobProductionPlanStatusChangesUpdateOptions(cmd *cobra.Command, args []string) (doJobProductionPlanStatusChangesUpdateOptions, error) {
	id := strings.TrimSpace(args[0])
	if id == "" {
		return doJobProductionPlanStatusChangesUpdateOptions{}, fmt.Errorf("status change id is required")
	}

	jsonOut, _ := cmd.Flags().GetBool("json")
	reasonType, _ := cmd.Flags().GetString("job-production-plan-cancellation-reason-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doJobProductionPlanStatusChangesUpdateOptions{
		BaseURL:                                 baseURL,
		Token:                                   token,
		JSON:                                    jsonOut,
		ID:                                      id,
		JobProductionPlanCancellationReasonType: reasonType,
	}, nil
}
