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

type doOpenDoorTeamMembershipsCreateOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	MembershipID     string
	OrganizationType string
	OrganizationID   string
}

func newDoOpenDoorTeamMembershipsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an open door team membership",
		Long: `Create an open door team membership.

Required flags:
  --membership         Membership ID
  --organization-type  Organization type (brokers|truckers|customers)
  --organization-id    Organization ID`,
		Example: `  # Create an open door team membership for a broker
  xbe do open-door-team-memberships create \
    --membership 456 \
    --organization-type brokers \
    --organization-id 123

  # Get JSON output
  xbe do open-door-team-memberships create \
    --membership 456 \
    --organization-type brokers \
    --organization-id 123 \
    --json`,
		Args: cobra.NoArgs,
		RunE: runDoOpenDoorTeamMembershipsCreate,
	}
	initDoOpenDoorTeamMembershipsCreateFlags(cmd)
	return cmd
}

func init() {
	doOpenDoorTeamMembershipsCmd.AddCommand(newDoOpenDoorTeamMembershipsCreateCmd())
}

func initDoOpenDoorTeamMembershipsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("membership", "", "Membership ID (required)")
	cmd.Flags().String("organization-type", "", "Organization type (brokers|truckers|customers) (required)")
	cmd.Flags().String("organization-id", "", "Organization ID (required)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoOpenDoorTeamMembershipsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoOpenDoorTeamMembershipsCreateOptions(cmd)
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

	membershipID := strings.TrimSpace(opts.MembershipID)
	if membershipID == "" {
		err := fmt.Errorf("--membership is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	organizationType := strings.TrimSpace(opts.OrganizationType)
	if organizationType == "" {
		err := fmt.Errorf("--organization-type is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	organizationID := strings.TrimSpace(opts.OrganizationID)
	if organizationID == "" {
		err := fmt.Errorf("--organization-id is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{
		"membership": map[string]any{
			"data": map[string]any{
				"type": "memberships",
				"id":   membershipID,
			},
		},
		"organization": map[string]any{
			"data": map[string]any{
				"type": organizationType,
				"id":   organizationID,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "open-door-team-memberships",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/open-door-team-memberships", jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Created open door team membership %s\n", details.ID)
	return nil
}

func parseDoOpenDoorTeamMembershipsCreateOptions(cmd *cobra.Command) (doOpenDoorTeamMembershipsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	membershipID, _ := cmd.Flags().GetString("membership")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doOpenDoorTeamMembershipsCreateOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		MembershipID:     membershipID,
		OrganizationType: organizationType,
		OrganizationID:   organizationID,
	}, nil
}
