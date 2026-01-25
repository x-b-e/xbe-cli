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

type tenderJobScheduleShiftCancellationsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderJobScheduleShiftCancellationDetails struct {
	ID                                   string `json:"id"`
	TenderJobScheduleShiftID             string `json:"tender_job_schedule_shift_id,omitempty"`
	StatusChangedByID                    string `json:"status_changed_by_id,omitempty"`
	StatusChangeComment                  string `json:"status_change_comment,omitempty"`
	IsReturned                           bool   `json:"is_returned,omitempty"`
	JobProductionPlanCancellationComment string `json:"job_production_plan_cancellation_comment,omitempty"`
	SkipTruckerNotifications             bool   `json:"skip_trucker_notifications,omitempty"`
}

func newTenderJobScheduleShiftCancellationsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender job schedule shift cancellation details",
		Long: `Show the full details of a tender job schedule shift cancellation.

Output Fields:
  ID             Cancellation identifier
  SHIFT          Tender job schedule shift ID
  CHANGED BY     User who changed the status
  COMMENT        Status change comment
  RETURNED       Whether the tender was returned
  JPP COMMENT    Job production plan cancellation comment
  SKIP NOTIFS    Whether trucker notifications were skipped

Arguments:
  <id>  Cancellation ID (required).

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a cancellation
  xbe view tender-job-schedule-shift-cancellations show 123

  # Output as JSON
  xbe view tender-job-schedule-shift-cancellations show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderJobScheduleShiftCancellationsShow,
	}
	initTenderJobScheduleShiftCancellationsShowFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftCancellationsCmd.AddCommand(newTenderJobScheduleShiftCancellationsShowCmd())
}

func initTenderJobScheduleShiftCancellationsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftCancellationsShow(cmd *cobra.Command, args []string) error {
	opts, err := parseTenderJobScheduleShiftCancellationsShowOptions(cmd)
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
		return fmt.Errorf("tender job schedule shift cancellation id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tender-job-schedule-shift-cancellations]", "tender-job-schedule-shift,status-change-comment,status-changed-by,is-returned,job-production-plan-cancellation-comment,skip-trucker-notifications")

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shift-cancellations/"+id, query)
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

	details := buildTenderJobScheduleShiftCancellationDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderJobScheduleShiftCancellationDetails(cmd, details)
}

func parseTenderJobScheduleShiftCancellationsShowOptions(cmd *cobra.Command) (tenderJobScheduleShiftCancellationsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftCancellationsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderJobScheduleShiftCancellationDetails(resp jsonAPISingleResponse) tenderJobScheduleShiftCancellationDetails {
	attrs := resp.Data.Attributes
	details := tenderJobScheduleShiftCancellationDetails{
		ID:                                   resp.Data.ID,
		StatusChangeComment:                  stringAttr(attrs, "status-change-comment"),
		IsReturned:                           boolAttr(attrs, "is-returned"),
		JobProductionPlanCancellationComment: stringAttr(attrs, "job-production-plan-cancellation-comment"),
		SkipTruckerNotifications:             boolAttr(attrs, "skip-trucker-notifications"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resp.Data.Relationships["status-changed-by"]; ok && rel.Data != nil {
		details.StatusChangedByID = rel.Data.ID
	}

	return details
}

func renderTenderJobScheduleShiftCancellationDetails(cmd *cobra.Command, details tenderJobScheduleShiftCancellationDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShiftID)
	}
	if details.StatusChangedByID != "" {
		fmt.Fprintf(out, "Status Changed By: %s\n", details.StatusChangedByID)
	}
	if details.StatusChangeComment != "" {
		fmt.Fprintf(out, "Status Change Comment: %s\n", details.StatusChangeComment)
	}
	fmt.Fprintf(out, "Is Returned: %t\n", details.IsReturned)
	fmt.Fprintf(out, "Skip Trucker Notifications: %t\n", details.SkipTruckerNotifications)
	if details.JobProductionPlanCancellationComment != "" {
		fmt.Fprintf(out, "Job Production Plan Cancellation Comment: %s\n", details.JobProductionPlanCancellationComment)
	}

	return nil
}
