package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type postChildrenListOptions struct {
	BaseURL    string
	Token      string
	JSON       bool
	NoAuth     bool
	Limit      int
	Offset     int
	Sort       string
	ParentPost string
	ChildPost  string
}

type postChildRow struct {
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

func newPostChildrenListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List post child links",
		Long: `List post child links with filtering and pagination.

Output Columns:
  ID      Post child link identifier
  PARENT  Parent post summary or ID
  CHILD   Child post summary or ID

Filters:
  --parent-post  Filter by parent post ID
  --child-post   Filter by child post ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List post child links
  xbe view post-children list

  # Filter by parent post
  xbe view post-children list --parent-post 123

  # Filter by child post
  xbe view post-children list --child-post 456

  # Output as JSON
  xbe view post-children list --json`,
		Args: cobra.NoArgs,
		RunE: runPostChildrenList,
	}
	initPostChildrenListFlags(cmd)
	return cmd
}

func init() {
	postChildrenCmd.AddCommand(newPostChildrenListCmd())
}

func initPostChildrenListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("parent-post", "", "Filter by parent post ID")
	cmd.Flags().String("child-post", "", "Filter by child post ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostChildrenList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePostChildrenListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[post-children]", "parent-post,child-post")
	query.Set("include", "parent-post,child-post")
	query.Set("fields[posts]", "post-type,status,published-at,short-text-content,creator-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[parent-post]", opts.ParentPost)
	setFilterIfPresent(query, "filter[child-post]", opts.ChildPost)

	body, _, err := client.Get(cmd.Context(), "/v1/post-children", query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildPostChildRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPostChildrenTable(cmd, rows)
}

func parsePostChildrenListOptions(cmd *cobra.Command) (postChildrenListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	parentPost, _ := cmd.Flags().GetString("parent-post")
	childPost, _ := cmd.Flags().GetString("child-post")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return postChildrenListOptions{
		BaseURL:    baseURL,
		Token:      token,
		JSON:       jsonOut,
		NoAuth:     noAuth,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		ParentPost: parentPost,
		ChildPost:  childPost,
	}, nil
}

func buildPostChildRows(resp jsonAPIResponse) []postChildRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]postChildRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPostChildRow(resource, included))
	}
	return rows
}

func postChildRowFromSingle(resp jsonAPISingleResponse) postChildRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildPostChildRow(resp.Data, included)
}

func buildPostChildRow(resource jsonAPIResource, included map[string]jsonAPIResource) postChildRow {
	row := postChildRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["parent-post"]; ok && rel.Data != nil {
		row.ParentPostID = rel.Data.ID
		if post, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := post.Attributes
			row.ParentPostType = stringAttr(attrs, "post-type")
			row.ParentPostStatus = stringAttr(attrs, "status")
			row.ParentPostPublishedAt = formatDateTime(stringAttr(attrs, "published-at"))
			row.ParentPostShortText = stringAttr(attrs, "short-text-content")
			row.ParentPostCreatorName = stringAttr(attrs, "creator-name")
		}
	}

	if rel, ok := resource.Relationships["child-post"]; ok && rel.Data != nil {
		row.ChildPostID = rel.Data.ID
		if post, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := post.Attributes
			row.ChildPostType = stringAttr(attrs, "post-type")
			row.ChildPostStatus = stringAttr(attrs, "status")
			row.ChildPostPublishedAt = formatDateTime(stringAttr(attrs, "published-at"))
			row.ChildPostShortText = stringAttr(attrs, "short-text-content")
			row.ChildPostCreatorName = stringAttr(attrs, "creator-name")
		}
	}

	return row
}

func renderPostChildrenTable(cmd *cobra.Command, rows []postChildRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No post child links found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPARENT\tCHILD")
	for _, row := range rows {
		parentLabel := firstNonEmpty(row.ParentPostShortText, row.ParentPostType, row.ParentPostID)
		childLabel := firstNonEmpty(row.ChildPostShortText, row.ChildPostType, row.ChildPostID)
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(parentLabel, 40),
			truncateString(childLabel, 40),
		)
	}
	return writer.Flush()
}
