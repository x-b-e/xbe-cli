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

type doActionItemTeamMembersCreateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ActionItem          string
	User                string
	IsResponsiblePerson bool
}

func newDoActionItemTeamMembersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an action item team member",
		Long: `Create an action item team member.

Required flags:
  --action-item            Action item ID
  --user                   User ID

Optional flags:
  --is-responsible-person  Whether the user is the responsible person`,
		Example: `  # Add a team member to an action item
  xbe do action-item-team-members create --action-item 123 --user 456

  # Create and mark as responsible person
  xbe do action-item-team-members create --action-item 123 --user 456 --is-responsible-person

  # Get JSON output
  xbe do action-item-team-members create --action-item 123 --user 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoActionItemTeamMembersCreate,
	}
	initDoActionItemTeamMembersCreateFlags(cmd)
	return cmd
}

func init() {
	doActionItemTeamMembersCmd.AddCommand(newDoActionItemTeamMembersCreateCmd())
}

func initDoActionItemTeamMembersCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("action-item", "", "Action item ID (required)")
	cmd.Flags().String("user", "", "User ID (required)")
	cmd.Flags().Bool("is-responsible-person", false, "Whether the user is the responsible person")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemTeamMembersCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoActionItemTeamMembersCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	actionItemID := strings.TrimSpace(opts.ActionItem)
	if actionItemID == "" {
		err := fmt.Errorf("--action-item is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	userID := strings.TrimSpace(opts.User)
	if userID == "" {
		err := fmt.Errorf("--user is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("is-responsible-person") {
		attributes["is-responsible-person"] = opts.IsResponsiblePerson
	}

	relationships := map[string]any{
		"action-item": map[string]any{
			"data": map[string]any{
				"type": "action-items",
				"id":   actionItemID,
			},
		},
		"user": map[string]any{
			"data": map[string]any{
				"type": "users",
				"id":   userID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "action-item-team-members",
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

	body, _, err := client.Post(cmd.Context(), "/v1/action-item-team-members", jsonBody)
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

	details := buildActionItemTeamMemberDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created action item team member %s\n", details.ID)
	return nil
}

func parseDoActionItemTeamMembersCreateOptions(cmd *cobra.Command) (doActionItemTeamMembersCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	actionItem, _ := cmd.Flags().GetString("action-item")
	user, _ := cmd.Flags().GetString("user")
	isResponsiblePerson, _ := cmd.Flags().GetBool("is-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemTeamMembersCreateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ActionItem:          actionItem,
		User:                user,
		IsResponsiblePerson: isResponsiblePerson,
	}, nil
}
