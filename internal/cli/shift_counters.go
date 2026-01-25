package cli

import "github.com/spf13/cobra"

var shiftCountersCmd = &cobra.Command{
	Use:     "shift-counters",
	Aliases: []string{"shift-counter"},
	Short:   "Browse shift counters",
	Long: `Browse shift counters.

Shift counters report how many accepted tender job schedule shifts start after a
minimum timestamp. Counters are generated on demand via the create command and
are not persisted.`,
	Example: `  # List shift counters (typically empty)
  xbe view shift-counters list`,
}

func init() {
	viewCmd.AddCommand(shiftCountersCmd)
}
