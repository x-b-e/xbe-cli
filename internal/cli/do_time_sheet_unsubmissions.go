package cli

import "github.com/spf13/cobra"

var doTimeSheetUnsubmissionsCmd = &cobra.Command{
	Use:   "time-sheet-unsubmissions",
	Short: "Unsubmit time sheets",
	Long: `Unsubmit time sheets.

Unsubmitting a time sheet moves it from submitted to editing status.

Commands:
  create  Unsubmit a time sheet`,
	Example: `  # Unsubmit a time sheet
  xbe do time-sheet-unsubmissions create --time-sheet 123

  # Unsubmit with a comment
  xbe do time-sheet-unsubmissions create --time-sheet 123 --comment "Needs edits"`,
}

func init() {
	doCmd.AddCommand(doTimeSheetUnsubmissionsCmd)
}
