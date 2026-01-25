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

type newslettersListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	IsPublished      string
	IsPublic         string
	Q                string
	HasOrganization  string
	Organization     string
	OrganizationType string
	BrokerID         int
	PublishedOn      string
	PublishedOnMin   string
	PublishedOnMax   string
	HasPublishedOn   string
}

func newNewslettersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List newsletters",
		Long: `List published newsletters with filtering and pagination.

Returns a list of newsletters matching the specified criteria. By default,
only published newsletters are shown, sorted by publication date (newest first).

Output Columns (table format):
  ID            Unique newsletter identifier
  PUBLISHED     Publication date
  SUMMARY       Brief summary of the newsletter content
  ORGANIZATION  The broker/organization that published the newsletter

Pagination:
  Use --limit and --offset to paginate through large result sets.
  The server has a default page size if --limit is not specified.

Filtering:
  Multiple filters can be combined. All filters use AND logic.`,
		Example: `  # List recent newsletters (default: published only)
  xbe view newsletters list

  # Search newsletters by keyword
  xbe view newsletters list --q "interest rates"

  # Filter by broker ID
  xbe view newsletters list --broker-id 123

  # Filter by date range
  xbe view newsletters list --published-on-min 2024-01-01 --published-on-max 2024-06-30

  # Filter by exact publication date
  xbe view newsletters list --published-on 2024-03-15

  # Show only public newsletters
  xbe view newsletters list --is-public true

  # Include unpublished newsletters (requires auth)
  xbe view newsletters list --is-published ""

  # Paginate results
  xbe view newsletters list --limit 20 --offset 40

  # Output as JSON for scripting
  xbe view newsletters list --json

  # Access without authentication (public content only)
  xbe view newsletters list --no-auth --is-public true`,
		RunE: runNewslettersList,
	}
	initNewslettersListFlags(cmd)
	return cmd
}

func init() {
	newslettersCmd.AddCommand(newNewslettersListCmd())
}

func initNewslettersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("is-published", "true", "Filter by published status (true/false)")
	cmd.Flags().String("is-public", "", "Filter by public status (true/false)")
	cmd.Flags().String("q", "", "Search newsletters")
	cmd.Flags().String("has-organization", "", "Filter by presence of organization (true/false)")
	cmd.Flags().String("organization", "", "Filter by organization (e.g., Broker|123)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker)")
	cmd.Flags().Int("broker-id", 0, "Filter by broker organization id")
	cmd.Flags().String("published-on", "", "Filter to newsletters published on this date (YYYY-MM-DD)")
	cmd.Flags().String("published-on-min", "", "Filter to newsletters published on or after this date (YYYY-MM-DD)")
	cmd.Flags().String("published-on-max", "", "Filter to newsletters published on or before this date (YYYY-MM-DD)")
	cmd.Flags().String("has-published-on", "", "Filter by presence of published-on date (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runNewslettersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseNewslettersListOptions(cmd)
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
	query.Set("sort", "-published-on")
	query.Set("include", "organization")
	query.Set("fields[newsletters]", "summary,published-on,organization")
	setFilterIfPresent(query, "filter[is-published]", opts.IsPublished)
	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	setFilterIfPresent(query, "filter[is-public]", opts.IsPublic)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[has-organization]", opts.HasOrganization)
	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	setFilterIfPresent(query, "filter[organization-type]", opts.OrganizationType)
	setFilterIfPresent(query, "filter[published-on]", opts.PublishedOn)
	setFilterIfPresent(query, "filter[published-on-min]", opts.PublishedOnMin)
	setFilterIfPresent(query, "filter[published-on-max]", opts.PublishedOnMax)
	setFilterIfPresent(query, "filter[has-published-on]", opts.HasPublishedOn)
	if opts.BrokerID > 0 {
		query.Set("filter[broker-id]", strconv.Itoa(opts.BrokerID))
	}

	body, _, err := client.Get(cmd.Context(), "/v1/newsletters", query)
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

	if opts.JSON {
		rows := buildNewsletterRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderNewslettersTable(cmd, resp)
}

func parseNewslettersListOptions(cmd *cobra.Command) (newslettersListOptions, error) {
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return newslettersListOptions{}, err
	}
	noAuth, err := cmd.Flags().GetBool("no-auth")
	if err != nil {
		return newslettersListOptions{}, err
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return newslettersListOptions{}, err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return newslettersListOptions{}, err
	}
	isPublished, err := cmd.Flags().GetString("is-published")
	if err != nil {
		return newslettersListOptions{}, err
	}
	isPublic, err := cmd.Flags().GetString("is-public")
	if err != nil {
		return newslettersListOptions{}, err
	}
	queryText, err := cmd.Flags().GetString("q")
	if err != nil {
		return newslettersListOptions{}, err
	}
	hasOrganization, err := cmd.Flags().GetString("has-organization")
	if err != nil {
		return newslettersListOptions{}, err
	}
	organization, err := cmd.Flags().GetString("organization")
	if err != nil {
		return newslettersListOptions{}, err
	}
	organizationType, err := cmd.Flags().GetString("organization-type")
	if err != nil {
		return newslettersListOptions{}, err
	}
	brokerID, err := cmd.Flags().GetInt("broker-id")
	if err != nil {
		return newslettersListOptions{}, err
	}
	publishedOn, err := cmd.Flags().GetString("published-on")
	if err != nil {
		return newslettersListOptions{}, err
	}
	publishedOnMin, err := cmd.Flags().GetString("published-on-min")
	if err != nil {
		return newslettersListOptions{}, err
	}
	publishedOnMax, err := cmd.Flags().GetString("published-on-max")
	if err != nil {
		return newslettersListOptions{}, err
	}
	hasPublishedOn, err := cmd.Flags().GetString("has-published-on")
	if err != nil {
		return newslettersListOptions{}, err
	}
	baseURL, err := cmd.Flags().GetString("base-url")
	if err != nil {
		return newslettersListOptions{}, err
	}
	token, err := cmd.Flags().GetString("token")
	if err != nil {
		return newslettersListOptions{}, err
	}

	return newslettersListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		IsPublished:      isPublished,
		IsPublic:         isPublic,
		Q:                queryText,
		HasOrganization:  hasOrganization,
		Organization:     organization,
		OrganizationType: organizationType,
		BrokerID:         brokerID,
		PublishedOn:      publishedOn,
		PublishedOnMin:   publishedOnMin,
		PublishedOnMax:   publishedOnMax,
		HasPublishedOn:   hasPublishedOn,
	}, nil
}

func renderNewslettersTable(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildNewsletterRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No newsletters found.")
		return nil
	}

	const tableSummaryMax = 160

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tPUBLISHED\tSUMMARY\tORGANIZATION")

	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n", row.ID, row.Published, truncateString(row.Summary, tableSummaryMax), row.Organization)
	}

	return writer.Flush()
}

type newsletterRow struct {
	ID           string `json:"id"`
	Published    string `json:"published"`
	Summary      string `json:"summary"`
	Organization string `json:"organization"`
}

func buildNewsletterRows(resp jsonAPIResponse) []newsletterRow {
	included := map[string]map[string]any{}
	for _, resource := range resp.Included {
		included[resourceKey(resource.Type, resource.ID)] = resource.Attributes
	}

	rows := make([]newsletterRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, newsletterRow{
			ID:           resource.ID,
			Published:    formatDate(stringAttr(resource.Attributes, "published-on")),
			Summary:      strings.TrimSpace(stringAttr(resource.Attributes, "summary")),
			Organization: resolveOrganization(resource, included),
		})
	}

	return rows
}
