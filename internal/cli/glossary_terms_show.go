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

type glossaryTermsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type glossaryTermDetails struct {
	ID         string `json:"id"`
	Term       string `json:"term"`
	Definition string `json:"definition"`
	Source     string `json:"source"`
	AudioURL   string `json:"audio_url,omitempty"`
}

func newGlossaryTermsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show glossary term details",
		Long: `Show the full details of a specific glossary term.

Retrieves and displays comprehensive information about a glossary term including
its full definition and source.

Output Fields (table format):
  ID          Unique glossary term identifier
  Term        The term being defined
  Definition  Full definition of the term
  Source      Source of the definition
  Audio URL   Link to audio version (if available)

Arguments:
  <id>          The glossary term ID (required). You can find IDs using the list command.`,
		Example: `  # View a glossary term by ID
  xbe view glossary-terms show 123

  # Get glossary term as JSON
  xbe view glossary-terms show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runGlossaryTermsShow,
	}
	initGlossaryTermsShowFlags(cmd)
	return cmd
}

func init() {
	glossaryTermsCmd.AddCommand(newGlossaryTermsShowCmd())
}

func initGlossaryTermsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runGlossaryTermsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseGlossaryTermsShowOptions(cmd)
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
		return fmt.Errorf("glossary term id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[glossary-terms]", "term,definition,source,audio-url")

	body, _, err := client.Get(cmd.Context(), "/v1/glossary-terms/"+id, query)
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

	details := buildGlossaryTermDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderGlossaryTermDetails(cmd, details)
}

func parseGlossaryTermsShowOptions(cmd *cobra.Command) (glossaryTermsShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return glossaryTermsShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return glossaryTermsShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return glossaryTermsShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return glossaryTermsShowOptions{}, err
	}

	return glossaryTermsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildGlossaryTermDetails(resp jsonAPISingleResponse) glossaryTermDetails {
	attrs := resp.Data.Attributes

	return glossaryTermDetails{
		ID:         resp.Data.ID,
		Term:       strings.TrimSpace(stringAttr(attrs, "term")),
		Definition: strings.TrimSpace(stringAttr(attrs, "definition")),
		Source:     stringAttr(attrs, "source"),
		AudioURL:   strings.TrimSpace(stringAttr(attrs, "audio-url")),
	}
}

func renderGlossaryTermDetails(cmd *cobra.Command, details glossaryTermDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Term != "" {
		fmt.Fprintf(out, "Term: %s\n", details.Term)
	}
	if details.Source != "" {
		fmt.Fprintf(out, "Source: %s\n", details.Source)
	}
	if details.AudioURL != "" {
		fmt.Fprintf(out, "Audio URL: %s\n", details.AudioURL)
	}
	if details.Definition != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Definition:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Definition)
	}

	return nil
}
