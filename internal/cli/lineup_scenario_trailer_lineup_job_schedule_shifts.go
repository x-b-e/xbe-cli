package cli

import "github.com/spf13/cobra"

var lineupScenarioTrailerLineupJobScheduleShiftsCmd = &cobra.Command{
	Use:     "lineup-scenario-trailer-lineup-job-schedule-shifts",
	Aliases: []string{"lineup-scenario-trailer-lineup-job-schedule-shift"},
	Short:   "Browse lineup scenario trailer lineup job schedule shifts",
	Long: `Browse lineup scenario trailer lineup job schedule shifts.

These records link lineup scenario trailers to lineup job schedule shifts within
lineup scenarios, with optional site distance minutes.

Commands:
  list    List records with filters
  show    Show record details`,
	Example: `  # List records
  xbe view lineup-scenario-trailer-lineup-job-schedule-shifts list

  # Show details for a record
  xbe view lineup-scenario-trailer-lineup-job-schedule-shifts show 123`,
}

func init() {
	viewCmd.AddCommand(lineupScenarioTrailerLineupJobScheduleShiftsCmd)
}
