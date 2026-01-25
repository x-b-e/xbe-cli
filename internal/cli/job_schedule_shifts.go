package cli

import "github.com/spf13/cobra"

var jobScheduleShiftsCmd = &cobra.Command{
	Use:     "job-schedule-shifts",
	Aliases: []string{"job-schedule-shift"},
	Short:   "View job schedule shifts",
	Long: `View job schedule shifts on the XBE platform.

Job schedule shifts define scheduled work windows tied to jobs.

Commands:
  list    List job schedule shifts with filtering
  show    Show job schedule shift details`,
	Example: `  # List job schedule shifts
  xbe view job-schedule-shifts list

  # Show a job schedule shift
  xbe view job-schedule-shifts show 123`,
}

func init() {
	viewCmd.AddCommand(jobScheduleShiftsCmd)
}
