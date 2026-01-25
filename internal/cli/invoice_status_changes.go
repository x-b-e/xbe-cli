package cli

import "github.com/spf13/cobra"

var invoiceStatusChangesCmd = &cobra.Command{
	Use:     "invoice-status-changes",
	Aliases: []string{"invoice-status-change"},
	Short:   "View invoice status changes",
	Long: `View invoice status changes.

Status changes capture updates to an invoice status, along with who made
the change, when it happened, and any comment.

Commands:
  list    List invoice status changes
  show    Show invoice status change details`,
	Example: `  # List invoice status changes
  xbe view invoice-status-changes list

  # Show an invoice status change
  xbe view invoice-status-changes show 123`,
}

func init() {
	viewCmd.AddCommand(invoiceStatusChangesCmd)
}
