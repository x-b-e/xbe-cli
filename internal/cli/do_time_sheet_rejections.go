package cli

import "github.com/spf13/cobra"

var doTimeSheetRejectionsCmd = &cobra.Command{
	Use:     "time-sheet-rejections",
	Aliases: []string{"time-sheet-rejection"},
	Short:   "Reject time sheets",
	Long: `Reject time sheets.

Time sheet rejections transition submitted time sheets to rejected.

Commands:
  create    Reject a time sheet`,
}

func init() {
	doCmd.AddCommand(doTimeSheetRejectionsCmd)
}
