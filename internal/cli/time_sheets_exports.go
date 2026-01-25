package cli

import "github.com/spf13/cobra"

var timeSheetsExportsCmd = &cobra.Command{
	Use:     "time-sheets-exports",
	Aliases: []string{"time-sheets-export"},
	Short:   "Browse time sheets exports",
	Long: `Browse time sheets exports.

Time sheets exports capture formatted export files generated from time sheets
and organization formatters.

Commands:
  list    List time sheets exports with filtering
  show    Show time sheets export details`,
	Example: `  # List time sheets exports
  xbe view time-sheets-exports list

  # Show a time sheets export
  xbe view time-sheets-exports show 123`,
}

func init() {
	viewCmd.AddCommand(timeSheetsExportsCmd)
}
