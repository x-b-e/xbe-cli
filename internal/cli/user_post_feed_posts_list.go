package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type userPostFeedPostsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	Score                  string
	FeedAtMin              string
	FeedAtMax              string
	IsFeedAt               string
	SubscriptionStartAtMin string
	SubscriptionStartAtMax string
	IsSubscriptionStartAt  string
	SubscriptionEndAtMin   string
	SubscriptionEndAtMax   string
	IsSubscriptionEndAt    string
	User                   string
	UserID                 string
	Post                   string
	PostType               string
	Creator                string
}

type userPostFeedPostRow struct {
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
}

func newUserPostFeedPostsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user post feed posts",
		Long: `List user post feed posts with filtering and pagination.

Output Columns:
  ID         Feed post identifier
  USER       User ID for the feed owner
  POST       Post ID
  FEED AT    When the post entered the feed
  SCORE      Ranking score (0-1)
  BOOKMARKED Whether the post is bookmarked
  VIEWED     Whether the user has viewed the post

Filters:
  --user                           Filter by user ID (relationship)
  --user-id                        Filter by user ID via feed membership
  --post                           Filter by post ID
  --post-type                      Filter by post type (comma-separated for multiple)
  --creator                        Filter by creator (e.g., User|123)
  --score                          Filter by score
  --feed-at-min                    Filter by feed-at on/after (ISO 8601)
  --feed-at-max                    Filter by feed-at on/before (ISO 8601)
  --is-feed-at                     Filter by presence of feed-at (true/false)
  --subscription-start-at-min      Filter by subscription start on/after (ISO 8601)
  --subscription-start-at-max      Filter by subscription start on/before (ISO 8601)
  --is-subscription-start-at       Filter by presence of subscription start (true/false)
  --subscription-end-at-min        Filter by subscription end on/after (ISO 8601)
  --subscription-end-at-max        Filter by subscription end on/before (ISO 8601)
  --is-subscription-end-at         Filter by presence of subscription end (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List user post feed posts
  xbe view user-post-feed-posts list

  # Filter by user
  xbe view user-post-feed-posts list --user 123

  # Filter by post
  xbe view user-post-feed-posts list --post 456

  # Filter by post type
  xbe view user-post-feed-posts list --post-type objective-status

  # Output as JSON
  xbe view user-post-feed-posts list --json`,
		Args: cobra.NoArgs,
		RunE: runUserPostFeedPostsList,
	}
	initUserPostFeedPostsListFlags(cmd)
	return cmd
}

func init() {
	userPostFeedPostsCmd.AddCommand(newUserPostFeedPostsListCmd())
}

func initUserPostFeedPostsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("user", "", "Filter by user ID (relationship)")
	cmd.Flags().String("user-id", "", "Filter by user ID via feed membership")
	cmd.Flags().String("post", "", "Filter by post ID")
	cmd.Flags().String("post-type", "", "Filter by post type (comma-separated for multiple)")
	cmd.Flags().String("creator", "", "Filter by creator (e.g., User|123)")
	cmd.Flags().String("score", "", "Filter by score")
	cmd.Flags().String("feed-at-min", "", "Filter by feed-at on/after (ISO 8601)")
	cmd.Flags().String("feed-at-max", "", "Filter by feed-at on/before (ISO 8601)")
	cmd.Flags().String("is-feed-at", "", "Filter by presence of feed-at (true/false)")
	cmd.Flags().String("subscription-start-at-min", "", "Filter by subscription start on/after (ISO 8601)")
	cmd.Flags().String("subscription-start-at-max", "", "Filter by subscription start on/before (ISO 8601)")
	cmd.Flags().String("is-subscription-start-at", "", "Filter by presence of subscription start (true/false)")
	cmd.Flags().String("subscription-end-at-min", "", "Filter by subscription end on/after (ISO 8601)")
	cmd.Flags().String("subscription-end-at-max", "", "Filter by subscription end on/before (ISO 8601)")
	cmd.Flags().String("is-subscription-end-at", "", "Filter by presence of subscription end (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserPostFeedPostsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseUserPostFeedPostsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-post-feed-posts]", "feed-at,score,is-viewed-by-user,is-bookmarked,subscription-start-at,subscription-end-at,user,post,user-post-feed,follow")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[user-id]", opts.UserID)
	setFilterIfPresent(query, "filter[post]", opts.Post)
	setFilterIfPresent(query, "filter[post-type]", opts.PostType)
	setFilterIfPresent(query, "filter[creator]", opts.Creator)
	setFilterIfPresent(query, "filter[score]", opts.Score)
	setFilterIfPresent(query, "filter[feed-at-min]", opts.FeedAtMin)
	setFilterIfPresent(query, "filter[feed-at-max]", opts.FeedAtMax)
	setFilterIfPresent(query, "filter[is-feed-at]", opts.IsFeedAt)
	setFilterIfPresent(query, "filter[subscription-start-at-min]", opts.SubscriptionStartAtMin)
	setFilterIfPresent(query, "filter[subscription-start-at-max]", opts.SubscriptionStartAtMax)
	setFilterIfPresent(query, "filter[is-subscription-start-at]", opts.IsSubscriptionStartAt)
	setFilterIfPresent(query, "filter[subscription-end-at-min]", opts.SubscriptionEndAtMin)
	setFilterIfPresent(query, "filter[subscription-end-at-max]", opts.SubscriptionEndAtMax)
	setFilterIfPresent(query, "filter[is-subscription-end-at]", opts.IsSubscriptionEndAt)

	body, _, err := client.Get(cmd.Context(), "/v1/user-post-feed-posts", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	rows := buildUserPostFeedPostRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderUserPostFeedPostsTable(cmd, rows)
}

func parseUserPostFeedPostsListOptions(cmd *cobra.Command) (userPostFeedPostsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	user, _ := cmd.Flags().GetString("user")
	userID, _ := cmd.Flags().GetString("user-id")
	post, _ := cmd.Flags().GetString("post")
	postType, _ := cmd.Flags().GetString("post-type")
	creator, _ := cmd.Flags().GetString("creator")
	score, _ := cmd.Flags().GetString("score")
	feedAtMin, _ := cmd.Flags().GetString("feed-at-min")
	feedAtMax, _ := cmd.Flags().GetString("feed-at-max")
	isFeedAt, _ := cmd.Flags().GetString("is-feed-at")
	subscriptionStartAtMin, _ := cmd.Flags().GetString("subscription-start-at-min")
	subscriptionStartAtMax, _ := cmd.Flags().GetString("subscription-start-at-max")
	isSubscriptionStartAt, _ := cmd.Flags().GetString("is-subscription-start-at")
	subscriptionEndAtMin, _ := cmd.Flags().GetString("subscription-end-at-min")
	subscriptionEndAtMax, _ := cmd.Flags().GetString("subscription-end-at-max")
	isSubscriptionEndAt, _ := cmd.Flags().GetString("is-subscription-end-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userPostFeedPostsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		User:                   user,
		UserID:                 userID,
		Post:                   post,
		PostType:               postType,
		Creator:                creator,
		Score:                  score,
		FeedAtMin:              feedAtMin,
		FeedAtMax:              feedAtMax,
		IsFeedAt:               isFeedAt,
		SubscriptionStartAtMin: subscriptionStartAtMin,
		SubscriptionStartAtMax: subscriptionStartAtMax,
		IsSubscriptionStartAt:  isSubscriptionStartAt,
		SubscriptionEndAtMin:   subscriptionEndAtMin,
		SubscriptionEndAtMax:   subscriptionEndAtMax,
		IsSubscriptionEndAt:    isSubscriptionEndAt,
	}, nil
}

func buildUserPostFeedPostRows(resp jsonAPIResponse) []userPostFeedPostRow {
	rows := make([]userPostFeedPostRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := userPostFeedPostRow{
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
		}
		rows = append(rows, row)
	}
	return rows
}

func renderUserPostFeedPostsTable(cmd *cobra.Command, rows []userPostFeedPostRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No user post feed posts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tUSER\tPOST\tFEED AT\tSCORE\tBOOKMARKED\tVIEWED")
	for _, row := range rows {
		score := ""
		if row.Score != nil {
			score = fmt.Sprintf("%.2f", *row.Score)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n",
			row.ID,
			row.UserID,
			row.PostID,
			row.FeedAt,
			score,
			row.IsBookmarked,
			row.IsViewedByUser,
		)
	}
	return writer.Flush()
}
