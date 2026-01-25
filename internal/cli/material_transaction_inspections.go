package cli

import "github.com/spf13/cobra"

var materialTransactionInspectionsCmd = &cobra.Command{
	Use:   "material-transaction-inspections",
	Short: "Browse and view material transaction inspections",
	Long: `Browse and view material transaction inspections.

Material transaction inspections capture delivery-site checks and inspection
outcomes tied to a material transaction.

Statuses:
  open    Inspection is in progress
  closed  Inspection is finalized

Strategies:
  delivery_site_personnel  Inspected by delivery site personnel

Commands:
  list    List inspections with filtering
  show    Show inspection details`,
	Example: `  # List recent inspections
  xbe view material-transaction-inspections list

  # Filter by status
  xbe view material-transaction-inspections list --status open

  # Show an inspection
  xbe view material-transaction-inspections show 123`,
}

func init() {
	viewCmd.AddCommand(materialTransactionInspectionsCmd)
}
