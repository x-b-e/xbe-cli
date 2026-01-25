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

type doUserPostFeedPostsUpdateOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	ID                  string
	IsBookmarked        bool
	SubscriptionStartAt string
	SubscriptionEndAt   string
}

func newDoUserPostFeedPostsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a user post feed post",
		Long: `Update an existing user post feed post.

Only the fields you specify will be updated. Fields not provided remain unchanged.

Arguments:
  <id>    The user post feed post ID (required)

Flags:
  --is-bookmarked          Set whether the post is bookmarked (true/false)
  --subscription-start-at  Subscription start timestamp (ISO 8601, empty to clear)
  --subscription-end-at    Subscription end timestamp (ISO 8601, empty to clear)

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Bookmark a feed post
  xbe do user-post-feed-posts update 123 --is-bookmarked=true

  # Set subscription window
  xbe do user-post-feed-posts update 123 \
    --subscription-start-at 2025-01-01T00:00:00Z \
    --subscription-end-at 2025-01-31T23:59:59Z

  # Get JSON output
  xbe do user-post-feed-posts update 123 --is-bookmarked=true --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoUserPostFeedPostsUpdate,
	}
	initDoUserPostFeedPostsUpdateFlags(cmd)
	return cmd
}

func init() {
	doUserPostFeedPostsCmd.AddCommand(newDoUserPostFeedPostsUpdateCmd())
}

func initDoUserPostFeedPostsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("is-bookmarked", false, "Whether the post is bookmarked")
	cmd.Flags().String("subscription-start-at", "", "Subscription start timestamp (ISO 8601, empty to clear)")
	cmd.Flags().String("subscription-end-at", "", "Subscription end timestamp (ISO 8601, empty to clear)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoUserPostFeedPostsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoUserPostFeedPostsUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("is-bookmarked") {
		attributes["is-bookmarked"] = opts.IsBookmarked
	}
	if cmd.Flags().Changed("subscription-start-at") {
		if strings.TrimSpace(opts.SubscriptionStartAt) == "" {
			attributes["subscription-start-at"] = nil
		} else {
			attributes["subscription-start-at"] = opts.SubscriptionStartAt
		}
	}
	if cmd.Flags().Changed("subscription-end-at") {
		if strings.TrimSpace(opts.SubscriptionEndAt) == "" {
			attributes["subscription-end-at"] = nil
		} else {
			attributes["subscription-end-at"] = opts.SubscriptionEndAt
		}
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         opts.ID,
			"type":       "user-post-feed-posts",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)
	body, _, err := client.Patch(cmd.Context(), "/v1/user-post-feed-posts/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated user post feed post %s\n", details.ID)
	return nil
}

func parseDoUserPostFeedPostsUpdateOptions(cmd *cobra.Command, args []string) (doUserPostFeedPostsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	isBookmarked, _ := cmd.Flags().GetBool("is-bookmarked")
	subscriptionStartAt, _ := cmd.Flags().GetString("subscription-start-at")
	subscriptionEndAt, _ := cmd.Flags().GetString("subscription-end-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doUserPostFeedPostsUpdateOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		ID:                  args[0],
		IsBookmarked:        isBookmarked,
		SubscriptionStartAt: subscriptionStartAt,
		SubscriptionEndAt:   subscriptionEndAt,
	}, nil
}
