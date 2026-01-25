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

type jobProductionPlanAlarmSubscribersShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanAlarmSubscriberDetails struct {
	ID                          string `json:"id"`
	JobProductionPlanAlarmID    string `json:"job_production_plan_alarm_id,omitempty"`
	JobProductionPlanID         string `json:"job_production_plan_id,omitempty"`
	SubscriberID                string `json:"subscriber_id,omitempty"`
	SubscriberName              string `json:"subscriber_name,omitempty"`
	SubscriberEmail             string `json:"subscriber_email,omitempty"`
	AlarmTons                   any    `json:"alarm_tons,omitempty"`
	AlarmBaseMaterialType       string `json:"alarm_base_material_type,omitempty"`
	AlarmMaxLatencyMinutes      any    `json:"alarm_max_latency_minutes,omitempty"`
	AlarmNote                   string `json:"alarm_note,omitempty"`
	AlarmPlannedAt              string `json:"alarm_planned_at,omitempty"`
	AlarmFulfilledAt            string `json:"alarm_fulfilled_at,omitempty"`
	AlarmFulfilledTransactionAt string `json:"alarm_fulfilled_transaction_at,omitempty"`
	AlarmPlanVarianceMinutes    any    `json:"alarm_plan_variance_minutes,omitempty"`
	AlarmGoodEnoughTimeZoneID   string `json:"alarm_good_enough_time_zone_id,omitempty"`
	AlarmCanUpdate              bool   `json:"alarm_can_update"`
	AlarmCanDestroy             bool   `json:"alarm_can_destroy"`
}

func newJobProductionPlanAlarmSubscribersShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan alarm subscriber details",
		Long: `Show the full details of a job production plan alarm subscriber.

Includes subscriber information plus the associated alarm details.

Arguments:
  <id>  The job production plan alarm subscriber ID (required).`,
		Example: `  # Show an alarm subscriber
  xbe view job-production-plan-alarm-subscribers show 123

  # Output as JSON
  xbe view job-production-plan-alarm-subscribers show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanAlarmSubscribersShow,
	}
	initJobProductionPlanAlarmSubscribersShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanAlarmSubscribersCmd.AddCommand(newJobProductionPlanAlarmSubscribersShowCmd())
}

func initJobProductionPlanAlarmSubscribersShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanAlarmSubscribersShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanAlarmSubscribersShowOptions(cmd)
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
		return fmt.Errorf("job production plan alarm subscriber id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[job-production-plan-alarm-subscribers]", "job-production-plan-alarm,subscriber")
	query.Set("include", "job-production-plan-alarm,subscriber")
	query.Set("fields[job-production-plan-alarms]", "job-production-plan,tons,base-material-type-fully-qualified-name,max-latency-minutes,note,planned-at,fulfilled-at,good-enough-time-zone-id,fulfilled-transaction-at,plan-variance-minutes,can-update,can-destroy")
	query.Set("fields[users]", "name,email-address")

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-alarm-subscribers/"+id, query)
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

	details := buildJobProductionPlanAlarmSubscriberDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanAlarmSubscriberDetails(cmd, details)
}

func parseJobProductionPlanAlarmSubscribersShowOptions(cmd *cobra.Command) (jobProductionPlanAlarmSubscribersShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanAlarmSubscribersShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanAlarmSubscriberDetails(resp jsonAPISingleResponse) jobProductionPlanAlarmSubscriberDetails {
	included := make(map[string]jsonAPIResource, len(resp.Included))
	for _, inc := range resp.Included {
		included[resourceKey(inc.Type, inc.ID)] = inc
	}

	details := jobProductionPlanAlarmSubscriberDetails{
		ID: resp.Data.ID,
	}

	if rel, ok := resp.Data.Relationships["job-production-plan-alarm"]; ok && rel.Data != nil {
		details.JobProductionPlanAlarmID = rel.Data.ID
		if alarm, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			attrs := alarm.Attributes
			details.AlarmTons = attrs["tons"]
			details.AlarmBaseMaterialType = stringAttr(attrs, "base-material-type-fully-qualified-name")
			details.AlarmMaxLatencyMinutes = attrs["max-latency-minutes"]
			details.AlarmNote = stringAttr(attrs, "note")
			details.AlarmPlannedAt = stringAttr(attrs, "planned-at")
			details.AlarmFulfilledAt = stringAttr(attrs, "fulfilled-at")
			details.AlarmFulfilledTransactionAt = stringAttr(attrs, "fulfilled-transaction-at")
			details.AlarmPlanVarianceMinutes = attrs["plan-variance-minutes"]
			details.AlarmGoodEnoughTimeZoneID = stringAttr(attrs, "good-enough-time-zone-id")
			details.AlarmCanUpdate = boolAttr(attrs, "can-update")
			details.AlarmCanDestroy = boolAttr(attrs, "can-destroy")

			if alarmRel, ok := alarm.Relationships["job-production-plan"]; ok && alarmRel.Data != nil {
				details.JobProductionPlanID = alarmRel.Data.ID
			}
		}
	}

	if rel, ok := resp.Data.Relationships["subscriber"]; ok && rel.Data != nil {
		details.SubscriberID = rel.Data.ID
		if user, ok := included[resourceKey(rel.Data.Type, rel.Data.ID)]; ok {
			details.SubscriberName = stringAttr(user.Attributes, "name")
			details.SubscriberEmail = stringAttr(user.Attributes, "email-address")
		}
	}

	return details
}

func renderJobProductionPlanAlarmSubscriberDetails(cmd *cobra.Command, details jobProductionPlanAlarmSubscriberDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanAlarmID != "" {
		fmt.Fprintf(out, "Job Production Plan Alarm: %s\n", details.JobProductionPlanAlarmID)
	}
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.SubscriberID != "" {
		fmt.Fprintf(out, "Subscriber ID: %s\n", details.SubscriberID)
	}
	if details.SubscriberName != "" {
		fmt.Fprintf(out, "Subscriber Name: %s\n", details.SubscriberName)
	}
	if details.SubscriberEmail != "" {
		fmt.Fprintf(out, "Subscriber Email: %s\n", details.SubscriberEmail)
	}
	if details.AlarmTons != nil {
		fmt.Fprintf(out, "Alarm Tons: %v\n", details.AlarmTons)
	}
	if details.AlarmBaseMaterialType != "" {
		fmt.Fprintf(out, "Alarm Base Material Type: %s\n", details.AlarmBaseMaterialType)
	}
	if details.AlarmMaxLatencyMinutes != nil {
		fmt.Fprintf(out, "Alarm Max Latency Minutes: %v\n", details.AlarmMaxLatencyMinutes)
	}
	if details.AlarmNote != "" {
		fmt.Fprintf(out, "Alarm Note: %s\n", details.AlarmNote)
	}
	if details.AlarmPlannedAt != "" {
		fmt.Fprintf(out, "Alarm Planned At: %s\n", details.AlarmPlannedAt)
	}
	if details.AlarmFulfilledAt != "" {
		fmt.Fprintf(out, "Alarm Fulfilled At: %s\n", details.AlarmFulfilledAt)
	}
	if details.AlarmFulfilledTransactionAt != "" {
		fmt.Fprintf(out, "Alarm Fulfilled Transaction At: %s\n", details.AlarmFulfilledTransactionAt)
	}
	if details.AlarmPlanVarianceMinutes != nil {
		fmt.Fprintf(out, "Alarm Plan Variance Minutes: %v\n", details.AlarmPlanVarianceMinutes)
	}
	if details.AlarmGoodEnoughTimeZoneID != "" {
		fmt.Fprintf(out, "Alarm Time Zone: %s\n", details.AlarmGoodEnoughTimeZoneID)
	}
	if details.AlarmCanUpdate {
		fmt.Fprintln(out, "Alarm Can Update: true")
	}
	if details.AlarmCanDestroy {
		fmt.Fprintln(out, "Alarm Can Destroy: true")
	}

	return nil
}
