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

type doAnswerFeedbacksUpdateOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	ID            string
	Score         float64
	Notes         string
	BetterContent string
}

func newDoAnswerFeedbacksUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update answer feedback",
		Long: `Update answer feedback fields.

Provide the answer feedback ID as an argument, then use flags to specify
which fields to update. Only specified fields will be modified.

Updatable fields:
  --score           Feedback score between 0 and 1
  --notes           Feedback notes
  --better-content  Suggested improved content

Note: Answer relationship cannot be changed after creation.
Only admin users can update answer feedbacks.

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update score
  xbe do answer-feedbacks update 123 --score 0.9

  # Update notes and better content
  xbe do answer-feedbacks update 123 --notes \"Refined\" --better-content \"Updated answer\"

  # Output as JSON
  xbe do answer-feedbacks update 123 --score 0.7 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoAnswerFeedbacksUpdate,
	}
	initDoAnswerFeedbacksUpdateFlags(cmd)
	return cmd
}

func init() {
	doAnswerFeedbacksCmd.AddCommand(newDoAnswerFeedbacksUpdateCmd())
}

func initDoAnswerFeedbacksUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Float64("score", 0, "Feedback score between 0 and 1")
	cmd.Flags().String("notes", "", "Feedback notes")
	cmd.Flags().String("better-content", "", "Suggested improved content")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoAnswerFeedbacksUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoAnswerFeedbacksUpdateOptions(cmd, args)
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

	attributes := map[string]any{}
	if cmd.Flags().Changed("score") {
		if opts.Score < 0 || opts.Score > 1 {
			err := fmt.Errorf("--score must be between 0 and 1")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		attributes["score"] = opts.Score
	}
	if cmd.Flags().Changed("notes") {
		attributes["notes"] = opts.Notes
	}
	if cmd.Flags().Changed("better-content") {
		attributes["better-content"] = opts.BetterContent
	}

	if len(attributes) == 0 {
		err := fmt.Errorf("no fields to update; specify --score, --notes, or --better-content")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":       "answer-feedbacks",
			"id":         opts.ID,
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/answer-feedbacks/"+opts.ID, jsonBody)
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

	fmt.Fprintf(cmd.OutOrStdout(), "Updated answer feedback %s\n", row.ID)
	return nil
}

func parseDoAnswerFeedbacksUpdateOptions(cmd *cobra.Command, args []string) (doAnswerFeedbacksUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	score, _ := cmd.Flags().GetFloat64("score")
	notes, _ := cmd.Flags().GetString("notes")
	betterContent, _ := cmd.Flags().GetString("better-content")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doAnswerFeedbacksUpdateOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		ID:            args[0],
		Score:         score,
		Notes:         notes,
		BetterContent: betterContent,
	}, nil
}
