package cli

import "github.com/spf13/cobra"

var doJobProductionPlanSafetyRiskCommunicationSuggestionsCmd = &cobra.Command{
	Use:   "job-production-plan-safety-risk-communication-suggestions",
	Short: "Generate safety risk communication suggestions",
	Long: `Create safety risk communication suggestions for job production plans.

Commands:
  create    Generate a safety risk communication suggestion
  delete    Delete a safety risk communication suggestion`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanSafetyRiskCommunicationSuggestionsCmd)
}
