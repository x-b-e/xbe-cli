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

type doPostsUpdateOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	ID          string
	TextContent string
	Status      string
	PublishedAt string
	IsPrivate   bool
}

func newDoPostsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a post",
		Long: `Update an existing post.

Note: post-type, creator, and subject-post cannot be changed after creation.

Optional flags:
  --text-content    Post content text
  --status          Post status (draft/published)
  --published-at    Publication date (ISO 8601)
  --private         Mark post as private`,
		Example: `  # Update post content
  xbe do posts update 123 --text-content "Updated content"

  # Publish a draft post
  xbe do posts update 123 --status published`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPostsUpdate,
	}
	initDoPostsUpdateFlags(cmd)
	return cmd
}

func init() {
	doPostsCmd.AddCommand(newDoPostsUpdateCmd())
}

func initDoPostsUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("text-content", "", "Post content text")
	cmd.Flags().String("status", "", "Post status (draft/published)")
	cmd.Flags().String("published-at", "", "Publication date (ISO 8601)")
	cmd.Flags().Bool("private", false, "Mark post as private")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPostsUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPostsUpdateOptions(cmd, args)
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

	if cmd.Flags().Changed("text-content") {
		attributes["text-content"] = opts.TextContent
	}
	if cmd.Flags().Changed("status") {
		attributes["status"] = opts.Status
	}
	if cmd.Flags().Changed("published-at") {
		attributes["published-at"] = opts.PublishedAt
	}
	if cmd.Flags().Changed("private") {
		attributes["is-private"] = opts.IsPrivate
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	data := map[string]any{
		"type":       "posts",
		"id":         opts.ID,
		"attributes": attributes,
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

	path := fmt.Sprintf("/v1/posts/%s", opts.ID)
	body, _, err := client.Patch(cmd.Context(), path, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated post %s (%s)\n", row.ID, row.PostType)
	return nil
}

func parseDoPostsUpdateOptions(cmd *cobra.Command, args []string) (doPostsUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	textContent, _ := cmd.Flags().GetString("text-content")
	status, _ := cmd.Flags().GetString("status")
	publishedAt, _ := cmd.Flags().GetString("published-at")
	isPrivate, _ := cmd.Flags().GetBool("private")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPostsUpdateOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		ID:          args[0],
		TextContent: textContent,
		Status:      status,
		PublishedAt: publishedAt,
		IsPrivate:   isPrivate,
	}, nil
}
