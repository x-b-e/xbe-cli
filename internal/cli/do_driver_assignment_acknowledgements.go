package cli

import "github.com/spf13/cobra"

var doDriverAssignmentAcknowledgementsCmd = &cobra.Command{
	Use:     "driver-assignment-acknowledgements",
	Aliases: []string{"driver-assignment-acknowledgement"},
	Short:   "Manage driver assignment acknowledgements",
	Long: `Commands for creating driver assignment acknowledgements.

Note: Driver assignment acknowledgements cannot be updated or deleted via the API.`,
}

func init() {
	doCmd.AddCommand(doDriverAssignmentAcknowledgementsCmd)
}
