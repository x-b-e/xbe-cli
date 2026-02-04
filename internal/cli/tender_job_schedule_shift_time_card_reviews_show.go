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

type tenderJobScheduleShiftTimeCardReviewsShowOptions struct {
	BaseURL string
	Token   string
	JSON    bool
	NoAuth  bool
}

type tenderJobScheduleShiftTimeCardReviewDetails struct {
	ID                        string `json:"id"`
	TenderJobScheduleShiftID  string `json:"tender_job_schedule_shift_id,omitempty"`
	Analysis                  string `json:"analysis,omitempty"`
	TimeCardStartAt           string `json:"time_card_start_at,omitempty"`
	TimeCardEndAt             string `json:"time_card_end_at,omitempty"`
	TimeCardDownMinutes       int    `json:"time_card_down_minutes,omitempty"`
	TimeCardStartAtConfidence bool   `json:"time_card_start_at_confidence"`
	TimeCardEndAtConfidence   bool   `json:"time_card_end_at_confidence"`
}

func newTenderJobScheduleShiftTimeCardReviewsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show tender job schedule shift time card review details",
		Long: `Show full details of a tender job schedule shift time card review.

Output Fields:
  ID                     Review identifier
  Tender Job Schedule Shift  Shift ID the review applies to
  Analysis               Review analysis summary
  Time Card Start/End    Suggested time card times
  Time Card Down Minutes Suggested down minutes
  Start/End Confidence   Time confidence flags

Arguments:
  <id>    Time card review ID (required)

Global flags (see xbe --help): --json, --base-url, --token, --no-auth`,
		Example: `  # Show a time card review
  xbe view tender-job-schedule-shift-time-card-reviews show 123

  # JSON output
  xbe view tender-job-schedule-shift-time-card-reviews show 123 --json`,
		Args: cobra.ExactArgs(1),
		RunE: runTenderJobScheduleShiftTimeCardReviewsShow,
	}
	initTenderJobScheduleShiftTimeCardReviewsShowFlags(cmd)
	return cmd
}

func init() {
	tenderJobScheduleShiftTimeCardReviewsCmd.AddCommand(newTenderJobScheduleShiftTimeCardReviewsShowCmd())
}

func initTenderJobScheduleShiftTimeCardReviewsShowFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().Bool("omit-null", false, "Omit null values in JSON output")
	cmd.Flags().Bool("no-auth", false, "Disable auth token lookup")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runTenderJobScheduleShiftTimeCardReviewsShow(cmd *cobra.Command, args []string) error {
	if handled, err := maybeHandleClientURLShow(cmd, args); err != nil {
		return err
	} else if handled {
		return nil
	}

	opts, err := parseTenderJobScheduleShiftTimeCardReviewsShowOptions(cmd)
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
		return fmt.Errorf("tender job schedule shift time card review id is required")
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	query := url.Values{}
	query.Set("fields[tender-job-schedule-shift-time-card-reviews]", "tender-job-schedule-shift,analysis,time-card-start-at,time-card-end-at,time-card-down-minutes,time-card-start-at-confidence,time-card-end-at-confidence")

	body, _, err := client.Get(cmd.Context(), "/v1/tender-job-schedule-shift-time-card-reviews/"+id, query)
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

	details := buildTenderJobScheduleShiftTimeCardReviewDetails(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), details)
	}

	return renderTenderJobScheduleShiftTimeCardReviewDetails(cmd, details)
}

func parseTenderJobScheduleShiftTimeCardReviewsShowOptions(cmd *cobra.Command) (tenderJobScheduleShiftTimeCardReviewsShowOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	noAuth, _ := cmd.Flags().GetBool("no-auth")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return tenderJobScheduleShiftTimeCardReviewsShowOptions{
		BaseURL: baseURL,
		Token:   token,
		JSON:    jsonOut,
		NoAuth:  noAuth,
	}, nil
}

func buildTenderJobScheduleShiftTimeCardReviewDetails(resp jsonAPISingleResponse) tenderJobScheduleShiftTimeCardReviewDetails {
	attrs := resp.Data.Attributes
	details := tenderJobScheduleShiftTimeCardReviewDetails{
		ID:                        resp.Data.ID,
		Analysis:                  strings.TrimSpace(stringAttr(attrs, "analysis")),
		TimeCardStartAt:           formatDateTime(stringAttr(attrs, "time-card-start-at")),
		TimeCardEndAt:             formatDateTime(stringAttr(attrs, "time-card-end-at")),
		TimeCardDownMinutes:       intAttr(attrs, "time-card-down-minutes"),
		TimeCardStartAtConfidence: boolAttr(attrs, "time-card-start-at-confidence"),
		TimeCardEndAtConfidence:   boolAttr(attrs, "time-card-end-at-confidence"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		details.TenderJobScheduleShiftID = rel.Data.ID
	}

	return details
}

func renderTenderJobScheduleShiftTimeCardReviewDetails(cmd *cobra.Command, details tenderJobScheduleShiftTimeCardReviewDetails) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "ID: %s\n", details.ID)
	if details.TenderJobScheduleShiftID != "" {
		fmt.Fprintf(out, "Tender Job Schedule Shift: %s\n", details.TenderJobScheduleShiftID)
	}
	fmt.Fprintf(out, "Analysis: %s\n", formatOptional(details.Analysis))
	fmt.Fprintf(out, "Time Card Start At: %s\n", formatOptional(details.TimeCardStartAt))
	fmt.Fprintf(out, "Time Card End At: %s\n", formatOptional(details.TimeCardEndAt))
	fmt.Fprintf(out, "Time Card Down Minutes: %d\n", details.TimeCardDownMinutes)
	fmt.Fprintf(out, "Start Time Confidence: %s\n", formatBool(details.TimeCardStartAtConfidence))
	fmt.Fprintf(out, "End Time Confidence: %s\n", formatBool(details.TimeCardEndAtConfidence))

	return nil
}
