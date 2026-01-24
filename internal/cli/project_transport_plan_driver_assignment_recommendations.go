package cli

import "github.com/spf13/cobra"

var projectTransportPlanDriverAssignmentRecommendationsCmd = &cobra.Command{
	Use:     "project-transport-plan-driver-assignment-recommendations",
	Aliases: []string{"project-transport-plan-driver-assignment-recommendation"},
	Short:   "Browse project transport plan driver assignment recommendations",
	Long: `Browse project transport plan driver assignment recommendations.

Recommendations rank candidate drivers for a project transport plan driver
assignment based on scoring rules.

Commands:
  list    List project transport plan driver assignment recommendations
  show    Show project transport plan driver assignment recommendation details`,
	Example: `  # List recommendations
  xbe view project-transport-plan-driver-assignment-recommendations list

  # Filter by project transport plan driver
  xbe view project-transport-plan-driver-assignment-recommendations list --project-transport-plan-driver 123

  # Show recommendation details
  xbe view project-transport-plan-driver-assignment-recommendations show 456`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanDriverAssignmentRecommendationsCmd)
}
