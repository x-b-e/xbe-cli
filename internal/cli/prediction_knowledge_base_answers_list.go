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

type predictionKnowledgeBaseAnswersListOptions struct {
	BaseURL                         string
	Token                           string
	JSON                            bool
	NoAuth                          bool
	Limit                           int
	Offset                          int
	PredictionKnowledgeBaseQuestion string
}

type predictionKnowledgeBaseAnswerRow struct {
	ID                                string `json:"id"`
	Content                           string `json:"content,omitempty"`
	PredictionKnowledgeBaseQuestionID string `json:"prediction_knowledge_base_question_id,omitempty"`
}

func newPredictionKnowledgeBaseAnswersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prediction knowledge base answers",
		Long: `List prediction knowledge base answers.

Output Columns:
  ID           Answer identifier
  QUESTION ID  Prediction knowledge base question ID
  CONTENT      Answer content (truncated)

Filters:
  --prediction-knowledge-base-question  Filter by prediction knowledge base question ID

Global flags (see xbe --help): --json, --no-auth, --limit, --offset, --base-url, --token`,
		Example: `  # List prediction knowledge base answers
  xbe view prediction-knowledge-base-answers list

  # Filter by question
  xbe view prediction-knowledge-base-answers list --prediction-knowledge-base-question 123

  # Output as JSON
  xbe view prediction-knowledge-base-answers list --json`,
		Args: cobra.NoArgs,
		RunE: runPredictionKnowledgeBaseAnswersList,
	}
	initPredictionKnowledgeBaseAnswersListFlags(cmd)
	return cmd
}

func init() {
	predictionKnowledgeBaseAnswersCmd.AddCommand(newPredictionKnowledgeBaseAnswersListCmd())
}

func initPredictionKnowledgeBaseAnswersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("prediction-knowledge-base-question", "", "Filter by prediction knowledge base question ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPredictionKnowledgeBaseAnswersList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePredictionKnowledgeBaseAnswersListOptions(cmd)
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
	query.Set("fields[prediction-knowledge-base-answers]", "content,prediction-knowledge-base-question")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[prediction-knowledge-base-question]", opts.PredictionKnowledgeBaseQuestion)

	body, _, err := client.Get(cmd.Context(), "/v1/prediction-knowledge-base-answers", query)
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

	rows := buildPredictionKnowledgeBaseAnswerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPredictionKnowledgeBaseAnswersTable(cmd, rows)
}

func parsePredictionKnowledgeBaseAnswersListOptions(cmd *cobra.Command) (predictionKnowledgeBaseAnswersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	question, _ := cmd.Flags().GetString("prediction-knowledge-base-question")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return predictionKnowledgeBaseAnswersListOptions{
		BaseURL:                         baseURL,
		Token:                           token,
		JSON:                            jsonOut,
		NoAuth:                          noAuth,
		Limit:                           limit,
		Offset:                          offset,
		PredictionKnowledgeBaseQuestion: question,
	}, nil
}

func buildPredictionKnowledgeBaseAnswerRows(resp jsonAPIResponse) []predictionKnowledgeBaseAnswerRow {
	rows := make([]predictionKnowledgeBaseAnswerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := predictionKnowledgeBaseAnswerRow{
			ID:                                resource.ID,
			Content:                           strings.TrimSpace(stringAttr(resource.Attributes, "content")),
			PredictionKnowledgeBaseQuestionID: relationshipIDFromMap(resource.Relationships, "prediction-knowledge-base-question"),
		}
		rows = append(rows, row)
	}
	return rows
}

func buildPredictionKnowledgeBaseAnswerRowFromSingle(resp jsonAPISingleResponse) predictionKnowledgeBaseAnswerRow {
	attrs := resp.Data.Attributes
	return predictionKnowledgeBaseAnswerRow{
		ID:                                resp.Data.ID,
		Content:                           strings.TrimSpace(stringAttr(attrs, "content")),
		PredictionKnowledgeBaseQuestionID: relationshipIDFromMap(resp.Data.Relationships, "prediction-knowledge-base-question"),
	}
}

func renderPredictionKnowledgeBaseAnswersTable(cmd *cobra.Command, rows []predictionKnowledgeBaseAnswerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prediction knowledge base answers found.")
		return nil
	}

	const contentMax = 60

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tQUESTION ID\tCONTENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.PredictionKnowledgeBaseQuestionID,
			truncateString(row.Content, contentMax),
		)
	}
	return writer.Flush()
}
