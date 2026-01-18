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

type pressReleasesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type pressReleaseDetails struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Headline    string `json:"headline"`
	Subheadline string `json:"subheadline,omitempty"`
	Released    string `json:"released"`
	Location    string `json:"location,omitempty"`
	Published   bool   `json:"published"`
	AudioURL    string `json:"audio_url,omitempty"`
	Body        string `json:"body"`
}

func newPressReleasesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show press release details",
		Long: `Show the full details of a specific press release.

Retrieves and displays comprehensive information about a press release including
its full body content, metadata, and publication status.

Output Fields (table format):
  ID           Unique press release identifier
  Slug         URL-friendly identifier
  Headline     Press release headline
  Subheadline  Secondary headline (if available)
  Released     Release date and time
  Location     Release location
  Published    Publication status
  Audio URL    Link to audio version (if available)
  Body         Full press release content

Arguments:
  <id>          The press release ID (required). You can find IDs using the list command.`,
		Example: `  # View a press release by ID
  xbe view press-releases show 123

  # Get press release as JSON
  xbe view press-releases show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runPressReleasesShow,
	}
	initPressReleasesShowFlags(cmd)
	return cmd
}

func init() {
	pressReleasesCmd.AddCommand(newPressReleasesShowCmd())
}

func initPressReleasesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPressReleasesShow(cmd *cobra.Command, args []string) error {
	opts, err := parsePressReleasesShowOptions(cmd)
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
		return fmt.Errorf("press release id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[press-releases]", "slug,headline,subheadline,body,released-at,location-name,published,audio-url")

	body, _, err := client.Get(cmd.Context(), "/v1/press-releases/"+id, query)
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

	details := buildPressReleaseDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderPressReleaseDetails(cmd, details)
}

func parsePressReleasesShowOptions(cmd *cobra.Command) (pressReleasesShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return pressReleasesShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return pressReleasesShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return pressReleasesShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return pressReleasesShowOptions{}, err
	}

	return pressReleasesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildPressReleaseDetails(resp jsonAPISingleResponse) pressReleaseDetails {
	attrs := resp.Data.Attributes

	return pressReleaseDetails{
		ID:          resp.Data.ID,
		Slug:        stringAttr(attrs, "slug"),
		Headline:    strings.TrimSpace(stringAttr(attrs, "headline")),
		Subheadline: strings.TrimSpace(stringAttr(attrs, "subheadline")),
		Released:    formatDate(stringAttr(attrs, "released-at")),
		Location:    strings.TrimSpace(stringAttr(attrs, "location-name")),
		Published:   boolAttr(attrs, "published"),
		AudioURL:    strings.TrimSpace(stringAttr(attrs, "audio-url")),
		Body:        strings.TrimSpace(stringAttr(attrs, "body")),
	}
}

func renderPressReleaseDetails(cmd *cobra.Command, details pressReleaseDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Slug != "" {
		fmt.Fprintf(out, "Slug: %s\n", details.Slug)
	}
	if details.Headline != "" {
		fmt.Fprintf(out, "Headline: %s\n", details.Headline)
	}
	if details.Subheadline != "" {
		fmt.Fprintf(out, "Subheadline: %s\n", details.Subheadline)
	}
	if details.Released != "" {
		fmt.Fprintf(out, "Released: %s\n", details.Released)
	}
	if details.Location != "" {
		fmt.Fprintf(out, "Location: %s\n", details.Location)
	}
	if details.Published {
		fmt.Fprintf(out, "Status: published\n")
	}
	if details.AudioURL != "" {
		fmt.Fprintf(out, "Audio URL: %s\n", details.AudioURL)
	}
	if details.Body != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Body:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, details.Body)
	}

	return nil
}
