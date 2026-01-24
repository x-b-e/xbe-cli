package cli

import "github.com/spf13/cobra"

var rawMaterialTransactionImportResultsCmd = &cobra.Command{
	Use:     "raw-material-transaction-import-results",
	Aliases: []string{"raw-material-transaction-import-result"},
	Short:   "Browse raw material transaction import results",
	Long: `Browse raw material transaction import results on the XBE platform.

Raw material transaction import results summarize recent import activity and
connection status for raw material transaction importers.

Commands:
  list    List raw material transaction import results
  show    Show raw material transaction import result details`,
	Example: `  # List import results
  xbe view raw-material-transaction-import-results list

  # Show import result details
  xbe view raw-material-transaction-import-results show 123`,
}

func init() {
	viewCmd.AddCommand(rawMaterialTransactionImportResultsCmd)
}
