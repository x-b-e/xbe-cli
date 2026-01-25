package cli

import "github.com/spf13/cobra"

var tenderJobScheduleShiftsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shifts",
	Aliases: []string{"tender-job-schedule-shift"},
	Short:   "Browse tender job schedule shifts",
	Long: `Browse tender job schedule shifts.

Tender job schedule shifts represent tendered shifts tied to job schedule shifts,
including driver, equipment, and tender assignment details.

Commands:
  list    List tender job schedule shifts with filters
  show    Show tender job schedule shift details`,
	Example: `  # List tender job schedule shifts
  xbe view tender-job-schedule-shifts list

  # Filter by tender
  xbe view tender-job-schedule-shifts list --tender 123

  # Filter by job schedule shift
  xbe view tender-job-schedule-shifts list --job-schedule-shift 456

  # Show shift details
  xbe view tender-job-schedule-shifts show 789`,
}

func init() {
	viewCmd.AddCommand(tenderJobScheduleShiftsCmd)
}
