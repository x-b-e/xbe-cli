package cli

import "github.com/spf13/cobra"

var doJobProductionPlanRecapGenerationsCmd = &cobra.Command{
	Use:     "job-production-plan-recap-generations",
	Aliases: []string{"job-production-plan-recap-generation"},
	Short:   "Generate job production plan recaps",
	Long:    "Commands for generating job production plan recaps.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanRecapGenerationsCmd)
}
