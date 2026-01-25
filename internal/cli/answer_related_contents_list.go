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

type answerRelatedContentsListOptions struct {
	BaseURL               string
	Token                 string
	JSON                  bool
	NoAuth                bool
	Limit                 int
	Offset                int
	Sort                  string
	Answer                string
	RelatedContentType    string
	RelatedContentID      string
	NotRelatedContentType string
	CreatedAtMin          string
	CreatedAtMax          string
	IsCreatedAt           string
	UpdatedAtMin          string
	UpdatedAtMax          string
	IsUpdatedAt           string
}

type answerRelatedContentRow struct {
	ID                 string  `json:"id"`
	AnswerID           string  `json:"answer_id,omitempty"`
	RelatedContentType string  `json:"related_content_type,omitempty"`
	RelatedContentID   string  `json:"related_content_id,omitempty"`
	Similarity         float64 `json:"similarity,omitempty"`
	CreatedAt          string  `json:"created_at,omitempty"`
	UpdatedAt          string  `json:"updated_at,omitempty"`
}

func newAnswerRelatedContentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List answer related contents",
		Long: `List answer related contents with filtering and pagination.

Answer related contents link answers to other content types and provide a
similarity score for each match.

Output Columns:
  ID               Related content link identifier
  ANSWER           Answer ID
  RELATED CONTENT  Related content type and ID
  SIMILARITY       Similarity score

Filters:
  --answer                    Filter by answer ID
  --related-content-type      Filter by related content type
  --related-content-id        Filter by related content ID (requires --related-content-type)
  --not-related-content-type  Exclude a related content type
  --created-at-min            Filter by created-at on/after (ISO 8601)
  --created-at-max            Filter by created-at on/before (ISO 8601)
  --is-created-at             Filter by has created-at (true/false)
  --updated-at-min            Filter by updated-at on/after (ISO 8601)
  --updated-at-max            Filter by updated-at on/before (ISO 8601)
  --is-updated-at             Filter by has updated-at (true/false)

Related content types:
  newsletters, glossary-terms, release-notes, press-releases, objectives,
  features, questions (also accepts class names like GlossaryTerm)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List related content links
  xbe view answer-related-contents list

  # Filter by answer
  xbe view answer-related-contents list --answer 123

  # Filter by related content
  xbe view answer-related-contents list --related-content-type newsletters --related-content-id 456

  # Filter by related content type only
  xbe view answer-related-contents list --related-content-type release-notes

  # Output as JSON
  xbe view answer-related-contents list --json`,
		Args: cobra.NoArgs,
		RunE: runAnswerRelatedContentsList,
	}
	initAnswerRelatedContentsListFlags(cmd)
	return cmd
}

func init() {
	answerRelatedContentsCmd.AddCommand(newAnswerRelatedContentsListCmd())
}

func initAnswerRelatedContentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("answer", "", "Filter by answer ID")
	cmd.Flags().String("related-content-type", "", "Filter by related content type")
	cmd.Flags().String("related-content-id", "", "Filter by related content ID")
	cmd.Flags().String("not-related-content-type", "", "Exclude a related content type")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runAnswerRelatedContentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseAnswerRelatedContentsListOptions(cmd)
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
	query.Set("fields[answer-related-contents]", "answer,related-content,similarity,created-at,updated-at")
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[answer]", opts.Answer)
	relatedContentType := normalizeRelatedContentTypeForFilter(opts.RelatedContentType)
	if relatedContentType != "" && opts.RelatedContentID != "" {
		query.Set("filter[related_content]", relatedContentType+"|"+opts.RelatedContentID)
	} else if opts.RelatedContentID != "" {
		err := fmt.Errorf("--related-content-id requires --related-content-type")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	} else {
		setFilterIfPresent(query, "filter[related_content_type]", relatedContentType)
	}
	setFilterIfPresent(query, "filter[not_related_content_type]", normalizeRelatedContentTypeForFilter(opts.NotRelatedContentType))
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/answer-related-contents", query)
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

	rows := buildAnswerRelatedContentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderAnswerRelatedContentsTable(cmd, rows)
}

func parseAnswerRelatedContentsListOptions(cmd *cobra.Command) (answerRelatedContentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	answer, _ := cmd.Flags().GetString("answer")
	relatedContentType, _ := cmd.Flags().GetString("related-content-type")
	relatedContentID, _ := cmd.Flags().GetString("related-content-id")
	notRelatedContentType, _ := cmd.Flags().GetString("not-related-content-type")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return answerRelatedContentsListOptions{
		BaseURL:               baseURL,
		Token:                 token,
		JSON:                  jsonOut,
		NoAuth:                noAuth,
		Limit:                 limit,
		Offset:                offset,
		Sort:                  sort,
		Answer:                answer,
		RelatedContentType:    relatedContentType,
		RelatedContentID:      relatedContentID,
		NotRelatedContentType: notRelatedContentType,
		CreatedAtMin:          createdAtMin,
		CreatedAtMax:          createdAtMax,
		IsCreatedAt:           isCreatedAt,
		UpdatedAtMin:          updatedAtMin,
		UpdatedAtMax:          updatedAtMax,
		IsUpdatedAt:           isUpdatedAt,
	}, nil
}

func buildAnswerRelatedContentRows(resp jsonAPIResponse) []answerRelatedContentRow {
	rows := make([]answerRelatedContentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := answerRelatedContentRow{
			ID:         resource.ID,
			Similarity: floatAttr(resource.Attributes, "similarity"),
			CreatedAt:  formatDateTime(stringAttr(resource.Attributes, "created-at")),
			UpdatedAt:  formatDateTime(stringAttr(resource.Attributes, "updated-at")),
		}

		if rel, ok := resource.Relationships["answer"]; ok && rel.Data != nil {
			row.AnswerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["related-content"]; ok && rel.Data != nil {
			row.RelatedContentType = rel.Data.Type
			row.RelatedContentID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderAnswerRelatedContentsTable(cmd *cobra.Command, rows []answerRelatedContentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No answer related contents found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tANSWER\tRELATED CONTENT\tSIMILARITY")
	for _, row := range rows {
		related := ""
		if row.RelatedContentType != "" && row.RelatedContentID != "" {
			related = row.RelatedContentType + "/" + row.RelatedContentID
		}
		similarity := ""
		if row.Similarity != 0 {
			similarity = fmt.Sprintf("%.4f", row.Similarity)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.AnswerID,
			related,
			similarity,
		)
	}
	return writer.Flush()
}
