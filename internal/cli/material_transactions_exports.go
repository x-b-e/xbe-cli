package cli

import "github.com/spf13/cobra"

var materialTransactionsExportsCmd = &cobra.Command{
	Use:     "material-transactions-exports",
	Aliases: []string{"material-transactions-export"},
	Short:   "Browse material transaction exports",
	Long: `Browse material transaction exports.

Material transaction exports generate formatted files for selected material
transactions using an organization formatter.

Commands:
  list    List material transaction exports
  show    Show material transaction export details`,
	Example: `  # List exports
  xbe view material-transactions-exports list

  # Show export details
  xbe view material-transactions-exports show 123`,
}

func init() {
	viewCmd.AddCommand(materialTransactionsExportsCmd)
}
