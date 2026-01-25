package cli

import "github.com/spf13/cobra"

var doTimeSheetApprovalsCmd = &cobra.Command{
	Use:     "time-sheet-approvals",
	Aliases: []string{"time-sheet-approval"},
	Short:   "Approve time sheets",
	Long:    "Commands for approving time sheets.",
}

func init() {
	doCmd.AddCommand(doTimeSheetApprovalsCmd)
}
