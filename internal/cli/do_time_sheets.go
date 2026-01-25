package cli

import "github.com/spf13/cobra"

var doTimeSheetsCmd = &cobra.Command{
	Use:     "time-sheets",
	Aliases: []string{"time-sheet"},
	Short:   "Manage time sheets",
	Long: `Manage time sheets.

Time sheets record worked time for shift sets, crew requirements, and work orders.

Commands:
  create    Create a time sheet
  update    Update a time sheet
  delete    Delete a time sheet`,
}

func init() {
	doCmd.AddCommand(doTimeSheetsCmd)
}
