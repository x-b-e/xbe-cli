package cli

import "github.com/spf13/cobra"

var rateAdjustmentsCmd = &cobra.Command{
	Use:     "rate-adjustments",
	Aliases: []string{"rate-adjustment"},
	Short:   "View rate adjustments",
	Long: `View rate adjustments on the XBE platform.

Rate adjustments connect a rate to a cost index and define how pricing
changes as the index moves.

Commands:
  list    List rate adjustments
  show    Show rate adjustment details`,
	Example: `  # List rate adjustments
  xbe view rate-adjustments list

  # Show a rate adjustment
  xbe view rate-adjustments show 123`,
}

func init() {
	viewCmd.AddCommand(rateAdjustmentsCmd)
}
