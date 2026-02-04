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

type incidentHeadlineSuggestionsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type incidentHeadlineSuggestionDetails struct {
	ID          string `json:"id"`
	IncidentID  string `json:"incident_id,omitempty"`
	IsAsync     bool   `json:"is_async"`
	Options     any    `json:"options,omitempty"`
	Prompt      string `json:"prompt,omitempty"`
	Response    any    `json:"response,omitempty"`
	IsFulfilled bool   `json:"is_fulfilled"`
	Suggestion  string `json:"suggestion,omitempty"`
}

func newIncidentHeadlineSuggestionsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show incident headline suggestion details",
		Long: `Show the full details of an incident headline suggestion.

Output Fields:
  ID
  Incident ID
  Is Async
  Options
  Prompt
  Response
  Is Fulfilled
  Suggestion

Arguments:
  <id>  The incident headline suggestion ID (required).`,
		Example: `  # Show a suggestion
  xbe view incident-headline-suggestions show 123

  # Output as JSON
  xbe view incident-headline-suggestions show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runIncidentHeadlineSuggestionsShow,
	}
	initIncidentHeadlineSuggestionsShowFlags(cmd)
	return cmd
}

func init() {
	incidentHeadlineSuggestionsCmd.AddCommand(newIncidentHeadlineSuggestionsShowCmd())
}

func initIncidentHeadlineSuggestionsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentHeadlineSuggestionsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseIncidentHeadlineSuggestionsShowOptions(cmd)
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
		return fmt.Errorf("incident headline suggestion id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-headline-suggestions]", "incident,is-async,options,prompt,response,is-fulfilled,suggestion")

	body, _, err := client.Get(cmd.Context(), "/v1/incident-headline-suggestions/"+id, query)
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

	details := buildIncidentHeadlineSuggestionDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderIncidentHeadlineSuggestionDetails(cmd, details)
}

func parseIncidentHeadlineSuggestionsShowOptions(cmd *cobra.Command) (incidentHeadlineSuggestionsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentHeadlineSuggestionsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildIncidentHeadlineSuggestionDetails(resp jsonAPISingleResponse) incidentHeadlineSuggestionDetails {
	resource := resp.Data
	details := incidentHeadlineSuggestionDetails{
		ID:          resource.ID,
		IsAsync:     boolAttr(resource.Attributes, "is-async"),
		Options:     resource.Attributes["options"],
		Prompt:      stringAttr(resource.Attributes, "prompt"),
		Response:    resource.Attributes["response"],
		IsFulfilled: boolAttr(resource.Attributes, "is-fulfilled"),
		Suggestion:  stringAttr(resource.Attributes, "suggestion"),
	}

	if rel, ok := resource.Relationships["incident"]; ok && rel.Data != nil {
		details.IncidentID = rel.Data.ID
	}

	return details
}

func renderIncidentHeadlineSuggestionDetails(cmd *cobra.Command, details incidentHeadlineSuggestionDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.IncidentID != "" {
		fmt.Fprintf(out, "Incident ID: %s\n", details.IncidentID)
	}
	fmt.Fprintf(out, "Is Async: %t\n", details.IsAsync)
	fmt.Fprintf(out, "Is Fulfilled: %t\n", details.IsFulfilled)

	if details.Prompt != "" {
		fmt.Fprintf(out, "Prompt: %s\n", details.Prompt)
	}
	if details.Suggestion != "" {
		fmt.Fprintf(out, "Suggestion: %s\n", details.Suggestion)
	}

	if formatted := formatSuggestionPayload(details.Options); formatted != "" {
		fmt.Fprintln(out, "Options:")
		fmt.Fprintln(out, formatted)
	}
	if formatted := formatSuggestionPayload(details.Response); formatted != "" {
		fmt.Fprintln(out, "Response:")
		fmt.Fprintln(out, formatted)
	}

	return nil
}

func formatSuggestionPayload(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		pretty, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(pretty)
	}
}
