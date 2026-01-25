package cli

import "github.com/spf13/cobra"

var doShiftTimeCardRequisitionsCmd = &cobra.Command{
	Use:     "shift-time-card-requisitions",
	Aliases: []string{"shift-time-card-requisition"},
	Short:   "Manage shift time card requisitions",
	Long:    "Commands for creating shift time card requisitions.",
}

func init() {
	doCmd.AddCommand(doShiftTimeCardRequisitionsCmd)
}
