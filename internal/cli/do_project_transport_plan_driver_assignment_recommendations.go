package cli

import "github.com/spf13/cobra"

var doProjectTransportPlanDriverAssignmentRecommendationsCmd = &cobra.Command{
	Use:     "project-transport-plan-driver-assignment-recommendations",
	Aliases: []string{"project-transport-plan-driver-assignment-recommendation"},
	Short:   "Generate driver assignment recommendations",
	Long: `Generate project transport plan driver assignment recommendations.

Recommendations rank candidate drivers for a project transport plan driver.

Commands:
  create    Generate driver assignment recommendations`,
	Example: `  # Generate recommendations for a driver assignment
  xbe do project-transport-plan-driver-assignment-recommendations create --project-transport-plan-driver 123`,
}

func init() {
	doCmd.AddCommand(doProjectTransportPlanDriverAssignmentRecommendationsCmd)
}
