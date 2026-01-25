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

type releaseNotesListOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	NoAuth                         bool
	Limit                          int
	Offset                         int
	IsPublished                    string
	IsArchived                     string
	Q                              string
	ReleasedOnMin                  string
	ReleasedOnMax                  string
	CreatedBy                      string
	HasNavigationInstructionsGuess string
}

func newReleaseNotesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List release notes",
		Long: `List release notes with filtering and pagination.

Returns a list of release notes matching the specified criteria, sorted by
release date (newest first).

Output Columns (table format):
  ID         Unique release note identifier
  RELEASED   Release date
  HEADLINE   Release note headline

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.`,
		Example: `  # List recent release notes
  xbe view release-notes list

  # Search release notes
  xbe view release-notes list --q "trucking"

  # Filter by published status
  xbe view release-notes list --is-published true

  # Filter by date range
  xbe view release-notes list --released-on-min 2024-01-01 --released-on-max 2024-06-30

  # Include archived release notes
  xbe view release-notes list --is-archived true

  # Paginate results
  xbe view release-notes list --limit 20 --offset 40

  # Output as JSON for scripting
  xbe view release-notes list --json`,
		RunE: runReleaseNotesList,
	}
	initReleaseNotesListFlags(cmd)
	return cmd
}

func init() {
	releaseNotesCmd.AddCommand(newReleaseNotesListCmd())
}

func initReleaseNotesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("is-published", "", "Filter by published status (true/false)")
	cmd.Flags().String("is-archived", "", "Filter by archived status (true/false, default: false)")
	cmd.Flags().String("q", "", "Search release notes")
	cmd.Flags().String("released-on-min", "", "Filter to release notes released on or after this date (YYYY-MM-DD)")
	cmd.Flags().String("released-on-max", "", "Filter to release notes released on or before this date (YYYY-MM-DD)")
	cmd.Flags().String("created-by", "", "Filter by creator user ID (comma-separated for multiple)")
	cmd.Flags().String("has-navigation-instructions-guess", "", "Filter by having navigation instructions guess (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runReleaseNotesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseReleaseNotesListOptions(cmd)
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

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("sort", "-released-on")
	query.Set("fields[release-notes]", "headline,released-on,is-published,is-archived")
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[is-published]", opts.IsPublished)
	setFilterIfPresent(query, "filter[is-archived]", opts.IsArchived)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[released-on-min]", opts.ReleasedOnMin)
	setFilterIfPresent(query, "filter[released-on-max]", opts.ReleasedOnMax)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[has-navigation-instructions-guess]", opts.HasNavigationInstructionsGuess)

	body, _, err := client.Get(cmd.Context(), "/v1/release-notes", query)
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

	if opts.JSON {
		rows := buildReleaseNoteRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderReleaseNotesTable(cmd, resp)
}

func parseReleaseNotesListOptions(cmd *cobra.Command) (releaseNotesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	isPublished, err := cmd.Flags().GetString("is-published")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	isArchived, err := cmd.Flags().GetString("is-archived")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	q, err := cmd.Flags().GetString("q")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	releasedOnMin, err := cmd.Flags().GetString("released-on-min")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	releasedOnMax, err := cmd.Flags().GetString("released-on-max")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	createdBy, err := cmd.Flags().GetString("created-by")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	hasNavigationInstructionsGuess, err := cmd.Flags().GetString("has-navigation-instructions-guess")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return releaseNotesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return releaseNotesListOptions{}, err
	}

	return releaseNotesListOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		NoAuth:                         noAuth,
		Limit:                          limit,
		Offset:                         offset,
		IsPublished:                    isPublished,
		IsArchived:                     isArchived,
		Q:                              q,
		ReleasedOnMin:                  releasedOnMin,
		ReleasedOnMax:                  releasedOnMax,
		CreatedBy:                      createdBy,
		HasNavigationInstructionsGuess: hasNavigationInstructionsGuess,
	}, nil
}

func renderReleaseNotesTable(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildReleaseNoteRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No release notes found.")
		return nil
	}

	const headlineMax = 80

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tRELEASED\tHEADLINE")

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", row.ID, row.Released, truncateString(row.Headline, headlineMax))
	}

	return writer.Flush()
}

type releaseNoteRow struct {
	ID          string `json:"id"`
	Released    string `json:"released"`
	Headline    string `json:"headline"`
	IsPublished bool   `json:"is_published"`
	IsArchived  bool   `json:"is_archived"`
}

func buildReleaseNoteRows(resp jsonAPIResponse) []releaseNoteRow {
	rows := make([]releaseNoteRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, releaseNoteRow{
			ID:          resource.ID,
			Released:    formatDate(stringAttr(resource.Attributes, "released-on")),
			Headline:    strings.TrimSpace(stringAttr(resource.Attributes, "headline")),
			IsPublished: boolAttr(resource.Attributes, "is-published"),
			IsArchived:  boolAttr(resource.Attributes, "is-archived"),
		})
	}

	return rows
}
