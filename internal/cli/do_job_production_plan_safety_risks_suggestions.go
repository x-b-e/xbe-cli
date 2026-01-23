package cli

import "github.com/spf13/cobra"

var doJobProductionPlanSafetyRisksSuggestionsCmd = &cobra.Command{
	Use:   "job-production-plan-safety-risks-suggestions",
	Short: "Generate job production plan safety risks suggestions",
	Long: `Generate job production plan safety risks suggestions.

Commands:
  create    Generate a safety risks suggestion`,
	Example: `  # Generate safety risks suggestions for a job production plan
  xbe do job-production-plan-safety-risks-suggestions create --job-production-plan 123

  # Generate with custom options
  xbe do job-production-plan-safety-risks-suggestions create \
    --job-production-plan 123 \
    --options '{"include_other_incidents":true}'

  # Generate synchronously (wait for risks)
  xbe do job-production-plan-safety-risks-suggestions create \
    --job-production-plan 123 \
    --is-async=false`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanSafetyRisksSuggestionsCmd)
}
