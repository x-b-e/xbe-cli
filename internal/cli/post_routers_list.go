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

type postRoutersListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Status  string
	Post    string
}

type postRouterRow struct {
	ID     string `json:"id"`
	Status string `json:"status,omitempty"`
	PostID string `json:"post_id,omitempty"`
}

func newPostRoutersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List post routers",
		Long: `List post routers with filtering and pagination.

Post routers analyze posts and enqueue background routing jobs.

Output Columns:
  ID      Post router identifier
  STATUS  Router status
  POST    Post ID

Filters:
  --status  Filter by status (queueing/analyzing/routed)
  --post    Filter by post ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List post routers
  xbe view post-routers list

  # Filter by status
  xbe view post-routers list --status analyzing

  # Filter by post
  xbe view post-routers list --post 123

  # Output as JSON
  xbe view post-routers list --json`,
		Args: cobra.NoArgs,
		RunE: runPostRoutersList,
	}
	initPostRoutersListFlags(cmd)
	return cmd
}

func init() {
	postRoutersCmd.AddCommand(newPostRoutersListCmd())
}

func initPostRoutersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status (queueing/analyzing/routed)")
	cmd.Flags().String("post", "", "Filter by post ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPostRoutersList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePostRoutersListOptions(cmd)
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
	query.Set("fields[post-routers]", "status,post")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[post]", opts.Post)

	body, _, err := client.Get(cmd.Context(), "/v1/post-routers", query)
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

	rows := buildPostRouterRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPostRoutersTable(cmd, rows)
}

func parsePostRoutersListOptions(cmd *cobra.Command) (postRoutersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	post, _ := cmd.Flags().GetString("post")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return postRoutersListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Status:  status,
		Post:    post,
	}, nil
}

func buildPostRouterRows(resp jsonAPIResponse) []postRouterRow {
	rows := make([]postRouterRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := postRouterRow{
			ID:     resource.ID,
			Status: stringAttr(attrs, "status"),
			PostID: relationshipIDFromMap(resource.Relationships, "post"),
		}

		rows = append(rows, row)
	}
	return rows
}

func renderPostRoutersTable(cmd *cobra.Command, rows []postRouterRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No post routers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tPOST")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.PostID,
		)
	}
	return writer.Flush()
}
