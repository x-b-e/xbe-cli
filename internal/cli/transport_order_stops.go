package cli

import "github.com/spf13/cobra"

var transportOrderStopsCmd = &cobra.Command{
	Use:     "transport-order-stops",
	Aliases: []string{"transport-order-stop"},
	Short:   "View transport order stops",
	Long: `View transport order stops.

Transport order stops define pickup and delivery points for transport orders,
including scheduling windows and stop metadata.

Commands:
  list  List transport order stops
  show  Show transport order stop details`,
}

func init() {
	viewCmd.AddCommand(transportOrderStopsCmd)
}
