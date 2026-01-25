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

type doProjectTransportPlansUpdateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	ID                            string
	Status                        string
	SkipActualization             bool
	SkipAssignmentRulesValidation bool
	AssignmentRuleOverrideReason  string
}

func newDoProjectTransportPlansUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan",
		Long: `Update a project transport plan.

Optional:
  --status                         Plan status (editing/approved/complete/cancelled)
  --skip-actualization             Skip actualization logic
  --skip-assignment-rules-validation Skip assignment rules validation
  --assignment-rule-override-reason  Override reason for assignment rules`,
		Example: `  # Update project transport plan status
  xbe do project-transport-plans update 123 --status approved

  # Update assignment rules metadata
  xbe do project-transport-plans update 123 \\
    --skip-assignment-rules-validation=false \\
    --assignment-rule-override-reason "Reviewed"`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlansUpdate,
	}
	initDoProjectTransportPlansUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlansCmd.AddCommand(newDoProjectTransportPlansUpdateCmd())
}

func initDoProjectTransportPlansUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("status", "", "Plan status")
	cmd.Flags().Bool("skip-actualization", false, "Skip actualization logic")
	cmd.Flags().Bool("skip-assignment-rules-validation", false, "Skip assignment rules validation")
	cmd.Flags().String("assignment-rule-override-reason", "", "Assignment rule override reason")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlansUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlansUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("skip-actualization") {
		attributes["skip-actualization"] = opts.SkipActualization
	}
	if cmd.Flags().Changed("skip-assignment-rules-validation") {
		attributes["skip-assignment-rules-validation"] = opts.SkipAssignmentRulesValidation
	}
	if cmd.Flags().Changed("assignment-rule-override-reason") {
		attributes["assignment-rule-override-reason"] = opts.AssignmentRuleOverrideReason
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify at least one field")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "project-transport-plans",
		"id":         opts.ID,
		"attributes": attributes,
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

	body, _, err := client.Patch(cmd.Context(), "/v1/project-transport-plans/"+opts.ID, jsonBody)
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

	row := projectTransportPlanRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlansUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlansUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	status, _ := cmd.Flags().GetString("status")
	skipActualization, _ := cmd.Flags().GetBool("skip-actualization")
	skipAssignmentRulesValidation, _ := cmd.Flags().GetBool("skip-assignment-rules-validation")
	assignmentRuleOverrideReason, _ := cmd.Flags().GetString("assignment-rule-override-reason")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlansUpdateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		ID:                            args[0],
		Status:                        status,
		SkipActualization:             skipActualization,
		SkipAssignmentRulesValidation: skipAssignmentRulesValidation,
		AssignmentRuleOverrideReason:  assignmentRuleOverrideReason,
	}, nil
}
