package cli

import "github.com/spf13/cobra"

var timeCardTimeChangesCmd = &cobra.Command{
	Use:   "time-card-time-changes",
	Short: "Browse time card time changes",
	Long: `Browse time card time changes on the XBE platform.

Time card time changes track requested adjustments to time card times.

Commands:
  list    List time card time changes
  show    Show time card time change details`,
	Example: `  # List time card time changes
  xbe view time-card-time-changes list

  # Show a time card time change
  xbe view time-card-time-changes show 123`,
}

func init() {
	viewCmd.AddCommand(timeCardTimeChangesCmd)
}
