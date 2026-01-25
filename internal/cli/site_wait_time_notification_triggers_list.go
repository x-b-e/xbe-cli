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

type siteWaitTimeNotificationTriggersListOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	NoAuth                 bool
	Limit                  int
	Offset                 int
	Sort                   string
	TenderJobScheduleShift string
	JobProductionPlan      string
	SiteType               string
	SiteID                 string
	EventAtMin             string
	EventAtMax             string
	IsEventAt              string
	CreatedAtMin           string
	CreatedAtMax           string
	UpdatedAtMin           string
	UpdatedAtMax           string
	IsCreatedAt            string
	IsUpdatedAt            string
}

type siteWaitTimeNotificationTriggerRow struct {
	ID                       string `json:"id"`
	SiteType                 string `json:"site_type,omitempty"`
	SiteID                   string `json:"site_id,omitempty"`
	SiteName                 string `json:"site_name,omitempty"`
	EventAt                  string `json:"event_at,omitempty"`
	ActualMinutes            string `json:"actual_minutes,omitempty"`
	ThresholdMinutes         string `json:"threshold_minutes,omitempty"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	JobProductionPlanID      string `json:"job_production_plan_id,omitempty"`
}

func newSiteWaitTimeNotificationTriggersListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List site wait time notification triggers",
		Long: `List site wait time notification triggers with filtering and pagination.

Site wait time notification triggers record excessive wait time events for job sites
and material sites that generate notifications.

Output Columns:
  ID           Trigger identifier
  SITE TYPE    Site type (job_site/material_site)
  SITE ID      Site ID
  SITE NAME    Site name
  EVENT AT     When the wait time was recorded
  ACTUAL MIN   Actual wait time in minutes
  THRESHOLD    Threshold minutes that triggered the notification
  SHIFT ID     Tender job schedule shift ID
  JOB PLAN     Job production plan ID

Filters:
  --tender-job-schedule-shift  Filter by tender job schedule shift ID
  --job-production-plan        Filter by job production plan ID
  --site-type                  Filter by site type (job_site/material_site)
  --site-id                    Filter by site ID
  --event-at-min               Filter by event-at on/after (ISO 8601)
  --event-at-max               Filter by event-at on/before (ISO 8601)
  --is-event-at                Filter by presence of event-at (true/false)
  --created-at-min             Filter by created-at on/after (ISO 8601)
  --created-at-max             Filter by created-at on/before (ISO 8601)
  --updated-at-min             Filter by updated-at on/after (ISO 8601)
  --updated-at-max             Filter by updated-at on/before (ISO 8601)
  --is-created-at              Filter by presence of created-at (true/false)
  --is-updated-at              Filter by presence of updated-at (true/false)

Global flags (see xbe --help): --json, --limit, --offset, --sort, --base-url, --token, --no-auth`,
		Example: `  # List site wait time notification triggers
  xbe view site-wait-time-notification-triggers list

  # Filter by job production plan
  xbe view site-wait-time-notification-triggers list --job-production-plan 123

  # Filter by site type
  xbe view site-wait-time-notification-triggers list --site-type job_site

  # Filter by event time range
  xbe view site-wait-time-notification-triggers list \
    --event-at-min 2025-01-01T00:00:00Z --event-at-max 2025-01-31T23:59:59Z

  # Output as JSON
  xbe view site-wait-time-notification-triggers list --json`,
		Args: cobra.NoArgs,
		RunE: runSiteWaitTimeNotificationTriggersList,
	}
	initSiteWaitTimeNotificationTriggersListFlags(cmd)
	return cmd
}

func init() {
	siteWaitTimeNotificationTriggersCmd.AddCommand(newSiteWaitTimeNotificationTriggersListCmd())
}

func initSiteWaitTimeNotificationTriggersListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().Int("limit", 50, "Page size")
	cmd.Flags().Int("offset", 0, "Page offset")
	cmd.Flags().String("sort", "", "Sort by field")
	cmd.Flags().String("tender-job-schedule-shift", "", "Filter by tender job schedule shift ID")
	cmd.Flags().String("job-production-plan", "", "Filter by job production plan ID")
	cmd.Flags().String("site-type", "", "Filter by site type (job_site/material_site)")
	cmd.Flags().String("site-id", "", "Filter by site ID")
	cmd.Flags().String("event-at-min", "", "Filter by event-at on/after (ISO 8601)")
	cmd.Flags().String("event-at-max", "", "Filter by event-at on/before (ISO 8601)")
	cmd.Flags().String("is-event-at", "", "Filter by presence of event-at (true/false)")
	cmd.Flags().String("created-at-min", "", "Filter by created-at on/after (ISO 8601)")
	cmd.Flags().String("created-at-max", "", "Filter by created-at on/before (ISO 8601)")
	cmd.Flags().String("updated-at-min", "", "Filter by updated-at on/after (ISO 8601)")
	cmd.Flags().String("updated-at-max", "", "Filter by updated-at on/before (ISO 8601)")
	cmd.Flags().String("is-created-at", "", "Filter by presence of created-at (true/false)")
	cmd.Flags().String("is-updated-at", "", "Filter by presence of updated-at (true/false)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSiteWaitTimeNotificationTriggersList(cmd *cobra.Command, _ []string) error {
	opts, err := parseSiteWaitTimeNotificationTriggersListOptions(cmd)
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
	query.Set("fields[site-wait-time-notification-triggers]", "site-type,site-id,site-name,event-at,actual-minutes,threshold-minutes,tender-job-schedule-shift,job-production-plan")

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
	setFilterIfPresent(query, "filter[job_production_plan]", opts.JobProductionPlan)
	setFilterIfPresent(query, "filter[site_type]", opts.SiteType)
	setFilterIfPresent(query, "filter[site_id]", opts.SiteID)
	setFilterIfPresent(query, "filter[event-at-min]", opts.EventAtMin)
	setFilterIfPresent(query, "filter[event-at-max]", opts.EventAtMax)
	setFilterIfPresent(query, "filter[is-event-at]", opts.IsEventAt)
	setFilterIfPresent(query, "filter[created-at-min]", opts.CreatedAtMin)
	setFilterIfPresent(query, "filter[created-at-max]", opts.CreatedAtMax)
	setFilterIfPresent(query, "filter[updated-at-min]", opts.UpdatedAtMin)
	setFilterIfPresent(query, "filter[updated-at-max]", opts.UpdatedAtMax)
	setFilterIfPresent(query, "filter[is-created-at]", opts.IsCreatedAt)
	setFilterIfPresent(query, "filter[is-updated-at]", opts.IsUpdatedAt)

	body, _, err := client.Get(cmd.Context(), "/v1/site-wait-time-notification-triggers", query)
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

	rows := buildSiteWaitTimeNotificationTriggerRows(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), rows)
	}

	return renderSiteWaitTimeNotificationTriggersTable(cmd, rows)
}

func parseSiteWaitTimeNotificationTriggersListOptions(cmd *cobra.Command) (siteWaitTimeNotificationTriggersListOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	sort, _ := cmd.Flags().GetString("sort")
	shift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	jobProductionPlan, _ := cmd.Flags().GetString("job-production-plan")
	siteType, _ := cmd.Flags().GetString("site-type")
	siteID, _ := cmd.Flags().GetString("site-id")
	eventAtMin, _ := cmd.Flags().GetString("event-at-min")
	eventAtMax, _ := cmd.Flags().GetString("event-at-max")
	isEventAt, _ := cmd.Flags().GetString("is-event-at")
	createdAtMin, _ := cmd.Flags().GetString("created-at-min")
	createdAtMax, _ := cmd.Flags().GetString("created-at-max")
	updatedAtMin, _ := cmd.Flags().GetString("updated-at-min")
	updatedAtMax, _ := cmd.Flags().GetString("updated-at-max")
	isCreatedAt, _ := cmd.Flags().GetString("is-created-at")
	isUpdatedAt, _ := cmd.Flags().GetString("is-updated-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return siteWaitTimeNotificationTriggersListOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		NoAuth:                 noAuth,
		Limit:                  limit,
		Offset:                 offset,
		Sort:                   sort,
		TenderJobScheduleShift: shift,
		JobProductionPlan:      jobProductionPlan,
		SiteType:               siteType,
		SiteID:                 siteID,
		EventAtMin:             eventAtMin,
		EventAtMax:             eventAtMax,
		IsEventAt:              isEventAt,
		CreatedAtMin:           createdAtMin,
		CreatedAtMax:           createdAtMax,
		UpdatedAtMin:           updatedAtMin,
		UpdatedAtMax:           updatedAtMax,
		IsCreatedAt:            isCreatedAt,
		IsUpdatedAt:            isUpdatedAt,
	}, nil
}

func buildSiteWaitTimeNotificationTriggerRows(resp jsonAPIResponse) []siteWaitTimeNotificationTriggerRow {
	rows := make([]siteWaitTimeNotificationTriggerRow, 0, len(resp.Data))
	for _, resource := range resp.Data {
		attrs := resource.Attributes
		row := siteWaitTimeNotificationTriggerRow{
			ID:               resource.ID,
			SiteType:         stringAttr(attrs, "site-type"),
			SiteID:           stringAttr(attrs, "site-id"),
			SiteName:         stringAttr(attrs, "site-name"),
			EventAt:          formatDateTime(stringAttr(attrs, "event-at")),
			ActualMinutes:    stringAttr(attrs, "actual-minutes"),
			ThresholdMinutes: stringAttr(attrs, "threshold-minutes"),
		}

		row.TenderJobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "tender-job-schedule-shift")
		row.JobProductionPlanID = relationshipIDFromMap(resource.Relationships, "job-production-plan")

		rows = append(rows, row)
	}
	return rows
}

func renderSiteWaitTimeNotificationTriggersTable(cmd *cobra.Command, rows []siteWaitTimeNotificationTriggerRow) error {
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No site wait time notification triggers found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tSITE_TYPE\tSITE_ID\tSITE_NAME\tEVENT_AT\tACTUAL_MIN\tTHRESHOLD\tSHIFT_ID\tJOB_PLAN")
	for _, row := range rows {
		siteName := truncateString(row.SiteName, 28)
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.SiteType,
			row.SiteID,
			siteName,
			row.EventAt,
			row.ActualMinutes,
			row.ThresholdMinutes,
			row.TenderJobScheduleShiftID,
			row.JobProductionPlanID,
		)
	}
	return writer.Flush()
}
