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

type jobProductionPlanAlarmSubscribersListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	JobProductionPlanAlarm string
	Subscriber             string
}

type jobProductionPlanAlarmSubscriberRow struct {
	ID                       string `json:"id"`
	JobProductionPlanAlarmID string `json:"job_production_plan_alarm_id,omitempty"`
	JobProductionPlanID      string `json:"job_production_plan_id,omitempty"`
	SubscriberID             string `json:"subscriber_id,omitempty"`
	SubscriberName           string `json:"subscriber_name,omitempty"`
	SubscriberEmail          string `json:"subscriber_email,omitempty"`
}

func newJobProductionPlanAlarmSubscribersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List job production plan alarm subscribers",
		Long: `List job production plan alarm subscribers with filtering and pagination.

Output Columns:
  ID          Subscriber identifier
  ALARM       Job production plan alarm ID
  PLAN        Job production plan ID (via alarm)
  SUBSCRIBER  Subscriber name/email

Filters:
  --job-production-plan-alarm  Filter by job production plan alarm ID
  --subscriber                 Filter by subscriber user ID

Sorting:
  Use --sort to specify sort order. Prefix with - for descending.

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List alarm subscribers
  xbe view job-production-plan-alarm-subscribers list

  # Filter by alarm
  xbe view job-production-plan-alarm-subscribers list --job-production-plan-alarm 123

  # Filter by subscriber
  xbe view job-production-plan-alarm-subscribers list --subscriber 456

  # Output as JSON
  xbe view job-production-plan-alarm-subscribers list --json`,
		Args: cobra.NoArgs,
		RunE: runJobProductionPlanAlarmSubscribersList,
	}
	initJobProductionPlanAlarmSubscribersListFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanAlarmSubscribersCmd.AddCommand(newJobProductionPlanAlarmSubscribersListCmd())
}

func initJobProductionPlanAlarmSubscribersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("job-production-plan-alarm", "", "Filter by job production plan alarm ID")
	cmd.Flags().String("subscriber", "", "Filter by subscriber user ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanAlarmSubscribersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseJobProductionPlanAlarmSubscribersListOptions(cmd)
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
	query.Set("fields[job-production-plan-alarm-subscribers]", "job-production-plan-alarm,subscriber")
	query.Set("include", "job-production-plan-alarm,subscriber")
	query.Set("fields[job-production-plan-alarms]", "job-production-plan")
	query.Set("fields[users]", "name,email-address")

	if opts.Limit > 0 {
		query.Set("page[limit]", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		query.Set("page[offset]", strconv.Itoa(opts.Offset))
	}
	if strings.TrimSpace(opts.Sort) != "" {
		query.Set("sort", opts.Sort)
	}
	setFilterIfPresent(query, "filter[job-production-plan-alarm]", opts.JobProductionPlanAlarm)
	setFilterIfPresent(query, "filter[subscriber]", opts.Subscriber)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-alarm-subscribers", query)
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

	rows := buildJobProductionPlanAlarmSubscriberRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderJobProductionPlanAlarmSubscribersTable(cmd, rows)
}

func parseJobProductionPlanAlarmSubscribersListOptions(cmd *cobra.Command) (jobProductionPlanAlarmSubscribersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	jobProductionPlanAlarm, _ := cmd.Flags().GetString("job-production-plan-alarm")
	subscriber, _ := cmd.Flags().GetString("subscriber")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanAlarmSubscribersListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		JobProductionPlanAlarm: jobProductionPlanAlarm,
		Subscriber:             subscriber,
	}, nil
}

func buildJobProductionPlanAlarmSubscriberRows(resp jsonAPIResponse) []jobProductionPlanAlarmSubscriberRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	rows := make([]jobProductionPlanAlarmSubscriberRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		rows = append(rows, buildJobProductionPlanAlarmSubscriberRow(resource, included))
	}
	return rows
}

func jobProductionPlanAlarmSubscriberRowFromSingle(resp jsonAPISingleResponse) jobProductionPlanAlarmSubscriberRow {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}
	return buildJobProductionPlanAlarmSubscriberRow(resp.Data, included)
}

func buildJobProductionPlanAlarmSubscriberRow(resource jsonAPIResource, included map[string]jsonAPIResource) jobProductionPlanAlarmSubscriberRow {
	row := jobProductionPlanAlarmSubscriberRow{
		ID: resource.ID,
	}

	if rel, ok := resource.Relationships["job-production-plan-alarm"]; ok && rel.Data != nil {
		row.JobProductionPlanAlarmID = rel.Data.ID
		if alarm, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			if alarmRel, ok := alarm.Relationships["job-production-plan"]; ok && alarmRel.Data != nil {
				row.JobProductionPlanID = alarmRel.Data.ID
			}
		}
	}

	if rel, ok := resource.Relationships["subscriber"]; ok && rel.Data != nil {
		row.SubscriberID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			row.SubscriberName = stringAttr(user.Attributes, "name")
			row.SubscriberEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return row
}

func renderJobProductionPlanAlarmSubscribersTable(cmd *cobra.Command, rows []jobProductionPlanAlarmSubscriberRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No job production plan alarm subscribers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tALARM\tPLAN\tSUBSCRIBER")
	for _, row := range rows {
		subscriber := firstNonEmpty(row.SubscriberName, row.SubscriberEmail, row.SubscriberID)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			row.ID,
			truncateString(row.JobProductionPlanAlarmID, 16),
			truncateString(row.JobProductionPlanID, 16),
			truncateString(subscriber, 24),
		)
	}
	return writer.Flush()
}
