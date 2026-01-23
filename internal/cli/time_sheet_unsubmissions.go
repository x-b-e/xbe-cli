package cli

import "github.com/spf13/cobra"

var timeSheetUnsubmissionsCmd = &cobra.Command{
	Use:   "time-sheet-unsubmissions",
	Short: "View time sheet unsubmissions",
	Long: `View time sheet unsubmissions.

Time sheet unsubmissions move a time sheet from submitted to editing.

Commands:
  list    List time sheet unsubmissions
  show    Show time sheet unsubmission details`,
	Example: `  # List time sheet unsubmissions
  xbe view time-sheet-unsubmissions list

  # Show a time sheet unsubmission
  xbe view time-sheet-unsubmissions show 123`,
}

func init() {
	viewCmd.AddCommand(timeSheetUnsubmissionsCmd)
}
