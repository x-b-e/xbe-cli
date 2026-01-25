package cli

import "github.com/spf13/cobra"

var timeSheetRejectionsCmd = &cobra.Command{
	Use:     "time-sheet-rejections",
	Aliases: []string{"time-sheet-rejection"},
	Short:   "View time sheet rejections",
	Long: `View time sheet rejections.

Time sheet rejections transition submitted time sheets to rejected.

Commands:
  list    List time sheet rejections
  show    Show time sheet rejection details`,
	Example: `  # List time sheet rejections
  xbe view time-sheet-rejections list

  # Show a time sheet rejection
  xbe view time-sheet-rejections show 123

  # Output JSON
  xbe view time-sheet-rejections list --json`,
}

func init() {
	viewCmd.AddCommand(timeSheetRejectionsCmd)
}
