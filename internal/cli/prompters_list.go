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

type promptersListOptions struct {
	BaseURL  string
	Token    string
	JSON     bool
	NoAuth   bool
	Limit    int
	Offset   int
	IsActive string
}

type prompterRow struct {
	ID                                                      string `json:"id"`
	Name                                                    string `json:"name,omitempty"`
	IsActive                                                bool   `json:"is_active"`
	ReleaseNoteGuessHasNavigationInstructionsPromptTemplate string `json:"release_note_guess_has_navigation_instructions_prompt_template,omitempty"`
	ReleaseNoteHeadlineSuggestionsPromptTemplate            string `json:"release_note_headline_suggestions_prompt_template,omitempty"`
	ReleaseNoteGlossaryTermSuggestionsPromptTemplate        string `json:"release_note_glossary_term_suggestions_prompt_template,omitempty"`
	JPPSafetyRisksSuggestionSuggestionPromptTemplate        string `json:"jpp_safety_risks_suggestion_suggestion_prompt_template,omitempty"`
	JPPSafetyRiskCommSuggestionSuggestionPromptTemplate     string `json:"jpp_safety_risk_comm_suggestion_suggestion_prompt_template,omitempty"`
	IncidentHeadlineSuggestionSuggestionPromptTemplate      string `json:"incident_headline_suggestion_suggestion_prompt_template,omitempty"`
	GlossaryTermDefinitionSuggestionsPromptTemplate         string `json:"glossary_term_definition_suggestions_prompt_template,omitempty"`
	CondensableCondensePromptTemplate                       string `json:"condensable_condense_prompt_template,omitempty"`
	AnswerAnswerPromptTemplate                              string `json:"answer_answer_prompt_template,omitempty"`
	ActionItemSummaryPromptTemplate                         string `json:"action_item_summary_prompt_template,omitempty"`
}

func newPromptersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List prompters",
		Long: `List prompters with filtering and pagination.

Prompters store prompt templates used across AI-assisted features such as
release notes, glossary terms, and safety risk suggestions.

Output Columns:
  ID       Prompter identifier
  NAME     Prompter name (if set)
  ACTIVE   Whether the prompter is active

Filters:
  --is-active   Filter by active status (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --base-url, --token, --no-auth`,
		Example: `  # List prompters
  xbe view prompters list

  # Filter active prompters
  xbe view prompters list --is-active true

  # Output as JSON
  xbe view prompters list --json`,
		RunE: runPromptersList,
	}
	initPromptersListFlags(cmd)
	return cmd
}

func init() {
	promptersCmd.AddCommand(newPromptersListCmd())
}

func initPromptersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("is-active", "", "Filter by active status (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPromptersList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePromptersListOptions(cmd)
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
	query.Set("sort", "name")
	query.Set("fields[prompters]", "name,is-active")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[is-active]", opts.IsActive)

	body, _, err := client.Get(cmd.Context(), "/v1/prompters", query)
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

	rows := buildPrompterRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPromptersTable(cmd, rows)
}

func parsePromptersListOptions(cmd *cobra.Command) (promptersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	isActive, _ := cmd.Flags().GetString("is-active")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return promptersListOptions{
		BaseURL:  baseURL,
		Token:    token,
		JSON:     jsonOut,
		NoAuth:   noAuth,
		Limit:    limit,
		Offset:   offset,
		IsActive: isActive,
	}, nil
}

func buildPrompterRows(resp jsonAPIResponse) []prompterRow {
	rows := make([]prompterRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, prompterRowFromResource(resource))
	}
	return rows
}

func prompterRowFromResource(resource jsonAPIResource) prompterRow {
	attrs := resource.Attributes
	return prompterRow{
		ID:       resource.ID,
		Name:     stringAttr(attrs, "name"),
		IsActive: boolAttr(attrs, "is-active"),
		ReleaseNoteGuessHasNavigationInstructionsPromptTemplate: stringAttr(attrs, "release-note-guess-has-navigation-instructions-prompt-template"),
		ReleaseNoteHeadlineSuggestionsPromptTemplate:            stringAttr(attrs, "release-note-headline-suggestions-prompt-template"),
		ReleaseNoteGlossaryTermSuggestionsPromptTemplate:        stringAttr(attrs, "release-note-glossary-term-suggestions-prompt-template"),
		JPPSafetyRisksSuggestionSuggestionPromptTemplate:        stringAttr(attrs, "jpp-safety-risks-suggestion-suggestion-prompt-template"),
		JPPSafetyRiskCommSuggestionSuggestionPromptTemplate:     stringAttr(attrs, "jpp-safety-risk-comm-suggestion-suggestion-prompt-template"),
		IncidentHeadlineSuggestionSuggestionPromptTemplate:      stringAttr(attrs, "incident-headline-suggestion-suggestion-prompt-template"),
		GlossaryTermDefinitionSuggestionsPromptTemplate:         stringAttr(attrs, "glossary-term-definition-suggestions-prompt-template"),
		CondensableCondensePromptTemplate:                       stringAttr(attrs, "condensable-condense-prompt-template"),
		AnswerAnswerPromptTemplate:                              stringAttr(attrs, "answer-answer-prompt-template"),
		ActionItemSummaryPromptTemplate:                         stringAttr(attrs, "action-item-summary-prompt-template"),
	}
}

func buildPrompterRowFromSingle(resp jsonAPISingleResponse) prompterRow {
	return prompterRowFromResource(resp.Data)
}

func renderPromptersTable(cmd *cobra.Command, rows []prompterRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No prompters found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNAME\tACTIVE")
	for _, row := range rows {
		active := "no"
		if row.IsActive {
			active = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			truncateString(row.Name, 40),
			active,
		)
	}
	return writer.Flush()
}
