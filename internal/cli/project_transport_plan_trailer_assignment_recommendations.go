package cli

import "github.com/spf13/cobra"

var projectTransportPlanTrailerAssignmentRecommendationsCmd = &cobra.Command{
	Use:     "project-transport-plan-trailer-assignment-recommendations",
	Aliases: []string{"project-transport-plan-trailer-assignment-recommendation"},
	Short:   "Browse project transport plan trailer assignment recommendations",
	Long: `Browse project transport plan trailer assignment recommendations.

Recommendations rank candidate trailers for a project transport plan trailer
assignment based on scoring rules.

Commands:
  list    List project transport plan trailer assignment recommendations
  show    Show project transport plan trailer assignment recommendation details`,
	Example: `  # List recommendations
  xbe view project-transport-plan-trailer-assignment-recommendations list

  # Filter by project transport plan trailer
  xbe view project-transport-plan-trailer-assignment-recommendations list --project-transport-plan-trailer 123

  # Show recommendation details
  xbe view project-transport-plan-trailer-assignment-recommendations show 456`,
}

func init() {
	viewCmd.AddCommand(projectTransportPlanTrailerAssignmentRecommendationsCmd)
}
