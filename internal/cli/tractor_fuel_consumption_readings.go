package cli

import "github.com/spf13/cobra"

var tractorFuelConsumptionReadingsCmd = &cobra.Command{
	Use:   "tractor-fuel-consumption-readings",
	Short: "Browse tractor fuel consumption readings",
	Long: `Browse tractor fuel consumption readings on the XBE platform.

Tractor fuel consumption readings capture fuel usage values for tractors,
optionally tied to a driver day and unit of measure.

Commands:
  list    List tractor fuel consumption readings
  show    Show tractor fuel consumption reading details`,
	Example: `  # List tractor fuel consumption readings
  xbe view tractor-fuel-consumption-readings list

  # Show a tractor fuel consumption reading
  xbe view tractor-fuel-consumption-readings show 123`,
}

func init() {
	viewCmd.AddCommand(tractorFuelConsumptionReadingsCmd)
}
