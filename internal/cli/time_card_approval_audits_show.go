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

type timeCardApprovalAuditsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type timeCardApprovalAuditDetails struct {
	ID         string `json:"id"`
	TimeCardID string `json:"time_card_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	IsBot      bool   `json:"is_bot"`
	Note       string `json:"note,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

func newTimeCardApprovalAuditsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show time card approval audit details",
		Long: `Show the full details of a time card approval audit.

Output Fields:
  ID         Audit identifier
  Time Card  Time card ID
  User       User ID (blank when bot)
  Is Bot     Whether audit was created by a bot
  Note       Audit note
  Created At Audit creation timestamp
  Updated At Audit last update timestamp

Arguments:
  <id>    The time card approval audit ID (required). You can find IDs using the list command.`,
		Example: `  # Show a time card approval audit
  xbe view time-card-approval-audits show 123

  # Get JSON output
  xbe view time-card-approval-audits show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTimeCardApprovalAuditsShow,
	}
	initTimeCardApprovalAuditsShowFlags(cmd)
	return cmd
}

func init() {
	timeCardApprovalAuditsCmd.AddCommand(newTimeCardApprovalAuditsShowCmd())
}

func initTimeCardApprovalAuditsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardApprovalAuditsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTimeCardApprovalAuditsShowOptions(cmd)
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
		return fmt.Errorf("time card approval audit id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-approval-audits/"+id, nil)
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

	details := buildTimeCardApprovalAuditDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTimeCardApprovalAuditDetails(cmd, details)
}

func parseTimeCardApprovalAuditsShowOptions(cmd *cobra.Command) (timeCardApprovalAuditsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardApprovalAuditsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTimeCardApprovalAuditDetails(resp jsonAPISingleResponse) timeCardApprovalAuditDetails {
	resource := resp.Data
	attrs := resource.Attributes
	results := timeCardApprovalAuditDetails{
		ID:        resource.ID,
		IsBot:     boolAttr(attrs, "is-bot"),
		Note:      stringAttr(attrs, "note"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		results.TimeCardID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		results.UserID = rel.Data.ID
	}

	return results
}

func renderTimeCardApprovalAuditDetails(cmd *cobra.Command, details timeCardApprovalAuditDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TimeCardID != "" {
		fmt.Fprintf(out, "Time Card: %s\n", details.TimeCardID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserID)
	}
	fmt.Fprintf(out, "Is Bot: %s\n", boolToYesNo(details.IsBot))
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
