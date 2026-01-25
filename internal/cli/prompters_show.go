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

type promptersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

func newPromptersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show prompter details",
		Long: `Show the full details of a specific prompter.

Prompters store prompt templates used for AI-assisted features like release
note summarization, glossary term definitions, and safety risk suggestions.

Output Fields:
  ID                                               Prompter identifier
  NAME                                             Prompter name (if set)
  ACTIVE                                           Whether the prompter is active
  RELEASE NOTE GUESS HAS NAVIGATION INSTRUCTIONS   Prompt template
  RELEASE NOTE HEADLINE SUGGESTIONS                Prompt template
  RELEASE NOTE GLOSSARY TERM SUGGESTIONS           Prompt template
  JPP SAFETY RISKS SUGGESTION                      Prompt template
  JPP SAFETY RISK COMM SUGGESTION                  Prompt template
  INCIDENT HEADLINE SUGGESTION                     Prompt template
  GLOSSARY TERM DEFINITION SUGGESTIONS             Prompt template
  CONDENSABLE CONDENSE                             Prompt template
  ANSWER ANSWER                                    Prompt template
  ACTION ITEM SUMMARY                              Prompt template

Arguments:
  <id>    The prompter ID (required). You can find IDs using the list command.

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # View a prompter
  xbe view prompters show 123

  # Get JSON output
  xbe view prompters show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPromptersShow,
	}
	initPromptersShowFlags(cmd)
	return cmd
}

func init() {
	promptersCmd.AddCommand(newPromptersShowCmd())
}

func initPromptersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPromptersShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePromptersShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("prompter id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[prompters]", strings.Join([]string{
		"name",
		"is-active",
		"release-note-guess-has-navigation-instructions-prompt-template",
		"release-note-headline-suggestions-prompt-template",
		"release-note-glossary-term-suggestions-prompt-template",
		"jpp-safety-risks-suggestion-suggestion-prompt-template",
		"jpp-safety-risk-comm-suggestion-suggestion-prompt-template",
		"incident-headline-suggestion-suggestion-prompt-template",
		"glossary-term-definition-suggestions-prompt-template",
		"condensable-condense-prompt-template",
		"answer-answer-prompt-template",
		"action-item-summary-prompt-template",
	}, ","))

	body, _, err := client.Get(cmd.Context(), "/v1/prompters/"+id, query)
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

	row := buildPrompterRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	return renderPrompterDetails(cmd, row)
}

func parsePromptersShowOptions(cmd *cobra.Command) (promptersShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return promptersShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return promptersShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return promptersShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return promptersShowOptions{}, err
	}

	return promptersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func renderPrompterDetails(cmd *cobra.Command, row prompterRow) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", row.ID)
	if row.Name != "" {
		fmt.Fprintf(out, "Name: %s\n", row.Name)
	}
	fmt.Fprintf(out, "Active: %t\n", row.IsActive)
	if row.ReleaseNoteGuessHasNavigationInstructionsPromptTemplate != "" {
		fmt.Fprintf(out, "Release Note Guess Has Navigation Instructions Prompt Template: %s\n", row.ReleaseNoteGuessHasNavigationInstructionsPromptTemplate)
	}
	if row.ReleaseNoteHeadlineSuggestionsPromptTemplate != "" {
		fmt.Fprintf(out, "Release Note Headline Suggestions Prompt Template: %s\n", row.ReleaseNoteHeadlineSuggestionsPromptTemplate)
	}
	if row.ReleaseNoteGlossaryTermSuggestionsPromptTemplate != "" {
		fmt.Fprintf(out, "Release Note Glossary Term Suggestions Prompt Template: %s\n", row.ReleaseNoteGlossaryTermSuggestionsPromptTemplate)
	}
	if row.JPPSafetyRisksSuggestionSuggestionPromptTemplate != "" {
		fmt.Fprintf(out, "JPP Safety Risks Suggestion Prompt Template: %s\n", row.JPPSafetyRisksSuggestionSuggestionPromptTemplate)
	}
	if row.JPPSafetyRiskCommSuggestionSuggestionPromptTemplate != "" {
		fmt.Fprintf(out, "JPP Safety Risk Comm Suggestion Prompt Template: %s\n", row.JPPSafetyRiskCommSuggestionSuggestionPromptTemplate)
	}
	if row.IncidentHeadlineSuggestionSuggestionPromptTemplate != "" {
		fmt.Fprintf(out, "Incident Headline Suggestion Prompt Template: %s\n", row.IncidentHeadlineSuggestionSuggestionPromptTemplate)
	}
	if row.GlossaryTermDefinitionSuggestionsPromptTemplate != "" {
		fmt.Fprintf(out, "Glossary Term Definition Suggestions Prompt Template: %s\n", row.GlossaryTermDefinitionSuggestionsPromptTemplate)
	}
	if row.CondensableCondensePromptTemplate != "" {
		fmt.Fprintf(out, "Condensable Condense Prompt Template: %s\n", row.CondensableCondensePromptTemplate)
	}
	if row.AnswerAnswerPromptTemplate != "" {
		fmt.Fprintf(out, "Answer Answer Prompt Template: %s\n", row.AnswerAnswerPromptTemplate)
	}
	if row.ActionItemSummaryPromptTemplate != "" {
		fmt.Fprintf(out, "Action Item Summary Prompt Template: %s\n", row.ActionItemSummaryPromptTemplate)
	}

	return nil
}
