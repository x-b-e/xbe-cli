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

type answersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type answerDetails struct {
	ID                            string   `json:"id"`
	Content                       string   `json:"content,omitempty"`
	Prompt                        string   `json:"prompt,omitempty"`
	QuestionID                    string   `json:"question_id,omitempty"`
	FeedbackID                    string   `json:"feedback_id,omitempty"`
	RelatedContentIDs             []string `json:"related_content_ids,omitempty"`
	RelatedGlossaryTermContentIDs []string `json:"related_glossary_term_content_ids,omitempty"`
	RelatedNewsletterContentIDs   []string `json:"related_newsletter_content_ids,omitempty"`
	RelatedReleaseNoteContentIDs  []string `json:"related_release_note_content_ids,omitempty"`
	RelatedQuestionContentIDs     []string `json:"related_question_content_ids,omitempty"`
	RelatedPressReleaseContentIDs []string `json:"related_press_release_content_ids,omitempty"`
	RelatedObjectiveContentIDs    []string `json:"related_objective_content_ids,omitempty"`
	RelatedFeatureContentIDs      []string `json:"related_feature_content_ids,omitempty"`
}

func newAnswersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show answer details",
		Long: `Show the full details of an answer.

Output Fields:
  ID                             Answer identifier
  QUESTION ID                    Associated question ID
  FEEDBACK ID                    Associated answer feedback ID
  PROMPT                         Prompt content
  CONTENT                        Answer content
  RELATED CONTENT IDS            Related content entries (answer-related-contents)
  RELATED GLOSSARY TERM IDS      Related glossary term content entries
  RELATED NEWSLETTER IDS         Related newsletter content entries
  RELATED RELEASE NOTE IDS       Related release note content entries
  RELATED QUESTION IDS           Related question content entries
  RELATED PRESS RELEASE IDS      Related press release content entries
  RELATED OBJECTIVE IDS          Related objective content entries
  RELATED FEATURE IDS            Related feature content entries

Global flags (see xbe --help): --json, --no-auth, --base-url, --token`,
		Example: `  # Show answer details
  xbe view answers show 123

  # Output as JSON
  xbe view answers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runAnswersShow,
	}
	initAnswersShowFlags(cmd)
	return cmd
}

func init() {
	answersCmd.AddCommand(newAnswersShowCmd())
}

func initAnswersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runAnswersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseAnswersShowOptions(cmd)
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
		return fmt.Errorf("answer id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[answers]", "content,prompt,question,feedback,related-contents,related-glossary-term-contents,related-newsletter-contents,related-release-note-contents,related-question-contents,related-press-release-contents,related-objective-contents,related-feature-contents")

	body, _, err := client.Get(cmd.Context(), "/v1/answers/"+id, query)
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

	details := buildAnswerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderAnswerDetails(cmd, details)
}

func parseAnswersShowOptions(cmd *cobra.Command) (answersShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return answersShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return answersShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return answersShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return answersShowOptions{}, err
	}

	return answersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildAnswerDetails(resp jsonAPISingleResponse) answerDetails {
	resource := resp.Data
	attrs := resource.Attributes
	return answerDetails{
		ID:                            resource.ID,
		Content:                       strings.TrimSpace(stringAttr(attrs, "content")),
		Prompt:                        strings.TrimSpace(stringAttr(attrs, "prompt")),
		QuestionID:                    relationshipIDFromMap(resource.Relationships, "question"),
		FeedbackID:                    relationshipIDFromMap(resource.Relationships, "feedback"),
		RelatedContentIDs:             relationshipIDsFromMap(resource.Relationships, "related-contents"),
		RelatedGlossaryTermContentIDs: relationshipIDsFromMap(resource.Relationships, "related-glossary-term-contents"),
		RelatedNewsletterContentIDs:   relationshipIDsFromMap(resource.Relationships, "related-newsletter-contents"),
		RelatedReleaseNoteContentIDs:  relationshipIDsFromMap(resource.Relationships, "related-release-note-contents"),
		RelatedQuestionContentIDs:     relationshipIDsFromMap(resource.Relationships, "related-question-contents"),
		RelatedPressReleaseContentIDs: relationshipIDsFromMap(resource.Relationships, "related-press-release-contents"),
		RelatedObjectiveContentIDs:    relationshipIDsFromMap(resource.Relationships, "related-objective-contents"),
		RelatedFeatureContentIDs:      relationshipIDsFromMap(resource.Relationships, "related-feature-contents"),
	}
}

func renderAnswerDetails(cmd *cobra.Command, details answerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.QuestionID != "" {
		fmt.Fprintf(out, "Question ID: %s\n", details.QuestionID)
	}
	if details.FeedbackID != "" {
		fmt.Fprintf(out, "Feedback ID: %s\n", details.FeedbackID)
	}
	if details.Prompt != "" {
		fmt.Fprintf(out, "Prompt: %s\n", details.Prompt)
	}
	if details.Content != "" {
		fmt.Fprintf(out, "Content: %s\n", details.Content)
	}
	if len(details.RelatedContentIDs) > 0 {
		fmt.Fprintf(out, "Related Content IDs: %s\n", strings.Join(details.RelatedContentIDs, ", "))
	}
	if len(details.RelatedGlossaryTermContentIDs) > 0 {
		fmt.Fprintf(out, "Related Glossary Term Content IDs: %s\n", strings.Join(details.RelatedGlossaryTermContentIDs, ", "))
	}
	if len(details.RelatedNewsletterContentIDs) > 0 {
		fmt.Fprintf(out, "Related Newsletter Content IDs: %s\n", strings.Join(details.RelatedNewsletterContentIDs, ", "))
	}
	if len(details.RelatedReleaseNoteContentIDs) > 0 {
		fmt.Fprintf(out, "Related Release Note Content IDs: %s\n", strings.Join(details.RelatedReleaseNoteContentIDs, ", "))
	}
	if len(details.RelatedQuestionContentIDs) > 0 {
		fmt.Fprintf(out, "Related Question Content IDs: %s\n", strings.Join(details.RelatedQuestionContentIDs, ", "))
	}
	if len(details.RelatedPressReleaseContentIDs) > 0 {
		fmt.Fprintf(out, "Related Press Release Content IDs: %s\n", strings.Join(details.RelatedPressReleaseContentIDs, ", "))
	}
	if len(details.RelatedObjectiveContentIDs) > 0 {
		fmt.Fprintf(out, "Related Objective Content IDs: %s\n", strings.Join(details.RelatedObjectiveContentIDs, ", "))
	}
	if len(details.RelatedFeatureContentIDs) > 0 {
		fmt.Fprintf(out, "Related Feature Content IDs: %s\n", strings.Join(details.RelatedFeatureContentIDs, ", "))
	}

	return nil
}
