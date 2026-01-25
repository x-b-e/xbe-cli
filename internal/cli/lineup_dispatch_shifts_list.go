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

type lineupDispatchShiftsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	LineupDispatch         string
	LineupJobScheduleShift string
	CreatedAtMin           string
	CreatedAtMax           string
	UpdatedAtMin           string
	UpdatedAtMax           string
}

type lineupDispatchShiftRow struct {
	ID                     string `json:"id"`
	LineupDispatchID       string `json:"lineup_dispatch_id,omitempty"`
	LineupJobScheduleShift string `json:"lineup_job_schedule_shift_id,omitempty"`
	FulfilledTruckerID     string `json:"fulfilled_trucker_id,omitempty"`
	CancelledAt            string `json:"cancelled_at,omitempty"`
}

func newLineupDispatchShiftsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup dispatch shifts",
		Long: `List lineup dispatch shifts.

Output Columns:
  ID              Lineup dispatch shift identifier
  DISPATCH        Lineup dispatch ID
  SCHEDULE SHIFT  Lineup job schedule shift ID
  FULFILLED BY    Fulfilled trucker ID
  CANCELLED AT    Cancellation timestamp

Filters:
  --lineup-dispatch            Filter by lineup dispatch ID
  --lineup-job-schedule-shift  Filter by lineup job schedule shift ID
  --created-at-min             Filter by created-at on/after (ISO 8601)
  --created-at-max             Filter by created-at on/before (ISO 8601)
  --updated-at-min             Filter by updated-at on/after (ISO 8601)
  --updated-at-max             Filter by updated-at on/before (ISO 8601)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup dispatch shifts
  xbe view lineup-dispatch-shifts list

  # Filter by lineup dispatch
  xbe view lineup-dispatch-shifts list --lineup-dispatch 123

  # Filter by lineup job schedule shift
  xbe view lineup-dispatch-shifts list --lineup-job-schedule-shift 456

  # Output as JSON
  xbe view lineup-dispatch-shifts list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupDispatchShiftsList,
	}
	initLineupDispatchShiftsListFlags(cmd)
	return cmd
}

func init() {
	lineupDispatchShiftsCmd.AddCommand(newLineupDispatchShiftsListCmd())
}

func initLineupDispatchShiftsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup-dispatch", "", "Filter by lineup dispatch ID")
	cmd.Flags().String("lineup-job-schedule-shift", "", "Filter by lineup job schedule shift ID")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupDispatchShiftsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupDispatchShiftsListOptions(cmd)
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
	query.Set("fields[lineup-dispatch-shifts]", "cancelled-at,lineup-dispatch,lineup-job-schedule-shift,fulfilled-trucker")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[lineup-dispatch]", opts.LineupDispatch)
	setFilterIfPresent(query, "filter[lineup-job-schedule-shift]", opts.LineupJobScheduleShift)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-dispatch-shifts", query)
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

	rows := buildLineupDispatchShiftRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupDispatchShiftsTable(cmd, rows)
}

func parseLineupDispatchShiftsListOptions(cmd *cobra.Command) (lineupDispatchShiftsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineupDispatch, _ := cmd.Flags().GetString("lineup-dispatch")
	lineupJobScheduleShift, _ := cmd.Flags().GetString("lineup-job-schedule-shift")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupDispatchShiftsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		LineupDispatch:         lineupDispatch,
		LineupJobScheduleShift: lineupJobScheduleShift,
		CreatedAtMin:           createdAtMin,
		CreatedAtMax:           createdAtMax,
		UpdatedAtMin:           updatedAtMin,
		UpdatedAtMax:           updatedAtMax,
	}, nil
}

func buildLineupDispatchShiftRows(resp jsonAPIResponse) []lineupDispatchShiftRow {
	rows := make([]lineupDispatchShiftRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := lineupDispatchShiftRow{
			ID:          resource.ID,
			CancelledAt: formatDateTime(stringAttr(attrs, "cancelled-at")),
		}

		if rel, ok := resource.Relationships["lineup-dispatch"]; ok && rel.Data != nil {
			row.LineupDispatchID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["lineup-job-schedule-shift"]; ok && rel.Data != nil {
			row.LineupJobScheduleShift = rel.Data.ID
		}
		if rel, ok := resource.Relationships["fulfilled-trucker"]; ok && rel.Data != nil {
			row.FulfilledTruckerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderLineupDispatchShiftsTable(cmd *cobra.Command, rows []lineupDispatchShiftRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup dispatch shifts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDISPATCH\tSCHEDULE SHIFT\tFULFILLED BY\tCANCELLED AT")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.LineupDispatchID,
			row.LineupJobScheduleShift,
			row.FulfilledTruckerID,
			row.CancelledAt,
		)
	}
	return writer.Flush()
}
