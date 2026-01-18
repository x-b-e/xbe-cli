package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type pressReleasesListOptions struct {
	BaseURL       string
	Token         string
	JSON          bool
	NoAuth        bool
	Published     string
	Slug          string
	ReleasedAtMin string
	ReleasedAtMax string
}

func newPressReleasesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List press releases",
		Long: `List press releases.

Returns all press releases matching the specified criteria, sorted by
release date (newest first).

Output Columns (table format):
  ID         Unique press release identifier
  RELEASED   Release date
  HEADLINE   Press release headline

Filtering:
  Multiple filters can be combined. All filters use AND logic.`,
		Example: `  # List all press releases
  xbe view press-releases list

  # Filter by published status
  xbe view press-releases list --published true

  # Filter by slug
  xbe view press-releases list --slug "company-announcement"

  # Filter by date range
  xbe view press-releases list --released-at-min 2024-01-01 --released-at-max 2024-06-30

  # Output as JSON for scripting
  xbe view press-releases list --json`,
		RunE: runPressReleasesList,
	}
	initPressReleasesListFlags(cmd)
	return cmd
}

func init() {
	pressReleasesCmd.AddCommand(newPressReleasesListCmd())
}

func initPressReleasesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("published", "", "Filter by published status (true/false)")
	cmd.Flags().String("slug", "", "Filter by slug")
	cmd.Flags().String("released-at-min", "", "Filter to press releases released on or after this date (YYYY-MM-DD)")
	cmd.Flags().String("released-at-max", "", "Filter to press releases released on or before this date (YYYY-MM-DD)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runPressReleasesList(cmd *cobra.Command, _ []string) error {
	opts, err := parsePressReleasesListOptions(cmd)
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
	query.Set("sort", "-released-at")
	query.Set("fields[press-releases]", "slug,headline,released-at,published")
	setFilterIfPresent(query, "filter[published]", opts.Published)
	setFilterIfPresent(query, "filter[slug]", opts.Slug)
	setFilterIfPresent(query, "filter[released-at-min]", opts.ReleasedAtMin)
	setFilterIfPresent(query, "filter[released-at-max]", opts.ReleasedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/press-releases", query)
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
		rows := buildPressReleaseRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderPressReleasesTable(cmd, resp)
}

func parsePressReleasesListOptions(cmd *cobra.Command) (pressReleasesListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return pressReleasesListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return pressReleasesListOptions{}, err
	}
	published, err := cmd.Flags().GetString("published")
	if err != nil {
		return pressReleasesListOptions{}, err
	}
	slug, err := cmd.Flags().GetString("slug")
	if err != nil {
		return pressReleasesListOptions{}, err
	}
	releasedAtMin, err := cmd.Flags().GetString("released-at-min")
	if err != nil {
		return pressReleasesListOptions{}, err
	}
	releasedAtMax, err := cmd.Flags().GetString("released-at-max")
	if err != nil {
		return pressReleasesListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return pressReleasesListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return pressReleasesListOptions{}, err
	}

	return pressReleasesListOptions{
		BaseURL:       baseURL,
		Token:         token,
		JSON:          jsonOut,
		NoAuth:        noAuth,
		Published:     published,
		Slug:          slug,
		ReleasedAtMin: releasedAtMin,
		ReleasedAtMax: releasedAtMax,
	}, nil
}

func renderPressReleasesTable(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildPressReleaseRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No press releases found.")
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

type pressReleaseRow struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Released  string `json:"released"`
	Headline  string `json:"headline"`
	Location  string `json:"location"`
	Published bool   `json:"published"`
}

func buildPressReleaseRows(resp jsonAPIResponse) []pressReleaseRow {
	rows := make([]pressReleaseRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, pressReleaseRow{
			ID:        resource.ID,
			Slug:      stringAttr(resource.Attributes, "slug"),
			Released:  formatDate(stringAttr(resource.Attributes, "released-at")),
			Headline:  strings.TrimSpace(stringAttr(resource.Attributes, "headline")),
			Location:  strings.TrimSpace(stringAttr(resource.Attributes, "location-name")),
			Published: boolAttr(resource.Attributes, "published"),
		})
	}

	return rows
}
