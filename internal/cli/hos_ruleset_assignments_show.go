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

type hosRulesetAssignmentsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type hosRulesetAssignmentDetails struct {
	ID          string `json:"id"`
	RuleSetID   string `json:"rule_set_id,omitempty"`
	Name        string `json:"name,omitempty"`
	EffectiveAt string `json:"effective_at,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	HosDayID    string `json:"hos_day_id,omitempty"`
	BrokerID    string `json:"broker_id,omitempty"`
}

func newHosRulesetAssignmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show HOS ruleset assignment details",
		Long: `Show the full details of a HOS ruleset assignment.

Output Fields:
  ID
  Rule Set ID
  Name
  Effective At
  User ID
  HOS Day ID
  Broker ID

Arguments:
  <id>    The HOS ruleset assignment ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a HOS ruleset assignment
  xbe view hos-ruleset-assignments show 123

  # Output as JSON
  xbe view hos-ruleset-assignments show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHosRulesetAssignmentsShow,
	}
	initHosRulesetAssignmentsShowFlags(cmd)
	return cmd
}

func init() {
	hosRulesetAssignmentsCmd.AddCommand(newHosRulesetAssignmentsShowCmd())
}

func initHosRulesetAssignmentsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosRulesetAssignmentsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseHosRulesetAssignmentsShowOptions(cmd)
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
		return fmt.Errorf("HOS ruleset assignment id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-ruleset-assignments/"+id, nil)
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

	details := buildHosRulesetAssignmentDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHosRulesetAssignmentDetails(cmd, details)
}

func parseHosRulesetAssignmentsShowOptions(cmd *cobra.Command) (hosRulesetAssignmentsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosRulesetAssignmentsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHosRulesetAssignmentDetails(resp jsonAPISingleResponse) hosRulesetAssignmentDetails {
	resource := resp.Data
	details := hosRulesetAssignmentDetails{
		ID:          resource.ID,
		RuleSetID:   stringAttr(resource.Attributes, "rule-set-id"),
		Name:        stringAttr(resource.Attributes, "name"),
		EffectiveAt: formatDateTime(stringAttr(resource.Attributes, "effective-at")),
	}

	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
		details.HosDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}

	return details
}

func renderHosRulesetAssignmentDetails(cmd *cobra.Command, details hosRulesetAssignmentDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RuleSetID != "" {
		fmt.Fprintf(out, "Rule Set ID: %s\n", details.RuleSetID)
	}
	if details.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", details.Name)
	}
	if details.EffectiveAt != "" {
		fmt.Fprintf(out, "Effective At: %s\n", details.EffectiveAt)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.HosDayID != "" {
		fmt.Fprintf(out, "HOS Day ID: %s\n", details.HosDayID)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}

	return nil
}
