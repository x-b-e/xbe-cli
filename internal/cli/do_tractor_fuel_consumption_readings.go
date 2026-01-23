package cli

import "github.com/spf13/cobra"

var doTractorFuelConsumptionReadingsCmd = &cobra.Command{
	Use:   "tractor-fuel-consumption-readings",
	Short: "Manage tractor fuel consumption readings",
	Long: `Manage tractor fuel consumption readings.

Commands:
  create   Create a tractor fuel consumption reading
  update   Update a tractor fuel consumption reading
  delete   Delete a tractor fuel consumption reading`,
	Example: `  # Create a tractor fuel consumption reading
  xbe do tractor-fuel-consumption-readings create --tractor 123 --unit-of-measure 45 --value 12.5

  # Update a tractor fuel consumption reading
  xbe do tractor-fuel-consumption-readings update 456 --value 10

  # Delete a tractor fuel consumption reading
  xbe do tractor-fuel-consumption-readings delete 456 --confirm`,
}

func init() {
	doCmd.AddCommand(doTractorFuelConsumptionReadingsCmd)
}
