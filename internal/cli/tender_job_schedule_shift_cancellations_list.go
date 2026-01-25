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

type tenderJobScheduleShiftCancellationsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
}

type tenderJobScheduleShiftCancellationRow struct {
	ID                                   string `json:"id"`
	TenderJobScheduleShiftID             string `json:"tender_job_schedule_shift_id,omitempty"`
	StatusChangedByID                    string `json:"status_changed_by_id,omitempty"`
	StatusChangeComment                  string `json:"status_change_comment,omitempty"`
	IsReturned                           bool   `json:"is_returned"`
	JobProductionPlanCancellationComment string `json:"job_production_plan_cancellation_comment,omitempty"`
	SkipTruckerNotifications             bool   `json:"skip_trucker_notifications"`
}

func newTenderJobScheduleShiftCancellationsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tender job schedule shift cancellations",
		Long: `List tender job schedule shift cancellations with pagination.

Output Columns:
  ID           Cancellation identifier
  SHIFT        Tender job schedule shift ID
  RETURNED     Whether the tender was returned
  CHANGED BY   User who changed the status
  SKIP NOTIFS  Whether trucker notifications were skipped
  COMMENT      Status change comment

Pagination:
  Use --limit and --offset to paginate through large result sets.

Filters:
  None

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List cancellations
  xbe view tender-job-schedule-shift-cancellations list

  # Paginate results
  xbe view tender-job-schedule-shift-cancellations list --limit 25 --offset 50

  # Output as JSON
  xbe view tender-job-schedule-shift-cancellations list --json`,
		Args: cobra.NoArgs,
		RunE: runTenderJobScheduleShiftCancellationsList,
	}
	initTenderJobScheduleShiftCancellationsListFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftCancellationsCmd.AddCommand(newTenderJobScheduleShiftCancellationsListCmd())
}

func initTenderJobScheduleShiftCancellationsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftCancellationsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseTenderJobScheduleShiftCancellationsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tender-job-schedule-shift-cancellations]", "tender-job-schedule-shift,status-change-comment,status-changed-by,is-returned,skip-trucker-notifications")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shift-cancellations", query)
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

	rows := buildTenderJobScheduleShiftCancellationRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderTenderJobScheduleShiftCancellationsTable(cmd, rows)
}

func parseTenderJobScheduleShiftCancellationsListOptions(cmd *cobra.Command) (tenderJobScheduleShiftCancellationsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftCancellationsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
	}, nil
}

func buildTenderJobScheduleShiftCancellationRows(resp jsonAPIResponse) []tenderJobScheduleShiftCancellationRow {
	rows := make([]tenderJobScheduleShiftCancellationRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := tenderJobScheduleShiftCancellationRow{
			ID:                       resource.ID,
			StatusChangeComment:      stringAttr(resource.Attributes, "status-change-comment"),
			IsReturned:               boolAttr(resource.Attributes, "is-returned"),
			SkipTruckerNotifications: boolAttr(resource.Attributes, "skip-trucker-notifications"),
		}
		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShiftID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["status-changed-by"]; ok && rel.Data != nil {
			row.StatusChangedByID = rel.Data.ID
		}
		rows = append(rows, row)
	}
	return rows
}

func renderTenderJobScheduleShiftCancellationsTable(cmd *cobra.Command, rows []tenderJobScheduleShiftCancellationRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tender job schedule shift cancellations found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSHIFT\tRETURNED\tCHANGED BY\tSKIP NOTIFS\tCOMMENT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.TenderJobScheduleShiftID,
			boolToYesNo(row.IsReturned),
			row.StatusChangedByID,
			boolToYesNo(row.SkipTruckerNotifications),
			truncateString(row.StatusChangeComment, 40),
		)
	}
	return writer.Flush()
}
