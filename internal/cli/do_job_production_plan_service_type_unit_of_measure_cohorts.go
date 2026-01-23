package cli

import "github.com/spf13/cobra"

var doJobProductionPlanServiceTypeUnitOfMeasureCohortsCmd = &cobra.Command{
	Use:     "job-production-plan-service-type-unit-of-measure-cohorts",
	Aliases: []string{"job-production-plan-service-type-unit-of-measure-cohort"},
	Short:   "Manage job production plan service type unit of measure cohorts",
	Long: `Create and delete job production plan service type unit of measure cohort links.

Commands:
  create    Create a cohort link
  delete    Delete a cohort link`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanServiceTypeUnitOfMeasureCohortsCmd)
}
