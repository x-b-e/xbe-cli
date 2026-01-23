package cli

import "github.com/spf13/cobra"

var doDriverAssignmentRefusalsCmd = &cobra.Command{
	Use:     "driver-assignment-refusals",
	Aliases: []string{"driver-assignment-refusal"},
	Short:   "Manage driver assignment refusals",
	Long:    "Commands for creating driver assignment refusals.",
}

func init() {
	doCmd.AddCommand(doDriverAssignmentRefusalsCmd)
}
