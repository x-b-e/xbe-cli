package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type postsListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Status         string
	PostType       string
	PublishedAtMin string
	PublishedAtMax string
	Creator        string
	SubjectPost    string
}

func newPostsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List posts",
		Long: `List posts with filtering and pagination.

Returns a list of posts matching the specified criteria, sorted by publication
date (newest first). Posts are displayed in a feed format with a content preview.

Pagination:
  Use --limit and --offset to paginate through large result sets.
  The server has a default page size if --limit is not specified.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Post Types:
  basic, notification, action, new_membership, post_summary, post_activity,
  objective_status, objective_completion, objective_status_scoreboard,
  objective_sales_responsible_person_assignment_issues,
  key_result_completion, key_result_status_scoreboard,
  key_result_customer_success_responsible_person_assignment_issues,
  job_production_plan_recap, job_production_plan_material_site_start_timing,
  customer_daily_job_production_plan_recap, customer_job_production_plan_schedule,
  customer_lineup_schedule, material_supplier_production_daily_recap,
  material_supplier_production_monthly_recap, trucker_shift_summary,
  trucking_time_card_administration_report_card, trucking_tender_acceptance_report_card,
  driver_day_recap, release_note_summary, proffer`,
		Example: `  # List recent posts
  xbe view posts list

  # Filter by status
  xbe view posts list --status published
  xbe view posts list --status draft

  # Filter by post type
  xbe view posts list --post-type basic
  xbe view posts list --post-type objective_status

  # Filter by date range
  xbe view posts list --published-at-min 2024-01-01 --published-at-max 2024-06-30

  # Filter by creator
  xbe view posts list --creator "User|123"

  # Paginate results
  xbe view posts list --limit 20 --offset 40

  # Output as JSON for scripting
  xbe view posts list --json

  # Access without authentication
  xbe view posts list --no-auth`,
		RunE: runPostsList,
	}
	initPostsListFlags(cmd)
	return cmd
}

func init() {
	postsCmd.AddCommand(newPostsListCmd())
}

func initPostsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("status", "", "Filter by status (draft/published)")
	cmd.Flags().String("post-type", "", "Filter by post type")
	cmd.Flags().String("published-at-min", "", "Filter to posts published on or after this date (YYYY-MM-DD)")
	cmd.Flags().String("published-at-max", "", "Filter to posts published on or before this date (YYYY-MM-DD)")
	cmd.Flags().String("creator", "", "Filter by creator (e.g., User|123)")
	cmd.Flags().String("subject-post", "", "Filter by subject post ID (comma-separated for multiple)")
	// NOTE: similar-to-text filter removed due to performance issues (OpenAI embedding calls too slow for posts)
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePostsListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "-published-at")
	query.Set("fields[posts]", "post-type,published-at,short-text-content,creator-name,status,creator,creator-parents")
	query.Set("include", "creator")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[post-type]", opts.PostType)
	setFilterIfPresent(query, "filter[published-at-min]", opts.PublishedAtMin)
	setFilterIfPresent(query, "filter[published-at-max]", opts.PublishedAtMax)
	setFilterIfPresent(query, "filter[creator]", opts.Creator)
	setFilterIfPresent(query, "filter[subject-post]", opts.SubjectPost)

	body, _, err := client.Get(cmd.Context(), "/v1/posts", query)
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

	if opts.JSON {
		rows := buildPostRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPostsFeed(cmd, resp)
}

func parsePostsListOptions(cmd *cobra.Command) (postsListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return postsListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return postsListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return postsListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return postsListOptions{}, err
	}
	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return postsListOptions{}, err
	}
	postType, err := cmd.Flags().GetString("post-type")
	if err != nil {
		return postsListOptions{}, err
	}
	publishedAtMin, err := cmd.Flags().GetString("published-at-min")
	if err != nil {
		return postsListOptions{}, err
	}
	publishedAtMax, err := cmd.Flags().GetString("published-at-max")
	if err != nil {
		return postsListOptions{}, err
	}
	creator, err := cmd.Flags().GetString("creator")
	if err != nil {
		return postsListOptions{}, err
	}
	subjectPost, err := cmd.Flags().GetString("subject-post")
	if err != nil {
		return postsListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return postsListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return postsListOptions{}, err
	}

	return postsListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Status:         status,
		PostType:       postType,
		PublishedAtMin: publishedAtMin,
		PublishedAtMax: publishedAtMax,
		Creator:        creator,
		SubjectPost:    subjectPost,
	}, nil
}

func renderPostsFeed(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildPostRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No posts found.")
		return nil
	}

	out := cmd.OutOrStdout()
	const contentPreviewMax = 400

	for i, row := range rows {
		// Header line: [ID] post_type
		fmt.Fprintf(out, "[%s] %s\n", row.ID, row.PostType)

		// Meta line: date • creator
		if row.Creator != "" {
			fmt.Fprintf(out, "%s • %s\n", row.Published, row.Creator)
		} else {
			fmt.Fprintf(out, "%s\n", row.Published)
		}

		// Content preview (stripped of markdown)
		if row.Content != "" {
			preview := stripMarkdown(row.Content)
			preview = truncateString(preview, contentPreviewMax)
			if preview != "" {
				fmt.Fprintf(out, "%s\n", preview)
			}
		}

		// Blank line between posts
		if i < len(rows)-1 {
			fmt.Fprintln(out)
		}
	}

	return nil
}

var (
	markdownHeaderRegex    = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	markdownBoldRegex      = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	markdownItalicRegex    = regexp.MustCompile(`\*([^*]+)\*`)
	markdownLinkRegex      = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	markdownTableRowRegex  = regexp.MustCompile(`(?m)^\|.+\|$`)
	markdownTableSepRegex  = regexp.MustCompile(`(?m)^\|[\s\-:│|]+`)
	markdownDashLineRegex  = regexp.MustCompile(`(?m)^[\s\-:│|]+$`)
	markdownMultiNewline   = regexp.MustCompile(`\n{3,}`)
	markdownMultiSpace     = regexp.MustCompile(`[ \t]+`)
	markdownLeadingNewline = regexp.MustCompile(`^\n+`)
)

func stripMarkdown(s string) string {
	// Decode HTML entities
	s = decodeHTMLEntities(s)
	// Remove entire markdown table rows (|...|) before removing pipes
	s = markdownTableRowRegex.ReplaceAllString(s, "")
	// Remove table separator lines (|---|---|...)
	s = markdownTableSepRegex.ReplaceAllString(s, "")
	// Remove markdown headers (# ## ### etc)
	s = markdownHeaderRegex.ReplaceAllString(s, "")
	// Convert **bold** to just the text
	s = markdownBoldRegex.ReplaceAllString(s, "$1")
	// Convert *italic* to just the text
	s = markdownItalicRegex.ReplaceAllString(s, "$1")
	// Convert [text](url) to just text
	s = markdownLinkRegex.ReplaceAllString(s, "$1")
	// Replace pipe separators with bullet
	s = strings.ReplaceAll(s, " | ", " • ")
	s = strings.ReplaceAll(s, "| ", "")
	s = strings.ReplaceAll(s, " |", "")
	// Collapse multiple spaces/tabs to single space
	s = markdownMultiSpace.ReplaceAllString(s, " ")
	// Remove lines that are just dashes/spaces (table separators, horizontal rules)
	s = markdownDashLineRegex.ReplaceAllString(s, "")
	// Collapse 3+ newlines to 2
	s = markdownMultiNewline.ReplaceAllString(s, "\n\n")
	// Remove leading newlines
	s = markdownLeadingNewline.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

type postRow struct {
	ID             string      `json:"id"`
	PostType       string      `json:"post_type"`
	Status         string      `json:"status"`
	Published      string      `json:"published"`
	Creator        string      `json:"creator"`
	CreatorType    string      `json:"creator_type"`
	CreatorParents interface{} `json:"creator_parents,omitempty"`
	Content        string      `json:"content"`
}

func buildPostRows(resp jsonAPIResponse) []postRow {
	rows := make([]postRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, postRow{
			ID:             resource.ID,
			PostType:       stringAttr(resource.Attributes, "post-type"),
			Status:         stringAttr(resource.Attributes, "status"),
			Published:      formatDate(stringAttr(resource.Attributes, "published-at")),
			Creator:        strings.TrimSpace(stringAttr(resource.Attributes, "creator-name")),
			CreatorType:    resolveCreatorType(resource),
			CreatorParents: resource.Attributes["creator-parents"],
			Content:        strings.TrimSpace(stringAttr(resource.Attributes, "short-text-content")),
		})
	}

	return rows
}

func resolveCreatorType(resource jsonAPIResource) string {
	rel, ok := resource.Relationships["creator"]
	if !ok || rel.Data == nil {
		return ""
	}
	return rel.Data.Type
}
