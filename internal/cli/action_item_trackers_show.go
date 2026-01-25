package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type actionItemTrackersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type actionItemTrackerUpdateRequestInfo struct {
	ID                string `json:"id"`
	RequestNote       string `json:"request_note,omitempty"`
	DueOn             string `json:"due_on,omitempty"`
	UpdateNote        string `json:"update_note,omitempty"`
	RequestedByID     string `json:"requested_by_id,omitempty"`
	RequestedByName   string `json:"requested_by_name,omitempty"`
	RequestedFromID   string `json:"requested_from_id,omitempty"`
	RequestedFromName string `json:"requested_from_name,omitempty"`
}

type actionItemTrackerDetails struct {
	ID                          string                               `json:"id"`
	Status                      string                               `json:"status,omitempty"`
	Priority                    int                                  `json:"priority,omitempty"`
	PriorityIndex               int                                  `json:"priority_index,omitempty"`
	PriorityPosition            string                               `json:"priority_position,omitempty"`
	DevEffortMinutes            int                                  `json:"dev_effort_minutes,omitempty"`
	DevEffortSize               string                               `json:"dev_effort_size,omitempty"`
	HasDueDateAgreement         bool                                 `json:"has_due_date_agreement"`
	IsUnplanned                 bool                                 `json:"is_unplanned"`
	ActionItemID                string                               `json:"action_item_id,omitempty"`
	ActionItemTitle             string                               `json:"action_item_title,omitempty"`
	ActionItemStatus            string                               `json:"action_item_status,omitempty"`
	ActionItemKind              string                               `json:"action_item_kind,omitempty"`
	DevAssigneeID               string                               `json:"dev_assignee_id,omitempty"`
	DevAssigneeName             string                               `json:"dev_assignee_name,omitempty"`
	CustomerSuccessAssigneeID   string                               `json:"customer_success_assignee_id,omitempty"`
	CustomerSuccessAssigneeName string                               `json:"customer_success_assignee_name,omitempty"`
	UpdateRequests              []actionItemTrackerUpdateRequestInfo `json:"update_requests,omitempty"`
}

func newActionItemTrackersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show action item tracker details",
		Long: `Show the full details of an action item tracker.

Output Fields:
  ID                       Tracker identifier
  Status                   Tracker status
  Priority                 Rank priority
  Priority Index           Index within active trackers
  Priority Position        Requested position change (if set)
  Dev Effort Minutes       Effort estimate in minutes
  Dev Effort Size          Size bucket (s, m, l, xl, xxl)
  Has Due Date Agreement   Due date agreed (true/false)
  Is Unplanned             Unplanned tracker (true/false)
  Action Item              Linked action item
  Dev Assignee             Development assignee
  CS Assignee              Customer success assignee
  Update Requests          Tracker update requests

Arguments:
  <id>    Action item tracker ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show an action item tracker
  xbe view action-item-trackers show 123

  # JSON output
  xbe view action-item-trackers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runActionItemTrackersShow,
	}
	initActionItemTrackersShowFlags(cmd)
	return cmd
}

func init() {
	actionItemTrackersCmd.AddCommand(newActionItemTrackersShowCmd())
}

func initActionItemTrackersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemTrackersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseActionItemTrackersShowOptions(cmd)
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

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("action item tracker id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[action-item-trackers]", "priority,priority-index,priority-position,dev-effort-minutes,has-due-date-agreement,is-unplanned,status,dev-effort-size,action-item,dev-assignee,customer-success-assignee,update-requests")
	query.Set("fields[action-items]", "title,status,kind")
	query.Set("fields[users]", "name")
	query.Set("fields[action-item-tracker-update-requests]", "request-note,due-on,update-note,requested-by,requested-from")
	query.Set("include", "action-item,dev-assignee,customer-success-assignee,update-requests,update-requests.requested-by,update-requests.requested-from")

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-trackers/"+id, query)
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

	details := buildActionItemTrackerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderActionItemTrackerDetails(cmd, details)
}

func parseActionItemTrackersShowOptions(cmd *cobra.Command) (actionItemTrackersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemTrackersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildActionItemTrackerDetails(resp jsonAPISingleResponse) actionItemTrackerDetails {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	attrs := resp.Data.Attributes
	details := actionItemTrackerDetails{
		ID:                  resp.Data.ID,
		Status:              stringAttr(attrs, "status"),
		Priority:            intAttr(attrs, "priority"),
		PriorityIndex:       intAttr(attrs, "priority-index"),
		PriorityPosition:    strings.TrimSpace(stringAttr(attrs, "priority-position")),
		DevEffortMinutes:    intAttr(attrs, "dev-effort-minutes"),
		DevEffortSize:       strings.TrimSpace(stringAttr(attrs, "dev-effort-size")),
		HasDueDateAgreement: boolAttr(attrs, "has-due-date-agreement"),
		IsUnplanned:         boolAttr(attrs, "is-unplanned"),
	}

	if rel, ok := resp.Data.Relationships["action-item"]; ok && rel.Data != nil {
		details.ActionItemID = rel.Data.ID
		if actionItem, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.ActionItemTitle = strings.TrimSpace(stringAttr(actionItem.Attributes, "title"))
			details.ActionItemStatus = stringAttr(actionItem.Attributes, "status")
			details.ActionItemKind = stringAttr(actionItem.Attributes, "kind")
		}
	}

	if rel, ok := resp.Data.Relationships["dev-assignee"]; ok && rel.Data != nil {
		details.DevAssigneeID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.DevAssigneeName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["customer-success-assignee"]; ok && rel.Data != nil {
		details.CustomerSuccessAssigneeID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.CustomerSuccessAssigneeName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	if rel, ok := resp.Data.Relationships["update-requests"]; ok && rel.raw != nil {
		var refs []jsonAPIResourceIdentifier
		if err := json.Unmarshal(rel.raw, &refs); err == nil {
			for _, ref := range refs {
				if updateReq, ok := included[resourceKey(ref.Type, ref.ID)]; ok {
					info := actionItemTrackerUpdateRequestInfo{
						ID:          updateReq.ID,
						RequestNote: strings.TrimSpace(stringAttr(updateReq.Attributes, "request-note")),
						DueOn:       formatDate(stringAttr(updateReq.Attributes, "due-on")),
						UpdateNote:  strings.TrimSpace(stringAttr(updateReq.Attributes, "update-note")),
					}
					if rbRel, ok := updateReq.Relationships["requested-by"]; ok && rbRel.Data != nil {
						info.RequestedByID = rbRel.Data.ID
						if user, ok := included[resourceKey(rbRel.Data.Type, rbRel.Data.ID)]; ok {
							info.RequestedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					if rfRel, ok := updateReq.Relationships["requested-from"]; ok && rfRel.Data != nil {
						info.RequestedFromID = rfRel.Data.ID
						if user, ok := included[resourceKey(rfRel.Data.Type, rfRel.Data.ID)]; ok {
							info.RequestedFromName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
						}
					}
					details.UpdateRequests = append(details.UpdateRequests, info)
				}
			}
		}
	}

	return details
}

func renderActionItemTrackerDetails(cmd *cobra.Command, details actionItemTrackerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.Status != "" {
		fmt.Fprintf(out, "Status: %s\n", details.Status)
	}
	fmt.Fprintf(out, "Priority: %d\n", details.Priority)
	fmt.Fprintf(out, "Priority Index: %d\n", details.PriorityIndex)
	if details.PriorityPosition != "" {
		fmt.Fprintf(out, "Priority Position: %s\n", details.PriorityPosition)
	}
	if details.DevEffortMinutes != 0 {
		fmt.Fprintf(out, "Dev Effort Minutes: %d\n", details.DevEffortMinutes)
	}
	if details.DevEffortSize != "" {
		fmt.Fprintf(out, "Dev Effort Size: %s\n", details.DevEffortSize)
	}
	fmt.Fprintf(out, "Has Due Date Agreement: %t\n", details.HasDueDateAgreement)
	fmt.Fprintf(out, "Is Unplanned: %t\n", details.IsUnplanned)

	writeLabelWithID(out, "Action Item", details.ActionItemTitle, details.ActionItemID)
	if details.ActionItemStatus != "" {
		fmt.Fprintf(out, "Action Item Status: %s\n", details.ActionItemStatus)
	}
	if details.ActionItemKind != "" {
		fmt.Fprintf(out, "Action Item Kind: %s\n", details.ActionItemKind)
	}
	writeLabelWithID(out, "Dev Assignee", details.DevAssigneeName, details.DevAssigneeID)
	writeLabelWithID(out, "CS Assignee", details.CustomerSuccessAssigneeName, details.CustomerSuccessAssigneeID)

	if len(details.UpdateRequests) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Update Requests (%d):\n", len(details.UpdateRequests))
		fmt.Fprintln(out, strings.Repeat("-", 40))
		for _, req := range details.UpdateRequests {
			label := req.RequestNote
			if label == "" {
				label = "(no request note)"
			}
			fmt.Fprintf(out, "  - %s (ID: %s)\n", label, req.ID)
			if req.DueOn != "" {
				fmt.Fprintf(out, "    Due On: %s\n", req.DueOn)
			}
			if req.UpdateNote != "" {
				fmt.Fprintf(out, "    Update Note: %s\n", req.UpdateNote)
			}
			if req.RequestedByName != "" || req.RequestedByID != "" {
				writeLabelWithID(out, "    Requested By", req.RequestedByName, req.RequestedByID)
			}
			if req.RequestedFromName != "" || req.RequestedFromID != "" {
				writeLabelWithID(out, "    Requested From", req.RequestedFromName, req.RequestedFromID)
			}
		}
	}

	return nil
}
