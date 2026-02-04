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

type newslettersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type newsletterDetails struct {
	ID           string   `json:"id"`
	Published    string   `json:"published"`
	Summary      string   `json:"summary"`
	Organization string   `json:"organization"`
	IsPublished  bool     `json:"is_published"`
	IsPublic     bool     `json:"is_public"`
	UserScopes   []string `json:"user_scopes"`
	AudioURL     string   `json:"audio_url"`
	Body         string   `json:"body"`
}

func newNewslettersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show newsletter details",
		Long: `Show the full details of a specific newsletter.

Retrieves and displays comprehensive information about a newsletter including
its full body content, metadata, and publication status.

Output Fields (table format):
  ID            Unique newsletter identifier
  Summary       Brief summary of the content
  Published     Publication date
  Organization  The broker/organization that published it
  Status        Publication flags (published, public)
  User Scopes   Access scopes for the authenticated user
  Audio URL     Link to audio version (if available)
  Body          Full newsletter content

Arguments:
  <id>          The newsletter ID (required). You can find IDs using the list command.`,
		Example: `  # View a newsletter by ID
  xbe view newsletters show 123

  # Get newsletter as JSON
  xbe view newsletters show 123 --json

  # View without authentication (only works for public newsletters)
  xbe view newsletters show 123 --no-auth`,
		Args: cobra.ExactArgs(1),
		RunE: runNewslettersShow,
	}
	initNewslettersShowFlags(cmd)
	return cmd
}

func init() {
	newslettersCmd.AddCommand(newNewslettersShowCmd())
}

func initNewslettersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNewslettersShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseNewslettersShowOptions(cmd)
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
		return fmt.Errorf("newsletter id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("include", "organization")
	query.Set("fields[newsletters]", "summary,published-on,is-public,is-published,body,audio-url,user-scopes,organization")

	body, _, err := client.Get(cmd.Context(), "/v1/newsletters/"+id, query)
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

	details := buildNewsletterDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderNewsletterDetails(cmd, details)
}

func parseNewslettersShowOptions(cmd *cobra.Command) (newslettersShowOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return newslettersShowOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return newslettersShowOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return newslettersShowOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return newslettersShowOptions{}, err
	}

	return newslettersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildNewsletterDetails(resp jsonAPISingleResponse) newsletterDetails {
	included := map[string]map[string]any{}
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource.Attributes
	}

	attrs := resp.Data.Attributes

	return newsletterDetails{
		ID:           resp.Data.ID,
		Published:    formatDate(stringAttr(attrs, "published-on")),
		Summary:      strings.TrimSpace(stringAttr(attrs, "summary")),
		Organization: resolveOrganization(resp.Data, included),
		IsPublished:  boolAttr(attrs, "is-published"),
		IsPublic:     boolAttr(attrs, "is-public"),
		UserScopes:   stringSliceAttr(attrs, "user-scopes"),
		AudioURL:     strings.TrimSpace(stringAttr(attrs, "audio-url")),
		Body:         strings.TrimSpace(stringAttr(attrs, "body")),
	}
}

func renderNewsletterDetails(cmd *cobra.Command, details newsletterDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Summary != "" {
		fmt.Fprintf(out, "Summary: %s\n", details.Summary)
	}
	if details.Published != "" {
		fmt.Fprintf(out, "Published: %s\n", details.Published)
	}
	if details.Organization != "" {
		fmt.Fprintf(out, "Organization: %s\n", details.Organization)
	}
	statusParts := make([]string, 0, 2)
	if details.IsPublished {
		statusParts = append(statusParts, "published")
	}
	if details.IsPublic {
		statusParts = append(statusParts, "public")
	}
	if len(statusParts) > 0 {
		fmt.Fprintf(out, "Status: %s\n", strings.Join(statusParts, ", "))
	}
	if len(details.UserScopes) > 0 {
		fmt.Fprintf(out, "User Scopes: %s\n", strings.Join(details.UserScopes, ", "))
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
