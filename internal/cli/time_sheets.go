package cli

import "github.com/spf13/cobra"

var timeSheetsCmd = &cobra.Command{
	Use:     "time-sheets",
	Aliases: []string{"time-sheet"},
	Short:   "View time sheets",
	Long: `View time sheets.

Time sheets track worked time for shift sets, crew requirements, and work orders.

Commands:
  list    List time sheets
  show    Show time sheet details`,
	Example: `  # List time sheets
  xbe view time-sheets list

  # Show a time sheet
  xbe view time-sheets show 123

  # Output JSON
  xbe view time-sheets list --json`,
}

func init() {
	viewCmd.AddCommand(timeSheetsCmd)
}
