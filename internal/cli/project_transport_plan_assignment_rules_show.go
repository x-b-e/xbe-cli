package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type projectTransportPlanAssignmentRulesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanAssignmentRuleDetails struct {
	ID        string `json:"id"`
	Rule      string `json:"rule,omitempty"`
	AssetType string `json:"asset_type,omitempty"`
	IsActive  bool   `json:"is_active"`
	LevelType string `json:"level_type,omitempty"`
	LevelID   string `json:"level_id,omitempty"`
	BrokerID  string `json:"broker_id,omitempty"`
	Broker    string `json:"broker,omitempty"`
}

func newProjectTransportPlanAssignmentRulesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan assignment rule details",
		Long: `Show the full details of a project transport plan assignment rule.

Output Fields:
  ID         Assignment rule identifier
  Rule       Rule text
  Asset Type Asset type (driver/tractor/trailer)
  Is Active  Whether the rule is active
  Level      Level type and ID
  Broker     Broker name and ID

Arguments:
  <id>    The assignment rule ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an assignment rule
  xbe view project-transport-plan-assignment-rules show 123

  # Get JSON output
  xbe view project-transport-plan-assignment-rules show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanAssignmentRulesShow,
	}
	initProjectTransportPlanAssignmentRulesShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanAssignmentRulesCmd.AddCommand(newProjectTransportPlanAssignmentRulesShowCmd())
}

func initProjectTransportPlanAssignmentRulesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanAssignmentRulesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanAssignmentRulesShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("project transport plan assignment rule id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "broker")
	query.Set("fields[brokers]", "company-name")

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-assignment-rules/"+id, query)
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

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildProjectTransportPlanAssignmentRuleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanAssignmentRuleDetails(cmd, details)
}

func parseProjectTransportPlanAssignmentRulesShowOptions(cmd *cobra.Command) (projectTransportPlanAssignmentRulesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanAssignmentRulesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanAssignmentRuleDetails(resp jsonAPISingleResponse) projectTransportPlanAssignmentRuleDetails {
	resource := resp.Data
	attrs := resource.Attributes

	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := projectTransportPlanAssignmentRuleDetails{
		ID:        resource.ID,
		Rule:      stringAttr(attrs, "rule"),
		AssetType: stringAttr(attrs, "asset-type"),
		IsActive:  boolAttr(attrs, "is-active"),
	}

	if rel, ok := resource.Relationships["level"]; ok && rel.Data != nil {
		details.LevelType = rel.Data.Type
		details.LevelID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
		if broker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.Broker = stringAttr(broker.Attributes, "company-name")
		}
	}

	return details
}

func renderProjectTransportPlanAssignmentRuleDetails(cmd *cobra.Command, details projectTransportPlanAssignmentRuleDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Rule != "" {
		fmt.Fprintf(out, "Rule: %s\n", details.Rule)
	}
	if details.AssetType != "" {
		fmt.Fprintf(out, "Asset Type: %s\n", details.AssetType)
	}
	fmt.Fprintf(out, "Is Active: %t\n", details.IsActive)
	if details.LevelType != "" && details.LevelID != "" {
		fmt.Fprintf(out, "Level: %s/%s\n", details.LevelType, details.LevelID)
	}
	if details.BrokerID != "" {
		if details.Broker != "" {
			fmt.Fprintf(out, "Broker: %s (%s)\n", details.Broker, details.BrokerID)
		} else {
			fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
		}
	}

	return nil
}
