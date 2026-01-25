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

type predictionKnowledgeBaseQuestionsListOptions struct {
	BaseURL                 string
	Token                   string
	JSON                    bool
	NoAuth                  bool
	Limit                   int
	Offset                  int
	Sort                    string
	PredictionKnowledgeBase string
	PredictionSubject       string
	Status                  string
	TaggedWith              string
	TaggedWithAny           string
	TaggedWithAll           string
	InTagCategory           string
}

type predictionKnowledgeBaseQuestionRow struct {
	ID                        string `json:"id"`
	Title                     string `json:"title,omitempty"`
	Status                    string `json:"status,omitempty"`
	PredictionKnowledgeBaseID string `json:"prediction_knowledge_base_id,omitempty"`
	PredictionSubjectID       string `json:"prediction_subject_id,omitempty"`
	CreatedByID               string `json:"created_by_id,omitempty"`
	AnswerID                  string `json:"answer_id,omitempty"`
}

func newPredictionKnowledgeBaseQuestionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction knowledge base questions",
		Long: `List prediction knowledge base questions with filtering and pagination.

Output Columns:
  ID              Question identifier
  TITLE           Question title (truncated)
  STATUS          Question status
  KNOWLEDGE BASE  Prediction knowledge base ID
  SUBJECT         Prediction subject ID (if present)
  CREATED BY      Creator user ID

Filters:
  --prediction-knowledge-base  Filter by prediction knowledge base ID
  --prediction-subject         Filter by prediction subject ID
  --status                     Filter by status (open/resolved/dismissed)
  --tagged-with                Filter by tag names (comma-separated, all required)
  --tagged-with-any            Filter by any tag names (comma-separated)
  --tagged-with-all            Filter by all tag names (comma-separated)
  --in-tag-category            Filter by tag category slug (comma-separated)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List prediction knowledge base questions
  xbe view prediction-knowledge-base-questions list

  # Filter by knowledge base
  xbe view prediction-knowledge-base-questions list --prediction-knowledge-base 123

  # Filter by status
  xbe view prediction-knowledge-base-questions list --status open

  # Filter by tag name
  xbe view prediction-knowledge-base-questions list --tagged-with "priority"

  # Output as JSON
  xbe view prediction-knowledge-base-questions list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionKnowledgeBaseQuestionsList,
	}
	initPredictionKnowledgeBaseQuestionsListFlags(cmd)
	return cmd
}

func init() {
	predictionKnowledgeBaseQuestionsCmd.AddCommand(newPredictionKnowledgeBaseQuestionsListCmd())
}

func initPredictionKnowledgeBaseQuestionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("prediction-knowledge-base", "", "Filter by prediction knowledge base ID")
	cmd.Flags().String("prediction-subject", "", "Filter by prediction subject ID")
	cmd.Flags().String("status", "", "Filter by status (open/resolved/dismissed)")
	cmd.Flags().String("tagged-with", "", "Filter by tag names (comma-separated, all required)")
	cmd.Flags().String("tagged-with-any", "", "Filter by any tag names (comma-separated)")
	cmd.Flags().String("tagged-with-all", "", "Filter by all tag names (comma-separated)")
	cmd.Flags().String("in-tag-category", "", "Filter by tag category slug (comma-separated)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionKnowledgeBaseQuestionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionKnowledgeBaseQuestionsListOptions(cmd)
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
	query.Set("fields[prediction-knowledge-base-questions]", "title,status,prediction-subject,prediction-knowledge-base,created-by,answer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[prediction_knowledge_base]", opts.PredictionKnowledgeBase)
	setFilterIfPresent(query, "filter[prediction_subject]", opts.PredictionSubject)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[tagged_with]", opts.TaggedWith)
	setFilterIfPresent(query, "filter[tagged_with_any]", opts.TaggedWithAny)
	setFilterIfPresent(query, "filter[tagged_with_all]", opts.TaggedWithAll)
	setFilterIfPresent(query, "filter[in_tag_category]", opts.InTagCategory)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-knowledge-base-questions", query)
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

	rows := buildPredictionKnowledgeBaseQuestionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionKnowledgeBaseQuestionsTable(cmd, rows)
}

func parsePredictionKnowledgeBaseQuestionsListOptions(cmd *cobra.Command) (predictionKnowledgeBaseQuestionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	predictionKnowledgeBase, _ := cmd.Flags().GetString("prediction-knowledge-base")
	predictionSubject, _ := cmd.Flags().GetString("prediction-subject")
	status, _ := cmd.Flags().GetString("status")
	taggedWith, _ := cmd.Flags().GetString("tagged-with")
	taggedWithAny, _ := cmd.Flags().GetString("tagged-with-any")
	taggedWithAll, _ := cmd.Flags().GetString("tagged-with-all")
	inTagCategory, _ := cmd.Flags().GetString("in-tag-category")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionKnowledgeBaseQuestionsListOptions{
		BaseURL:                 baseURL,
		Token:                   token,
		JSON:                    jsonOut,
		NoAuth:                  noAuth,
		Limit:                   limit,
		Offset:                  offset,
		Sort:                    sort,
		PredictionKnowledgeBase: predictionKnowledgeBase,
		PredictionSubject:       predictionSubject,
		Status:                  status,
		TaggedWith:              taggedWith,
		TaggedWithAny:           taggedWithAny,
		TaggedWithAll:           taggedWithAll,
		InTagCategory:           inTagCategory,
	}, nil
}

func buildPredictionKnowledgeBaseQuestionRows(resp jsonAPIResponse) []predictionKnowledgeBaseQuestionRow {
	rows := make([]predictionKnowledgeBaseQuestionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildPredictionKnowledgeBaseQuestionRow(resource))
	}
	return rows
}

func predictionKnowledgeBaseQuestionRowFromSingle(resp jsonAPISingleResponse) predictionKnowledgeBaseQuestionRow {
	return buildPredictionKnowledgeBaseQuestionRow(resp.Data)
}

func buildPredictionKnowledgeBaseQuestionRow(resource jsonAPIResource) predictionKnowledgeBaseQuestionRow {
	row := predictionKnowledgeBaseQuestionRow{
		ID:     resource.ID,
		Title:  stringAttr(resource.Attributes, "title"),
		Status: stringAttr(resource.Attributes, "status"),
	}

	if rel, ok := resource.Relationships["prediction-knowledge-base"]; ok && rel.Data != nil {
		row.PredictionKnowledgeBaseID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["prediction-subject"]; ok && rel.Data != nil {
		row.PredictionSubjectID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["answer"]; ok && rel.Data != nil {
		row.AnswerID = rel.Data.ID
	}

	return row
}

func renderPredictionKnowledgeBaseQuestionsTable(cmd *cobra.Command, rows []predictionKnowledgeBaseQuestionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction knowledge base questions found.")
		return nil
	}

	const titleMax = 40
	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTITLE\tSTATUS\tKNOWLEDGE BASE\tSUBJECT\tCREATED BY")
	for _, row := range rows {
		knowledgeBase := row.PredictionKnowledgeBaseID
		if knowledgeBase == "" {
			knowledgeBase = "-"
		}
		subject := row.PredictionSubjectID
		if subject == "" {
			subject = "-"
		}
		createdBy := row.CreatedByID
		if createdBy == "" {
			createdBy = "-"
		}
		status := row.Status
		if status == "" {
			status = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Title, titleMax),
			status,
			knowledgeBase,
			subject,
			createdBy,
		)
	}
	return writer.Flush()
}
