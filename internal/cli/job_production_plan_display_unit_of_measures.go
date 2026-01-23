package cli

import "github.com/spf13/cobra"

var jobProductionPlanDisplayUnitOfMeasuresCmd = &cobra.Command{
	Use:     "job-production-plan-display-unit-of-measures",
	Aliases: []string{"job-production-plan-display-unit-of-measure"},
	Short:   "View job production plan display unit of measures",
	Long: `View job production plan display unit of measures.

Display unit of measures control which units appear on job production plan
reporting and their relative importance order.

Commands:
  list    List job production plan display unit of measures
  show    Show job production plan display unit of measure details`,
	Example: `  # List display unit of measures
  xbe view job-production-plan-display-unit-of-measures list

  # Show a display unit of measure
  xbe view job-production-plan-display-unit-of-measures show 123`,
}

func init() {
	viewCmd.AddCommand(jobProductionPlanDisplayUnitOfMeasuresCmd)
}
