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

type incidentTagIncidentsListOptions struct {
	BaseURL     string
	Token       string
	JSON        bool
	NoAuth      bool
	Limit       int
	Offset      int
	Sort        string
	Incident    string
	IncidentTag string
}

type incidentTagIncidentRow struct {
	ID               string `json:"id"`
	IncidentID       string `json:"incident_id,omitempty"`
	IncidentHeadline string `json:"incident_headline,omitempty"`
	IncidentStatus   string `json:"incident_status,omitempty"`
	IncidentTagID    string `json:"incident_tag_id,omitempty"`
	IncidentTagSlug  string `json:"incident_tag_slug,omitempty"`
	IncidentTagName  string `json:"incident_tag_name,omitempty"`
}

func newIncidentTagIncidentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List incident tag incident links",
		Long: `List incident tag incident links with filtering and pagination.

Output Columns:
  ID        Incident tag incident link identifier
  INCIDENT  Incident headline or ID
  TAG       Incident tag name or slug
  STATUS    Incident status

Filters:
  --incident      Filter by incident ID
  --incident-tag  Filter by incident tag ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List incident tag incidents
  xbe view incident-tag-incidents list

  # Filter by incident
  xbe view incident-tag-incidents list --incident 123

  # Filter by incident tag
  xbe view incident-tag-incidents list --incident-tag 456

  # Output as JSON
  xbe view incident-tag-incidents list --json`,
		Args: cobra.NoArgs,
		RunE: runIncidentTagIncidentsList,
	}
	initIncidentTagIncidentsListFlags(cmd)
	return cmd
}

func init() {
	incidentTagIncidentsCmd.AddCommand(newIncidentTagIncidentsListCmd())
}

func initIncidentTagIncidentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("incident", "", "Filter by incident ID")
	cmd.Flags().String("incident-tag", "", "Filter by incident tag ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runIncidentTagIncidentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseIncidentTagIncidentsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run 'xbe auth login' first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[incident-tag-incidents]", "incident,incident-tag")
	query.Set("include", "incident,incident-tag")
	query.Set("fields[incidents]", "headline,status")
	query.Set("fields[incident-tags]", "slug,name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[incident]", opts.Incident)
	setFilterIfPresent(query, "filter[incident-tag]", opts.IncidentTag)

	body, _, err := client.Get(cmd.Context(), "/v1/incident-tag-incidents", query)
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

	rows := buildIncidentTagIncidentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderIncidentTagIncidentsTable(cmd, rows)
}

func parseIncidentTagIncidentsListOptions(cmd *cobra.Command) (incidentTagIncidentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	incident, _ := cmd.Flags().GetString("incident")
	incidentTag, _ := cmd.Flags().GetString("incident-tag")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return incidentTagIncidentsListOptions{
		BaseURL:     baseURL,
		Token:       token,
		JSON:        jsonOut,
		NoAuth:      noAuth,
		Limit:       limit,
		Offset:      offset,
		Sort:        sort,
		Incident:    incident,
		IncidentTag: incidentTag,
	}, nil
}

func buildIncidentTagIncidentRows(resp jsonAPIResponse) []incidentTagIncidentRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]incidentTagIncidentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildIncidentTagIncidentRow(resource, included))
	}
	return rows
}

func incidentTagIncidentRowFromSingle(resp jsonAPISingleResponse) incidentTagIncidentRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildIncidentTagIncidentRow(resp.Data, included)
}

func buildIncidentTagIncidentRow(resource jsonAPIResource, included map[string]jsonAPIResource) incidentTagIncidentRow {
	row := incidentTagIncidentRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["incident"]; ok && rel.Data != nil {
		row.IncidentID = rel.Data.ID
		if incident, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.IncidentHeadline = stringAttr(incident.Attributes, "headline")
			row.IncidentStatus = stringAttr(incident.Attributes, "status")
		}
	}

	if rel, ok := resource.Relationships["incident-tag"]; ok && rel.Data != nil {
		row.IncidentTagID = rel.Data.ID
		if tag, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.IncidentTagSlug = stringAttr(tag.Attributes, "slug")
			row.IncidentTagName = stringAttr(tag.Attributes, "name")
		}
	}

	return row
}

func renderIncidentTagIncidentsTable(cmd *cobra.Command, rows []incidentTagIncidentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No incident tag incidents found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tINCIDENT\tTAG\tSTATUS")
	for _, row := range rows {
		incident := firstNonEmpty(row.IncidentHeadline, row.IncidentID)
		tag := firstNonEmpty(row.IncidentTagName, row.IncidentTagSlug, row.IncidentTagID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(incident, 32),
			truncateString(tag, 24),
			truncateString(row.IncidentStatus, 12),
		)
	}
	return writer.Flush()
}
