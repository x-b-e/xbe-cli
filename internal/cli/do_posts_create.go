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

type doPostsCreateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	PostType    string
	TextContent string
	Status      string
	PublishedAt string
	CreatorType string
	CreatorID   string
	SubjectPost string
	IsPrivate   bool
}

func newDoPostsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new post",
		Long: `Create a new post.

Post types:
  basic, notification, action, new_membership, post_summary, post_activity,
  objective_status, objective_completion, job_production_plan_recap, proffer

Status values:
  draft, published

Required flags:
  --post-type       Post type
  --text-content    Post content text

Optional flags:
  --status          Post status (draft/published)
  --published-at    Publication date (ISO 8601)
  --creator-type    Creator type (users, brokers, etc.)
  --creator-id      Creator ID
  --subject-post    Subject post ID (for replies)
  --private         Mark post as private`,
		Example: `  # Create a basic post
  xbe do posts create --post-type basic --text-content "Hello world"

  # Create a published post
  xbe do posts create --post-type basic --text-content "Published post" --status published

  # Create a post as a specific user
  xbe do posts create --post-type basic --text-content "Content" --creator-type users --creator-id 123`,
		RunE: runDoPostsCreate,
	}
	initDoPostsCreateFlags(cmd)
	return cmd
}

func init() {
	doPostsCmd.AddCommand(newDoPostsCreateCmd())
}

func initDoPostsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("post-type", "", "Post type (required)")
	cmd.Flags().String("text-content", "", "Post content text (required)")
	cmd.Flags().String("status", "", "Post status (draft/published)")
	cmd.Flags().String("published-at", "", "Publication date (ISO 8601)")
	cmd.Flags().String("creator-type", "", "Creator type (users, brokers, etc.)")
	cmd.Flags().String("creator-id", "", "Creator ID")
	cmd.Flags().String("subject-post", "", "Subject post ID (for replies)")
	cmd.Flags().Bool("private", false, "Mark post as private")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")

	cmd.MarkFlagRequired("post-type")
	cmd.MarkFlagRequired("text-content")
}

func runDoPostsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoPostsCreateOptions(cmd)
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

	attributes := map[string]any{
		"post-type":    opts.PostType,
		"text-content": opts.TextContent,
	}

	if opts.Status != "" {
		attributes["status"] = opts.Status
	}
	if opts.PublishedAt != "" {
		attributes["published-at"] = opts.PublishedAt
	}
	if cmd.Flags().Changed("private") {
		attributes["is-private"] = opts.IsPrivate
	}

	relationships := map[string]any{}

	if opts.CreatorType != "" && opts.CreatorID != "" {
		relationships["creator"] = map[string]any{
			"data": map[string]any{
				"type": opts.CreatorType,
				"id":   opts.CreatorID,
			},
		}
	}

	if opts.SubjectPost != "" {
		relationships["subject-post"] = map[string]any{
			"data": map[string]any{
				"type": "posts",
				"id":   opts.SubjectPost,
			},
		}
	}

	data := map[string]any{
		"type":       "posts",
		"attributes": attributes,
	}

	if len(relationships) > 0 {
		data["relationships"] = relationships
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/posts", jsonBody)
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

	row := postRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created post %s (%s)\n", row.ID, row.PostType)
	return nil
}

func parseDoPostsCreateOptions(cmd *cobra.Command) (doPostsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	postType, _ := cmd.Flags().GetString("post-type")
	textContent, _ := cmd.Flags().GetString("text-content")
	status, _ := cmd.Flags().GetString("status")
	publishedAt, _ := cmd.Flags().GetString("published-at")
	creatorType, _ := cmd.Flags().GetString("creator-type")
	creatorID, _ := cmd.Flags().GetString("creator-id")
	subjectPost, _ := cmd.Flags().GetString("subject-post")
	isPrivate, _ := cmd.Flags().GetBool("private")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPostsCreateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		PostType:    postType,
		TextContent: textContent,
		Status:      status,
		PublishedAt: publishedAt,
		CreatorType: creatorType,
		CreatorID:   creatorID,
		SubjectPost: subjectPost,
		IsPrivate:   isPrivate,
	}, nil
}

func postRowFromSingle(resp jsonAPISingleResponse) postRow {
	return postRow{
		ID:        resp.Data.ID,
		PostType:  stringAttr(resp.Data.Attributes, "post-type"),
		Status:    stringAttr(resp.Data.Attributes, "status"),
		Published: formatDate(stringAttr(resp.Data.Attributes, "published-at")),
		Content:   strings.TrimSpace(stringAttr(resp.Data.Attributes, "short-text-content")),
	}
}
