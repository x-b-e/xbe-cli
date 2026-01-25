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

type timeSheetStatusChangesListOptions struct {
	BaseURL   string
	Token     string
	JSON      bool
	NoAuth    bool
	Limit     int
	Offset    int
	Sort      string
	TimeSheet string
	Status    string
}

type timeSheetStatusChangeRow struct {
	ID          string `json:"id"`
	TimeSheetID string `json:"time_sheet_id,omitempty"`
	Status      string `json:"status,omitempty"`
	ChangedAt   string `json:"changed_at,omitempty"`
	ChangedByID string `json:"changed_by_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newTimeSheetStatusChangesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheet status changes",
		Long: `List time sheet status changes.

Output Columns:
  ID          Status change ID
  TIME SHEET  Time sheet ID
  STATUS      Status after the change
  CHANGED AT  Timestamp for the status change
  CHANGED BY  User who made the change (if present)
  COMMENT     Optional comment

Filters:
  --time-sheet   Filter by time sheet ID
  --status       Filter by status

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time sheet status changes
  xbe view time-sheet-status-changes list

  # Filter by time sheet
  xbe view time-sheet-status-changes list --time-sheet 123

  # Filter by status
  xbe view time-sheet-status-changes list --status approved

  # Output as JSON
  xbe view time-sheet-status-changes list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetStatusChangesList,
	}
	initTimeSheetStatusChangesListFlags(cmd)
	return cmd
}

func init() {
	timeSheetStatusChangesCmd.AddCommand(newTimeSheetStatusChangesListCmd())
}

func initTimeSheetStatusChangesListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("time-sheet", "", "Filter by time sheet ID")
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetStatusChangesList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetStatusChangesListOptions(cmd)
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
	query.Set("fields[time-sheet-status-changes]", "status,changed-at,comment,time-sheet,changed-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[time_sheet]", opts.TimeSheet)
	setFilterIfPresent(query, "filter[status]", opts.Status)

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-status-changes", query)
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

	handled, err := renderSparseListIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	rows := buildTimeSheetStatusChangeRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetStatusChangesTable(cmd, rows)
}

func parseTimeSheetStatusChangesListOptions(cmd *cobra.Command) (timeSheetStatusChangesListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	timeSheet, _ := cmd.Flags().GetString("time-sheet")
	status, _ := cmd.Flags().GetString("status")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetStatusChangesListOptions{
		BaseURL:   baseURL,
		Token:     token,
		JSON:      jsonOut,
		NoAuth:    noAuth,
		Limit:     limit,
		Offset:    offset,
		Sort:      sort,
		TimeSheet: timeSheet,
		Status:    status,
	}, nil
}

func buildTimeSheetStatusChangeRows(resp jsonAPIResponse) []timeSheetStatusChangeRow {
	rows := make([]timeSheetStatusChangeRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTimeSheetStatusChangeRow(resource))
	}
	return rows
}

func buildTimeSheetStatusChangeRow(resource jsonAPIResource) timeSheetStatusChangeRow {
	attrs := resource.Attributes
	row := timeSheetStatusChangeRow{
		ID:        resource.ID,
		Status:    stringAttr(attrs, "status"),
		ChangedAt: formatDateTime(stringAttr(attrs, "changed-at")),
		Comment:   stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		row.TimeSheetID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["changed-by"]; ok && rel.Data != nil {
		row.ChangedByID = rel.Data.ID
	}

	return row
}

func buildTimeSheetStatusChangeRowFromSingle(resp jsonAPISingleResponse) timeSheetStatusChangeRow {
	return buildTimeSheetStatusChangeRow(resp.Data)
}

func renderTimeSheetStatusChangesTable(cmd *cobra.Command, rows []timeSheetStatusChangeRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time sheet status changes found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME SHEET\tSTATUS\tCHANGED AT\tCHANGED BY\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TimeSheetID,
			row.Status,
			row.ChangedAt,
			row.ChangedByID,
			truncateString(row.Comment, 50),
		)
	}
	return writer.Flush()
}
