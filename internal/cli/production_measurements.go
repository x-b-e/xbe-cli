package cli

import "github.com/spf13/cobra"

var productionMeasurementsCmd = &cobra.Command{
	Use:   "production-measurements",
	Short: "Browse production measurements",
	Long: `Browse production measurements on the XBE platform.

Production measurements capture width, depth, length, speed, density, and
pass counts for job production plan segments to calculate volume and rates.

Commands:
  list    List production measurements
  show    Show production measurement details`,
	Example: `  # List production measurements
  xbe view production-measurements list

  # Show a production measurement
  xbe view production-measurements show 123`,
}

func init() {
	viewCmd.AddCommand(productionMeasurementsCmd)
}
