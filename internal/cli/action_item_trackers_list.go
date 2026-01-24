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

type actionItemTrackersListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type actionItemTrackerRow struct {
	ID                          string `json:"id"`
	Status                      string `json:"status,omitempty"`
	Priority                    int    `json:"priority,omitempty"`
	DevEffortSize               string `json:"dev_effort_size,omitempty"`
	ActionItemID                string `json:"action_item_id,omitempty"`
	ActionItemTitle             string `json:"action_item_title,omitempty"`
	DevAssigneeID               string `json:"dev_assignee_id,omitempty"`
	DevAssigneeName             string `json:"dev_assignee_name,omitempty"`
	CustomerSuccessAssigneeID   string `json:"customer_success_assignee_id,omitempty"`
	CustomerSuccessAssigneeName string `json:"customer_success_assignee_name,omitempty"`
}

func newActionItemTrackersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List action item trackers",
		Long: `List action item trackers with pagination.

Output Columns:
  ID           Tracker identifier
  STATUS       Tracker status
  PRIORITY     Priority rank (lower is higher priority)
  ACTION ITEM  Linked action item title
  DEV ASSIGNEE Development assignee
  CS ASSIGNEE  Customer success assignee

Pagination:
  Use --limit and --offset to paginate through large result sets.

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.
  Default: priority

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List action item trackers
  xbe view action-item-trackers list

  # Sort by priority descending
  xbe view action-item-trackers list --sort -priority

  # Paginate results
  xbe view action-item-trackers list --limit 50 --offset 100

  # JSON output
  xbe view action-item-trackers list --json`,
		Args: cobra.NoArgs,
		RunE: runActionItemTrackersList,
	}
	initActionItemTrackersListFlags(cmd)
	return cmd
}

func init() {
	actionItemTrackersCmd.AddCommand(newActionItemTrackersListCmd())
}

func initActionItemTrackersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order (default: priority)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemTrackersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseActionItemTrackersListOptions(cmd)
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
	query.Set("fields[action-item-trackers]", "priority,status,dev-effort-size,action-item,dev-assignee,customer-success-assignee")
	query.Set("fields[action-items]", "title")
	query.Set("fields[users]", "name")
	query.Set("include", "action-item,dev-assignee,customer-success-assignee")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	} else {
		query.Set("sort", "priority")
	}

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-trackers", query)
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

	rows := buildActionItemTrackerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderActionItemTrackersTable(cmd, rows)
}

func parseActionItemTrackersListOptions(cmd *cobra.Command) (actionItemTrackersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemTrackersListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildActionItemTrackerRows(resp jsonAPIResponse) []actionItemTrackerRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]actionItemTrackerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := actionItemTrackerRow{
			ID:            resource.ID,
			Status:        stringAttr(resource.Attributes, "status"),
			Priority:      intAttr(resource.Attributes, "priority"),
			DevEffortSize: stringAttr(resource.Attributes, "dev-effort-size"),
		}

		if rel, ok := resource.Relationships["action-item"]; ok && rel.Data != nil {
			row.ActionItemID = rel.Data.ID
			if actionItem, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.ActionItemTitle = strings.TrimSpace(stringAttr(actionItem.Attributes, "title"))
			}
		}

		if rel, ok := resource.Relationships["dev-assignee"]; ok && rel.Data != nil {
			row.DevAssigneeID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.DevAssigneeName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		if rel, ok := resource.Relationships["customer-success-assignee"]; ok && rel.Data != nil {
			row.CustomerSuccessAssigneeID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.CustomerSuccessAssigneeName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func renderActionItemTrackersTable(cmd *cobra.Command, rows []actionItemTrackerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No action item trackers found.")
		return nil
	}

	const titleMax = 40
	const nameMax = 20

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tPRIORITY\tACTION ITEM\tDEV ASSIGNEE\tCS ASSIGNEE")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%d\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			row.Priority,
			truncateString(row.ActionItemTitle, titleMax),
			truncateString(row.DevAssigneeName, nameMax),
			truncateString(row.CustomerSuccessAssigneeName, nameMax),
		)
	}

	return writer.Flush()
}
