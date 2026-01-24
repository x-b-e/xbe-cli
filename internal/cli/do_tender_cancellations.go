package cli

import "github.com/spf13/cobra"

var doTenderCancellationsCmd = &cobra.Command{
	Use:     "tender-cancellations",
	Aliases: []string{"tender-cancellation"},
	Short:   "Cancel tenders",
	Long: `Cancel tenders.

Cancellations move tenders to cancelled status.

Commands:
  create    Cancel a tender`,
	Example: `  # Cancel a tender
  xbe do tender-cancellations create --tender 123 --comment "Cancelled"`,
}

func init() {
	doCmd.AddCommand(doTenderCancellationsCmd)
}
