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

type openDoorTeamMembershipsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type openDoorTeamMembershipDetails struct {
	ID               string `json:"id"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	MembershipID     string `json:"membership_id,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

func newOpenDoorTeamMembershipsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show open door team membership details",
		Long: `Show the full details of an open door team membership.

Output Fields:
  ID
  Organization Type
  Organization ID
  Membership ID
  Created At
  Updated At

Arguments:
  <id>    The open door team membership ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an open door team membership
  xbe view open-door-team-memberships show 123

  # Get JSON output
  xbe view open-door-team-memberships show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runOpenDoorTeamMembershipsShow,
	}
	initOpenDoorTeamMembershipsShowFlags(cmd)
	return cmd
}

func init() {
	openDoorTeamMembershipsCmd.AddCommand(newOpenDoorTeamMembershipsShowCmd())
}

func initOpenDoorTeamMembershipsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOpenDoorTeamMembershipsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseOpenDoorTeamMembershipsShowOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("open door team membership id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[open-door-team-memberships]", "created-at,updated-at,organization,membership")

	body, _, err := client.Get(cmd.Context(), "/v1/open-door-team-memberships/"+id, query)
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

	return renderOpenDoorTeamMembershipDetails(cmd, details)
}

func parseOpenDoorTeamMembershipsShowOptions(cmd *cobra.Command) (openDoorTeamMembershipsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return openDoorTeamMembershipsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildOpenDoorTeamMembershipDetails(resp jsonAPISingleResponse) openDoorTeamMembershipDetails {
	attrs := resp.Data.Attributes
	details := openDoorTeamMembershipDetails{
		ID:           resp.Data.ID,
		MembershipID: relationshipIDFromMap(resp.Data.Relationships, "membership"),
		CreatedAt:    formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:    formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["organization"]; ok && rel.Data != nil {
		details.OrganizationType = rel.Data.Type
		details.OrganizationID = rel.Data.ID
	}

	return details
}

func renderOpenDoorTeamMembershipDetails(cmd *cobra.Command, details openDoorTeamMembershipDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.OrganizationType != "" {
		fmt.Fprintf(out, "Organization Type: %s\n", details.OrganizationType)
	}
	if details.OrganizationID != "" {
		fmt.Fprintf(out, "Organization ID: %s\n", details.OrganizationID)
	}
	if details.MembershipID != "" {
		fmt.Fprintf(out, "Membership ID: %s\n", details.MembershipID)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
