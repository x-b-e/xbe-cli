package cli

import "github.com/spf13/cobra"

var doTimeSheetNoShowsCmd = &cobra.Command{
	Use:     "time-sheet-no-shows",
	Aliases: []string{"time-sheet-no-show"},
	Short:   "Manage time sheet no-shows",
	Long:    "Commands for managing time sheet no-shows.",
}

func init() {
	doCmd.AddCommand(doTimeSheetNoShowsCmd)
}
