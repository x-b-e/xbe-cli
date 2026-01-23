package cli

import "github.com/spf13/cobra"

var lineupScenarioLineupJobScheduleShiftsCmd = &cobra.Command{
	Use:     "lineup-scenario-lineup-job-schedule-shifts",
	Aliases: []string{"lineup-scenario-lineup-job-schedule-shift"},
	Short:   "Browse lineup scenario lineup job schedule shifts",
	Long: `Browse lineup scenario lineup job schedule shifts on the XBE platform.

Lineup scenario lineup job schedule shifts connect lineup scenarios to lineup job
schedule shifts.

Commands:
  list    List lineup scenario lineup job schedule shifts with filtering and pagination
  show    Show lineup scenario lineup job schedule shift details`,
	Example: `  # List lineup scenario lineup job schedule shifts
  xbe view lineup-scenario-lineup-job-schedule-shifts list

  # Show a lineup scenario lineup job schedule shift
  xbe view lineup-scenario-lineup-job-schedule-shifts show 123

  # Output as JSON
  xbe view lineup-scenario-lineup-job-schedule-shifts list --json`,
}

func init() {
	viewCmd.AddCommand(lineupScenarioLineupJobScheduleShiftsCmd)
}
