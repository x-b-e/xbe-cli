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

type actionItemLineItemsListOptions struct {
	BaseURL           string
	Token             string
	JSON              bool
	NoAuth            bool
	Limit             int
	Offset            int
	Sort              string
	Status            string
	DueOn             string
	DueOnMin          string
	DueOnMax          string
	HasDueOn          string
	ResponsiblePerson string
	ActionItem        string
}

type actionItemLineItemRow struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	Status                string `json:"status"`
	DueOn                 string `json:"due_on,omitempty"`
	ResponsiblePersonID   string `json:"responsible_person_id,omitempty"`
	ResponsiblePersonName string `json:"responsible_person_name,omitempty"`
	ActionItemID          string `json:"action_item_id,omitempty"`
	ActionItemTitle       string `json:"action_item_title,omitempty"`
}

func newActionItemLineItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List action item line items",
		Long: `List action item line items with filtering and pagination.

Output Columns:
  ID           Action item line item ID
  STATUS       Status (open/closed)
  TITLE        Line item title (truncated)
  DUE ON       Due date
  RESPONSIBLE  Responsible person (if assigned)
  ACTION ITEM  Parent action item

Filters:
  --action-item         Filter by action item ID (comma-separated for multiple)
  --responsible-person  Filter by responsible person user ID (comma-separated for multiple)
  --status              Filter by status (comma-separated: open,closed)
  --due-on              Filter by due date (YYYY-MM-DD)
  --due-on-min          Filter by minimum due date (YYYY-MM-DD)
  --due-on-max          Filter by maximum due date (YYYY-MM-DD)
  --has-due-on          Filter by due date presence (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List action item line items
  xbe view action-item-line-items list

  # Filter by action item
  xbe view action-item-line-items list --action-item 123

  # Filter by responsible person
  xbe view action-item-line-items list --responsible-person 456

  # Filter by status
  xbe view action-item-line-items list --status open

  # Filter by due date range
  xbe view action-item-line-items list --due-on-min 2025-01-01 --due-on-max 2025-01-31

  # Output as JSON
  xbe view action-item-line-items list --json`,
		Args: cobra.NoArgs,
		RunE: runActionItemLineItemsList,
	}
	initActionItemLineItemsListFlags(cmd)
	return cmd
}

func init() {
	actionItemLineItemsCmd.AddCommand(newActionItemLineItemsListCmd())
}

func initActionItemLineItemsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort order")
	cmd.Flags().String("status", "", "Filter by status (comma-separated: open,closed)")
	cmd.Flags().String("due-on", "", "Filter by due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-min", "", "Filter by minimum due date (YYYY-MM-DD)")
	cmd.Flags().String("due-on-max", "", "Filter by maximum due date (YYYY-MM-DD)")
	cmd.Flags().String("has-due-on", "", "Filter by due date presence (true/false)")
	cmd.Flags().String("responsible-person", "", "Filter by responsible person user ID (comma-separated for multiple)")
	cmd.Flags().String("action-item", "", "Filter by action item ID (comma-separated for multiple)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemLineItemsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseActionItemLineItemsListOptions(cmd)
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
	query.Set("fields[action-item-line-items]", "title,status,due-on,responsible-person,action-item,created-at,updated-at")
	query.Set("include", "responsible-person,action-item")
	query.Set("fields[users]", "name")
	query.Set("fields[action-items]", "title")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[due-on]", opts.DueOn)
	setFilterIfPresent(query, "filter[due-on-min]", opts.DueOnMin)
	setFilterIfPresent(query, "filter[due-on-max]", opts.DueOnMax)
	setFilterIfPresent(query, "filter[has-due-on]", opts.HasDueOn)
	setFilterIfPresent(query, "filter[responsible-person]", opts.ResponsiblePerson)
	setFilterIfPresent(query, "filter[action-item]", opts.ActionItem)

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-line-items", query)
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

	rows := buildActionItemLineItemRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderActionItemLineItemsTable(cmd, rows)
}

func parseActionItemLineItemsListOptions(cmd *cobra.Command) (actionItemLineItemsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	dueOn, _ := cmd.Flags().GetString("due-on")
	dueOnMin, _ := cmd.Flags().GetString("due-on-min")
	dueOnMax, _ := cmd.Flags().GetString("due-on-max")
	hasDueOn, _ := cmd.Flags().GetString("has-due-on")
	responsiblePerson, _ := cmd.Flags().GetString("responsible-person")
	actionItem, _ := cmd.Flags().GetString("action-item")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemLineItemsListOptions{
		BaseURL:           baseURL,
		Token:             token,
		JSON:              jsonOut,
		NoAuth:            noAuth,
		Limit:             limit,
		Offset:            offset,
		Sort:              sort,
		Status:            status,
		DueOn:             dueOn,
		DueOnMin:          dueOnMin,
		DueOnMax:          dueOnMax,
		HasDueOn:          hasDueOn,
		ResponsiblePerson: responsiblePerson,
		ActionItem:        actionItem,
	}, nil
}

func buildActionItemLineItemRows(resp jsonAPIResponse) []actionItemLineItemRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		key := resourceKey(inc.Type, inc.ID)
		included[key] = inc
	}

	rows := make([]actionItemLineItemRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := actionItemLineItemRow{
			ID:     resource.ID,
			Title:  strings.TrimSpace(stringAttr(resource.Attributes, "title")),
			Status: stringAttr(resource.Attributes, "status"),
			DueOn:  formatDate(stringAttr(resource.Attributes, "due-on")),
		}

		if rel, ok := resource.Relationships["responsible-person"]; ok && rel.Data != nil {
			row.ResponsiblePersonID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.ResponsiblePersonName = strings.TrimSpace(stringAttr(inc.Attributes, "name"))
			}
		}

		if rel, ok := resource.Relationships["action-item"]; ok && rel.Data != nil {
			row.ActionItemID = rel.Data.ID
			key := resourceKey(rel.Data.Type, rel.Data.ID)
			if inc, ok := included[key]; ok {
				row.ActionItemTitle = strings.TrimSpace(stringAttr(inc.Attributes, "title"))
			}
		}

		rows = append(rows, row)
	}

	return rows
}

func buildActionItemLineItemRowFromSingle(resp jsonAPISingleResponse) actionItemLineItemRow {
	rows := buildActionItemLineItemRows(jsonAPIResponse{Data: []jsonAPIResource{resp.Data}, Included: resp.Included})
	if len(rows) > 0 {
		return rows[0]
	}
	return actionItemLineItemRow{ID: resp.Data.ID}
}

func renderActionItemLineItemsTable(cmd *cobra.Command, rows []actionItemLineItemRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No action item line items found.")
		return nil
	}

	const titleMax = 40
	const responsibleMax = 20
	const actionItemMax = 30

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tTITLE\tDUE ON\tRESPONSIBLE\tACTION ITEM")
	for _, row := range rows {
		responsible := row.ResponsiblePersonName
		if responsible == "" {
			responsible = row.ResponsiblePersonID
		}
		actionItem := row.ActionItemTitle
		if actionItem == "" {
			actionItem = row.ActionItemID
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Title, titleMax),
			row.DueOn,
			truncateString(responsible, responsibleMax),
			truncateString(actionItem, actionItemMax),
		)
	}
	return writer.Flush()
}
