package cli

import "github.com/spf13/cobra"

var lineupDispatchShiftsCmd = &cobra.Command{
	Use:     "lineup-dispatch-shifts",
	Aliases: []string{"lineup-dispatch-shift"},
	Short:   "Browse lineup dispatch shifts",
	Long: `Browse lineup dispatch shifts.

Lineup dispatch shifts connect lineup dispatches to lineup job schedule shifts
and track fulfillment or cancellation.`,
	Example: `  # List lineup dispatch shifts
  xbe view lineup-dispatch-shifts list

  # Filter by lineup dispatch
  xbe view lineup-dispatch-shifts list --lineup-dispatch 123

  # Show a lineup dispatch shift
  xbe view lineup-dispatch-shifts show 456`,
}

func init() {
	viewCmd.AddCommand(lineupDispatchShiftsCmd)
}
