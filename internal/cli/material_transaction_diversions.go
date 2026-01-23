package cli

import "github.com/spf13/cobra"

var materialTransactionDiversionsCmd = &cobra.Command{
	Use:     "material-transaction-diversions",
	Aliases: []string{"material-transaction-diversion"},
	Short:   "Browse material transaction diversions",
	Long: `Browse material transaction diversions.

Diversions reroute a material transaction to a new job site or delivery date,
optionally with driver instructions and explicit diverted tonnage.

Commands:
  list    List material transaction diversions with filtering
  show    Show diversion details`,
	Example: `  # List diversions
  xbe view material-transaction-diversions list

  # Filter by material transaction
  xbe view material-transaction-diversions list --material-transaction 123

  # View a diversion
  xbe view material-transaction-diversions show 456`,
}

func init() {
	viewCmd.AddCommand(materialTransactionDiversionsCmd)
}
