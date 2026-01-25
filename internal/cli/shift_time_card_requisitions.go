package cli

import "github.com/spf13/cobra"

var shiftTimeCardRequisitionsCmd = &cobra.Command{
	Use:     "shift-time-card-requisitions",
	Aliases: []string{"shift-time-card-requisition"},
	Short:   "View shift time card requisitions",
	Long:    "Commands for viewing shift time card requisitions.",
}

func init() {
	viewCmd.AddCommand(shiftTimeCardRequisitionsCmd)
}
