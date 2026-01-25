package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type postsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type postDetails struct {
	ID          string `json:"id"`
	PostType    string `json:"post_type"`
	Status      string `json:"status"`
	Published   string `json:"published"`
	Creator     string `json:"creator"`
	CreatorType string `json:"creator_type"`
	Content     string `json:"content"`
}

func newPostsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show post details",
		Long: `Show the full details of a specific post.

Retrieves and displays comprehensive information about a post including
its full content and metadata.

Output Fields:
  ID              Unique post identifier
  Type            The type of post
  Status          Publication status (draft/published)
  Published       Publication date
  Creator         Name of the creator
  Content         Full post content

Arguments:
  <id>            The post ID (required). You can find IDs using the list command.`,
		Example: `  # View a post by ID
  xbe view posts show 123

  # Get post as JSON
  xbe view posts show 123 --json

  # View without authentication
  xbe view posts show 123 --no-auth`,
		Args: cobra.ExactArgs(1),
		RunE: runPostsShow,
	}
	initPostsShowFlags(cmd)
	return cmd
}

func init() {
	postsCmd.AddCommand(newPostsShowCmd())
}

func initPostsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePostsShowOptions(cmd)
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
		return fmt.Errorf("post id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[posts]", "post-type,published-at,status,text-content,short-text-content,creator-name,creator")
	query.Set("include", "creator")

	body, _, err := client.Get(cmd.Context(), "/v1/posts/"+id, query)
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

	details := buildPostDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPostDetails(cmd, details)
}

func parsePostsShowOptions(cmd *cobra.Command) (postsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return postsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return postsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return postsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return postsShowOptions{}, err
	}

	return postsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPostDetails(resp jsonAPISingleResponse) postDetails {
	attrs := resp.Data.Attributes

	// Prefer full text-content, fall back to short-text-content
	content := strings.TrimSpace(stringAttr(attrs, "text-content"))
	if content == "" {
		content = strings.TrimSpace(stringAttr(attrs, "short-text-content"))
	}

	// Get creator type from relationship
	creatorType := ""
	if rel, ok := resp.Data.Relationships["creator"]; ok && rel.Data != nil {
		creatorType = rel.Data.Type
	}

	return postDetails{
		ID:          resp.Data.ID,
		PostType:    stringAttr(attrs, "post-type"),
		Status:      stringAttr(attrs, "status"),
		Published:   formatDate(stringAttr(attrs, "published-at")),
		Creator:     strings.TrimSpace(stringAttr(attrs, "creator-name")),
		CreatorType: creatorType,
		Content:     content,
	}
}

func renderPostDetails(cmd *cobra.Command, details postDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.PostType != "" {
		fmt.Fprintf(out, "Type: %s\n", details.PostType)
	}
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	if details.Published != "" {
		fmt.Fprintf(out, "Published: %s\n", details.Published)
	}
	if details.Creator != "" {
		fmt.Fprintf(out, "Creator: %s\n", details.Creator)
	}
	if details.Content != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Content:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, cleanMarkdownFull(details.Content))
	}

	return nil
}

var (
	showHeaderRegex    = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	showBoldRegex      = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	showItalicRegex    = regexp.MustCompile(`\*([^*]+)\*`)
	showLinkRegex      = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	showTableRowRegex  = regexp.MustCompile(`(?m)^\|.+\|$`)
	showTableSepRegex  = regexp.MustCompile(`(?m)^\|[\s\-:│|]+`)
	showDashLineRegex  = regexp.MustCompile(`(?m)^[\s\-:│|]+$`)
	showMultiNewline   = regexp.MustCompile(`\n{4,}`)
	showMultiSpace     = regexp.MustCompile(`[ \t]+`)
	showLeadingNewline = regexp.MustCompile(`^\n+`)
)

// cleanMarkdownFull cleans markdown for full content display, preserving line structure
func cleanMarkdownFull(s string) string {
	// Decode HTML entities
	s = decodeHTMLEntities(s)
	// Remove entire markdown table rows (|...|)
	s = showTableRowRegex.ReplaceAllString(s, "")
	// Remove table separator lines (|---|---|...)
	s = showTableSepRegex.ReplaceAllString(s, "")
	// Remove markdown headers (# ## ### etc)
	s = showHeaderRegex.ReplaceAllString(s, "")
	// Convert **bold** to just the text
	s = showBoldRegex.ReplaceAllString(s, "$1")
	// Convert *italic* to just the text
	s = showItalicRegex.ReplaceAllString(s, "$1")
	// Convert [text](url) to just text
	s = showLinkRegex.ReplaceAllString(s, "$1")
	// Replace pipe separators with bullet
	s = strings.ReplaceAll(s, " | ", " • ")
	s = strings.ReplaceAll(s, "| ", "")
	s = strings.ReplaceAll(s, " |", "")
	// Collapse multiple spaces/tabs to single space
	s = showMultiSpace.ReplaceAllString(s, " ")
	// Remove lines that are just dashes/spaces
	s = showDashLineRegex.ReplaceAllString(s, "")
	// Collapse 4+ newlines to 3 (keep some paragraph spacing)
	s = showMultiNewline.ReplaceAllString(s, "\n\n\n")
	// Remove leading newlines
	s = showLeadingNewline.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
