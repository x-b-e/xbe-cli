package cli

import "github.com/spf13/cobra"

var jobProductionPlanSafetyRiskCommunicationSuggestionsCmd = &cobra.Command{
	Use:   "job-production-plan-safety-risk-communication-suggestions",
	Short: "Browse safety risk communication suggestions",
	Long: `Browse job production plan safety risk communication suggestions on the XBE platform.

Safety risk communication suggestions are generated plans for communicating
safety risks and remediation strategies to crews.

Commands:
  list    List safety risk communication suggestions
  show    Show safety risk communication suggestion details`,
	Example: `  # List suggestions for a job production plan
  xbe view job-production-plan-safety-risk-communication-suggestions list --job-production-plan 123

  # Show suggestion details
  xbe view job-production-plan-safety-risk-communication-suggestions show 456`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanSafetyRiskCommunicationSuggestionsCmd)
}
