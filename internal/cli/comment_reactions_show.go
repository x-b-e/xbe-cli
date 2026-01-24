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

type commentReactionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type commentReactionDetails struct {
	ID                        string `json:"id"`
	CommentID                 string `json:"comment_id,omitempty"`
	CommentBody               string `json:"comment_body,omitempty"`
	CommentIsAdminOnly        bool   `json:"comment_is_admin_only"`
	ReactionClassificationID  string `json:"reaction_classification_id,omitempty"`
	ReactionLabel             string `json:"reaction_label,omitempty"`
	ReactionUTF8              string `json:"reaction_utf8,omitempty"`
	ReactionExternalReference string `json:"reaction_external_reference,omitempty"`
	CreatedByID               string `json:"created_by_id,omitempty"`
	CreatedByName             string `json:"created_by_name,omitempty"`
	CreatedByEmail            string `json:"created_by_email,omitempty"`
	CreatedAt                 string `json:"created_at,omitempty"`
	UpdatedAt                 string `json:"updated_at,omitempty"`
}

func newCommentReactionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show comment reaction details",
		Long: `Show the full details of a comment reaction.

Output Fields:
  ID                     Resource identifier
  Comment                Comment ID
  Comment Body           Comment text
  Comment Admin Only     Whether the comment is admin-only
  Reaction               Reaction label/emoji and ID
  Reaction Label         Reaction label/name
  Reaction Emoji         Reaction emoji
  Reaction External Ref  Reaction external reference
  Created By             User who reacted
  Created At             Reaction creation time
  Updated At             Reaction last update time

Arguments:
  <id>  The comment reaction ID (required).`,
		Example: `  # Show a comment reaction
  xbe view comment-reactions show 123

  # Output as JSON
  xbe view comment-reactions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runCommentReactionsShow,
	}
	initCommentReactionsShowFlags(cmd)
	return cmd
}

func init() {
	commentReactionsCmd.AddCommand(newCommentReactionsShowCmd())
}

func initCommentReactionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommentReactionsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseCommentReactionsShowOptions(cmd)
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
		return fmt.Errorf("comment reaction id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[comment-reactions]", "comment,created-by,reaction-classification,created-at,updated-at")
	query.Set("include", "comment,created-by,reaction-classification")
	query.Set("fields[comments]", "body,is-admin-only")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[reaction-classifications]", "label,utf8,external-reference")

	body, _, err := client.Get(cmd.Context(), "/v1/comment-reactions/"+id, query)
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

	details := buildCommentReactionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderCommentReactionDetails(cmd, details)
}

func parseCommentReactionsShowOptions(cmd *cobra.Command) (commentReactionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commentReactionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildCommentReactionDetails(resp jsonAPISingleResponse) commentReactionDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := commentReactionDetails{
		ID:        resp.Data.ID,
		CreatedAt: formatDateTime(stringAttr(resp.Data.Attributes, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(resp.Data.Attributes, "updated-at")),
	}

	if rel, ok := resp.Data.Relationships["comment"]; ok && rel.Data != nil {
		details.CommentID = rel.Data.ID
		if comment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CommentBody = stringAttr(comment.Attributes, "body")
			details.CommentIsAdminOnly = boolAttr(comment.Attributes, "is-admin-only")
		}
	}

	if rel, ok := resp.Data.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CreatedByName = stringAttr(user.Attributes, "name")
			details.CreatedByEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	if rel, ok := resp.Data.Relationships["reaction-classification"]; ok && rel.Data != nil {
		details.ReactionClassificationID = rel.Data.ID
		if reaction, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ReactionLabel = stringAttr(reaction.Attributes, "label")
			details.ReactionUTF8 = stringAttr(reaction.Attributes, "utf8")
			details.ReactionExternalReference = stringAttr(reaction.Attributes, "external-reference")
		}
	}

	return details
}

func renderCommentReactionDetails(cmd *cobra.Command, details commentReactionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CommentID != "" {
		fmt.Fprintf(out, "Comment: %s\n", details.CommentID)
	}
	if details.CommentBody != "" {
		fmt.Fprintf(out, "Comment Body: %s\n", details.CommentBody)
	}
	if details.CommentIsAdminOnly {
		fmt.Fprintln(out, "Comment Admin Only: yes")
	}

	if details.ReactionClassificationID != "" || details.ReactionLabel != "" || details.ReactionUTF8 != "" {
		reactionLabel := firstNonEmpty(details.ReactionLabel, details.ReactionClassificationID)
		if details.ReactionUTF8 != "" {
			reactionLabel = strings.TrimSpace(details.ReactionUTF8 + " " + reactionLabel)
		}
		fmt.Fprintf(out, "Reaction: %s\n", formatRelated(reactionLabel, details.ReactionClassificationID))
	}
	if details.ReactionLabel != "" {
		fmt.Fprintf(out, "Reaction Label: %s\n", details.ReactionLabel)
	}
	if details.ReactionUTF8 != "" {
		fmt.Fprintf(out, "Reaction Emoji: %s\n", details.ReactionUTF8)
	}
	if details.ReactionExternalReference != "" {
		fmt.Fprintf(out, "Reaction External Ref: %s\n", details.ReactionExternalReference)
	}

	if details.CreatedByID != "" || details.CreatedByName != "" || details.CreatedByEmail != "" {
		label := firstNonEmpty(details.CreatedByName, details.CreatedByEmail)
		fmt.Fprintf(out, "Created By: %s\n", formatRelated(label, details.CreatedByID))
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
