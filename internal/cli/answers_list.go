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

type answersListOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	Limit           int
	Offset          int
	Question        string
	WithoutFeedback string
	WithFeedback    string
}

type answerRow struct {
	ID         string `json:"id"`
	Content    string `json:"content,omitempty"`
	Prompt     string `json:"prompt,omitempty"`
	QuestionID string `json:"question_id,omitempty"`
	FeedbackID string `json:"feedback_id,omitempty"`
}

func newAnswersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List answers",
		Long: `List answers with filtering and pagination.

Output Columns:
  ID        Answer identifier
  QUESTION  Question ID
  FEEDBACK  Answer feedback ID (if any)
  CONTENT   Answer content (truncated)

Filters:
  --question          Filter by question ID
  --with-feedback     Filter by presence of feedback (true/false)
  --without-feedback  Filter by absence of feedback (true/false)

Global flags (see xbe --help): --json, --no-auth, --limit, --offset, --base-url, --token`,
		Example: `  # List answers
  xbe view answers list

  # Filter by question
  xbe view answers list --question 123

  # Filter answers without feedback
  xbe view answers list --without-feedback true

  # Output as JSON
  xbe view answers list --json`,
		RunE: runAnswersList,
	}
	initAnswersListFlags(cmd)
	return cmd
}

func init() {
	answersCmd.AddCommand(newAnswersListCmd())
}

func initAnswersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("question", "", "Filter by question ID")
	cmd.Flags().String("without-feedback", "", "Filter by absence of feedback (true/false)")
	cmd.Flags().String("with-feedback", "", "Filter by presence of feedback (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runAnswersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseAnswersListOptions(cmd)
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
	query.Set("fields[answers]", "content,prompt,question,feedback")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[question]", opts.Question)
	setFilterIfPresent(query, "filter[without-feedback]", opts.WithoutFeedback)
	setFilterIfPresent(query, "filter[with-feedback]", opts.WithFeedback)

	body, _, err := client.Get(cmd.Context(), "/v1/answers", query)
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

	rows := buildAnswerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderAnswersTable(cmd, rows)
}

func parseAnswersListOptions(cmd *cobra.Command) (answersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	question, _ := cmd.Flags().GetString("question")
	withoutFeedback, _ := cmd.Flags().GetString("without-feedback")
	withFeedback, _ := cmd.Flags().GetString("with-feedback")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return answersListOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		Limit:           limit,
		Offset:          offset,
		Question:        question,
		WithoutFeedback: withoutFeedback,
		WithFeedback:    withFeedback,
	}, nil
}

func buildAnswerRows(resp jsonAPIResponse) []answerRow {
	rows := make([]answerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := answerRow{
			ID:         resource.ID,
			Content:    strings.TrimSpace(stringAttr(resource.Attributes, "content")),
			Prompt:     strings.TrimSpace(stringAttr(resource.Attributes, "prompt")),
			QuestionID: relationshipIDFromMap(resource.Relationships, "question"),
			FeedbackID: relationshipIDFromMap(resource.Relationships, "feedback"),
		}
		rows = append(rows, row)
	}
	return rows
}

func buildAnswerRowFromSingle(resp jsonAPISingleResponse) answerRow {
	attrs := resp.Data.Attributes
	return answerRow{
		ID:         resp.Data.ID,
		Content:    strings.TrimSpace(stringAttr(attrs, "content")),
		Prompt:     strings.TrimSpace(stringAttr(attrs, "prompt")),
		QuestionID: relationshipIDFromMap(resp.Data.Relationships, "question"),
		FeedbackID: relationshipIDFromMap(resp.Data.Relationships, "feedback"),
	}
}

func renderAnswersTable(cmd *cobra.Command, rows []answerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No answers found.")
		return nil
	}

	const contentMax = 60

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tQUESTION\tFEEDBACK\tCONTENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			row.QuestionID,
			row.FeedbackID,
			truncateString(row.Content, contentMax),
		)
	}
	return writer.Flush()
}
