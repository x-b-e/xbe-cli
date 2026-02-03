package cli

import "github.com/spf13/cobra"

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Aggregate data for analysis (pivot tables, totals, statistics)",
	Long: `Aggregate large datasets for analysis.

Summary commands work like pivot tables, grouping and aggregating data to produce
totals, averages, and other statistics. Use these when you need to analyze trends
or compare metrics across dimensions.

Resources:
  lane-summary                         Aggregate hauling/cycle data by origin, destination, etc.
  material-transaction-summary         Aggregate material transactions by site, customer, date, etc.
  material-site-reading-summary        Aggregate material site readings by time bucket, site, measure, etc.
  shift-summary                        Aggregate shift data by driver, trucker, date, etc.
  job-production-plan-summary          Aggregate job production plan data by customer, project, etc.
  driver-day-summary                   Aggregate driver day data by driver, trucker, date, etc.
  public-praise-summary                Aggregate public praise data by recipient, giver, etc.
  device-location-event-summary        Aggregate device location events by device, user, etc.
  transport-summary                    Aggregate transport data by entity type (orders, plans, etc.)
  transport-order-efficiency-summary   Aggregate transport order efficiency by customer, driver, etc.
  ptp-summary                          Aggregate project transport plans by broker, strategy, etc.
  ptp-driver-summary                   Aggregate PTP driver data by driver, customer, etc.
  ptp-trailer-summary                  Aggregate PTP trailer data by trailer, customer, etc.
  ptp-event-summary                    Aggregate PTP events by event type, broker, etc.
  ptp-event-time-summary               Aggregate PTP event durations by event type, location, etc.
  ptp-expected-event-time-summary      Aggregate PTP expected event time accuracy by event type, lead time, etc.`,
	Example: `  # Summarize hauling data by origin and destination
  xbe summarize lane-summary create --group-by origin,destination --filter broker=123

  # Summarize material transactions by site for a date range
  xbe summarize material-transaction-summary create \
    --group-by material_site \
    --filter broker=123 --filter date_min=2025-01-01

  # Summarize shift data by driver
  xbe summarize shift-summary create --start-on 2025-01-01 --end-on 2025-01-31 \
    --group-by driver --filter broker=123

  # Get totals without grouping
  xbe summarize lane-summary create --group-by "" --filter broker=123`,
	Annotations: map[string]string{"group": GroupCore},
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}
