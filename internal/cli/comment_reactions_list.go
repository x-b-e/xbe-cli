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

type commentReactionsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
	NotID        string
}

type commentReactionRow struct {
	ID                        string `json:"id"`
	CommentID                 string `json:"comment_id,omitempty"`
	CommentBody               string `json:"comment_body,omitempty"`
	CommentIsAdminOnly        bool   `json:"comment_is_admin_only"`
	ReactionClassificationID  string `json:"reaction_classification_id,omitempty"`
	ReactionLabel             string `json:"reaction_label,omitempty"`
	ReactionUTF8              string `json:"reaction_utf8,omitempty"`
	ReactionExternalReference string `json:"reaction_external_reference,omitempty"`
	CreatedByID               string `json:"created_by_id,omitempty"`
	CreatedByName             string `json:"created_by_name,omitempty"`
	CreatedByEmail            string `json:"created_by_email,omitempty"`
	CreatedAt                 string `json:"created_at,omitempty"`
	UpdatedAt                 string `json:"updated_at,omitempty"`
}

func newCommentReactionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comment reactions",
		Long: `List comment reactions with filtering and pagination.

Output Columns:
  ID        Comment reaction identifier
  COMMENT   Comment body or ID
  REACTION  Reaction label/emoji
  CREATED BY User who reacted
  CREATED   Reaction timestamp

Filters:
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --is-created-at   Filter by has created-at (true/false)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)
  --is-updated-at   Filter by has updated-at (true/false)
  --not-id          Exclude reactions by ID (comma-separated)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List comment reactions
  xbe view comment-reactions list

  # Filter by created time
  xbe view comment-reactions list --created-at-min 2024-01-01T00:00:00Z

  # Output as JSON
  xbe view comment-reactions list --json`,
		Args: cobra.NoArgs,
		RunE: runCommentReactionsList,
	}
	initCommentReactionsListFlags(cmd)
	return cmd
}

func init() {
	commentReactionsCmd.AddCommand(newCommentReactionsListCmd())
}

func initCommentReactionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("not-id", "", "Exclude reactions by ID (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runCommentReactionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseCommentReactionsListOptions(cmd)
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
	query.Set("fields[comment-reactions]", "comment,created-by,reaction-classification,created-at,updated-at")
	query.Set("include", "comment,created-by,reaction-classification")
	query.Set("fields[comments]", "body,is-admin-only")
	query.Set("fields[users]", "name,email-address")
	query.Set("fields[reaction-classifications]", "label,utf8,external-reference")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)
	setFilterIfPresent(query, "filter[not-id]", opts.NotID)

	body, _, err := client.Get(cmd.Context(), "/v1/comment-reactions", query)
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

	rows := buildCommentReactionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderCommentReactionsTable(cmd, rows)
}

func parseCommentReactionsListOptions(cmd *cobra.Command) (commentReactionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	notID, _ := cmd.Flags().GetString("not-id")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return commentReactionsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
		NotID:        notID,
	}, nil
}

func buildCommentReactionRows(resp jsonAPIResponse) []commentReactionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]commentReactionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := commentReactionRow{
			ID:        resource.ID,
			CreatedAt: formatDateTime(stringAttr(resource.Attributes, "created-at")),
			UpdatedAt: formatDateTime(stringAttr(resource.Attributes, "updated-at")),
		}

		if rel, ok := resource.Relationships["comment"]; ok && rel.Data != nil {
			row.CommentID = rel.Data.ID
			if comment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CommentBody = stringAttr(comment.Attributes, "body")
				row.CommentIsAdminOnly = boolAttr(comment.Attributes, "is-admin-only")
			}
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CreatedByName = stringAttr(user.Attributes, "name")
				row.CreatedByEmail = stringAttr(user.Attributes, "email-address")
			}
		}

		if rel, ok := resource.Relationships["reaction-classification"]; ok && rel.Data != nil {
			row.ReactionClassificationID = rel.Data.ID
			if reaction, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ReactionLabel = stringAttr(reaction.Attributes, "label")
				row.ReactionUTF8 = stringAttr(reaction.Attributes, "utf8")
				row.ReactionExternalReference = stringAttr(reaction.Attributes, "external-reference")
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func commentReactionRowFromSingle(resp jsonAPISingleResponse) commentReactionRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	resource := resp.Data
	row := commentReactionRow{
		ID:        resource.ID,
		CreatedAt: formatDateTime(stringAttr(resource.Attributes, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(resource.Attributes, "updated-at")),
	}

	if rel, ok := resource.Relationships["comment"]; ok && rel.Data != nil {
		row.CommentID = rel.Data.ID
		if comment, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.CommentBody = stringAttr(comment.Attributes, "body")
			row.CommentIsAdminOnly = boolAttr(comment.Attributes, "is-admin-only")
		}
	}

	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.CreatedByName = stringAttr(user.Attributes, "name")
			row.CreatedByEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	if rel, ok := resource.Relationships["reaction-classification"]; ok && rel.Data != nil {
		row.ReactionClassificationID = rel.Data.ID
		if reaction, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ReactionLabel = stringAttr(reaction.Attributes, "label")
			row.ReactionUTF8 = stringAttr(reaction.Attributes, "utf8")
			row.ReactionExternalReference = stringAttr(reaction.Attributes, "external-reference")
		}
	}

	return row
}

func renderCommentReactionsTable(cmd *cobra.Command, rows []commentReactionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No comment reactions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tCOMMENT\tREACTION\tCREATED BY\tCREATED")
	for _, row := range rows {
		comment := firstNonEmpty(truncateString(row.CommentBody, 40), row.CommentID)
		reaction := firstNonEmpty(row.ReactionLabel, row.ReactionClassificationID)
		if row.ReactionUTF8 != "" {
			if reaction != "" {
				reaction = row.ReactionUTF8 + " " + reaction
			} else {
				reaction = row.ReactionUTF8
			}
		}
		createdBy := firstNonEmpty(row.CreatedByName, row.CreatedByEmail, row.CreatedByID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			comment,
			truncateString(reaction, 24),
			truncateString(createdBy, 24),
			formatDateTime(row.CreatedAt),
		)
	}
	return writer.Flush()
}
