package cli

import "github.com/spf13/cobra"

var doJobProductionPlanJobSiteChangesCmd = &cobra.Command{
	Use:     "job-production-plan-job-site-changes",
	Aliases: []string{"job-production-plan-job-site-change"},
	Short:   "Manage job production plan job site changes",
	Long:    "Commands for creating job production plan job site changes.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanJobSiteChangesCmd)
}
