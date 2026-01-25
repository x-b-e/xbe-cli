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

type doProjectTransportPlansCreateOptions struct {
	BaseURL                       string
	Token                         string
	JSON                          bool
	Project                       string
	SkipActualization             bool
	SkipAssignmentRulesValidation bool
	AssignmentRuleOverrideReason  string
}

func newDoProjectTransportPlansCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan",
		Long: `Create a project transport plan.

Required:
  --project  Project ID

Optional:
  --skip-actualization              Skip actualization logic
  --skip-assignment-rules-validation Skip assignment rules validation
  --assignment-rule-override-reason  Override reason for assignment rules`,
		Example: `  # Create a project transport plan
  xbe do project-transport-plans create --project 123

  # Create with assignment override metadata
  xbe do project-transport-plans create \\
    --project 123 \\
    --skip-assignment-rules-validation \\
    --assignment-rule-override-reason "Manual override"`,
		RunE: runDoProjectTransportPlansCreate,
	}
	initDoProjectTransportPlansCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlansCmd.AddCommand(newDoProjectTransportPlansCreateCmd())
}

func initDoProjectTransportPlansCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().Bool("skip-actualization", false, "Skip actualization logic")
	cmd.Flags().Bool("skip-assignment-rules-validation", false, "Skip assignment rules validation")
	cmd.Flags().String("assignment-rule-override-reason", "", "Assignment rule override reason")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	_ = cmd.MarkFlagRequired("project")
}

func runDoProjectTransportPlansCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlansCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Project) == "" {
		err := fmt.Errorf("--project is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("skip-actualization") {
		attributes["skip-actualization"] = opts.SkipActualization
	}
	if cmd.Flags().Changed("skip-assignment-rules-validation") {
		attributes["skip-assignment-rules-validation"] = opts.SkipAssignmentRulesValidation
	}
	if cmd.Flags().Changed("assignment-rule-override-reason") {
		attributes["assignment-rule-override-reason"] = opts.AssignmentRuleOverrideReason
	}

	relationships := map[string]any{
		"project": map[string]any{
			"data": map[string]any{
				"type": "projects",
				"id":   opts.Project,
			},
		},
	}

	data := map[string]any{
		"type":          "project-transport-plans",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plans", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlansCreateOptions(cmd *cobra.Command) (doProjectTransportPlansCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	project, _ := cmd.Flags().GetString("project")
	skipActualization, _ := cmd.Flags().GetBool("skip-actualization")
	skipAssignmentRulesValidation, _ := cmd.Flags().GetBool("skip-assignment-rules-validation")
	assignmentRuleOverrideReason, _ := cmd.Flags().GetString("assignment-rule-override-reason")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlansCreateOptions{
		BaseURL:                       baseURL,
		Token:                         token,
		JSON:                          jsonOut,
		Project:                       project,
		SkipActualization:             skipActualization,
		SkipAssignmentRulesValidation: skipAssignmentRulesValidation,
		AssignmentRuleOverrideReason:  assignmentRuleOverrideReason,
	}, nil
}
