package cli

import "github.com/spf13/cobra"

var materialTransactionPreloadsCmd = &cobra.Command{
	Use:   "material-transaction-preloads",
	Short: "Browse material transaction preloads",
	Long: `Browse material transaction preloads on the XBE platform.

Material transaction preloads link a trailer to a material transaction before loading,
tracking when the preload occurred and the estimated preload minutes.

Commands:
  list    List material transaction preloads
  show    Show material transaction preload details`,
	Example: `  # List preloads
  xbe view material-transaction-preloads list

  # Show preload details
  xbe view material-transaction-preloads show 123`,
}

func init() {
	viewCmd.AddCommand(materialTransactionPreloadsCmd)
}
