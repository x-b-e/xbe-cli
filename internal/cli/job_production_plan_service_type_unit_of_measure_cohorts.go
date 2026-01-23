package cli

import "github.com/spf13/cobra"

var jobProductionPlanServiceTypeUnitOfMeasureCohortsCmd = &cobra.Command{
	Use:     "job-production-plan-service-type-unit-of-measure-cohorts",
	Aliases: []string{"job-production-plan-service-type-unit-of-measure-cohort"},
	Short:   "View job production plan service type unit of measure cohorts",
	Long: `View job production plan service type unit of measure cohorts.

Job production plan service type unit of measure cohorts link a job production plan
with a service type unit of measure cohort.

Commands:
  list    List job production plan service type unit of measure cohorts
  show    Show job production plan service type unit of measure cohort details`,
	Example: `  # List job production plan service type unit of measure cohorts
  xbe view job-production-plan-service-type-unit-of-measure-cohorts list

  # Show a specific cohort link
  xbe view job-production-plan-service-type-unit-of-measure-cohorts show 123

  # Output JSON
  xbe view job-production-plan-service-type-unit-of-measure-cohorts list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanServiceTypeUnitOfMeasureCohortsCmd)
}
