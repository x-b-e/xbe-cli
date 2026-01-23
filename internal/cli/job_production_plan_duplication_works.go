package cli

import "github.com/spf13/cobra"

var jobProductionPlanDuplicationWorksCmd = &cobra.Command{
	Use:   "job-production-plan-duplication-works",
	Short: "View job production plan duplication work",
	Long: `View job production plan duplication work items.

These records track async job production plan duplication requests and outcomes.

Commands:
  list    List duplication work records with filtering
  show    Show duplication work details`,
	Example: `  # List duplication work
  xbe view job-production-plan-duplication-works list

  # Filter by template ID
  xbe view job-production-plan-duplication-works list --job-production-plan-template-id 123

  # Show duplication work details
  xbe view job-production-plan-duplication-works show 456`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanDuplicationWorksCmd)
}
