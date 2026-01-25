package cli

import "github.com/spf13/cobra"

var doLineupScenarioTrailerLineupJobScheduleShiftsCmd = &cobra.Command{
	Use:     "lineup-scenario-trailer-lineup-job-schedule-shifts",
	Aliases: []string{"lineup-scenario-trailer-lineup-job-schedule-shift"},
	Short:   "Manage lineup scenario trailer lineup job schedule shifts",
	Long: `Create, update, and delete lineup scenario trailer lineup job schedule shifts.

These records link lineup scenario trailers to lineup job schedule shifts within
lineup scenarios.

Commands:
  create    Create a record
  update    Update a record
  delete    Delete a record`,
}

func init() {
	doCmd.AddCommand(doLineupScenarioTrailerLineupJobScheduleShiftsCmd)
}
