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

type doOpenDoorTeamMembershipsUpdateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	ID               string
	MembershipID     string
	OrganizationType string
	OrganizationID   string
}

func newDoOpenDoorTeamMembershipsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an open door team membership",
		Long: `Update an existing open door team membership.

Arguments:
  <id>    The open door team membership ID (required)

Flags:
  --membership         Membership ID
  --organization-type  Organization type (brokers|truckers|customers)
  --organization-id    Organization ID`,
		Example: `  # Update relationships
  xbe do open-door-team-memberships update 123 \
    --membership 456 \
    --organization-type brokers \
    --organization-id 789

  # Get JSON output
  xbe do open-door-team-memberships update 123 --membership 456 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoOpenDoorTeamMembershipsUpdate,
	}
	initDoOpenDoorTeamMembershipsUpdateFlags(cmd)
	return cmd
}

func init() {
	doOpenDoorTeamMembershipsCmd.AddCommand(newDoOpenDoorTeamMembershipsUpdateCmd())
}

func initDoOpenDoorTeamMembershipsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("membership", "", "Membership ID")
	cmd.Flags().String("organization-type", "", "Organization type (brokers|truckers|customers)")
	cmd.Flags().String("organization-id", "", "Organization ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOpenDoorTeamMembershipsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoOpenDoorTeamMembershipsUpdateOptions(cmd, args)
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

	relationships := map[string]any{}

	if cmd.Flags().Changed("membership") {
		if strings.TrimSpace(opts.MembershipID) == "" {
			err := fmt.Errorf("--membership cannot be empty")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["membership"] = map[string]any{
			"data": map[string]any{
				"type": "memberships",
				"id":   opts.MembershipID,
			},
		}
	}

	orgTypeChanged := cmd.Flags().Changed("organization-type")
	orgIDChanged := cmd.Flags().Changed("organization-id")
	if orgTypeChanged || orgIDChanged {
		if strings.TrimSpace(opts.OrganizationType) == "" || strings.TrimSpace(opts.OrganizationID) == "" {
			err := fmt.Errorf("--organization-type and --organization-id are required together")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		relationships["organization"] = map[string]any{
			"data": map[string]any{
				"type": opts.OrganizationType,
				"id":   opts.OrganizationID,
			},
		}
	}

	if len(relationships) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "open-door-team-memberships",
			"id":            opts.ID,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/open-door-team-memberships/"+opts.ID, jsonBody)
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

	details := buildOpenDoorTeamMembershipDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated open door team membership %s\n", details.ID)
	return nil
}

func parseDoOpenDoorTeamMembershipsUpdateOptions(cmd *cobra.Command, args []string) (doOpenDoorTeamMembershipsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	membershipID, _ := cmd.Flags().GetString("membership")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOpenDoorTeamMembershipsUpdateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		ID:               args[0],
		MembershipID:     membershipID,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
	}, nil
}
