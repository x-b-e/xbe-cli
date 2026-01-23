package cli

import "github.com/spf13/cobra"

var driverAssignmentAcknowledgementsCmd = &cobra.Command{
	Use:     "driver-assignment-acknowledgements",
	Aliases: []string{"driver-assignment-acknowledgement"},
	Short:   "Browse driver assignment acknowledgements",
	Long: `Browse driver assignment acknowledgements.

Driver assignment acknowledgements record when a driver (or authorized user)
acknowledges a tender job schedule shift assignment.

Commands:
  list    List acknowledgements with filtering and pagination
  show    Show full details of an acknowledgement`,
	Example: `  # List acknowledgements
  xbe view driver-assignment-acknowledgements list

  # Filter by tender job schedule shift
  xbe view driver-assignment-acknowledgements list --tender-job-schedule-shift 123

  # Filter by driver
  xbe view driver-assignment-acknowledgements list --driver 456

  # Show an acknowledgement
  xbe view driver-assignment-acknowledgements show 789`,
}

func init() {
	viewCmd.AddCommand(driverAssignmentAcknowledgementsCmd)
}
