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

type userCreatorFeedsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type userCreatorFeedDetails struct {
	ID                        string   `json:"id"`
	UserID                    string   `json:"user_id,omitempty"`
	UserName                  string   `json:"user_name,omitempty"`
	UserEmail                 string   `json:"user_email,omitempty"`
	CreatedAt                 string   `json:"created_at,omitempty"`
	UpdatedAt                 string   `json:"updated_at,omitempty"`
	UserCreatorFeedCreatorIDs []string `json:"user_creator_feed_creator_ids,omitempty"`
}

func newUserCreatorFeedsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show user creator feed details",
		Long: `Show the full details of a user creator feed.

User creator feeds track which creators appear in a user's creator feed.

Output Fields:
  ID            Feed identifier
  User          User name, email, and ID
  Created At    Feed creation timestamp
  Updated At    Feed last update timestamp
  Feed Creators User creator feed creator IDs

Arguments:
  <id>  The user creator feed ID (required). Use the list command to find IDs.`,
		Example: `  # Show a user creator feed
  xbe view user-creator-feeds show 123

  # Output as JSON
  xbe view user-creator-feeds show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUserCreatorFeedsShow,
	}
	initUserCreatorFeedsShowFlags(cmd)
	return cmd
}

func init() {
	userCreatorFeedsCmd.AddCommand(newUserCreatorFeedsShowCmd())
}

func initUserCreatorFeedsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserCreatorFeedsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseUserCreatorFeedsShowOptions(cmd)
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
		return fmt.Errorf("user creator feed id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-creator-feeds]", "user,created-at,updated-at,user-creator-feed-creators")
	query.Set("include", "user,user-creator-feed-creators")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[user-creator-feed-creators]", "order,creator-name,creator-parents,creator-avatar-url,creator,follow")

	body, _, err := client.Get(cmd.Context(), "/v1/user-creator-feeds/"+id, query)
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

	details := buildUserCreatorFeedDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUserCreatorFeedDetails(cmd, details)
}

func parseUserCreatorFeedsShowOptions(cmd *cobra.Command) (userCreatorFeedsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userCreatorFeedsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUserCreatorFeedDetails(resp jsonAPISingleResponse) userCreatorFeedDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := userCreatorFeedDetails{
		ID:        resp.Data.ID,
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.UserName = stringAttr(user.Attributes, "name")
			details.UserEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	if rel, ok := resp.Data.Relationships["user-creator-feed-creators"]; ok {
		details.UserCreatorFeedCreatorIDs = relationshipIDList(rel)
	}
	if len(details.UserCreatorFeedCreatorIDs) == 0 {
		for _, inc := range resp.Included {
			if inc.Type == "user-creator-feed-creators" {
				details.UserCreatorFeedCreatorIDs = append(details.UserCreatorFeedCreatorIDs, inc.ID)
			}
		}
	}

	return details
}

func renderUserCreatorFeedDetails(cmd *cobra.Command, details userCreatorFeedDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.UserID != "" || details.UserName != "" || details.UserEmail != "" {
		userLabel := formatRelated(firstNonEmpty(details.UserName, details.UserEmail), details.UserID)
		fmt.Fprintf(out, "User: %s\n", userLabel)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}
	if len(details.UserCreatorFeedCreatorIDs) > 0 {
		fmt.Fprintf(out, "Feed Creators: %s\n", strings.Join(details.UserCreatorFeedCreatorIDs, ", "))
	}

	return nil
}
