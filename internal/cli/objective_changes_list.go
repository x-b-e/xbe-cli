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

type objectiveChangesListOptions struct {
	BaseURL          string
	Token            string
	JSON             bool
	NoAuth           bool
	Limit            int
	Offset           int
	Sort             string
	Objective        string
	Organization     string
	OrganizationID   string
	OrganizationType string
	Broker           string
	ChangedBy        string
}

type objectiveChangeRow struct {
	ID               string `json:"id"`
	ObjectiveID      string `json:"objective_id,omitempty"`
	OrganizationType string `json:"organization_type,omitempty"`
	OrganizationID   string `json:"organization_id,omitempty"`
	BrokerID         string `json:"broker_id,omitempty"`
	ChangedByID      string `json:"changed_by_id,omitempty"`
	StartOnOld       string `json:"start_on_old,omitempty"`
	StartOnNew       string `json:"start_on_new,omitempty"`
	EndOnOld         string `json:"end_on_old,omitempty"`
	EndOnNew         string `json:"end_on_new,omitempty"`
}

func newObjectiveChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List objective changes",
		Long: `List objective changes with filtering and pagination.

Output Columns:
  ID            Objective change identifier
  OBJECTIVE     Objective ID
  ORGANIZATION  Organization (Type/ID)
  BROKER        Broker ID
  START OLD     Previous start date
  START NEW     Updated start date
  END OLD       Previous end date
  END NEW       Updated end date
  CHANGED BY    User who made the change

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  --objective          Filter by objective ID
  --organization       Filter by organization (format: Type|ID, e.g., Broker|123)
  --organization-id    Filter by organization ID (requires --organization-type or Type|ID)
  --organization-type  Filter by organization type (e.g., Broker, Customer, Trucker)
  --broker             Filter by broker ID
  --changed-by         Filter by changed-by user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List objective changes
  xbe view objective-changes list

  # Filter by objective
  xbe view objective-changes list --objective 123

  # Filter by organization
  xbe view objective-changes list --organization "Broker|123"

  # Filter by broker and changed-by user
  xbe view objective-changes list --broker 123 --changed-by 456

  # Output as JSON
  xbe view objective-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runObjectiveChangesList,
	}
	initObjectiveChangesListFlags(cmd)
	return cmd
}

func init() {
	objectiveChangesCmd.AddCommand(newObjectiveChangesListCmd())
}

func initObjectiveChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("objective", "", "Filter by objective ID")
	cmd.Flags().String("organization", "", "Filter by organization (format: Type|ID, e.g., Broker|123)")
	cmd.Flags().String("organization-id", "", "Filter by organization ID (requires --organization-type or Type|ID)")
	cmd.Flags().String("organization-type", "", "Filter by organization type (e.g., Broker, Customer, Trucker)")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("changed-by", "", "Filter by changed-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runObjectiveChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseObjectiveChangesListOptions(cmd)
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
	query.Set("fields[objective-changes]", "objective,start-on-old,start-on-new,end-on-old,end-on-new,organization,broker,changed-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[objective]", opts.Objective)
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
	setFilterIfPresent(query, "filter[changed_by]", opts.ChangedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/objective-changes", query)
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

	rows := buildObjectiveChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderObjectiveChangesTable(cmd, rows)
}

func parseObjectiveChangesListOptions(cmd *cobra.Command) (objectiveChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	objective, _ := cmd.Flags().GetString("objective")
	organization, _ := cmd.Flags().GetString("organization")
	organizationID, _ := cmd.Flags().GetString("organization-id")
	organizationType, _ := cmd.Flags().GetString("organization-type")
	broker, _ := cmd.Flags().GetString("broker")
	changedBy, _ := cmd.Flags().GetString("changed-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return objectiveChangesListOptions{
		BaseURL:          baseURL,
		Token:            token,
		JSON:             jsonOut,
		NoAuth:           noAuth,
		Limit:            limit,
		Offset:           offset,
		Sort:             sort,
		Objective:        objective,
		Organization:     organization,
		OrganizationID:   organizationID,
		OrganizationType: organizationType,
		Broker:           broker,
		ChangedBy:        changedBy,
	}, nil
}

func buildObjectiveChangeRows(resp jsonAPIResponse) []objectiveChangeRow {
	rows := make([]objectiveChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := objectiveChangeRow{
			ID:         resource.ID,
			StartOnOld: formatDate(stringAttr(resource.Attributes, "start-on-old")),
			StartOnNew: formatDate(stringAttr(resource.Attributes, "start-on-new")),
			EndOnOld:   formatDate(stringAttr(resource.Attributes, "end-on-old")),
			EndOnNew:   formatDate(stringAttr(resource.Attributes, "end-on-new")),
		}
		if rel, ok := resource.Relationships["objective"]; ok && rel.Data != nil {
			row.ObjectiveID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["organization"]; ok && rel.Data != nil {
			row.OrganizationType = rel.Data.Type
			row.OrganizationID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
			row.ChangedByID = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderObjectiveChangesTable(cmd *cobra.Command, rows []objectiveChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No objective changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tOBJECTIVE\tORGANIZATION\tBROKER\tSTART OLD\tSTART NEW\tEND OLD\tEND NEW\tCHANGED BY")
	for _, row := range rows {
		organization := formatPolymorphic(row.OrganizationType, row.OrganizationID)
		if organization == "" {
			organization = "-"
		}
		broker := row.BrokerID
		if broker == "" {
			broker = "-"
		}
		changedBy := row.ChangedByID
		if changedBy == "" {
			changedBy = "-"
		}
		objective := row.ObjectiveID
		if objective == "" {
			objective = "-"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			objective,
			organization,
			broker,
			row.StartOnOld,
			row.StartOnNew,
			row.EndOnOld,
			row.EndOnNew,
			changedBy,
		)
	}
	return writer.Flush()
}
