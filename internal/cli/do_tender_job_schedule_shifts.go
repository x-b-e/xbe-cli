package cli

import "github.com/spf13/cobra"

var doTenderJobScheduleShiftsCmd = &cobra.Command{
	Use:     "tender-job-schedule-shifts",
	Aliases: []string{"tender-job-schedule-shift"},
	Short:   "Manage tender job schedule shifts",
	Long: `Manage tender job schedule shifts.

Commands:
  create  Create a tender job schedule shift
  update  Update a tender job schedule shift
  delete  Delete a tender job schedule shift`,
	Example: `  # Create a tender job schedule shift
  xbe do tender-job-schedule-shifts create --tender-type broker-tenders --tender-id 123 --job-schedule-shift 456 --material-transaction-status open

  # Update a shift
  xbe do tender-job-schedule-shifts update 789 --seller-operations-contact 456

  # Delete a shift
  xbe do tender-job-schedule-shifts delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doTenderJobScheduleShiftsCmd)
}
