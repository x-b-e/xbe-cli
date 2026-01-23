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

type hosViolationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type hosViolationDetails struct {
	ID                string `json:"id"`
	RegulationSetCode string `json:"regulation_set_code,omitempty"`
	ViolationType     string `json:"violation_type,omitempty"`
	StartAt           string `json:"start_at,omitempty"`
	EndAt             string `json:"end_at,omitempty"`
	RuleID            string `json:"rule_id,omitempty"`
	RuleName          string `json:"rule_name,omitempty"`
	Detail            string `json:"detail,omitempty"`
	BrokerID          string `json:"broker_id,omitempty"`
	HosDayID          string `json:"hos_day_id,omitempty"`
	UserID            string `json:"user_id,omitempty"`
}

func newHosViolationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show HOS violation details",
		Long: `Show the full details of an HOS violation.

Output Fields:
  ID
  Regulation Set Code
  Violation Type
  Start At
  End At
  Rule ID
  Rule Name
  Detail
  Broker ID
  HOS Day ID
  User ID

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The violation ID (required). You can find IDs using the list command.`,
		Example: `  # Show a violation
  xbe view hos-violations show 123

  # Get JSON output
  xbe view hos-violations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHosViolationsShow,
	}
	initHosViolationsShowFlags(cmd)
	return cmd
}

func init() {
	hosViolationsCmd.AddCommand(newHosViolationsShowCmd())
}

func initHosViolationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosViolationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseHosViolationsShowOptions(cmd)
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
		return fmt.Errorf("hos violation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[hos-violations]", "regulation-set-code,violation-type,start-at,end-at,rule-id,rule-name,detail")

	body, _, err := client.Get(cmd.Context(), "/v1/hos-violations/"+id, query)
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

	details := buildHosViolationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHosViolationDetails(cmd, details)
}

func parseHosViolationsShowOptions(cmd *cobra.Command) (hosViolationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosViolationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHosViolationDetails(resp jsonAPISingleResponse) hosViolationDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := hosViolationDetails{
		ID:                resource.ID,
		RegulationSetCode: stringAttr(attrs, "regulation-set-code"),
		ViolationType:     stringAttr(attrs, "violation-type"),
		StartAt:           formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:             formatDateTime(stringAttr(attrs, "end-at")),
		RuleID:            stringAttr(attrs, "rule-id"),
		RuleName:          stringAttr(attrs, "rule-name"),
		Detail:            stringAttr(attrs, "detail"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
		details.HosDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}

	return details
}

func renderHosViolationDetails(cmd *cobra.Command, details hosViolationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.RegulationSetCode != "" {
		fmt.Fprintf(out, "Regulation Set Code: %s\n", details.RegulationSetCode)
	}
	if details.ViolationType != "" {
		fmt.Fprintf(out, "Violation Type: %s\n", details.ViolationType)
	}
	if details.StartAt != "" {
		fmt.Fprintf(out, "Start At: %s\n", details.StartAt)
	}
	if details.EndAt != "" {
		fmt.Fprintf(out, "End At: %s\n", details.EndAt)
	}
	if details.RuleID != "" {
		fmt.Fprintf(out, "Rule ID: %s\n", details.RuleID)
	}
	if details.RuleName != "" {
		fmt.Fprintf(out, "Rule Name: %s\n", details.RuleName)
	}
	if details.Detail != "" {
		fmt.Fprintf(out, "Detail: %s\n", details.Detail)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.HosDayID != "" {
		fmt.Fprintf(out, "HOS Day ID: %s\n", details.HosDayID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}

	return nil
}
