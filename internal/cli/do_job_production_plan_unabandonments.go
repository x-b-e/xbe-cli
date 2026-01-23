package cli

import "github.com/spf13/cobra"

var doJobProductionPlanUnabandonmentsCmd = &cobra.Command{
	Use:     "job-production-plan-unabandonments",
	Aliases: []string{"job-production-plan-unabandonment"},
	Short:   "Unabandon job production plans",
	Long:    "Commands for unabandoning job production plans.",
}

func init() {
	doCmd.AddCommand(doJobProductionPlanUnabandonmentsCmd)
}
