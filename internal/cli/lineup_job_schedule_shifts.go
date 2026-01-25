package cli

import "github.com/spf13/cobra"

var lineupJobScheduleShiftsCmd = &cobra.Command{
	Use:   "lineup-job-schedule-shifts",
	Short: "View lineup job schedule shifts",
	Long: `Browse lineup job schedule shifts on the XBE platform.

Lineup job schedule shifts tie job schedule shifts to lineups, capturing
trucker assignments, trailer classifications, and dispatch readiness.

Commands:
  list    List lineup job schedule shifts
  show    Show lineup job schedule shift details`,
	Example: `  # List lineup job schedule shifts
  xbe view lineup-job-schedule-shifts list

  # Show a lineup job schedule shift
  xbe view lineup-job-schedule-shifts show 123`,
}

func init() {
	viewCmd.AddCommand(lineupJobScheduleShiftsCmd)
}
