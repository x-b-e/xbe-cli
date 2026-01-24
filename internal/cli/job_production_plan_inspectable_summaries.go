package cli

import "github.com/spf13/cobra"

var jobProductionPlanInspectableSummariesCmd = &cobra.Command{
	Use:     "job-production-plan-inspectable-summaries",
	Aliases: []string{"job-production-plan-inspectable-summary"},
	Short:   "Browse job production plan inspectable summaries",
	Long: `Browse job production plan inspectable summaries.

Inspectable summaries provide inspection-ready job production plan context,
including schedule details, site information, and inspection eligibility.

Commands:
  list    List inspectable summaries
  show    Show inspectable summary details`,
	Example: `  # List inspectable summaries
  xbe view job-production-plan-inspectable-summaries list

  # Show a job production plan inspectable summary
  xbe view job-production-plan-inspectable-summaries show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanInspectableSummariesCmd)
}
