package cli

import "github.com/spf13/cobra"

var truckerShiftSetsCmd = &cobra.Command{
	Use:     "trucker-shift-sets",
	Aliases: []string{"trucker-shift-set"},
	Short:   "Browse trucker shift sets",
	Long: `Browse trucker shift sets (driver days) on the XBE platform.

Trucker shift sets group one or more shifts for a trucker/driver on a date and
track time sheets, constraints, and equipment.

Commands:
  list    List trucker shift sets with filtering
  show    Show trucker shift set details`,
	Example: `  # List trucker shift sets
  xbe view trucker-shift-sets list

  # Show a trucker shift set
  xbe view trucker-shift-sets show 123`,
}

func init() {
	viewCmd.AddCommand(truckerShiftSetsCmd)
}
