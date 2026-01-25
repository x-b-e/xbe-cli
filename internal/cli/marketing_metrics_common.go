package cli

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/xbe-inc/xbe-cli/internal/api"
)

type marketingMetricsListRow struct {
	ID                string  `json:"id"`
	ShiftCount        int     `json:"shift_count"`
	DriverDayCount    int     `json:"driver_day_count"`
	TonsSum           float64 `json:"tons_sum"`
	DriverCount       int     `json:"driver_count"`
	UserCount         int     `json:"user_count"`
	TruckerCount      int     `json:"trucker_count"`
	IncidentCount     int     `json:"incident_count"`
	ActiveBranchCount int     `json:"active_branch_count"`
}

type marketingMetricsDetails struct {
	ID                       string  `json:"id"`
	ShiftCount               int     `json:"shift_count"`
	DriverDayCount           int     `json:"driver_day_count"`
	TonsSum                  float64 `json:"tons_sum"`
	DriverCount              int     `json:"driver_count"`
	UserCount                int     `json:"user_count"`
	ForemanCount             int     `json:"foreman_count"`
	MaterialTransactionCount int     `json:"material_transaction_count"`
	JobProductionPlanCount   int     `json:"job_production_plan_count"`
	NotificationCount        int     `json:"notification_count"`
	BroadcastMessageCount    int     `json:"broadcast_message_count"`
	TimeCardCount            int     `json:"time_card_count"`
	InvoiceCount             int     `json:"invoice_count"`
	TruckerCount             int     `json:"trucker_count"`
	IncidentCount            int     `json:"incident_count"`
	TripMilesAvg             float64 `json:"trip_miles_avg"`
	ActiveBranchCount        int     `json:"active_branch_count"`
	TransportationCostPerTon float64 `json:"transportation_cost_per_ton"`
	NightJobPct              float64 `json:"night_job_pct"`
}

func fetchMarketingMetrics(cmd *cobra.Command, client *api.Client, query url.Values) (marketingMetricsDetails, error) {
	requestBody := map[string]any{
		"data": map[string]any{
			"type": "marketing-metrics",
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return marketingMetricsDetails{}, err
	}

	var body []byte
	if query == nil || len(query) == 0 {
		body, _, err = client.Post(cmd.Context(), "/v1/marketing-metrics", jsonBody)
	} else {
		body, _, err = client.PostWithQuery(cmd.Context(), "/v1/marketing-metrics", query, jsonBody)
	}
	if err != nil {
		if len(body) > 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), string(body))
		}
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return marketingMetricsDetails{}, err
	}

	var resp jsonAPISingleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return marketingMetricsDetails{}, err
	}

	return buildMarketingMetricsDetails(resp), nil
}

func buildMarketingMetricsDetails(resp jsonAPISingleResponse) marketingMetricsDetails {
	attrs := resp.Data.Attributes

	return marketingMetricsDetails{
		ID:                       resp.Data.ID,
		ShiftCount:               intAttrWithFallback(attrs, "shift-count", "shift_count"),
		DriverDayCount:           intAttrWithFallback(attrs, "driver-day-count", "driver_day_count"),
		TonsSum:                  floatAttrWithFallback(attrs, "tons-sum", "tons_sum"),
		DriverCount:              intAttrWithFallback(attrs, "driver-count", "driver_count"),
		UserCount:                intAttrWithFallback(attrs, "user-count", "user_count"),
		ForemanCount:             intAttrWithFallback(attrs, "foreman-count", "foreman_count"),
		MaterialTransactionCount: intAttrWithFallback(attrs, "material-transaction-count", "material_transaction_count"),
		JobProductionPlanCount:   intAttrWithFallback(attrs, "job-production-plan-count", "job_production_plan_count"),
		NotificationCount:        intAttrWithFallback(attrs, "notification-count", "notification_count"),
		BroadcastMessageCount:    intAttrWithFallback(attrs, "broadcast-message-count", "broadcast_message_count"),
		TimeCardCount:            intAttrWithFallback(attrs, "time-card-count", "time_card_count"),
		InvoiceCount:             intAttrWithFallback(attrs, "invoice-count", "invoice_count"),
		TruckerCount:             intAttrWithFallback(attrs, "trucker-count", "trucker_count"),
		IncidentCount:            intAttrWithFallback(attrs, "incident-count", "incident_count"),
		TripMilesAvg:             floatAttrWithFallback(attrs, "trip-miles-avg", "trip_miles_avg"),
		ActiveBranchCount:        intAttrWithFallback(attrs, "active-branch-count", "active_branch_count"),
		TransportationCostPerTon: floatAttrWithFallback(attrs, "transportation-cost-per-ton", "transportation_cost_per_ton"),
		NightJobPct:              floatAttrWithFallback(attrs, "night-job-pct", "night_job_pct"),
	}
}

func buildMarketingMetricsListRow(details marketingMetricsDetails) marketingMetricsListRow {
	return marketingMetricsListRow{
		ID:                details.ID,
		ShiftCount:        details.ShiftCount,
		DriverDayCount:    details.DriverDayCount,
		TonsSum:           details.TonsSum,
		DriverCount:       details.DriverCount,
		UserCount:         details.UserCount,
		TruckerCount:      details.TruckerCount,
		IncidentCount:     details.IncidentCount,
		ActiveBranchCount: details.ActiveBranchCount,
	}
}

func intAttrWithFallback(attrs map[string]any, key, fallback string) int {
	if attrs == nil {
		return 0
	}
	if _, ok := attrs[key]; ok {
		return intAttr(attrs, key)
	}
	return intAttr(attrs, fallback)
}

func floatAttrWithFallback(attrs map[string]any, key, fallback string) float64 {
	if attrs == nil {
		return 0
	}
	if _, ok := attrs[key]; ok {
		return floatAttr(attrs, key)
	}
	return floatAttr(attrs, fallback)
}
