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

type jobScheduleShiftSplitsListOptions struct {
	BaseURL             string
	Token               string
	JSON                bool
	NoAuth              bool
	Limit               int
	Offset              int
	Sort                string
	JobScheduleShift    string
	NewJobScheduleShift string
}

type jobScheduleShiftSplitRow struct {
	ID                               string  `json:"id"`
	JobScheduleShiftID               string  `json:"job_schedule_shift_id,omitempty"`
	NewJobScheduleShiftID            string  `json:"new_job_schedule_shift_id,omitempty"`
	ExpectedMaterialTransactionCount int     `json:"expected_material_transaction_count,omitempty"`
	ExpectedMaterialTransactionTons  float64 `json:"expected_material_transaction_tons,omitempty"`
	NewStartAt                       string  `json:"new_start_at,omitempty"`
}

func newJobScheduleShiftSplitsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job schedule shift splits",
		Long: `List job schedule shift splits.

Output Columns:
  ID          Split identifier
  JOB SHIFT   Original job schedule shift ID
  NEW SHIFT   New job schedule shift ID (if created)
  COUNT       Expected material transaction count
  TONS        Expected material transaction tons
  NEW START   New start time for the created shift

Filters:
  --job-schedule-shift      Filter by job schedule shift ID
  --new-job-schedule-shift  Filter by new job schedule shift ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List shift splits
  xbe view job-schedule-shift-splits list

  # Filter by original shift
  xbe view job-schedule-shift-splits list --job-schedule-shift 123

  # Output as JSON
  xbe view job-schedule-shift-splits list --json`,
		Args: cobra.NoArgs,
		RunE: runJobScheduleShiftSplitsList,
	}
	initJobScheduleShiftSplitsListFlags(cmd)
	return cmd
}

func init() {
	jobScheduleShiftSplitsCmd.AddCommand(newJobScheduleShiftSplitsListCmd())
}

func initJobScheduleShiftSplitsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-schedule-shift", "", "Filter by job schedule shift ID")
	cmd.Flags().String("new-job-schedule-shift", "", "Filter by new job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobScheduleShiftSplitsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobScheduleShiftSplitsListOptions(cmd)
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
	query.Set("fields[job-schedule-shift-splits]", "expected-material-transaction-count,expected-material-transaction-tons,new-start-at,job-schedule-shift,new-job-schedule-shift")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[job-schedule-shift]", opts.JobScheduleShift)
	setFilterIfPresent(query, "filter[new-job-schedule-shift]", opts.NewJobScheduleShift)

	body, _, err := client.Get(cmd.Context(), "/v1/job-schedule-shift-splits", query)
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

	rows := buildJobScheduleShiftSplitRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobScheduleShiftSplitsTable(cmd, rows)
}

func parseJobScheduleShiftSplitsListOptions(cmd *cobra.Command) (jobScheduleShiftSplitsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	newJobScheduleShift, _ := cmd.Flags().GetString("new-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobScheduleShiftSplitsListOptions{
		BaseURL:             baseURL,
		Token:               token,
		JSON:                jsonOut,
		NoAuth:              noAuth,
		Limit:               limit,
		Offset:              offset,
		Sort:                sort,
		JobScheduleShift:    jobScheduleShift,
		NewJobScheduleShift: newJobScheduleShift,
	}, nil
}

func buildJobScheduleShiftSplitRows(resp jsonAPIResponse) []jobScheduleShiftSplitRow {
	rows := make([]jobScheduleShiftSplitRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildJobScheduleShiftSplitRow(resource))
	}
	return rows
}

func buildJobScheduleShiftSplitRow(resource jsonAPIResource) jobScheduleShiftSplitRow {
	attrs := resource.Attributes
	row := jobScheduleShiftSplitRow{
		ID:                               resource.ID,
		ExpectedMaterialTransactionCount: intAttr(attrs, "expected-material-transaction-count"),
		ExpectedMaterialTransactionTons:  floatAttr(attrs, "expected-material-transaction-tons"),
		NewStartAt:                       formatDateTime(stringAttr(attrs, "new-start-at")),
	}

	if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
		row.JobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["new-job-schedule-shift"]; ok && rel.Data != nil {
		row.NewJobScheduleShiftID = rel.Data.ID
	}

	return row
}

func buildJobScheduleShiftSplitRowFromSingle(resp jsonAPISingleResponse) jobScheduleShiftSplitRow {
	return buildJobScheduleShiftSplitRow(resp.Data)
}

func renderJobScheduleShiftSplitsTable(cmd *cobra.Command, rows []jobScheduleShiftSplitRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job schedule shift splits found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tJOB SHIFT\tNEW SHIFT\tCOUNT\tTONS\tNEW START")
	for _, row := range rows {
		count := formatOptionalInt(row.ExpectedMaterialTransactionCount)
		tons := formatOptionalFloat(row.ExpectedMaterialTransactionTons)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.JobScheduleShiftID,
			row.NewJobScheduleShiftID,
			count,
			tons,
			row.NewStartAt,
		)
	}
	return writer.Flush()
}

func formatOptionalInt(value int) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}
