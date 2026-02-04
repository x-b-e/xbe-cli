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

type questionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type questionDetails struct {
	ID                                  string   `json:"id"`
	Content                             string   `json:"content,omitempty"`
	Source                              string   `json:"source,omitempty"`
	IsPublic                            bool     `json:"is_public"`
	IsPublicToOrganizationChildren      bool     `json:"is_public_to_organization_children"`
	IgnoreOrganizationScopedNewsletters bool     `json:"ignore_organization_scoped_newsletters"`
	Motivation                          string   `json:"motivation,omitempty"`
	MotivationGuess                     string   `json:"motivation_guess,omitempty"`
	IsTriaged                           bool     `json:"is_triaged"`
	TriagedAt                           string   `json:"triaged_at,omitempty"`
	BestAnswerContent                   string   `json:"best_answer_content,omitempty"`
	AskedByID                           string   `json:"asked_by_id,omitempty"`
	CreatedByID                         string   `json:"created_by_id,omitempty"`
	AssignedToID                        string   `json:"assigned_to_id,omitempty"`
	PublicOrganizationScopeType         string   `json:"public_organization_scope_type,omitempty"`
	PublicOrganizationScopeID           string   `json:"public_organization_scope_id,omitempty"`
	AnswerID                            string   `json:"answer_id,omitempty"`
	AnswerIDs                           []string `json:"answer_ids,omitempty"`
}

func newQuestionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show question details",
		Long: `Show the full details of a question.

Output Fields:
  ID                                  Question identifier
  Content                             Question text
  Source                              Question source (app/link)
  Motivation                          Motivation label (serious/silly)
  Motivation Guess                    Suggested motivation
  Best Answer Content                 Best answer content (if any)
  Triaged At                          Triaged timestamp
  Is Public                           Whether the question is public
  Is Public to Organization Children  Whether public to org children
  Ignore Organization Scoped Newsletters  Ignore newsletter scoping
  Is Triaged                          Whether the question is triaged
  Asked By                            Asking user ID
  Created By                          Creator user ID
  Assigned To                         Assigned user ID
  Public Organization Scope           Scope type and ID
  Answer                              Latest answer ID
  Answers                             Answer IDs

Arguments:
  <id>  The question ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show question details
  xbe view questions show 123

  # Output as JSON
  xbe view questions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runQuestionsShow,
	}
	initQuestionsShowFlags(cmd)
	return cmd
}

func init() {
	questionsCmd.AddCommand(newQuestionsShowCmd())
}

func initQuestionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runQuestionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseQuestionsShowOptions(cmd)
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
		return fmt.Errorf("question id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[questions]", strings.Join([]string{
		"content",
		"source",
		"is-public",
		"is-public-to-organization-children",
		"ignore-organization-scoped-newsletters",
		"motivation",
		"motivation-guess",
		"is-triaged",
		"triaged-at",
		"best-answer-content",
		"asked-by",
		"created-by",
		"assigned-to",
		"public-organization-scope",
		"answer",
		"answers",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/questions/"+id, query)
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

	details := buildQuestionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderQuestionDetails(cmd, details)
}

func parseQuestionsShowOptions(cmd *cobra.Command) (questionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return questionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildQuestionDetails(resp jsonAPISingleResponse) questionDetails {
	attrs := resp.Data.Attributes
	resource := resp.Data

	details := questionDetails{
		ID:                                  resource.ID,
		Content:                             stringAttr(attrs, "content"),
		Source:                              stringAttr(attrs, "source"),
		IsPublic:                            boolAttr(attrs, "is-public"),
		IsPublicToOrganizationChildren:      boolAttr(attrs, "is-public-to-organization-children"),
		IgnoreOrganizationScopedNewsletters: boolAttr(attrs, "ignore-organization-scoped-newsletters"),
		Motivation:                          stringAttr(attrs, "motivation"),
		MotivationGuess:                     stringAttr(attrs, "motivation-guess"),
		IsTriaged:                           boolAttr(attrs, "is-triaged"),
		TriagedAt:                           formatDateTime(stringAttr(attrs, "triaged-at")),
		BestAnswerContent:                   stringAttr(attrs, "best-answer-content"),
	}

	if rel, ok := resource.Relationships["asked-by"]; ok && rel.Data != nil {
		details.AskedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		details.CreatedByID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["assigned-to"]; ok && rel.Data != nil {
		details.AssignedToID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["public-organization-scope"]; ok && rel.Data != nil {
		details.PublicOrganizationScopeType = rel.Data.Type
		details.PublicOrganizationScopeID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["answer"]; ok && rel.Data != nil {
		details.AnswerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["answers"]; ok {
		details.AnswerIDs = relationshipIDsToStrings(rel)
	}

	return details
}

func renderQuestionDetails(cmd *cobra.Command, details questionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Content != "" {
		fmt.Fprintf(out, "Content: %s\n", details.Content)
	}
	if details.Source != "" {
		fmt.Fprintf(out, "Source: %s\n", details.Source)
	}
	if details.Motivation != "" {
		fmt.Fprintf(out, "Motivation: %s\n", details.Motivation)
	}
	if details.MotivationGuess != "" {
		fmt.Fprintf(out, "Motivation Guess: %s\n", details.MotivationGuess)
	}
	if details.BestAnswerContent != "" {
		fmt.Fprintf(out, "Best Answer Content: %s\n", details.BestAnswerContent)
	}
	if details.TriagedAt != "" {
		fmt.Fprintf(out, "Triaged At: %s\n", details.TriagedAt)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	fmt.Fprintf(out, "  Is Public: %s\n", formatBool(details.IsPublic))
	fmt.Fprintf(out, "  Is Public to Organization Children: %s\n", formatBool(details.IsPublicToOrganizationChildren))
	fmt.Fprintf(out, "  Ignore Organization Scoped Newsletters: %s\n", formatBool(details.IgnoreOrganizationScopedNewsletters))
	fmt.Fprintf(out, "  Is Triaged: %s\n", formatBool(details.IsTriaged))

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Relationships:")
	fmt.Fprintln(out, strings.Repeat("-", 40))
	if details.AskedByID != "" {
		fmt.Fprintf(out, "  Asked By: %s\n", details.AskedByID)
	}
	if details.CreatedByID != "" {
		fmt.Fprintf(out, "  Created By: %s\n", details.CreatedByID)
	}
	if details.AssignedToID != "" {
		fmt.Fprintf(out, "  Assigned To: %s\n", details.AssignedToID)
	}
	if details.PublicOrganizationScopeType != "" && details.PublicOrganizationScopeID != "" {
		fmt.Fprintf(out, "  Public Organization Scope: %s/%s\n", details.PublicOrganizationScopeType, details.PublicOrganizationScopeID)
	}
	if details.AnswerID != "" {
		fmt.Fprintf(out, "  Answer: %s\n", details.AnswerID)
	}
	if len(details.AnswerIDs) > 0 {
		fmt.Fprintf(out, "  Answers: %s\n", strings.Join(details.AnswerIDs, ", "))
	}

	return nil
}
