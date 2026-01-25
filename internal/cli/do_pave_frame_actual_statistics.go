package cli

import "github.com/spf13/cobra"

var doPaveFrameActualStatisticsCmd = &cobra.Command{
	Use:   "pave-frame-actual-statistics",
	Short: "Manage pave frame actual statistics",
	Long: `Create, update, and delete pave frame actual statistics.

Statistics calculate paving windows based on temperature, precipitation,
and work-day thresholds near a location.

Commands:
  create  Create a pave frame actual statistic
  update  Update a pave frame actual statistic
  delete  Delete a pave frame actual statistic`,
	Example: `  # Create a statistic
  xbe do pave-frame-actual-statistics create --latitude 41.88 --longitude -87.62 \
    --hour-minimum-temp-f 45 --hour-maximum-precip-in 0.1 --window-minimum-paving-hour-pct 0.6 \
    --agg-level month --work-days 1,2,3,4,5

  # Update a statistic
  xbe do pave-frame-actual-statistics update 123 --window night --agg-level week`,
}

func init() {
	doCmd.AddCommand(doPaveFrameActualStatisticsCmd)
}
