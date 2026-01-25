package cli

import "github.com/spf13/cobra"

var doJobScheduleShiftsCmd = &cobra.Command{
	Use:     "job-schedule-shifts",
	Aliases: []string{"job-schedule-shift"},
	Short:   "Manage job schedule shifts",
	Long: `Create, update, and delete job schedule shifts.

Job schedule shifts define scheduled work windows tied to jobs.

Commands:
  create    Create a new job schedule shift
  update    Update an existing job schedule shift
  delete    Delete a job schedule shift`,
}

func init() {
	doCmd.AddCommand(doJobScheduleShiftsCmd)
}
