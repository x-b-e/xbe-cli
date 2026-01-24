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

type doProjectTransportPlanAssignmentRulesUpdateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	ID        string
	Rule      string
	AssetType string
	IsActive  bool
}

func newDoProjectTransportPlanAssignmentRulesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project transport plan assignment rule",
		Long: `Update an existing project transport plan assignment rule.

Optional flags:
  --rule        Rule text
  --asset-type  Asset type (driver/tractor/trailer)
  --is-active   Whether the rule is active

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update rule text
  xbe do project-transport-plan-assignment-rules update 123 --rule "Updated rule"

  # Deactivate a rule
  xbe do project-transport-plan-assignment-rules update 123 --is-active=false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoProjectTransportPlanAssignmentRulesUpdate,
	}
	initDoProjectTransportPlanAssignmentRulesUpdateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanAssignmentRulesCmd.AddCommand(newDoProjectTransportPlanAssignmentRulesUpdateCmd())
}

func initDoProjectTransportPlanAssignmentRulesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rule", "", "Rule text")
	cmd.Flags().String("asset-type", "", "Asset type (driver/tractor/trailer)")
	cmd.Flags().Bool("is-active", false, "Whether the rule is active")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoProjectTransportPlanAssignmentRulesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoProjectTransportPlanAssignmentRulesUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("rule") {
		attributes["rule"] = opts.Rule
	}
	if cmd.Flags().Changed("asset-type") {
		attributes["asset-type"] = opts.AssetType
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "project-transport-plan-assignment-rules",
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

	path := fmt.Sprintf("/v1/project-transport-plan-assignment-rules/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	row := buildProjectTransportPlanAssignmentRuleRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated project transport plan assignment rule %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanAssignmentRulesUpdateOptions(cmd *cobra.Command, args []string) (doProjectTransportPlanAssignmentRulesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rule, _ := cmd.Flags().GetString("rule")
	assetType, _ := cmd.Flags().GetString("asset-type")
	isActive, _ := cmd.Flags().GetBool("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanAssignmentRulesUpdateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		ID:        args[0],
		Rule:      rule,
		AssetType: assetType,
		IsActive:  isActive,
	}, nil
}
