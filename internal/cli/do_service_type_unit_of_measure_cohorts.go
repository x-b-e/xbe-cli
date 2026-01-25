package cli

import "github.com/spf13/cobra"

var doServiceTypeUnitOfMeasureCohortsCmd = &cobra.Command{
	Use:   "service-type-unit-of-measure-cohorts",
	Short: "Manage service type unit of measure cohorts",
	Long: `Manage service type unit of measure cohorts on the XBE platform.

Commands:
  create    Create a new cohort
  update    Update an existing cohort
  delete    Delete a cohort`,
	Example: `  # Create a cohort
  xbe do service-type-unit-of-measure-cohorts create --customer 123 --trigger 456 --service-type-unit-of-measure-ids 789

  # Update a cohort name
  xbe do service-type-unit-of-measure-cohorts update 123 --name "Updated cohort"

  # Delete a cohort (requires --confirm)
  xbe do service-type-unit-of-measure-cohorts delete 123 --confirm`,
}

func init() {
	doCmd.AddCommand(doServiceTypeUnitOfMeasureCohortsCmd)
}
