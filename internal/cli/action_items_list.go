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
	BaseURL                     string
	Token                       string
	JSON                        bool
	NoAuth                      bool
	Limit                       int
	Offset                      int
	Status                      string
	Kind                        string
	Project                     string
	Tracker                     string
	Broker                      string
	Sort                        string
	Q                           string
	DueOn                       string
	DueOnMin                    string
	DueOnMax                    string
	CompletedOn                 string
	CompletedOnMin              string
	CompletedOnMax              string
	IsCompleted                 string
	ResponsiblePerson           string
	ResponsibleOrganization     string
	TeamMember                  string
	CreatedBy                   string
	Priority                    string
	IsDeleted                   string
	ParentActionItem            string
	Incident                    string
	RootCause                   string
	Meeting                     string
	IsUnplanned                 string
	RequiresXBEFeature          string
	Source                      string
	ExpectedCostAmountMin       string
	ExpectedCostAmountMax       string
	ExpectedBenefitAmountMin    string
	ExpectedBenefitAmountMax    string
	ExpectedNetBenefitAmountMin string
	ExpectedNetBenefitAmountMax string
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
  Multiple filters can be combined. Different filter types use AND logic (intersection),
  while multiple values within a filter use OR logic (union).

  Comma-separated values:
    --status and --kind accept comma-separated lists. For example:
      --status in_progress,ready_for_work
    This matches items that are in_progress OR ready_for_work.

  Cross-filter intersection:
    When combining different filters, they are intersected:
      --status in_progress --kind feature
    This matches items that are in_progress AND of kind feature.

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

  # Filter by multiple statuses (comma-separated, matches any)
  xbe view action-items list --status in_progress,ready_for_work

  # Filter for all incomplete items (everything except complete)
  xbe view action-items list --status editing,ready_for_work,in_progress,in_verification,on_hold

  # Filter by kind
  xbe view action-items list --kind bug_fix
  xbe view action-items list --kind feature

  # Filter by multiple kinds (comma-separated, matches any)
  xbe view action-items list --kind feature,bug_fix

  # Combine filters (intersection: must match both)
  xbe view action-items list --status in_progress --kind feature

  # Combine with multiple values (in_progress OR ready_for_work) AND (feature OR bug_fix)
  xbe view action-items list --status in_progress,ready_for_work --kind feature,bug_fix

  # Filter by project
  xbe view action-items list --project 123

  # Filter by tracker
  xbe view action-items list --tracker 456

  # Filter by broker
  xbe view action-items list --broker 49

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
	cmd.Flags().String("status", "", "Filter by status (comma-separated: editing,ready_for_work,in_progress,in_verification,complete,on_hold)")
	cmd.Flags().String("kind", "", "Filter by kind (comma-separated: feature,integration,sombrero,bug_fix,change_management,data_seeding,training)")
	cmd.Flags().String("project", "", "Filter by project ID (comma-separated for multiple)")
	cmd.Flags().String("tracker", "", "Filter by tracker ID (comma-separated for multiple)")
	cmd.Flags().String("broker", "", "Filter by broker ID (comma-separated for multiple)")
	cmd.Flags().String("sort", "", "Sort order (default: action_item_tracker.priority,id)")
	cmd.Flags().String("q", "", "Full-text search")
	cmd.Flags().String("due-on", "", "Filter by due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-min", "", "Filter by minimum due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-max", "", "Filter by maximum due date (YYYY-MM-DD)")
	cmd.Flags().String("completed-on", "", "Filter by completed date (YYYY-MM-DD)")
	cmd.Flags().String("completed-on-min", "", "Filter by minimum completed date (YYYY-MM-DD)")
	cmd.Flags().String("completed-on-max", "", "Filter by maximum completed date (YYYY-MM-DD)")
	cmd.Flags().String("is-completed", "", "Filter by completion status (true/false)")
	cmd.Flags().String("responsible-person", "", "Filter by responsible person user ID (comma-separated for multiple)")
	cmd.Flags().String("responsible-organization", "", "Filter by responsible organization (Type|ID, comma-separated for multiple)")
	cmd.Flags().String("team-member", "", "Filter by team member user ID (comma-separated for multiple)")
	cmd.Flags().String("created-by", "", "Filter by creator user ID (comma-separated for multiple)")
	cmd.Flags().String("priority", "", "Filter by priority")
	cmd.Flags().String("is-deleted", "", "Filter by deleted status (true/false)")
	cmd.Flags().String("parent-action-item", "", "Filter by parent action item ID (comma-separated for multiple)")
	cmd.Flags().String("incident", "", "Filter by incident ID (comma-separated for multiple)")
	cmd.Flags().String("root-cause", "", "Filter by root cause ID (comma-separated for multiple)")
	cmd.Flags().String("meeting", "", "Filter by meeting ID (comma-separated for multiple)")
	cmd.Flags().String("is-unplanned", "", "Filter by unplanned status (true/false)")
	cmd.Flags().String("requires-xbe-feature", "", "Filter by XBE feature requirement (true/false)")
	cmd.Flags().String("source", "", "Filter by source (Type|ID, comma-separated for multiple)")
	cmd.Flags().String("expected-cost-amount-min", "", "Filter by minimum expected cost amount")
	cmd.Flags().String("expected-cost-amount-max", "", "Filter by maximum expected cost amount")
	cmd.Flags().String("expected-benefit-amount-min", "", "Filter by minimum expected benefit amount")
	cmd.Flags().String("expected-benefit-amount-max", "", "Filter by maximum expected benefit amount")
	cmd.Flags().String("expected-net-benefit-amount-min", "", "Filter by minimum expected net benefit amount")
	cmd.Flags().String("expected-net-benefit-amount-max", "", "Filter by maximum expected net benefit amount")
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
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[q]", opts.Q)
	setFilterIfPresent(query, "filter[due-on]", opts.DueOn)
	setFilterIfPresent(query, "filter[due-on-min]", opts.DueOnMin)
	setFilterIfPresent(query, "filter[due-on-max]", opts.DueOnMax)
	setFilterIfPresent(query, "filter[completed-on]", opts.CompletedOn)
	setFilterIfPresent(query, "filter[completed-on-min]", opts.CompletedOnMin)
	setFilterIfPresent(query, "filter[completed-on-max]", opts.CompletedOnMax)
	setFilterIfPresent(query, "filter[is-completed]", opts.IsCompleted)
	setFilterIfPresent(query, "filter[responsible-person]", opts.ResponsiblePerson)
	setFilterIfPresent(query, "filter[responsible-organization]", opts.ResponsibleOrganization)
	setFilterIfPresent(query, "filter[team-member]", opts.TeamMember)
	setFilterIfPresent(query, "filter[created-by]", opts.CreatedBy)
	setFilterIfPresent(query, "filter[priority]", opts.Priority)
	setFilterIfPresent(query, "filter[is-deleted]", opts.IsDeleted)
	setFilterIfPresent(query, "filter[parent-action-item]", opts.ParentActionItem)
	setFilterIfPresent(query, "filter[incident]", opts.Incident)
	setFilterIfPresent(query, "filter[root-cause]", opts.RootCause)
	setFilterIfPresent(query, "filter[meeting]", opts.Meeting)
	setFilterIfPresent(query, "filter[is-unplanned]", opts.IsUnplanned)
	setFilterIfPresent(query, "filter[requires-xbe-feature]", opts.RequiresXBEFeature)
	setFilterIfPresent(query, "filter[source]", opts.Source)
	setFilterIfPresent(query, "filter[expected-cost-amount-min]", opts.ExpectedCostAmountMin)
	setFilterIfPresent(query, "filter[expected-cost-amount-max]", opts.ExpectedCostAmountMax)
	setFilterIfPresent(query, "filter[expected-benefit-amount-min]", opts.ExpectedBenefitAmountMin)
	setFilterIfPresent(query, "filter[expected-benefit-amount-max]", opts.ExpectedBenefitAmountMax)
	setFilterIfPresent(query, "filter[expected-net-benefit-amount-min]", opts.ExpectedNetBenefitAmountMin)
	setFilterIfPresent(query, "filter[expected-net-benefit-amount-max]", opts.ExpectedNetBenefitAmountMax)

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
	broker, _ := cmd.Flags().GetString("broker")
	sort, _ := cmd.Flags().GetString("sort")
	q, _ := cmd.Flags().GetString("q")
	dueOn, _ := cmd.Flags().GetString("due-on")
	dueOnMin, _ := cmd.Flags().GetString("due-on-min")
	dueOnMax, _ := cmd.Flags().GetString("due-on-max")
	completedOn, _ := cmd.Flags().GetString("completed-on")
	completedOnMin, _ := cmd.Flags().GetString("completed-on-min")
	completedOnMax, _ := cmd.Flags().GetString("completed-on-max")
	isCompleted, _ := cmd.Flags().GetString("is-completed")
	responsiblePerson, _ := cmd.Flags().GetString("responsible-person")
	responsibleOrganization, _ := cmd.Flags().GetString("responsible-organization")
	teamMember, _ := cmd.Flags().GetString("team-member")
	createdBy, _ := cmd.Flags().GetString("created-by")
	priority, _ := cmd.Flags().GetString("priority")
	isDeleted, _ := cmd.Flags().GetString("is-deleted")
	parentActionItem, _ := cmd.Flags().GetString("parent-action-item")
	incident, _ := cmd.Flags().GetString("incident")
	rootCause, _ := cmd.Flags().GetString("root-cause")
	meeting, _ := cmd.Flags().GetString("meeting")
	isUnplanned, _ := cmd.Flags().GetString("is-unplanned")
	requiresXBEFeature, _ := cmd.Flags().GetString("requires-xbe-feature")
	source, _ := cmd.Flags().GetString("source")
	expectedCostAmountMin, _ := cmd.Flags().GetString("expected-cost-amount-min")
	expectedCostAmountMax, _ := cmd.Flags().GetString("expected-cost-amount-max")
	expectedBenefitAmountMin, _ := cmd.Flags().GetString("expected-benefit-amount-min")
	expectedBenefitAmountMax, _ := cmd.Flags().GetString("expected-benefit-amount-max")
	expectedNetBenefitAmountMin, _ := cmd.Flags().GetString("expected-net-benefit-amount-min")
	expectedNetBenefitAmountMax, _ := cmd.Flags().GetString("expected-net-benefit-amount-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemsListOptions{
		BaseURL:                     baseURL,
		Token:                       token,
		JSON:                        jsonOut,
		NoAuth:                      noAuth,
		Limit:                       limit,
		Offset:                      offset,
		Status:                      status,
		Kind:                        kind,
		Project:                     project,
		Tracker:                     tracker,
		Broker:                      broker,
		Sort:                        sort,
		Q:                           q,
		DueOn:                       dueOn,
		DueOnMin:                    dueOnMin,
		DueOnMax:                    dueOnMax,
		CompletedOn:                 completedOn,
		CompletedOnMin:              completedOnMin,
		CompletedOnMax:              completedOnMax,
		IsCompleted:                 isCompleted,
		ResponsiblePerson:           responsiblePerson,
		ResponsibleOrganization:     responsibleOrganization,
		TeamMember:                  teamMember,
		CreatedBy:                   createdBy,
		Priority:                    priority,
		IsDeleted:                   isDeleted,
		ParentActionItem:            parentActionItem,
		Incident:                    incident,
		RootCause:                   rootCause,
		Meeting:                     meeting,
		IsUnplanned:                 isUnplanned,
		RequiresXBEFeature:          requiresXBEFeature,
		Source:                      source,
		ExpectedCostAmountMin:       expectedCostAmountMin,
		ExpectedCostAmountMax:       expectedCostAmountMax,
		ExpectedBenefitAmountMin:    expectedBenefitAmountMin,
		ExpectedBenefitAmountMax:    expectedBenefitAmountMax,
		ExpectedNetBenefitAmountMin: expectedNetBenefitAmountMin,
		ExpectedNetBenefitAmountMax: expectedNetBenefitAmountMax,
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
