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

type userCreatorFeedCreatorsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type userCreatorFeedCreatorDetails struct {
	ID                string `json:"id"`
	Order             int    `json:"order,omitempty"`
	CreatorName       string `json:"creator_name,omitempty"`
	CreatorType       string `json:"creator_type,omitempty"`
	CreatorID         string `json:"creator_id,omitempty"`
	CreatorAvatarURL  string `json:"creator_avatar_url,omitempty"`
	CreatorParents    any    `json:"creator_parents,omitempty"`
	UserCreatorFeedID string `json:"user_creator_feed_id,omitempty"`
	UserID            string `json:"user_id,omitempty"`
	FollowID          string `json:"follow_id,omitempty"`
}

func newUserCreatorFeedCreatorsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show user creator feed creator details",
		Long: `Show the full details of a user creator feed creator.

Output Fields:
  ID                User creator feed creator identifier
  Order             Position in the user creator feed
  Creator Name      Creator display name
  Creator           Creator type and ID
  Creator Avatar URL Creator avatar URL (if available)
  Creator Parents   Creator parent references
  User Creator Feed User creator feed ID
  User              User ID
  Follow            Follow ID (if present)

Arguments:
  <id>    User creator feed creator ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a user creator feed creator
  xbe view user-creator-feed-creators show 123

  # JSON output
  xbe view user-creator-feed-creators show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runUserCreatorFeedCreatorsShow,
	}
	initUserCreatorFeedCreatorsShowFlags(cmd)
	return cmd
}

func init() {
	userCreatorFeedCreatorsCmd.AddCommand(newUserCreatorFeedCreatorsShowCmd())
}

func initUserCreatorFeedCreatorsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runUserCreatorFeedCreatorsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseUserCreatorFeedCreatorsShowOptions(cmd)
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
		return fmt.Errorf("user creator feed creator id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[user-creator-feed-creators]", "order,creator-name,creator-parents,creator-avatar-url,creator,user,user-creator-feed,follow")

	body, _, err := client.Get(cmd.Context(), "/v1/user-creator-feed-creators/"+id, query)
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

	details := buildUserCreatorFeedCreatorDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderUserCreatorFeedCreatorDetails(cmd, details)
}

func parseUserCreatorFeedCreatorsShowOptions(cmd *cobra.Command) (userCreatorFeedCreatorsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return userCreatorFeedCreatorsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildUserCreatorFeedCreatorDetails(resp jsonAPISingleResponse) userCreatorFeedCreatorDetails {
	attrs := resp.Data.Attributes
	details := userCreatorFeedCreatorDetails{
		ID:               resp.Data.ID,
		Order:            intAttr(attrs, "order"),
		CreatorName:      strings.TrimSpace(stringAttr(attrs, "creator-name")),
		CreatorAvatarURL: strings.TrimSpace(stringAttr(attrs, "creator-avatar-url")),
		CreatorParents:   attrs["creator-parents"],
	}

	if rel, ok := resp.Data.Relationships["creator"]; ok && rel.Data != nil {
		details.CreatorType = rel.Data.Type
		details.CreatorID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["user-creator-feed"]; ok && rel.Data != nil {
		details.UserCreatorFeedID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["follow"]; ok && rel.Data != nil {
		details.FollowID = rel.Data.ID
	}

	return details
}

func renderUserCreatorFeedCreatorDetails(cmd *cobra.Command, details userCreatorFeedCreatorDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	fmt.Fprintf(out, "Order: %d\n", details.Order)
	fmt.Fprintf(out, "Creator Name: %s\n", formatOptional(details.CreatorName))
	creatorReference := formatCreatorReference(details.CreatorType, details.CreatorID)
	if creatorReference != "" {
		fmt.Fprintf(out, "Creator: %s\n", creatorReference)
	}
	if details.CreatorAvatarURL != "" {
		fmt.Fprintf(out, "Creator Avatar URL: %s\n", formatOptional(details.CreatorAvatarURL))
	}
	if details.UserCreatorFeedID != "" {
		fmt.Fprintf(out, "User Creator Feed: %s\n", details.UserCreatorFeedID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User: %s\n", details.UserID)
	}
	if details.FollowID != "" {
		fmt.Fprintf(out, "Follow: %s\n", details.FollowID)
	}

	if details.CreatorParents != nil {
		fmt.Fprintln(out, "Creator Parents:")
		formatted := formatAny(details.CreatorParents)
		if formatted == "" {
			fmt.Fprintln(out, "  (none)")
		} else {
			fmt.Fprintln(out, indentLines(formatted, "  "))
		}
	}

	return nil
}
