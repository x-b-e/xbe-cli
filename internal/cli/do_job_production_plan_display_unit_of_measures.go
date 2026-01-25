package cli

import "github.com/spf13/cobra"

var doJobProductionPlanDisplayUnitOfMeasuresCmd = &cobra.Command{
	Use:     "job-production-plan-display-unit-of-measures",
	Aliases: []string{"job-production-plan-display-unit-of-measure"},
	Short:   "Manage job production plan display unit of measures",
	Long: `Manage job production plan display unit of measures.

Display unit of measures determine which units are shown for a job production
plan and how they are ordered by importance.

Commands:
  create  Add a display unit of measure
  update  Update the importance position
  delete  Remove a display unit of measure`,
	Example: `  # Add a unit of measure to a job production plan
  xbe do job-production-plan-display-unit-of-measures create \
    --job-production-plan 123 \
    --unit-of-measure 456 \
    --importance-position 0

  # Update importance position
  xbe do job-production-plan-display-unit-of-measures update 789 --importance-position 1

  # Delete a display unit of measure
  xbe do job-production-plan-display-unit-of-measures delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doJobProductionPlanDisplayUnitOfMeasuresCmd)
}
