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

type nativeAppReleasesListOptions struct {
	BaseURL        string
	Token          string
	JSON           bool
	NoAuth         bool
	Limit          int
	Offset         int
	Sort           string
	GitTag         string
	GitSHA         string
	BuildNumber    string
	ReleaseStatus  string
	ReleaseChannel string
}

type nativeAppReleaseRow struct {
	ID          string `json:"id"`
	GitTag      string `json:"git_tag,omitempty"`
	GitSHA      string `json:"git_sha,omitempty"`
	BuildNumber string `json:"build_number,omitempty"`
}

func newNativeAppReleasesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List native app releases",
		Long: `List native app releases with filtering and pagination.

Output Columns:
  ID       Native app release identifier
  GIT TAG  Git tag (if set)
  GIT SHA  Git commit SHA
  BUILD    Build number

Filters:
  --git-tag          Filter by git tag
  --git-sha          Filter by git SHA
  --build-number     Filter by build number
  --release-status   Filter by release status (e.g. uploaded, released)
  --release-channel  Filter by release channel (apple-app-store, google-play-store)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List native app releases
  xbe view native-app-releases list

  # Filter by git SHA
  xbe view native-app-releases list --git-sha abc123

  # Filter by release channel and status
  xbe view native-app-releases list --release-channel apple-app-store --release-status released

  # Paginate results
  xbe view native-app-releases list --limit 20 --offset 40

  # Output as JSON
  xbe view native-app-releases list --json`,
		Args: cobra.NoArgs,
		RunE: runNativeAppReleasesList,
	}
	initNativeAppReleasesListFlags(cmd)
	return cmd
}

func init() {
	nativeAppReleasesCmd.AddCommand(newNativeAppReleasesListCmd())
}

func initNativeAppReleasesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("git-tag", "", "Filter by git tag")
	cmd.Flags().String("git-sha", "", "Filter by git SHA")
	cmd.Flags().String("build-number", "", "Filter by build number")
	cmd.Flags().String("release-status", "", "Filter by release status")
	cmd.Flags().String("release-channel", "", "Filter by release channel")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNativeAppReleasesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseNativeAppReleasesListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[native-app-releases]", "git-tag,git-sha,build-number")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[git-tag]", opts.GitTag)
	setFilterIfPresent(query, "filter[git-sha]", opts.GitSHA)
	setFilterIfPresent(query, "filter[build-number]", opts.BuildNumber)
	setFilterIfPresent(query, "filter[release-status]", opts.ReleaseStatus)
	setFilterIfPresent(query, "filter[release-channel]", opts.ReleaseChannel)

	body, _, err := client.Get(cmd.Context(), "/v1/native-app-releases", query)
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

	rows := buildNativeAppReleaseRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderNativeAppReleasesTable(cmd, rows)
}

func parseNativeAppReleasesListOptions(cmd *cobra.Command) (nativeAppReleasesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	gitTag, _ := cmd.Flags().GetString("git-tag")
	gitSHA, _ := cmd.Flags().GetString("git-sha")
	buildNumber, _ := cmd.Flags().GetString("build-number")
	releaseStatus, _ := cmd.Flags().GetString("release-status")
	releaseChannel, _ := cmd.Flags().GetString("release-channel")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return nativeAppReleasesListOptions{
		BaseURL:        baseURL,
		Token:          token,
		JSON:           jsonOut,
		NoAuth:         noAuth,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
		GitTag:         gitTag,
		GitSHA:         gitSHA,
		BuildNumber:    buildNumber,
		ReleaseStatus:  releaseStatus,
		ReleaseChannel: releaseChannel,
	}, nil
}

func buildNativeAppReleaseRows(resp jsonAPIResponse) []nativeAppReleaseRow {
	rows := make([]nativeAppReleaseRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, nativeAppReleaseRowFromResource(resource))
	}
	return rows
}

func nativeAppReleaseRowFromResource(resource jsonAPIResource) nativeAppReleaseRow {
	attrs := resource.Attributes
	return nativeAppReleaseRow{
		ID:          resource.ID,
		GitTag:      strings.TrimSpace(stringAttr(attrs, "git-tag")),
		GitSHA:      strings.TrimSpace(stringAttr(attrs, "git-sha")),
		BuildNumber: strings.TrimSpace(stringAttr(attrs, "build-number")),
	}
}

func nativeAppReleaseRowFromSingle(resp jsonAPISingleResponse) nativeAppReleaseRow {
	return nativeAppReleaseRowFromResource(resp.Data)
}

func renderNativeAppReleasesTable(cmd *cobra.Command, rows []nativeAppReleaseRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No native app releases found.")
		return nil
	}

	const shaMax = 12

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tGIT TAG\tGIT SHA\tBUILD")

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", row.ID, row.GitTag, truncateString(row.GitSHA, shaMax), row.BuildNumber)
	}

	return writer.Flush()
}
