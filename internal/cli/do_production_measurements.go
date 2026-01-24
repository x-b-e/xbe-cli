package cli

import "github.com/spf13/cobra"

var doProductionMeasurementsCmd = &cobra.Command{
	Use:   "production-measurements",
	Short: "Manage production measurements",
	Long: `Create, update, and delete production measurements.

Production measurements capture dimensional and speed inputs used to compute
production volumes and rates for job production plan segments.

Commands:
  create    Create a production measurement
  update    Update a production measurement
  delete    Delete a production measurement`,
}

func init() {
	doCmd.AddCommand(doProductionMeasurementsCmd)
}
