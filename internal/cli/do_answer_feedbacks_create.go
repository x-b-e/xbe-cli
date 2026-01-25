package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type doAnswerFeedbacksCreateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	Answer        string
	Score         float64
	Notes         string
	BetterContent string
}

func newDoAnswerFeedbacksCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create answer feedback",
		Long: `Create answer feedback for an answer.

Required flags:
  --answer          Answer ID (required)

Optional flags:
  --score           Feedback score between 0 and 1
  --notes           Feedback notes
  --better-content  Suggested improved content

Note: Only admin users can create answer feedbacks.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Create answer feedback with a score
  xbe do answer-feedbacks create --answer 123 --score 0.8 --notes "Good answer"

  # Provide improved content suggestion
  xbe do answer-feedbacks create --answer 123 --better-content "Reworded answer"

  # Output as JSON
  xbe do answer-feedbacks create --answer 123 --score 0.7 --json`,
		Args: cobra.NoArgs,
		RunE: runDoAnswerFeedbacksCreate,
	}
	initDoAnswerFeedbacksCreateFlags(cmd)
	return cmd
}

func init() {
	doAnswerFeedbacksCmd.AddCommand(newDoAnswerFeedbacksCreateCmd())
}

func initDoAnswerFeedbacksCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("answer", "", "Answer ID (required)")
	cmd.Flags().Float64("score", 0, "Feedback score between 0 and 1")
	cmd.Flags().String("notes", "", "Feedback notes")
	cmd.Flags().String("better-content", "", "Suggested improved content")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoAnswerFeedbacksCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoAnswerFeedbacksCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
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

	if strings.TrimSpace(opts.Answer) == "" {
		err := fmt.Errorf("--answer is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if cmd.Flags().Changed("score") {
		if opts.Score < 0 || opts.Score > 1 {
			err := fmt.Errorf("--score must be between 0 and 1")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	attributes := map[string]any{}
	if cmd.Flags().Changed("score") {
		attributes["score"] = opts.Score
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("better-content") {
		attributes["better-content"] = opts.BetterContent
	}

	relationships := map[string]any{
		"answer": map[string]any{
			"data": map[string]any{
				"type": "answers",
				"id":   opts.Answer,
			},
		},
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "answer-feedbacks",
			"attributes":    attributes,
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/answer-feedbacks", jsonBody)
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

	row := buildAnswerFeedbackRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created answer feedback %s\n", row.ID)
	return nil
}

func parseDoAnswerFeedbacksCreateOptions(cmd *cobra.Command) (doAnswerFeedbacksCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	answer, _ := cmd.Flags().GetString("answer")
	score, _ := cmd.Flags().GetFloat64("score")
	notes, _ := cmd.Flags().GetString("notes")
	betterContent, _ := cmd.Flags().GetString("better-content")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doAnswerFeedbacksCreateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		Answer:        answer,
		Score:         score,
		Notes:         notes,
		BetterContent: betterContent,
	}, nil
}
