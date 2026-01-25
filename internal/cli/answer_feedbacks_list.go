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

type answerFeedbacksListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Answer  string
}

type answerFeedbackRow struct {
	ID            string   `json:"id"`
	Score         *float64 `json:"score,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	BetterContent string   `json:"better_content,omitempty"`
	AnswerID      string   `json:"answer_id,omitempty"`
}

func newAnswerFeedbacksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List answer feedbacks",
		Long: `List answer feedbacks for answers.

Output Columns:
  ID              Answer feedback identifier
  SCORE           Feedback score (0 to 1)
  NOTES           Feedback notes (truncated)
  BETTER CONTENT  Suggested improved content (truncated)
  ANSWER ID       Answer ID

Filters:
  --answer    Filter by answer ID

Global flags (see xbe --help): --json, --no-auth, --limit, --offset, --base-url, --token`,
		Example: `  # List answer feedbacks
  xbe view answer-feedbacks list

  # Filter by answer
  xbe view answer-feedbacks list --answer 123

  # Output as JSON
  xbe view answer-feedbacks list --json`,
		RunE: runAnswerFeedbacksList,
	}
	initAnswerFeedbacksListFlags(cmd)
	return cmd
}

func init() {
	answerFeedbacksCmd.AddCommand(newAnswerFeedbacksListCmd())
}

func initAnswerFeedbacksListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("answer", "", "Filter by answer ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runAnswerFeedbacksList(cmd *cobra.Command, _ []string) error {
	opts, err := parseAnswerFeedbacksListOptions(cmd)
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
	query.Set("fields[answer-feedbacks]", "score,notes,better-content,answer")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	setFilterIfPresent(query, "filter[answer]", opts.Answer)

	body, _, err := client.Get(cmd.Context(), "/v1/answer-feedbacks", query)
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

	rows := buildAnswerFeedbackRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderAnswerFeedbacksTable(cmd, rows)
}

func parseAnswerFeedbacksListOptions(cmd *cobra.Command) (answerFeedbacksListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	answer, _ := cmd.Flags().GetString("answer")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return answerFeedbacksListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Answer:  answer,
	}, nil
}

func buildAnswerFeedbackRows(resp jsonAPIResponse) []answerFeedbackRow {
	rows := make([]answerFeedbackRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := answerFeedbackRow{
			ID:            resource.ID,
			Score:         floatAttrPointer(resource.Attributes, "score"),
			Notes:         strings.TrimSpace(stringAttr(resource.Attributes, "notes")),
			BetterContent: strings.TrimSpace(stringAttr(resource.Attributes, "better-content")),
			AnswerID:      relationshipIDFromMap(resource.Relationships, "answer"),
		}
		rows = append(rows, row)
	}
	return rows
}

func buildAnswerFeedbackRowFromSingle(resp jsonAPISingleResponse) answerFeedbackRow {
	attrs := resp.Data.Attributes
	return answerFeedbackRow{
		ID:            resp.Data.ID,
		Score:         floatAttrPointer(attrs, "score"),
		Notes:         strings.TrimSpace(stringAttr(attrs, "notes")),
		BetterContent: strings.TrimSpace(stringAttr(attrs, "better-content")),
		AnswerID:      relationshipIDFromMap(resp.Data.Relationships, "answer"),
	}
}

func renderAnswerFeedbacksTable(cmd *cobra.Command, rows []answerFeedbackRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No answer feedbacks found.")
		return nil
	}

	const notesMax = 40
	const contentMax = 40

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSCORE\tNOTES\tBETTER CONTENT\tANSWER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			formatAnswerFeedbackScore(row.Score),
			truncateString(row.Notes, notesMax),
			truncateString(row.BetterContent, contentMax),
			row.AnswerID,
		)
	}
	return writer.Flush()
}

func formatAnswerFeedbackScore(score *float64) string {
	if score == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *score)
}
