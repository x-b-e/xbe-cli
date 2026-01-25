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

type hosAvailabilitySnapshotsListOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
	Limit   int
	Offset  int
	Sort    string
	Broker  string
	HosDay  string
	User    string
	Driver  string
}

type hosAvailabilitySnapshotRow struct {
	ID                         string `json:"id"`
	CapturedAt                 string `json:"captured_at,omitempty"`
	WorkStatus                 string `json:"work_status,omitempty"`
	IsAvailable                bool   `json:"is_available"`
	WorkdaySecondsRemaining    string `json:"workday_seconds_remaining,omitempty"`
	DrivingSecondsRemaining    string `json:"driving_seconds_remaining,omitempty"`
	DutySecondsRemaining       string `json:"duty_seconds_remaining,omitempty"`
	CycleSecondsRemainingToday string `json:"cycle_seconds_remaining_today,omitempty"`
	BreakSecondsRemaining      string `json:"break_seconds_remaining,omitempty"`
	BrokerID                   string `json:"broker_id,omitempty"`
	HosDayID                   string `json:"hos_day_id,omitempty"`
	UserID                     string `json:"user_id,omitempty"`
}

func newHosAvailabilitySnapshotsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HOS availability snapshots",
		Long: `List HOS availability snapshots with filtering and pagination.

HOS availability snapshots capture remaining hours-of-service availability
for a driver at a point in time.

Output Columns:
  ID         Snapshot identifier
  DRIVER     Driver user ID
  HOS DAY    HOS day ID
  CAPTURED   Snapshot timestamp
  STATUS     Work status at capture time
  AVAILABLE  Whether the driver is available
  WORKDAY    Workday seconds remaining
  DRIVING    Driving seconds remaining
  DUTY       Duty seconds remaining
  CYCLE      Cycle seconds remaining (today)
  BREAK      Break seconds remaining

Filters:
  --broker   Filter by broker ID
  --hos-day  Filter by HOS day ID
  --user     Filter by user ID
  --driver   Filter by driver user ID (alias for user)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List snapshots
  xbe view hos-availability-snapshots list

  # Filter by driver
  xbe view hos-availability-snapshots list --driver 123

  # Filter by broker
  xbe view hos-availability-snapshots list --broker 456

  # Output as JSON
  xbe view hos-availability-snapshots list --json`,
		Args: cobra.NoArgs,
		RunE: runHosAvailabilitySnapshotsList,
	}
	initHosAvailabilitySnapshotsListFlags(cmd)
	return cmd
}

func init() {
	hosAvailabilitySnapshotsCmd.AddCommand(newHosAvailabilitySnapshotsListCmd())
}

func initHosAvailabilitySnapshotsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("hos-day", "", "Filter by HOS day ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("driver", "", "Filter by driver user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosAvailabilitySnapshotsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHosAvailabilitySnapshotsListOptions(cmd)
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
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run \"xbe auth login\" first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[hos-availability-snapshots]", "captured-at,work-status,is-available,workday-seconds-remaining,driving-seconds-remaining,duty-seconds-remaining,cycle-seconds-remaining-today,break-seconds-remaining")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[hos_day]", opts.HosDay)
	setFilterIfPresent(query, "filter[user]", opts.User)
	setFilterIfPresent(query, "filter[driver]", opts.Driver)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-availability-snapshots", query)
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

	rows := buildHosAvailabilitySnapshotRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHosAvailabilitySnapshotsTable(cmd, rows)
}

func parseHosAvailabilitySnapshotsListOptions(cmd *cobra.Command) (hosAvailabilitySnapshotsListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	broker, _ := cmd.Flags().GetString("broker")
	hosDay, _ := cmd.Flags().GetString("hos-day")
	user, _ := cmd.Flags().GetString("user")
	driver, _ := cmd.Flags().GetString("driver")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosAvailabilitySnapshotsListOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
		Limit:   limit,
		Offset:  offset,
		Sort:    sort,
		Broker:  broker,
		HosDay:  hosDay,
		User:    user,
		Driver:  driver,
	}, nil
}

func buildHosAvailabilitySnapshotRows(resp jsonAPIResponse) []hosAvailabilitySnapshotRow {
	rows := make([]hosAvailabilitySnapshotRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := hosAvailabilitySnapshotRow{
			ID:                         resource.ID,
			CapturedAt:                 formatDateTime(stringAttr(attrs, "captured-at")),
			WorkStatus:                 stringAttr(attrs, "work-status"),
			IsAvailable:                boolAttr(attrs, "is-available"),
			WorkdaySecondsRemaining:    stringAttr(attrs, "workday-seconds-remaining"),
			DrivingSecondsRemaining:    stringAttr(attrs, "driving-seconds-remaining"),
			DutySecondsRemaining:       stringAttr(attrs, "duty-seconds-remaining"),
			CycleSecondsRemainingToday: stringAttr(attrs, "cycle-seconds-remaining-today"),
			BreakSecondsRemaining:      stringAttr(attrs, "break-seconds-remaining"),
		}

		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
			row.HosDayID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderHosAvailabilitySnapshotsTable(cmd *cobra.Command, rows []hosAvailabilitySnapshotRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No HOS availability snapshots found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDRIVER\tHOS DAY\tCAPTURED\tSTATUS\tAVAILABLE\tWORKDAY\tDRIVING\tDUTY\tCYCLE\tBREAK")
	for _, row := range rows {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.UserID,
			row.HosDayID,
			row.CapturedAt,
			row.WorkStatus,
			formatYesNo(row.IsAvailable),
			row.WorkdaySecondsRemaining,
			row.DrivingSecondsRemaining,
			row.DutySecondsRemaining,
			row.CycleSecondsRemainingToday,
			row.BreakSecondsRemaining,
		)
	}
	return writer.Flush()
}
