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

type actionItemTeamMembersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type actionItemTeamMemberDetails struct {
	ID                  string `json:"id"`
	ActionItemID        string `json:"action_item_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
	IsResponsiblePerson bool   `json:"is_responsible_person"`
}

func newActionItemTeamMembersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show action item team member details",
		Long: `Show the full details of an action item team member.

Output Fields:
  ID
  Action Item ID
  User ID
  Is Responsible Person

Arguments:
  <id>    The team member ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a team member
  xbe view action-item-team-members show 123

  # JSON output
  xbe view action-item-team-members show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runActionItemTeamMembersShow,
	}
	initActionItemTeamMembersShowFlags(cmd)
	return cmd
}

func init() {
	actionItemTeamMembersCmd.AddCommand(newActionItemTeamMembersShowCmd())
}

func initActionItemTeamMembersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemTeamMembersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseActionItemTeamMembersShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("action item team member id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[action-item-team-members]", "is-responsible-person,action-item,user")

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-team-members/"+id, query)
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

	details := buildActionItemTeamMemberDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderActionItemTeamMemberDetails(cmd, details)
}

func parseActionItemTeamMembersShowOptions(cmd *cobra.Command) (actionItemTeamMembersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemTeamMembersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildActionItemTeamMemberDetails(resp jsonAPISingleResponse) actionItemTeamMemberDetails {
	attrs := resp.Data.Attributes
	return actionItemTeamMemberDetails{
		ID:                  resp.Data.ID,
		ActionItemID:        relationshipIDFromMap(resp.Data.Relationships, "action-item"),
		UserID:              relationshipIDFromMap(resp.Data.Relationships, "user"),
		IsResponsiblePerson: boolAttr(attrs, "is-responsible-person"),
	}
}

func renderActionItemTeamMemberDetails(cmd *cobra.Command, details actionItemTeamMemberDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ActionItemID != "" {
		fmt.Fprintf(out, "Action Item ID: %s\n", details.ActionItemID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	fmt.Fprintf(out, "Is Responsible Person: %s\n", formatBool(details.IsResponsiblePerson))

	return nil
}
