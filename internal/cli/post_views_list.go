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

type postViewsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	Post         string
	Viewer       string
	ViewedAtMin  string
	ViewedAtMax  string
	IsViewedAt   string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type postViewRow struct {
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

func newPostViewsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List post views",
		Long: `List post views with filtering and pagination.

Output Columns:
  ID        Post view identifier
  POST      Post summary or ID
  VIEWER    Viewer name or ID
  VIEWED AT View timestamp

Filters:
  --post            Filter by post ID
  --viewer          Filter by viewer user ID
  --viewed-at-min   Filter by viewed-at on/after (ISO 8601)
  --viewed-at-max   Filter by viewed-at on/before (ISO 8601)
  --is-viewed-at    Filter by has viewed-at (true/false)
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --is-created-at   Filter by has created-at (true/false)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)
  --is-updated-at   Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List post views
  xbe view post-views list

  # Filter by post
  xbe view post-views list --post 123

  # Filter by viewer
  xbe view post-views list --viewer 456

  # Filter by viewed time
  xbe view post-views list --viewed-at-min 2024-01-01T00:00:00Z --viewed-at-max 2024-01-31T23:59:59Z

  # Output as JSON
  xbe view post-views list --json`,
		Args: cobra.NoArgs,
		RunE: runPostViewsList,
	}
	initPostViewsListFlags(cmd)
	return cmd
}

func init() {
	postViewsCmd.AddCommand(newPostViewsListCmd())
}

func initPostViewsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("post", "", "Filter by post ID")
	cmd.Flags().String("viewer", "", "Filter by viewer user ID")
	cmd.Flags().String("viewed-at-min", "", "Filter by viewed-at on/after (ISO 8601)")
	cmd.Flags().String("viewed-at-max", "", "Filter by viewed-at on/before (ISO 8601)")
	cmd.Flags().String("is-viewed-at", "", "Filter by has viewed-at (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostViewsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePostViewsListOptions(cmd)
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
	query.Set("fields[post-views]", "post,viewer,viewed-at,created-at,updated-at")
	query.Set("include", "post,viewer")
	query.Set("fields[posts]", "post-type,status,published-at,short-text-content,creator-name")
	query.Set("fields[users]", "name,email-address")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[post]", opts.Post)
	setFilterIfPresent(query, "filter[viewer]", opts.Viewer)
	setFilterIfPresent(query, "filter[viewed-at-min]", opts.ViewedAtMin)
	setFilterIfPresent(query, "filter[viewed-at-max]", opts.ViewedAtMax)
	setFilterIfPresent(query, "filter[is-viewed-at]", opts.IsViewedAt)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/post-views", query)
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

	rows := buildPostViewRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPostViewsTable(cmd, rows)
}

func parsePostViewsListOptions(cmd *cobra.Command) (postViewsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	post, _ := cmd.Flags().GetString("post")
	viewer, _ := cmd.Flags().GetString("viewer")
	viewedAtMin, _ := cmd.Flags().GetString("viewed-at-min")
	viewedAtMax, _ := cmd.Flags().GetString("viewed-at-max")
	isViewedAt, _ := cmd.Flags().GetString("is-viewed-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return postViewsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		Post:         post,
		Viewer:       viewer,
		ViewedAtMin:  viewedAtMin,
		ViewedAtMax:  viewedAtMax,
		IsViewedAt:   isViewedAt,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildPostViewRows(resp jsonAPIResponse) []postViewRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]postViewRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPostViewRow(resource, included))
	}
	return rows
}

func postViewRowFromSingle(resp jsonAPISingleResponse) postViewRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildPostViewRow(resp.Data, included)
}

func buildPostViewRow(resource jsonAPIResource, included map[string]jsonAPIResource) postViewRow {
	row := postViewRow{
		ID:        resource.ID,
		ViewedAt:  formatDateTime(stringAttr(resource.Attributes, "viewed-at")),
		CreatedAt: formatDateTime(stringAttr(resource.Attributes, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(resource.Attributes, "updated-at")),
	}

	if rel, ok := resource.Relationships["post"]; ok && rel.Data != nil {
		row.PostID = rel.Data.ID
		if post, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := post.Attributes
			row.PostType = stringAttr(attrs, "post-type")
			row.PostStatus = stringAttr(attrs, "status")
			row.PostPublishedAt = formatDateTime(stringAttr(attrs, "published-at"))
			row.PostShortText = stringAttr(attrs, "short-text-content")
			row.PostCreatorName = stringAttr(attrs, "creator-name")
		}
	}

	if rel, ok := resource.Relationships["viewer"]; ok && rel.Data != nil {
		row.ViewerID = rel.Data.ID
		if viewer, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ViewerName = stringAttr(viewer.Attributes, "name")
			row.ViewerEmail = stringAttr(viewer.Attributes, "email-address")
		}
	}

	return row
}

func renderPostViewsTable(cmd *cobra.Command, rows []postViewRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No post views found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPOST\tVIEWER\tVIEWED AT")
	for _, row := range rows {
		postLabel := firstNonEmpty(row.PostShortText, row.PostType, row.PostID)
		viewerLabel := firstNonEmpty(row.ViewerName, row.ViewerEmail, row.ViewerID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(postLabel, 40),
			truncateString(viewerLabel, 30),
			row.ViewedAt,
		)
	}
	return writer.Flush()
}
