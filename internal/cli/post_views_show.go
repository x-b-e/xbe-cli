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

type postViewsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type postViewDetails struct {
	ID              string `json:"id"`
	PostID          string `json:"post_id,omitempty"`
	PostType        string `json:"post_type,omitempty"`
	PostStatus      string `json:"post_status,omitempty"`
	PostPublishedAt string `json:"post_published_at,omitempty"`
	PostShortText   string `json:"post_short_text,omitempty"`
	PostCreatorName string `json:"post_creator_name,omitempty"`
	ViewerID        string `json:"viewer_id,omitempty"`
	ViewerName      string `json:"viewer_name,omitempty"`
	ViewerEmail     string `json:"viewer_email,omitempty"`
	ViewedAt        string `json:"viewed_at,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

func newPostViewsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show post view details",
		Long: `Show the full details of a post view.

Output Fields:
  ID              Resource identifier
  Post            Post summary and ID
  Post Type       Post type
  Post Status     Post status
  Post Published  Post published timestamp
  Post Creator    Post creator name
  Post Summary    Post short text
  Viewer          Viewer name or ID
  Viewer Name     Viewer name
  Viewer Email    Viewer email
  Viewed At       View timestamp
  Created At      Post view creation time
  Updated At      Post view last update time

Arguments:
  <id>  The post view ID (required).`,
		Example: `  # Show a post view
  xbe view post-views show 123

  # Output as JSON
  xbe view post-views show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPostViewsShow,
	}
	initPostViewsShowFlags(cmd)
	return cmd
}

func init() {
	postViewsCmd.AddCommand(newPostViewsShowCmd())
}

func initPostViewsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostViewsShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePostViewsShowOptions(cmd)
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
		return fmt.Errorf("post view id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[post-views]", "post,viewer,viewed-at,created-at,updated-at")
	query.Set("include", "post,viewer")
	query.Set("fields[posts]", "post-type,status,published-at,short-text-content,creator-name")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/post-views/"+id, query)
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

	details := buildPostViewDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPostViewDetails(cmd, details)
}

func parsePostViewsShowOptions(cmd *cobra.Command) (postViewsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return postViewsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPostViewDetails(resp jsonAPISingleResponse) postViewDetails {
	row := postViewRowFromSingle(resp)
	return postViewDetails{
		ID:              row.ID,
		PostID:          row.PostID,
		PostType:        row.PostType,
		PostStatus:      row.PostStatus,
		PostPublishedAt: row.PostPublishedAt,
		PostShortText:   row.PostShortText,
		PostCreatorName: row.PostCreatorName,
		ViewerID:        row.ViewerID,
		ViewerName:      row.ViewerName,
		ViewerEmail:     row.ViewerEmail,
		ViewedAt:        row.ViewedAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func renderPostViewDetails(cmd *cobra.Command, details postViewDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)

	if details.PostID != "" || details.PostShortText != "" || details.PostType != "" {
		label := firstNonEmpty(details.PostShortText, details.PostType)
		fmt.Fprintf(out, "Post: %s\n", formatRelated(label, details.PostID))
	}
	if details.PostType != "" {
		fmt.Fprintf(out, "Post Type: %s\n", details.PostType)
	}
	if details.PostStatus != "" {
		fmt.Fprintf(out, "Post Status: %s\n", details.PostStatus)
	}
	if details.PostPublishedAt != "" {
		fmt.Fprintf(out, "Post Published: %s\n", details.PostPublishedAt)
	}
	if details.PostCreatorName != "" {
		fmt.Fprintf(out, "Post Creator: %s\n", details.PostCreatorName)
	}
	if details.PostShortText != "" {
		fmt.Fprintf(out, "Post Summary: %s\n", details.PostShortText)
	}

	if details.ViewerID != "" || details.ViewerName != "" || details.ViewerEmail != "" {
		label := firstNonEmpty(details.ViewerName, details.ViewerEmail)
		fmt.Fprintf(out, "Viewer: %s\n", formatRelated(label, details.ViewerID))
	}
	if details.ViewerName != "" {
		fmt.Fprintf(out, "Viewer Name: %s\n", details.ViewerName)
	}
	if details.ViewerEmail != "" {
		fmt.Fprintf(out, "Viewer Email: %s\n", details.ViewerEmail)
	}

	if details.ViewedAt != "" {
		fmt.Fprintf(out, "Viewed At: %s\n", details.ViewedAt)
	}
	if details.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", details.CreatedAt)
	}
	if details.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", details.UpdatedAt)
	}

	return nil
}
