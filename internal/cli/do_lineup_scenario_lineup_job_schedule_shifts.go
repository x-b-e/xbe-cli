package cli

import "github.com/spf13/cobra"

var doLineupScenarioLineupJobScheduleShiftsCmd = &cobra.Command{
	Use:     "lineup-scenario-lineup-job-schedule-shifts",
	Aliases: []string{"lineup-scenario-lineup-job-schedule-shift"},
	Short:   "Manage lineup scenario lineup job schedule shifts",
	Long: `Manage lineup scenario lineup job schedule shifts.

Commands:
  create    Create a lineup scenario lineup job schedule shift
  delete    Delete a lineup scenario lineup job schedule shift`,
	Example: `  # Create a lineup scenario lineup job schedule shift
  xbe do lineup-scenario-lineup-job-schedule-shifts create --lineup-scenario 123 --lineup-job-schedule-shift 456

  # Delete a lineup scenario lineup job schedule shift
  xbe do lineup-scenario-lineup-job-schedule-shifts delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doLineupScenarioLineupJobScheduleShiftsCmd)
}
