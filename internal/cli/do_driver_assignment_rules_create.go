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

type doDriverAssignmentRulesCreateOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	Rule      string
	IsActive  bool
	LevelType string
	LevelID   string
}

func newDoDriverAssignmentRulesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a driver assignment rule",
		Long: `Create a driver assignment rule.

Required flags:
  --rule        Rule text
  --level-type  Level type (Broker, JobScheduleShift, Project, JobProductionPlan, MaterialSupplier, MaterialSite, MaterialType, Trucker, JobSite)
  --level-id    Level ID

Optional flags:
  --is-active   Whether the rule is active

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create a broker-level rule
  xbe do driver-assignment-rules create \
    --rule "Drivers must be assigned by 6am" \
    --level-type Broker \
    --level-id 123 \
    --is-active

  # Create a project-level rule
  xbe do driver-assignment-rules create \
    --rule "Only certified drivers" \
    --level-type Project \
    --level-id 456`,
		Args: cobra.NoArgs,
		RunE: runDoDriverAssignmentRulesCreate,
	}
	initDoDriverAssignmentRulesCreateFlags(cmd)
	return cmd
}

func init() {
	doDriverAssignmentRulesCmd.AddCommand(newDoDriverAssignmentRulesCreateCmd())
}

func initDoDriverAssignmentRulesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("rule", "", "Rule text (required)")
	cmd.Flags().Bool("is-active", false, "Whether the rule is active")
	cmd.Flags().String("level-type", "", "Level type (Broker, JobScheduleShift, Project, JobProductionPlan, MaterialSupplier, MaterialSite, MaterialType, Trucker, JobSite)")
	cmd.Flags().String("level-id", "", "Level ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDriverAssignmentRulesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDriverAssignmentRulesCreateOptions(cmd)
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

	if strings.TrimSpace(opts.Rule) == "" {
		err := fmt.Errorf("--rule is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.LevelType == "" {
		err := fmt.Errorf("--level-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.LevelID == "" {
		err := fmt.Errorf("--level-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	levelType, err := parseDriverAssignmentRuleLevelType(opts.LevelType)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{
		"rule": opts.Rule,
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}

	relationships := map[string]any{
		"level": map[string]any{
			"data": map[string]any{
				"type": levelType,
				"id":   opts.LevelID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "driver-assignment-rules",
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

	body, _, err := client.Post(cmd.Context(), "/v1/driver-assignment-rules", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created driver assignment rule %s\n", row.ID)
	return nil
}

func parseDoDriverAssignmentRulesCreateOptions(cmd *cobra.Command) (doDriverAssignmentRulesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	rule, _ := cmd.Flags().GetString("rule")
	isActive, _ := cmd.Flags().GetBool("is-active")
	levelType, _ := cmd.Flags().GetString("level-type")
	levelID, _ := cmd.Flags().GetString("level-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDriverAssignmentRulesCreateOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		Rule:      rule,
		IsActive:  isActive,
		LevelType: levelType,
		LevelID:   levelID,
	}, nil
}
