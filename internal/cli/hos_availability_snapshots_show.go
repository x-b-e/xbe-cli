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

type hosAvailabilitySnapshotsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type hosAvailabilitySnapshotDetails struct {
	ID                                  string `json:"id"`
	CapturedAt                          string `json:"captured_at,omitempty"`
	WorkStatus                          string `json:"work_status,omitempty"`
	IsAvailable                         bool   `json:"is_available"`
	WorkdaySecondsRemaining             string `json:"workday_seconds_remaining,omitempty"`
	DrivingSecondsRemaining             string `json:"driving_seconds_remaining,omitempty"`
	DutySecondsRemaining                string `json:"duty_seconds_remaining,omitempty"`
	CycleSecondsRemainingToday          string `json:"cycle_seconds_remaining_today,omitempty"`
	CycleSecondsAvailableTomorrow       string `json:"cycle_seconds_available_tomorrow,omitempty"`
	BreakSecondsRemaining               string `json:"break_seconds_remaining,omitempty"`
	ProjectedAsOf                       string `json:"projected_as_of,omitempty"`
	ProjectedWorkdaySecondsRemaining    string `json:"projected_workday_seconds_remaining,omitempty"`
	ProjectedDrivingSecondsRemaining    string `json:"projected_driving_seconds_remaining,omitempty"`
	ProjectedCycleSecondsRemainingToday string `json:"projected_cycle_seconds_remaining_today,omitempty"`
	BrokerID                            string `json:"broker_id,omitempty"`
	HosDayID                            string `json:"hos_day_id,omitempty"`
	UserID                              string `json:"user_id,omitempty"`
}

func newHosAvailabilitySnapshotsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show HOS availability snapshot details",
		Long: `Show the full details of an HOS availability snapshot.

Output Fields:
  ID
  Captured At
  Work Status
  Available
  Workday Seconds Remaining
  Driving Seconds Remaining
  Duty Seconds Remaining
  Cycle Seconds Remaining (Today)
  Cycle Seconds Available (Tomorrow)
  Break Seconds Remaining
  Projected As Of
  Projected Workday Seconds Remaining
  Projected Driving Seconds Remaining
  Projected Cycle Seconds Remaining (Today)
  Broker ID
  HOS Day ID
  User ID

Global flags (see xbe --help): --json, --base-url, --token, --no-auth

Arguments:
  <id>    The snapshot ID (required). You can find IDs using the list command.`,
		Example: `  # Show a snapshot
  xbe view hos-availability-snapshots show 123

  # Get JSON output
  xbe view hos-availability-snapshots show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runHosAvailabilitySnapshotsShow,
	}
	initHosAvailabilitySnapshotsShowFlags(cmd)
	return cmd
}

func init() {
	hosAvailabilitySnapshotsCmd.AddCommand(newHosAvailabilitySnapshotsShowCmd())
}

func initHosAvailabilitySnapshotsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runHosAvailabilitySnapshotsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseHosAvailabilitySnapshotsShowOptions(cmd)
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
		return fmt.Errorf("hos availability snapshot id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[hos-availability-snapshots]", "captured-at,work-status,is-available,workday-seconds-remaining,driving-seconds-remaining,duty-seconds-remaining,cycle-seconds-remaining-today,cycle-seconds-available-tomorrow,break-seconds-remaining,projected-as-of,projected-workday-seconds-remaining,projected-driving-seconds-remaining,projected-cycle-seconds-remaining-today")

	body, _, err := client.Get(cmd.Context(), "/v1/hos-availability-snapshots/"+id, query)
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

	details := buildHosAvailabilitySnapshotDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderHosAvailabilitySnapshotDetails(cmd, details)
}

func parseHosAvailabilitySnapshotsShowOptions(cmd *cobra.Command) (hosAvailabilitySnapshotsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return hosAvailabilitySnapshotsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildHosAvailabilitySnapshotDetails(resp jsonAPISingleResponse) hosAvailabilitySnapshotDetails {
	resource := resp.Data
	attrs := resource.Attributes

	details := hosAvailabilitySnapshotDetails{
		ID:                                  resource.ID,
		CapturedAt:                          formatDateTime(stringAttr(attrs, "captured-at")),
		WorkStatus:                          stringAttr(attrs, "work-status"),
		IsAvailable:                         boolAttr(attrs, "is-available"),
		WorkdaySecondsRemaining:             stringAttr(attrs, "workday-seconds-remaining"),
		DrivingSecondsRemaining:             stringAttr(attrs, "driving-seconds-remaining"),
		DutySecondsRemaining:                stringAttr(attrs, "duty-seconds-remaining"),
		CycleSecondsRemainingToday:          stringAttr(attrs, "cycle-seconds-remaining-today"),
		CycleSecondsAvailableTomorrow:       stringAttr(attrs, "cycle-seconds-available-tomorrow"),
		BreakSecondsRemaining:               stringAttr(attrs, "break-seconds-remaining"),
		ProjectedAsOf:                       formatDateTime(stringAttr(attrs, "projected-as-of")),
		ProjectedWorkdaySecondsRemaining:    stringAttr(attrs, "projected-workday-seconds-remaining"),
		ProjectedDrivingSecondsRemaining:    stringAttr(attrs, "projected-driving-seconds-remaining"),
		ProjectedCycleSecondsRemainingToday: stringAttr(attrs, "projected-cycle-seconds-remaining-today"),
	}

	if rel, ok := resource.Relationships["broker"]; ok && rel.Data != nil {
		details.BrokerID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hos-day"]; ok && rel.Data != nil {
		details.HosDayID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["user"]; ok && rel.Data != nil {
		details.UserID = rel.Data.ID
	}

	return details
}

func renderHosAvailabilitySnapshotDetails(cmd *cobra.Command, details hosAvailabilitySnapshotDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.CapturedAt != "" {
		fmt.Fprintf(out, "Captured At: %s\n", details.CapturedAt)
	}
	if details.WorkStatus != "" {
		fmt.Fprintf(out, "Work Status: %s\n", details.WorkStatus)
	}
	fmt.Fprintf(out, "Available: %s\n", formatYesNo(details.IsAvailable))
	if details.WorkdaySecondsRemaining != "" {
		fmt.Fprintf(out, "Workday Seconds Remaining: %s\n", details.WorkdaySecondsRemaining)
	}
	if details.DrivingSecondsRemaining != "" {
		fmt.Fprintf(out, "Driving Seconds Remaining: %s\n", details.DrivingSecondsRemaining)
	}
	if details.DutySecondsRemaining != "" {
		fmt.Fprintf(out, "Duty Seconds Remaining: %s\n", details.DutySecondsRemaining)
	}
	if details.CycleSecondsRemainingToday != "" {
		fmt.Fprintf(out, "Cycle Seconds Remaining (Today): %s\n", details.CycleSecondsRemainingToday)
	}
	if details.CycleSecondsAvailableTomorrow != "" {
		fmt.Fprintf(out, "Cycle Seconds Available (Tomorrow): %s\n", details.CycleSecondsAvailableTomorrow)
	}
	if details.BreakSecondsRemaining != "" {
		fmt.Fprintf(out, "Break Seconds Remaining: %s\n", details.BreakSecondsRemaining)
	}
	if details.ProjectedAsOf != "" {
		fmt.Fprintf(out, "Projected As Of: %s\n", details.ProjectedAsOf)
	}
	if details.ProjectedWorkdaySecondsRemaining != "" {
		fmt.Fprintf(out, "Projected Workday Seconds Remaining: %s\n", details.ProjectedWorkdaySecondsRemaining)
	}
	if details.ProjectedDrivingSecondsRemaining != "" {
		fmt.Fprintf(out, "Projected Driving Seconds Remaining: %s\n", details.ProjectedDrivingSecondsRemaining)
	}
	if details.ProjectedCycleSecondsRemainingToday != "" {
		fmt.Fprintf(out, "Projected Cycle Seconds Remaining (Today): %s\n", details.ProjectedCycleSecondsRemainingToday)
	}
	if details.BrokerID != "" {
		fmt.Fprintf(out, "Broker ID: %s\n", details.BrokerID)
	}
	if details.HosDayID != "" {
		fmt.Fprintf(out, "HOS Day ID: %s\n", details.HosDayID)
	}
	if details.UserID != "" {
		fmt.Fprintf(out, "User ID: %s\n", details.UserID)
	}

	return nil
}
