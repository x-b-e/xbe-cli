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

type doDownMinutesEstimatesCreateOptions struct {
	BaseURL                string
	Token                  string
	JSON                   bool
	TenderJobScheduleShift string
	TimeCardStartAt        string
	TimeCardEndAt          string
}

type downMinutesEstimateRow struct {
	ID                           string  `json:"id"`
	TimeCardStartAt              string  `json:"time_card_start_at,omitempty"`
	TimeCardEndAt                string  `json:"time_card_end_at,omitempty"`
	Minutes                      float64 `json:"minutes,omitempty"`
	SquanderedMinutes            float64 `json:"squandered_minutes,omitempty"`
	BeforeTimeCardDownMinutes    float64 `json:"before_time_card_down_minutes,omitempty"`
	TimeCardDownMinutes          float64 `json:"time_card_down_minutes,omitempty"`
	AfterTimeCardDownMinutes     float64 `json:"after_time_card_down_minutes,omitempty"`
	IsDownTimeIncidentWithoutEnd bool    `json:"is_down_time_incident_without_end_at,omitempty"`
	BeforeTimeCardCredited       float64 `json:"before_time_card_credited_minutes,omitempty"`
	TimeCardCredited             float64 `json:"time_card_credited_minutes,omitempty"`
	AfterTimeCardCredited        float64 `json:"after_time_card_credited_minutes,omitempty"`
	TotalCreditedMinutes         float64 `json:"total_credited_minutes,omitempty"`
	TenderJobScheduleShiftID     string  `json:"tender_job_schedule_shift_id,omitempty"`
}

func newDoDownMinutesEstimatesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Estimate down minutes for a shift",
		Long: `Estimate down minutes for a tender job schedule shift.

Required flags:
  --tender-job-schedule-shift  Tender job schedule shift ID (required)

Optional flags:
  --time-card-start-at         Time card start time (ISO-8601)
  --time-card-end-at           Time card end time (ISO-8601)`,
		Example: `  # Estimate down minutes for a shift
  xbe do down-minutes-estimates create --tender-job-schedule-shift 123

  # Provide a time card window
  xbe do down-minutes-estimates create --tender-job-schedule-shift 123 --time-card-start-at 2025-01-01T08:00:00Z --time-card-end-at 2025-01-01T12:00:00Z

  # Output as JSON
  xbe do down-minutes-estimates create --tender-job-schedule-shift 123 --json`,
		Args: cobra.NoArgs,
		RunE: runDoDownMinutesEstimatesCreate,
	}
	initDoDownMinutesEstimatesCreateFlags(cmd)
	return cmd
}

func init() {
	doDownMinutesEstimatesCmd.AddCommand(newDoDownMinutesEstimatesCreateCmd())
}

func initDoDownMinutesEstimatesCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("tender-job-schedule-shift", "", "Tender job schedule shift ID (required)")
	cmd.Flags().String("time-card-start-at", "", "Time card start time (ISO-8601)")
	cmd.Flags().String("time-card-end-at", "", "Time card end time (ISO-8601)")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoDownMinutesEstimatesCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoDownMinutesEstimatesCreateOptions(cmd)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	if strings.TrimSpace(opts.Token) == "" {
		if token, _, err := auth.ResolveToken(opts.BaseURL, ""); err == nil {
			opts.Token = token
		} else if errors.Is(err, auth.ErrNotFound) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Authentication required. Run xbe auth login first.")
			return err
		} else {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	if strings.TrimSpace(opts.TenderJobScheduleShift) == "" {
		err := fmt.Errorf("--tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	attributes := map[string]any{}
	if strings.TrimSpace(opts.TimeCardStartAt) != "" {
		attributes["time-card-start-at"] = opts.TimeCardStartAt
	}
	if strings.TrimSpace(opts.TimeCardEndAt) != "" {
		attributes["time-card-end-at"] = opts.TimeCardEndAt
	}

	relationships := map[string]any{
		"tender-job-schedule-shift": map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.TenderJobScheduleShift,
			},
		},
	}

	data := map[string]any{
		"type":          "down-minutes-estimates",
		"relationships": relationships,
	}
	if len(attributes) > 0 {
		data["attributes"] = attributes
	}

	requestBody := map[string]any{
		"data": data,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/down-minutes-estimates", jsonBody)
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

	row := downMinutesEstimateRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	if strings.TrimSpace(row.ID) == "" {
		fmt.Fprintln(cmd.OutOrStdout(), "Created down minutes estimate")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created down minutes estimate %s\n", row.ID)
	return nil
}

func parseDoDownMinutesEstimatesCreateOptions(cmd *cobra.Command) (doDownMinutesEstimatesCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	tenderJobScheduleShift, _ := cmd.Flags().GetString("tender-job-schedule-shift")
	timeCardStartAt, _ := cmd.Flags().GetString("time-card-start-at")
	timeCardEndAt, _ := cmd.Flags().GetString("time-card-end-at")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doDownMinutesEstimatesCreateOptions{
		BaseURL:                baseURL,
		Token:                  token,
		JSON:                   jsonOut,
		TenderJobScheduleShift: tenderJobScheduleShift,
		TimeCardStartAt:        timeCardStartAt,
		TimeCardEndAt:          timeCardEndAt,
	}, nil
}

func downMinutesEstimateRowFromSingle(resp jsonAPISingleResponse) downMinutesEstimateRow {
	attrs := resp.Data.Attributes
	row := downMinutesEstimateRow{
		ID:                           resp.Data.ID,
		TimeCardStartAt:              formatDateTime(stringAttr(attrs, "time-card-start-at")),
		TimeCardEndAt:                formatDateTime(stringAttr(attrs, "time-card-end-at")),
		Minutes:                      floatAttr(attrs, "minutes"),
		SquanderedMinutes:            floatAttr(attrs, "squandered-minutes"),
		BeforeTimeCardDownMinutes:    floatAttr(attrs, "before-time-card-down-minutes"),
		TimeCardDownMinutes:          floatAttr(attrs, "time-card-down-minutes"),
		AfterTimeCardDownMinutes:     floatAttr(attrs, "after-time-card-down-minutes"),
		IsDownTimeIncidentWithoutEnd: boolAttr(attrs, "is-down-time-incident-without-end-at"),
		BeforeTimeCardCredited:       floatAttr(attrs, "before-time-card-credited-minutes"),
		TimeCardCredited:             floatAttr(attrs, "time-card-credited-minutes"),
		AfterTimeCardCredited:        floatAttr(attrs, "after-time-card-credited-minutes"),
		TotalCreditedMinutes:         floatAttr(attrs, "total-credited-minutes"),
	}

	if rel, ok := resp.Data.Relationships["tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.TenderJobScheduleShiftID = rel.Data.ID
	}

	return row
}
