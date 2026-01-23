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

type hosEventsListOptions struct {
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

type hosEventRow struct {
	ID              string `json:"id"`
	EventType       string `json:"event_type,omitempty"`
	WorkStatus      string `json:"work_status,omitempty"`
	OccurredAt      string `json:"occurred_at,omitempty"`
	DurationSeconds *int   `json:"duration_seconds,omitempty"`
	UserID          string `json:"user_id,omitempty"`
	HosDayID        string `json:"hos_day_id,omitempty"`
	BrokerID        string `json:"broker_id,omitempty"`
}

func newHosEventsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List HOS events",
		Long: `List HOS (hours-of-service) events.

Output Columns:
  ID         HOS event identifier
  EVENT      Event type
  WORK       Work status
  OCCURRED   Event timestamp
  DURATION   Duration in seconds
  USER       User ID
  HOS DAY    HOS day ID
  BROKER     Broker ID

Filters:
  --broker   Filter by broker ID
  --hos-day  Filter by HOS day ID
  --user     Filter by user ID
  --driver   Filter by driver (alias for user) ID

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List HOS events
  xbe view hos-events list

  # Filter by driver
  xbe view hos-events list --driver 123

  # Filter by HOS day
  xbe view hos-events list --hos-day 456

  # JSON output
  xbe view hos-events list --json`,
		Args: cobra.NoArgs,
		RunE: runHosEventsList,
	}
	initHosEventsListFlags(cmd)
	return cmd
}

func init() {
	hosEventsCmd.AddCommand(newHosEventsListCmd())
}

func initHosEventsListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 0, "Limit results")
	cmd.Flags().Int("offset", 0, "Offset results")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("broker", "", "Filter by broker ID")
	cmd.Flags().String("hos-day", "", "Filter by HOS day ID")
	cmd.Flags().String("user", "", "Filter by user ID")
	cmd.Flags().String("driver", "", "Filter by driver (alias for user) ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosEventsList(cmd *cobra.Command, _ []string) error {
	opts, err := parseHosEventsListOptions(cmd)
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
	query.Set("fields[hos-events]", "event-type,work-status,occurred-at,start-at,duration-seconds")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if opts.Sort != "" {
		query.Set("sort", opts.Sort)
	}

	if opts.Driver != "" {
		if opts.User != "" && opts.User != opts.Driver {
			return fmt.Errorf("driver and user filters must match when both are set")
		}
		opts.User = opts.Driver
	}

	setFilterIfPresent(query, "filter[broker]", opts.Broker)
	setFilterIfPresent(query, "filter[hos_day]", opts.HosDay)
	setFilterIfPresent(query, "filter[user]", opts.User)

	body, _, err := client.Get(cmd.Context(), "/v1/hos-events", query)
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

	rows := buildHosEventRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderHosEventsTable(cmd, rows)
}

func parseHosEventsListOptions(cmd *cobra.Command) (hosEventsListOptions, error) {
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

	return hosEventsListOptions{
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

func buildHosEventRows(resp jsonAPIResponse) []hosEventRow {
	rows := make([]hosEventRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		occurredAt := parseTimeAttr(resource.Attributes, "occurred-at")
		if occurredAt == nil {
			occurredAt = parseTimeAttr(resource.Attributes, "start-at")
		}
		row := hosEventRow{
			ID:         resource.ID,
			EventType:  stringAttr(resource.Attributes, "event-type"),
			WorkStatus: stringAttr(resource.Attributes, "work-status"),
			OccurredAt: formatTimeValue(occurredAt),
		}

		if duration, ok := intAttrValue(resource.Attributes, "duration-seconds"); ok {
			row.DurationSeconds = &duration
		}

		if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
			row.UserID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
			row.HosDayID = rel.Data.ID
		}
		if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
			row.BrokerID = rel.Data.ID
		}

		rows = append(rows, row)
	}
	return rows
}

func renderHosEventsTable(cmd *cobra.Command, rows []hosEventRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No HOS events found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, 32, 0)
	fmt.Fprintln(writer, "ID\tEVENT\tWORK\tOCCURRED\tDURATION\tUSER\tHOS DAY\tBROKER")
	for _, row := range rows {
		duration := ""
		if row.DurationSeconds != nil {
			duration = strconv.Itoa(*row.DurationSeconds)
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.EventType,
			row.WorkStatus,
			row.OccurredAt,
			duration,
			row.UserID,
			row.HosDayID,
			row.BrokerID,
		)
	}
	return writer.Flush()
}

func intAttrValue(attrs map[string]any, key string) (int, bool) {
	if attrs == nil {
		return 0, false
	}
	value, ok := attrs[key]
	if !ok || value == nil {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case string:
		if strings.TrimSpace(typed) == "" {
			return 0, false
		}
		var result int
		if _, err := fmt.Sscanf(typed, "%d", &result); err == nil {
			return result, true
		}
	}
	return 0, false
}
