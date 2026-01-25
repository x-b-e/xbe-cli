package cli

import "github.com/spf13/cobra"

var doTimeCardApprovalsCmd = &cobra.Command{
	Use:     "time-card-approvals",
	Aliases: []string{"time-card-approval"},
	Short:   "Approve time cards",
	Long: `Approve time cards.

Approvals move submitted time cards to approved status.

Commands:
  create    Approve a time card`,
	Example: `  # Approve a time card
  xbe do time-card-approvals create --time-card 123 --comment "Approved"`,
}

func init() {
	doCmd.AddCommand(doTimeCardApprovalsCmd)
}
