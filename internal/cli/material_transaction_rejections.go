package cli

import "github.com/spf13/cobra"

var materialTransactionRejectionsCmd = &cobra.Command{
	Use:     "material-transaction-rejections",
	Aliases: []string{"material-transaction-rejection"},
	Short:   "View material transaction rejections",
	Long: `View material transaction rejections.

Rejections record a status change to rejected for a material transaction and may
include a comment.

Commands:
  list    List material transaction rejections
  show    Show material transaction rejection details`,
	Example: `  # List rejections
  xbe view material-transaction-rejections list

  # Show a rejection
  xbe view material-transaction-rejections show 123`,
}

func init() {
	viewCmd.AddCommand(materialTransactionRejectionsCmd)
}
