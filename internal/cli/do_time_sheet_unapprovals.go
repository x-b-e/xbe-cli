package cli

import "github.com/spf13/cobra"

var doTimeSheetUnapprovalsCmd = &cobra.Command{
	Use:     "time-sheet-unapprovals",
	Aliases: []string{"time-sheet-unapproval"},
	Short:   "Unapprove time sheets",
	Long:    "Commands for unapproving time sheets.",
}

func init() {
	doCmd.AddCommand(doTimeSheetUnapprovalsCmd)
}
