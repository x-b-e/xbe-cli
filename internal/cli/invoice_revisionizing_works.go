package cli

import "github.com/spf13/cobra"

var invoiceRevisionizingWorksCmd = &cobra.Command{
	Use:   "invoice-revisionizing-works",
	Short: "View invoice revisionizing work",
	Long: `View invoice revisionizing work items.

These records track bulk invoice revisionizing requests and outcomes.

Commands:
  list    List revisionizing work records with filtering
  show    Show revisionizing work details`,
	Example: `  # List invoice revisionizing work
  xbe view invoice-revisionizing-works list

  # Filter by broker
  xbe view invoice-revisionizing-works list --broker 123

  # Show revisionizing work details
  xbe view invoice-revisionizing-works show 456`,
}

func init() {
	viewCmd.AddCommand(invoiceRevisionizingWorksCmd)
}
