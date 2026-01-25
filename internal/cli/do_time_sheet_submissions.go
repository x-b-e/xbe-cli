package cli

import "github.com/spf13/cobra"

var doTimeSheetSubmissionsCmd = &cobra.Command{
	Use:   "time-sheet-submissions",
	Short: "Submit time sheets",
	Long: `Submit time sheets.

Submitting a time sheet moves it from editing or rejected to submitted.
Time sheets must have a duration before submission.

Commands:
  create  Submit a time sheet`,
	Example: `  # Submit a time sheet
  xbe do time-sheet-submissions create --time-sheet 123

  # Submit with a comment
  xbe do time-sheet-submissions create --time-sheet 123 --comment "Ready for approval"`,
}

func init() {
	doCmd.AddCommand(doTimeSheetSubmissionsCmd)
}
