package cli

import "github.com/spf13/cobra"

var doLineupJobScheduleShiftsCmd = &cobra.Command{
	Use:   "lineup-job-schedule-shifts",
	Short: "Manage lineup job schedule shifts",
	Long: `Create, update, and delete lineup job schedule shifts.

Lineup job schedule shifts connect job schedule shifts to lineups, capturing
trucker assignments and dispatch readiness.

Commands:
  create    Create a lineup job schedule shift
  update    Update a lineup job schedule shift
  delete    Delete a lineup job schedule shift`,
}

func init() {
	doCmd.AddCommand(doLineupJobScheduleShiftsCmd)
}
