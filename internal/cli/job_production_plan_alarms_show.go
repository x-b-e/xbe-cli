package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
	"github.com/xbe-inc/xbe-cli/internal/auth"
)

type jobProductionPlanAlarmsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type jobProductionPlanAlarmDetails struct {
	ID                                 string   `json:"id"`
	JobProductionPlanID                string   `json:"job_production_plan_id,omitempty"`
	Tons                               float64  `json:"tons,omitempty"`
	BaseMaterialTypeFullyQualifiedName string   `json:"base_material_type_fully_qualified_name,omitempty"`
	MaxLatencyMinutes                  int      `json:"max_latency_minutes,omitempty"`
	Note                               string   `json:"note,omitempty"`
	PlannedAt                          string   `json:"planned_at,omitempty"`
	FulfilledAt                        string   `json:"fulfilled_at,omitempty"`
	GoodEnoughTimeZoneID               string   `json:"good_enough_time_zone_id,omitempty"`
	FulfilledTransactionAt             string   `json:"fulfilled_transaction_at,omitempty"`
	PlanVarianceMinutes                int      `json:"plan_variance_minutes,omitempty"`
	CanUpdate                          bool     `json:"can_update"`
	CanDestroy                         bool     `json:"can_destroy"`
	JobProductionPlanAlarmSubscribers  []string `json:"job_production_plan_alarm_subscribers,omitempty"`
}

func newJobProductionPlanAlarmsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show job production plan alarm details",
		Long: `Show the full details of a job production plan alarm.

Output Fields:
  ID                Alarm identifier
  Job Production Plan  Associated job production plan ID
  Tons              Tonnage trigger
  Base Material Type Base material type fully qualified name
  Max Latency Minutes Maximum latency in minutes
  Note              Alarm note
  Planned At         Planned fulfillment timestamp
  Fulfilled At       Fulfillment timestamp
  Good Enough Time Zone ID Time zone identifier
  Fulfilled Transaction At Fulfilled transaction timestamp
  Plan Variance Minutes Planned vs actual variance in minutes
  Can Update         Whether the alarm can be updated
  Can Destroy        Whether the alarm can be deleted
  Subscribers        Alarm subscriber IDs

Arguments:
  <id>    The alarm ID (required). You can find IDs using the list command.`,
		Example: `  # Show an alarm
  xbe view job-production-plan-alarms show 123

  # Get JSON output
  xbe view job-production-plan-alarms show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runJobProductionPlanAlarmsShow,
	}
	initJobProductionPlanAlarmsShowFlags(cmd)
	return cmd
}

func init() {
	jobProductionPlanAlarmsCmd.AddCommand(newJobProductionPlanAlarmsShowCmd())
}

func initJobProductionPlanAlarmsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runJobProductionPlanAlarmsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseJobProductionPlanAlarmsShowOptions(cmd)
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
		return fmt.Errorf("job production plan alarm id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/job-production-plan-alarms/"+id, nil)
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

	details := buildJobProductionPlanAlarmDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderJobProductionPlanAlarmDetails(cmd, details)
}

func parseJobProductionPlanAlarmsShowOptions(cmd *cobra.Command) (jobProductionPlanAlarmsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return jobProductionPlanAlarmsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildJobProductionPlanAlarmDetails(resp jsonAPISingleResponse) jobProductionPlanAlarmDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := jobProductionPlanAlarmDetails{
		ID:                                 resource.ID,
		Tons:                               floatAttr(attrs, "tons"),
		BaseMaterialTypeFullyQualifiedName: stringAttr(attrs, "base-material-type-fully-qualified-name"),
		MaxLatencyMinutes:                  intAttr(attrs, "max-latency-minutes"),
		Note:                               stringAttr(attrs, "note"),
		PlannedAt:                          formatDateTime(stringAttr(attrs, "planned-at")),
		FulfilledAt:                        formatDateTime(stringAttr(attrs, "fulfilled-at")),
		GoodEnoughTimeZoneID:               stringAttr(attrs, "good-enough-time-zone-id"),
		FulfilledTransactionAt:             formatDateTime(stringAttr(attrs, "fulfilled-transaction-at")),
		PlanVarianceMinutes:                intAttr(attrs, "plan-variance-minutes"),
		CanUpdate:                          boolAttr(attrs, "can-update"),
		CanDestroy:                         boolAttr(attrs, "can-destroy"),
	}

	if rel, ok := resource.Relationships["job-production-plan"]; ok && rel.Data != nil {
		details.JobProductionPlanID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["job-production-plan-alarm-subscribers"]; ok {
		details.JobProductionPlanAlarmSubscribers = alarmSubscriberIDs(rel)
	}

	return details
}

func alarmSubscriberIDs(rel jsonAPIRelationship) []string {
	if rel.Data != nil {
		return []string{rel.Data.ID}
	}
	identifiers := relationshipIDs(rel)
	if len(identifiers) == 0 {
		return nil
	}
	ids := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		if identifier.ID != "" {
			ids = append(ids, identifier.ID)
		}
	}
	return ids
}

func renderJobProductionPlanAlarmDetails(cmd *cobra.Command, details jobProductionPlanAlarmDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.JobProductionPlanID != "" {
		fmt.Fprintf(out, "Job Production Plan: %s\n", details.JobProductionPlanID)
	}
	if details.Tons > 0 {
		fmt.Fprintf(out, "Tons: %s\n", formatTons(details.Tons))
	}
	if details.BaseMaterialTypeFullyQualifiedName != "" {
		fmt.Fprintf(out, "Base Material Type: %s\n", details.BaseMaterialTypeFullyQualifiedName)
	}
	if details.MaxLatencyMinutes > 0 {
		fmt.Fprintf(out, "Max Latency Minutes: %d\n", details.MaxLatencyMinutes)
	}
	if details.PlannedAt != "" {
		fmt.Fprintf(out, "Planned At: %s\n", details.PlannedAt)
	}
	if details.FulfilledAt != "" {
		fmt.Fprintf(out, "Fulfilled At: %s\n", details.FulfilledAt)
	}
	if details.FulfilledTransactionAt != "" {
		fmt.Fprintf(out, "Fulfilled Transaction At: %s\n", details.FulfilledTransactionAt)
	}
	if details.PlanVarianceMinutes != 0 {
		fmt.Fprintf(out, "Plan Variance Minutes: %d\n", details.PlanVarianceMinutes)
	}
	if details.GoodEnoughTimeZoneID != "" {
		fmt.Fprintf(out, "Good Enough Time Zone ID: %s\n", details.GoodEnoughTimeZoneID)
	}
	if details.Note != "" {
		fmt.Fprintf(out, "Note: %s\n", details.Note)
	}
	fmt.Fprintf(out, "Can Update: %t\n", details.CanUpdate)
	fmt.Fprintf(out, "Can Destroy: %t\n", details.CanDestroy)
	if len(details.JobProductionPlanAlarmSubscribers) > 0 {
		fmt.Fprintf(out, "Subscribers: %s\n", strings.Join(details.JobProductionPlanAlarmSubscribers, ", "))
	}

	return nil
}
