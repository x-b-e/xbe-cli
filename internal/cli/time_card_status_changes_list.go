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

type timeCardStatusChangesListOptions struct {
	BaseURL      string
	Token        string
	JSON         bool
	NoAuth       bool
	Limit        int
	Offset       int
	Sort         string
	TimeCardID   string
	Status       string
	CreatedAtMin string
	CreatedAtMax string
	IsCreatedAt  string
	UpdatedAtMin string
	UpdatedAtMax string
	IsUpdatedAt  string
}

type timeCardStatusChangeRow struct {
	ID            string `json:"id"`
	TimeCardID    string `json:"time_card_id,omitempty"`
	Status        string `json:"status,omitempty"`
	ChangedAt     string `json:"changed_at,omitempty"`
	Comment       string `json:"comment,omitempty"`
	ChangedByID   string `json:"changed_by_id,omitempty"`
	ChangedByName string `json:"changed_by_name,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

func newTimeCardStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time card status changes",
		Long: `List time card status changes.

Output Columns:
  ID          Status change identifier
  TIME CARD   Time card ID
  STATUS      Status value
  CHANGED AT  When the status change occurred
  CHANGED BY  User who made the change (when available)
  COMMENT     Status change comment

Filters:
  --time-card       Filter by time card ID
  --status          Filter by status value
  --created-at-min  Filter by created-at on/after (ISO 8601)
  --created-at-max  Filter by created-at on/before (ISO 8601)
  --is-created-at   Filter by presence of created-at (true/false)
  --updated-at-min  Filter by updated-at on/after (ISO 8601)
  --updated-at-max  Filter by updated-at on/before (ISO 8601)
  --is-updated-at   Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List status changes
  xbe view time-card-status-changes list

  # Filter by time card
  xbe view time-card-status-changes list --time-card 123

  # Filter by status
  xbe view time-card-status-changes list --status submitted

  # Filter by created-at range
  xbe view time-card-status-changes list \
    --created-at-min 2026-01-23T00:00:00Z \
    --created-at-max 2026-01-24T00:00:00Z

  # Output as JSON
  xbe view time-card-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeCardStatusChangesList,
	}
	initTimeCardStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	timeCardStatusChangesCmd.AddCommand(newTimeCardStatusChangesListCmd())
}

func initTimeCardStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-card", "", "Filter by time card ID")
	cmd.Flags().String("status", "", "Filter by status value")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeCardStatusChangesListOptions(cmd)
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
	query.Set("fields[time-card-status-changes]", "status,changed-at,comment,created-at,updated-at,time-card,changed-by")
	query.Set("fields[users]", "name")
	query.Set("include", "changed-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[time_card]", opts.TimeCardID)
	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[created_at_min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created_at_max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[is_created_at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[updated_at_min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated_at_max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is_updated_at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-status-changes", query)
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

	rows := buildTimeCardStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeCardStatusChangesTable(cmd, rows)
}

func parseTimeCardStatusChangesListOptions(cmd *cobra.Command) (timeCardStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeCardID, _ := cmd.Flags().GetString("time-card")
	status, _ := cmd.Flags().GetString("status")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardStatusChangesListOptions{
		BaseURL:      baseURL,
		Token:        token,
		JSON:         jsonOut,
		NoAuth:       noAuth,
		Limit:        limit,
		Offset:       offset,
		Sort:         sort,
		TimeCardID:   timeCardID,
		Status:       status,
		CreatedAtMin: createdAtMin,
		CreatedAtMax: createdAtMax,
		IsCreatedAt:  isCreatedAt,
		UpdatedAtMin: updatedAtMin,
		UpdatedAtMax: updatedAtMax,
		IsUpdatedAt:  isUpdatedAt,
	}, nil
}

func buildTimeCardStatusChangeRows(resp jsonAPIResponse) []timeCardStatusChangeRow {
	included := make(map[string]jsonAPIResource)
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]timeCardStatusChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTimeCardStatusChangeRow(resource, included))
	}

	return rows
}

func buildTimeCardStatusChangeRow(resource jsonAPIResource, included map[string]jsonAPIResource) timeCardStatusChangeRow {
	attrs := resource.Attributes
	row := timeCardStatusChangeRow{
		ID:        resource.ID,
		Status:    strings.TrimSpace(stringAttr(attrs, "status")),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
		CreatedAt: formatDateTime(stringAttr(attrs, "created-at")),
		UpdatedAt: formatDateTime(stringAttr(attrs, "updated-at")),
	}

	if rel, ok := resource.Relationships["time-card"]; ok && rel.Data != nil {
		row.TimeCardID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		row.ChangedByID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.ChangedByName = strings.TrimSpace(stringAttr(user.Attributes, "name"))
		}
	}

	return row
}

func renderTimeCardStatusChangesTable(cmd *cobra.Command, rows []timeCardStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time card status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME CARD\tSTATUS\tCHANGED AT\tCHANGED BY\tCOMMENT")
	for _, row := range rows {
		changedBy := firstNonEmpty(row.ChangedByName, row.ChangedByID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TimeCardID,
			row.Status,
			row.ChangedAt,
			truncateString(changedBy, 18),
			truncateString(row.Comment, 32),
		)
	}

	return writer.Flush()
}
