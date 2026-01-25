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

type changeRequestsListOptions struct {
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
	Broker              string
	CreatedBy           string
}

type changeRequestRow struct {
	ID               string `json:"id"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	CreatedByID      string `json:"created_by_id,omitempty"`
	RequestsCount    int    `json:"requests_count,omitempty"`
}

func newChangeRequestsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List change requests",
		Long: `List change requests.

Output Columns:
  ID            Change request identifier
  ORGANIZATION  Organization type and ID
  BROKER        Broker ID
  CREATED BY    User ID who created the change request
  REQUESTS      Number of request items

Filters:
  --organization            Filter by organization (format: Type|ID, e.g., Broker|123)
  --organization-id         Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type       Filter by organization type (e.g., Broker, Customer, Trucker)
  --not-organization-type   Exclude organization type
  --broker                  Filter by broker ID
  --created-by              Filter by creator user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List change requests
  xbe view change-requests list

  # Filter by organization
  xbe view change-requests list --organization "Broker|123"

  # Filter by broker and creator
  xbe view change-requests list --broker 123 --created-by 456

  # Output as JSON
  xbe view change-requests list --json`,
		Args: cobra.NoArgs,
		RunE: runChangeRequestsList,
	}
	initChangeRequestsListFlags(cmd)
	return cmd
}

func init() {
	changeRequestsCmd.AddCommand(newChangeRequestsListCmd())
}

func initChangeRequestsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("organization", "", "Filter by organization (format: Type|ID, e.g., Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer, Trucker)")
	cmd.Flags().String("not-organization-type", "", "Exclude organization type")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by creator user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runChangeRequestsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseChangeRequestsListOptions(cmd)
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
	query.Set("fields[change-requests]", "requests,organization,broker,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[organization]", opts.Organization)
	organizationIDFilter, err := buildOrganizationIDFilter(opts.OrganizationType, opts.OrganizationID)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if organizationIDFilter != "" {
		query.Set("filter[organization_id]", organizationIDFilter)
	} else {
		setFilterIfPresent(query, "filter[organization_type]", opts.OrganizationType)
	}
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/change-requests", query)
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

	rows := buildChangeRequestRows(resp)
	if strings.TrimSpace(opts.NotOrganizationType) != "" {
		rows = filterChangeRequestsByOrganizationType(rows, opts.NotOrganizationType)
	}
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderChangeRequestsTable(cmd, rows)
}

func parseChangeRequestsListOptions(cmd *cobra.Command) (changeRequestsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	organization, _ := cmd.Flags().GetString("organization")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	notOrganizationType, _ := cmd.Flags().GetString("not-organization-type")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return changeRequestsListOptions{
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
		Broker:              broker,
		CreatedBy:           createdBy,
	}, nil
}

func buildChangeRequestRows(resp jsonAPIResponse) []changeRequestRow {
	rows := make([]changeRequestRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := changeRequestRow{
			ID:            resource.ID,
			RequestsCount: requestCountFromAny(resource.Attributes["requests"]),
		}

		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func filterChangeRequestsByOrganizationType(rows []changeRequestRow, organizationType string) []changeRequestRow {
	filterType := normalizeOrganizationType(organizationType)
	if filterType == "" {
		return rows
	}
	filtered := make([]changeRequestRow, 0, len(rows))
	for _, row := range rows {
		if normalizeOrganizationType(row.OrganizationType) != filterType {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func renderChangeRequestsTable(cmd *cobra.Command, rows []changeRequestRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No change requests found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tORGANIZATION\tBROKER\tCREATED BY\tREQUESTS")
	for _, row := range rows {
		organization := formatPolymorphic(row.OrganizationType, row.OrganizationID)
		if organization == "" {
			organization = "-"
		}
		broker := row.BrokerID
		if broker == "" {
			broker = "-"
		}
		createdBy := row.CreatedByID
		if createdBy == "" {
			createdBy = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%d\n",
			row.ID,
			organization,
			broker,
			createdBy,
			row.RequestsCount,
		)
	}
	return writer.Flush()
}

func requestCountFromAny(value any) int {
	if value == nil {
		return 0
	}
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []map[string]any:
		return len(typed)
	default:
		return 1
	}
}
