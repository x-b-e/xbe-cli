package cli

import "github.com/spf13/cobra"

var projectPhaseDatesEstimatesCmd = &cobra.Command{
	Use:   "project-phase-dates-estimates",
	Short: "View project phase dates estimates",
	Long: `View project phase dates estimates on the XBE platform.

Project phase dates estimates capture estimated start and end dates for a project
phase within a project estimate set.

Commands:
  list    List date estimates
  show    Show date estimate details`,
	Example: `  # List date estimates
  xbe view project-phase-dates-estimates list

  # Filter by project phase
  xbe view project-phase-dates-estimates list --project-phase 123

  # Show a date estimate
  xbe view project-phase-dates-estimates show 456`,
}

func init() {
	viewCmd.AddCommand(projectPhaseDatesEstimatesCmd)
}
