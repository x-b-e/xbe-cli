package cli

import "github.com/spf13/cobra"

var serviceTypeUnitOfMeasureCohortsCmd = &cobra.Command{
	Use:   "service-type-unit-of-measure-cohorts",
	Short: "Browse service type unit of measure cohorts",
	Long: `Browse service type unit of measure cohorts.

Service type unit of measure cohorts group service type unit of measures
for a customer and define a trigger that selects the cohort.

Commands:
  list    List cohorts with filtering and pagination
  show    Show cohort details`,
	Example: `  # List cohorts
  xbe view service-type-unit-of-measure-cohorts list

  # Show a cohort
  xbe view service-type-unit-of-measure-cohorts show 123`,
}

func init() {
	viewCmd.AddCommand(serviceTypeUnitOfMeasureCohortsCmd)
}
