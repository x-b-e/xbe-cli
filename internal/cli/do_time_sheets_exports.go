package cli

import "github.com/spf13/cobra"

var doTimeSheetsExportsCmd = &cobra.Command{
	Use:     "time-sheets-exports",
	Aliases: []string{"time-sheets-export"},
	Short:   "Export time sheets",
	Long: `Export time sheets.

Time sheets exports format approved time sheets through organization formatters
and generate export files.

Commands:
  create    Create a time sheets export`,
	Example: `  # Create a time sheets export
  xbe do time-sheets-exports create --organization-formatter 123 --time-sheet-ids 456,789`,
}

func init() {
	doCmd.AddCommand(doTimeSheetsExportsCmd)
}
