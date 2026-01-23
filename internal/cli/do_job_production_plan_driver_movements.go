package cli

import "github.com/spf13/cobra"

var doJobProductionPlanDriverMovementsCmd = &cobra.Command{
	Use:   "job-production-plan-driver-movements",
	Short: "Generate job production plan driver movements",
	Long: `Generate driver movement details for a job production plan.

Driver movements compute location segment data for a driver on a job production
plan. Provide a job production plan and either a driver or tender job schedule
shift to scope the results.

Commands:
  create    Generate driver movement details`,
	Example: `  # Generate movement for a driver
  xbe do job-production-plan-driver-movements create --job-production-plan 123 --driver 456

  # Generate movement for a shift and bust cache
  xbe do job-production-plan-driver-movements create \
    --job-production-plan 123 \
    --tender-job-schedule-shift 789 \
    --bust-cache`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanDriverMovementsCmd)
}
