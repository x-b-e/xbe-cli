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

type driverDayTripsAdjustmentsListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	Status                 string
	DriverDay              string
	TenderJobScheduleShift string
	Trucker                string
	Broker                 string
	CreatedBy              string
}

type driverDayTripsAdjustmentRow struct {
	ID                     string `json:"id"`
	Status                 string `json:"status,omitempty"`
	Description            string `json:"description,omitempty"`
	DriverDayID            string `json:"driver_day_id,omitempty"`
	TenderJobScheduleShift string `json:"tender_job_schedule_shift_id,omitempty"`
	TruckerID              string `json:"trucker_id,omitempty"`
	BrokerID               string `json:"broker_id,omitempty"`
	CreatedByID            string `json:"created_by_id,omitempty"`
}

func newDriverDayTripsAdjustmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List driver day trips adjustments",
		Long: `List driver day trips adjustments.

Output Columns:
  ID            Adjustment identifier
  STATUS        Adjustment status
  DESCRIPTION   Adjustment description (truncated)
  SHIFT ID      Tender job schedule shift ID
  DRIVER DAY    Driver day ID
  TRUCKER       Trucker ID
  BROKER        Broker ID

Filters:
  --status                    Filter by status (e.g., editing)
  --driver-day                Filter by driver day ID
  --tender-job-schedule-shift Filter by tender job schedule shift ID
  --trucker                   Filter by trucker ID
  --broker                    Filter by broker ID
  --created-by                Filter by created-by user ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List adjustments
  xbe view driver-day-trips-adjustments list

  # Filter by status
  xbe view driver-day-trips-adjustments list --status editing

  # Filter by driver day
  xbe view driver-day-trips-adjustments list --driver-day 123

  # Output as JSON
  xbe view driver-day-trips-adjustments list --json`,
		Args: cobra.NoArgs,
		RunE: runDriverDayTripsAdjustmentsList,
	}
	initDriverDayTripsAdjustmentsListFlags(cmd)
	return cmd
}

func init() {
	driverDayTripsAdjustmentsCmd.AddCommand(newDriverDayTripsAdjustmentsListCmd())
}

func initDriverDayTripsAdjustmentsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("status", "", "Filter by status (e.g., editing)")
	cmd.Flags().String("driver-day", "", "Filter by driver day ID")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("trucker", "", "Filter by trucker ID")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("created-by", "", "Filter by created-by user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDriverDayTripsAdjustmentsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseDriverDayTripsAdjustmentsListOptions(cmd)
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

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[status]", opts.Status)
	setFilterIfPresent(query, "filter[driver_day]", opts.DriverDay)
	setFilterIfPresent(query, "filter[tender_job_schedule_shift]", opts.TenderJobScheduleShift)
	setFilterIfPresent(query, "filter[trucker]", opts.Trucker)
	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[created_by]", opts.CreatedBy)

	body, _, err := client.Get(cmd.Context(), "/v1/driver-day-trips-adjustments", query)
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

	rows := buildDriverDayTripsAdjustmentRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderDriverDayTripsAdjustmentsTable(cmd, rows)
}

func parseDriverDayTripsAdjustmentsListOptions(cmd *cobra.Command) (driverDayTripsAdjustmentsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	status, _ := cmd.Flags().GetString("status")
	driverDay, _ := cmd.Flags().GetString("driver-day")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	trucker, _ := cmd.Flags().GetString("trucker")
	broker, _ := cmd.Flags().GetString("broker")
	createdBy, _ := cmd.Flags().GetString("created-by")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return driverDayTripsAdjustmentsListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		Status:                 status,
		DriverDay:              driverDay,
		TenderJobScheduleShift: tenderJobScheduleShift,
		Trucker:                trucker,
		Broker:                 broker,
		CreatedBy:              createdBy,
	}, nil
}

func buildDriverDayTripsAdjustmentRows(resp jsonAPIResponse) []driverDayTripsAdjustmentRow {
	rows := make([]driverDayTripsAdjustmentRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildDriverDayTripsAdjustmentRow(resource))
	}
	return rows
}

func buildDriverDayTripsAdjustmentRow(resource jsonAPIResource) driverDayTripsAdjustmentRow {
	attrs := resource.Attributes
	row := driverDayTripsAdjustmentRow{
		ID:          resource.ID,
		Status:      stringAttr(attrs, "status"),
		Description: stringAttr(attrs, "description"),
	}

	if rel, ok := resource.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShift = rel.Data.ID
	}
	if rel, ok := resource.Relationships["driver-day"]; ok && rel.Data != nil {
		row.DriverDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["trucker"]; ok && rel.Data != nil {
		row.TruckerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		row.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["created-by"]; ok && rel.Data != nil {
		row.CreatedByID = rel.Data.ID
	}

	return row
}

func driverDayTripsAdjustmentRowFromSingle(resp jsonAPISingleResponse) driverDayTripsAdjustmentRow {
	return buildDriverDayTripsAdjustmentRow(resp.Data)
}

func renderDriverDayTripsAdjustmentsTable(cmd *cobra.Command, rows []driverDayTripsAdjustmentRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No driver day trips adjustments found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSTATUS\tDESCRIPTION\tSHIFT ID\tDRIVER DAY\tTRUCKER\tBROKER")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.Status,
			truncateString(row.Description, 40),
			row.TenderJobScheduleShift,
			row.DriverDayID,
			row.TruckerID,
			row.BrokerID,
		)
	}
	return writer.Flush()
}
