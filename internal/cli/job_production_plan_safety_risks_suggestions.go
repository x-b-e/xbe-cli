package cli

import "github.com/spf13/cobra"

var jobProductionPlanSafetyRisksSuggestionsCmd = &cobra.Command{
	Use:     "job-production-plan-safety-risks-suggestions",
	Aliases: []string{"job-production-plan-safety-risks-suggestion"},
	Short:   "Browse job production plan safety risks suggestions",
	Long: `Browse job production plan safety risks suggestions on the XBE platform.

Safety risks suggestions capture AI-generated safety risk lists for a job
production plan. Suggestions are usually generated asynchronously.

Commands:
  list    List safety risks suggestions with filtering and pagination
  show    Show safety risks suggestion details`,
	Example: `  # List safety risks suggestions
  xbe view job-production-plan-safety-risks-suggestions list

  # Filter by job production plan
  xbe view job-production-plan-safety-risks-suggestions list --job-production-plan 123

  # Show a suggestion
  xbe view job-production-plan-safety-risks-suggestions show 456

  # Output as JSON
  xbe view job-production-plan-safety-risks-suggestions list --json`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanSafetyRisksSuggestionsCmd)
}
