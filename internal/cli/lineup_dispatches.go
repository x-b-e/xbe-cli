package cli

import "github.com/spf13/cobra"

var lineupDispatchesCmd = &cobra.Command{
	Use:     "lineup-dispatches",
	Aliases: []string{"lineup-dispatch"},
	Short:   "Browse lineup dispatches",
	Long: `Browse lineup dispatches.

Lineup dispatches represent dispatch events for a lineup, tracking
fulfillment progress and tendering behavior.`,
	Example: `  # List lineup dispatches
  xbe view lineup-dispatches list

  # Filter by lineup
  xbe view lineup-dispatches list --lineup 123

  # Show a lineup dispatch
  xbe view lineup-dispatches show 456`,
}

func init() {
	viewCmd.AddCommand(lineupDispatchesCmd)
}
