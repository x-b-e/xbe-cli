package cli

import "github.com/spf13/cobra"

var equipmentUtilizationReadingsCmd = &cobra.Command{
	Use:     "equipment-utilization-readings",
	Aliases: []string{"equipment-utilization-reading"},
	Short:   "View equipment utilization readings",
	Long: `View equipment utilization readings.

Equipment utilization readings capture reported odometer and hourmeter values
for equipment. Readings can be filtered by equipment, business unit, user,
source, and reported-at range.

Commands:
  list    List readings with filtering
  show    Show reading details`,
}

func init() {
	viewCmd.AddCommand(equipmentUtilizationReadingsCmd)
}
