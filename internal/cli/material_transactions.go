package cli

import "github.com/spf13/cobra"

var materialTransactionsCmd = &cobra.Command{
	Use:   "material-transactions",
	Short: "Browse and view material transactions",
	Long: `Browse and view material transactions on the XBE platform.

Material transactions track the movement of materials (asphalt, concrete, aggregates,
etc.) between sites. Each transaction represents a load delivered, with details about
quantities, timing, and the parties involved.

Statuses:
  editing      Draft state, user can modify
  submitted    Submitted for approval
  accepted     Approved and final
  rejected     Rejected, can be resubmitted
  unmatched    System couldn't match to a job/shift
  denied       Explicitly denied (final)
  invalidated  Deleted/invalidated (final)

Commands:
  list    List material transactions with filtering
  show    View the full details of a specific transaction`,
	Example: `  # List recent material transactions
  xbe view material-transactions list

  # List transactions for a specific date
  xbe view material-transactions list --date 2025-01-15

  # List transactions in a date range
  xbe view material-transactions list --date-min 2025-01-01 --date-max 2025-01-31

  # Filter by status
  xbe view material-transactions list --status accepted

  # Search by ticket number
  xbe view material-transactions list --ticket-number T12345

  # View a specific transaction
  xbe view material-transactions show 123`,
}

func init() {
	viewCmd.AddCommand(materialTransactionsCmd)
}
