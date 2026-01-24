package cli

import "github.com/spf13/cobra"

var doProjectPhaseDatesEstimatesCmd = &cobra.Command{
	Use:   "project-phase-dates-estimates",
	Short: "Manage project phase dates estimates",
	Long: `Create, update, and delete project phase dates estimates.

Project phase dates estimates define estimated start and end dates for a
project phase within a project estimate set.

Commands:
  create  Create a new dates estimate
  update  Update an existing dates estimate
  delete  Delete a dates estimate`,
	Example: `  # Create a dates estimate
  xbe do project-phase-dates-estimates create \
    --project-phase 123 \
    --project-estimate-set 456 \
    --start-date 2025-01-01 \
    --end-date 2025-01-15

  # Update the end date
  xbe do project-phase-dates-estimates update 789 --end-date 2025-02-01

  # Delete a dates estimate
  xbe do project-phase-dates-estimates delete 789 --confirm`,
}

func init() {
	doCmd.AddCommand(doProjectPhaseDatesEstimatesCmd)
}
