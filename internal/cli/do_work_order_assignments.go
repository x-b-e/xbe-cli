package cli

import "github.com/spf13/cobra"

var doWorkOrderAssignmentsCmd = &cobra.Command{
	Use:     "work-order-assignments",
	Aliases: []string{"work-order-assignment"},
	Short:   "Manage work order assignments",
	Long:    "Commands for creating, updating, and deleting work order assignments.",
}

func init() {
	doCmd.AddCommand(doWorkOrderAssignmentsCmd)
}
