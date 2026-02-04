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

type projectTransportPlanPlannedEventTimeSchedulesShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type projectTransportPlanPlannedEventTimeScheduleDetails struct {
	ID                        string `json:"id"`
	ProjectTransportPlanID    string `json:"project_transport_plan_id,omitempty"`
	TransportOrderID          string `json:"transport_order_id,omitempty"`
	PlanData                  any    `json:"plan_data,omitempty"`
	RespectProvidedEventTimes bool   `json:"respect_provided_event_times"`
	Schedule                  any    `json:"schedule,omitempty"`
	Warnings                  any    `json:"warnings,omitempty"`
	Success                   bool   `json:"success"`
}

func newProjectTransportPlanPlannedEventTimeSchedulesShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show project transport plan planned event time schedule details",
		Long: `Show the full details of a project transport plan planned event time schedule.

Output Fields:
  ID
  Project Transport Plan ID
  Transport Order ID
  Plan Data
  Respect Provided Event Times
  Schedule
  Warnings
  Success

Arguments:
  <id>    The schedule ID (required). You can find IDs using the list command.`,
		Example: `  # Show schedule details
  xbe view project-transport-plan-planned-event-time-schedules show 123

  # JSON output
  xbe view project-transport-plan-planned-event-time-schedules show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runProjectTransportPlanPlannedEventTimeSchedulesShow,
	}
	initProjectTransportPlanPlannedEventTimeSchedulesShowFlags(cmd)
	return cmd
}

func init() {
	projectTransportPlanPlannedEventTimeSchedulesCmd.AddCommand(newProjectTransportPlanPlannedEventTimeSchedulesShowCmd())
}

func initProjectTransportPlanPlannedEventTimeSchedulesShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runProjectTransportPlanPlannedEventTimeSchedulesShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseProjectTransportPlanPlannedEventTimeSchedulesShowOptions(cmd)
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
		return fmt.Errorf("project transport plan planned event time schedule id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/project-transport-plan-planned-event-time-schedules/"+id, nil)
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

	details := buildProjectTransportPlanPlannedEventTimeScheduleDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderProjectTransportPlanPlannedEventTimeScheduleDetails(cmd, details)
}

func parseProjectTransportPlanPlannedEventTimeSchedulesShowOptions(cmd *cobra.Command) (projectTransportPlanPlannedEventTimeSchedulesShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return projectTransportPlanPlannedEventTimeSchedulesShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildProjectTransportPlanPlannedEventTimeScheduleDetails(resp jsonAPISingleResponse) projectTransportPlanPlannedEventTimeScheduleDetails {
	attrs := resp.Data.Attributes
	details := projectTransportPlanPlannedEventTimeScheduleDetails{
		ID:                        resp.Data.ID,
		RespectProvidedEventTimes: boolAttr(attrs, "respect-provided-event-times"),
		Success:                   boolAttr(attrs, "success"),
	}

	if planData, ok := attrs["plan-data"]; ok {
		details.PlanData = planData
	}
	if schedule, ok := attrs["schedule"]; ok {
		details.Schedule = schedule
	}
	if warnings, ok := attrs["warnings"]; ok {
		details.Warnings = warnings
	}

	if rel, ok := resp.Data.Relationships["project-transport-plan"]; ok && rel.Data != nil {
		details.ProjectTransportPlanID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["transport-order"]; ok && rel.Data != nil {
		details.TransportOrderID = rel.Data.ID
	}

	return details
}

func renderProjectTransportPlanPlannedEventTimeScheduleDetails(cmd *cobra.Command, details projectTransportPlanPlannedEventTimeScheduleDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.ProjectTransportPlanID != "" {
		fmt.Fprintf(out, "Project Transport Plan ID: %s\n", details.ProjectTransportPlanID)
	}
	if details.TransportOrderID != "" {
		fmt.Fprintf(out, "Transport Order ID: %s\n", details.TransportOrderID)
	}

	fmt.Fprintf(out, "Respect Provided Event Times: %t\n", details.RespectProvidedEventTimes)
	fmt.Fprintf(out, "Success: %t\n", details.Success)

	if formatted := formatJSONValue(details.PlanData); formatted != "" {
		fmt.Fprintf(out, "Plan Data: %s\n", formatted)
	}
	if formatted := formatJSONValue(details.Schedule); formatted != "" {
		fmt.Fprintf(out, "Schedule: %s\n", formatted)
	}
	if formatted := formatJSONValue(details.Warnings); formatted != "" {
		fmt.Fprintf(out, "Warnings: %s\n", formatted)
	}

	return nil
}
