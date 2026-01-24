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

type proffersListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Limit         int
	Offset        int
	Sort          string
	Kind          string
	CreatedBy     string
	SimilarToText string
}

type profferRow struct {
	ID                  string `json:"id"`
	Title               string `json:"title,omitempty"`
	Kind                string `json:"kind,omitempty"`
	CreatedByID         string `json:"created_by_id,omitempty"`
	CreatedByName       string `json:"created_by_name,omitempty"`
	LikeCount           int    `json:"like_count"`
	ModerationStatus    string `json:"moderation_status,omitempty"`
	Similarity          string `json:"similarity,omitempty"`
	HasCurrentUserLiked bool   `json:"has_current_user_liked"`
}

func newProffersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List proffers",
		Long: `List proffers (feature suggestions).

Output Columns:
  ID          Proffer identifier
  TITLE       Proffer title
  KIND        Proffer kind
  CREATED BY  Creator (name or ID)
  LIKES       Like count
  MODERATION  Moderation status
  SIMILARITY  Similarity score (when using --similar-to-text)

Filters:
  --kind             Filter by proffer kind (hot_feed_post/make_it_so_action)
  --created-by       Filter by created-by user ID
  --similar-to-text  Find proffers similar to the supplied text (uses OpenAI embeddings)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List proffers
  xbe view proffers list

  # Filter by kind
  xbe view proffers list --kind hot_feed_post

  # Filter by creator
  xbe view proffers list --created-by 123

  # Find similar proffers (may take longer)
  xbe view proffers list --similar-to-text "Add export to CSV"

  # Output as JSON
  xbe view proffers list --json`,
		Args: cobra.NoArgs,
		RunE: runProffersList,
	}
	initProffersListFlags(cmd)
	return cmd
}

func init() {
	proffersCmd.AddCommand(newProffersListCmd())
}

func initProffersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("kind", "", "Filter by proffer kind")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("similar-to-text", "", "Find proffers similar to the supplied text")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProffersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseProffersListOptions(cmd)
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
	query.Set("fields[proffers]", "title,kind,created-by-name,like-count,moderation-status,similarity,has-current-user-liked,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[similar_to_text]", opts.SimilarToText)

	body, _, err := client.Get(cmd.Context(), "/v1/proffers", query)
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

	rows := buildProfferRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderProffersTable(cmd, rows)
}

func parseProffersListOptions(cmd *cobra.Command) (proffersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	kind, _ := cmd.Flags().GetString("kind")
	createdBy, _ := cmd.Flags().GetString("created-by")
	similarToText, _ := cmd.Flags().GetString("similar-to-text")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return proffersListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Limit:         limit,
		Offset:        offset,
		Sort:          sort,
		Kind:          kind,
		CreatedBy:     createdBy,
		SimilarToText: similarToText,
	}, nil
}

func buildProfferRows(resp jsonAPIResponse) []profferRow {
	rows := make([]profferRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := profferRow{
			ID:                  resource.ID,
			Title:               strings.TrimSpace(stringAttr(attrs, "title")),
			Kind:                stringAttr(attrs, "kind"),
			CreatedByName:       stringAttr(attrs, "created-by-name"),
			LikeCount:           intAttr(attrs, "like-count"),
			ModerationStatus:    stringAttr(attrs, "moderation-status"),
			Similarity:          stringAttr(attrs, "similarity"),
			HasCurrentUserLiked: boolAttr(attrs, "has-current-user-liked"),
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderProffersTable(cmd *cobra.Command, rows []profferRow) error {
	out := cmd.OutOrStdout()
	if len(rows) == 0 {
		fmt.Fprintln(out, "No proffers found.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tKIND\tCREATED BY\tLIKES\tMODERATION\tSIMILARITY")
	for _, row := range rows {
		createdBy := firstNonEmpty(row.CreatedByName, row.CreatedByID)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			row.ID,
			truncateString(row.Title, 40),
			row.Kind,
			createdBy,
			row.LikeCount,
			row.ModerationStatus,
			row.Similarity,
		)
	}
	return w.Flush()
}
