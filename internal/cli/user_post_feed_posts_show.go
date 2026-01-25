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

type userPostFeedPostsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type userPostFeedPostDetails struct {
	ID                  string   `json:"id"`
	UserID              string   `json:"user_id,omitempty"`
	UserPostFeedID      string   `json:"user_post_feed_id,omitempty"`
	PostID              string   `json:"post_id,omitempty"`
	FollowID            string   `json:"follow_id,omitempty"`
	FeedAt              string   `json:"feed_at,omitempty"`
	Score               *float64 `json:"score,omitempty"`
	IsViewedByUser      bool     `json:"is_viewed_by_user"`
	IsBookmarked        bool     `json:"is_bookmarked"`
	SubscriptionStartAt string   `json:"subscription_start_at,omitempty"`
	SubscriptionEndAt   string   `json:"subscription_end_at,omitempty"`
	CreatedAt           string   `json:"created_at,omitempty"`
	UpdatedAt           string   `json:"updated_at,omitempty"`
}

func newUserPostFeedPostsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show user post feed post details",
		Long: `Show the full details of a user post feed post.

Output Fields:
  ID
  User ID
  User Post Feed ID
  Post ID
  Follow ID
  Feed At
  Score
  Is Viewed By User
  Is Bookmarked
  Subscription Start At
  Subscription End At
  Created At
  Updated At

Arguments:
  <id>    The user post feed post ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a user post feed post
  xbe view user-post-feed-posts show 123

  # Get JSON output
  xbe view user-post-feed-posts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUserPostFeedPostsShow,
	}
	initUserPostFeedPostsShowFlags(cmd)
	return cmd
}

func init() {
	userPostFeedPostsCmd.AddCommand(newUserPostFeedPostsShowCmd())
}

func initUserPostFeedPostsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserPostFeedPostsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseUserPostFeedPostsShowOptions(cmd)
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
		return fmt.Errorf("user post feed post id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-post-feed-posts]", "feed-at,score,is-viewed-by-user,is-bookmarked,subscription-start-at,subscription-end-at,created-at,updated-at,user,post,user-post-feed,follow")

	body, _, err := client.Get(cmd.Context(), "/v1/user-post-feed-posts/"+id, query)
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

	details := buildUserPostFeedPostDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUserPostFeedPostDetails(cmd, details)
}

func parseUserPostFeedPostsShowOptions(cmd *cobra.Command) (userPostFeedPostsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userPostFeedPostsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUserPostFeedPostDetails(resp jsonAPISingleResponse) userPostFeedPostDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return userPostFeedPostDetails{
		ID:                  resource.ID,
		UserID:              relationshipIDFromMap(resource.Relationships, "user"),
		UserPostFeedID:      relationshipIDFromMap(resource.Relationships, "user-post-feed"),
		PostID:              relationshipIDFromMap(resource.Relationships, "post"),
		FollowID:            relationshipIDFromMap(resource.Relationships, "follow"),
		FeedAt:              formatDateTime(stringAttr(attrs, "feed-at")),
		Score:               floatAttrPointer(attrs, "score"),
		IsViewedByUser:      boolAttr(attrs, "is-viewed-by-user"),
		IsBookmarked:        boolAttr(attrs, "is-bookmarked"),
		SubscriptionStartAt: formatDateTime(stringAttr(attrs, "subscription-start-at")),
		SubscriptionEndAt:   formatDateTime(stringAttr(attrs, "subscription-end-at")),
		CreatedAt:           formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:           formatDateTime(stringAttr(attrs, "updated-at")),
	}
}

func renderUserPostFeedPostDetails(cmd *cobra.Command, details userPostFeedPostDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}
	if details.UserPostFeedID != "" {
		fmt.Fprintf(out, "User Post Feed ID: %s\n", details.UserPostFeedID)
	}
	if details.PostID != "" {
		fmt.Fprintf(out, "Post ID: %s\n", details.PostID)
	}
	if details.FollowID != "" {
		fmt.Fprintf(out, "Follow ID: %s\n", details.FollowID)
	}
	if details.FeedAt != "" {
		fmt.Fprintf(out, "Feed At: %s\n", details.FeedAt)
	}
	if details.Score != nil {
		fmt.Fprintf(out, "Score: %.2f\n", *details.Score)
	}
	fmt.Fprintf(out, "Is Viewed By User: %t\n", details.IsViewedByUser)
	fmt.Fprintf(out, "Is Bookmarked: %t\n", details.IsBookmarked)
	if details.SubscriptionStartAt != "" {
		fmt.Fprintf(out, "Subscription Start At: %s\n", details.SubscriptionStartAt)
	}
	if details.SubscriptionEndAt != "" {
		fmt.Fprintf(out, "Subscription End At: %s\n", details.SubscriptionEndAt)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
