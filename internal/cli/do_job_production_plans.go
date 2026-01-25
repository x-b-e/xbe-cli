package cli

import "github.com/spf13/cobra"

var doJobProductionPlansCmd = &cobra.Command{
	Use:     "job-production-plans",
	Aliases: []string{"job-production-plan", "jpp", "jpps"},
	Short:   "Manage job production plans",
	Long:    `Create and update job production plans.`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlansCmd)
}
