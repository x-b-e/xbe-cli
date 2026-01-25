package cli

import "github.com/spf13/cobra"

var jobProductionPlanServiceTypeUnitOfMeasuresCmd = &cobra.Command{
	Use:     "job-production-plan-service-type-unit-of-measures",
	Aliases: []string{"job-production-plan-service-type-unit-of-measure"},
	Short:   "View job production plan service type unit of measures",
	Long: `View job production plan service type unit of measures.

Job production plan service type unit of measures configure step sizes and
invoice exclusions per service type unit of measure on a job production plan.

Commands:
  list    List job production plan service type unit of measures
  show    Show job production plan service type unit of measure details`,
	Example: `  # List job production plan service type unit of measures
  xbe view job-production-plan-service-type-unit-of-measures list

  # Show a job production plan service type unit of measure
  xbe view job-production-plan-service-type-unit-of-measures show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanServiceTypeUnitOfMeasuresCmd)
}
