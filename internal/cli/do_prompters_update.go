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

type doPromptersUpdateOptions struct {
	BaseURL                                                 string
	Token                                                   string
	JSON                                                    bool
	Name                                                    string
	IsActive                                                bool
	ReleaseNoteGuessHasNavigationInstructionsPromptTemplate string
	ReleaseNoteHeadlineSuggestionsPromptTemplate            string
	ReleaseNoteGlossaryTermSuggestionsPromptTemplate        string
	JPPSafetyRisksSuggestionSuggestionPromptTemplate        string
	JPPSafetyRiskCommSuggestionSuggestionPromptTemplate     string
	IncidentHeadlineSuggestionSuggestionPromptTemplate      string
	GlossaryTermDefinitionSuggestionsPromptTemplate         string
	CondensableCondensePromptTemplate                       string
	AnswerAnswerPromptTemplate                              string
	ActionItemSummaryPromptTemplate                         string
}

func newDoPromptersUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a prompter",
		Long: `Update an existing prompter.

Only the fields you specify will be updated. Fields not provided will remain unchanged.

Arguments:
  <id>    The prompter ID (required)

Optional flags:
  --name                                                    Prompter name
  --is-active                                               Set as active (true/false)
  --release-note-guess-has-navigation-instructions-prompt-template   Template for release note navigation instructions
  --release-note-headline-suggestions-prompt-template       Template for release note headline suggestions
  --release-note-glossary-term-suggestions-prompt-template   Template for release note glossary term suggestions
  --jpp-safety-risks-suggestion-suggestion-prompt-template   Template for JPP safety risks suggestions
  --jpp-safety-risk-comm-suggestion-suggestion-prompt-template Template for JPP safety risk communication suggestions
  --incident-headline-suggestion-suggestion-prompt-template  Template for incident headline suggestions
  --glossary-term-definition-suggestions-prompt-template     Template for glossary term definition suggestions
  --condensable-condense-prompt-template                     Template for content condensation
  --answer-answer-prompt-template                            Template for answer generation
  --action-item-summary-prompt-template                      Template for action item summaries

Global flags (see xbe --help): --json, --base-url, --token`,
		Example: `  # Update a prompter name
  xbe do prompters update 123 --name "Updated Name"

  # Update a template
  xbe do prompters update 123 --release-note-headline-suggestions-prompt-template "New template"

  # Deactivate a prompter
  xbe do prompters update 123 --is-active=false

  # Get JSON output
  xbe do prompters update 123 --name "Updated" --json`,
		Args: cobra.ExactArgs(1),
		RunE: runDoPromptersUpdate,
	}
	initDoPromptersUpdateFlags(cmd)
	return cmd
}

func init() {
	doPromptersCmd.AddCommand(newDoPromptersUpdateCmd())
}

func initDoPromptersUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("name", "", "Prompter name")
	cmd.Flags().Bool("is-active", false, "Set as active")
	cmd.Flags().String("release-note-guess-has-navigation-instructions-prompt-template", "", "Template for release note navigation instructions")
	cmd.Flags().String("release-note-headline-suggestions-prompt-template", "", "Template for release note headline suggestions")
	cmd.Flags().String("release-note-glossary-term-suggestions-prompt-template", "", "Template for release note glossary term suggestions")
	cmd.Flags().String("jpp-safety-risks-suggestion-suggestion-prompt-template", "", "Template for JPP safety risks suggestions")
	cmd.Flags().String("jpp-safety-risk-comm-suggestion-suggestion-prompt-template", "", "Template for JPP safety risk communication suggestions")
	cmd.Flags().String("incident-headline-suggestion-suggestion-prompt-template", "", "Template for incident headline suggestions")
	cmd.Flags().String("glossary-term-definition-suggestions-prompt-template", "", "Template for glossary term definition suggestions")
	cmd.Flags().String("condensable-condense-prompt-template", "", "Template for content condensation")
	cmd.Flags().String("answer-answer-prompt-template", "", "Template for answer generation")
	cmd.Flags().String("action-item-summary-prompt-template", "", "Template for action item summaries")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoPromptersUpdate(cmd *cobra.Command, args []string) error {
	opts, err := parseDoPromptersUpdateOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("prompter id is required")
	}

	attributes := map[string]any{}

	if cmd.Flags().Changed("name") {
		attributes["name"] = opts.Name
	}
	if cmd.Flags().Changed("is-active") {
		attributes["is-active"] = opts.IsActive
	}
	if cmd.Flags().Changed("release-note-guess-has-navigation-instructions-prompt-template") {
		attributes["release-note-guess-has-navigation-instructions-prompt-template"] = opts.ReleaseNoteGuessHasNavigationInstructionsPromptTemplate
	}
	if cmd.Flags().Changed("release-note-headline-suggestions-prompt-template") {
		attributes["release-note-headline-suggestions-prompt-template"] = opts.ReleaseNoteHeadlineSuggestionsPromptTemplate
	}
	if cmd.Flags().Changed("release-note-glossary-term-suggestions-prompt-template") {
		attributes["release-note-glossary-term-suggestions-prompt-template"] = opts.ReleaseNoteGlossaryTermSuggestionsPromptTemplate
	}
	if cmd.Flags().Changed("jpp-safety-risks-suggestion-suggestion-prompt-template") {
		attributes["jpp-safety-risks-suggestion-suggestion-prompt-template"] = opts.JPPSafetyRisksSuggestionSuggestionPromptTemplate
	}
	if cmd.Flags().Changed("jpp-safety-risk-comm-suggestion-suggestion-prompt-template") {
		attributes["jpp-safety-risk-comm-suggestion-suggestion-prompt-template"] = opts.JPPSafetyRiskCommSuggestionSuggestionPromptTemplate
	}
	if cmd.Flags().Changed("incident-headline-suggestion-suggestion-prompt-template") {
		attributes["incident-headline-suggestion-suggestion-prompt-template"] = opts.IncidentHeadlineSuggestionSuggestionPromptTemplate
	}
	if cmd.Flags().Changed("glossary-term-definition-suggestions-prompt-template") {
		attributes["glossary-term-definition-suggestions-prompt-template"] = opts.GlossaryTermDefinitionSuggestionsPromptTemplate
	}
	if cmd.Flags().Changed("condensable-condense-prompt-template") {
		attributes["condensable-condense-prompt-template"] = opts.CondensableCondensePromptTemplate
	}
	if cmd.Flags().Changed("answer-answer-prompt-template") {
		attributes["answer-answer-prompt-template"] = opts.AnswerAnswerPromptTemplate
	}
	if cmd.Flags().Changed("action-item-summary-prompt-template") {
		attributes["action-item-summary-prompt-template"] = opts.ActionItemSummaryPromptTemplate
	}

	if len(attributes) == 0 {
		err := errors.New("no fields to update")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"id":         id,
			"type":       "prompters",
			"attributes": attributes,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Patch(cmd.Context(), "/v1/prompters/"+id, jsonBody)
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

	if row.Name != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Updated prompter %s (%s)\n", row.ID, row.Name)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Updated prompter %s\n", row.ID)
	return nil
}

func parseDoPromptersUpdateOptions(cmd *cobra.Command) (doPromptersUpdateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	name, _ := cmd.Flags().GetString("name")
	isActive, _ := cmd.Flags().GetBool("is-active")
	releaseNoteGuessHasNavigationInstructionsPromptTemplate, _ := cmd.Flags().GetString("release-note-guess-has-navigation-instructions-prompt-template")
	releaseNoteHeadlineSuggestionsPromptTemplate, _ := cmd.Flags().GetString("release-note-headline-suggestions-prompt-template")
	releaseNoteGlossaryTermSuggestionsPromptTemplate, _ := cmd.Flags().GetString("release-note-glossary-term-suggestions-prompt-template")
	jppSafetyRisksSuggestionSuggestionPromptTemplate, _ := cmd.Flags().GetString("jpp-safety-risks-suggestion-suggestion-prompt-template")
	jppSafetyRiskCommSuggestionSuggestionPromptTemplate, _ := cmd.Flags().GetString("jpp-safety-risk-comm-suggestion-suggestion-prompt-template")
	incidentHeadlineSuggestionSuggestionPromptTemplate, _ := cmd.Flags().GetString("incident-headline-suggestion-suggestion-prompt-template")
	glossaryTermDefinitionSuggestionsPromptTemplate, _ := cmd.Flags().GetString("glossary-term-definition-suggestions-prompt-template")
	condensableCondensePromptTemplate, _ := cmd.Flags().GetString("condensable-condense-prompt-template")
	answerAnswerPromptTemplate, _ := cmd.Flags().GetString("answer-answer-prompt-template")
	actionItemSummaryPromptTemplate, _ := cmd.Flags().GetString("action-item-summary-prompt-template")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doPromptersUpdateOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		Name:     name,
		IsActive: isActive,
		ReleaseNoteGuessHasNavigationInstructionsPromptTemplate: releaseNoteGuessHasNavigationInstructionsPromptTemplate,
		ReleaseNoteHeadlineSuggestionsPromptTemplate:            releaseNoteHeadlineSuggestionsPromptTemplate,
		ReleaseNoteGlossaryTermSuggestionsPromptTemplate:        releaseNoteGlossaryTermSuggestionsPromptTemplate,
		JPPSafetyRisksSuggestionSuggestionPromptTemplate:        jppSafetyRisksSuggestionSuggestionPromptTemplate,
		JPPSafetyRiskCommSuggestionSuggestionPromptTemplate:     jppSafetyRiskCommSuggestionSuggestionPromptTemplate,
		IncidentHeadlineSuggestionSuggestionPromptTemplate:      incidentHeadlineSuggestionSuggestionPromptTemplate,
		GlossaryTermDefinitionSuggestionsPromptTemplate:         glossaryTermDefinitionSuggestionsPromptTemplate,
		CondensableCondensePromptTemplate:                       condensableCondensePromptTemplate,
		AnswerAnswerPromptTemplate:                              answerAnswerPromptTemplate,
		ActionItemSummaryPromptTemplate:                         actionItemSummaryPromptTemplate,
	}, nil
}
