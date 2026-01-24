package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanTrailerAssignmentRecommendationsCmd = &cobra.Command{
	Use:     "project-transport-plan-trailer-assignment-recommendations",
	Aliases: []string{"project-transport-plan-trailer-assignment-recommendation"},
	Short:   "Generate trailer assignment recommendations",
	Long: `Generate project transport plan trailer assignment recommendations.

Recommendations rank candidate trailers for a project transport plan trailer.

Commands:
  create    Generate trailer assignment recommendations`,
	Example: `  # Generate recommendations for a trailer assignment
  xbe do project-transport-plan-trailer-assignment-recommendations create --project-transport-plan-trailer 123`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanTrailerAssignmentRecommendationsCmd)
}
