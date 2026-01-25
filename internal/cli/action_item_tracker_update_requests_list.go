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

type actionItemTrackerUpdateRequestsListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type actionItemTrackerUpdateRequestRow struct {
	ID                  string `json:"id"`
	ActionItemTrackerID string `json:"action_item_tracker_id,omitempty"`
	RequestedByID       string `json:"requested_by_id,omitempty"`
	RequestedByName     string `json:"requested_by_name,omitempty"`
	RequestedFromID     string `json:"requested_from_id,omitempty"`
	RequestedFromName   string `json:"requested_from_name,omitempty"`
	RequestNote         string `json:"request_note,omitempty"`
	DueOn               string `json:"due_on,omitempty"`
	UpdateNote          string `json:"update_note,omitempty"`
	Status              string `json:"status,omitempty"`
}

func newActionItemTrackerUpdateRequestsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List action item tracker update requests",
		Long: `List action item tracker update requests.

Output Columns:
  ID             Update request identifier
  TRACKER        Action item tracker ID
  REQUESTED BY   User who requested the update
  REQUESTED FROM User who should provide the update
  DUE ON         Requested due date
  STATUS         pending or fulfilled
  REQUEST NOTE   Requested update note (truncated)

Filters:
  --created-at-min   Filter by created-at on/after (ISO 8601)
  --created-at-max   Filter by created-at on/before (ISO 8601)
  --is-created-at    Filter by has created-at (true/false)
  --updated-at-min   Filter by updated-at on/after (ISO 8601)
  --updated-at-max   Filter by updated-at on/before (ISO 8601)
  --is-updated-at    Filter by has updated-at (true/false)

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List update requests
  xbe view action-item-tracker-update-requests list

  # Filter by created-at window
  xbe view action-item-tracker-update-requests list \
    --created-at-min 2024-01-01T00:00:00Z \
    --created-at-max 2024-12-31T23:59:59Z

  # Output as JSON
  xbe view action-item-tracker-update-requests list --json`,
		Args: cobra.NoArgs,
		RunE: runActionItemTrackerUpdateRequestsList,
	}
	initActionItemTrackerUpdateRequestsListFlags(cmd)
	return cmd
}

func init() {
	actionItemTrackerUpdateRequestsCmd.AddCommand(newActionItemTrackerUpdateRequestsListCmd())
}

func initActionItemTrackerUpdateRequestsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by has created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by has updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runActionItemTrackerUpdateRequestsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseActionItemTrackerUpdateRequestsListOptions(cmd)
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
	query.Set("fields[action-item-tracker-update-requests]", "request-note,due-on,update-note,action-item-tracker,requested-by,requested-from")
	query.Set("include", "requested-by,requested-from")
	query.Set("fields[users]", "name")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/action-item-tracker-update-requests", query)
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

	rows := buildActionItemTrackerUpdateRequestRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderActionItemTrackerUpdateRequestsTable(cmd, rows)
}

func parseActionItemTrackerUpdateRequestsListOptions(cmd *cobra.Command) (actionItemTrackerUpdateRequestsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return actionItemTrackerUpdateRequestsListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildActionItemTrackerUpdateRequestRows(resp jsonAPIResponse) []actionItemTrackerUpdateRequestRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]actionItemTrackerUpdateRequestRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := actionItemTrackerUpdateRequestRow{
			ID:          resource.ID,
			RequestNote: strings.TrimSpace(stringAttr(resource.Attributes, "request-note")),
			DueOn:       formatDate(stringAttr(resource.Attributes, "due-on")),
			UpdateNote:  strings.TrimSpace(stringAttr(resource.Attributes, "update-note")),
		}

		if strings.TrimSpace(row.UpdateNote) != "" {
			row.Status = "fulfilled"
		} else {
			row.Status = "pending"
		}

		if rel, ok := resource.Relationships["action-item-tracker"]; ok && rel.Data != nil {
			row.ActionItemTrackerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["requested-by"]; ok && rel.Data != nil {
			row.RequestedByID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.RequestedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}
		if rel, ok := resource.Relationships["requested-from"]; ok && rel.Data != nil {
			row.RequestedFromID = rel.Data.ID
			if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
				row.RequestedFromName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
			}
		}

		rows = append(rows, row)
	}
	return rows
}

func buildActionItemTrackerUpdateRequestRowFromSingle(resp jsonAPISingleResponse) actionItemTrackerUpdateRequestRow {
	row := actionItemTrackerUpdateRequestRow{
		ID:          resp.Data.ID,
		RequestNote: strings.TrimSpace(stringAttr(resp.Data.Attributes, "request-note")),
		DueOn:       formatDate(stringAttr(resp.Data.Attributes, "due-on")),
		UpdateNote:  strings.TrimSpace(stringAttr(resp.Data.Attributes, "update-note")),
	}

	if strings.TrimSpace(row.UpdateNote) != "" {
		row.Status = "fulfilled"
	} else {
		row.Status = "pending"
	}

	requestedByType := ""
	requestedFromType := ""
	if rel, ok := resp.Data.Relationships["action-item-tracker"]; ok && rel.Data != nil {
		row.ActionItemTrackerID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["requested-by"]; ok && rel.Data != nil {
		row.RequestedByID = rel.Data.ID
		requestedByType = rel.Data.Type
	}
	if rel, ok := resp.Data.Relationships["requested-from"]; ok && rel.Data != nil {
		row.RequestedFromID = rel.Data.ID
		requestedFromType = rel.Data.Type
	}

	if len(resp.Included) == 0 {
		return row
	}

	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	if row.RequestedByID != "" && requestedByType != "" {
		key := resourceKey(requestedByType, row.RequestedByID)
		if user, ok := included[key]; ok {
			row.RequestedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}
	if row.RequestedFromID != "" && requestedFromType != "" {
		key := resourceKey(requestedFromType, row.RequestedFromID)
		if user, ok := included[key]; ok {
			row.RequestedFromName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return row
}

func renderActionItemTrackerUpdateRequestsTable(cmd *cobra.Command, rows []actionItemTrackerUpdateRequestRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No action item tracker update requests found.")
		return nil
	}

	const nameMax = 20
	const noteMax = 40

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTRACKER\tREQUESTED BY\tREQUESTED FROM\tDUE ON\tSTATUS\tREQUEST NOTE")
	for _, row := range rows {
		requestedBy := row.RequestedByName
		if requestedBy == "" {
			requestedBy = row.RequestedByID
		}
		requestedFrom := row.RequestedFromName
		if requestedFrom == "" {
			requestedFrom = row.RequestedFromID
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ActionItemTrackerID,
			truncateString(requestedBy, nameMax),
			truncateString(requestedFrom, nameMax),
			row.DueOn,
			row.Status,
			truncateString(row.RequestNote, noteMax),
		)
	}
	return writer.Flush()
}
