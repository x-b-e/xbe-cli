package cli

import "github.com/spf13/cobra"

var driverAssignmentRefusalsCmd = &cobra.Command{
	Use:     "driver-assignment-refusals",
	Aliases: []string{"driver-assignment-refusal"},
	Short:   "Browse driver assignment refusals",
	Long: `Browse driver assignment refusals.

Driver assignment refusals record when a driver declines a tender job schedule
shift assignment.

Commands:
  list    List refusals with filtering and pagination
  show    Show full details of a refusal`,
	Example: `  # List refusals
  xbe view driver-assignment-refusals list

  # Filter by tender job schedule shift
  xbe view driver-assignment-refusals list --tender-job-schedule-shift 123

  # Filter by driver
  xbe view driver-assignment-refusals list --driver 456

  # Show refusal details
  xbe view driver-assignment-refusals show 789`,
}

func init() {
	viewCmd.AddCommand(driverAssignmentRefusalsCmd)
}
