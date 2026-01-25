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

type postChildrenShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type postChildDetails struct {
	ID                    string `json:"id"`
	ParentPostID          string `json:"parent_post_id,omitempty"`
	ParentPostType        string `json:"parent_post_type,omitempty"`
	ParentPostStatus      string `json:"parent_post_status,omitempty"`
	ParentPostPublishedAt string `json:"parent_post_published_at,omitempty"`
	ParentPostShortText   string `json:"parent_post_short_text,omitempty"`
	ParentPostCreatorName string `json:"parent_post_creator_name,omitempty"`
	ChildPostID           string `json:"child_post_id,omitempty"`
	ChildPostType         string `json:"child_post_type,omitempty"`
	ChildPostStatus       string `json:"child_post_status,omitempty"`
	ChildPostPublishedAt  string `json:"child_post_published_at,omitempty"`
	ChildPostShortText    string `json:"child_post_short_text,omitempty"`
	ChildPostCreatorName  string `json:"child_post_creator_name,omitempty"`
}

func newPostChildrenShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show post child link details",
		Long: `Show the full details of a post child link.

Output Fields:
  ID                     Resource identifier
  Parent Post            Parent post summary and ID
  Parent Post Type       Parent post type
  Parent Post Status     Parent post status
  Parent Post Published  Parent post published timestamp
  Parent Post Creator    Parent post creator name
  Parent Post Summary    Parent post short text
  Child Post             Child post summary and ID
  Child Post Type        Child post type
  Child Post Status      Child post status
  Child Post Published   Child post published timestamp
  Child Post Creator     Child post creator name
  Child Post Summary     Child post short text

Arguments:
  <id>  The post child ID (required).`,
		Example: `  # Show a post child link
  xbe view post-children show 123

  # Output as JSON
  xbe view post-children show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPostChildrenShow,
	}
	initPostChildrenShowFlags(cmd)
	return cmd
}

func init() {
	postChildrenCmd.AddCommand(newPostChildrenShowCmd())
}

func initPostChildrenShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostChildrenShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePostChildrenShowOptions(cmd)
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
		return fmt.Errorf("post child id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[post-children]", "parent-post,child-post")
	query.Set("include", "parent-post,child-post")
	query.Set("fields[posts]", "post-type,status,published-at,short-text-content,creator-name")

	body, _, err := client.Get(cmd.Context(), "/v1/post-children/"+id, query)
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

	details := buildPostChildDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPostChildDetails(cmd, details)
}

func parsePostChildrenShowOptions(cmd *cobra.Command) (postChildrenShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return postChildrenShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPostChildDetails(resp jsonAPISingleResponse) postChildDetails {
	row := postChildRowFromSingle(resp)
	return postChildDetails{
		ID:                    row.ID,
		ParentPostID:          row.ParentPostID,
		ParentPostType:        row.ParentPostType,
		ParentPostStatus:      row.ParentPostStatus,
		ParentPostPublishedAt: row.ParentPostPublishedAt,
		ParentPostShortText:   row.ParentPostShortText,
		ParentPostCreatorName: row.ParentPostCreatorName,
		ChildPostID:           row.ChildPostID,
		ChildPostType:         row.ChildPostType,
		ChildPostStatus:       row.ChildPostStatus,
		ChildPostPublishedAt:  row.ChildPostPublishedAt,
		ChildPostShortText:    row.ChildPostShortText,
		ChildPostCreatorName:  row.ChildPostCreatorName,
	}
}

func renderPostChildDetails(cmd *cobra.Command, details postChildDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	if details.ParentPostID != "" || details.ParentPostShortText != "" || details.ParentPostType != "" {
		label := firstNonEmpty(details.ParentPostShortText, details.ParentPostType)
		fmt.Fprintf(out, "Parent Post: %s\n", formatRelated(label, details.ParentPostID))
	}
	if details.ParentPostType != "" {
		fmt.Fprintf(out, "Parent Post Type: %s\n", details.ParentPostType)
	}
	if details.ParentPostStatus != "" {
		fmt.Fprintf(out, "Parent Post Status: %s\n", details.ParentPostStatus)
	}
	if details.ParentPostPublishedAt != "" {
		fmt.Fprintf(out, "Parent Post Published: %s\n", details.ParentPostPublishedAt)
	}
	if details.ParentPostCreatorName != "" {
		fmt.Fprintf(out, "Parent Post Creator: %s\n", details.ParentPostCreatorName)
	}
	if details.ParentPostShortText != "" {
		fmt.Fprintf(out, "Parent Post Summary: %s\n", details.ParentPostShortText)
	}

	if details.ChildPostID != "" || details.ChildPostShortText != "" || details.ChildPostType != "" {
		label := firstNonEmpty(details.ChildPostShortText, details.ChildPostType)
		fmt.Fprintf(out, "Child Post: %s\n", formatRelated(label, details.ChildPostID))
	}
	if details.ChildPostType != "" {
		fmt.Fprintf(out, "Child Post Type: %s\n", details.ChildPostType)
	}
	if details.ChildPostStatus != "" {
		fmt.Fprintf(out, "Child Post Status: %s\n", details.ChildPostStatus)
	}
	if details.ChildPostPublishedAt != "" {
		fmt.Fprintf(out, "Child Post Published: %s\n", details.ChildPostPublishedAt)
	}
	if details.ChildPostCreatorName != "" {
		fmt.Fprintf(out, "Child Post Creator: %s\n", details.ChildPostCreatorName)
	}
	if details.ChildPostShortText != "" {
		fmt.Fprintf(out, "Child Post Summary: %s\n", details.ChildPostShortText)
	}

	return nil
}
