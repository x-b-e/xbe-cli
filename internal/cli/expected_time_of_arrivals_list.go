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

type expectedTimeOfArrivalsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
	JobScheduleShift       string
	ExpectedAtMin          string
	ExpectedAtMax          string
	IsExpectedAt           string
	Unsure                 string
	CreatedBy              string
}

type expectedTimeOfArrivalRow struct {
	ID                     string `json:"id"`
	ExpectedAt             string `json:"expected_at,omitempty"`
	Note                   string `json:"note,omitempty"`
	Unsure                 bool   `json:"unsure"`
	TenderJobScheduleShift string `json:"tender_job_schedule_shift_id,omitempty"`
	JobScheduleShift       string `json:"job_schedule_shift_id,omitempty"`
	CreatedBy              string `json:"created_by_id,omitempty"`
}

func newExpectedTimeOfArrivalsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List expected time of arrivals",
		Long: `List expected time of arrival updates for tender job schedule shifts.

Output Columns:
  ID            Expected time of arrival identifier
  EXPECTED AT   Expected arrival timestamp
  UNSURE        Whether the arrival time is unsure
  NOTE          Notes for the arrival (truncated)
  TENDER SHIFT  Tender job schedule shift ID
  JOB SHIFT     Job schedule shift ID
  CREATED BY    User who created the record

Filters:
  --tender-job-schedule-shift  Filter by tender job schedule shift ID
  --job-schedule-shift         Filter by job schedule shift ID
  --expected-at-min            Filter by minimum expected arrival time (ISO 8601)
  --expected-at-max            Filter by maximum expected arrival time (ISO 8601)
  --is-expected-at             Filter by has expected arrival time (true/false)
  --unsure                     Filter by unsure status (true/false)
  --created-by                 Filter by created-by user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List expected time of arrivals
  xbe view expected-time-of-arrivals list

  # Filter by tender job schedule shift
  xbe view expected-time-of-arrivals list --tender-job-schedule-shift 123

  # Filter by expected arrival range
  xbe view expected-time-of-arrivals list --expected-at-min 2025-01-01T00:00:00Z --expected-at-max 2025-01-31T23:59:59Z

  # Filter unsure ETAs
  xbe view expected-time-of-arrivals list --unsure true

  # Output as JSON
  xbe view expected-time-of-arrivals list --json`,
		Args: cobra.NoArgs,
		RunE: runExpectedTimeOfArrivalsList,
	}
	initExpectedTimeOfArrivalsListFlags(cmd)
	return cmd
}

func init() {
	expectedTimeOfArrivalsCmd.AddCommand(newExpectedTimeOfArrivalsListCmd())
}

func initExpectedTimeOfArrivalsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("job-schedule-shift", "", "Filter by job schedule shift ID")
	cmd.Flags().String("expected-at-min", "", "Filter by minimum expected arrival time (ISO 8601)")
	cmd.Flags().String("expected-at-max", "", "Filter by maximum expected arrival time (ISO 8601)")
	cmd.Flags().String("is-expected-at", "", "Filter by has expected arrival time (true/false)")
	cmd.Flags().String("unsure", "", "Filter by unsure status (true/false)")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runExpectedTimeOfArrivalsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseExpectedTimeOfArrivalsListOptions(cmd)
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
	query.Set("fields[expected-time-of-arrivals]", "expected-at,note,unsure,tender-job-schedule-shift,job-schedule-shift,created-by")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[job_schedule_shift]", opts.JobScheduleShift)
	setFilterIfPresent(query, "filter[expected_at_min]", opts.ExpectedAtMin)
	setFilterIfPresent(query, "filter[expected_at_max]", opts.ExpectedAtMax)
	setFilterIfPresent(query, "filter[is_expected_at]", opts.IsExpectedAt)
	setFilterIfPresent(query, "filter[unsure]", opts.Unsure)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/expected-time-of-arrivals", query)
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

	rows := buildExpectedTimeOfArrivalRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderExpectedTimeOfArrivalsTable(cmd, rows)
}

func parseExpectedTimeOfArrivalsListOptions(cmd *cobra.Command) (expectedTimeOfArrivalsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	expectedAtMin, _ := cmd.Flags().GetString("expected-at-min")
	expectedAtMax, _ := cmd.Flags().GetString("expected-at-max")
	isExpectedAt, _ := cmd.Flags().GetString("is-expected-at")
	unsure, _ := cmd.Flags().GetString("unsure")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return expectedTimeOfArrivalsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TenderJobScheduleShift: tenderJobScheduleShift,
		JobScheduleShift:       jobScheduleShift,
		ExpectedAtMin:          expectedAtMin,
		ExpectedAtMax:          expectedAtMax,
		IsExpectedAt:           isExpectedAt,
		Unsure:                 unsure,
		CreatedBy:              createdBy,
	}, nil
}

func buildExpectedTimeOfArrivalRows(resp jsonAPIResponse) []expectedTimeOfArrivalRow {
	rows := make([]expectedTimeOfArrivalRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		row := expectedTimeOfArrivalRow{
			ID:         resource.ID,
			ExpectedAt: formatDateTime(stringAttr(resource.Attributes, "expected-at")),
			Note:       stringAttr(resource.Attributes, "note"),
			Unsure:     boolAttr(resource.Attributes, "unsure"),
		}

		if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
			row.TenderJobScheduleShift = rel.Data.ID
		}
		if rel, ok := resource.Relationships["job-schedule-shift"]; ok && rel.Data != nil {
			row.JobScheduleShift = rel.Data.ID
		}
		if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
			row.CreatedBy = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderExpectedTimeOfArrivalsTable(cmd *cobra.Command, rows []expectedTimeOfArrivalRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No expected time of arrivals found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tEXPECTED AT\tUNSURE\tNOTE\tTENDER SHIFT\tJOB SHIFT\tCREATED BY")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.ExpectedAt,
			formatBool(row.Unsure),
			truncateString(row.Note, 30),
			row.TenderJobScheduleShift,
			row.JobScheduleShift,
			row.CreatedBy,
		)
	}
	return writer.Flush()
}
