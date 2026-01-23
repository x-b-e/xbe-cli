package cli

import "github.com/spf13/cobra"

var materialTransactionStatusChangesCmd = &cobra.Command{
	Use:     "material-transaction-status-changes",
	Aliases: []string{"material-transaction-status-change"},
	Short:   "View material transaction status changes",
	Long: `View material transaction status changes.

Material transaction status changes capture status updates, who made the change,
when the change occurred, and any optional comments.

Commands:
  list    List material transaction status changes with filtering
  show    Show material transaction status change details`,
	Example: `  # List status changes
  xbe view material-transaction-status-changes list

  # Filter by material transaction
  xbe view material-transaction-status-changes list --material-transaction 123

  # Filter by status
  xbe view material-transaction-status-changes list --status accepted

  # Show a status change
  xbe view material-transaction-status-changes show 456`,
}

func init() {
	viewCmd.AddCommand(materialTransactionStatusChangesCmd)
}
