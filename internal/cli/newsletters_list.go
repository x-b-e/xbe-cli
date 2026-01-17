package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
)

type newslettersListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
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

type jsonAPIResponse struct {
	Data     []jsonAPIResource `json:"data"`
	Included []jsonAPIResource `json:"included"`
}

type jsonAPIResource struct {
	ID            string                         `json:"id"`
	Type          string                         `json:"type"`
	Attributes    map[string]any                 `json:"attributes"`
	Relationships map[string]jsonAPIRelationship `json:"relationships"`
}

type jsonAPIRelationship struct {
	Data *jsonAPIResourceIdentifier `json:"data"`
}

type jsonAPIResourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

var newslettersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List newsletters",
	Long:  "List published newsletters up to today.",
	RunE:  runNewslettersList,
}

func init() {
	newslettersCmd.AddCommand(newslettersListCmd)

	newslettersListCmd.Flags().Bool("json", false, "Output JSON")
	newslettersListCmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	newslettersListCmd.Flags().Int("offset", 0, "Page offset")
	newslettersListCmd.Flags().String("is-published", "true", "Filter by published status (true/false)")
	newslettersListCmd.Flags().String("is-public", "", "Filter by public status (true/false)")
	newslettersListCmd.Flags().String("q", "", "Search newsletters")
	newslettersListCmd.Flags().String("has-organization", "", "Filter by presence of organization (true/false)")
	newslettersListCmd.Flags().String("organization", "", "Filter by organization (e.g., Broker|123)")
	newslettersListCmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker)")
	newslettersListCmd.Flags().Int("broker-id", 0, "Filter by broker organization id")
	newslettersListCmd.Flags().String("published-on", "", "Filter to newsletters published on this date (YYYY-MM-DD)")
	newslettersListCmd.Flags().String("published-on-min", "", "Filter to newsletters published on or after this date (YYYY-MM-DD)")
	newslettersListCmd.Flags().String("published-on-max", "", "Filter to newsletters published on or before this date (YYYY-MM-DD)")
	newslettersListCmd.Flags().String("has-published-on", "", "Filter by presence of published-on date (true/false)")
	newslettersListCmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	newslettersListCmd.Flags().String("token", defaultToken(), "API token (optional)")
}

func runNewslettersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseNewslettersListOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
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

func setFilterIfPresent(query url.Values, key, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		query.Set(key, value)
	}
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

func resolveOrganization(resource jsonAPIResource, included map[string]map[string]any) string {
	rel, ok := resource.Relationships["organization"]
	if !ok || rel.Data == nil {
		return "XBE Horizon"
	}

	key := resourceKey(rel.Data.Type, rel.Data.ID)
	if attrs, ok := included[key]; ok {
		name := firstNonEmpty(
			stringAttr(attrs, "company-name"),
			stringAttr(attrs, "name"),
			stringAttr(attrs, "title"),
		)
		if name != "" {
			return name
		}
	}

	return fmt.Sprintf("%s:%s", rel.Data.Type, rel.Data.ID)
}

func resourceKey(typ, id string) string {
	return typ + "|" + id
}

func stringAttr(attrs map[string]any, key string) string {
	if attrs == nil {
		return ""
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func truncateString(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	if max < 4 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func formatDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", value); err == nil {
		return value
	}
	return value
}

func writeJSON(out io.Writer, value any) error {
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if _, err := out.Write(pretty); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out)
	return err
}

func defaultBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("XBE_BASE_URL")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("XBE_API_BASE_URL")); value != "" {
		return value
	}
	return "https://server.x-b-e.com"
}

func defaultToken() string {
	if value := strings.TrimSpace(os.Getenv("XBE_TOKEN")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("XBE_API_TOKEN")); value != "" {
		return value
	}
	return ""
}
