package cli

import "github.com/spf13/cobra"

var timeSheetNoShowsCmd = &cobra.Command{
	Use:     "time-sheet-no-shows",
	Aliases: []string{"time-sheet-no-show"},
	Short:   "View time sheet no-shows",
	Long: `View time sheet no-shows.

No-shows record when a time sheet is marked as a no-show with a reason.

Commands:
  list    List time sheet no-shows
  show    Show time sheet no-show details`,
	Example: `  # List time sheet no-shows
  xbe view time-sheet-no-shows list

  # Show a time sheet no-show
  xbe view time-sheet-no-shows show 123`,
}

func init() {
	viewCmd.AddCommand(timeSheetNoShowsCmd)
}
