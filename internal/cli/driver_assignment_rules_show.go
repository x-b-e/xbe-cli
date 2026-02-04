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

type driverAssignmentRulesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type driverAssignmentRuleDetails struct {
	ID        string `json:"id"`
	Rule      string `json:"rule,omitempty"`
	IsActive  bool   `json:"is_active"`
	LevelType string `json:"level_type,omitempty"`
	LevelID   string `json:"level_id,omitempty"`
	BrokerID  string `json:"broker_id,omitempty"`
}

func newDriverAssignmentRulesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show driver assignment rule details",
		Long: `Show the full details of a driver assignment rule.

Output Fields:
  ID
  Rule
  Is Active
  Level Type
  Level ID
  Broker ID

Arguments:
  <id>    The driver assignment rule ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a driver assignment rule
  xbe view driver-assignment-rules show 123

  # Output as JSON
  xbe view driver-assignment-rules show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDriverAssignmentRulesShow,
	}
	initDriverAssignmentRulesShowFlags(cmd)
	return cmd
}

func init() {
	driverAssignmentRulesCmd.AddCommand(newDriverAssignmentRulesShowCmd())
}

func initDriverAssignmentRulesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverAssignmentRulesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseDriverAssignmentRulesShowOptions(cmd)
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
		return fmt.Errorf("driver assignment rule id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-assignment-rules/"+id, nil)
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

	details := buildDriverAssignmentRuleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderDriverAssignmentRuleDetails(cmd, details)
}

func parseDriverAssignmentRulesShowOptions(cmd *cobra.Command) (driverAssignmentRulesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverAssignmentRulesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildDriverAssignmentRuleDetails(resp jsonAPISingleResponse) driverAssignmentRuleDetails {
	resource := resp.Data
	details := driverAssignmentRuleDetails{
		ID:       resource.ID,
		Rule:     stringAttr(resource.Attributes, "rule"),
		IsActive: boolAttr(resource.Attributes, "is-active"),
	}

	if rel, ok := resource.Relationships["level"]; ok && rel.Data != nil {
		details.LevelType = rel.Data.Type
		details.LevelID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}

	return details
}

func renderDriverAssignmentRuleDetails(cmd *cobra.Command, details driverAssignmentRuleDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Rule != "" {
		fmt.Fprintf(out, "Rule: %s\n", details.Rule)
	}
	fmt.Fprintf(out, "Is Active: %t\n", details.IsActive)
	if details.LevelType != "" {
		fmt.Fprintf(out, "Level Type: %s\n", details.LevelType)
	}
	if details.LevelID != "" {
		fmt.Fprintf(out, "Level ID: %s\n", details.LevelID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}

	return nil
}
