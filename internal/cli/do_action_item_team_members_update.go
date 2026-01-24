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

type doActionItemTeamMembersUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	IsResponsiblePerson bool
}

func newDoActionItemTeamMembersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an action item team member",
		Long: `Update an existing action item team member.

Arguments:
  <id>    The team member ID (required)

Flags:
  --is-responsible-person  Whether the user is the responsible person`,
		Example: `  # Mark a team member as responsible
  xbe do action-item-team-members update 123 --is-responsible-person

  # Clear responsible person flag
  xbe do action-item-team-members update 123 --is-responsible-person=false

  # Get JSON output
  xbe do action-item-team-members update 123 --is-responsible-person --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoActionItemTeamMembersUpdate,
	}
	initDoActionItemTeamMembersUpdateFlags(cmd)
	return cmd
}

func init() {
	doActionItemTeamMembersCmd.AddCommand(newDoActionItemTeamMembersUpdateCmd())
}

func initDoActionItemTeamMembersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-responsible-person", false, "Whether the user is the responsible person")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoActionItemTeamMembersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoActionItemTeamMembersUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("is-responsible-person") {
		attributes["is-responsible-person"] = opts.IsResponsiblePerson
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "action-item-team-members",
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

	body, _, err := client.Patch(cmd.Context(), "/v1/action-item-team-members/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated action item team member %s\n", details.ID)
	return nil
}

func parseDoActionItemTeamMembersUpdateOptions(cmd *cobra.Command, args []string) (doActionItemTeamMembersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isResponsiblePerson, _ := cmd.Flags().GetBool("is-responsible-person")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doActionItemTeamMembersUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		IsResponsiblePerson: isResponsiblePerson,
	}, nil
}
