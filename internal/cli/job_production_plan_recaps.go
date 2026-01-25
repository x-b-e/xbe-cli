package cli

import "github.com/spf13/cobra"

var jobProductionPlanRecapsCmd = &cobra.Command{
	Use:   "job-production-plan-recaps",
	Short: "Browse job production plan recaps",
	Long: `Browse job production plan recaps on the XBE platform.

Job production plan recaps contain generated markdown summaries for a plan.

Commands:
  list    List job production plan recaps
  show    Show job production plan recap details`,
	Example: `  # List recaps
  xbe view job-production-plan-recaps list

  # Show recap details
  xbe view job-production-plan-recaps show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanRecapsCmd)
}
