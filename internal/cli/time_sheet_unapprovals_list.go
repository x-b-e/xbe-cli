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

type timeSheetUnapprovalsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type timeSheetUnapprovalRow struct {
	ID          string `json:"id"`
	TimeSheetID string `json:"time_sheet_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newTimeSheetUnapprovalsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheet unapprovals",
		Long: `List time sheet unapprovals.

Output Columns:
  ID          Unapproval identifier
  TIME SHEET  Time sheet ID
  COMMENT     Unapproval comment (if present)

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List unapprovals
  xbe view time-sheet-unapprovals list

  # Paginate results
  xbe view time-sheet-unapprovals list --limit 25 --offset 50

  # Output as JSON
  xbe view time-sheet-unapprovals list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetUnapprovalsList,
	}
	initTimeSheetUnapprovalsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetUnapprovalsCmd.AddCommand(newTimeSheetUnapprovalsListCmd())
}

func initTimeSheetUnapprovalsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetUnapprovalsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetUnapprovalsListOptions(cmd)
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
	query.Set("fields[time-sheet-unapprovals]", "comment,time-sheet")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-unapprovals", query)
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

	rows := buildTimeSheetUnapprovalRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetUnapprovalsTable(cmd, rows)
}

func parseTimeSheetUnapprovalsListOptions(cmd *cobra.Command) (timeSheetUnapprovalsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetUnapprovalsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTimeSheetUnapprovalRows(resp jsonAPIResponse) []timeSheetUnapprovalRow {
	rows := make([]timeSheetUnapprovalRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := timeSheetUnapprovalRow{
			ID:      resource.ID,
			Comment: stringAttr(attrs, "comment"),
		}

		if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
			row.TimeSheetID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func buildTimeSheetUnapprovalRowFromSingle(resp jsonAPISingleResponse) timeSheetUnapprovalRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := timeSheetUnapprovalRow{
		ID:      resource.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		row.TimeSheetID = rel.Data.ID
	}

	return row
}

func renderTimeSheetUnapprovalsTable(cmd *cobra.Command, rows []timeSheetUnapprovalRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time sheet unapprovals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME SHEET\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.TimeSheetID,
			truncateString(row.Comment, 50),
		)
	}
	return writer.Flush()
}
