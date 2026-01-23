package cli

import "github.com/spf13/cobra"

var doTimeCardSubmissionsCmd = &cobra.Command{
	Use:     "time-card-submissions",
	Aliases: []string{"time-card-submission"},
	Short:   "Submit time cards",
	Long: `Submit time cards.

Submissions move editing or rejected time cards to submitted status.

Commands:
  create    Submit a time card`,
	Example: `  # Submit a time card
  xbe do time-card-submissions create --time-card 123 --comment "Submitted"`,
}

func init() {
	doCmd.AddCommand(doTimeCardSubmissionsCmd)
}
