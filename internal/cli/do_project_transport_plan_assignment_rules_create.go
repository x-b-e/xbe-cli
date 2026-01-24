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

type doProjectTransportPlanAssignmentRulesCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	Rule      string
	AssetType string
	LevelType string
	LevelID   string
	IsActive  bool
}

func newDoProjectTransportPlanAssignmentRulesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project transport plan assignment rule",
		Long: `Create a project transport plan assignment rule.

Required flags:
  --rule        Rule text
  --asset-type  Asset type (driver/tractor/trailer)
  --level-type  Level type (JSON:API type, e.g., brokers)
  --level-id    Level ID

Optional flags:
  --is-active   Whether the rule is active

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create an assignment rule
  xbe do project-transport-plan-assignment-rules create \
    --rule "Prefer local drivers" \
    --asset-type driver \
    --level-type brokers \
    --level-id 123

  # Create an inactive rule
  xbe do project-transport-plan-assignment-rules create \
    --rule "Hold tractors for late shift" \
    --asset-type tractor \
    --level-type brokers \
    --level-id 123 \
    --is-active=false`,
		Args: cobra.NoArgs,
		RunE: runDoProjectTransportPlanAssignmentRulesCreate,
	}
	initDoProjectTransportPlanAssignmentRulesCreateFlags(cmd)
	return cmd
}

func init() {
	doProjectTransportPlanAssignmentRulesCmd.AddCommand(newDoProjectTransportPlanAssignmentRulesCreateCmd())
}

func initDoProjectTransportPlanAssignmentRulesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rule", "", "Rule text (required)")
	cmd.Flags().String("asset-type", "", "Asset type (driver/tractor/trailer) (required)")
	cmd.Flags().String("level-type", "", "Level type (JSON:API type, required)")
	cmd.Flags().String("level-id", "", "Level ID (required)")
	cmd.Flags().Bool("is-active", false, "Whether the rule is active")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("rule")
	cmd.MarkFlagRequired("asset-type")
	cmd.MarkFlagRequired("level-type")
	cmd.MarkFlagRequired("level-id")
}

func runDoProjectTransportPlanAssignmentRulesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoProjectTransportPlanAssignmentRulesCreateOptions(cmd)
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

	attributes := map[string]any{
		"rule":       opts.Rule,
		"asset-type": opts.AssetType,
	}

	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}

	relationships := map[string]any{
		"level": map[string]any{
			"data": map[string]any{
				"type": opts.LevelType,
				"id":   opts.LevelID,
			},
		},
	}

	data := map[string]any{
		"type":          "project-transport-plan-assignment-rules",
		"attributes":    attributes,
		"relationships": relationships,
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

	body, _, err := client.Post(cmd.Context(), "/v1/project-transport-plan-assignment-rules", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created project transport plan assignment rule %s\n", row.ID)
	return nil
}

func parseDoProjectTransportPlanAssignmentRulesCreateOptions(cmd *cobra.Command) (doProjectTransportPlanAssignmentRulesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rule, _ := cmd.Flags().GetString("rule")
	assetType, _ := cmd.Flags().GetString("asset-type")
	levelType, _ := cmd.Flags().GetString("level-type")
	levelID, _ := cmd.Flags().GetString("level-id")
	isActive, _ := cmd.Flags().GetBool("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doProjectTransportPlanAssignmentRulesCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		Rule:      rule,
		AssetType: assetType,
		LevelType: levelType,
		LevelID:   levelID,
		IsActive:  isActive,
	}, nil
}
