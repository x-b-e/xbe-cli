package cli

import "github.com/spf13/cobra"

var tenderJobScheduleShiftDriversCmd = &cobra.Command{
	Use:     "tender-job-schedule-shift-drivers",
	Aliases: []string{"tender-job-schedule-shift-driver"},
	Short:   "Browse tender job schedule shift drivers",
	Long: `Browse tender job schedule shift drivers.

Tender job schedule shift drivers link driver users to tender job schedule shifts.

Commands:
  list    List shift drivers with filters
  show    Show shift driver details`,
	Example: `  # List shift drivers
  xbe view tender-job-schedule-shift-drivers list

  # Filter by tender job schedule shift
  xbe view tender-job-schedule-shift-drivers list --tender-job-schedule-shift 123

  # Filter by user
  xbe view tender-job-schedule-shift-drivers list --user 456

  # Show shift driver details
  xbe view tender-job-schedule-shift-drivers show 789`,
}

func init() {
	viewCmd.AddCommand(tenderJobScheduleShiftDriversCmd)
}
