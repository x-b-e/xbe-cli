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

type lineupDispatchShiftsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type lineupDispatchShiftDetails struct {
	ID                     string `json:"id"`
	LineupDispatchID       string `json:"lineup_dispatch_id,omitempty"`
	LineupJobScheduleShift string `json:"lineup_job_schedule_shift_id,omitempty"`
	FulfilledTruckerID     string `json:"fulfilled_trucker_id,omitempty"`
	CancelledAt            string `json:"cancelled_at,omitempty"`
}

func newLineupDispatchShiftsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show lineup dispatch shift details",
		Long: `Show the full details of a lineup dispatch shift.

Output Fields:
  ID              Lineup dispatch shift identifier
  DISPATCH        Lineup dispatch ID
  SCHEDULE SHIFT  Lineup job schedule shift ID
  FULFILLED BY    Fulfilled trucker ID
  CANCELLED AT    Cancellation timestamp

Arguments:
  <id>  The lineup dispatch shift ID (required).`,
		Example: `  # Show a lineup dispatch shift
  xbe view lineup-dispatch-shifts show 123

  # JSON output
  xbe view lineup-dispatch-shifts show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runLineupDispatchShiftsShow,
	}
	initLineupDispatchShiftsShowFlags(cmd)
	return cmd
}

func init() {
	lineupDispatchShiftsCmd.AddCommand(newLineupDispatchShiftsShowCmd())
}

func initLineupDispatchShiftsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runLineupDispatchShiftsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseLineupDispatchShiftsShowOptions(cmd)
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
		return fmt.Errorf("lineup dispatch shift id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Get(cmd.Context(), "/v1/lineup-dispatch-shifts/"+id, nil)
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

	details := buildLineupDispatchShiftDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderLineupDispatchShiftDetails(cmd, details)
}

func parseLineupDispatchShiftsShowOptions(cmd *cobra.Command) (lineupDispatchShiftsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return lineupDispatchShiftsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildLineupDispatchShiftDetails(resp jsonAPISingleResponse) lineupDispatchShiftDetails {
	resource := resp.Data
	attrs := resource.Attributes
	details := lineupDispatchShiftDetails{
		ID:          resource.ID,
		CancelledAt: formatDateTime(stringAttr(attrs, "cancelled-at")),
	}

	if rel, ok := resource.Relationships["lineup-dispatch"]; ok && rel.Data != nil {
		details.LineupDispatchID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["lineup-job-schedule-shift"]; ok && rel.Data != nil {
		details.LineupJobScheduleShift = rel.Data.ID
	}
	if rel, ok := resource.Relationships["fulfilled-trucker"]; ok && rel.Data != nil {
		details.FulfilledTruckerID = rel.Data.ID
	}

	return details
}

func renderLineupDispatchShiftDetails(cmd *cobra.Command, details lineupDispatchShiftDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.LineupDispatchID != "" {
		fmt.Fprintf(out, "Lineup Dispatch: %s\n", details.LineupDispatchID)
	}
	if details.LineupJobScheduleShift != "" {
		fmt.Fprintf(out, "Schedule Shift: %s\n", details.LineupJobScheduleShift)
	}
	if details.FulfilledTruckerID != "" {
		fmt.Fprintf(out, "Fulfilled By: %s\n", details.FulfilledTruckerID)
	}
	if details.CancelledAt != "" {
		fmt.Fprintf(out, "Cancelled At: %s\n", details.CancelledAt)
	}

	return nil
}
