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

type tenderJobScheduleShiftTimeCardReviewsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
}

type tenderJobScheduleShiftTimeCardReviewRow struct {
	ID                        string `json:"id"`
	TenderJobScheduleShiftID  string `json:"tender_job_schedule_shift_id,omitempty"`
	TimeCardStartAt           string `json:"time_card_start_at,omitempty"`
	TimeCardEndAt             string `json:"time_card_end_at,omitempty"`
	TimeCardDownMinutes       int    `json:"time_card_down_minutes,omitempty"`
	TimeCardStartAtConfidence bool   `json:"time_card_start_at_confidence"`
	TimeCardEndAtConfidence   bool   `json:"time_card_end_at_confidence"`
}

func newTenderJobScheduleShiftTimeCardReviewsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender job schedule shift time card reviews",
		Long: `List tender job schedule shift time card reviews.

Output Columns:
  ID          Review identifier
  SHIFT       Tender job schedule shift ID
  START       Suggested time card start time
  END         Suggested time card end time
  DOWN        Suggested down minutes
  START CONF  Start time confidence
  END CONF    End time confidence

Filters:
  --tender-job-schedule-shift   Filter by tender job schedule shift ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List time card reviews
  xbe view tender-job-schedule-shift-time-card-reviews list

  # Filter by shift
  xbe view tender-job-schedule-shift-time-card-reviews list --tender-job-schedule-shift 123

  # JSON output
  xbe view tender-job-schedule-shift-time-card-reviews list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderJobScheduleShiftTimeCardReviewsList,
	}
	initTenderJobScheduleShiftTimeCardReviewsListFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftTimeCardReviewsCmd.AddCommand(newTenderJobScheduleShiftTimeCardReviewsListCmd())
}

func initTenderJobScheduleShiftTimeCardReviewsListFlags(cmd *cobra.Command) {
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

func runTenderJobScheduleShiftTimeCardReviewsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderJobScheduleShiftTimeCardReviewsListOptions(cmd)
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
	query.Set("fields[tender-job-schedule-shift-time-card-reviews]", "tender-job-schedule-shift,time-card-start-at,time-card-end-at,time-card-down-minutes,time-card-start-at-confidence,time-card-end-at-confidence")

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

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shift-time-card-reviews", query)
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

	rows := buildTenderJobScheduleShiftTimeCardReviewRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderJobScheduleShiftTimeCardReviewsTable(cmd, rows)
}

func parseTenderJobScheduleShiftTimeCardReviewsListOptions(cmd *cobra.Command) (tenderJobScheduleShiftTimeCardReviewsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftTimeCardReviewsListOptions{
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

func buildTenderJobScheduleShiftTimeCardReviewRows(resp jsonAPIResponse) []tenderJobScheduleShiftTimeCardReviewRow {
	rows := make([]tenderJobScheduleShiftTimeCardReviewRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildTenderJobScheduleShiftTimeCardReviewRow(resource))
	}
	return rows
}

func buildTenderJobScheduleShiftTimeCardReviewRow(resource jsonAPIResource) tenderJobScheduleShiftTimeCardReviewRow {
	attrs := resource.Attributes
	row := tenderJobScheduleShiftTimeCardReviewRow{
		ID:                        resource.ID,
		TimeCardStartAt:           formatDateTime(stringAttr(attrs, "time-card-start-at")),
		TimeCardEndAt:             formatDateTime(stringAttr(attrs, "time-card-end-at")),
		TimeCardDownMinutes:       intAttr(attrs, "time-card-down-minutes"),
		TimeCardStartAtConfidence: boolAttr(attrs, "time-card-start-at-confidence"),
		TimeCardEndAtConfidence:   boolAttr(attrs, "time-card-end-at-confidence"),
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}

	return row
}

func renderTenderJobScheduleShiftTimeCardReviewsTable(cmd *cobra.Command, rows []tenderJobScheduleShiftTimeCardReviewRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender job schedule shift time card reviews found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT\tSTART\tEND\tDOWN\tSTART CONF\tEND CONF")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
			row.ID,
			row.TenderJobScheduleShiftID,
			truncateString(row.TimeCardStartAt, 20),
			truncateString(row.TimeCardEndAt, 20),
			row.TimeCardDownMinutes,
			formatBool(row.TimeCardStartAtConfidence),
			formatBool(row.TimeCardEndAtConfidence),
		)
	}
	return writer.Flush()
}
