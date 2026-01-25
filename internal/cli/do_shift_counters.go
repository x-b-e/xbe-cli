package cli

import "github.com/spf13/cobra"

var doShiftCountersCmd = &cobra.Command{
	Use:     "shift-counters",
	Aliases: []string{"shift-counter"},
	Short:   "Count accepted shifts",
	Long: `Count accepted tender job schedule shifts on the XBE platform.

Commands:
  create    Count accepted shifts after a minimum start timestamp`,
	Example: `  # Count accepted shifts (default start)
  xbe do shift-counters create

  # Count accepted shifts after a date
  xbe do shift-counters create --start-at-min 2025-01-01T00:00:00Z`,
}

func init() {
	doCmd.AddCommand(doShiftCountersCmd)
}
