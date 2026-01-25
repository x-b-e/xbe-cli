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

type timeCardPreApprovalsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
}

type timeCardPreApprovalRow struct {
	ID                                 string `json:"id"`
	TenderJobScheduleShiftID           string `json:"tender_job_schedule_shift_id,omitempty"`
	CreatedByID                        string `json:"created_by_id,omitempty"`
	ExplicitStartAt                    string `json:"explicit_start_at,omitempty"`
	ExplicitEndAt                      string `json:"explicit_end_at,omitempty"`
	ShouldAutomaticallyCreateAndSubmit bool   `json:"should_automatically_create_and_submit"`
	SubmitAt                           string `json:"submit_at,omitempty"`
	CanUpdate                          bool   `json:"can_update"`
}

func newTimeCardPreApprovalsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time card pre-approvals",
		Long: `List time card pre-approvals.

Output Columns:
  ID          Pre-approval identifier
  SHIFT       Tender job schedule shift ID
  START       Explicit start time (if set)
  END         Explicit end time (if set)
  AUTO        Whether auto-submit is enabled
  SUBMIT AT   Calculated submission time
  CAN UPDATE  Whether the pre-approval can be updated
  CREATED BY  User who created the pre-approval

Filters:
  --tender-job-schedule-shift   Filter by tender job schedule shift ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time card pre-approvals
  xbe view time-card-pre-approvals list

  # Filter by tender job schedule shift
  xbe view time-card-pre-approvals list --tender-job-schedule-shift 123

  # JSON output
  xbe view time-card-pre-approvals list --json`,
		Args: cobra.NoArgs,
		RunE: runTimeCardPreApprovalsList,
	}
	initTimeCardPreApprovalsListFlags(cmd)
	return cmd
}

func init() {
	timeCardPreApprovalsCmd.AddCommand(newTimeCardPreApprovalsListCmd())
}

func initTimeCardPreApprovalsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTimeCardPreApprovalsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTimeCardPreApprovalsListOptions(cmd)
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
	query.Set("fields[time-card-pre-approvals]", "tender-job-schedule-shift,created-by,explicit-start-at,explicit-end-at,should-automatically-create-and-submit,submit-at,can-update")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)

	body, _, err := client.Get(cmd.Context(), "/v1/time-card-pre-approvals", query)
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

	rows := buildTimeCardPreApprovalRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTimeCardPreApprovalsTable(cmd, rows)
}

func parseTimeCardPreApprovalsListOptions(cmd *cobra.Command) (timeCardPreApprovalsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return timeCardPreApprovalsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TenderJobScheduleShift: tenderJobScheduleShift,
	}, nil
}

func buildTimeCardPreApprovalRows(resp jsonAPIResponse) []timeCardPreApprovalRow {
	rows := make([]timeCardPreApprovalRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := buildTimeCardPreApprovalRow(resource)
		rows = append(rows, row)
	}
	return rows
}

func buildTimeCardPreApprovalRow(resource jsonAPIResource) timeCardPreApprovalRow {
	attrs := resource.Attributes
	row := timeCardPreApprovalRow{
		ID:                                 resource.ID,
		ExplicitStartAt:                    formatDateTime(stringAttr(attrs, "explicit-start-at")),
		ExplicitEndAt:                      formatDateTime(stringAttr(attrs, "explicit-end-at")),
		ShouldAutomaticallyCreateAndSubmit: boolAttr(attrs, "should-automatically-create-and-submit"),
		SubmitAt:                           formatDateTime(stringAttr(attrs, "submit-at")),
		CanUpdate:                          boolAttr(attrs, "can-update"),
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func renderTimeCardPreApprovalsTable(cmd *cobra.Command, rows []timeCardPreApprovalRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No time card pre-approvals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT\tSTART\tEND\tAUTO\tSUBMIT AT\tCAN UPDATE\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderJobScheduleShiftID,
			truncateString(row.ExplicitStartAt, 20),
			truncateString(row.ExplicitEndAt, 20),
			formatBool(row.ShouldAutomaticallyCreateAndSubmit),
			truncateString(row.SubmitAt, 20),
			formatBool(row.CanUpdate),
			row.CreatedByID,
		)
	}
	return writer.Flush()
}
