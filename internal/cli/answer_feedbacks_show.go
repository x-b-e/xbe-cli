package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type answerFeedbacksShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type answerFeedbackDetails struct {
	ID            string   `json:"id"`
	Score         *float64 `json:"score,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	BetterContent string   `json:"better_content,omitempty"`
	AnswerID      string   `json:"answer_id,omitempty"`
}

func newAnswerFeedbacksShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show answer feedback details",
		Long: `Show the full details of an answer feedback record.

Output Fields:
  ID              Answer feedback identifier
  ANSWER ID       Associated answer ID
  SCORE           Feedback score (0 to 1)
  NOTES           Feedback notes
  BETTER CONTENT  Suggested improved content

Global flags (see xbe --help): --json, --no-auth, --base-url, --token`,
		Example: `  # Show answer feedback details
  xbe view answer-feedbacks show 123

  # Output as JSON
  xbe view answer-feedbacks show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runAnswerFeedbacksShow,
	}
	initAnswerFeedbacksShowFlags(cmd)
	return cmd
}

func init() {
	answerFeedbacksCmd.AddCommand(newAnswerFeedbacksShowCmd())
}

func initAnswerFeedbacksShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runAnswerFeedbacksShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseAnswerFeedbacksShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("answer feedback id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[answer-feedbacks]", "score,notes,better-content,answer")

	body, _, err := client.Get(cmd.Context(), "/v1/answer-feedbacks/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildAnswerFeedbackDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderAnswerFeedbackDetails(cmd, details)
}

func parseAnswerFeedbacksShowOptions(cmd *cobra.Command) (answerFeedbacksShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return answerFeedbacksShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return answerFeedbacksShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return answerFeedbacksShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return answerFeedbacksShowOptions{}, err
	}

	return answerFeedbacksShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildAnswerFeedbackDetails(resp jsonAPISingleResponse) answerFeedbackDetails {
	attrs := resp.Data.Attributes
	return answerFeedbackDetails{
		ID:            resp.Data.ID,
		Score:         floatAttrPointer(attrs, "score"),
		Notes:         strings.TrimSpace(stringAttr(attrs, "notes")),
		BetterContent: strings.TrimSpace(stringAttr(attrs, "better-content")),
		AnswerID:      relationshipIDFromMap(resp.Data.Relationships, "answer"),
	}
}

func renderAnswerFeedbackDetails(cmd *cobra.Command, details answerFeedbackDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.AnswerID != "" {
		fmt.Fprintf(out, "Answer ID: %s\n", details.AnswerID)
	}
	if details.Score != nil {
		fmt.Fprintf(out, "Score: %s\n", formatAnswerFeedbackScore(details.Score))
	}
	if details.Notes != "" {
		fmt.Fprintf(out, "Notes: %s\n", details.Notes)
	}
	if details.BetterContent != "" {
		fmt.Fprintf(out, "Better content: %s\n", details.BetterContent)
	}

	return nil
}
