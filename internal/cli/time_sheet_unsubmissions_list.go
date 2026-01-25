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

type timeSheetUnsubmissionsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type timeSheetUnsubmissionRow struct {
	ID          string `json:"id"`
	TimeSheetID string `json:"time_sheet_id,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

func newTimeSheetUnsubmissionsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time sheet unsubmissions",
		Long: `List time sheet unsubmissions.

Output Columns:
  ID         Unsubmission identifier
  TIME SHEET Time sheet ID
  COMMENT    Status change comment

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time sheet unsubmissions
  xbe view time-sheet-unsubmissions list

  # Output as JSON
  xbe view time-sheet-unsubmissions list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeSheetUnsubmissionsList,
	}
	initTimeSheetUnsubmissionsListFlags(cmd)
	return cmd
}

func init() {
	timeSheetUnsubmissionsCmd.AddCommand(newTimeSheetUnsubmissionsListCmd())
}

func initTimeSheetUnsubmissionsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeSheetUnsubmissionsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeSheetUnsubmissionsListOptions(cmd)
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
	query.Set("fields[time-sheet-unsubmissions]", "comment,time-sheet")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/time-sheet-unsubmissions", query)
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

	rows := buildTimeSheetUnsubmissionRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeSheetUnsubmissionsTable(cmd, rows)
}

func parseTimeSheetUnsubmissionsListOptions(cmd *cobra.Command) (timeSheetUnsubmissionsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeSheetUnsubmissionsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTimeSheetUnsubmissionRows(resp jsonAPIResponse) []timeSheetUnsubmissionRow {
	rows := make([]timeSheetUnsubmissionRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := timeSheetUnsubmissionRow{
			ID:      resource.ID,
			Comment: stringAttr(resource.Attributes, "comment"),
		}
		if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
			row.TimeSheetID = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func buildTimeSheetUnsubmissionRowFromSingle(resp jsonAPISingleResponse) timeSheetUnsubmissionRow {
	resource := resp.Data
	row := timeSheetUnsubmissionRow{
		ID:      resource.ID,
		Comment: stringAttr(resource.Attributes, "comment"),
	}
	if rel, ok := resource.Relationships["time-sheet"]; ok && rel.Data != nil {
		row.TimeSheetID = rel.Data.ID
	}
	return row
}

func renderTimeSheetUnsubmissionsTable(cmd *cobra.Command, rows []timeSheetUnsubmissionRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time sheet unsubmissions found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tTIME SHEET\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			row.ID,
			row.TimeSheetID,
			truncateString(row.Comment, 30),
		)
	}
	return writer.Flush()
}
