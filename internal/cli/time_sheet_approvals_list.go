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

type timeSheetApprovalsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type timeSheetApprovalRow struct {
	ID          string `json:"id"`
	TimeSheetID string `json:"time_sheet_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newTimeSheetApprovalsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheet approvals",
		Long: `List time sheet approvals.

Output Columns:
  ID          Approval identifier
  TIME SHEET  Time sheet ID
  COMMENT     Approval comment (if present)

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List approvals
  xbe view time-sheet-approvals list

  # Paginate results
  xbe view time-sheet-approvals list --limit 25 --offset 50

  # Output as JSON
  xbe view time-sheet-approvals list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetApprovalsList,
	}
	initTimeSheetApprovalsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetApprovalsCmd.AddCommand(newTimeSheetApprovalsListCmd())
}

func initTimeSheetApprovalsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetApprovalsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetApprovalsListOptions(cmd)
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
	query.Set("fields[time-sheet-approvals]", "comment,time-sheet")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-approvals", query)
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

	rows := buildTimeSheetApprovalRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetApprovalsTable(cmd, rows)
}

func parseTimeSheetApprovalsListOptions(cmd *cobra.Command) (timeSheetApprovalsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetApprovalsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTimeSheetApprovalRows(resp jsonAPIResponse) []timeSheetApprovalRow {
	rows := make([]timeSheetApprovalRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := timeSheetApprovalRow{
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

func buildTimeSheetApprovalRowFromSingle(resp jsonAPISingleResponse) timeSheetApprovalRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := timeSheetApprovalRow{
		ID:      resource.ID,
		Comment: stringAttr(attrs, "comment"),
	}

	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		row.TimeSheetID = rel.Data.ID
	}

	return row
}

func renderTimeSheetApprovalsTable(cmd *cobra.Command, rows []timeSheetApprovalRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time sheet approvals found.")
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
