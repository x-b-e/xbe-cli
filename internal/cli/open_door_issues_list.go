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

type openDoorIssuesListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	Organization        string
	OrganizationType    string
	OrganizationID      string
	NotOrganizationType string
	CreatedAtMin        string
	CreatedAtMax        string
	UpdatedAtMin        string
	UpdatedAtMax        string
}

type openDoorIssueRow struct {
	ID               string `json:"id"`
	Status           string `json:"status,omitempty"`
	Description      string `json:"description,omitempty"`
	Organization     string `json:"organization,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	ReportedByID     string `json:"reported_by_id,omitempty"`
	ReportedByName   string `json:"reported_by_name,omitempty"`
}

func newOpenDoorIssuesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List open door issues",
		Long: `List open door issues with filtering and pagination.

Open door issues capture concerns reported by users for broker, customer,
or trucker organizations.

Output Columns:
  ID           Open door issue ID
  STATUS       Current status
  DESCRIPTION  Issue description (truncated)
  ORG          Organization name (falls back to type/id)
  REPORTED BY  Reporting user name (falls back to ID)

Filters:
  --organization          Filter by organization (Type|ID, e.g. Broker|123)
  --organization-type     Filter by organization type (Broker, Customer, Trucker)
  --organization-id       Filter by organization ID (use with --organization-type)
  --not-organization-type Exclude by organization type
  --created-at-min        Filter by created-at on/after (ISO 8601)
  --created-at-max        Filter by created-at on/before (ISO 8601)
  --updated-at-min        Filter by updated-at on/after (ISO 8601)
  --updated-at-max        Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List open door issues
  xbe view open-door-issues list

  # Filter by organization
  xbe view open-door-issues list --organization "Broker|123"

  # Filter by organization type
  xbe view open-door-issues list --organization-type Broker --organization-id 123

  # Output as JSON
  xbe view open-door-issues list --json`,
		Args: cobra.NoArgs,
		RunE: runOpenDoorIssuesList,
	}
	initOpenDoorIssuesListFlags(cmd)
	return cmd
}

func init() {
	openDoorIssuesCmd.AddCommand(newOpenDoorIssuesListCmd())
}

func initOpenDoorIssuesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization", "", "Filter by organization (Type|ID, e.g. Broker|123)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (Broker, Customer, Trucker)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (use with --organization-type)")
	cmd.Flags().String("not-organization-type", "", "Exclude by organization type")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runOpenDoorIssuesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseOpenDoorIssuesListOptions(cmd)
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
	query.Set("fields[open-door-issues]", "status,description,created-at,updated-at,organization,reported-by")
	query.Set("include", "organization,reported-by")
	query.Set("fields[users]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	organizationFilter := strings.TrimSpace(opts.Organization)
	if organizationFilter == "" && opts.OrganizationType != "" && opts.OrganizationID != "" {
		organizationFilter = opts.OrganizationType + "|" + opts.OrganizationID
	}
	if organizationFilter != "" {
		query.Set("filter[organization]", organizationFilter)
		if opts.OrganizationID != "" {
			query.Set("filter[organization-id]", organizationFilter)
		}
	} else if opts.OrganizationID != "" {
		return fmt.Errorf("--organization-id requires --organization-type or --organization")
	}
	setFilterIfPresent(query, "filter[organization-type]", opts.OrganizationType)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/open-door-issues", query)
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

	rows := buildOpenDoorIssueRows(resp)
	if opts.NotOrganizationType != "" {
		rows = filterOpenDoorIssuesByNotOrganizationType(rows, opts.NotOrganizationType)
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderOpenDoorIssuesTable(cmd, rows)
}

func parseOpenDoorIssuesListOptions(cmd *cobra.Command) (openDoorIssuesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organization, _ := cmd.Flags().GetString("organization")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	notOrganizationType, _ := cmd.Flags().GetString("not-organization-type")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return openDoorIssuesListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		Organization:        organization,
		OrganizationType:    organizationType,
		OrganizationID:      organizationID,
		NotOrganizationType: notOrganizationType,
		CreatedAtMin:        createdAtMin,
		CreatedAtMax:        createdAtMax,
		UpdatedAtMin:        updatedAtMin,
		UpdatedAtMax:        updatedAtMax,
	}, nil
}

func buildOpenDoorIssueRows(resp jsonAPIResponse) []openDoorIssueRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]openDoorIssueRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := openDoorIssueRow{
			ID:          resource.ID,
			Status:      strings.TrimSpace(stringAttr(resource.Attributes, "status")),
			Description: strings.TrimSpace(stringAttr(resource.Attributes, "description")),
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationID = rel.Data.ID
			row.OrganizationType = rel.Data.Type
			if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.Organization = strings.TrimSpace(stringAttr(org.Attributes, "company-name"))
				if row.Organization == "" {
					row.Organization = strings.TrimSpace(stringAttr(org.Attributes, "name"))
				}
			}
		}

		if rel, ok := resource.Relationships["reported-by"]; ok && rel.Data != nil {
			row.ReportedByID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ReportedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderOpenDoorIssuesTable(cmd *cobra.Command, rows []openDoorIssueRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No open door issues found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tDESCRIPTION\tORG\tREPORTED BY")
	for _, row := range rows {
		orgDisplay := row.Organization
		if orgDisplay == "" {
			orgDisplay = formatResourceRef(row.OrganizationType, row.OrganizationID)
		}
		reportedBy := row.ReportedByName
		if reportedBy == "" {
			reportedBy = row.ReportedByID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Description, 40),
			truncateString(orgDisplay, 25),
			truncateString(reportedBy, 20),
		)
	}
	return writer.Flush()
}

func filterOpenDoorIssuesByNotOrganizationType(rows []openDoorIssueRow, notOrganizationType string) []openDoorIssueRow {
	normalized := normalizeOrganizationType(notOrganizationType)
	if normalized == "" {
		return rows
	}

	filtered := rows[:0]
	for _, row := range rows {
		if strings.EqualFold(row.OrganizationType, normalized) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func normalizeOrganizationType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	orgType, _, err := parseOrganization(value + "|0")
	if err != nil {
		return strings.ToLower(value)
	}
	return orgType
}
