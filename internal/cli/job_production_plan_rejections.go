package cli

import "github.com/spf13/cobra"

var jobProductionPlanRejectionsCmd = &cobra.Command{
	Use:     "job-production-plan-rejections",
	Aliases: []string{"job-production-plan-rejection"},
	Short:   "View job production plan rejections",
	Long: `View job production plan rejections.

Rejections record a status change to rejected for a job production plan and may
include a comment.

Commands:
  list    List job production plan rejections
  show    Show job production plan rejection details`,
	Example: `  # List rejections
  xbe view job-production-plan-rejections list

  # Show a rejection
  xbe view job-production-plan-rejections show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanRejectionsCmd)
}
