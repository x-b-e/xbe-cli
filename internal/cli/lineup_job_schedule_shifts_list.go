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

type lineupJobScheduleShiftsListOptions struct {
	BaseURL                             string
	Token                               string
	JSON                                bool
	NoAuth                              bool
	Limit                               int
	Offset                              int
	Sort                                string
	Lineup                              string
	JobScheduleShift                    string
	Trucker                             string
	TrailerClassification               string
	IsReadyToDispatch                   string
	WithoutLineupDispatchShift          string
	TrailerClassificationEquivalentType string
}

type lineupJobScheduleShiftRow struct {
	ID                                  string `json:"id"`
	LineupID                            string `json:"lineup_id,omitempty"`
	JobScheduleShiftID                  string `json:"job_schedule_shift_id,omitempty"`
	TruckerID                           string `json:"trucker_id,omitempty"`
	DriverID                            string `json:"driver_id,omitempty"`
	TrailerClassificationID             string `json:"trailer_classification_id,omitempty"`
	TrailerClassificationEquivalentType string `json:"trailer_classification_equivalent_type,omitempty"`
	IsBrokered                          bool   `json:"is_brokered"`
	IsReadyToDispatch                   bool   `json:"is_ready_to_dispatch"`
	HasLineupDispatchShift              bool   `json:"has_lineup_dispatch_shift"`
	ExcludeFromLineupScenarios          bool   `json:"exclude_from_lineup_scenarios"`
	IsExpectingTimeCard                 bool   `json:"is_expecting_time_card"`
}

func newLineupJobScheduleShiftsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lineup job schedule shifts",
		Long: `List lineup job schedule shifts with filtering and pagination.

Output Columns:
  ID           Lineup job schedule shift ID
  LINEUP       Lineup ID
  JOB_SHIFT    Job schedule shift ID
  TRUCKER      Trucker ID
  TRAILER      Trailer classification ID
  READY        Ready to dispatch (yes/no)
  BROKERED     Brokered shift (yes/no)
  DISPATCH     Has lineup dispatch shift (yes/no)

Filters:
  --lineup                                 Filter by lineup ID
  --job-schedule-shift                     Filter by job schedule shift ID
  --trucker                                Filter by trucker ID
  --trailer-classification                 Filter by trailer classification ID
  --is-ready-to-dispatch                   Filter by ready to dispatch (true/false)
  --without-lineup-dispatch-shift          Filter by missing lineup dispatch shift (true/false)
  --trailer-classification-equivalent-type Filter by trailer classification equivalent type

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List lineup job schedule shifts
  xbe view lineup-job-schedule-shifts list

  # Filter by lineup
  xbe view lineup-job-schedule-shifts list --lineup 123

  # Filter by ready to dispatch
  xbe view lineup-job-schedule-shifts list --is-ready-to-dispatch true

  # Output as JSON
  xbe view lineup-job-schedule-shifts list --json`,
		Args: cobra.NoArgs,
		RunE: runLineupJobScheduleShiftsList,
	}
	initLineupJobScheduleShiftsListFlags(cmd)
	return cmd
}

func init() {
	lineupJobScheduleShiftsCmd.AddCommand(newLineupJobScheduleShiftsListCmd())
}

func initLineupJobScheduleShiftsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("lineup", "", "Filter by lineup ID")
	cmd.Flags().String("job-schedule-shift", "", "Filter by job schedule shift ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("trailer-classification", "", "Filter by trailer classification ID")
	cmd.Flags().String("is-ready-to-dispatch", "", "Filter by ready to dispatch (true/false)")
	cmd.Flags().String("without-lineup-dispatch-shift", "", "Filter by missing lineup dispatch shift (true/false)")
	cmd.Flags().String("trailer-classification-equivalent-type", "", "Filter by trailer classification equivalent type")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupJobScheduleShiftsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseLineupJobScheduleShiftsListOptions(cmd)
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
	query.Set("fields[lineup-job-schedule-shifts]", strings.Join([]string{
		"trailer-classification-equivalent-type",
		"is-brokered",
		"is-ready-to-dispatch",
		"exclude-from-lineup-scenarios",
		"is-expecting-time-card",
		"has-lineup-dispatch-shift",
		"lineup",
		"job-schedule-shift",
		"trucker",
		"driver",
		"trailer-classification",
	}, ","))

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[lineup]", opts.Lineup)
	setFilterIfPresent(query, "filter[job-schedule-shift]", opts.JobScheduleShift)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[trailer-classification]", opts.TrailerClassification)
	setFilterIfPresent(query, "filter[is-ready-to-dispatch]", opts.IsReadyToDispatch)
	setFilterIfPresent(query, "filter[without-lineup-dispatch-shift]", opts.WithoutLineupDispatchShift)
	setFilterIfPresent(query, "filter[trailer-classification-equivalent-type]", opts.TrailerClassificationEquivalentType)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-job-schedule-shifts", query)
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

	rows := buildLineupJobScheduleShiftRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderLineupJobScheduleShiftsTable(cmd, rows)
}

func parseLineupJobScheduleShiftsListOptions(cmd *cobra.Command) (lineupJobScheduleShiftsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	lineup, _ := cmd.Flags().GetString("lineup")
	jobScheduleShift, _ := cmd.Flags().GetString("job-schedule-shift")
	trucker, _ := cmd.Flags().GetString("trucker")
	trailerClassification, _ := cmd.Flags().GetString("trailer-classification")
	isReadyToDispatch, _ := cmd.Flags().GetString("is-ready-to-dispatch")
	withoutLineupDispatchShift, _ := cmd.Flags().GetString("without-lineup-dispatch-shift")
	trailerClassificationEquivalentType, _ := cmd.Flags().GetString("trailer-classification-equivalent-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupJobScheduleShiftsListOptions{
		BaseURL:                             baseURL,
		Token:                               token,
		JSON:                                jsonOut,
		NoAuth:                              noAuth,
		Limit:                               limit,
		Offset:                              offset,
		Sort:                                sort,
		Lineup:                              lineup,
		JobScheduleShift:                    jobScheduleShift,
		Trucker:                             trucker,
		TrailerClassification:               trailerClassification,
		IsReadyToDispatch:                   isReadyToDispatch,
		WithoutLineupDispatchShift:          withoutLineupDispatchShift,
		TrailerClassificationEquivalentType: trailerClassificationEquivalentType,
	}, nil
}

func buildLineupJobScheduleShiftRows(resp jsonAPIResponse) []lineupJobScheduleShiftRow {
	rows := make([]lineupJobScheduleShiftRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := lineupJobScheduleShiftRow{
			ID:                                  resource.ID,
			TrailerClassificationEquivalentType: stringAttr(attrs, "trailer-classification-equivalent-type"),
			IsBrokered:                          boolAttr(attrs, "is-brokered"),
			IsReadyToDispatch:                   boolAttr(attrs, "is-ready-to-dispatch"),
			HasLineupDispatchShift:              boolAttr(attrs, "has-lineup-dispatch-shift"),
			ExcludeFromLineupScenarios:          boolAttr(attrs, "exclude-from-lineup-scenarios"),
			IsExpectingTimeCard:                 boolAttr(attrs, "is-expecting-time-card"),
		}

		row.LineupID = relationshipIDFromMap(resource.Relationships, "lineup")
		row.JobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "job-schedule-shift")
		row.TruckerID = relationshipIDFromMap(resource.Relationships, "trucker")
		row.DriverID = relationshipIDFromMap(resource.Relationships, "driver")
		row.TrailerClassificationID = relationshipIDFromMap(resource.Relationships, "trailer-classification")

		rows = append(rows, row)
	}
	return rows
}

func renderLineupJobScheduleShiftsTable(cmd *cobra.Command, rows []lineupJobScheduleShiftRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No lineup job schedule shifts found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tLINEUP\tJOB_SHIFT\tTRUCKER\tTRAILER\tREADY\tBROKERED\tDISPATCH")
	for _, row := range rows {
		ready := "no"
		if row.IsReadyToDispatch {
			ready = "yes"
		}
		brokered := "no"
		if row.IsBrokered {
			brokered = "yes"
		}
		dispatch := "no"
		if row.HasLineupDispatchShift {
			dispatch = "yes"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.LineupID,
			row.JobScheduleShiftID,
			row.TruckerID,
			row.TrailerClassificationID,
			ready,
			brokered,
			dispatch,
		)
	}
	return writer.Flush()
}
