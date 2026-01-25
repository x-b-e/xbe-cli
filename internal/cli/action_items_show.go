package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type actionItemsShowOptions struct {
	BaseURL         string
	Token           string
	JSON            bool
	NoAuth          bool
	ShowAllComments bool
}

type actionItemDetails struct {
	// Core fields
	ID          string `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	DueOn       string `json:"due_on,omitempty"`

	// Assignment
	ResponsiblePersonID   string `json:"responsible_person_id,omitempty"`
	ResponsiblePersonName string `json:"responsible_person_name,omitempty"`
	ResponsibleOrgID      string `json:"responsible_org_id,omitempty"`
	ResponsibleOrgType    string `json:"responsible_org_type,omitempty"`
	ResponsibleOrgName    string `json:"responsible_org_name,omitempty"`

	// Project context
	ProjectID   string `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`

	// Tracker info
	TrackerID          string `json:"tracker_id,omitempty"`
	TrackerPriority    int    `json:"tracker_priority,omitempty"`
	TrackerDevAssignee string `json:"tracker_dev_assignee,omitempty"`
	TrackerCSAssignee  string `json:"tracker_cs_assignee,omitempty"`

	// Meeting context
	MeetingID   string `json:"meeting_id,omitempty"`
	MeetingName string `json:"meeting_name,omitempty"`

	// Root cause
	RootCauseID   string `json:"root_cause_id,omitempty"`
	RootCauseName string `json:"root_cause_name,omitempty"`

	// Hierarchy
	ParentID    string            `json:"parent_id,omitempty"`
	ParentTitle string            `json:"parent_title,omitempty"`
	Children    []actionItemChild `json:"children,omitempty"`

	// Incident (nested)
	Incident *incidentInfo `json:"incident,omitempty"`

	// Arrays
	TeamMembers []teamMember     `json:"team_members,omitempty"`
	LineItems   []lineItem       `json:"line_items,omitempty"`
	KeyResults  []keyResultLink  `json:"key_results,omitempty"`
	Comments    []comment        `json:"comments,omitempty"`
	Attachments []fileAttachment `json:"attachments,omitempty"`
}

type actionItemChild struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type incidentInfo struct {
	ID       string `json:"id"`
	Subject  string `json:"subject,omitempty"`
	JobPlan  string `json:"job_plan,omitempty"`
	Customer string `json:"customer,omitempty"`
}

type teamMember struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

type lineItem struct {
	ID                    string `json:"id"`
	Title                 string `json:"title,omitempty"`
	Status                string `json:"status,omitempty"`
	DueOn                 string `json:"due_on,omitempty"`
	CreatedAt             string `json:"created_at,omitempty"`
	Description           string `json:"description,omitempty"`
	ResponsiblePersonID   string `json:"responsible_person_id,omitempty"`
	ResponsiblePersonName string `json:"responsible_person_name,omitempty"`
}

type keyResultLink struct {
	ID            string `json:"id"`
	KeyResultID   string `json:"key_result_id"`
	KeyResultName string `json:"key_result_name,omitempty"`
	ObjectiveName string `json:"objective_name,omitempty"`
}

type comment struct {
	ID          string `json:"id"`
	Body        string `json:"body"`
	CreatedAt   string `json:"created_at"`
	CreatedByID string `json:"created_by_id,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
}

type fileAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
}

func newActionItemsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show action item details",
		Long: `Show the full details of a specific action item.

Retrieves and displays comprehensive information about an action item
including assignment, project context, team members, comments, and more.

Output Sections (table format):
  Core Info       ID, title, status, kind, due date
  Assignment      Responsible person and organization
  Project         Associated project
  Tracker         Priority and assignees
  Hierarchy       Parent and child action items
  Incident        Linked incident with job/customer info
  Team Members    Users assigned to this item
  Line Items      Sub-tasks with responsible persons
  Key Results     Linked OKRs
  Comments        Discussion thread
  Attachments     Attached files
  Description     Full description text

Arguments:
  <id>          The action item ID (required). Find IDs using the list command.`,
		Example: `  # View an action item by ID
  xbe view action-items show 6195

  # Get action item as JSON
  xbe view action-items show 6195 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runActionItemsShow,
	}
	initActionItemsShowFlags(cmd)
	return cmd
}

func init() {
	actionItemsCmd.AddCommand(newActionItemsShowCmd())
}

func initActionItemsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Bool("show-all-comments", false, "Show all comments (default shows 3 most recent)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseActionItemsShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("action item id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	// CLI-optimized include (excludes reactions, update-requests, deep broker nesting)
	query.Set("include", "responsible-organization,responsible-person,project,tracker,tracker.dev-assignee,tracker.customer-success-assignee,meeting,root-cause,parent-action-item,child-action-items,action-item-team-members,action-item-team-members.user,action-item-line-items,action-item-line-items.responsible-person,action-item-key-results,action-item-key-results.key-result,action-item-key-results.key-result.objective,incident,incident.job-production-plan,incident.job-production-plan.customer,comments,comments.created-by,file-attachments,file-attachments.created-by")

	// Sparse fieldsets for known stable types
	query.Set("fields[users]", "name")
	query.Set("fields[projects]", "name")
	query.Set("fields[action-item-trackers]", "priority")
	query.Set("fields[job-production-plans]", "job-number,job-name")
	query.Set("fields[customers]", "company-name")
	query.Set("fields[brokers]", "company-name")
	query.Set("fields[truckers]", "company-name")
	query.Set("fields[material-suppliers]", "name")
	query.Set("fields[developers]", "name")
	// Note: Other resource types return all fields (meetings, root-causes, incidents,
	// action-item-team-members, action-item-line-items, action-item-key-results,
	// key-results, objectives, comments, file-attachments)

	body, _, err := client.Get(cmd.Context(), "/v1/action-items/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildActionItemDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderActionItemDetails(cmd, details, opts)
}

func parseActionItemsShowOptions(cmd *cobra.Command) (actionItemsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	showAllComments, _ := cmd.Flags().GetBool("show-all-comments")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemsShowOptions{
		BaseURL:         baseURL,
		Token:           token,
		JSON:            jsonOut,
		NoAuth:          noAuth,
		ShowAllComments: showAllComments,
	}, nil
}

func buildActionItemDetails(resp jsonAPISingleResponse) actionItemDetails {
	attrs := resp.Data.Attributes

	// Build included lookup
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	details := actionItemDetails{
		ID:          resp.Data.ID,
		Title:       strings.TrimSpace(stringAttr(attrs, "title")),
		Status:      stringAttr(attrs, "status"),
		Kind:        stringAttr(attrs, "kind"),
		Description: strings.TrimSpace(stringAttr(attrs, "description")),
		CreatedAt:   formatDate(stringAttr(attrs, "created-at")),
		UpdatedAt:   formatDate(stringAttr(attrs, "updated-at")),
		DueOn:       formatDate(stringAttr(attrs, "due-on")),
	}

	// Responsible person
	if rel, ok := resp.Data.Relationships["responsible-person"]; ok && rel.Data != nil {
		details.ResponsiblePersonID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ResponsiblePersonName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	// Responsible organization (polymorphic)
	if rel, ok := resp.Data.Relationships["responsible-organization"]; ok && rel.Data != nil {
		details.ResponsibleOrgID = rel.Data.ID
		details.ResponsibleOrgType = rel.Data.Type
		if org, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ResponsibleOrgName = firstNonEmpty(
				stringAttr(org.Attributes, "company-name"),
				stringAttr(org.Attributes, "name"),
			)
		}
	}

	// Project
	if rel, ok := resp.Data.Relationships["project"]; ok && rel.Data != nil {
		details.ProjectID = rel.Data.ID
		if proj, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ProjectName = strings.TrimSpace(stringAttr(proj.Attributes, "name"))
		}
	}

	// Tracker with nested assignees
	if rel, ok := resp.Data.Relationships["tracker"]; ok && rel.Data != nil {
		details.TrackerID = rel.Data.ID
		if tracker, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			if priority, ok := tracker.Attributes["priority"]; ok && priority != nil {
				if p, ok := priority.(float64); ok {
					details.TrackerPriority = int(p)
				}
			}
			// Dev assignee
			if devRel, ok := tracker.Relationships["dev-assignee"]; ok && devRel.Data != nil {
				if dev, ok := included[resourceKey(devRel.Data.Type, devRel.Data.ID)]; ok {
					details.TrackerDevAssignee = strings.TrimSpace(stringAttr(dev.Attributes, "name"))
				}
			}
			// CS assignee
			if csRel, ok := tracker.Relationships["customer-success-assignee"]; ok && csRel.Data != nil {
				if cs, ok := included[resourceKey(csRel.Data.Type, csRel.Data.ID)]; ok {
					details.TrackerCSAssignee = strings.TrimSpace(stringAttr(cs.Attributes, "name"))
				}
			}
		}
	}

	// Meeting
	if rel, ok := resp.Data.Relationships["meeting"]; ok && rel.Data != nil {
		details.MeetingID = rel.Data.ID
		if meeting, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.MeetingName = strings.TrimSpace(stringAttr(meeting.Attributes, "name"))
		}
	}

	// Root cause
	if rel, ok := resp.Data.Relationships["root-cause"]; ok && rel.Data != nil {
		details.RootCauseID = rel.Data.ID
		if rc, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.RootCauseName = strings.TrimSpace(stringAttr(rc.Attributes, "name"))
		}
	}

	// Parent action item
	if rel, ok := resp.Data.Relationships["parent-action-item"]; ok && rel.Data != nil {
		details.ParentID = rel.Data.ID
		if parent, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ParentTitle = strings.TrimSpace(stringAttr(parent.Attributes, "title"))
		}
	}

	// Child action items (array)
	if rel, ok := resp.Data.Relationships["child-action-items"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if child, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					details.Children = append(details.Children, actionItemChild{
						ID:     child.ID,
						Title:  strings.TrimSpace(stringAttr(child.Attributes, "title")),
						Status: stringAttr(child.Attributes, "status"),
					})
				}
			}
		}
	}

	// Incident with nested job-production-plan and customer
	if rel, ok := resp.Data.Relationships["incident"]; ok && rel.Data != nil {
		if incident, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			info := &incidentInfo{
				ID:      incident.ID,
				Subject: strings.TrimSpace(stringAttr(incident.Attributes, "subject")),
			}
			// Job production plan
			if jppRel, ok := incident.Relationships["job-production-plan"]; ok && jppRel.Data != nil {
				if jpp, ok := included[resourceKey(jppRel.Data.Type, jppRel.Data.ID)]; ok {
					jobNumber := stringAttr(jpp.Attributes, "job-number")
					jobName := stringAttr(jpp.Attributes, "job-name")
					if jobNumber != "" && jobName != "" {
						info.JobPlan = fmt.Sprintf("%s - %s", jobNumber, jobName)
					} else {
						info.JobPlan = firstNonEmpty(jobNumber, jobName)
					}
					// Customer
					if custRel, ok := jpp.Relationships["customer"]; ok && custRel.Data != nil {
						if cust, ok := included[resourceKey(custRel.Data.Type, custRel.Data.ID)]; ok {
							info.Customer = strings.TrimSpace(stringAttr(cust.Attributes, "company-name"))
						}
					}
				}
			}
			details.Incident = info
		}
	}

	// Team members (array with nested user)
	if rel, ok := resp.Data.Relationships["action-item-team-members"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if tm, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					member := teamMember{ID: tm.ID}
					if userRel, ok := tm.Relationships["user"]; ok && userRel.Data != nil {
						member.UserID = userRel.Data.ID
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							member.UserName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.TeamMembers = append(details.TeamMembers, member)
				}
			}
		}
	}

	// Line items (array with nested responsible-person)
	if rel, ok := resp.Data.Relationships["action-item-line-items"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if li, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					item := lineItem{
						ID:          li.ID,
						Title:       strings.TrimSpace(stringAttr(li.Attributes, "title")),
						Status:      stringAttr(li.Attributes, "status"),
						DueOn:       formatDate(stringAttr(li.Attributes, "due-on")),
						CreatedAt:   stringAttr(li.Attributes, "created-at"),
						Description: strings.TrimSpace(stringAttr(li.Attributes, "description")),
					}
					if rpRel, ok := li.Relationships["responsible-person"]; ok && rpRel.Data != nil {
						item.ResponsiblePersonID = rpRel.Data.ID
						if rp, ok := included[resourceKey(rpRel.Data.Type, rpRel.Data.ID)]; ok {
							item.ResponsiblePersonName = strings.TrimSpace(stringAttr(rp.Attributes, "name"))
						}
					}
					details.LineItems = append(details.LineItems, item)
				}
			}
		}
		// Sort line items by created_at (oldest first)
		sort.Slice(details.LineItems, func(i, j int) bool {
			return details.LineItems[i].CreatedAt < details.LineItems[j].CreatedAt
		})
	}

	// Key results (array with nested key-result and objective)
	if rel, ok := resp.Data.Relationships["action-item-key-results"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if aikr, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					link := keyResultLink{ID: aikr.ID}
					if krRel, ok := aikr.Relationships["key-result"]; ok && krRel.Data != nil {
						link.KeyResultID = krRel.Data.ID
						if kr, ok := included[resourceKey(krRel.Data.Type, krRel.Data.ID)]; ok {
							link.KeyResultName = strings.TrimSpace(stringAttr(kr.Attributes, "name"))
							// Objective
							if objRel, ok := kr.Relationships["objective"]; ok && objRel.Data != nil {
								if obj, ok := included[resourceKey(objRel.Data.Type, objRel.Data.ID)]; ok {
									link.ObjectiveName = strings.TrimSpace(stringAttr(obj.Attributes, "name"))
								}
							}
						}
					}
					details.KeyResults = append(details.KeyResults, link)
				}
			}
		}
	}

	// Comments (array with nested created-by)
	if rel, ok := resp.Data.Relationships["comments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if c, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					cmt := comment{
						ID:        c.ID,
						Body:      strings.TrimSpace(stringAttr(c.Attributes, "body")),
						CreatedAt: formatDateTime(stringAttr(c.Attributes, "created-at")),
					}
					if userRel, ok := c.Relationships["created-by"]; ok && userRel.Data != nil {
						cmt.CreatedByID = userRel.Data.ID
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							cmt.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.Comments = append(details.Comments, cmt)
				}
			}
		}
	}

	// File attachments (array with nested created-by)
	if rel, ok := resp.Data.Relationships["file-attachments"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if fa, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					att := fileAttachment{
						ID:          fa.ID,
						Filename:    strings.TrimSpace(stringAttr(fa.Attributes, "filename")),
						ContentType: stringAttr(fa.Attributes, "content-type"),
					}
					if userRel, ok := fa.Relationships["created-by"]; ok && userRel.Data != nil {
						if user, ok := included[resourceKey(userRel.Data.Type, userRel.Data.ID)]; ok {
							att.CreatedBy = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.Attachments = append(details.Attachments, att)
				}
			}
		}
	}

	return details
}

func renderActionItemDetails(cmd *cobra.Command, d actionItemDetails, opts actionItemsShowOptions) error {
	out := cmd.OutOrStdout()

	// Core info
	fmt.Fprintf(out, "ID: %s\n", d.ID)
	fmt.Fprintf(out, "Title: %s\n", d.Title)
	fmt.Fprintf(out, "Status: %s\n", d.Status)
	fmt.Fprintf(out, "Kind: %s\n", d.Kind)
	if d.DueOn != "" {
		fmt.Fprintf(out, "Due: %s\n", d.DueOn)
	}
	if d.CreatedAt != "" {
		fmt.Fprintf(out, "Created: %s\n", d.CreatedAt)
	}
	if d.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated: %s\n", d.UpdatedAt)
	}

	// Assignment
	if d.ResponsiblePersonID != "" || d.ResponsibleOrgID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Assignment:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.ResponsiblePersonName != "" {
			fmt.Fprintf(out, "  Responsible: %s (ID: %s)\n", d.ResponsiblePersonName, d.ResponsiblePersonID)
		} else if d.ResponsiblePersonID != "" {
			fmt.Fprintf(out, "  Responsible: (ID: %s)\n", d.ResponsiblePersonID)
		}
		if d.ResponsibleOrgName != "" {
			fmt.Fprintf(out, "  Organization: %s (%s, ID: %s)\n", d.ResponsibleOrgName, d.ResponsibleOrgType, d.ResponsibleOrgID)
		} else if d.ResponsibleOrgID != "" {
			fmt.Fprintf(out, "  Organization: (%s, ID: %s)\n", d.ResponsibleOrgType, d.ResponsibleOrgID)
		}
	}

	// Project
	if d.ProjectID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Project:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.ProjectName != "" {
			fmt.Fprintf(out, "  Name: %s (ID: %s)\n", d.ProjectName, d.ProjectID)
		} else {
			fmt.Fprintf(out, "  ID: %s\n", d.ProjectID)
		}
	}

	// Tracker
	if d.TrackerID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Tracker:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintf(out, "  ID: %s\n", d.TrackerID)
		if d.TrackerPriority != 0 {
			fmt.Fprintf(out, "  Priority: %d\n", d.TrackerPriority)
		}
		if d.TrackerDevAssignee != "" {
			fmt.Fprintf(out, "  Dev Assignee: %s\n", d.TrackerDevAssignee)
		}
		if d.TrackerCSAssignee != "" {
			fmt.Fprintf(out, "  CS Assignee: %s\n", d.TrackerCSAssignee)
		}
	}

	// Meeting
	if d.MeetingID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Meeting:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.MeetingName != "" {
			fmt.Fprintf(out, "  Name: %s (ID: %s)\n", d.MeetingName, d.MeetingID)
		} else {
			fmt.Fprintf(out, "  ID: %s\n", d.MeetingID)
		}
	}

	// Root cause
	if d.RootCauseID != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Root Cause:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.RootCauseName != "" {
			fmt.Fprintf(out, "  Name: %s (ID: %s)\n", d.RootCauseName, d.RootCauseID)
		} else {
			fmt.Fprintf(out, "  ID: %s\n", d.RootCauseID)
		}
	}

	// Hierarchy
	if d.ParentID != "" || len(d.Children) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Hierarchy:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		if d.ParentID != "" {
			if d.ParentTitle != "" {
				fmt.Fprintf(out, "  Parent: %s (ID: %s)\n", d.ParentTitle, d.ParentID)
			} else {
				fmt.Fprintf(out, "  Parent ID: %s\n", d.ParentID)
			}
		}
		if len(d.Children) > 0 {
			fmt.Fprintf(out, "  Children (%d):\n", len(d.Children))
			for _, child := range d.Children {
				fmt.Fprintf(out, "    - [%s] %s (ID: %s)\n", child.Status, child.Title, child.ID)
			}
		}
	}

	// Incident
	if d.Incident != nil {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Incident:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintf(out, "  ID: %s\n", d.Incident.ID)
		if d.Incident.Subject != "" {
			fmt.Fprintf(out, "  Subject: %s\n", d.Incident.Subject)
		}
		if d.Incident.JobPlan != "" {
			fmt.Fprintf(out, "  Job: %s\n", d.Incident.JobPlan)
		}
		if d.Incident.Customer != "" {
			fmt.Fprintf(out, "  Customer: %s\n", d.Incident.Customer)
		}
	}

	// Team members
	if len(d.TeamMembers) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Team Members (%d):\n", len(d.TeamMembers))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, tm := range d.TeamMembers {
			if tm.UserName != "" {
				fmt.Fprintf(out, "  - %s (ID: %s)\n", tm.UserName, tm.UserID)
			} else {
				fmt.Fprintf(out, "  - User ID: %s\n", tm.UserID)
			}
		}
	}

	// Line items
	if len(d.LineItems) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Line Items (%d):\n", len(d.LineItems))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for i, li := range d.LineItems {
			// Show title with status if available
			title := li.Title
			if title == "" {
				title = li.Description
			}
			if title == "" {
				title = "(no title)"
			}
			if li.Status != "" {
				fmt.Fprintf(out, "  %d. [%s] %s\n", i+1, li.Status, title)
			} else {
				fmt.Fprintf(out, "  %d. %s\n", i+1, title)
			}
			if li.DueOn != "" {
				fmt.Fprintf(out, "     Due: %s\n", li.DueOn)
			}
			if li.ResponsiblePersonName != "" {
				fmt.Fprintf(out, "     Responsible: %s\n", li.ResponsiblePersonName)
			}
		}
	}

	// Key results
	if len(d.KeyResults) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Key Results (%d):\n", len(d.KeyResults))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, kr := range d.KeyResults {
			if kr.ObjectiveName != "" {
				fmt.Fprintf(out, "  - %s (Objective: %s)\n", kr.KeyResultName, kr.ObjectiveName)
			} else {
				fmt.Fprintf(out, "  - %s\n", kr.KeyResultName)
			}
		}
	}

	// Comments (show 3 most recent by default, all with --show-all-comments)
	if len(d.Comments) > 0 {
		totalComments := len(d.Comments)
		commentsToShow := d.Comments

		// By default, show only 3 most recent (comments are ordered newest first from API)
		if !opts.ShowAllComments && totalComments > 3 {
			// Take the first 3 (most recent)
			commentsToShow = d.Comments[:3]
			fmt.Fprintln(out, "")
			fmt.Fprintf(out, "Comments (showing 3 of %d, use --show-all-comments for all):\n", totalComments)
		} else {
			fmt.Fprintln(out, "")
			fmt.Fprintf(out, "Comments (%d):\n", totalComments)
		}
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, c := range commentsToShow {
			if c.CreatedBy != "" {
				fmt.Fprintf(out, "  [%s] %s:\n", c.CreatedAt, c.CreatedBy)
			} else {
				fmt.Fprintf(out, "  [%s]:\n", c.CreatedAt)
			}
			// Indent comment body
			lines := strings.Split(c.Body, "\n")
			for _, line := range lines {
				fmt.Fprintf(out, "    %s\n", line)
			}
			fmt.Fprintln(out, "")
		}
	}

	// Attachments
	if len(d.Attachments) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Attachments (%d):\n", len(d.Attachments))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, att := range d.Attachments {
			if att.ContentType != "" {
				fmt.Fprintf(out, "  - %s (%s)\n", att.Filename, att.ContentType)
			} else {
				fmt.Fprintf(out, "  - %s\n", att.Filename)
			}
			if att.CreatedBy != "" {
				fmt.Fprintf(out, "    Uploaded by: %s\n", att.CreatedBy)
			}
		}
	}

	// Description (at the end since it can be long)
	if d.Description != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description:")
		fmt.Fprintln(out, strings.Repeat("-", 40))
		fmt.Fprintln(out, d.Description)
	}

	return nil
}
