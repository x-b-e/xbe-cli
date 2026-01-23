package cli

import "github.com/spf13/cobra"

var materialTransactionInspectionRejectionsCmd = &cobra.Command{
	Use:     "material-transaction-inspection-rejections",
	Aliases: []string{"material-transaction-inspection-rejection"},
	Short:   "View material transaction inspection rejections",
	Long: `View material transaction inspection rejections.

Material transaction inspection rejections record rejected quantities and
notes for inspection results.

Commands:
  list    List material transaction inspection rejections with filtering
  show    Show material transaction inspection rejection details`,
	Example: `  # List inspection rejections
  xbe view material-transaction-inspection-rejections list

  # Filter by inspection
  xbe view material-transaction-inspection-rejections list --material-transaction-inspection 123

  # Show a rejection
  xbe view material-transaction-inspection-rejections show 456`,
}

func init() {
	viewCmd.AddCommand(materialTransactionInspectionRejectionsCmd)
}
