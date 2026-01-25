package cli

import "github.com/spf13/cobra"

var doMaterialTransactionDiversionsCmd = &cobra.Command{
	Use:     "material-transaction-diversions",
	Aliases: []string{"material-transaction-diversion"},
	Short:   "Manage material transaction diversions",
	Long: `Create, update, or delete material transaction diversions.

Diversions reroute a material transaction to a new job site or delivery date.

Commands:
  create    Create a material transaction diversion
  update    Update a material transaction diversion
  delete    Delete a material transaction diversion`,
	Example: `  # Create a diversion
  xbe do material-transaction-diversions create --material-transaction 123 --new-delivery-date 2025-01-02

  # Update a diversion
  xbe do material-transaction-diversions update 456 --driver-instructions "Call dispatch"

  # Delete a diversion
  xbe do material-transaction-diversions delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doMaterialTransactionDiversionsCmd)
}
