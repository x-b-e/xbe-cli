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

type doBuiltTimeCardsCreateOptions struct {
	BaseURL                        string
	Token                          string
	JSON                           bool
	BrokerTenderJobScheduleShift   string
	CustomerTenderJobScheduleShift string
}

type builtTimeCardRow struct {
	ID                                 string  `json:"id"`
	StartAt                            string  `json:"start_at,omitempty"`
	EndAt                              string  `json:"end_at,omitempty"`
	TonsQuantity                       float64 `json:"tons_quantity,omitempty"`
	HoursQuantity                      float64 `json:"hours_quantity,omitempty"`
	DownMinutes                        float64 `json:"down_minutes,omitempty"`
	JobProductionPlanTimeCardCreatable bool    `json:"job_production_plan_time_card_creatable,omitempty"`
	SubmittedTravelMinutes             float64 `json:"submitted_travel_minutes,omitempty"`
	SubmittedBy                        string  `json:"submitted_by,omitempty"`
	TimeZoneID                         string  `json:"time_zone_id,omitempty"`
	BrokerTenderJobScheduleShiftID     string  `json:"broker_tender_job_schedule_shift_id,omitempty"`
	CustomerTenderJobScheduleShiftID   string  `json:"customer_tender_job_schedule_shift_id,omitempty"`
	TonsServiceTypeUnitOfMeasureID     string  `json:"tons_service_type_unit_of_measure_id,omitempty"`
	HoursServiceTypeUnitOfMeasureID    string  `json:"hours_service_type_unit_of_measure_id,omitempty"`
}

func newDoBuiltTimeCardsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Build a time card from a shift",
		Long: `Build a time card from a broker or customer tender job schedule shift.

Required flags:
  --broker-tender-job-schedule-shift    Broker tender job schedule shift ID
  --customer-tender-job-schedule-shift  Customer tender job schedule shift ID

Provide one of the shift flags above.`,
		Example: `  # Build a time card from a broker tender shift
  xbe do built-time-cards create --broker-tender-job-schedule-shift 123

  # Build using a customer tender shift
  xbe do built-time-cards create --customer-tender-job-schedule-shift 456 --json`,
		Args: cobra.NoArgs,
		RunE: runDoBuiltTimeCardsCreate,
	}
	initDoBuiltTimeCardsCreateFlags(cmd)
	return cmd
}

func init() {
	doBuiltTimeCardsCmd.AddCommand(newDoBuiltTimeCardsCreateCmd())
}

func initDoBuiltTimeCardsCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output JSON")
	cmd.Flags().String("broker-tender-job-schedule-shift", "", "Broker tender job schedule shift ID")
	cmd.Flags().String("customer-tender-job-schedule-shift", "", "Customer tender job schedule shift ID")
	cmd.Flags().String("base-url", defaultBaseURL(), "API base URL")
	cmd.Flags().String("token", "", "API token (optional)")
}

func runDoBuiltTimeCardsCreate(cmd *cobra.Command, _ []string) error {
	opts, err := parseDoBuiltTimeCardsCreateOptions(cmd)
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

	if strings.TrimSpace(opts.BrokerTenderJobScheduleShift) == "" && strings.TrimSpace(opts.CustomerTenderJobScheduleShift) == "" {
		err := fmt.Errorf("either --broker-tender-job-schedule-shift or --customer-tender-job-schedule-shift is required")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	relationships := map[string]any{}
	if strings.TrimSpace(opts.BrokerTenderJobScheduleShift) != "" {
		relationships["broker-tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.BrokerTenderJobScheduleShift,
			},
		}
	}
	if strings.TrimSpace(opts.CustomerTenderJobScheduleShift) != "" {
		relationships["customer-tender-job-schedule-shift"] = map[string]any{
			"data": map[string]any{
				"type": "tender-job-schedule-shifts",
				"id":   opts.CustomerTenderJobScheduleShift,
			},
		}
	}

	requestBody := map[string]any{
		"data": map[string]any{
			"type":          "built-time-cards",
			"relationships": relationships,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	client := api.NewClient(opts.BaseURL, opts.Token)

	body, _, err := client.Post(cmd.Context(), "/v1/built-time-cards", jsonBody)
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

	row := builtTimeCardRowFromSingle(resp)
	if opts.JSON {
		return writeJSON(cmd.OutOrStdout(), row)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created built time card %s\n", row.ID)
	return nil
}

func builtTimeCardRowFromSingle(resp jsonAPISingleResponse) builtTimeCardRow {
	resource := resp.Data
	attrs := resource.Attributes
	row := builtTimeCardRow{
		ID:                                 resource.ID,
		StartAt:                            formatDateTime(stringAttr(attrs, "start-at")),
		EndAt:                              formatDateTime(stringAttr(attrs, "end-at")),
		TonsQuantity:                       floatAttr(attrs, "tons-quantity"),
		HoursQuantity:                      floatAttr(attrs, "hours-quantity"),
		DownMinutes:                        floatAttr(attrs, "down-minutes"),
		JobProductionPlanTimeCardCreatable: boolAttr(attrs, "job-production-plan-time-card-creatable"),
		SubmittedTravelMinutes:             floatAttr(attrs, "submitted-travel-minutes"),
		SubmittedBy:                        stringAttr(attrs, "submitted-by"),
		TimeZoneID:                         stringAttr(attrs, "time-zone-id"),
	}

	if rel, ok := resource.Relationships["broker-tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.BrokerTenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["customer-tender-job-schedule-shift"]; ok && rel.Data != nil {
		row.CustomerTenderJobScheduleShiftID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["tons-service-type-unit-of-measure"]; ok && rel.Data != nil {
		row.TonsServiceTypeUnitOfMeasureID = rel.Data.ID
	}
	if rel, ok := resource.Relationships["hours-service-type-unit-of-measure"]; ok && rel.Data != nil {
		row.HoursServiceTypeUnitOfMeasureID = rel.Data.ID
	}

	return row
}

func parseDoBuiltTimeCardsCreateOptions(cmd *cobra.Command) (doBuiltTimeCardsCreateOptions, error) {
	jsonOut, _ := cmd.Flags().GetBool("json")
	brokerTenderJobScheduleShift, _ := cmd.Flags().GetString("broker-tender-job-schedule-shift")
	customerTenderJobScheduleShift, _ := cmd.Flags().GetString("customer-tender-job-schedule-shift")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")

	return doBuiltTimeCardsCreateOptions{
		BaseURL:                        baseURL,
		Token:                          token,
		JSON:                           jsonOut,
		BrokerTenderJobScheduleShift:   brokerTenderJobScheduleShift,
		CustomerTenderJobScheduleShift: customerTenderJobScheduleShift,
	}, nil
}
