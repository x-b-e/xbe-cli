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

type postRouterJobsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	PostWorkerClassName string
	PostRouter          string
	Post                string
}

type postRouterJobRow struct {
	ID                  string `json:"id"`
	PostRouterID        string `json:"post_router_id,omitempty"`
	PostID              string `json:"post_id,omitempty"`
	PostWorkerClassName string `json:"post_worker_class_name,omitempty"`
	PostWorkerJID       string `json:"post_worker_jid,omitempty"`
	IsNotWorthPosting   bool   `json:"is_not_worth_posting,omitempty"`
}

func newPostRouterJobsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List post router jobs",
		Long: `List post router jobs with filtering and pagination.

Post router jobs track background worker jobs created for routed posts.

Output Columns:
  ID                 Post router job identifier
  POST_ROUTER        Post router ID
  POST               Post ID
  WORKER_CLASS       Post worker class name
  WORKER_JID         Background worker job ID
  NOT_WORTH_POSTING  Whether the post was skipped

Filters:
  --post-worker-class-name  Filter by post worker class name
  --post-router             Filter by post router ID
  --post                    Filter by post ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List post router jobs
  xbe view post-router-jobs list

  # Filter by post router
  xbe view post-router-jobs list --post-router 123

  # Filter by post
  xbe view post-router-jobs list --post 456

  # Filter by worker class
  xbe view post-router-jobs list --post-worker-class-name "Posters::FooWorker"

  # Output as JSON
  xbe view post-router-jobs list --json`,
		Args: cobra.NoArgs,
		RunE: runPostRouterJobsList,
	}
	initPostRouterJobsListFlags(cmd)
	return cmd
}

func init() {
	postRouterJobsCmd.AddCommand(newPostRouterJobsListCmd())
}

func initPostRouterJobsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("post-worker-class-name", "", "Filter by post worker class name")
	cmd.Flags().String("post-router", "", "Filter by post router ID")
	cmd.Flags().String("post", "", "Filter by post ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostRouterJobsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePostRouterJobsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[post-router-jobs]", "post-worker-class-name,post-worker-jid,is-not-worth-posting,post-router,post")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[post-worker-class-name]", opts.PostWorkerClassName)
	setFilterIfPresent(query, "filter[post-router]", opts.PostRouter)
	setFilterIfPresent(query, "filter[post]", opts.Post)

	body, _, err := client.Get(cmd.Context(), "/v1/post-router-jobs", query)
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

	rows := buildPostRouterJobRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPostRouterJobsTable(cmd, rows)
}

func parsePostRouterJobsListOptions(cmd *cobra.Command) (postRouterJobsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	postWorkerClassName, _ := cmd.Flags().GetString("post-worker-class-name")
	postRouter, _ := cmd.Flags().GetString("post-router")
	post, _ := cmd.Flags().GetString("post")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return postRouterJobsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		PostWorkerClassName: postWorkerClassName,
		PostRouter:          postRouter,
		Post:                post,
	}, nil
}

func buildPostRouterJobRows(resp jsonAPIResponse) []postRouterJobRow {
	rows := make([]postRouterJobRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := postRouterJobRow{
			ID:                  resource.ID,
			PostWorkerClassName: stringAttr(attrs, "post-worker-class-name"),
			PostWorkerJID:       stringAttr(attrs, "post-worker-jid"),
			IsNotWorthPosting:   boolAttr(attrs, "is-not-worth-posting"),
			PostRouterID:        relationshipIDFromMap(resource.Relationships, "post-router"),
			PostID:              relationshipIDFromMap(resource.Relationships, "post"),
		}

		rows = append(rows, row)
	}
	return rows
}

func renderPostRouterJobsTable(cmd *cobra.Command, rows []postRouterJobRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No post router jobs found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tPOST_ROUTER\tPOST\tWORKER_CLASS\tWORKER_JID\tNOT_WORTH_POSTING")
	for _, row := range rows {
		notWorthPosting := ""
		if row.IsNotWorthPosting {
			notWorthPosting = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.PostRouterID,
			row.PostID,
			row.PostWorkerClassName,
			row.PostWorkerJID,
			notWorthPosting,
		)
	}
	return writer.Flush()
}
