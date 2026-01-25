package cli

import "github.com/spf13/cobra"

var doTimeCardRejectionsCmd = &cobra.Command{
	Use:     "time-card-rejections",
	Aliases: []string{"time-card-rejection"},
	Short:   "Reject time cards",
	Long: `Reject time cards.

Rejections move submitted time cards to rejected status.

Commands:
  create    Reject a time card`,
	Example: `  # Reject a time card
  xbe do time-card-rejections create --time-card 123 --comment "Missing ticket"`,
}

func init() {
	doCmd.AddCommand(doTimeCardRejectionsCmd)
}
