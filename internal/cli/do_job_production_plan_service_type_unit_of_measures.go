package cli

import "github.com/spf13/cobra"

var doJobProductionPlanServiceTypeUnitOfMeasuresCmd = &cobra.Command{
	Use:     "job-production-plan-service-type-unit-of-measures",
	Aliases: []string{"job-production-plan-service-type-unit-of-measure"},
	Short:   "Manage job production plan service type unit of measures",
	Long: `Manage job production plan service type unit of measures.

Job production plan service type unit of measures configure step sizes and
invoice exclusions per service type unit of measure on a job production plan.

Commands:
  create  Add a service type unit of measure to a job production plan
  update  Update step size settings and relationships
  delete  Remove a service type unit of measure from a job production plan`,
	Example: `  # Add a service type unit of measure to a job production plan
  xbe do job-production-plan-service-type-unit-of-measures create \
    --job-production-plan 123 \
    --service-type-unit-of-measure 456 \
    --step-size no_step

  # Update step size settings
  xbe do job-production-plan-service-type-unit-of-measures update 789 \
    --step-size ceiling \
    --explicit-step-size-target "Tons" \
    --exclude-from-time-card-invoices

  # Delete a job production plan service type unit of measure
  xbe do job-production-plan-service-type-unit-of-measures delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanServiceTypeUnitOfMeasuresCmd)
}
