package cli

import "github.com/spf13/cobra"

var doJobProductionPlanScrappagesCmd = &cobra.Command{
	Use:   "job-production-plan-scrappages",
	Short: "Manage job production plan scrappages",
	Long: `Create job production plan scrappages.

Scrappages move approved job production plans to a scrapped status.

Commands:
  create    Scrap a job production plan`,
	Example: `  # Scrap a job production plan
  xbe do job-production-plan-scrappages create --job-production-plan 123 --comment "Plan cancelled"`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanScrappagesCmd)
}
