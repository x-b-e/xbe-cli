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

type doDriverAssignmentRulesUpdateOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	ID       string
	Rule     string
	IsActive bool
}

func newDoDriverAssignmentRulesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a driver assignment rule",
		Long: `Update a driver assignment rule.

Note: The level relationship cannot be changed after creation.

Optional flags:
  --rule        Rule text
  --is-active   Whether the rule is active

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update rule text
  xbe do driver-assignment-rules update 123 --rule "Updated rule"

  # Disable a rule
  xbe do driver-assignment-rules update 123 --is-active=false`,
		Args: cobra.ExactArgs(1),
		RunE: runDoDriverAssignmentRulesUpdate,
	}
	initDoDriverAssignmentRulesUpdateFlags(cmd)
	return cmd
}

func init() {
	doDriverAssignmentRulesCmd.AddCommand(newDoDriverAssignmentRulesUpdateCmd())
}

func initDoDriverAssignmentRulesUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rule", "", "Rule text")
	cmd.Flags().Bool("is-active", false, "Whether the rule is active")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverAssignmentRulesUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoDriverAssignmentRulesUpdateOptions(cmd, args)
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
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no attributes to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "driver-assignment-rules",
		"id":         opts.ID,
		"attributes": attributes,
	}

	requestBody := map[string]any{"data": data}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/driver-assignment-rules/"+opts.ID, jsonBody)
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

	row := buildDriverAssignmentRuleRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated driver assignment rule %s\n", row.ID)
	return nil
}

func parseDoDriverAssignmentRulesUpdateOptions(cmd *cobra.Command, args []string) (doDriverAssignmentRulesUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rule, _ := cmd.Flags().GetString("rule")
	isActive, _ := cmd.Flags().GetBool("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverAssignmentRulesUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		ID:       args[0],
		Rule:     rule,
		IsActive: isActive,
	}, nil
}
