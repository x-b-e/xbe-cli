package cli

import "github.com/spf13/cobra"

var paveFrameActualStatisticsCmd = &cobra.Command{
	Use:   "pave-frame-actual-statistics",
	Short: "View pave frame actual statistics",
	Long: `View pave frame actual statistics on the XBE platform.

Pave frame actual statistics summarize historical paving window conditions
near a specified location using configurable temperature, precipitation,
and work-day thresholds.

Commands:
  list    List pave frame actual statistics
  show    Show pave frame actual statistic details`,
	Example: `  # List statistics
  xbe view pave-frame-actual-statistics list

  # Show a statistic
  xbe view pave-frame-actual-statistics show 123

  # Output JSON
  xbe view pave-frame-actual-statistics list --json`,
}

func init() {
	viewCmd.AddCommand(paveFrameActualStatisticsCmd)
}
