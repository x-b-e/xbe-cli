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

type releaseNotesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type releaseNoteDetails struct {
	ID          string   `json:"id"`
	Headline    string   `json:"headline"`
	Released    string   `json:"released"`
	IsPublished bool     `json:"is_published"`
	IsArchived  bool     `json:"is_archived"`
	Scopes      []string `json:"scopes"`
	AudioURL    string   `json:"audio_url,omitempty"`
	Description string   `json:"description"`
}

func newReleaseNotesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show release note details",
		Long: `Show the full details of a specific release note.

Retrieves and displays comprehensive information about a release note including
its full description, release date, and metadata.

Output Fields (table format):
  ID           Unique release note identifier
  Headline     Release note headline
  Released     Release date
  Status       Publication and archive status
  Scopes       Access scopes
  Audio URL    Link to audio version (if available)
  Description  Full release note content

Arguments:
  <id>          The release note ID (required). You can find IDs using the list command.`,
		Example: `  # View a release note by ID
  xbe view release-notes show 123

  # Get release note as JSON
  xbe view release-notes show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runReleaseNotesShow,
	}
	initReleaseNotesShowFlags(cmd)
	return cmd
}

func init() {
	releaseNotesCmd.AddCommand(newReleaseNotesShowCmd())
}

func initReleaseNotesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runReleaseNotesShow(cmd *cobra.Command, args []string) error {
	opts, err := parseReleaseNotesShowOptions(cmd)
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
		return fmt.Errorf("release note id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[release-notes]", "headline,description,released-on,is-published,is-archived,scopes,audio-url")

	body, _, err := client.Get(cmd.Context(), "/v1/release-notes/"+id, query)
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

	details := buildReleaseNoteDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderReleaseNoteDetails(cmd, details)
}

func parseReleaseNotesShowOptions(cmd *cobra.Command) (releaseNotesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return releaseNotesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return releaseNotesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return releaseNotesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return releaseNotesShowOptions{}, err
	}

	return releaseNotesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildReleaseNoteDetails(resp jsonAPISingleResponse) releaseNoteDetails {
	attrs := resp.Data.Attributes

	return releaseNoteDetails{
		ID:          resp.Data.ID,
		Headline:    strings.TrimSpace(stringAttr(attrs, "headline")),
		Released:    formatDate(stringAttr(attrs, "released-on")),
		IsPublished: boolAttr(attrs, "is-published"),
		IsArchived:  boolAttr(attrs, "is-archived"),
		Scopes:      stringSliceAttr(attrs, "scopes"),
		AudioURL:    strings.TrimSpace(stringAttr(attrs, "audio-url")),
		Description: strings.TrimSpace(stringAttr(attrs, "description")),
	}
}

func renderReleaseNoteDetails(cmd *cobra.Command, details releaseNoteDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Headline != "" {
		fmt.Fprintf(out, "Headline: %s\n", details.Headline)
	}
	if details.Released != "" {
		fmt.Fprintf(out, "Released: %s\n", details.Released)
	}
	statusParts := make([]string, 0, 2)
	if details.IsPublished {
		statusParts = append(statusParts, "published")
	}
	if details.IsArchived {
		statusParts = append(statusParts, "archived")
	}
	if len(statusParts) > 0 {
		fmt.Fprintf(out, "Status: %s\n", strings.Join(statusParts, ", "))
	}
	if len(details.Scopes) > 0 {
		fmt.Fprintf(out, "Scopes: %s\n", strings.Join(details.Scopes, ", "))
	}
	if details.AudioURL != "" {
		fmt.Fprintf(out, "Audio URL: %s\n", details.AudioURL)
	}
	if details.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Description)
	}

	return nil
}
