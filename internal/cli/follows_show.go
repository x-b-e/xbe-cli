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

type followsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type followDetails struct {
	ID                        string   `json:"id"`
	FollowerID                string   `json:"follower_id,omitempty"`
	FollowerName              string   `json:"follower_name,omitempty"`
	FollowerEmail             string   `json:"follower_email,omitempty"`
	CreatorType               string   `json:"creator_type,omitempty"`
	CreatorID                 string   `json:"creator_id,omitempty"`
	UserCreatorFeedCreatorIDs []string `json:"user_creator_feed_creator_ids,omitempty"`
	CreatedAt                 string   `json:"created_at,omitempty"`
	UpdatedAt                 string   `json:"updated_at,omitempty"`
}

func newFollowsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show follow details",
		Long: `Show the full details of a follow relationship.

Arguments:
  <id>  The follow ID (required).`,
		Example: `  # Show a follow
  xbe view follows show 123

  # Output as JSON
  xbe view follows show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runFollowsShow,
	}
	initFollowsShowFlags(cmd)
	return cmd
}

func init() {
	followsCmd.AddCommand(newFollowsShowCmd())
}

func initFollowsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runFollowsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseFollowsShowOptions(cmd)
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
		return fmt.Errorf("follow id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[follows]", "follower,creator,created-at,updated-at,user-creator-feed-creators")
	query.Set("include", "follower")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/follows/"+id, query)
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

	details := buildFollowDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderFollowDetails(cmd, details)
}

func parseFollowsShowOptions(cmd *cobra.Command) (followsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return followsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildFollowDetails(resp jsonAPISingleResponse) followDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := followDetails{
		ID:        resp.Data.ID,
		CreatedAt: formatDateTime(stringAttr(resp.Data.Attributes, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(resp.Data.Attributes, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["follower"]; ok && rel.Data != nil {
		details.FollowerID = rel.Data.ID
		if follower, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.FollowerName = stringAttr(follower.Attributes, "name")
			details.FollowerEmail = stringAttr(follower.Attributes, "email-address")
		}
	}

	if rel, ok := resp.Data.Relationships["creator"]; ok && rel.Data != nil {
		details.CreatorType = rel.Data.Type
		details.CreatorID = rel.Data.ID
	}

	if rel, ok := resp.Data.Relationships["user-creator-feed-creators"]; ok {
		details.UserCreatorFeedCreatorIDs = relationshipIDList(rel)
	}

	return details
}

func renderFollowDetails(cmd *cobra.Command, details followDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.FollowerID != "" || details.FollowerName != "" || details.FollowerEmail != "" {
		label := firstNonEmpty(details.FollowerName, details.FollowerEmail)
		fmt.Fprintf(out, "Follower: %s\n", formatRelated(label, details.FollowerID))
	}
	if details.CreatorType != "" || details.CreatorID != "" {
		fmt.Fprintf(out, "Creator: %s\n", formatPolymorphic(details.CreatorType, details.CreatorID))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if len(details.UserCreatorFeedCreatorIDs) > 0 {
		fmt.Fprintf(out, "User Creator Feed Creators: %s\n", strings.Join(details.UserCreatorFeedCreatorIDs, ", "))
	}

	return nil
}
