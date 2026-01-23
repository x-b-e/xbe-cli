package cli

import "github.com/spf13/cobra"

var doTimeCardUnsubmissionsCmd = &cobra.Command{
	Use:   "time-card-unsubmissions",
	Short: "Unsubmit time cards",
	Long: `Unsubmit time cards.

Unsubmitting a time card moves it from submitted to editing status.

Commands:
  create  Unsubmit a time card`,
	Example: `  # Unsubmit a time card
  xbe do time-card-unsubmissions create --time-card 123

  # Unsubmit with a comment
  xbe do time-card-unsubmissions create --time-card 123 --comment "Needs edits"`,
}

func init() {
	doCmd.AddCommand(doTimeCardUnsubmissionsCmd)
}
