package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type siteWaitTimeNotificationTriggersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type siteWaitTimeNotificationTriggerDetails struct {
	ID                       string `json:"id"`
	SiteType                 string `json:"site_type,omitempty"`
	SiteID                   string `json:"site_id,omitempty"`
	SiteName                 string `json:"site_name,omitempty"`
	Latitude                 string `json:"latitude,omitempty"`
	Longitude                string `json:"longitude,omitempty"`
	EventAt                  string `json:"event_at,omitempty"`
	ActualMinutes            string `json:"actual_minutes,omitempty"`
	ThresholdMinutes         string `json:"threshold_minutes,omitempty"`
	TenderJobScheduleShiftID string `json:"tender_job_schedule_shift_id,omitempty"`
	JobProductionPlanID      string `json:"job_production_plan_id,omitempty"`
}

func newSiteWaitTimeNotificationTriggersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show site wait time notification trigger details",
		Long: `Show the full details of a site wait time notification trigger.

Output Fields:
  ID
  Site Type
  Site ID
  Site Name
  Latitude
  Longitude
  Event At
  Actual Minutes
  Threshold Minutes
  Tender Job Schedule Shift ID
  Job Production Plan ID

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The site wait time notification trigger ID (required). You can find IDs using the list command.`,
		Example: `  # Show a site wait time notification trigger
  xbe view site-wait-time-notification-triggers show 123

  # Get JSON output
  xbe view site-wait-time-notification-triggers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runSiteWaitTimeNotificationTriggersShow,
	}
	initSiteWaitTimeNotificationTriggersShowFlags(cmd)
	return cmd
}

func init() {
	siteWaitTimeNotificationTriggersCmd.AddCommand(newSiteWaitTimeNotificationTriggersShowCmd())
}

func initSiteWaitTimeNotificationTriggersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runSiteWaitTimeNotificationTriggersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseSiteWaitTimeNotificationTriggersShowOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if opts.NoAuth {
		opts.Token = ""
	} else if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if !errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("site wait time notification trigger id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[site-wait-time-notification-triggers]", "site-type,site-id,site-name,latitude,longitude,event-at,actual-minutes,threshold-minutes,tender-job-schedule-shift,job-production-plan")

	body, _, err := client.Get(cmd.Context(), "/v1/site-wait-time-notification-triggers/"+id, query)
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	handled, err := renderSparseShowIfRequested(cmd, resp)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}
	if handled {
		return nil
	}

	details := buildSiteWaitTimeNotificationTriggerDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderSiteWaitTimeNotificationTriggerDetails(cmd, details)
}

func parseSiteWaitTimeNotificationTriggersShowOptions(cmd *cobra.Command) (siteWaitTimeNotificationTriggersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return siteWaitTimeNotificationTriggersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildSiteWaitTimeNotificationTriggerDetails(resp jsonAPISingleResponse) siteWaitTimeNotificationTriggerDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := siteWaitTimeNotificationTriggerDetails{
		ID:               resource.ID,
		SiteType:         stringAttr(attrs, "site-type"),
		SiteID:           stringAttr(attrs, "site-id"),
		SiteName:         stringAttr(attrs, "site-name"),
		Latitude:         stringAttr(attrs, "latitude"),
		Longitude:        stringAttr(attrs, "longitude"),
		EventAt:          formatDateTime(stringAttr(attrs, "event-at")),
		ActualMinutes:    stringAttr(attrs, "actual-minutes"),
		ThresholdMinutes: stringAttr(attrs, "threshold-minutes"),
	}

	details.TenderJobScheduleShiftID = relationshipIDFromMap(resource.Relationships, "tender-job-schedule-shift")
	details.JobProductionPlanID = relationshipIDFromMap(resource.Relationships, "job-production-plan")

	return details
}

func renderSiteWaitTimeNotificationTriggerDetails(cmd *cobra.Command, details siteWaitTimeNotificationTriggerDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.SiteType != "" {
		fmt.Fprintf(out, "Site Type: %s\n", details.SiteType)
	}
	if details.SiteID != "" {
		fmt.Fprintf(out, "Site ID: %s\n", details.SiteID)
	}
	if details.SiteName != "" {
		fmt.Fprintf(out, "Site Name: %s\n", details.SiteName)
	}
	if details.Latitude != "" {
		fmt.Fprintf(out, "Latitude: %s\n", details.Latitude)
	}
	if details.Longitude != "" {
		fmt.Fprintf(out, "Longitude: %s\n", details.Longitude)
	}
	if details.EventAt != "" {
		fmt.Fprintf(out, "Event At: %s\n", details.EventAt)
	}
	if details.ActualMinutes != "" {
		fmt.Fprintf(out, "Actual Minutes: %s\n", details.ActualMinutes)
	}
	if details.ThresholdMinutes != "" {
		fmt.Fprintf(out, "Threshold Minutes: %s\n", details.ThresholdMinutes)
	}
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift ID: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan ID: %s\n", details.JobProductionPlanID)
	}

	return nil
}
