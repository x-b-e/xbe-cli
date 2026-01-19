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

type actionItemsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Status  string
	Kind    string
	Project string
	Tracker string
	Sort    string
}

func newActionItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List action items",
		Long: `List action items with filtering and pagination.

Returns a list of action items matching the specified criteria. Action items
represent trackable work such as tasks, bugs, features, and integrations.

Output Columns (table format):
  ID            Unique action item identifier
  STATUS        Current status (open, in_progress, etc.)
  KIND          Type of work (feature, bug_fix, etc.)
  TITLE         Action item title (truncated)
  RESPONSIBLE   Assigned person or organization
  PROJECT       Associated project name

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filtering:
  Multiple filters can be combined. All filters use AND logic.

Status Values:
  editing, ready_for_work, in_progress, in_verification, complete, on_hold

Kind Values:
  feature, integration, sombrero, bug_fix, change_management, data_seeding, training

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.
  Default: action_item_tracker.priority,id`,
		Example: `  # List all action items
  xbe view action-items list

  # Filter by status
  xbe view action-items list --status in_progress
  xbe view action-items list --status ready_for_work

  # Filter by kind
  xbe view action-items list --kind bug_fix
  xbe view action-items list --kind feature

  # Combine filters
  xbe view action-items list --status in_progress --kind feature

  # Filter by project
  xbe view action-items list --project 123

  # Filter by tracker
  xbe view action-items list --tracker 456

  # Sort by created date (descending)
  xbe view action-items list --sort -created-at

  # Sort by priority
  xbe view action-items list --sort action_item_tracker.priority

  # Paginate results
  xbe view action-items list --limit 50 --offset 100

  # Output as JSON
  xbe view action-items list --json`,
		RunE: runActionItemsList,
	}
	initActionItemsListFlags(cmd)
	return cmd
}

func init() {
	actionItemsCmd.AddCommand(newActionItemsListCmd())
}

func initActionItemsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Page size (defaults to server default)")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("status", "", "Filter by status (editing/ready_for_work/in_progress/in_verification/complete/on_hold)")
	cmd.Flags().String("kind", "", "Filter by kind (feature/integration/sombrero/bug_fix/change_management/data_seeding/training)")
	cmd.Flags().String("project", "", "Filter by project ID")
	cmd.Flags().String("tracker", "", "Filter by tracker ID")
	cmd.Flags().String("sort", "", "Sort order (default: action_item_tracker.priority,id)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseActionItemsListOptions(cmd)
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
	query.Set("include", "created-by,responsible-person,responsible-organization,project,tracker")
	// Sparse fieldsets for included resources
	query.Set("fields[users]", "name")
	query.Set("fields[projects]", "name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")
	query.Set("fields[action-item-trackers]", "priority")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}

	// Apply filters
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[kind]", opts.Kind)
	setFilterIfPresent(query, "filter[project]", opts.Project)
	setFilterIfPresent(query, "filter[tracker]", opts.Tracker)

	// Apply sort
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "action_item_tracker.priority,id")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/action-items", query)
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
		rows := buildActionItemRows(resp)
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderActionItemsList(cmd, resp)
}

func parseActionItemsListOptions(cmd *cobra.Command) (actionItemsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	status, _ := cmd.Flags().GetString("status")
	kind, _ := cmd.Flags().GetString("kind")
	project, _ := cmd.Flags().GetString("project")
	tracker, _ := cmd.Flags().GetString("tracker")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Status:  status,
		Kind:    kind,
		Project: project,
		Tracker: tracker,
		Sort:    sort,
	}, nil
}

type actionItemRow struct {
	ID                    string `json:"id"`
	Status                string `json:"status"`
	Kind                  string `json:"kind"`
	Title                 string `json:"title"`
	CreatedByID           string `json:"created_by_id,omitempty"`
	CreatedByName         string `json:"created_by_name,omitempty"`
	ResponsiblePersonID   string `json:"responsible_person_id,omitempty"`
	ResponsiblePersonName string `json:"responsible_person_name,omitempty"`
	ResponsibleOrgID      string `json:"responsible_org_id,omitempty"`
	ResponsibleOrgType    string `json:"responsible_org_type,omitempty"`
	ResponsibleOrgName    string `json:"responsible_org_name,omitempty"`
	ProjectID             string `json:"project_id,omitempty"`
	ProjectName           string `json:"project_name,omitempty"`
	TrackerID             string `json:"tracker_id,omitempty"`
	TrackerName           string `json:"tracker_name,omitempty"`
	Priority              int    `json:"priority,omitempty"`
}

func buildActionItemRows(resp jsonAPIResponse) []actionItemRow {
	// Build included map for lookups
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]actionItemRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := actionItemRow{
			ID:     resource.ID,
			Status: stringAttr(resource.Attributes, "status"),
			Kind:   stringAttr(resource.Attributes, "kind"),
			Title:  strings.TrimSpace(stringAttr(resource.Attributes, "title")),
		}

		// Get created-by user info
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedByID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.CreatedByName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			}
		}

		// Get responsible person info
		if rel, ok := resource.Relationships["responsible-person"]; ok && rel.Data != nil {
			row.ResponsiblePersonID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.ResponsiblePersonName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			}
		}

		// Get responsible organization info (polymorphic)
		if rel, ok := resource.Relationships["responsible-organization"]; ok && rel.Data != nil {
			row.ResponsibleOrgID = rel.Data.ID
			row.ResponsibleOrgType = rel.Data.Type
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.ResponsibleOrgName = firstNonEmpty(
					stringAttr(inc.Attributes, "company-name"),
					stringAttr(inc.Attributes, "name"),
				)
			}
		}

		// Get project info
		if rel, ok := resource.Relationships["project"]; ok && rel.Data != nil {
			row.ProjectID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.ProjectName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			}
		}

		// Get tracker info
		if rel, ok := resource.Relationships["tracker"]; ok && rel.Data != nil {
			row.TrackerID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.TrackerName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
				if priority, ok := inc.Attributes["priority"]; ok {
					if p, ok := priority.(float64); ok {
						row.Priority = int(p)
					}
				}
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func renderActionItemsList(cmd *cobra.Command, resp jsonAPIResponse) error {
	rows := buildActionItemRows(resp)
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No action items found.")
		return nil
	}

	const titleMax = 40
	const responsibleMax = 20
	const projectMax = 20

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tKIND\tTITLE\tRESPONSIBLE\tPROJECT")
	for _, row := range rows {
		// Determine responsible display (prefer person over org)
		responsible := row.ResponsiblePersonName
		if responsible == "" {
			responsible = row.ResponsibleOrgName
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Kind,
			truncateString(row.Title, titleMax),
			truncateString(responsible, responsibleMax),
			truncateString(row.ProjectName, projectMax),
		)
	}
	return writer.Flush()
}
