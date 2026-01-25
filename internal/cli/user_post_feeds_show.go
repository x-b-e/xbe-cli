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

type userPostFeedsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type userPostFeedDetails struct {
	ID                      string   `json:"id"`
	UserID                  string   `json:"user_id,omitempty"`
	UserName                string   `json:"user_name,omitempty"`
	UserEmail               string   `json:"user_email,omitempty"`
	EnableVectorIndexing    bool     `json:"enable_vector_indexing"`
	OpenAIVectorStoreOpenAI string   `json:"open_ai_vector_store_open_ai_id,omitempty"`
	CreatedAt               string   `json:"created_at,omitempty"`
	UpdatedAt               string   `json:"updated_at,omitempty"`
	UserPostFeedPostIDs     []string `json:"user_post_feed_post_ids,omitempty"`
}

func newUserPostFeedsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show user post feed details",
		Long: `Show the full details of a user post feed.

User post feeds track posts shown in a user's feed and vector indexing settings.

Output Fields:
  ID                       Feed identifier
  User                     User name, email, and ID
  Vector Indexing Enabled  Whether vector indexing is enabled
  OpenAI Vector Store ID   OpenAI vector store identifier
  Created At               Feed creation timestamp
  Updated At               Feed last update timestamp
  Feed Posts               User post feed post IDs

Arguments:
  <id>  The user post feed ID (required). Use the list command to find IDs.`,
		Example: `  # Show a user post feed
  xbe view user-post-feeds show 123

  # Output as JSON
  xbe view user-post-feeds show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUserPostFeedsShow,
	}
	initUserPostFeedsShowFlags(cmd)
	return cmd
}

func init() {
	userPostFeedsCmd.AddCommand(newUserPostFeedsShowCmd())
}

func initUserPostFeedsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserPostFeedsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseUserPostFeedsShowOptions(cmd)
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
		return fmt.Errorf("user post feed id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-post-feeds]", "user,enable-vector-indexing,open-ai-vector-store-open-ai-id,created-at,updated-at,user-post-feed-posts")
	query.Set("include", "user,user-post-feed-posts")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[user-post-feed-posts]", "post,feed-at,is-viewed-by-user")

	body, _, err := client.Get(cmd.Context(), "/v1/user-post-feeds/"+id, query)
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

	details := buildUserPostFeedDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUserPostFeedDetails(cmd, details)
}

func parseUserPostFeedsShowOptions(cmd *cobra.Command) (userPostFeedsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userPostFeedsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUserPostFeedDetails(resp jsonAPISingleResponse) userPostFeedDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := userPostFeedDetails{
		ID:                      resp.Data.ID,
		EnableVectorIndexing:    boolAttr(attrs, "enable-vector-indexing"),
		OpenAIVectorStoreOpenAI: stringAttr(attrs, "open-ai-vector-store-open-ai-id"),
		CreatedAt:               formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt:               formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = stringAttr(user.Attributes, "name")
			details.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	if rel, ok := resp.Data.Relationships["user-post-feed-posts"]; ok {
		details.UserPostFeedPostIDs = relationshipIDList(rel)
	}
	if len(details.UserPostFeedPostIDs) == 0 {
		for _, inc := range resp.Included {
			if inc.Type == "user-post-feed-posts" {
				details.UserPostFeedPostIDs = append(details.UserPostFeedPostIDs, inc.ID)
			}
		}
	}

	return details
}

func renderUserPostFeedDetails(cmd *cobra.Command, details userPostFeedDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.UserID != "" || details.UserName != "" || details.UserEmail != "" {
		userLabel := formatRelated(firstNonEmpty(details.UserName, details.UserEmail), details.UserID)
		fmt.Fprintf(out, "User: %s\n", userLabel)
	}
	fmt.Fprintf(out, "Vector Indexing Enabled: %s\n", formatBoolLabel(details.EnableVectorIndexing))
	if details.OpenAIVectorStoreOpenAI != "" {
		fmt.Fprintf(out, "OpenAI Vector Store ID: %s\n", details.OpenAIVectorStoreOpenAI)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if len(details.UserPostFeedPostIDs) > 0 {
		fmt.Fprintf(out, "Feed Posts: %s\n", strings.Join(details.UserPostFeedPostIDs, ", "))
	}

	return nil
}

func formatBoolLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
