package cli

import "github.com/spf13/cobra"

var driverDayShortfallCalculationsCmd = &cobra.Command{
	Use:     "driver-day-shortfall-calculations",
	Aliases: []string{"driver-day-shortfall-calculation"},
	Short:   "Browse driver day shortfall calculations",
	Long: `Browse driver day shortfall calculations.

Driver day shortfall calculations allocate shortfall quantities across a set of
time cards and constraints. These calculations are generated on demand via the
create command and are not persisted.`,
	Example: `  # List calculations (typically empty)
  xbe view driver-day-shortfall-calculations list

  # Show a calculation
  xbe view driver-day-shortfall-calculations show <id>`,
}

func init() {
	viewCmd.AddCommand(driverDayShortfallCalculationsCmd)
}
